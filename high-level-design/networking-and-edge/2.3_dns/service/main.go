package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	id := os.Getenv("SERVER_ID")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "server:%s\n", id)
		log.Printf("[%s] served request", id)
	})

	log.Printf("Service %s on :8080", id)
	http.ListenAndServe(":8080", nil)
}
