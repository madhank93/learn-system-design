package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// testMux is created once for all tests; handlers are registered against it.
var testMux *http.ServeMux

func TestMain(m *testing.M) {
	testMux = http.NewServeMux()
	setupRoutes(testMux)
	m.Run()
}

// resetState clears the package-level store and version counter.
func resetState() {
	mu.Lock()
	store = map[string]entry{}
	mu.Unlock()
	atomic.StoreInt64(&version, 0)
}

// serve is a thin helper that dispatches a request through testMux and returns
// the recorded response.
func serve(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testMux.ServeHTTP(rr, req)
	return rr
}

func TestWriteCreatesEntry(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	rr := serve(httptest.NewRequest(http.MethodPost, "/write?key=k1&val=v1", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var e entry
	if err := json.NewDecoder(rr.Body).Decode(&e); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if e.Key != "k1" {
		t.Errorf("expected key=k1, got %q", e.Key)
	}
	if e.Val != "v1" {
		t.Errorf("expected val=v1, got %q", e.Val)
	}
	if e.Version != 1 {
		t.Errorf("expected version=1, got %d", e.Version)
	}
}

func TestWriteIncrementsVersion(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	write := func(key, val string) entry {
		rr := serve(httptest.NewRequest(http.MethodPost, "/write?key="+key+"&val="+val, nil))
		var e entry
		json.NewDecoder(rr.Body).Decode(&e)
		return e
	}

	e1 := write("k1", "v1")
	e2 := write("k2", "v2")

	if e1.Version != 1 {
		t.Errorf("expected first write version=1, got %d", e1.Version)
	}
	if e2.Version != 2 {
		t.Errorf("expected second write version=2, got %d", e2.Version)
	}
}

func TestReadReturnsStoredEntry(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	serve(httptest.NewRequest(http.MethodPost, "/write?key=k1&val=v1", nil))

	rr := serve(httptest.NewRequest(http.MethodGet, "/read?key=k1", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var e entry
	if err := json.NewDecoder(rr.Body).Decode(&e); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if e.Key != "k1" {
		t.Errorf("expected key=k1, got %q", e.Key)
	}
	if e.Val != "v1" {
		t.Errorf("expected val=v1, got %q", e.Val)
	}
}

func TestReadNotFound(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	rr := serve(httptest.NewRequest(http.MethodGet, "/read?key=missing", nil))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestLogReturnsAllEntries(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	for _, k := range []string{"k1", "k2", "k3"} {
		serve(httptest.NewRequest(http.MethodPost, "/write?key="+k+"&val=v", nil))
	}

	rr := serve(httptest.NewRequest(http.MethodGet, "/log?since=0", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entries []entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestLogSinceFilters(t *testing.T) {
	resetState()
	t.Cleanup(resetState)

	for _, k := range []string{"k1", "k2", "k3"} {
		serve(httptest.NewRequest(http.MethodPost, "/write?key="+k+"&val=v", nil))
	}

	rr := serve(httptest.NewRequest(http.MethodGet, "/log?since=2", nil))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entries []entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry (version>2), got %d", len(entries))
	}
	if entries[0].Version != 3 {
		t.Errorf("expected version=3, got %d", entries[0].Version)
	}
}
