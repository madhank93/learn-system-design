+++
title = "Caching"
description = "Cache placement strategies, eviction policies, cache-aside vs write-through — with Go and Redis examples."
weight = 2
sort_by = "weight"
taxonomies = { tags = ["Go", "Redis", "Docker", "Distributed Systems"], categories = ["hld"] }

[extra]
toc_expand = true
+++

## What Is Caching?

Caching is the practice of storing the result of an expensive computation or I/O operation in fast storage so subsequent requests can be served without repeating that work. In distributed systems, the canonical slow operation is a database query; a cache (typically an in-memory store like Redis) sits in front of it and absorbs the majority of reads.

{% mermaid() %}
flowchart LR
    C([Client]) --> Cache[(Cache)]
    Cache -->|Cache HIT| C
    Cache -->|Cache MISS| DB[(Database)]
    DB --> Cache
    Cache -->|Return & populate| C
{% end %}

---

## Where to Cache

Every layer between the user and the origin database is an opportunity to cache. Choosing the right layer depends on what you are caching, how stale it can be, and who benefits.

| Layer | Examples | What it caches | Latency saving |
|---|---|---|---|
| Client-side | Browser cache, mobile app cache | Static assets, API responses | Eliminates network round-trip entirely |
| CDN | Cloudflare, CloudFront, Fastly | HTML, images, JS/CSS, API responses | ~50–200 ms (edge PoP vs origin) |
| Reverse proxy / Edge | Nginx proxy_cache, Varnish | Full HTTP responses, partial pages | ~5–50 ms (avoids app tier) |
| Application layer | Redis, Memcached | DB rows, computed aggregates, sessions | ~1–10 ms (RAM vs disk) |
| Database query cache | MySQL query cache, PostgreSQL pg_prewarm | Query result sets | ~5–50 ms (avoids parse + plan + I/O) |

{% alert_info() %}
Database-level query caches (e.g., MySQL's now-removed query cache) are rarely worth the complexity. Application-layer caches (Redis) give you more control over granularity, TTL, and invalidation.
{% end %}

{% mermaid() %}
flowchart LR
    User([User])
    CDN[CDN / Edge]
    RP[Reverse Proxy\nNginx / Varnish]
    App[App Server]
    ACache[(Redis / Memcached)]
    DB[(Database)]

    User --> CDN
    CDN -->|Cache MISS| RP
    RP -->|Cache MISS| App
    App -->|Cache MISS| ACache
    ACache -->|Cache MISS| DB
    DB -->|Result| ACache
    ACache -->|Result| App
    App -->|Result| RP
    RP -->|Result| CDN
    CDN -->|Result| User
{% end %}

---

## Cache Strategies

### Cache-Aside (Lazy Loading)

The application owns the cache interaction. On a read, it checks the cache first; on a miss it fetches from the database and populates the cache before returning.

{% mermaid() %}
sequenceDiagram
    participant App
    participant Cache
    participant DB

    App->>Cache: GET user:42
    Cache-->>App: MISS
    App->>DB: SELECT * FROM users WHERE id=42
    DB-->>App: row
    App->>Cache: SET user:42 {row} EX 300
    App-->>App: return row
{% end %}

### Read-Through

The cache itself is responsible for loading from the database on a miss. The application only ever talks to the cache; the cache provider (or a sidecar library) handles the fallback. Common in ORM-level caches.

{% mermaid() %}
sequenceDiagram
    participant App
    participant Cache
    participant DB

    App->>Cache: GET user:42
    Cache->>DB: SELECT * FROM users WHERE id=42 (internal)
    DB-->>Cache: row
    Cache-->>App: row (now cached)
{% end %}

### Write-Through

Every write goes to the cache and the database synchronously before the request completes. Keeps cache warm but adds write latency.

{% mermaid() %}
sequenceDiagram
    participant App
    participant Cache
    participant DB

    App->>Cache: SET user:42 {data}
    Cache->>DB: UPDATE users SET ... WHERE id=42
    DB-->>Cache: OK
    Cache-->>App: OK
{% end %}

### Write-Behind (Write-Back)

Writes land in the cache immediately; the cache asynchronously flushes to the database. Lowest write latency but risks data loss if the cache crashes before flushing.

{% mermaid() %}
sequenceDiagram
    participant App
    participant Cache
    participant DB

    App->>Cache: SET user:42 {data}
    Cache-->>App: OK (immediate)
    Note over Cache,DB: async flush (seconds later)
    Cache->>DB: UPDATE users SET ... WHERE id=42
    DB-->>Cache: OK
{% end %}

{% alert_warning() %}
Write-behind introduces a durability gap. If Redis goes down between the write acknowledgement and the DB flush, data is lost. Only use it when you can tolerate occasional loss (counters, analytics) or when Redis is configured with AOF persistence and you accept replay lag.
{% end %}

### Strategy Comparison

| Strategy | Write Latency | Read Latency | Consistency | Best For |
|---|---|---|---|---|
| Cache-Aside | DB latency | Cache hit: fast, miss: DB | Eventual | General-purpose reads |
| Read-Through | DB latency | Same as cache-aside | Eventual | ORM / library-managed reads |
| Write-Through | Cache + DB | Fast (always warm) | Strong | Read-heavy, write-tolerant |
| Write-Behind | Cache only (fast) | Fast (always warm) | Eventual / lossy | High-write, loss-tolerant |

---

## Eviction Policies

When memory is full, the cache must evict entries. The policy determines which entries are removed.

**LRU (Least Recently Used)** — evicts the entry that has not been accessed for the longest time. Works well for temporal locality; most web workloads fit this model.

**LFU (Least Frequently Used)** — evicts the entry accessed the fewest times over a window. Better for Zipfian distributions (popular items stay hot longer) but higher bookkeeping cost.

**TTL (Time-To-Live)** — entries expire automatically after a fixed duration regardless of access frequency. Simplest to reason about; guarantees freshness bounds.

| Policy | Complexity | Best For |
|---|---|---|
| LRU | O(1) with linked hash map | General reads, temporal locality |
| LFU | O(log n) with min-heap | Skewed popularity distributions |
| TTL | O(1) | Freshness-critical data (tokens, prices) |
| Random | O(1) | Approximation, very low overhead |

Redis supports `allkeys-lru`, `allkeys-lfu`, `volatile-lru`, `volatile-lfu`, `volatile-ttl`, and `allkeys-random` as `maxmemory-policy` values.

---

## Trade-offs and Failure Modes

| Concern | Issue | Mitigation |
|---|---|---|
| Cache stampede | Many requests hit DB simultaneously on cold cache | Probabilistic early expiry (PER), Redis locks (`SET NX`), request coalescing |
| Stale data | Cache returns outdated rows after a write | TTL tuning, event-driven invalidation (publish on write) |
| Cold start | Fresh deployment, empty cache, full DB load | Pre-warm cache at startup, gradual traffic ramp |
| Memory pressure | Evictions spike, hit ratio drops | Monitor `evicted_keys`, right-size `maxmemory`, tune eviction policy |
| Network partition | App cannot reach Redis | Circuit breaker (fail open to DB), health checks, fallback logic |
| Cache poisoning | Corrupted or malicious data stored in cache | Validate before caching, short TTLs for critical data |

**Cache stampede** occurs when a high-traffic key expires and hundreds of goroutines simultaneously attempt to recompute it. Solutions: use `SET key value NX EX ttl` (only one writer wins), or probabilistic early expiry — recompute before expiry with probability that increases as TTL drains.

**Stale data** is the fundamental trade-off of caching. Choose TTL based on acceptable staleness. For strict consistency, invalidate the cache key on every write (cache-aside + delete on write).

**Cold start** after a deployment or failover can overwhelm the database. Pre-warm by replaying recent keys from the DB or a snapshot, or use read-through with connection pooling to coalesce burst reads.

**Cache eviction under pressure** means the cache is too small for the working set. Monitor `hit_rate = keyspace_hits / (keyspace_hits + keyspace_misses)`. Target > 90% for most caches; below 80% indicates under-provisioning.

**Network partition to Redis** means every request falls through to the database. Wrap Redis calls in a circuit breaker (e.g., `sony/gobreaker`) and fall back gracefully. Never let a Redis outage take down the application.

---

## Go Implementation: Cache-Aside with Redis

### HTTP handler

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const cacheTTL = 5 * time.Minute

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// mockDB simulates a database read.
func mockDB(ctx context.Context, id string) (*User, error) {
	// Simulate DB latency
	time.Sleep(10 * time.Millisecond)
	if id == "404" {
		return nil, errors.New("user not found")
	}
	return &User{ID: id, Name: "Alice", Email: "alice@example.com"}, nil
}

func userHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		cacheKey := fmt.Sprintf("user:%s", id)

		// 1. Check cache
		val, err := rdb.Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache HIT
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(val))
			return
		}
		if !errors.Is(err, redis.Nil) {
			// Redis error — log and fall through to DB
			log.Printf("redis GET error for key %s: %v", cacheKey, err)
		}

		// 2. Cache MISS — fetch from DB
		user, err := mockDB(ctx, id)
		if err != nil {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// 3. Serialise
		data, err := json.Marshal(user)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		// 4. Populate cache (best-effort — do not fail request on error)
		if setErr := rdb.Set(ctx, cacheKey, data, cacheTTL).Err(); setErr != nil {
			log.Printf("redis SET error for key %s: %v", cacheKey, setErr)
		}

		w.Header().Set("X-Cache", "MISS")
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /user/{id}", userHandler(rdb))

	log.Println("server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
```

Key design decisions:
- Redis errors on `GET` fall through to the database rather than returning 500 — the cache is an optimisation, not a hard dependency.
- Redis errors on `SET` are logged but do not fail the request — partial cache population is acceptable.
- `X-Cache: HIT/MISS` headers make behaviour observable during testing.

### docker-compose.yml

```yaml
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      redis:
        condition: service_healthy
    environment:
      - REDIS_ADDR=redis:6379

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    command: ["redis-server", "--maxmemory", "256mb", "--maxmemory-policy", "allkeys-lru"]
```

### Dockerfile (minimal)

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
```

---

## Cache Hit Ratio and Observability

A cache with a 50% hit ratio halves database load. A cache at 95% hit ratio reduces database load by 20x. Hit ratio is the single most important cache health metric.

```
hit_ratio = keyspace_hits / (keyspace_hits + keyspace_misses)
```

Redis exposes these counters via `INFO stats`:

```
keyspace_hits:9812
keyspace_misses:512
evicted_keys:0
used_memory_human:48.32M
maxmemory_human:256.00M
```

**Key metrics to track:**

| Metric | Redis INFO field | Alert threshold |
|---|---|---|
| Hit ratio | `keyspace_hits` / total | < 90% |
| Evictions per second | `evicted_keys` delta | > 0 sustained |
| Memory usage | `used_memory` / `maxmemory` | > 80% |
| Connected clients | `connected_clients` | Near `maxclients` |
| Expired keys | `expired_keys` | Sudden spike indicates mass TTL expiry |

{% alert_info() %}
Use `redis-cli --stat` for a live per-second view, or scrape `/metrics` via `redis_exporter` for Prometheus.
{% end %}

{% mermaid() %}
flowchart LR
    Redis[(Redis)] -->|INFO stats\nevery 15s| Exporter[redis_exporter]
    Exporter -->|/metrics| Prometheus[(Prometheus)]
    Prometheus --> Grafana[Grafana Dashboard]
    Prometheus -->|AlertManager| PD[PagerDuty / Slack]
{% end %}

Recommended Grafana panels: hit ratio over time, evictions/sec, memory utilisation %, top keys by access frequency (use `redis-cli --hotkeys` with LFU policy).

---

## How to Practice

- **Instrument cache-aside manually**: wire a Go service to Postgres with a Redis cache-aside layer; use `redis-cli monitor` to watch cache reads and writes in real time.
- **Measure hit ratio impact**: run `wrk` or `k6` against the service, observe `INFO stats` before and after pre-warming the cache, and chart the DB QPS reduction.
- **Simulate cache stampede**: set a single key with a 1-second TTL, then fire 500 concurrent requests at expiry time; observe DB spike; implement mutex-based coalescing with `SET NX` and measure the difference.
- **Test eviction policies**: fill Redis to `maxmemory` with synthetic keys under `allkeys-lru` vs `allkeys-lfu`, then replay a Zipfian access pattern; compare hit ratios.
- **Implement write-through**: extend the Go handler to write to both Redis and the DB on `POST /user`; add a chaos monkey that randomly drops Redis writes; verify the DB stays consistent.
- **Build a pre-warmer**: write a Go script that queries `SELECT id FROM users ORDER BY updated_at DESC LIMIT 10000` at startup and populates Redis; measure cold-start DB load before and after.
- **Circuit breaker**: wrap the Redis client with `sony/gobreaker`; kill the Redis container mid-load-test; verify the application degrades gracefully to direct DB reads with no 5xx responses.

---

## References

1. Kleppmann, Martin. *Designing Data-Intensive Applications*. O'Reilly Media, 2017.
2. Redis documentation — [https://redis.io/docs/](https://redis.io/docs/)
3. Donne Martin. *System Design Primer* — [https://github.com/donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer)
4. AWS ElastiCache documentation — [https://docs.aws.amazon.com/elasticache/](https://docs.aws.amazon.com/elasticache/)
