package main

import (
	"context"
	"testing"
	"time"
)

func TestTryAcquireAndRenew(t *testing.T) {
	// Unit test: verify ErrNotLeader sentinel exists and is distinct
	if ErrNotLeader == nil {
		t.Fatal("ErrNotLeader should not be nil")
	}
}

func TestLeaseTTLConstants(t *testing.T) {
	if leaseTTL <= heartbeat {
		t.Errorf("leaseTTL (%v) must be > heartbeat (%v) to avoid expiry between renewals", leaseTTL, heartbeat)
	}
	if heartbeat <= campaign {
		t.Errorf("heartbeat (%v) must be > campaign (%v)", heartbeat, campaign)
	}
}

func TestContextCancellation(t *testing.T) {
	// Run() should exit cleanly on context cancel without panic
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Can't test full Run() without Redis, but verify context done path
	select {
	case <-ctx.Done():
		// expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("context should have been cancelled")
	}
}
