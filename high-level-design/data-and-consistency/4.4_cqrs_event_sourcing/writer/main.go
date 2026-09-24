package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

const stream = "events"

type command struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	Payload string `json:"payload"`
}

func main() {
	rdb = redis.NewClient(&redis.Options{Addr: "redis:6379"})

	http.HandleFunc("/command", func(w http.ResponseWriter, r *http.Request) {
		var cmd command
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ctx := context.Background()
		id, err := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: stream,
			Values: map[string]any{
				"type":        cmd.Type,
				"user_id":     cmd.UserID,
				"payload":     cmd.Payload,
				"occurred_at": time.Now().UTC().Format(time.RFC3339),
			},
		}).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, `{"event_id":%q,"type":%q}`+"\n", id, cmd.Type)
		log.Printf("appended %s type:%s user:%s", id, cmd.Type, cmd.UserID)
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		msgs, _ := rdb.XRange(ctx, stream, "-", "+").Result()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(msgs)
	})

	log.Println("Writer on :8080")
	http.ListenAndServe(":8080", nil)
}
