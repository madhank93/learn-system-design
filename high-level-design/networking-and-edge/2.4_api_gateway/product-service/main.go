package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"service":"product-service","products":["widget","gadget","gizmo"]}`)
		log.Printf("GET /products")
	})

	http.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/products/"):]
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"product-service","id":"%s","name":"Product %s"}`+"\n", id, id)
	})

	log.Println("product-service listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] product-service: %v", err)
	}
}
