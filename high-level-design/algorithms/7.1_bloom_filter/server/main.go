package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

var (
	mu sync.RWMutex
	bf = NewBloomFilter(10000, 0.01) // 10k expected items, 1% FP rate
)

func main() {
	// POST /add?item=X — add item to filter
	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		if item == "" {
			http.Error(w, "missing ?item=", 400)
			return
		}
		mu.Lock()
		bf.Add(item)
		mu.Unlock()
		fmt.Fprintf(w, `{"added":%q,"fp_rate":"%.4f"}`+"\n", item, bf.FalsePositiveRate())
		log.Printf("added: %s", item)
	})

	// GET /check?item=X — check membership
	http.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		item := r.URL.Query().Get("item")
		mu.RLock()
		found := bf.Contains(item)
		mu.RUnlock()
		verdict := "probably_yes"
		if !found {
			verdict = "definitely_no"
		}
		fmt.Fprintf(w, `{"item":%q,"result":%q,"note":%q}`+"\n",
			item, verdict,
			"false positives possible; false negatives impossible")
		log.Printf("check: %s → %s", item, verdict)
	})

	// GET /stats — filter stats
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		fp := bf.FalsePositiveRate()
		size := bf.size
		hfns := bf.hashFns
		mu.RUnlock()
		json.NewEncoder(w).Encode(map[string]any{
			"bit_array_size":  size,
			"hash_functions":  hfns,
			"fp_rate_current": fmt.Sprintf("%.4f", fp),
		})
	})

	log.Println("Bloom filter server on :8080")
	http.ListenAndServe(":8080", nil)
}
