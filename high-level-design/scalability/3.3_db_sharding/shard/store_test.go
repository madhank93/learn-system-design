package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	s := &store{data: make(map[string]string)}
	mux := http.NewServeMux()
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("val")
		s.mu.Lock()
		s.data[key] = val
		s.mu.Unlock()
		w.WriteHeader(200)
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		s.mu.RLock()
		val, ok := s.data[key]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Write([]byte(val))
	})

	// set
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/set?key=name&val=alice", nil))
	if rr.Code != 200 {
		t.Fatalf("set: expected 200, got %d", rr.Code)
	}

	// get
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/get?key=name", nil))
	if rr.Code != 200 {
		t.Fatalf("get: expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "alice" {
		t.Fatalf("get: expected 'alice', got %q", rr.Body.String())
	}
}

func TestGetNotFound(t *testing.T) {
	s := &store{data: make(map[string]string)}
	mux := http.NewServeMux()
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		s.mu.RLock()
		_, ok := s.data[key]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
		}
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/get?key=missing", nil))
	if rr.Code != 404 {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestKeysEndpoint(t *testing.T) {
	s := &store{data: map[string]string{"a": "1", "b": "2", "c": "3"}}
	mux := http.NewServeMux()
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		keys := make([]string, 0, len(s.data))
		for k := range s.data {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]any{"count": len(keys)})
	})

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, httptest.NewRequest("GET", "/keys", nil))
	if rr.Code != 200 {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if count := int(resp["count"].(float64)); count != 3 {
		t.Fatalf("expected count=3, got %d", count)
	}
}
