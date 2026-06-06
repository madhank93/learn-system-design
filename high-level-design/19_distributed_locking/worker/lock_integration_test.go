//go:build integration

package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func startRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()
	container, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatalf("failed to start Redis container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("failed to get Redis endpoint: %v", err)
	}
	return redis.NewClient(&redis.Options{Addr: endpoint})
}

func TestIntegrationAcquireAndRelease(t *testing.T) {
	rdb := startRedis(t)
	ctx := context.Background()

	lock, err := TryAcquire(ctx, rdb, "lock:test", 10*time.Second)
	if err != nil {
		t.Fatalf("expected acquire to succeed, got: %v", err)
	}
	if err := lock.Release(ctx); err != nil {
		t.Fatalf("expected release to succeed, got: %v", err)
	}
}

func TestIntegrationDoubleLockFails(t *testing.T) {
	rdb := startRedis(t)
	ctx := context.Background()

	lock, err := TryAcquire(ctx, rdb, "lock:double", 10*time.Second)
	if err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	defer lock.Release(ctx)

	_, err2 := TryAcquire(ctx, rdb, "lock:double", 10*time.Second)
	if !errors.Is(err2, ErrLockNotAcquired) {
		t.Fatalf("expected ErrLockNotAcquired, got: %v", err2)
	}
}

func TestIntegrationReleaseByOwnerOnly(t *testing.T) {
	rdb := startRedis(t)
	ctx := context.Background()

	lock1, _ := TryAcquire(ctx, rdb, "lock:owner", 10*time.Second)

	// Forge a lock with wrong token — should not release
	faker := &Lock{rdb: rdb, key: "lock:owner", token: "wrong-token"}
	faker.Release(ctx)

	// Real owner can still release
	if err := lock1.Release(ctx); err != nil {
		t.Fatalf("owner should still be able to release: %v", err)
	}

	// After real release, lock is free
	lock2, err := TryAcquire(ctx, rdb, "lock:owner", 10*time.Second)
	if err != nil {
		t.Fatalf("lock should be free after real release: %v", err)
	}
	lock2.Release(ctx)
}

func TestIntegrationLockAutoExpires(t *testing.T) {
	rdb := startRedis(t)
	ctx := context.Background()

	// Acquire with very short TTL
	lock, err := TryAcquire(ctx, rdb, "lock:expiry", 500*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire failed: %v", err)
	}
	_ = lock // don't release — let it expire

	// Wait for TTL to expire
	time.Sleep(700 * time.Millisecond)

	// Should be acquirable again
	lock2, err := TryAcquire(ctx, rdb, "lock:expiry", 10*time.Second)
	if err != nil {
		t.Fatalf("expected lock to be free after TTL expiry: %v", err)
	}
	lock2.Release(ctx)
}

func TestIntegrationConcurrentWorkers(t *testing.T) {
	rdb := startRedis(t)
	ctx := context.Background()

	// 5 goroutines compete for the same lock — exactly 1 should win
	wins := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			lock, err := TryAcquire(ctx, rdb, "lock:concurrent", 5*time.Second)
			if err == nil {
				wins <- true
				time.Sleep(50 * time.Millisecond)
				lock.Release(ctx)
			} else {
				wins <- false
			}
		}()
	}

	results := make([]bool, 5)
	for i := 0; i < 5; i++ {
		results[i] = <-wins
	}

	acquired := 0
	for _, w := range results {
		if w {
			acquired++
		}
	}
	if acquired == 0 {
		t.Fatal("expected at least one goroutine to acquire the lock")
	}
}
