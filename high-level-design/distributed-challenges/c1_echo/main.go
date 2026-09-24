// Challenge 1: Echo
// Every "echo" message must be replied to with "echo_ok" and the same echo field.
// Run: ./maelstrom test -w echo --bin ./c1_echo --node-count 1 --time-limit 10
package main

import (
	"encoding/json"
	"log"

	maelstrom "example.com/maelstrom"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		body["type"] = "echo_ok"
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
