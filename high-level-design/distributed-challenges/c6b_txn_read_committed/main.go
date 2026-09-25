// Challenge 6b: Totally-Available Transactions (Read Committed)
// Same as 6a but with read-committed isolation: reads only see committed values.
// Writes within a transaction are buffered locally and only visible after commit.
//
// Run: ./maelstrom test -w txn-rw-register --bin ./c6b_txn_read_committed
//
//	--node-count 2 --time-limit 20 --rate 1000 --concurrency 2n
//	--consistency-models read-committed --availability total
package main

import (
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	// TODO: implement read-committed isolation.
	// Use a write buffer per transaction; flush to main store on commit.
	// Reads within the transaction see the write buffer + committed store.

	_ = n

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
