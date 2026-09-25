package main

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

// ── Vertical vs Horizontal ────────────────────────────────────

// VerticalScaling models the cost/capacity curve of bigger machines.
// In practice, doubling CPU costs >2x past a threshold.
func VerticalCostMultiplier(cpuCores int) float64 {
	// Rough model: linear to 16 cores, super-linear beyond
	if cpuCores <= 16 {
		return float64(cpuCores)
	}
	return 16 + math.Pow(float64(cpuCores-16), 1.5)
}

// ── Amdahl's Law ──────────────────────────────────────────────

// AmdahlSpeedup returns max theoretical speedup given fraction p of
// parallelizable work and n processors.
// Speedup = 1 / ((1-p) + p/n)
func AmdahlSpeedup(parallelFraction float64, processors int) float64 {
	serial := 1 - parallelFraction
	return 1.0 / (serial + parallelFraction/float64(processors))
}

// ── Stateless vs Stateful ─────────────────────────────────────

// StatefulHandler stores session in-process — cannot scale horizontally
// without sticky sessions or shared state.
type StatefulHandler struct {
	mu       sync.Mutex
	sessions map[string]int // userID → request count
}

func NewStatefulHandler() *StatefulHandler {
	return &StatefulHandler{sessions: make(map[string]int)}
}

func (h *StatefulHandler) Handle(userID string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[userID]++
	return h.sessions[userID]
}

// StatelessHandler carries no per-user state — any instance handles any request.
// All state goes to an external store (Redis, DB).
type StatelessHandler struct {
	// State lives in Redis/DB, not here.
	// Each request is self-contained: auth token carries identity.
}

func (h *StatelessHandler) Handle(userID string, sessionCount int) int {
	// sessionCount comes FROM the external store, not from instance memory.
	return sessionCount + 1
}

// ── Throughput benchmark ──────────────────────────────────────

// WorkerThroughput measures how many tasks/sec N workers complete.
func WorkerThroughput(workers int, duration time.Duration) float64 {
	var counter int64
	var wg sync.WaitGroup

	stop := make(chan struct{})

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					// Simulate a 1µs task (CPU-bound)
					time.Sleep(time.Microsecond)
					atomic.AddInt64(&counter, 1)
				}
			}
		}()
	}

	time.Sleep(duration)
	close(stop)
	wg.Wait()

	return float64(counter) / duration.Seconds()
}

func main() {
	fmt.Println("=== Amdahl's Law: Max Speedup with N processors ===")
	fmt.Printf("%-25s", "Parallel fraction")
	for _, n := range []int{2, 4, 8, 16, 32, 64} {
		fmt.Printf("  %3d CPUs", n)
	}
	fmt.Println()
	fmt.Println(string(make([]byte, 75)))
	for _, p := range []float64{0.50, 0.75, 0.90, 0.95, 0.99} {
		fmt.Printf("p=%-22.0f%%", p*100)
		for _, n := range []int{2, 4, 8, 16, 32, 64} {
			fmt.Printf("  %5.1fx  ", AmdahlSpeedup(p, n))
		}
		fmt.Println()
	}

	fmt.Println("\n=== Vertical Scaling Cost Multiplier ===")
	fmt.Println("(doubling cores beyond 16 costs super-linearly)")
	for _, cores := range []int{1, 2, 4, 8, 16, 32, 64, 128} {
		fmt.Printf("  %3d cores → cost multiplier %.1fx\n", cores, VerticalCostMultiplier(cores))
	}

	fmt.Println("\n=== Horizontal Scaling: Throughput vs Workers ===")
	dur := 50 * time.Millisecond
	for _, w := range []int{1, 2, 4, 8} {
		tps := WorkerThroughput(w, dur)
		fmt.Printf("  %2d workers → %.0f tasks/sec\n", w, tps)
	}
}
