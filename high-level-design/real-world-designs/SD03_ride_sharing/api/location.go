package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	driverKey    = "drivers:geo"
	statusPrefix = "driver:status:"
	ttl          = 30 * time.Second // driver location expires if no heartbeat
)

type Driver struct {
	ID     string  `json:"id"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	DistKm float64 `json:"dist_km,omitempty"`
	Status string  `json:"status"`
}

type LocationStore struct {
	rdb *redis.Client
}

func NewLocationStore(rdb *redis.Client) *LocationStore {
	return &LocationStore{rdb: rdb}
}

// UpdateLocation sets or updates a driver's location.
func (s *LocationStore) UpdateLocation(ctx context.Context, id string, lat, lng float64, status string) error {
	pipe := s.rdb.Pipeline()
	pipe.GeoAdd(ctx, driverKey, &redis.GeoLocation{
		Name:      id,
		Latitude:  lat,
		Longitude: lng,
	})
	pipe.Set(ctx, statusPrefix+id, status, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// NearbyDrivers returns drivers within radius km, sorted by distance.
func (s *LocationStore) NearbyDrivers(ctx context.Context, lat, lng, radiusKm float64) ([]*Driver, error) {
	results, err := s.rdb.GeoRadius(ctx, driverKey, lng, lat, &redis.GeoRadiusQuery{
		Radius:    radiusKm,
		Unit:      "km",
		WithDist:  true,
		WithCoord: true,
		Count:     50,
		Sort:      "ASC",
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	drivers := make([]*Driver, 0, len(results))
	for _, r := range results {
		status, _ := s.rdb.Get(ctx, statusPrefix+r.Name).Result()
		if status == "" {
			status = "offline" // TTL expired
		}
		drivers = append(drivers, &Driver{
			ID:     r.Name,
			Lat:    r.Latitude,
			Lng:    r.Longitude,
			DistKm: math.Round(r.Dist*100) / 100,
			Status: status,
		})
	}
	return drivers, nil
}

// DriverStatus returns the current status of a driver.
func (s *LocationStore) DriverStatus(ctx context.Context, id string) (string, error) {
	status, err := s.rdb.Get(ctx, statusPrefix+id).Result()
	if errors.Is(err, redis.Nil) {
		return "offline", nil
	}
	return status, err
}

// HaversineKm computes distance between two lat/lng points in km.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// ETA estimates arrival time assuming avg speed 30 km/h.
func ETA(distKm float64) time.Duration {
	minutes := distKm / 30.0 * 60
	return time.Duration(minutes) * time.Minute
}

// ActiveDriverCount returns total active drivers in the geo index.
func (s *LocationStore) ActiveDriverCount(ctx context.Context) int64 {
	n, _ := s.rdb.ZCard(ctx, driverKey).Result()
	return n
}

// FormatETA formats a duration as "X min".
func FormatETA(d time.Duration) string {
	return fmt.Sprintf("%.0f min", d.Minutes())
}
