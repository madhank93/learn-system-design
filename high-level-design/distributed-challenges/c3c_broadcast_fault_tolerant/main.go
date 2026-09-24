// Challenge 3c: Fault-Tolerant Broadcast
// Same as 3b but retry peer sends until acknowledged.
// Strategy: use RPC with a 1-second timeout. Retry in a goroutine until broadcast_ok received.
// Run: ./maelstrom test -w broadcast --bin ./c3c_broadcast_fault_tolerant --node-count 5
//
//	--time-limit 20 --rate 10 --nemesis partition
package main

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

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

	// reliableSend retries an RPC to dest until it receives broadcast_ok.
	reliableSend := func(dest string, message int) {
		for {
			done := make(chan struct{}, 1)
			if err := n.RPC(dest, map[string]any{
				"type":    "broadcast",
				"message": message,
			}, func(_ maelstrom.Message) error {
				close(done)
				return nil
			}); err != nil {
				log.Printf("rpc to %s: %v — retry in 1s", dest, err)
				time.Sleep(time.Second)
				continue
			}
			select {
			case <-done:
				return
			case <-time.After(time.Second):
				// No ack within 1s — likely partitioned. Retry.
			}
		}
	}

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
			for _, nbr := range nbrs {
				if nbr == msg.Src {
					continue
				}
				go reliableSend(nbr, body.Message)
			}
		}

		if !strings.HasPrefix(msg.Src, "n") {
			return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
		}
		// Peer nodes sending via RPC expect a reply too.
		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
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
