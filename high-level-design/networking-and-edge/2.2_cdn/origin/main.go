package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// Static assets — long TTL, immutable
	http.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(60 * time.Millisecond) // simulate object store latency
		w.Header().Set("Cache-Control", "public, max-age=300, immutable")
		fmt.Fprintf(w, "<asset path=%q generated=%d>static content</asset>\n",
			r.URL.Path, time.Now().UnixMilli())
		log.Printf("ORIGIN /static/ %s", r.URL.Path)
	})

	// Dynamic content — never cache
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(80 * time.Millisecond) // simulate DB read
		w.Header().Set("Cache-Control", "private, no-store")
		fmt.Fprintf(w, `{"path":%q,"ts":%d,"note":"dynamic, not cacheable"}`+"\n",
			r.URL.Path, time.Now().UnixMilli())
		log.Printf("ORIGIN /api/ %s", r.URL.Path)
	})

	log.Println("Origin on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] origin: %v", err)
	}
}
