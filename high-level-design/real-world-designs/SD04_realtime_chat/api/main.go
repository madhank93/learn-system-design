package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/websocket"
)

var hub *Hub

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	hub = NewHub(NewRedisStore(rdb))

	http.HandleFunc("/rooms", handleRooms)
	http.HandleFunc("/history/", handleHistory)
	http.HandleFunc("/presence/", handlePresence)
	http.Handle("/ws", websocket.Handler(handleWS))

	log.Println("chat server :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// GET  /rooms         → list active rooms
// POST /rooms?name=X  → create / join room (returns room info)
func handleRooms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodPost {
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, `{"error":"name required"}`, 400)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"room": name, "status": "ready"})
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"rooms": hub.ListRooms()})
}

// GET /history/:room?n=50
func handleHistory(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/history/")
	n := 50
	if v := r.URL.Query().Get("n"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			n = parsed
		}
	}
	msgs, err := hub.store.History(r.Context(), room, n)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"room": room, "messages": msgs})
}

// GET /presence/:room
func handlePresence(w http.ResponseWriter, r *http.Request) {
	room := strings.TrimPrefix(r.URL.Path, "/presence/")
	users, err := hub.store.Presence(r.Context(), room)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"room":  room,
		"users": users,
		"count": len(users),
	})
}

// WS /ws?room=general&user=alice
func handleWS(ws *websocket.Conn) {
	room := ws.Request().URL.Query().Get("room")
	user := ws.Request().URL.Query().Get("user")
	if room == "" || user == "" {
		ws.Close()
		return
	}

	log.Printf("[WS] %s joined %s", user, room)
	c := hub.Join(room, user, ws)
	defer func() {
		hub.Leave(room, user)
		log.Printf("[WS] %s left %s", user, room)
	}()

	// announce join
	joinMsg := Message{
		ID:   newMsgID(),
		Room: room,
		User: "system",
		Text: user + " joined",
		TS:   timeNow(),
	}
	hub.store.AddMessage(ws.Request().Context(), room, joinMsg)
	hub.Broadcast(joinMsg)

	_ = c // keep reference; pump goroutine is started inside Join

	// read loop
	for {
		var raw string
		if err := websocket.Message.Receive(ws, &raw); err != nil {
			break
		}
		msg := Message{
			ID:   newMsgID(),
			Room: room,
			User: user,
			Text: raw,
			TS:   timeNow(),
		}
		_ = hub.store.AddMessage(ws.Request().Context(), room, msg)
		hub.Broadcast(msg)
	}

	// announce leave
	leaveMsg := Message{
		ID:   newMsgID(),
		Room: room,
		User: "system",
		Text: user + " left",
		TS:   timeNow(),
	}
	hub.store.AddMessage(ws.Request().Context(), room, leaveMsg)
	hub.Broadcast(leaveMsg)
}

func timeNow() int64 { return time.Now().Unix() }
