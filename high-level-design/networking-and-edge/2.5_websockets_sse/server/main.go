package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// SSE broadcaster — fans out events to all connected SSE clients.
type broker struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newBroker() *broker {
	return &broker{clients: make(map[chan string]struct{})}
}

func (b *broker) subscribe() chan string {
	ch := make(chan string, 10)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *broker) unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	b.mu.Unlock()
}

func (b *broker) broadcast(msg string) {
	b.mu.Lock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default: // drop if client is slow
		}
	}
	b.mu.Unlock()
}

var sseBroker = newBroker()

func main() {
	// WebSocket echo + broadcast
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		remoteAddr := ws.RemoteAddr().String()
		log.Printf("[WS] connected: %s", remoteAddr)
		defer func() {
			log.Printf("[WS] disconnected: %s", remoteAddr)
			ws.Close()
		}()

		for {
			var msg string
			if err := websocket.Message.Receive(ws, &msg); err != nil {
				break
			}
			log.Printf("[WS] recv from %s: %s", remoteAddr, msg)
			echo := fmt.Sprintf("echo: %s (server time: %s)", msg, time.Now().Format("15:04:05"))
			websocket.Message.Send(ws, echo)
			// Also broadcast to SSE clients
			sseBroker.broadcast(fmt.Sprintf("[WS→SSE] %s", msg))
		}
	}))

	// SSE stream — server pushes events every 2s + on WS messages
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		ch := sseBroker.subscribe()
		defer sseBroker.unsubscribe(ch)

		// Send initial connection event
		fmt.Fprintf(w, "data: connected at %s\n\n", time.Now().Format("15:04:05"))
		flusher.Flush()

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		log.Printf("[SSE] client connected: %s", r.RemoteAddr)
		for {
			select {
			case <-r.Context().Done():
				log.Printf("[SSE] client disconnected: %s", r.RemoteAddr)
				return
			case msg := <-ch:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			case t := <-ticker.C:
				fmt.Fprintf(w, "data: heartbeat %s\n\n", t.Format("15:04:05"))
				flusher.Flush()
			}
		}
	})

	// POST /publish — push a message to all SSE clients via HTTP
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		if msg == "" {
			http.Error(w, "missing ?msg=", 400)
			return
		}
		sseBroker.broadcast(msg)
		fmt.Fprintf(w, "published: %s\n", msg)
		log.Printf("[HTTP] broadcast: %s", msg)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Println("Server on :8080 — WebSocket: /ws  SSE: /events  Publish: /publish")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
