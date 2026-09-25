// Challenge 4: Grow-Only Counter (G-Counter CRDT)
// Implement a distributed counter that only increments.
// Uses the Maelstrom seq-kv service for durable storage.
//
// Protocol:
//   - "add" {delta: N} → increment counter by N, reply "add_ok"
//   - "read" → reply "read_ok" {value: total}
//
// Strategy: each node keeps its own counter in seq-kv under key "<node_id>".
// To read the global total, sum all per-node counters.
// CAS loop ensures increments are atomic even under concurrent writers.
//
// Run: ./maelstrom test -w g-counter --bin ./c4_grow_only_counter --node-count 3
//
//	--time-limit 20 --rate 100 --nemesis partition
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: implement using n.RPC("seq-kv", ...) to:
	//   read  → {"type":"read", "key": nodeID}
	//   write → {"type":"write", "key": nodeID, "value": newVal}
	//   cas   → {"type":"cas", "key": nodeID, "from": old, "to": new, "create_if_not_exists": true}

	_ = n // suppress unused warning during development

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
