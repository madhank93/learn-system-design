# Distributed Challenges — Fly.io Gossip Glomers

Hands-on distributed systems challenges from [fly.io/dist-sys](https://fly.io/dist-sys/), implemented in Go.

Each challenge is a standalone Go binary that communicates with the **Maelstrom** testing framework over stdin/stdout using newline-delimited JSON.

---

## Setup

### 1. Install Java 11+

Maelstrom requires a JVM.

```bash
brew install openjdk@21   # macOS
```

### 2. Download Maelstrom

```bash
curl -L https://github.com/jepsen-io/maelstrom/releases/download/v0.2.3/maelstrom.tar.bz2 | tar -xj
# Binary is at ./maelstrom/maelstrom
```

### 3. Build a challenge

```bash
cd distributed-challenges/c1_echo
go build -o c1_echo .
```

Or use the justfile recipes from the repo root:

```bash
just dc-build-all       # build all challenge binaries
just dc-c1              # build + run C1 echo test
```

### 4. Run with Maelstrom

```bash
./maelstrom/maelstrom test -w echo \
  --bin ./distributed-challenges/c1_echo/c1_echo \
  --node-count 1 --time-limit 10
```

---

## Message Protocol

All messages are JSON objects on a single line:

```json
{"src": "c1", "dest": "n1", "body": {"type": "echo", "msg_id": 1, "echo": "hello"}}
{"src": "n1", "dest": "c1", "body": {"type": "echo_ok", "msg_id": 1, "in_reply_to": 1, "echo": "hello"}}
```

Every node receives an `init` message first:

```json
{"type": "init", "msg_id": 1, "node_id": "n3", "node_ids": ["n1", "n2", "n3"]}
```

The shared `maelstrom` library handles `init` automatically.

---

## Challenges

| # | Challenge | Status | Key concept |
|---|-----------|--------|-------------|
| 1 | [Echo](c1_echo/) | ✅ Done | Basic request-reply |
| 2 | [Unique IDs](c2_unique_ids/) | ✅ Done | Coordination-free ID generation |
| 3a | [Broadcast — Single Node](c3a_broadcast_single/) | ✅ Done | In-memory message store |
| 3b | [Broadcast — Multi Node](c3b_broadcast_multi/) | ✅ Done | Peer gossip, flood-fill |
| 3c | [Broadcast — Fault Tolerant](c3c_broadcast_fault_tolerant/) | ✅ Done | RPC retry on partition |
| 3d | [Broadcast — Efficient](c3d_broadcast_efficient/) | ✅ Done | Batched periodic gossip |
| 4 | [Grow-Only Counter](c4_grow_only_counter/) | 🚧 TODO | G-Counter CRDT, seq-kv |
| 5a | [Kafka Log — Single](c5a_kafka_single/) | 🚧 TODO | Append-only log |
| 5b | [Kafka Log — Multi](c5b_kafka_multi/) | 🚧 TODO | lin-kv backed distributed log |
| 5c | [Kafka Log — Efficient](c5c_kafka_efficient/) | 🚧 TODO | Caching, reduce lin-kv RTTs |
| 6a | [Transactions — Read Uncommitted](c6a_txn_read_uncommitted/) | 🚧 TODO | Buffered writes, total availability |
| 6b | [Transactions — Read Committed](c6b_txn_read_committed/) | 🚧 TODO | Snapshot isolation |

---

## Shared Library

[`maelstrom/`](maelstrom/) — Node protocol library:

- `Node.Handle(type, fn)` — register a message handler
- `Node.Reply(req, body)` — send a response to the sender
- `Node.Send(dest, body)` — fire-and-forget to any node
- `Node.RPC(dest, body, callback)` — send and register a one-shot reply handler
- `Node.Run()` — main read loop (blocks until stdin closes)

Init is handled automatically. Handlers run in goroutines; use mutexes for shared state.

---

## Key Concepts by Challenge

### C2: Unique IDs without coordination

```
ID = "<node_id>-<atomic_counter>"
```

Node IDs are unique per cluster. Per-node counter ensures uniqueness within the node. No coordination, no lock, globally unique.

### C3b: Broadcast flood-fill

On receiving a new message: forward to all neighbors except the sender. Deduplication via `seen` set prevents infinite loops. Works but chatty — O(N) messages per broadcast.

### C3c: Fault-tolerant broadcast

Same as 3b but use RPC (tracks `msg_id → callback`). Keep retrying every second until the neighbor acknowledges. Survives network partitions.

### C3d: Efficient broadcast

Instead of N individual messages per broadcast, accumulate in a `seen` set and gossip the full set to each neighbor every 200ms. Trades per-message immediacy for batching efficiency. Passes Maelstrom's ≤30 messages/op and ≤600ms latency targets.

---

## References

- [Fly.io Gossip Glomers challenges](https://fly.io/dist-sys/)
- [Maelstrom documentation](https://github.com/jepsen-io/maelstrom/tree/main/doc)
- [Kyle Kingsbury (aphyr) workbook](https://github.com/aphyr/distsys-class)
