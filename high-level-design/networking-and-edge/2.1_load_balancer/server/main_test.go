package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func do(h http.HandlerFunc, target string, cookie string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", target, nil)
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "session_id", Value: cookie})
	}
	h(rec, req)
	return rec
}

func TestHandlerEchoesServerAndSession(t *testing.T) {
	h := newHandler("7", "us")
	body := do(h, "/", "alice").Body.String()
	want := "Server 7 (us) | Session: alice | Visit: 1"
	if !strings.Contains(body, want) {
		t.Fatalf("body = %q, want contains %q", body, want)
	}
}

func TestVisitCountIncrementsPerSession(t *testing.T) {
	h := newHandler("1", "eu")
	do(h, "/", "bob")
	body := do(h, "/", "bob").Body.String()
	if !strings.Contains(body, "Visit: 2") {
		t.Fatalf("second visit body = %q, want Visit: 2", body)
	}
}

func TestSleepParamDelaysResponse(t *testing.T) {
	h := newHandler("1", "us")
	start := time.Now()
	do(h, "/?sleep=60ms", "")
	if e := time.Since(start); e < 60*time.Millisecond {
		t.Fatalf("elapsed %v, want >= 60ms", e)
	}
}

func TestInvalidSleepIsIgnored(t *testing.T) {
	h := newHandler("1", "us")
	start := time.Now()
	do(h, "/?sleep=banana", "")
	if e := time.Since(start); e > 500*time.Millisecond {
		t.Fatalf("elapsed %v, want fast path", e)
	}
}

func TestConcurrentRequestsNoRace(t *testing.T) {
	h := newHandler("1", "us")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			do(h, "/", fmt.Sprintf("user%d", i%5))
		}(i)
	}
	wg.Wait()
}
