// Challenge 3a: Single-Node Broadcast
// Store broadcast messages locally. No propagation needed (one node).
// Run: ./maelstrom test -w broadcast --bin ./c3a_broadcast_single --node-count 1 --time-limit 20 --rate 10
package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	var mu sync.Mutex
	var messages []int

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body struct {
			Message int `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		mu.Lock()
		messages = append(messages, body.Message)
		mu.Unlock()
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		mu.Lock()
		out := make([]int, len(messages))
		copy(out, messages)
		mu.Unlock()
		return n.Reply(msg, map[string]any{
			"type":     "read_ok",
			"messages": out,
		})
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
