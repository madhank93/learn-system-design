package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	lockKey = "critical-section"
	lockTTL = 10 * time.Second
)

func newRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func doWork(workerID string) {
	log.Printf("[%s] working for 1s...", workerID)
	time.Sleep(1 * time.Second)
	log.Printf("[%s] work done", workerID)
}

func main() {
	workerID := os.Getenv("WORKER_ID")
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := newRedisClient(redisAddr)

	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		start := time.Now()

		lock, err := TryAcquire(ctx, rdb, lockKey, lockTTL)
		if err != nil {
			if errors.Is(err, ErrLockNotAcquired) {
				http.Error(w, fmt.Sprintf("[%s] lock held by another worker — try again", workerID), 409)
				return
			}
			http.Error(w, err.Error(), 500)
			return
		}

		log.Printf("[%s] acquired lock", workerID)
		defer func() {
			lock.Release(ctx)
			log.Printf("[%s] released lock", workerID)
		}()

		doWork(workerID)
		elapsed := time.Since(start)
		fmt.Fprintf(w, "[%s] completed critical section in %v\n", workerID, elapsed)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		val, err := rdb.Get(ctx, lockKey).Result()
		if errors.Is(err, redis.Nil) {
			fmt.Fprintln(w, "lock: FREE")
		} else if err != nil {
			fmt.Fprintf(w, "lock: error: %v\n", err)
		} else {
			fmt.Fprintf(w, "lock: HELD (token prefix: %.8s...)\n", val)
		}
	})

	log.Printf("Worker %s on :8080 (redis: %s)", workerID, redisAddr)
	http.ListenAndServe(":8080", nil)
}
