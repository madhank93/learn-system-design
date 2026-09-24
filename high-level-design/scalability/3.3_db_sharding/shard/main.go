package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
)

type store struct {
	mu   sync.RWMutex
	data map[string]string
}

func main() {
	id := os.Getenv("SHARD_ID")
	s := &store{data: make(map[string]string)}

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		s.mu.RLock()
		val, ok := s.data[key]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "shard:%s key:%s val:%s\n", id, key, val)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		val := r.URL.Query().Get("val")
		s.mu.Lock()
		s.data[key] = val
		s.mu.Unlock()
		fmt.Fprintf(w, "shard:%s stored key:%s\n", id, key)
		log.Printf("[shard %s] set key:%s", id, key)
	})

	http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		keys := make([]string, 0, len(s.data))
		for k := range s.data {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
		sort.Strings(keys)
		json.NewEncoder(w).Encode(map[string]any{"shard": id, "keys": keys, "count": len(keys)})
	})

	log.Printf("Shard %s listening on :8080", id)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] shard %s: %v", id, err)
	}
}
