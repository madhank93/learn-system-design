package main

import (
	"strings"
	"testing"
	"time"
)

func newCB(threshold int, timeout time.Duration) *CB {
	return &CB{threshold: threshold, timeout: timeout}
}

func TestCBInitialStateClosed(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	if cb.state != stateClosed {
		t.Fatalf("expected stateClosed, got %v", cb.state)
	}
	if !cb.allow() {
		t.Fatal("expected allow() == true in CLOSED state")
	}
}

func TestCBOpensAfterThreshold(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	cb.record(false)
	cb.record(false)
	if cb.state != stateClosed {
		t.Fatal("expected still CLOSED after 2 failures")
	}
	cb.record(false)
	if cb.state != stateOpen {
		t.Fatalf("expected stateOpen after threshold failures, got %v", cb.state)
	}
}

func TestCBOpenRejectsRequests(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	cb.record(false)
	cb.record(false)
	cb.record(false)
	if cb.state != stateOpen {
		t.Fatal("CB should be OPEN")
	}
	if cb.allow() {
		t.Fatal("expected allow() == false in OPEN state")
	}
}

func TestCBResetsOnSuccess(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	cb.record(false)
	cb.record(false)
	cb.record(true)
	if cb.failures != 0 {
		t.Fatalf("expected failures == 0 after success, got %d", cb.failures)
	}
	if cb.state != stateClosed {
		t.Fatalf("expected stateClosed after success, got %v", cb.state)
	}
}

func TestCBHalfOpenAfterTimeout(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	// Force into OPEN state
	cb.record(false)
	cb.record(false)
	cb.record(false)
	if cb.state != stateOpen {
		t.Fatal("CB should be OPEN")
	}
	// Simulate timeout elapsed by backdating openedAt
	cb.openedAt = time.Now().Add(-10 * time.Second)

	allowed := cb.allow()
	if !allowed {
		t.Fatal("expected allow() == true after timeout elapses")
	}
	if cb.state != stateHalfOpen {
		t.Fatalf("expected stateHalfOpen after timeout, got %v", cb.state)
	}
}

func TestCBClosesOnHalfOpenSuccess(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	// Force into HALF_OPEN directly
	cb.state = stateHalfOpen

	cb.record(true)
	if cb.state != stateClosed {
		t.Fatalf("expected stateClosed after success in HALF_OPEN, got %v", cb.state)
	}
	if cb.failures != 0 {
		t.Fatalf("expected failures == 0, got %d", cb.failures)
	}
}

func TestCBReOpensOnHalfOpenFailure(t *testing.T) {
	cb := newCB(3, 5*time.Second)
	// Force into HALF_OPEN directly
	cb.state = stateHalfOpen

	cb.record(false)
	if cb.state != stateOpen {
		t.Fatalf("expected stateOpen after failure in HALF_OPEN, got %v", cb.state)
	}
}

func TestCBStatusStrings(t *testing.T) {
	cb := newCB(3, 5*time.Second)

	// CLOSED
	s := cb.status()
	if s != "CLOSED" {
		t.Fatalf("expected CLOSED, got %q", s)
	}

	// HALF_OPEN
	cb.state = stateHalfOpen
	s = cb.status()
	if s != "HALF_OPEN" {
		t.Fatalf("expected HALF_OPEN, got %q", s)
	}

	// OPEN — openedAt is very recent so timeout has not elapsed
	cb.state = stateOpen
	cb.openedAt = time.Now()
	s = cb.status()
	if !strings.HasPrefix(s, "OPEN(") || !strings.HasSuffix(s, "s left)") {
		t.Fatalf("expected OPEN(...s left), got %q", s)
	}
}
