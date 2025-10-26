package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	id := os.Getenv("SERVER_ID")
	region := os.Getenv("REGION")
	if region == "" {
		region = "global"
	}

	visits := make(map[string]int)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sessionID := "anonymous"
		if cookie, err := r.Cookie("session_id"); err == nil {
			sessionID = cookie.Value
		} else {
			sessionID = "anonymous"
		}

		visits[sessionID]++
		count := visits[sessionID]

		response := fmt.Sprintf("Server %s (%s) | Session: %s | Visit: %d\n",
			id, region, sessionID, count)

		fmt.Fprint(w, response)
		log.Printf("[INFO] Session %s → Server %s (visit #%d)", sessionID, id, count)
	})

	log.Printf("Server %s (%s) started on :8080", id, region)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Printf("[FATAL] Server %s failed: %v", id, err)
	}
}
