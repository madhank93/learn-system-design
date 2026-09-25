// Challenge 6a: Totally-Available Transactions (Read Uncommitted)
// Implement a KV store with multi-key transactions.
// Availability is total: every request must succeed, even under partition.
// Read-uncommitted: reads may return values written by in-progress transactions.
//
// Protocol:
//   - "txn" {txn: [[op, key, val], ...]} → apply ops, reply "txn_ok" {txn: [[op, key, result], ...]}
//   - op "r": read key, return current value (or null)
//   - op "w": write key=val, return val
//
// Run: ./maelstrom test -w txn-rw-register --bin ./c6a_txn_read_uncommitted
//
//	--node-count 2 --time-limit 20 --rate 1000 --concurrency 2n
//	--consistency-models read-uncommitted --availability total
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: implement using in-memory map[int]int store per node.
	// Transactions apply ops sequentially; no cross-node coordination needed for read-uncommitted.

	_ = n

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
