// Challenge 2: Unique ID Generation
// Generate globally unique IDs without coordination.
// Strategy: node_id + monotonic counter → unique across all nodes.
// Run: ./maelstrom test -w unique-ids --bin ./c2_unique_ids --node-count 3
//
//	--availability total --nemesis partition --time-limit 30 --rate 1000
package main

import (
	"fmt"
	"log"
	"sync/atomic"

	maelstrom "example.com/maelstrom"
)

var seq uint64

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		// node_id is unique per node; seq is unique within the node.
		// Together they form a globally unique ID with no coordination.
		id := fmt.Sprintf("%s-%d", n.ID(), atomic.AddUint64(&seq, 1))
		return n.Reply(msg, map[string]any{
			"type": "generate_ok",
			"id":   id,
		})
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
