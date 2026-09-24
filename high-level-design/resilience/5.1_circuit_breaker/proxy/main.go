package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type cbState int

const (
	stateClosed   cbState = iota
	stateOpen
	stateHalfOpen
)

type CB struct {
	mu        sync.Mutex
	state     cbState
	failures  int
	threshold int
	timeout   time.Duration
	openedAt  time.Time
}

func (cb *CB) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Since(cb.openedAt) > cb.timeout {
			log.Println("[CB] OPEN → HALF_OPEN")
			cb.state = stateHalfOpen
			return true
		}
		return false
	case stateHalfOpen:
		return true
	}
	return false
}

func (cb *CB) record(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if success {
		if cb.state == stateHalfOpen {
			log.Println("[CB] probe succeeded → CLOSED")
		}
		cb.state = stateClosed
		cb.failures = 0
		return
	}
	cb.failures++
	if cb.state == stateHalfOpen || cb.failures >= cb.threshold {
		log.Printf("[CB] → OPEN (failures=%d)", cb.failures)
		cb.state = stateOpen
		cb.openedAt = time.Now()
	}
}

func (cb *CB) status() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return "CLOSED"
	case stateOpen:
		remaining := cb.timeout - time.Since(cb.openedAt)
		return fmt.Sprintf("OPEN(%.0fs left)", remaining.Seconds())
	case stateHalfOpen:
		return "HALF_OPEN"
	}
	return "UNKNOWN"
}

func main() {
	upstream := os.Getenv("UPSTREAM")
	threshold, _ := strconv.Atoi(os.Getenv("THRESHOLD"))
	if threshold == 0 {
		threshold = 3
	}
	timeoutSec, _ := strconv.Atoi(os.Getenv("TIMEOUT_SEC"))
	if timeoutSec == 0 {
		timeoutSec = 5
	}

	cb := &CB{threshold: threshold, timeout: time.Duration(timeoutSec) * time.Second}
	log.Printf("Circuit breaker proxy → %s (threshold=%d, timeout=%ds)", upstream, threshold, timeoutSec)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !cb.allow() {
			w.Header().Set("X-Circuit", cb.status())
			http.Error(w, "circuit open — fast fail", http.StatusServiceUnavailable)
			return
		}
		resp, err := http.Get(upstream + r.RequestURI)
		if err != nil || resp.StatusCode >= 500 {
			cb.record(false)
			code := http.StatusBadGateway
			if resp != nil {
				code = resp.StatusCode
			}
			w.Header().Set("X-Circuit", cb.status())
			http.Error(w, "upstream error", code)
			return
		}
		cb.record(true)
		defer resp.Body.Close()
		w.Header().Set("X-Circuit", cb.status())
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	http.HandleFunc("/cb/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, cb.status())
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("[FATAL] proxy: %v", err)
	}
}
