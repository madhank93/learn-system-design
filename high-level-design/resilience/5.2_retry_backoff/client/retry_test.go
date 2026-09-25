package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errTransient = errors.New("transient error")

func TestRetrySucceedsOnFirstAttempt(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 10 * time.Millisecond, Multiplier: 2.0}
	attempts, err := Do(context.Background(), cfg, func() error { return nil })
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(attempts) != 1 {
		t.Fatalf("expected 1 attempt, got %d", len(attempts))
	}
}

func TestRetrySucceedsOnSecondAttempt(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond, Multiplier: 2.0}
	calls := 0
	attempts, err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 2 {
			return errTransient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if len(attempts) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(attempts))
	}
}

func TestRetryExhausted(t *testing.T) {
	cfg := Config{MaxAttempts: 3, BaseDelay: 1 * time.Millisecond, Multiplier: 2.0}
	attempts, err := Do(context.Background(), cfg, func() error { return errTransient })
	if err == nil {
		t.Fatal("expected error after all attempts exhausted")
	}
	if len(attempts) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(attempts))
	}
}

func TestBackoffDelayIncreases(t *testing.T) {
	cfg := Config{BaseDelay: 100 * time.Millisecond, MaxDelay: 10 * time.Second, Multiplier: 2.0}
	d1 := backoffDelay(1, cfg)
	d2 := backoffDelay(2, cfg)
	d3 := backoffDelay(3, cfg)
	if !(d1 <= d2 && d2 <= d3) {
		t.Errorf("delays should be non-decreasing: %v %v %v", d1, d2, d3)
	}
}

func TestBackoffDelayMaxCapped(t *testing.T) {
	cfg := Config{BaseDelay: 1 * time.Second, MaxDelay: 3 * time.Second, Multiplier: 2.0}
	for i := 1; i <= 10; i++ {
		d := backoffDelay(i, cfg)
		if d > cfg.MaxDelay {
			t.Errorf("attempt %d: delay %v exceeded max %v", i, d, cfg.MaxDelay)
		}
	}
}

func TestRetryContextCancelled(t *testing.T) {
	cfg := Config{MaxAttempts: 10, BaseDelay: 100 * time.Millisecond, Multiplier: 2.0}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := Do(ctx, cfg, func() error { return errTransient })
	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}
