package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

var streamer *Streamer

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	streamer = NewStreamer(NewRedisVideoStore(rdb))

	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/manifest/", handleManifest)
	http.HandleFunc("/chunk/", handleChunk)
	http.HandleFunc("/watch/", handleWatch)
	http.HandleFunc("/video/", handleVideo)
	http.HandleFunc("/feed", handleFeed)

	log.Println("video streaming server :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// POST /upload?title=T&owner=O&size=N
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	title := r.URL.Query().Get("title")
	owner := r.URL.Query().Get("owner")
	sizeStr := r.URL.Query().Get("size")
	if title == "" || owner == "" {
		http.Error(w, "title and owner required", 400)
		return
	}
	size := 10 * 1024 * 1024 // default 10 MB
	if sizeStr != "" {
		if n, err := strconv.Atoi(sizeStr); err == nil {
			size = n
		}
	}
	v, err := streamer.Upload(r.Context(), title, owner, size)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(v)
}

// GET /manifest/:videoID?res=720p
func handleManifest(w http.ResponseWriter, r *http.Request) {
	videoID := strings.TrimPrefix(r.URL.Path, "/manifest/")
	res := r.URL.Query().Get("res")
	if res == "" {
		res = "720p"
	}
	m, err := streamer.Manifest(r.Context(), videoID, res)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

// GET /chunk/:videoID/:chunkNum?res=720p
func handleChunk(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/chunk/"), "/")
	if len(parts) != 2 {
		http.Error(w, "bad path: /chunk/:videoID/:num", 400)
		return
	}
	chunkNum, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "bad chunk number", 400)
		return
	}
	res := r.URL.Query().Get("res")
	if res == "" {
		res = "720p"
	}
	chunk, err := streamer.GetChunk(r.Context(), parts[0], chunkNum, res)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chunk)
}

// POST /watch/:videoID?user=alice
func handleWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", 405)
		return
	}
	videoID := strings.TrimPrefix(r.URL.Path, "/watch/")
	userID := r.URL.Query().Get("user")
	if userID == "" {
		userID = "anonymous"
	}
	views, err := streamer.Watch(r.Context(), videoID, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"video_id": videoID, "views": views})
}

// GET /video/:videoID
func handleVideo(w http.ResponseWriter, r *http.Request) {
	videoID := strings.TrimPrefix(r.URL.Path, "/video/")
	v, err := streamer.store.GetVideo(r.Context(), videoID)
	if err != nil {
		http.Error(w, err.Error(), 404)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

// GET /feed?n=10
func handleFeed(w http.ResponseWriter, r *http.Request) {
	n := 10
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	videos, err := streamer.store.TopVideos(r.Context(), n)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"videos": videos, "count": len(videos)})
}
