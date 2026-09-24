package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var secret []byte

// In-memory auth code store (production: Redis with TTL)
var authCodes = map[string]string{} // code → username

func main() {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		sec = "dev-secret-change-me"
	}
	secret = []byte(sec)

	// Step 1: /authorize?user=alice&client_id=app1 → auth code
	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			http.Error(w, "missing user", 400)
			return
		}
		code := fmt.Sprintf("code-%s-%d", user, len(authCodes))
		authCodes[code] = user
		json.NewEncoder(w).Encode(map[string]string{"code": code, "user": user})
		log.Printf("authorized user=%s code=%s", user, code)
	})

	// Step 2: /token?code=... → JWT access token
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		user, ok := authCodes[code]
		if !ok {
			http.Error(w, `{"error":"invalid_code"}`, 401)
			return
		}
		delete(authCodes, code) // one-time use
		token, err := IssueToken(secret, user, "user", 30*time.Minute)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": token,
			"token_type":   "Bearer",
			"expires_in":   "1800",
		})
		log.Printf("issued token for user=%s", user)
	})

	// /validate — resource servers call this to verify a token
	http.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
		claims, err := ValidateToken(secret, token)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 401)
			return
		}
		json.NewEncoder(w).Encode(claims)
	})

	log.Println("auth-server on :8080")
	http.ListenAndServe(":8080", nil)
}
