// Package maelstrom implements the Maelstrom node protocol for Gossip Glomers challenges.
// Messages are newline-delimited JSON on stdin/stdout.
package maelstrom

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

// Message is the top-level Maelstrom envelope.
type Message struct {
	Src  string          `json:"src"`
	Dest string          `json:"dest"`
	Body json.RawMessage `json:"body"`
}

// MessageBody contains the common fields every message body carries.
type MessageBody struct {
	Type      string `json:"type"`
	MsgID     int    `json:"msg_id,omitempty"`
	InReplyTo int    `json:"in_reply_to,omitempty"`
}

// HandlerFunc handles an incoming message.
type HandlerFunc func(msg Message) error

// Node is a Maelstrom node. Handlers are called in separate goroutines.
type Node struct {
	mu        sync.Mutex
	id        string
	nodeIDs   []string
	handlers  map[string]HandlerFunc
	callbacks map[int]HandlerFunc // keyed by msg_id of the sent RPC
	msgSeq    uint64
	out       *json.Encoder
}

// NewNode creates a Node ready to use. Call Handle to register message types, then Run.
func NewNode() *Node {
	return &Node{
		handlers:  make(map[string]HandlerFunc),
		callbacks: make(map[int]HandlerFunc),
		out:       json.NewEncoder(os.Stdout),
	}
}

// ID returns this node's ID (available after init is received).
func (n *Node) ID() string {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.id
}

// NodeIDs returns all node IDs in the cluster (available after init).
func (n *Node) NodeIDs() []string {
	n.mu.Lock()
	defer n.mu.Unlock()
	ids := make([]string, len(n.nodeIDs))
	copy(ids, n.nodeIDs)
	return ids
}

// Handle registers a handler for messages of the given type.
func (n *Node) Handle(msgType string, fn HandlerFunc) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.handlers[msgType] = fn
}

// Reply sends a response to the source of req. Sets in_reply_to automatically.
func (n *Node) Reply(req Message, body any) error {
	var reqBody MessageBody
	_ = json.Unmarshal(req.Body, &reqBody)
	m := toMap(body)
	if reqBody.MsgID != 0 {
		m["in_reply_to"] = reqBody.MsgID
	}
	return n.writeMsg(req.Src, m)
}

// Send sends a message to dest with no reply tracking.
func (n *Node) Send(dest string, body any) error {
	return n.writeMsg(dest, toMap(body))
}

// RPC sends a message to dest and registers handler for the reply.
// handler is called once when the reply arrives.
func (n *Node) RPC(dest string, body any, handler HandlerFunc) error {
	id := int(atomic.AddUint64(&n.msgSeq, 1))
	n.mu.Lock()
	n.callbacks[id] = handler
	n.mu.Unlock()
	m := toMap(body)
	m["msg_id"] = id
	return n.writeMsg(dest, m)
}

// Run reads messages from stdin until EOF. Blocks.
func (n *Node) Run() error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		var msg Message
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			log.Printf("maelstrom: unmarshal: %v", err)
			continue
		}
		var body MessageBody
		_ = json.Unmarshal(msg.Body, &body)

		if body.Type == "init" {
			// Handled synchronously — must complete before any other message is processed.
			if err := n.handleInit(msg); err != nil {
				return err
			}
			continue
		}

		if body.InReplyTo != 0 {
			n.mu.Lock()
			cb, ok := n.callbacks[body.InReplyTo]
			if ok {
				delete(n.callbacks, body.InReplyTo)
			}
			n.mu.Unlock()
			if ok {
				go func(h HandlerFunc, m Message) {
					if err := h(m); err != nil {
						log.Printf("maelstrom: rpc callback: %v", err)
					}
				}(cb, msg)
			}
			continue
		}

		n.mu.Lock()
		handler, ok := n.handlers[body.Type]
		n.mu.Unlock()

		if !ok {
			log.Printf("maelstrom: no handler for %q", body.Type)
			continue
		}

		go func(h HandlerFunc, m Message, t string) {
			if err := h(m); err != nil {
				log.Printf("maelstrom: handler %q: %v", t, err)
			}
		}(handler, msg, body.Type)
	}
	return scanner.Err()
}

func (n *Node) handleInit(msg Message) error {
	var body struct {
		NodeID  string   `json:"node_id"`
		NodeIDs []string `json:"node_ids"`
	}
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}
	n.mu.Lock()
	n.id = body.NodeID
	n.nodeIDs = body.NodeIDs
	n.mu.Unlock()
	return n.Reply(msg, map[string]any{"type": "init_ok"})
}

func (n *Node) writeMsg(dest string, body map[string]any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.out.Encode(Message{
		Src:  n.id,
		Dest: dest,
		Body: json.RawMessage(b),
	})
}

// toMap converts any value to map[string]any via JSON round-trip.
func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}
