// Challenge 5c: Efficient Kafka-Style Log
// Optimize 5b: reduce lin-kv round trips using local caching and batching.
// Cache recent entries locally; write-through to lin-kv only when needed.
//
// Run: ./maelstrom test -w kafka --bin ./c5c_kafka_efficient --node-count 2 --time-limit 20 --rate 1000
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: build on 5b with local read cache to reduce lin-kv latency.
	// Trade-off: may return slightly stale reads (still linearizable per key).

	_ = n

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
