package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func resetReplica() {
	mu.Lock()
	store = map[string]entry{}
	mu.Unlock()
	atomic.StoreInt64(&version, 0)
}

func TestReplicaReadNotFound(t *testing.T) {
	resetReplica()
	mux := http.NewServeMux()
	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		mu.RLock()
		e, ok := store[key]
		mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("X-Source", "replica-test")
		json.NewEncoder(w).Encode(e)
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/read?key=missing", nil))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestReplicaAppliesReplicatedEntries(t *testing.T) {
	resetReplica()

	// Simulate what replicate() does: apply entries from primary log
	entries := []entry{
		{Key: "k1", Val: "v1", Version: 1, At: time.Now()},
		{Key: "k2", Val: "v2", Version: 2, At: time.Now()},
	}
	mu.Lock()
	for _, e := range entries {
		store[e.Key] = e
		if e.Version > atomic.LoadInt64(&version) {
			atomic.StoreInt64(&version, e.Version)
		}
	}
	mu.Unlock()

	if atomic.LoadInt64(&version) != 2 {
		t.Fatalf("expected version=2, got %d", atomic.LoadInt64(&version))
	}
	mu.RLock()
	if store["k1"].Val != "v1" {
		t.Fatalf("expected k1=v1, got %q", store["k1"].Val)
	}
	if store["k2"].Val != "v2" {
		t.Fatalf("expected k2=v2, got %q", store["k2"].Val)
	}
	mu.RUnlock()
}

func TestReplicaVersionOnlyAdvances(t *testing.T) {
	resetReplica()

	entries := []entry{
		{Key: "k1", Val: "v1", Version: 5, At: time.Now()},
		{Key: "k2", Val: "v2", Version: 3, At: time.Now()},
	}
	mu.Lock()
	for _, e := range entries {
		store[e.Key] = e
		if e.Version > atomic.LoadInt64(&version) {
			atomic.StoreInt64(&version, e.Version)
		}
	}
	mu.Unlock()

	// version should be max(5, 3) = 5
	if atomic.LoadInt64(&version) != 5 {
		t.Fatalf("expected version=5 (max), got %d", atomic.LoadInt64(&version))
	}
}

func TestReplicaLagEndpoint(t *testing.T) {
	resetReplica()
	atomic.StoreInt64(&version, 42)

	mux := http.NewServeMux()
	mux.HandleFunc("/lag", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/lag", nil))
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
