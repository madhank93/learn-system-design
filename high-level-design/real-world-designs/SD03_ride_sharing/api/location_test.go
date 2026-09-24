package main

import (
	"math"
	"testing"
	"time"
)

func TestHaversineKnownDistance(t *testing.T) {
	// SF to Oakland ≈ 13 km
	d := HaversineKm(37.7749, -122.4194, 37.8044, -122.2712)
	if math.Abs(d-13.0) > 1.0 {
		t.Errorf("SF→Oakland: expected ~13km, got %.2f", d)
	}
}

func TestHaversineSamePoint(t *testing.T) {
	d := HaversineKm(48.8566, 2.3522, 48.8566, 2.3522)
	if d > 0.001 {
		t.Errorf("same point: expected 0, got %.6f", d)
	}
}

func TestETACalculation(t *testing.T) {
	// 15 km at 30 km/h = 30 minutes
	eta := ETA(15.0)
	if math.Abs(eta.Minutes()-30.0) > 0.1 {
		t.Errorf("15km ETA: expected 30min, got %.1fmin", eta.Minutes())
	}
}

func TestETAZeroDistance(t *testing.T) {
	eta := ETA(0)
	if eta != 0 {
		t.Errorf("0km ETA: expected 0, got %v", eta)
	}
}

func TestFormatETA(t *testing.T) {
	s := FormatETA(30 * time.Minute)
	if s != "30 min" {
		t.Errorf("expected '30 min', got %q", s)
	}
}

func TestHaversineSymmetry(t *testing.T) {
	d1 := HaversineKm(51.5074, -0.1278, 48.8566, 2.3522) // London to Paris
	d2 := HaversineKm(48.8566, 2.3522, 51.5074, -0.1278) // Paris to London
	if math.Abs(d1-d2) > 0.001 {
		t.Errorf("haversine not symmetric: d1=%.4f d2=%.4f", d1, d2)
	}
}
