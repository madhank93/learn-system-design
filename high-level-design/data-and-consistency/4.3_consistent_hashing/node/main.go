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
	id := os.Getenv("NODE_ID")
	s := &store{data: make(map[string]string)}

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key, val := r.URL.Query().Get("key"), r.URL.Query().Get("val")
		s.mu.Lock()
		s.data[key] = val
		s.mu.Unlock()
		fmt.Fprintf(w, "node:%s stored key:%s\n", id, key)
		log.Printf("[%s] set %s=%s", id, key, val)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		s.mu.RLock()
		val, ok := s.data[key]
		s.mu.RUnlock()
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "node:%s key:%s val:%s\n", id, key, val)
	})

	http.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		s.mu.RLock()
		keys := make([]string, 0, len(s.data))
		for k := range s.data {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
		sort.Strings(keys)
		json.NewEncoder(w).Encode(map[string]any{"node": id, "keys": keys, "count": len(keys)})
	})

	log.Printf("Node %s on :8080", id)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] node %s: %v", id, err)
	}
}
