// Challenge 5b: Multi-Node Kafka-Style Log
// Same as 5a but distributed across multiple nodes.
// Use lin-kv (linearizable KV) for durable, consistent log storage.
// Each log key is stored in lin-kv; offsets are monotonically assigned via CAS.
//
// Run: ./maelstrom test -w kafka --bin ./c5b_kafka_multi --node-count 2 --time-limit 20 --rate 1000
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: implement using n.RPC("lin-kv", ...) for log storage and offset assignment.
	// Key scheme: "<log-key>-entries" → []int, "<log-key>-offset" → int (next write offset)

	_ = n

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
