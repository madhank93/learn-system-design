package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type entry struct {
	Key     string    `json:"key"`
	Val     string    `json:"val"`
	Version int64     `json:"version"`
	At      time.Time `json:"at"`
}

var (
	mu      sync.RWMutex
	store   = map[string]entry{}
	version int64
)

func setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		key, val := r.URL.Query().Get("key"), r.URL.Query().Get("val")
		v := atomic.AddInt64(&version, 1)
		e := entry{Key: key, Val: val, Version: v, At: time.Now()}
		mu.Lock()
		store[key] = e
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(e)
		log.Printf("[primary] write key:%s v:%d", key, v)
	})

	mux.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		mu.RLock()
		e, ok := store[key]
		mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("X-Source", "primary")
		json.NewEncoder(w).Encode(e)
	})

	mux.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		var since int64
		fmt.Sscanf(r.URL.Query().Get("since"), "%d", &since)
		mu.RLock()
		var entries []entry
		for _, e := range store {
			if e.Version > since {
				entries = append(entries, e)
			}
		}
		mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	})
}

func main() {
	setupRoutes(http.DefaultServeMux)
	log.Println("Primary on :8080")
	http.ListenAndServe(":8080", nil)
}
