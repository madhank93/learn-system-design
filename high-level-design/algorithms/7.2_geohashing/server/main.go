package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb = redis.NewClient(&redis.Options{Addr: redisAddr})

	// GET /encode?lat=37.7&lng=-122.4&precision=6 → geohash string
	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		prec, _ := strconv.Atoi(r.URL.Query().Get("precision"))
		if prec == 0 {
			prec = 6
		}
		hash := Encode(lat, lng, prec)
		json.NewEncoder(w).Encode(map[string]any{
			"lat": lat, "lng": lng, "precision": prec, "geohash": hash,
		})
	})

	// GET /decode?hash=9q8yy9 → lat/lng center
	http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
		hash := r.URL.Query().Get("hash")
		lat, lng := Decode(hash)
		json.NewEncoder(w).Encode(map[string]any{
			"geohash": hash, "lat": lat, "lng": lng,
		})
	})

	// POST /location?id=driver1&lat=37.7&lng=-122.4 — add to Redis GEO index
	http.HandleFunc("/location", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		ctx := context.Background()
		rdb.GeoAdd(ctx, "locations", &redis.GeoLocation{
			Name: id, Longitude: lng, Latitude: lat,
		})
		hash := Encode(lat, lng, 6)
		fmt.Fprintf(w, `{"id":%q,"geohash":%q}`+"\n", id, hash)
		log.Printf("added %s at %.4f,%.4f hash=%s", id, lat, lng, hash)
	})

	// GET /nearby?lat=37.7&lng=-122.4&km=5 — find nearby from Redis GEO
	http.HandleFunc("/nearby", func(w http.ResponseWriter, r *http.Request) {
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		km, _ := strconv.ParseFloat(r.URL.Query().Get("km"), 64)
		if km == 0 {
			km = 5
		}
		ctx := context.Background()
		results, err := rdb.GeoRadius(ctx, "locations", lng, lat, &redis.GeoRadiusQuery{
			Radius:    km,
			Unit:      "km",
			WithDist:  true,
			WithCoord: true,
			Count:     20,
			Sort:      "ASC",
		}).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	})

	log.Println("Geohash server on :8080")
	http.ListenAndServe(":8080", nil)
}
