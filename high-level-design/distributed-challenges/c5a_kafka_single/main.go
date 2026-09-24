// Challenge 5a: Single-Node Kafka-Style Log
// Implement an append-only log for a single node (no replication).
//
// Protocol:
//   - "send" {key: K, msg: V}       → append V to log K, reply "send_ok" {offset: N}
//   - "poll" {offsets: {K: N}}       → reply "poll_ok" {msgs: {K: [[offset, val], ...]}}
//   - "commit_offsets" {offsets: {K: N}} → persist committed offset per key, reply "commit_offsets_ok"
//   - "list_committed_offsets" {keys: [K]} → reply "list_committed_offsets_ok" {offsets: {K: N}}
//
// Run: ./maelstrom test -w kafka --bin ./c5a_kafka_single --node-count 1 --time-limit 20 --rate 1000
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: implement in-memory log map[string][]int and committed offsets map[string]int

	_ = n

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
