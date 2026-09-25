package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

var index = NewBTree()

func main() {
	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		key, _ := strconv.Atoi(r.URL.Query().Get("key"))
		val := r.URL.Query().Get("val")
		index.Insert(key, val)
		fmt.Fprintf(w, `{"inserted":{"key":%d,"val":%q}}`+"\n", key, val)
		log.Printf("insert key=%d val=%s", key, val)
	})

	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		key, _ := strconv.Atoi(r.URL.Query().Get("key"))
		val, ok := index.Search(key)
		if !ok {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"key": key, "val": val})
	})

	http.HandleFunc("/range", func(w http.ResponseWriter, r *http.Request) {
		lo, _ := strconv.Atoi(r.URL.Query().Get("lo"))
		hi, _ := strconv.Atoi(r.URL.Query().Get("hi"))
		results := index.RangeSearch(lo, hi)
		json.NewEncoder(w).Encode(map[string]any{"lo": lo, "hi": hi, "results": results, "count": len(results)})
	})

	log.Println("B-Tree index server on :8080")
	http.ListenAndServe(":8080", nil)
}
