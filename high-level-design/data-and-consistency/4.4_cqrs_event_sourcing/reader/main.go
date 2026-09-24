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
const viewPrefix = "view:"

func project(ctx context.Context) {
	lastID := "0"
	for {
		msgs, err := rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{stream, lastID},
			Count:   100,
			Block:   2000 * time.Millisecond,
		}).Result()
		if err != nil && err != redis.Nil {
			time.Sleep(time.Second)
			continue
		}
		for _, s := range msgs {
			for _, msg := range s.Messages {
				uid := fmt.Sprint(msg.Values["user_id"])
				typ := fmt.Sprint(msg.Values["type"])
				rdb.HSet(ctx, viewPrefix+uid,
					"last_event", typ,
					"updated_at", msg.Values["occurred_at"],
				)
				rdb.HIncrBy(ctx, viewPrefix+uid, "event_count", 1)
				log.Printf("projected %s → user:%s type:%s", msg.ID, uid, typ)
				lastID = msg.ID
			}
		}
	}
}

func main() {
	rdb = redis.NewClient(&redis.Options{Addr: "redis:6379"})
	go project(context.Background())

	http.HandleFunc("/state/", func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Path[len("/state/"):]
		ctx := context.Background()
		data, err := rdb.HGetAll(ctx, viewPrefix+uid).Result()
		if err != nil || len(data) == 0 {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		data["user_id"] = uid
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	log.Println("Reader on :8080")
	http.ListenAndServe(":8080", nil)
}
