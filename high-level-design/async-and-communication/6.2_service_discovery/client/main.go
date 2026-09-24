package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

var counter int64

type instance struct {
	Address string `json:"Address"`
	Port    int    `json:"Port"`
	ID      string `json:"ID"`
}

func discover(consulAddr, name string) []instance {
	resp, err := http.Get(fmt.Sprintf(
		"http://%s/v1/health/service/%s?passing=true", consulAddr, name,
	))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	var results []struct {
		Service instance `json:"Service"`
	}
	json.NewDecoder(resp.Body).Decode(&results)
	out := make([]instance, len(results))
	for i, r := range results {
		out[i] = r.Service
	}
	return out
}

func main() {
	consul := os.Getenv("CONSUL_ADDR")
	if consul == "" {
		consul = "consul:8500"
	}
	svcName := os.Getenv("SERVICE_NAME")

	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		instances := discover(consul, svcName)
		if len(instances) == 0 {
			http.Error(w, "no healthy instances", http.StatusServiceUnavailable)
			return
		}
		n := atomic.AddInt64(&counter, 1)
		svc := instances[int(n%int64(len(instances)))]
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/", svc.Address, svc.Port))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(w, "instances:%d called:%s:%d → %s",
			len(instances), svc.Address, svc.Port, body)
	})

	log.Println("Client on :8080")
	http.ListenAndServe(":8080", nil)
}
