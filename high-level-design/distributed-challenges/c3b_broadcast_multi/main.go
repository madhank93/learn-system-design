// Challenge 3b: Multi-Node Broadcast
// Propagate broadcast messages to all peers. No fault tolerance yet.
// Strategy: on receiving a new message, forward it to all neighbors except the sender.
// Run: ./maelstrom test -w broadcast --bin ./c3b_broadcast_multi --node-count 5
//
//	--time-limit 20 --rate 10
package main

import (
	"encoding/json"
	"log"
	"strings"
	"sync"

	maelstrom "example.com/maelstrom"
)

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
		fresh := !seen[body.Message]
		seen[body.Message] = true
		nbrs := make([]string, len(neighbors))
		copy(nbrs, neighbors)
		mu.Unlock()

		if fresh {
			// Forward to all neighbors except the sender to avoid loops.
			for _, nbr := range nbrs {
				if nbr == msg.Src {
					continue
				}
				dest := nbr
				if err := n.Send(dest, map[string]any{
					"type":    "broadcast",
					"message": body.Message,
				}); err != nil {
					log.Printf("send to %s: %v", dest, err)
				}
			}
		}

		// Reply only to clients (not to peer nodes propagating gossip).
		if !strings.HasPrefix(msg.Src, "n") {
			return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
		}
		return nil
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

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
