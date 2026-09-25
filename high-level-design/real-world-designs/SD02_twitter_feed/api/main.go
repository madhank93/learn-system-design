package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var feed *Feed

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	feed = NewFeed(rdb)

	// POST /tweet?user=alice&text=Hello+world
	http.HandleFunc("/tweet", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		text := r.URL.Query().Get("text")
		if user == "" || text == "" {
			http.Error(w, "missing user or text", 400)
			return
		}
		t, err := feed.Post(r.Context(), user, text)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
		log.Printf("tweet by %s: %s", user, text)
	})

	// POST /follow?from=bob&to=alice
	http.HandleFunc("/follow", func(w http.ResponseWriter, r *http.Request) {
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		feed.Follow(r.Context(), from, to)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"follower": from, "followee": to,
			"followee_followers": feed.FollowerCount(r.Context(), to),
		})
	})

	// GET /timeline/:user?limit=20
	http.HandleFunc("/timeline/", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Path[len("/timeline/"):]
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit == 0 {
			limit = 20
		}
		tweets, err := feed.Timeline(r.Context(), user, limit)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tweets)
	})

	log.Println("Twitter feed API on :8080")
	http.ListenAndServe(":8080", nil)
}
