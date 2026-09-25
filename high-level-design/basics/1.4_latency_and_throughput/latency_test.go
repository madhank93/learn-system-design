package main

import (
	"math"
	"testing"
)

func TestQPSScale(t *testing.T) {
	// 1M DAU, 10 req/day → ~115 QPS
	got := QPS(1_000_000, 10)
	want := 115.7
	if math.Abs(got-want) > 1 {
		t.Errorf("QPS(1M, 10) = %.1f, want ~%.1f", got, want)
	}
}

func TestQPSHighVolume(t *testing.T) {
	// 100M DAU, 100 req/day → ~115k QPS
	got := QPS(100_000_000, 100)
	if got < 100_000 || got > 130_000 {
		t.Errorf("QPS(100M, 100) = %.0f, expected ~115k", got)
	}
}

func TestStorageGrowth(t *testing.T) {
	// 1000 events/sec, 1KB each → 86.4 GB/day
	got := StorageGrowthGB(1000, KB)
	want := 86.4
	if math.Abs(got-want) > 1 {
		t.Errorf("StorageGrowthGB(1000, 1KB) = %.1f GB/day, want %.1f", got, want)
	}
}

func TestBandwidthMbps(t *testing.T) {
	// 1000 QPS, 1KB response → 8 Mbps
	got := BandwidthMbps(1000, KB)
	want := 8.0
	if math.Abs(got-want) > 0.1 {
		t.Errorf("BandwidthMbps(1000, 1KB) = %.2f, want %.2f", got, want)
	}
}

func TestLittlesLawConcurrency(t *testing.T) {
	// 100 req/s, 0.1s latency → 10 concurrent
	ll := LittlesLaw{ArrivalRate: 100, AvgLatency: 0.1}
	got := ll.ConcurrentRequests()
	if math.Abs(got-10) > 0.001 {
		t.Errorf("L=λW: want 10, got %.2f", got)
	}
}

func TestLittlesLawMaxQPS(t *testing.T) {
	// max 50 concurrent, 0.1s latency → max 500 QPS
	ll := LittlesLaw{ArrivalRate: 0, AvgLatency: 0.1}
	got := ll.MaxSustainableQPS(50)
	if math.Abs(got-500) > 0.001 {
		t.Errorf("MaxQPS: want 500, got %.2f", got)
	}
}

func TestUtilisationCeiling(t *testing.T) {
	// 10ms service time → 100 QPS ceiling
	got := UtilisationCeiling(10)
	if math.Abs(got-100) > 0.001 {
		t.Errorf("UtilisationCeiling(10ms) = %.1f, want 100", got)
	}

	// 1ms service time → 1000 QPS ceiling
	got = UtilisationCeiling(1)
	if math.Abs(got-1000) > 0.001 {
		t.Errorf("UtilisationCeiling(1ms) = %.1f, want 1000", got)
	}
}

func TestPercentileOrdering(t *testing.T) {
	// P99 > P95 > P50 for any distribution
	p50 := PercentileLatency(20, 10, 50)
	p95 := PercentileLatency(20, 10, 95)
	p99 := PercentileLatency(20, 10, 99)

	if p50 >= p95 {
		t.Errorf("P50 (%.1f) should be < P95 (%.1f)", p50, p95)
	}
	if p95 >= p99 {
		t.Errorf("P95 (%.1f) should be < P99 (%.1f)", p95, p99)
	}
}

func TestP50NearMean(t *testing.T) {
	// For log-normal, P50 ≈ median, which is close to mean for low variance
	p50 := PercentileLatency(20, 5, 50)
	if p50 < 15 || p50 > 25 {
		t.Errorf("P50 (%.1f) should be near mean 20ms", p50)
	}
}
