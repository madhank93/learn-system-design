package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

var shortener *Shortener

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	shortener = NewShortener(rdb)

	// POST /shorten?url=https://example.com
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		url := r.URL.Query().Get("url")
		if url == "" {
			http.Error(w, "missing ?url=", 400)
			return
		}
		ctx := r.Context()
		code, err := shortener.Shorten(ctx, url)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"short":    fmt.Sprintf("%s/%s", baseURL, code),
			"code":     code,
			"original": url,
		})
		log.Printf("shortened %s → %s", url, code)
	})

	// GET /:code → 302 redirect
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[1:]
		if code == "" || code == "favicon.ico" {
			http.NotFound(w, r)
			return
		}
		ctx := r.Context()
		url, err := shortener.Resolve(ctx, code)
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "short link not found", 404)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
		log.Printf("redirect %s → %s", code, url)
	})

	// GET /stats/:code
	http.HandleFunc("/stats/", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Path[len("/stats/"):]
		ctx := r.Context()
		clicks, err := shortener.Stats(ctx, code)
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", 404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"code": code, "clicks": clicks})
	})

	log.Printf("URL shortener on :8080 (base=%s)", baseURL)
	http.ListenAndServe(":8080", nil)
}
