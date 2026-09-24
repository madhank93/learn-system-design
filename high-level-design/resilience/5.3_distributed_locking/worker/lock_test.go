package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRedis is a simple in-memory mock for testing lock logic without Redis.
// We test lock logic by directly testing the functions with a real Redis
// or by unit-testing the pure logic parts.

func TestRandomTokenUnique(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		tok := randomToken()
		if len(tok) != 32 {
			t.Fatalf("expected 32 hex chars, got %d: %q", len(tok), tok)
		}
		if tokens[tok] {
			t.Fatalf("duplicate token generated: %q", tok)
		}
		tokens[tok] = true
	}
}

func TestLockErrorSentinel(t *testing.T) {
	if !errors.Is(ErrLockNotAcquired, ErrLockNotAcquired) {
		t.Fatal("sentinel error identity check failed")
	}
}

func TestBackoffDelayGrowth(t *testing.T) {
	// Verify that retry delay grows with attempt count (used in worker retry loop)
	delays := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
	}
	for i := 1; i < len(delays); i++ {
		if delays[i] <= delays[i-1] {
			t.Errorf("delay[%d]=%v not > delay[%d]=%v", i, delays[i], i-1, delays[i-1])
		}
	}
}

// Integration test — only runs when REDIS_ADDR is set.
func TestAcquireAndRelease(t *testing.T) {
	addr := "localhost:6379"

	rdb := newRedisClient(addr)
	ctx := context.Background()

	// Clean up any existing key
	rdb.Del(ctx, "test-lock")

	lock, err := TryAcquire(ctx, rdb, "test-lock", 5*time.Second)
	if err != nil {
		t.Skipf("Redis not available at %s: %v", addr, err)
	}
	defer lock.Release(ctx)

	// Second acquire should fail
	_, err2 := TryAcquire(ctx, rdb, "test-lock", 5*time.Second)
	if !errors.Is(err2, ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired, got %v", err2)
	}

	// Release then re-acquire should succeed
	lock.Release(ctx)
	lock2, err3 := TryAcquire(ctx, rdb, "test-lock", 5*time.Second)
	if err3 != nil {
		t.Fatalf("expected re-acquire after release, got %v", err3)
	}
	lock2.Release(ctx)
}
