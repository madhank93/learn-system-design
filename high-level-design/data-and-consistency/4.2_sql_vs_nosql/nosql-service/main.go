package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Document is a flexible map — any fields allowed.
type Document map[string]any

var (
	mu    sync.RWMutex
	store = map[string]Document{}
)

func main() {
	// PUT /doc/:id — upsert a document (any JSON shape)
	http.HandleFunc("/doc/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/doc/"):]
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}
		if r.Method == "PUT" || r.Method == "POST" {
			var doc Document
			if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			doc["_id"] = id
			mu.Lock()
			store[id] = doc
			mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(doc)
			log.Printf("upserted doc:%s fields:%d", id, len(doc))
			return
		}
		// GET
		mu.RLock()
		doc, ok := store[id]
		mu.RUnlock()
		if !ok {
			http.Error(w, "not found", 404)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	})

	// GET /docs — list all documents (no JOIN needed — data already denormalized)
	http.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		mu.RLock()
		docs := make([]Document, 0, len(store))
		for _, d := range store {
			docs = append(docs, d)
		}
		mu.RUnlock()
		w.Header().Set("X-Query-Type", "full-scan")
		json.NewEncoder(w).Encode(docs)
	})

	// GET /search?field=name&val=Alice — simple field scan
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		field := r.URL.Query().Get("field")
		val := r.URL.Query().Get("val")
		mu.RLock()
		var results []Document
		for _, d := range store {
			if fmt.Sprint(d[field]) == val {
				results = append(results, d)
			}
		}
		mu.RUnlock()
		w.Header().Set("X-Query-Type", "field-scan")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("NoSQL service on :8080")
	http.ListenAndServe(":8080", nil)
}
