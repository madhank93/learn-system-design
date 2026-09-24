# Latency and Throughput

**Latency and throughput are the two axes on which every system's performance is measured, and conflating them is one of the most common mistakes in system design.** Optimizing for average latency while ignoring p99 tail latency leaves 1% of users with a miserable experience — at 10,000 RPS that is 100 broken users every second. Little's Law reveals the surprising truth that concurrency, arrival rate, and latency are locked together in a single equation, which directly determines how large your thread pools, connection pools, and worker queues must be. Back-of-envelope estimation turns abstract scale numbers ("1 million users") into concrete capacity figures that reveal whether a design will work before a single line of code is written.

---

## Core Definitions

| Term | Definition | Unit | Example |
|---|---|---|---|
| **Latency** | Time for a single request to complete end-to-end | ms, µs, ns | API call takes 45ms |
| **Throughput** | Requests (or bytes) completed per unit time | RPS, QPS, MB/s | 5,000 requests/second |
| **Concurrency** | Requests in-flight simultaneously at any moment | count | 200 concurrent HTTP requests |
| **Bandwidth** | Maximum data transfer rate of a link | Mbps, GB/s | 1 Gbps NIC |
| **Utilization** | Fraction of capacity being used | 0.0–1.0 (0–100%) | CPU at 72% |

**Latency and throughput are not inverses.** A batch pipeline can have high throughput (millions of records/hour) and high latency (each record takes 30 seconds). A single-threaded serial handler can have low throughput (50 RPS) and low latency (2ms per request). You optimize them with different levers.

---

## Little's Law

The most important formula in queueing theory — it holds for any stable system in steady state:

```
L = λ × W
```

| Symbol | Meaning | Example |
|---|---|---|
| **L** | Average number of requests in the system (concurrency) | 200 in-flight requests |
| **λ** (lambda) | Arrival rate (throughput) | 1,000 RPS |
| **W** | Average time spent in system (latency) | 200ms |

**L = λ × W** → `200 = 1,000 × 0.2` ✓

```mermaid
flowchart LR
    Arrive["Requests\narriving\nλ = 1,000 RPS"] --> Queue["Queue +\nActive Processing\nL = 200 concurrent"]
    Queue --> Leave["Responses\nleaving\nλ = 1,000 RPS"]
    Queue -.->|"W = 200ms avg latency"| Timing["Time in System"]

    classDef input fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef queue fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef output fill:#27AE60,stroke:#1E8449,color:#fff
    classDef note fill:#ECF0F1,stroke:#BDC3C7,color:#333

    class Arrive input
    class Queue queue
    class Leave output
    class Timing note
```

### Worked Examples

**Example 1: Size a connection pool**
- Service receives 500 RPS
- Average DB query latency = 20ms
- Required pool size: L = 500 × 0.020 = **10 connections minimum**
- Add 50% headroom for bursts: **15–20 connections**

**Example 2: Find max sustainable throughput**
- Service can hold max 100 concurrent goroutines
- Average request latency = 50ms
- Max QPS = L/W = 100 / 0.050 = **2,000 RPS**

**Example 3: Diagnose a latency spike**
- Previously: 1,000 RPS, 50ms latency → 50 concurrent
- After deploy: 1,000 RPS, 500ms latency → 500 concurrent (10× more)
- Little's Law says concurrency grew 10×. Something is holding connections open: connection leak? slow dependency? timeout misconfiguration?

**Rearranged forms:**
- Max throughput: `λ_max = L_max / W`
- Expected latency: `W = L / λ`
- Required concurrency: `L = λ × W`

---

## Back-of-Envelope Estimation

A structured methodology for capacity planning:

```mermaid
flowchart TD
    Users["Start with Users\n(MAU / DAU)"] --> Activity["Estimate Activity\n(requests per user per day)"]
    Activity --> QPS["Calculate QPS\nQPS = DAU × req/day ÷ 86,400"]
    QPS --> Peak["Apply Peak Multiplier\npeak = avg × 2–3×"]
    Peak --> Storage["Estimate Storage\nbytes/event × events/day × retention"]
    Storage --> Bandwidth["Estimate Bandwidth\nQPS × response_size × 8 bits"]
    Bandwidth --> Infra["Size Infrastructure\ninstances = peak_QPS ÷ QPS_per_instance"]

    classDef step fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef calc fill:#27AE60,stroke:#1E8449,color:#fff

    class Users,Activity step
    class QPS,Peak,Storage,Bandwidth,Infra calc
```

### Key Multipliers

| Scenario | Formula | Example |
|---|---|---|
| **QPS from DAU** | `DAU × req/day ÷ 86,400` | 10M DAU × 30 req/day ÷ 86,400 = **3,472 QPS** |
| **Peak QPS** | `avg_QPS × 2–3×` | 3,472 × 3 = **~10K QPS peak** |
| **Storage/day** | `events/sec × bytes/event × 86,400` | 1K/s × 500B × 86,400 = **43 GB/day** |
| **Storage/year** | `GB/day × 365` | 43 × 365 = **~15 TB/year** |
| **Bandwidth** | `QPS × response_bytes × 8 ÷ 1e6` Mbps | 3,472 × 2,000B × 8 ÷ 1e6 = **55 Mbps** |
| **Instances needed** | `peak_QPS ÷ QPS_per_instance` | 10K ÷ 500 = **20 instances** |

**Twitter-scale example:** 300M MAU, 20% DAU = 60M DAU. 8 tweets read per user per day → 60M × 8 ÷ 86,400 ≈ **5,556 avg RPS** (~17K peak). Write ratio ~1:100 (reads >> writes) → ~55 writes/sec. Storage: 55 writes/sec × 500 bytes × 86,400 = **2.4 GB/day** of raw tweet text.

---

## Latency Percentiles

Average latency is a lie. Use percentiles:

```mermaid
flowchart LR
    subgraph Dist["Latency Distribution (example: API endpoint)"]
        direction TB
        P50["p50 = 12ms\n(median — half faster)"]
        P75["p75 = 28ms"]
        P90["p90 = 55ms"]
        P95["p95 = 95ms\n(SLA baseline)"]
        P99["p99 = 340ms\n(tail latency)"]
        P999["p999 = 1,800ms\n(worst outliers)"]
    end

    P50 --> P75 --> P90 --> P95 --> P99 --> P999

    classDef good fill:#27AE60,stroke:#1E8449,color:#fff
    classDef warn fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff

    class P50,P75 good
    class P90,P95 warn
    class P99,P999 bad
```

| Percentile | Meaning | At 10K RPS this affects | Use case |
|---|---|---|---|
| **p50** | Half of requests faster than this | 5,000 users/second | Typical user experience baseline |
| **p95** | 95% of requests at or below | 500 users/second | SLA definition, capacity planning |
| **p99** | 1% of requests slower | 100 users/second | "Tail latency" — real users, real pain |
| **p999** | 0.1% of requests slower | 10 users/second | Timeout/retry budget configuration |

**Why p99, not p50?** At 10K RPS, a p99 of 340ms means 100 users per second experience 340ms+ responses. Over one minute, that is 6,000 users with a degraded experience. Mean latency of 20ms masks this completely.

**Tail latency amplification:** A request that fans out to 100 backend calls completes when the *slowest* returns. If each backend has 1% chance of being slow (p99), the probability that *at least one* of 100 is slow = `1 - 0.99^100 = 63%`. The fan-out p99 becomes the single-backend p99.99. This is why Google's "hedged requests" pattern (send a duplicate request to a second server if the first doesn't respond within p95) is valuable at scale.

---

## Latency Reduction Levers

```mermaid
flowchart TD
    Slow["Slow Response\n(high W in Little's Law)"] --> Q1{"Where is\nthe time?"}

    Q1 -->|"repeated computation"| Cache["Caching\n(in-memory, Redis, CDN)\nReduce: DB/compute calls"]
    Q1 -->|"connection setup overhead"| Pool["Connection Pooling\nReduce: 3-RTT TLS handshake\namortized across requests"]
    Q1 -->|"blocking serial calls"| Async["Async / Parallel I/O\nReduce: fan-out latency\n(parallel vs sequential)"]
    Q1 -->|"far from user"| CDN2["CDN / Edge\nReduce: 300ms → 5ms\n(co-locate data with users)"]
    Q1 -->|"slow full computation"| Precomp["Pre-computation\nReduce: compute at write time\nnot read time"]

    Cache --> Result["Lower W\n→ Lower L (Little's Law)\n→ Higher max λ"]
    Pool --> Result
    Async --> Result
    CDN2 --> Result
    Precomp --> Result

    classDef problem fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef lever fill:#4A90D9,stroke:#2C6FAC,color:#fff
    classDef result fill:#27AE60,stroke:#1E8449,color:#fff
    classDef decision fill:#E67E22,stroke:#CA6F1E,color:#fff

    class Slow problem
    class Cache,Pool,Async,CDN2,Precomp lever
    class Result result
    class Q1 decision
```

---

## Throughput vs Latency Tradeoff

Batching is the classic lever that increases throughput at the cost of latency:

```mermaid
flowchart LR
    subgraph NoBatch["No Batching"]
        direction TB
        NB1["Request 1 → Write to DB immediately\nLatency: 5ms\nThroughput: 200 writes/sec"]
        NB2["Request 2 → Write to DB immediately"]
        NB3["Request 3 → Write to DB immediately"]
    end

    subgraph Batch["With Batching (100ms window)"]
        direction TB
        BUF["Buffer: collect 500 requests\nover 100ms window"] --> BW["Batch write to DB\n(500 rows in 1 transaction)"]
        BW --> Result2["Latency: 50–150ms avg\nThroughput: 5,000 writes/sec\n(25× improvement)"]
    end

    classDef slow fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef fast fill:#27AE60,stroke:#1E8449,color:#fff
    classDef buffer fill:#4A90D9,stroke:#2C6FAC,color:#fff

    class NB1,NB2,NB3 slow
    class BUF buffer
    class BW,Result2 fast
```

| Strategy | Effect on Latency | Effect on Throughput | Example |
|---|---|---|---|
| **Batching** | Increases (buffer wait) | Large increase | Kafka producer batch, DB bulk insert |
| **Async processing** | Appears lower (instant ack) | Increases | Message queue, fire-and-forget |
| **Caching** | Decreases (cache hit) | Increases | Redis, CDN, in-process LRU |
| **Connection pooling** | Decreases (no handshake) | Increases | DB pool, HTTP keep-alive |
| **Serializable isolation** | Increases (lock contention) | Decreases | Strong ACID transactions |
| **Hedged requests** | Decreases p99 | Slight decrease (extra load) | Duplicate request to backup server |

---

## Utilization and the Queue Cliff

```mermaid
flowchart LR
    U30["30% Utilization\nQueue depth ≈ 0\nLatency ≈ service time"] --> U60
    U60["60% Utilization\nQueue depth: small\nLatency: +20%"] --> U80
    U80["80% Utilization\nQueue depth: growing\nLatency: +80%"] --> U90
    U90["90% Utilization\nQueue depth: large\nLatency: 3–5× service time"] --> U100
    U100["100% Utilization\nQueue: unbounded growth\nLatency: → ∞ (system saturated)"]

    classDef good fill:#27AE60,stroke:#1E8449,color:#fff
    classDef warn fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff

    class U30,U60 good
    class U80 warn
    class U90,U100 bad
```

From queueing theory (M/M/1 queue): latency grows non-linearly as utilization approaches 100%. **Design target: 50–70% utilization headroom.** A system running at 80% CPU has very little burst capacity before queues explode.

`Max sustainable QPS (U=0.7) = 0.7 / service_time_seconds`

Example: 5ms service time → max 140 QPS per thread at 70% utilization. For 10K QPS, need `10,000 / 140 ≈ 72 threads/workers`.

---

## Interview Angles

**Q1: Estimate QPS for Twitter's home timeline.**
300M MAU, ~60M DAU (20%). Each user reads ~15 timeline refreshes/day × 20 tweets per refresh = 300 requests/day/user. QPS = 60M × 300 ÷ 86,400 ≈ **208K avg RPS**, ~600K peak. Write side: ~500M tweets/day = ~5,800 writes/sec. Fanout (precompute timelines) is the architectural challenge at this scale.

**Q2: Why optimize p99 latency rather than p50?**
At high QPS, p99 affects many real users per second. p50 (median) hides the tail — a p50 of 10ms with p99 of 2 seconds means 1% of users wait 2 seconds. Tail latency amplifies in fan-out architectures (see tail latency amplification). SLAs are typically defined on p99, not mean, because that is what users actually experience at the extremes.

**Q3: How does Little's Law help you size infrastructure?**
L = λ × W. If you know your target QPS (λ) and your expected latency (W), you can calculate required concurrency (L) — and therefore thread pool size, connection pool size, and number of worker processes. It is also a diagnostic tool: if L jumps unexpectedly, either λ increased or W increased (latency regression).

**Q4: What is tail latency amplification and how do you mitigate it?**
In a fan-out to N services, overall latency = max(all N). With N=100 and 1% chance of slowness per service, 63% of requests will be slow. Mitigations: hedged requests (duplicate request if no response within p75), aggressive timeouts per service, reducing fan-out width, and circuit breakers to fast-fail known-slow services.

**Q5: How does batching improve throughput at the cost of latency?**
Batching collects multiple requests into a single operation (one DB transaction for 500 rows instead of 500 individual writes). The per-item overhead (connection setup, transaction overhead, network round trips) is amortized across the batch — throughput goes up dramatically. But each item must wait for the batch window to fill before processing — adding buffer wait time to latency.

---

**References**

- [Little's Law — Wikipedia](https://en.wikipedia.org/wiki/Little%27s_law)
- [Latency Numbers Every Programmer Should Know](https://gist.github.com/jboner/2841832)
- [The Tail at Scale — Dean & Barroso, Google](https://dl.acm.org/doi/10.1145/2408776.2408794)
- [Systems Performance — Brendan Gregg](https://www.brendangregg.com/systems-performance-2nd-edition-book.html)
- [Queueing Theory for Dummies](https://www.johndcook.com/blog/2009/01/30/queueing-theory-basics/)

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/1.4_latency_and_throughput/`](../../../docs/diagrams/src/1.4_latency_and_throughput/).
> Regenerate: `just diagrams-topic 1.4_latency_and_throughput`.

### Overview

A request's latency budget itemized hop by hop: ISP 20ms, CDN 5ms, LB 1ms, app 50ms, DB 10ms — 86ms total, and where to attack it.

![Latency And Throughput — overview](../../../docs/diagrams/svg/1.4_latency_and_throughput/overview/index.svg)

**`cache_hit` scenario:** the CDN answers directly and the app/DB hops disappear — total latency collapses from 86ms to ~25ms, the single biggest lever in the pipeline.

![Latency And Throughput — cache_hit scenario](../../../docs/diagrams/svg/1.4_latency_and_throughput/overview/cache_hit.svg)

**`db_slow` scenario:** the 10ms DB hop degrades to 800ms (missing index, lock contention) and dominates everything — latency budgets are only as good as the worst hop.

![Latency And Throughput — db_slow scenario](../../../docs/diagrams/svg/1.4_latency_and_throughput/overview/db_slow.svg)

### Request Flow

Throughput scaling: how many requests per second the same pipeline sustains, and what saturates first.

![Latency And Throughput — request flow](../../../docs/diagrams/svg/1.4_latency_and_throughput/request-flow/index.svg)

**`thundering_herd` scenario:** a cache entry expires under load and all 4 app servers stampede the database simultaneously — the motivation for request coalescing and staggered TTLs.

![Latency And Throughput — thundering_herd scenario](../../../docs/diagrams/svg/1.4_latency_and_throughput/request-flow/thundering_herd.svg)
