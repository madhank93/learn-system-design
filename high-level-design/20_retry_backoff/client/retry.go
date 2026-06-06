package main

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Config configures the retry policy.
type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
	Jitter      bool
}

// DefaultConfig is a sensible production default.
var DefaultConfig = Config{
	MaxAttempts: 5,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    10 * time.Second,
	Multiplier:  2.0,
	Jitter:      true,
}

// Attempt represents one retry attempt.
type Attempt struct {
	Number int
	Delay  time.Duration
	Err    error
}

// Do retries fn according to cfg. Returns the last error if all attempts fail.
// fn should return nil on success or an error on failure.
func Do(ctx context.Context, cfg Config, fn func() error) ([]Attempt, error) {
	var attempts []Attempt
	for i := 1; i <= cfg.MaxAttempts; i++ {
		err := fn()
		delay := backoffDelay(i, cfg)
		attempts = append(attempts, Attempt{Number: i, Delay: delay, Err: err})
		if err == nil {
			return attempts, nil
		}
		if i == cfg.MaxAttempts {
			break
		}
		// Wait or context cancel
		select {
		case <-ctx.Done():
			return attempts, ctx.Err()
		case <-time.After(delay):
		}
	}
	return attempts, attempts[len(attempts)-1].Err
}

// backoffDelay computes delay for attempt i using exponential backoff with optional jitter.
func backoffDelay(attempt int, cfg Config) time.Duration {
	exp := math.Pow(cfg.Multiplier, float64(attempt-1))
	delay := time.Duration(float64(cfg.BaseDelay) * exp)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	if cfg.Jitter && delay > 0 {
		// Full jitter: random in [0, delay]
		delay = time.Duration(rand.Int63n(int64(delay) + 1))
	}
	return delay
}
