package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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

func replicate(primaryAddr string) {
	for {
		v := atomic.LoadInt64(&version)
		resp, err := http.Get(fmt.Sprintf("http://%s/log?since=%d", primaryAddr, v))
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		var entries []entry
		json.NewDecoder(resp.Body).Decode(&entries)
		resp.Body.Close()
		if len(entries) > 0 {
			mu.Lock()
			for _, e := range entries {
				store[e.Key] = e
				if e.Version > atomic.LoadInt64(&version) {
					atomic.StoreInt64(&version, e.Version)
				}
			}
			mu.Unlock()
			log.Printf("[replica] synced %d entries, version=%d", len(entries), atomic.LoadInt64(&version))
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func main() {
	id := os.Getenv("REPLICA_ID")
	primary := os.Getenv("PRIMARY_ADDR")
	if primary == "" {
		primary = "primary:8080"
	}

	go replicate(primary)

	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		mu.RLock()
		e, ok := store[key]
		mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("X-Source", id)
		json.NewEncoder(w).Encode(e)
	})

	http.HandleFunc("/lag", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "replica:%s version:%d\n", id, atomic.LoadInt64(&version))
	})

	log.Printf("Replica %s on :8080", id)
	http.ListenAndServe(":8080", nil)
}
