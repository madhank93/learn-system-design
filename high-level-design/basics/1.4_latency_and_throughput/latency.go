package main

import (
	"fmt"
	"math"
	"time"
)

// ── Back-of-envelope estimation helpers ──────────────────────

const (
	KB = 1_000
	MB = 1_000_000
	GB = 1_000_000_000

	SecondsPerDay  = 86_400
	SecondsPerYear = 31_536_000
)

// QPS returns queries-per-second given daily active users and
// average requests per user per day.
func QPS(dauUsers, requestsPerUserPerDay int) float64 {
	return float64(dauUsers) * float64(requestsPerUserPerDay) / SecondsPerDay
}

// StorageGrowthGB returns GB/day given events/sec and bytes/event.
func StorageGrowthGB(eventsPerSec float64, bytesPerEvent int) float64 {
	return eventsPerSec * float64(bytesPerEvent) * SecondsPerDay / GB
}

// BandwidthMbps returns Mbps for a given QPS and response size.
func BandwidthMbps(qps float64, responseSizeBytes int) float64 {
	return qps * float64(responseSizeBytes) * 8 / MB
}

// ── Little's Law ──────────────────────────────────────────────

// LittlesLaw: L = λW
// L = avg items in system
// λ = arrival rate (items/sec)
// W = avg time in system (seconds)
//
// Use: given latency W and arrival rate λ, how many concurrent requests L?
// If L exceeds server capacity → queue builds → latency blows up.

type LittlesLaw struct {
	ArrivalRate float64 // λ — requests/sec
	AvgLatency  float64 // W — seconds
}

func (l LittlesLaw) ConcurrentRequests() float64 { return l.ArrivalRate * l.AvgLatency }

func (l LittlesLaw) MaxSustainableQPS(maxConcurrent float64) float64 {
	return maxConcurrent / l.AvgLatency
}

// ── P99 latency model ─────────────────────────────────────────

// PercentileLatency estimates Pxx from mean and std-dev assuming log-normal distribution.
// Log-normal is a better fit for latency than normal (latency is always positive, right-skewed).
func PercentileLatency(meanMs, stdDevMs, percentile float64) float64 {
	// Convert to log-normal parameters
	variance := stdDevMs * stdDevMs
	mu := math.Log(meanMs*meanMs / math.Sqrt(variance+meanMs*meanMs))
	sigma := math.Sqrt(math.Log(1 + variance/(meanMs*meanMs)))
	// Approximate inverse normal CDF using rational approximation
	z := normalInvCDF(percentile / 100.0)
	return math.Exp(mu + sigma*z)
}

// normalInvCDF — Beasley-Springer-Moro approximation for inverse normal CDF.
func normalInvCDF(p float64) float64 {
	a := []float64{2.515517, 0.802853, 0.010328}
	b := []float64{1.432788, 0.189269, 0.001308}
	t := math.Sqrt(-2.0 * math.Log(math.Min(p, 1-p)))
	num := a[0] + t*(a[1]+t*a[2])
	den := 1 + t*(b[0]+t*(b[1]+t*b[2]))
	sign := 1.0
	if p < 0.5 {
		sign = -1
	}
	return sign * (t - num/den)
}

// ── Throughput ceiling: Utilisation law ──────────────────────

// UtilisationCeiling: at what arrival rate does utilisation hit 100%?
// U = λ × S  where S = service time per request
// If U → 1, queue → ∞ and latency → ∞
func UtilisationCeiling(serviceTimeMs float64) float64 {
	return 1000.0 / serviceTimeMs // max QPS before U=1
}

func main() {
	fmt.Println("=== Back-of-Envelope: Twitter-scale example ===")
	dau := 200_000_000
	readsPerUser := 100
	writesPerUser := 1
	tweetSizeBytes := 280

	readQPS := QPS(dau, readsPerUser)
	writeQPS := QPS(dau, writesPerUser)
	storage := StorageGrowthGB(writeQPS, tweetSizeBytes)
	bandwidth := BandwidthMbps(readQPS, tweetSizeBytes)

	fmt.Printf("  DAU: %d\n", dau)
	fmt.Printf("  Read  QPS: %.0f\n", readQPS)
	fmt.Printf("  Write QPS: %.0f\n", writeQPS)
	fmt.Printf("  Storage growth: %.0f GB/day\n", storage)
	fmt.Printf("  Outbound bandwidth: %.0f Mbps\n", bandwidth)

	fmt.Println("\n=== Little's Law: concurrent requests at different loads ===")
	latency := 50 * time.Millisecond
	for _, qps := range []float64{100, 500, 1000, 5000, 10000} {
		ll := LittlesLaw{ArrivalRate: qps, AvgLatency: latency.Seconds()}
		fmt.Printf("  QPS=%-6.0f  P50=50ms → %.0f concurrent requests in flight\n",
			qps, ll.ConcurrentRequests())
	}

	fmt.Println("\n=== P50/P95/P99/P999 latency (mean=20ms, σ=10ms) ===")
	for _, pct := range []float64{50, 75, 90, 95, 99, 99.9} {
		est := PercentileLatency(20, 10, pct)
		fmt.Printf("  P%-5.1f → %.1f ms\n", pct, est)
	}

	fmt.Println("\n=== Utilisation ceiling (max QPS before queue blowup) ===")
	for _, svcMs := range []float64{1, 5, 10, 50, 100} {
		fmt.Printf("  service_time=%-6.0fms → max QPS ≈ %.0f\n",
			svcMs, UtilisationCeiling(svcMs))
	}
}
