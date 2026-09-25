package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	id := os.Getenv("NODE_ID")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	node := NewNode(id, redisAddr)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// HTTP: expose current status
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, node.Status(r.Context()))
	})
	go http.ListenAndServe(":8080", nil)

	log.Printf("[%s] starting election loop", id)
	node.Run(ctx)
}
