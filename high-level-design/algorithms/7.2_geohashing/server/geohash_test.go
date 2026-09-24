package main

import (
	"math"
	"testing"
)

func TestEncodeKnownValues(t *testing.T) {
	// San Francisco: 37.7749, -122.4194
	hash := Encode(37.7749, -122.4194, 6)
	if hash != "9q8yyk" {
		t.Errorf("SF hash: expected 9q8yyk, got %q", hash)
	}
}

func TestDecodeRoundTrip(t *testing.T) {
	lat, lng := 48.8566, 2.3522 // Paris
	hash := Encode(lat, lng, 9)
	dlat, dlng := Decode(hash)
	// Precision 9 ≈ ±2.4m
	if math.Abs(dlat-lat) > 0.001 {
		t.Errorf("lat round-trip error: got %f, want %f", dlat, lat)
	}
	if math.Abs(dlng-lng) > 0.001 {
		t.Errorf("lng round-trip error: got %f, want %f", dlng, lng)
	}
}

func TestPrecisionLength(t *testing.T) {
	for p := 1; p <= 12; p++ {
		h := Encode(0, 0, p)
		if len(h) != p {
			t.Errorf("precision %d: expected length %d, got %d", p, p, len(h))
		}
	}
}

func TestEquatorPrimeMeridian(t *testing.T) {
	hash := Encode(0, 0, 5)
	if len(hash) != 5 {
		t.Errorf("expected 5-char hash, got %q", hash)
	}
}

func TestHighPrecision(t *testing.T) {
	lat, lng := 51.5074, -0.1278 // London
	h8 := Encode(lat, lng, 8)
	h4 := Encode(lat, lng, 4)
	// Longer hash should start with shorter hash
	if h8[:4] != h4 {
		t.Errorf("h8[:4]=%q should equal h4=%q", h8[:4], h4)
	}
}
