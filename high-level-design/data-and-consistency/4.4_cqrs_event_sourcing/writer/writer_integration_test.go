//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func startRedisForCQRS(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Redis endpoint error: %v", err)
	}
	return redis.NewClient(&redis.Options{Addr: endpoint})
}

func TestIntegrationCommandAppendsToStream(t *testing.T) {
	rdb = startRedisForCQRS(t)
	ctx := context.Background()

	mux := http.NewServeMux()
	mux.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		var cmd command
		json.NewDecoder(r.Body).Decode(&cmd)
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{
				"type":    cmd.Type,
				"user_id": cmd.UserID,
				"payload": cmd.Payload,
			},
		})
		w.WriteHeader(200)
	})

	body, _ := json.Marshal(command{Type: "user.created", UserID: "u1", Payload: "alice"})
	req := httptest.NewRequest("POST", "/command", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	// Verify event is in the stream
	msgs, err := rdb.XRange(ctx, stream, "-", "+").Result()
	if err != nil {
		t.Fatalf("XRANGE error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 event in stream, got %d", len(msgs))
	}
	if msgs[0].Values["type"] != "user.created" {
		t.Fatalf("expected type=user.created, got %v", msgs[0].Values["type"])
	}
	if msgs[0].Values["user_id"] != "u1" {
		t.Fatalf("expected user_id=u1, got %v", msgs[0].Values["user_id"])
	}
}

func TestIntegrationMultipleEventsInStream(t *testing.T) {
	rdb = startRedisForCQRS(t)
	ctx := context.Background()

	events := []command{
		{Type: "user.created", UserID: "u1", Payload: "alice"},
		{Type: "user.updated", UserID: "u1", Payload: "alice@example.com"},
		{Type: "user.created", UserID: "u2", Payload: "bob"},
	}

	for _, e := range events {
		rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{"type": e.Type, "user_id": e.UserID, "payload": e.Payload},
		})
	}

	msgs, _ := rdb.XRange(ctx, stream, "-", "+").Result()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 events, got %d", len(msgs))
	}

	// Events are ordered (append-only log)
	if msgs[0].Values["type"] != "user.created" {
		t.Fatalf("first event should be user.created")
	}
	if msgs[1].Values["type"] != "user.updated" {
		t.Fatalf("second event should be user.updated")
	}
}
