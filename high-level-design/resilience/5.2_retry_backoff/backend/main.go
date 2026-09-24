package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

var failRate int32 // percentage 0-100

func main() {
	rateStr := os.Getenv("FAIL_RATE")
	if r, err := strconv.Atoi(rateStr); err == nil {
		atomic.StoreInt32(&failRate, int32(r))
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rate := atomic.LoadInt32(&failRate)
		if rand.Intn(100) < int(rate) {
			http.Error(w, "simulated server error", http.StatusInternalServerError)
			log.Printf("FAIL (rate=%d%%)", rate)
			return
		}
		fmt.Fprintln(w, "OK")
		log.Printf("OK (rate=%d%%)", rate)
	})

	http.HandleFunc("/admin/fail-rate", func(w http.ResponseWriter, r *http.Request) {
		rStr := r.URL.Query().Get("rate")
		if rate, err := strconv.Atoi(rStr); err == nil {
			atomic.StoreInt32(&failRate, int32(rate))
		}
		fmt.Fprintf(w, "fail rate: %d%%\n", atomic.LoadInt32(&failRate))
	})

	log.Printf("Backend on :8080 (initial fail rate: %d%%)", atomic.LoadInt32(&failRate))
	http.ListenAndServe(":8080", nil)
}
