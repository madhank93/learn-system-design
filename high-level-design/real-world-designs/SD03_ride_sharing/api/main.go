package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var store *LocationStore

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	store = NewLocationStore(rdb)

	// POST /driver/location?id=d1&lat=37.7&lng=-122.4&status=available
	http.HandleFunc("/driver/location", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		status := r.URL.Query().Get("status")
		if status == "" {
			status = "available"
		}
		if id == "" {
			http.Error(w, "missing id", 400)
			return
		}
		ctx := r.Context()
		if err := store.UpdateLocation(ctx, id, lat, lng, status); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"id": id, "lat": lat, "lng": lng, "status": status,
		})
		log.Printf("driver %s at %.4f,%.4f status=%s", id, lat, lng, status)
	})

	// GET /riders/nearby?lat=37.7&lng=-122.4&km=5
	http.HandleFunc("/riders/nearby", func(w http.ResponseWriter, r *http.Request) {
		lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
		lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
		km, _ := strconv.ParseFloat(r.URL.Query().Get("km"), 64)
		if km == 0 {
			km = 5
		}
		ctx := r.Context()
		drivers, err := store.NearbyDrivers(ctx, lat, lng, km)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		// Annotate with ETA
		type DriverWithETA struct {
			*Driver
			ETA string `json:"eta"`
		}
		results := make([]DriverWithETA, len(drivers))
		for i, d := range drivers {
			results[i] = DriverWithETA{
				Driver: d,
				ETA:    FormatETA(ETA(d.DistKm)),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"rider_lat": lat, "rider_lng": lng,
			"radius_km": km,
			"drivers":   results,
			"count":     len(results),
		})
	})

	// GET /driver/:id/status
	http.HandleFunc("/driver/", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Path[len("/driver/"):]
		status, _ := store.DriverStatus(r.Context(), id)
		json.NewEncoder(w).Encode(map[string]string{"id": id, "status": status})
	})

	// GET /stats
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		n := store.ActiveDriverCount(r.Context())
		json.NewEncoder(w).Encode(map[string]any{"active_drivers": n})
	})

	log.Println("Ride sharing API on :8080")
	http.ListenAndServe(":8080", nil)
}
