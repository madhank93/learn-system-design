package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	id := os.Getenv("SERVER_ID")
	if id == "" {
		id = "server"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Server %s | path:%s\n", id, r.URL.Path)
		log.Printf("[%s] %s %s", id, r.Method, r.URL.Path)
	})

	log.Printf("Server %s listening on :8080", id)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] %v", err)
	}
}
