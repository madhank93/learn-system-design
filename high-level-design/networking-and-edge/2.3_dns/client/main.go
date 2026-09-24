package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	http.HandleFunc("/resolve", func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("host")
		addrs, err := net.LookupHost(host)
		if err != nil {
			http.Error(w, fmt.Sprintf("lookup %s: %v", host, err), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "host:%s → %v\n", host, addrs)
		log.Printf("resolved %s → %v", host, addrs)
	})

	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		host := r.URL.Query().Get("host")
		resp, err := http.Get("http://" + host + ":8080/")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		addrs, _ := net.LookupHost(host)
		fmt.Fprintf(w, "resolved:%v → %s", addrs, body)
	})

	log.Println("Client on :8080")
	http.ListenAndServe(":8080", nil)
}
