package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"service":"user-service","users":["alice","bob","carol"]}`)
		log.Printf("GET /users (from %s)", r.Header.Get("X-Real-IP"))
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/users/"):]
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"user-service","id":"%s","name":"User %s"}`+"\n", id, id)
	})

	log.Println("user-service listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] user-service: %v", err)
	}
}
