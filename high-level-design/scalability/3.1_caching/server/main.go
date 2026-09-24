package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	id := os.Getenv("SERVER_ID")
	region := os.Getenv("REGION")
	if region == "" {
		region = "global"
	}

	http.HandleFunc("/item", func(w http.ResponseWriter, r *http.Request) {
		itemID := r.URL.Query().Get("id")
		time.Sleep(80 * time.Millisecond) // simulate DB read latency
		w.Header().Set("Cache-Control", "public, max-age=30")
		fmt.Fprintf(w, "Server %s (%s) | item:%s | generated:%d\n",
			id, region, itemID, time.Now().UnixMilli())
		log.Printf("[%s] served item:%s", id, itemID)
	})

	log.Printf("Server %s (%s) listening on :8080", id, region)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] server %s: %v", id, err)
	}
}
