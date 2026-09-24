//go:build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func startRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	c, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("redis container: %v", err)
	}
	t.Cleanup(func() { c.Terminate(ctx) })
	ep, _ := c.Endpoint(ctx, "")
	return redis.NewClient(&redis.Options{Addr: ep})
}

func TestIntegrationProjectorBuildsView(t *testing.T) {
	rdb = startRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Seed 2 events for user u1
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"type": "user.created", "user_id": "u1", "payload": "alice", "occurred_at": time.Now().Format(time.RFC3339)},
	})
	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]any{"type": "user.updated", "user_id": "u1", "payload": "alice@example.com", "occurred_at": time.Now().Format(time.RFC3339)},
	})

	// Run projector
	go project(ctx)
	time.Sleep(300 * time.Millisecond) // let projector process

	// Verify materialized view
	data, err := rdb.HGetAll(ctx, viewPrefix+"u1").Result()
	if err != nil || len(data) == 0 {
		t.Fatalf("expected view for u1, got err=%v data=%v", err, data)
	}
	if data["last_event"] != "user.updated" {
		t.Fatalf("expected last_event=user.updated, got %q", data["last_event"])
	}
	count := data["event_count"]
	if count != "2" {
		t.Fatalf("expected event_count=2, got %q", count)
	}
}

func TestIntegrationProjectorMultipleUsers(t *testing.T) {
	rdb = startRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	users := []struct{ id, typ string }{
		{"u1", "user.created"},
		{"u2", "user.created"},
		{"u1", "user.updated"},
	}
	for _, e := range users {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{"type": e.typ, "user_id": e.id, "payload": "data", "occurred_at": time.Now().Format(time.RFC3339)},
		})
	}

	go project(ctx)
	time.Sleep(300 * time.Millisecond)

	u1Data, _ := rdb.HGetAll(ctx, viewPrefix+"u1").Result()
	u2Data, _ := rdb.HGetAll(ctx, viewPrefix+"u2").Result()

	if u1Data["event_count"] != "2" {
		t.Errorf("u1: expected event_count=2, got %q", u1Data["event_count"])
	}
	if u2Data["event_count"] != "1" {
		t.Errorf("u2: expected event_count=1, got %q", u2Data["event_count"])
	}
}

func TestIntegrationStateHTTPHandler(t *testing.T) {
	rdb = startRedis(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Pre-populate view directly (bypass projector for speed)
	rdb.HSet(ctx, viewPrefix+"u99", "last_event", "user.created", "event_count", "1", "updated_at", time.Now().Format(time.RFC3339))

	mux := http.NewServeMux()
	mux.HandleFunc("/state/", func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Path[len("/state/"):]
		data, err := rdb.HGetAll(ctx, viewPrefix+uid).Result()
		if err != nil || len(data) == 0 {
			http.Error(w, "not found", 404)
			return
		}
		data["user_id"] = uid
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/state/u99", nil))
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var result map[string]string
	json.Unmarshal(rr.Body.Bytes(), &result)
	if result["user_id"] != "u99" {
		t.Errorf("expected user_id=u99, got %q", result["user_id"])
	}
	if result["last_event"] != "user.created" {
		t.Errorf("expected last_event=user.created, got %q", result["last_event"])
	}
}

func TestIntegrationStateHTTPHandlerNotFound(t *testing.T) {
	rdb = startRedis(t)
	ctx := context.Background()
	_ = ctx

	mux := http.NewServeMux()
	mux.HandleFunc("/state/", func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Path[len("/state/"):]
		data, err := rdb.HGetAll(context.Background(), viewPrefix+uid).Result()
		if err != nil || len(data) == 0 {
			http.Error(w, "not found", 404)
			return
		}
		fmt.Fprint(w, "found")
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/state/nonexistent", nil))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
