package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var authServerURL string

func validateToken(token string) (map[string]any, error) {
	req, _ := http.NewRequest("GET", authServerURL+"/validate", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid token: %s", body)
	}
	var claims map[string]any
	json.Unmarshal(body, &claims)
	return claims, nil
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"missing token"}`, 401)
			return
		}
		token := authHeader[7:]
		claims, err := validateToken(token)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, 401)
			return
		}
		log.Printf("[resource] request from sub=%v", claims["sub"])
		r.Header.Set("X-User", fmt.Sprint(claims["sub"]))
		next(w, r)
	}
}

func main() {
	authServerURL = os.Getenv("AUTH_SERVER_URL")
	if authServerURL == "" {
		authServerURL = "http://auth-server:8080"
	}

	http.HandleFunc("/profile", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-User")
		json.NewEncoder(w).Encode(map[string]string{
			"user":    user,
			"message": "access granted to protected resource",
		})
	}))

	log.Println("resource-server on :8081")
	http.ListenAndServe(":8081", nil)
}
