# High-Level Design

25 concept topics + 5 real-world designs + 12 distributed challenges.  
All Go. All runnable. Grouped by domain, numbered for learning order (x.y).

```bash
just unit-test          # all Go unit tests, no Docker
just integration-test   # Testcontainers integration tests
just test-all           # docker compose up → test → down per topic
just <topic>-up/test/down
just dc-build-all       # build all Gossip Glomers binaries
```

---

## 1. Basics

Prerequisites and architectural mental models.

| # | Topic | Key concepts |
|---|-------|-------------|
| 1.1 | [Monoliths vs Microservices](basics/1.1_monoliths_vs_microservices/) | Decomposition, trade-offs, when to split |
| 1.2 | [Networking Fundamentals](basics/1.2_networking_fundamentals/) | Latency numbers, TCP/TLS RTT cost, HTTP vs QUIC, DNS |
| 1.3 | [Scaling Fundamentals](basics/1.3_scaling_fundamentals/) | Amdahl's Law, vertical vs horizontal, stateless design |
| 1.4 | [Latency and Throughput](basics/1.4_latency_and_throughput/) | Little's Law, back-of-envelope estimation, p99 tail latency |
| 1.5 | [Availability and SLAs](basics/1.5_availability_and_slas/) | Nines table, serial/parallel availability, error budget, MTTR |
| 1.6 | [ACID vs BASE](basics/1.6_acid_vs_base/) | Isolation levels, atomic transfers, eventual consistency, sagas |

---

## 2. Networking & Edge

How traffic enters, routes, and reaches your services.

| # | Topic | Key concepts |
|---|-------|-------------|
| 2.1 | [Load Balancing](networking-and-edge/2.1_load_balancer/) | Round-robin, least-conn, consistent hashing, health checks |
| 2.2 | [CDN](networking-and-edge/2.2_cdn/) | Edge caching, cache-control, origin pull, TTL |
| 2.3 | [DNS](networking-and-edge/2.3_dns/) | Resolution, TTL, GeoDNS, health-based routing |
| 2.4 | [API Gateway](networking-and-edge/2.4_api_gateway/) | Auth, routing, versioning, rate-limit enforcement |
| 2.5 | [WebSockets / SSE](networking-and-edge/2.5_websockets_sse/) | Full-duplex, server-sent events, fan-out |

---

## 3. Scalability

Patterns for handling more load without rewriting everything.

| # | Topic | Key concepts |
|---|-------|-------------|
| 3.1 | [Caching](scalability/3.1_caching/) | Read-through, write-through, eviction, TTL, stampede |
| 3.2 | [Rate Limiting](scalability/3.2_rate_limiting/) | Token bucket, sliding window, per-key limits |
| 3.3 | [Database Sharding](scalability/3.3_db_sharding/) | Horizontal partitioning, shard key, hot spots |
| 3.4 | [Database Replication](scalability/3.4_db_replication/) | Primary/replica, read scaling, replication lag |

---

## 4. Data & Consistency

Correctness guarantees when data lives on multiple nodes.

| # | Topic | Key concepts |
|---|-------|-------------|
| 4.1 | [CAP Theorem](data-and-consistency/4.1_cap_theorem/) | Consistency vs availability under partition |
| 4.2 | [SQL vs NoSQL](data-and-consistency/4.2_sql_vs_nosql/) | Schema trade-offs, query patterns, when to pick which |
| 4.3 | [Consistent Hashing](data-and-consistency/4.3_consistent_hashing/) | Virtual nodes, minimal rehashing, load distribution |
| 4.4 | [CQRS & Event Sourcing](data-and-consistency/4.4_cqrs_event_sourcing/) | Read/write separation, event log as source of truth |
| 4.5 | [Leader Election](data-and-consistency/4.5_leader_election/) | Raft, generation numbers, split-brain prevention |

---

## 5. Resilience

Keeping systems alive when dependencies fail.

| # | Topic | Key concepts |
|---|-------|-------------|
| 5.1 | [Circuit Breaker](resilience/5.1_circuit_breaker/) | Closed/open/half-open states, failure threshold |
| 5.2 | [Retry / Backoff](resilience/5.2_retry_backoff/) | Exponential backoff, jitter, idempotency |
| 5.3 | [Distributed Locking](resilience/5.3_distributed_locking/) | Redlock, fencing tokens, lock TTL |

---

## 6. Async & Communication

Decoupling producers from consumers; observing what happens in between.

| # | Topic | Key concepts |
|---|-------|-------------|
| 6.1 | [Message Queues](async-and-communication/6.1_message_queues/) | At-least-once, dead letter queue, backpressure |
| 6.2 | [Service Discovery](async-and-communication/6.2_service_discovery/) | Client-side vs server-side, Consul, health checks |
| 6.3 | [Distributed Tracing](async-and-communication/6.3_distributed_tracing/) | Spans, trace context propagation, OpenTelemetry |

---

## 7. Algorithms

Data structures and spatial algorithms that underpin system design decisions.

| # | Topic | Key concepts |
|---|-------|-------------|
| 7.1 | [Bloom Filter](algorithms/7.1_bloom_filter/) | Probabilistic membership, false positive rate, k-hash |
| 7.2 | [Geohashing](algorithms/7.2_geohashing/) | Coordinate encoding, proximity search, precision vs size |
| 7.3 | [Indexing](algorithms/7.3_indexing/) | B-tree, LSM-tree, write vs read trade-offs |

---

## 8. Security

Authentication and authorization patterns every API needs.

| # | Topic | Key concepts |
|---|-------|-------------|
| 8.1 | [OAuth / JWT](security/8.1_oauth_jwt/) | Token issuance, signature verification, expiry, scopes |

---

## Real-World Designs

Full-stack implementations — Redis, WebSockets, Docker Compose.

| # | System | Stack | What it covers |
|---|--------|-------|----------------|
| SD01 | [URL Shortener](real-world-designs/SD01_url_shortener/) | Go + Redis | Hash generation, redirect, click tracking |
| SD02 | [Twitter Feed](real-world-designs/SD02_twitter_feed/) | Go + Redis | Fan-out on write, timeline, follow graph |
| SD03 | [Ride Sharing](real-world-designs/SD03_ride_sharing/) | Go + Redis GEO | Geospatial driver matching, trip state machine |
| SD04 | [Real-time Chat](real-world-designs/SD04_realtime_chat/) | Go + WebSocket + Redis pub/sub | Hub pattern, rooms, presence, message history |
| SD05 | [Video Streaming](real-world-designs/SD05_video_streaming/) | Go + Redis | Chunked upload, manifest, adaptive bitrate |

```bash
just url-shortener-up/test/down
just twitter-feed-up/test/down
just ride-sharing-up/test/down
just realtime-chat-up/test/down
just video-streaming-up/test/down
```

---

## Distributed Challenges

[Fly.io Gossip Glomers](https://fly.io/dist-sys/) — stdin/stdout distributed systems problems tested by Maelstrom.  
Requires: Java 11+, [Maelstrom binary](https://github.com/jepsen-io/maelstrom/releases).

| # | Challenge | Status | Concept |
|---|-----------|--------|---------|
| C1 | [Echo](distributed-challenges/c1_echo/) | ✅ | Request-reply protocol |
| C2 | [Unique IDs](distributed-challenges/c2_unique_ids/) | ✅ | Coordination-free ID generation |
| C3a | [Broadcast — Single Node](distributed-challenges/c3a_broadcast_single/) | ✅ | In-memory message store |
| C3b | [Broadcast — Multi Node](distributed-challenges/c3b_broadcast_multi/) | ✅ | Flood-fill gossip, deduplication |
| C3c | [Broadcast — Fault Tolerant](distributed-challenges/c3c_broadcast_fault_tolerant/) | ✅ | RPC retry on partition |
| C3d | [Broadcast — Efficient](distributed-challenges/c3d_broadcast_efficient/) | ✅ | Batched periodic gossip (≤30 msg/op) |
| C4 | [Grow-Only Counter](distributed-challenges/c4_grow_only_counter/) | 🚧 | G-Counter CRDT, seq-kv |
| C5a | [Kafka Log — Single](distributed-challenges/c5a_kafka_single/) | 🚧 | Append-only log |
| C5b | [Kafka Log — Multi](distributed-challenges/c5b_kafka_multi/) | 🚧 | lin-kv backed distributed log |
| C5c | [Kafka Log — Efficient](distributed-challenges/c5c_kafka_efficient/) | 🚧 | Cached log, fewer round trips |
| C6a | [Transactions — Read Uncommitted](distributed-challenges/c6a_txn_read_uncommitted/) | 🚧 | Totally-available KV transactions |
| C6b | [Transactions — Read Committed](distributed-challenges/c6b_txn_read_committed/) | 🚧 | Snapshot isolation |

```bash
just dc-build-all   # build all binaries
just dc-c1          # build C1 + print maelstrom run command
just dc-c3d         # build C3d + print maelstrom run command
```
