package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var failing int32

func main() {
	id := os.Getenv("SERVER_ID")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&failing) == 1 {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("[%s] 500 (failing mode)", id)
			return
		}
		fmt.Fprintf(w, "Server %s OK\n", id)
		log.Printf("[%s] 200 OK", id)
	})

	http.HandleFunc("/admin/fail", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt32(&failing, 1)
		fmt.Fprintln(w, "now failing")
		log.Printf("[%s] → failing mode", id)
	})

	http.HandleFunc("/admin/recover", func(w http.ResponseWriter, r *http.Request) {
		atomic.StoreInt32(&failing, 0)
		fmt.Fprintln(w, "now healthy")
		log.Printf("[%s] → healthy mode", id)
	})

	log.Printf("Backend %s on :8080", id)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] backend %s: %v", id, err)
	}
}
