package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const maxSleep = 10 * time.Second

// newHandler returns the echo handler. sleep=<duration> query param delays the
// response (capped at maxSleep) so labs can simulate slow backends.
func newHandler(id, region string) http.HandlerFunc {
	var mu sync.Mutex
	visits := make(map[string]int)

	return func(w http.ResponseWriter, r *http.Request) {
		if d, err := time.ParseDuration(r.URL.Query().Get("sleep")); err == nil && d > 0 {
			if d > maxSleep {
				d = maxSleep
			}
			time.Sleep(d)
		}

		sessionID := "anonymous"
		if cookie, err := r.Cookie("session_id"); err == nil {
			sessionID = cookie.Value
		}

		mu.Lock()
		visits[sessionID]++
		count := visits[sessionID]
		mu.Unlock()

		fmt.Fprintf(w, "Server %s (%s) | Session: %s | Visit: %d\n",
			id, region, sessionID, count)
		log.Printf("[INFO] Session %s → Server %s (visit #%d)", sessionID, id, count)
	}
}

func main() {
	id := os.Getenv("SERVER_ID")
	region := os.Getenv("REGION")
	if region == "" {
		region = "global"
	}

	http.HandleFunc("/", newHandler(id, region))

	log.Printf("Server %s (%s) started on :8080", id, region)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("[FATAL] Server %s failed: %v", id, err)
	}
}
