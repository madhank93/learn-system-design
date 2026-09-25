// Challenge 3d: Efficient Broadcast (Part I)
// Reduce message overhead: batch gossip every 200ms instead of forwarding each message immediately.
// Goal: messages-per-op ≤ 30, median latency ≤ 400ms, max latency ≤ 600ms.
//
// Strategy:
//   - Clients send "broadcast" messages → store locally, reply immediately.
//   - Background goroutine sends all seen messages to each neighbor every 200ms.
//   - Peers receive "gossip" with a batch and merge into their seen set.
//
// Run: ./maelstrom test -w broadcast --bin ./c3d_broadcast_efficient --node-count 25
//
//	--time-limit 20 --rate 100 --latency 100
package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	maelstrom "example.com/maelstrom"
)

const gossipInterval = 200 * time.Millisecond

func main() {
	n := maelstrom.NewNode()

	var mu sync.Mutex
	seen := map[int]bool{}
	var neighbors []string

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		mu.Lock()
		neighbors = body.Topology[n.ID()]
		mu.Unlock()
		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body struct {
			Message int `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		mu.Lock()
		seen[body.Message] = true
		mu.Unlock()
		// Only clients (not peers) expect a broadcast_ok reply.
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	// gossip is a batched peer message carrying multiple values at once.
	n.Handle("gossip", func(msg maelstrom.Message) error {
		var body struct {
			Messages []int `json:"messages"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		mu.Lock()
		for _, m := range body.Messages {
			seen[m] = true
		}
		mu.Unlock()
		return nil // no reply needed for gossip
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		mu.Lock()
		msgs := make([]int, 0, len(seen))
		for m := range seen {
			msgs = append(msgs, m)
		}
		mu.Unlock()
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": msgs,
		})
	})

	// Background gossip: every 200ms, send all seen messages to each neighbor.
	// This amortizes per-message network overhead into periodic batches.
	go func() {
		ticker := time.NewTicker(gossipInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			msgs := make([]int, 0, len(seen))
			for m := range seen {
				msgs = append(msgs, m)
			}
			nbrs := make([]string, len(neighbors))
			copy(nbrs, neighbors)
			mu.Unlock()

			if len(msgs) == 0 || len(nbrs) == 0 {
				continue
			}
			for _, nbr := range nbrs {
				if err := n.Send(nbr, map[string]any{
					"type":     "gossip",
					"messages": msgs,
				}); err != nil {
					log.Printf("gossip to %s: %v", nbr, err)
				}
			}
		}
	}()

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
