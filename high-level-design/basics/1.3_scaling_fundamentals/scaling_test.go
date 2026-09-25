package main

import (
	"testing"
	"time"
)

func TestAmdahlLawDiminishingReturns(t *testing.T) {
	// Diminishing returns: the RATIO speedup(2n)/speedup(n) decreases as n grows.
	// ratio at n=1: speedup(2)/speedup(1)
	// ratio at n=8: speedup(16)/speedup(8)
	ratio1 := AmdahlSpeedup(0.9, 2) / AmdahlSpeedup(0.9, 1)
	ratio8 := AmdahlSpeedup(0.9, 16) / AmdahlSpeedup(0.9, 8)
	if ratio8 >= ratio1 {
		t.Errorf("Amdahl diminishing returns: ratio 2/1=%.3f should exceed ratio 16/8=%.3f",
			ratio1, ratio8)
	}
}

func TestAmdahlFullyParallel(t *testing.T) {
	// p=1.0: speedup = N processors
	for _, n := range []int{1, 2, 4, 8, 16} {
		got := AmdahlSpeedup(1.0, n)
		want := float64(n)
		if got < want*0.999 || got > want*1.001 {
			t.Errorf("p=1.0 n=%d: want %.1f got %.1f", n, want, got)
		}
	}
}

func TestAmdahlFullySerial(t *testing.T) {
	// p=0.0: no speedup regardless of processors
	for _, n := range []int{1, 2, 4, 8, 16} {
		got := AmdahlSpeedup(0.0, n)
		if got < 0.999 || got > 1.001 {
			t.Errorf("p=0.0 n=%d: want 1.0 got %.4f", n, got)
		}
	}
}

func TestAmdahlCeiling(t *testing.T) {
	// p=0.9: max speedup never exceeds 10x regardless of CPUs
	for _, n := range []int{16, 32, 64, 128, 1024} {
		got := AmdahlSpeedup(0.9, n)
		if got > 10.01 {
			t.Errorf("p=0.9 n=%d: speedup %.2f exceeds theoretical max 10x", n, got)
		}
	}
}

func TestVerticalCostSuperLinear(t *testing.T) {
	// Past 16 cores, cost grows faster than linear
	cost16 := VerticalCostMultiplier(16)
	cost32 := VerticalCostMultiplier(32)
	// linear would be 2x, super-linear should be >2x
	if cost32/cost16 <= 2.0 {
		t.Errorf("expected super-linear cost past 16 cores, got ratio %.2f", cost32/cost16)
	}
}

func TestStatefulHandlerTracksPerUser(t *testing.T) {
	h := NewStatefulHandler()
	for i := 1; i <= 5; i++ {
		count := h.Handle("alice")
		if count != i {
			t.Fatalf("alice call %d: want count=%d got %d", i, i, count)
		}
	}
	// bob starts at 1 independently
	if got := h.Handle("bob"); got != 1 {
		t.Fatalf("bob first call: want 1 got %d", got)
	}
}

func TestStatelessHandlerNoSharedState(t *testing.T) {
	h := &StatelessHandler{}
	// Two calls with same externalCount produce same result
	r1 := h.Handle("alice", 5)
	r2 := h.Handle("bob", 5)
	if r1 != r2 {
		t.Errorf("stateless handler should be deterministic: %d != %d", r1, r2)
	}
}

func TestWorkerThroughputScales(t *testing.T) {
	dur := 30 * time.Millisecond
	t1 := WorkerThroughput(1, dur)
	t4 := WorkerThroughput(4, dur)
	// 4 workers should produce meaningfully more throughput
	if t4 < t1*1.5 {
		t.Errorf("expected 4 workers > 1.5x throughput of 1 worker, got 1=%.0f 4=%.0f", t1, t4)
	}
}
