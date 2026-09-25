package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var backendURL string

func callBackend() error {
	resp, err := http.Get(backendURL + "/")
	if err != nil {
		return fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	backendURL = os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://backend:8080"
	}

	cfg := Config{
		MaxAttempts: 5,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
		Jitter:      true,
	}

	http.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		attempts, err := Do(context.Background(), cfg, callBackend)
		elapsed := time.Since(start)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"attempts":%d,"total_ms":%d,"success":%v}`,
			len(attempts), elapsed.Milliseconds(), err == nil)
		if err != nil {
			fmt.Fprintf(w, ` // final error: %v`, err)
		}
		fmt.Fprintln(w)

		for _, a := range attempts {
			status := "OK"
			if a.Err != nil {
				status = a.Err.Error()
			}
			log.Printf("  attempt %d: %s (delay_before_next: %v)", a.Number, status, a.Delay)
		}
	})

	log.Println("Retry client on :8081")
	http.ListenAndServe(":8081", nil)
}
