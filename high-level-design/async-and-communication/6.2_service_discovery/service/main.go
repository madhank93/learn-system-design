package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func register(consulAddr, id, name, addr string, port int) {
	body, _ := json.Marshal(map[string]any{
		"ID":      id,
		"Name":    name,
		"Address": addr,
		"Port":    port,
		"Check": map[string]string{
			"HTTP":                           fmt.Sprintf("http://%s:%d/health", addr, port),
			"Interval":                       "5s",
			"DeregisterCriticalServiceAfter": "15s",
		},
	})
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest("PUT",
			fmt.Sprintf("http://%s/v1/agent/service/register", consulAddr),
			bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err == nil && resp.StatusCode == 200 {
			log.Printf("[%s] registered with Consul", id)
			return
		}
		log.Printf("[%s] consul not ready, retry %d", id, i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("[%s] could not register with Consul", id)
}

func main() {
	id := os.Getenv("SERVICE_ID")
	name := os.Getenv("SERVICE_NAME")
	addr := os.Getenv("SERVICE_ADDR")
	consul := os.Getenv("CONSUL_ADDR")
	if consul == "" {
		consul = "consul:8500"
	}

	register(consul, id, name, addr, 8080)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":%q,"name":%q}`+"\n", id, name)
		log.Printf("[%s] served request", id)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	log.Printf("[%s] listening on :8080", id)
	http.ListenAndServe(":8080", nil)
}
