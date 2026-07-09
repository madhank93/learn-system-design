# Learn System Design

Hands-on system design in Go. Every concept is a runnable service — not just notes.

```
high-level-design/     ← 25 HLD topics, 5 real-world designs, 12 distributed challenges
low-level-design/      ← OOP design problems
astro-site/            ← docs site (interactive sims + guided labs)
go.work                ← single workspace, all modules
```

---

## Quick Start

```bash
# Prerequisites: Go 1.25+, Docker, just
brew install just

cd high-level-design

just unit-test          # all Go unit tests, no Docker required (~23 modules)
just integration-test   # Testcontainers integration tests (Docker needed)
just <topic>-up         # start a topic's Docker Compose stack
just <topic>-test       # run its test script
just <topic>-down       # tear down
just dc-build-all       # build all Gossip Glomers challenge binaries
```

---

## High-Level Design

All topics live under `high-level-design/`. Numbered for learning order (x.y).  
Each topic = Go code + optional Docker Compose stack + README article.

### 1. Basics — Prerequisites

| # | Topic | Key concepts |
|---|-------|-------------|
| 1.1 | [Monoliths vs Microservices](high-level-design/basics/1.1_monoliths_vs_microservices/) | Decomposition, strangler fig, distributed monolith |
| 1.2 | [Networking Fundamentals](high-level-design/basics/1.2_networking_fundamentals/) | Latency numbers, TCP/TLS RTT cost, HTTP/2 vs QUIC |
| 1.3 | [Scaling Fundamentals](high-level-design/basics/1.3_scaling_fundamentals/) | Amdahl's Law, vertical vs horizontal, stateless design |
| 1.4 | [Latency and Throughput](high-level-design/basics/1.4_latency_and_throughput/) | Little's Law, back-of-envelope estimation, p99 tail latency |
| 1.5 | [Availability and SLAs](high-level-design/basics/1.5_availability_and_slas/) | Nines table, serial/parallel availability, error budget |
| 1.6 | [ACID vs BASE](high-level-design/basics/1.6_acid_vs_base/) | Isolation levels, atomic transfer, eventual consistency |

### 2. Networking & Edge

| # | Topic | Key concepts |
|---|-------|-------------|
| 2.1 | [Load Balancing](high-level-design/networking-and-edge/2.1_load_balancer/) | Round-robin, least-conn, consistent hashing, health checks |
| 2.2 | [CDN](high-level-design/networking-and-edge/2.2_cdn/) | Edge caching, cache-control, origin pull, TTL |
| 2.3 | [DNS](high-level-design/networking-and-edge/2.3_dns/) | Resolution, TTL, GeoDNS, health-based routing |
| 2.4 | [API Gateway](high-level-design/networking-and-edge/2.4_api_gateway/) | Auth, routing, versioning, rate-limit enforcement |
| 2.5 | [WebSockets / SSE](high-level-design/networking-and-edge/2.5_websockets_sse/) | Full-duplex, server-sent events, fan-out |

### 3. Scalability

| # | Topic | Key concepts |
|---|-------|-------------|
| 3.1 | [Caching](high-level-design/scalability/3.1_caching/) | Read-through, write-through, eviction, TTL, stampede |
| 3.2 | [Rate Limiting](high-level-design/scalability/3.2_rate_limiting/) | Token bucket, sliding window, per-key limits |
| 3.3 | [Database Sharding](high-level-design/scalability/3.3_db_sharding/) | Horizontal partitioning, shard key, hot spots |
| 3.4 | [Database Replication](high-level-design/scalability/3.4_db_replication/) | Primary/replica, read scaling, replication lag |

### 4. Data & Consistency

| # | Topic | Key concepts |
|---|-------|-------------|
| 4.1 | [CAP Theorem](high-level-design/data-and-consistency/4.1_cap_theorem/) | Consistency vs availability under partition |
| 4.2 | [SQL vs NoSQL](high-level-design/data-and-consistency/4.2_sql_vs_nosql/) | Schema trade-offs, query patterns, when to choose |
| 4.3 | [Consistent Hashing](high-level-design/data-and-consistency/4.3_consistent_hashing/) | Virtual nodes, minimal rehashing, load distribution |
| 4.4 | [CQRS & Event Sourcing](high-level-design/data-and-consistency/4.4_cqrs_event_sourcing/) | Read/write separation, event log as source of truth |
| 4.5 | [Leader Election](high-level-design/data-and-consistency/4.5_leader_election/) | Raft, generation numbers, split-brain prevention |

### 5. Resilience

| # | Topic | Key concepts |
|---|-------|-------------|
| 5.1 | [Circuit Breaker](high-level-design/resilience/5.1_circuit_breaker/) | Closed/open/half-open states, failure threshold |
| 5.2 | [Retry / Backoff](high-level-design/resilience/5.2_retry_backoff/) | Exponential backoff, jitter, idempotency |
| 5.3 | [Distributed Locking](high-level-design/resilience/5.3_distributed_locking/) | Redlock, fencing tokens, lock TTL |

### 6. Async & Communication

| # | Topic | Key concepts |
|---|-------|-------------|
| 6.1 | [Message Queues](high-level-design/async-and-communication/6.1_message_queues/) | At-least-once, dead-letter queue, backpressure |
| 6.2 | [Service Discovery](high-level-design/async-and-communication/6.2_service_discovery/) | Client-side vs server-side, Consul, health checks |
| 6.3 | [Distributed Tracing](high-level-design/async-and-communication/6.3_distributed_tracing/) | Spans, trace context propagation, OpenTelemetry |

### 7. Algorithms

| # | Topic | Key concepts |
|---|-------|-------------|
| 7.1 | [Bloom Filter](high-level-design/algorithms/7.1_bloom_filter/) | Probabilistic membership, false positive rate, k-hash |
| 7.2 | [Geohashing](high-level-design/algorithms/7.2_geohashing/) | Coordinate encoding, proximity search, precision |
| 7.3 | [Indexing](high-level-design/algorithms/7.3_indexing/) | B-tree, LSM-tree, write vs read trade-offs |

### 8. Security

| # | Topic | Key concepts |
|---|-------|-------------|
| 8.1 | [OAuth / JWT](high-level-design/security/8.1_oauth_jwt/) | Token issuance, signature verification, expiry, scopes |

---

## Real-World Designs

Full-stack implementations with Redis, WebSockets, and Docker Compose.  
Located at `high-level-design/real-world-designs/`.

| # | System | Stack | What it covers |
|---|--------|-------|----------------|
| SD01 | [URL Shortener](high-level-design/real-world-designs/SD01_url_shortener/) | Go + Redis | Hash generation, redirect, click tracking |
| SD02 | [Twitter Feed](high-level-design/real-world-designs/SD02_twitter_feed/) | Go + Redis | Fan-out on write, timeline, follow graph |
| SD03 | [Ride Sharing](high-level-design/real-world-designs/SD03_ride_sharing/) | Go + Redis GEO | Geospatial driver matching, trip state machine |
| SD04 | [Real-time Chat](high-level-design/real-world-designs/SD04_realtime_chat/) | Go + WebSocket + Redis pub/sub | Hub pattern, rooms, presence, message history |
| SD05 | [Video Streaming](high-level-design/real-world-designs/SD05_video_streaming/) | Go + Redis | Chunked upload, manifest, adaptive bitrate |

```bash
just url-shortener-up      just url-shortener-test      just url-shortener-down
just twitter-feed-up       just twitter-feed-test       just twitter-feed-down
just ride-sharing-up       just ride-sharing-test       just ride-sharing-down
just realtime-chat-up      just realtime-chat-test      just realtime-chat-down
just video-streaming-up    just video-streaming-test    just video-streaming-down
```

---

## Distributed Challenges

[Fly.io Gossip Glomers](https://fly.io/dist-sys/) — stdin/stdout distributed systems challenges tested by Maelstrom.  
Located at `high-level-design/distributed-challenges/`.  
Requires: Java 11+, [Maelstrom binary](https://github.com/jepsen-io/maelstrom/releases).

| # | Challenge | Status | Concept |
|---|-----------|--------|---------|
| C1 | [Echo](high-level-design/distributed-challenges/c1_echo/) | ✅ | Request-reply protocol |
| C2 | [Unique IDs](high-level-design/distributed-challenges/c2_unique_ids/) | ✅ | Coordination-free ID generation |
| C3a | [Broadcast — Single Node](high-level-design/distributed-challenges/c3a_broadcast_single/) | ✅ | In-memory message store |
| C3b | [Broadcast — Multi Node](high-level-design/distributed-challenges/c3b_broadcast_multi/) | ✅ | Flood-fill gossip, deduplication |
| C3c | [Broadcast — Fault Tolerant](high-level-design/distributed-challenges/c3c_broadcast_fault_tolerant/) | ✅ | RPC with retry on partition |
| C3d | [Broadcast — Efficient](high-level-design/distributed-challenges/c3d_broadcast_efficient/) | ✅ | Batched periodic gossip (≤30 msg/op) |
| C4 | [Grow-Only Counter](high-level-design/distributed-challenges/c4_grow_only_counter/) | 🚧 | G-Counter CRDT, seq-kv service |
| C5a | [Kafka Log — Single](high-level-design/distributed-challenges/c5a_kafka_single/) | 🚧 | Append-only log |
| C5b | [Kafka Log — Multi](high-level-design/distributed-challenges/c5b_kafka_multi/) | 🚧 | Distributed log via lin-kv |
| C5c | [Kafka Log — Efficient](high-level-design/distributed-challenges/c5c_kafka_efficient/) | 🚧 | Cached log, fewer round trips |
| C6a | [Transactions — Read Uncommitted](high-level-design/distributed-challenges/c6a_txn_read_uncommitted/) | 🚧 | Totally-available KV transactions |
| C6b | [Transactions — Read Committed](high-level-design/distributed-challenges/c6b_txn_read_committed/) | 🚧 | Snapshot isolation |

```bash
just dc-build-all   # build all challenge binaries
just dc-c1          # build C1 + print maelstrom run command
just dc-c3d         # build C3d + print maelstrom run command
```

---

## Low-Level Design

Object-oriented design problems.

| Problem | Location |
|---------|----------|
| [Parking Lot](low-level-design/parking-lot/) | Multi-level, slot types, vehicle tracking |

---

## Resources

**Books**
- [Designing Data-Intensive Applications](https://dataintensive.net/) — Kleppmann
- [Building Microservices](https://www.oreilly.com/library/view/building-microservices-2nd/9781492034018/) — Sam Newman
- [System Design Interview](https://www.amazon.com/System-Design-Interview-insiders-Second/dp/B08CMF2CQF) — Alex Xu

**References**
- [System Design Primer](https://github.com/donnemartin/system-design-primer)
- [karanpratapsingh/system-design](https://github.com/karanpratapsingh/system-design)
- [roadmap.sh/system-design](https://roadmap.sh/system-design)
- [Fly.io Gossip Glomers](https://fly.io/dist-sys/)
- [Maelstrom docs](https://github.com/jepsen-io/maelstrom/tree/main/doc)

**Videos**
- [Gaurav Sen](https://www.youtube.com/@gkcs)
- [Hussein Nasser](https://www.youtube.com/@hnasr)
- [Arpit Bhayani](https://www.youtube.com/@AsliEngineering)
- [MIT Distributed Systems](https://www.youtube.com/playlist?list=PLrw6a1wE39_tb2fErI4-WkMbsvGQk9_UB)
- [Bala G — Distributed Systems](https://www.youtube.com/@gopalbala12)
