package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	shards   []string
	strategy string
)

func hashShard(key string) string {
	h := md5.Sum([]byte(key))
	return shards[int(h[0])%len(shards)]
}

func rangeShard(key string) string {
	id, err := strconv.Atoi(key)
	if err != nil {
		return shards[0]
	}
	// 0–332 → shard[0], 333–665 → shard[1], 666+ → shard[2]
	idx := id / 333
	if idx >= len(shards) {
		idx = len(shards) - 1
	}
	return shards[idx]
}

func route(key string) string {
	if strategy == "range" {
		return rangeShard(key)
	}
	return hashShard(key)
}

func forward(w http.ResponseWriter, r *http.Request, shard string) {
	url := fmt.Sprintf("http://%s%s", shard, r.RequestURI)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	strategy = os.Getenv("STRATEGY")
	shards = strings.Split(os.Getenv("SHARDS"), ",")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing ?key=", http.StatusBadRequest)
			return
		}
		shard := route(key)
		log.Printf("[%s] key:%s → %s", strategy, key, shard)
		forward(w, r, shard)
	})

	log.Printf("Router (%s) on :8080 | shards: %v", strategy, shards)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] router: %v", err)
	}
}
