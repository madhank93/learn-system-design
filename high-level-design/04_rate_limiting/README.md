+++
title = "Rate Limiting"
description = "Token bucket, sliding window, and distributed rate limiting — with Go and Redis examples."
weight = 3
sort_by = "weight"
taxonomies = { tags = ["Go", "Redis", "Docker", "Distributed Systems"], categories = ["hld"] }

[extra]
toc_expand = true
+++

## What Is Rate Limiting?

Rate limiting controls how many requests a client can make to a service within a given time window. Without it, a single misbehaving client — malicious or accidental — can exhaust resources for every other user.

Four concrete reasons to enforce rate limits:

- **DDoS protection** — cap the blast radius of volumetric attacks before they reach your application layer.
- **Fair use** — prevent one tenant from monopolising shared infrastructure in a multi-tenant system.
- **Cost control** — downstream API calls, LLM tokens, or database writes often cost money per unit; rate limits cap the bill.
- **SLA enforcement** — tiered plans (free / pro / enterprise) carry different throughput guarantees; rate limiting is the mechanism that upholds them.

{% mermaid() %}
flowchart LR
    C([Client]) --> RL{Rate Limiter}
    RL -- allowed --> S([Service])
    RL -- rejected --> R([429 Too Many Requests])
    style R fill:#f66,color:#fff
    style S fill:#2a9d8f,color:#fff
{% end %}

---

## Where to Enforce Rate Limits

Rate limiting can live at multiple layers of your stack. Choosing the right layer — or combining layers — depends on your threat model and operational constraints.

| Layer | Pros | Cons |
|---|---|---|
| **Client-side SDK** | Reduces network traffic; immediate feedback to caller | Trivially bypassed by a malicious client; no server-side protection |
| **API Gateway** (Kong, AWS APIGW) | Centralised policy; no code changes in services | Single point of configuration drift; adds latency on the hot path |
| **Reverse proxy** (Nginx, Envoy) | Works for all traffic regardless of language/framework | Per-request overhead in the proxy; limited to L7 signals |
| **Application middleware** | Fine-grained control (per-user, per-endpoint, per-plan) | Scattered logic across services; requires shared state store |
| **Service mesh** (Istio, Linkerd) | Transparent; policy-as-code via control plane | Operational complexity; Envoy filter configuration is verbose |

Combining layers is common: an API Gateway applies a coarse global cap, while application middleware enforces per-user quotas.

{% mermaid() %}
flowchart TD
    Internet([Internet]) --> GW[API Gateway\nglobal cap]
    GW --> RP[Reverse Proxy\nIP-level cap]
    RP --> MW[App Middleware\nper-user quota]
    MW --> SVC[Business Service]
    MW --> Redis[(Redis\nshared counters)]
{% end %}

---

## Rate Limiting Algorithms

### Fixed Window Counter

Divide time into fixed windows (e.g., 0–59 s, 60–119 s). Increment a counter per window; reject once the counter exceeds the limit.

**Problem — boundary burst.** A client can fire `limit` requests at `t=59 s` and another `limit` requests at `t=60 s`. In a one-second span the service sees `2 × limit` requests, yet no rule was violated.

{% mermaid() %}
gantt
    title Fixed Window Boundary Burst
    dateFormat  s
    axisFormat  %S s

    section Window 1 (0–60 s)
    100 requests at t=58–59 :crit, w1, 58, 2s

    section Window 2 (60–120 s)
    100 requests at t=60–61 :crit, w2, 60, 2s
{% end %}

```
Window 1 [0–60s]:   ░░░░░░░░░░░░░░░░░░░░░░░░░░████  ← 100 req at t≈59
Window 2 [60–120s]: ████░░░░░░░░░░░░░░░░░░░░░░░░░░  ← 100 req at t≈60
                                                   ↑ 200 req in ~2s
```

**Storage:** one counter + one expiry per (user, window). O(1) space.

---

### Sliding Window Log

Store the exact timestamp of every accepted request. On each new request, evict all timestamps older than `now - window_size`, then count what remains. Reject if count ≥ limit.

**Accuracy:** exact — no boundary burst problem.

**Cost:** O(limit) memory per user in the worst case; each check is O(log n) with a sorted set.

{% mermaid() %}
flowchart LR
    subgraph Fixed Window
        direction TB
        FW1[Window 1\n0–60s\n100 req] --> FW2[Window 2\n60–120s\n100 req]
    end

    subgraph Sliding Window Log
        direction TB
        SW[Timestamps sorted\nEvict older than now-60s\nCount remaining]
    end

    Fixed Window -->|Boundary burst possible| B((200 req / 2s))
    Sliding Window Log -->|Always ≤ 100 req / 60s| G((Accurate))
    style B fill:#f66,color:#fff
    style G fill:#2a9d8f,color:#fff
{% end %}

---

### Token Bucket

A bucket holds up to `capacity` tokens. Tokens are added at a constant rate `R` (tokens/second). Each incoming request consumes one token. If the bucket is empty, the request is rejected.

**Controlled bursting:** a client that has been idle accumulates tokens up to `capacity`, then can fire a burst of that size. After the burst, throughput is capped at rate `R`.

{% mermaid() %}
flowchart TD
    T([Token Refiller\nR tokens/sec]) -->|add tokens| B[(Bucket\ncurrent: N\nmax: capacity)]
    R([Incoming Request]) --> D{N > 0?}
    B --> D
    D -- yes --> Con[Consume 1 token\nAllow request]
    D -- no --> Rej[Reject\n429]
    style Rej fill:#f66,color:#fff
    style Con fill:#2a9d8f,color:#fff
{% end %}

**Key parameters:**
- `capacity` — maximum burst size.
- `rate` — steady-state throughput (tokens/sec).

---

### Sliding Window Counter

A hybrid that avoids the memory cost of the log while being more accurate than fixed windows.

**Formula:**

```
count = prev_window_count × (1 - elapsed / window_size) + current_window_count
```

Where `elapsed` is how far into the current window we are. This linearly weights the previous window's contribution by the overlap fraction.

**Example:** window = 60 s, limit = 100, prev = 84 req, curr = 36 req, elapsed = 15 s.

```
count = 84 × (1 - 15/60) + 36 = 84 × 0.75 + 36 = 63 + 36 = 99  → allowed
```

{% mermaid() %}
flowchart LR
    P[Previous window\ncount: 84] -- "× 0.75 overlap" --> W[Weighted\ncount: 63]
    C[Current window\ncount: 36] --> SUM[Sum = 99\n≤ 100 limit\nAllow]
    W --> SUM
    style SUM fill:#2a9d8f,color:#fff
{% end %}

**Storage:** two counters per user. O(1) space, O(1) check.

---

### Algorithm Comparison

| Algorithm | Memory | Accuracy | Burst Handling | Complexity | Best For |
|---|---|---|---|---|---|
| Fixed Window Counter | O(1) | Low (boundary burst) | None | Simple | Coarse global caps |
| Sliding Window Log | O(limit) | Exact | None (strict) | Medium | Low-volume, strict APIs |
| Token Bucket | O(1) | High | Controlled burst | Medium | Most production APIs |
| Sliding Window Counter | O(1) | Good approximation | Partial | Medium | High-scale, memory-sensitive |

---

## Distributed Rate Limiting

A single in-memory counter breaks the moment you run more than one instance of your service. Each pod has its own counter; a client can hit `n × limit` requests across `n` pods.

**Shared state is required.** Redis is the de-facto choice: it is fast, single-threaded (no race between reads and writes within a single command), and supports atomic Lua scripts.

### Option 1 — `INCR` + `EXPIRE`

```redis
INCR user:123:2024060512   # key = user:window
EXPIRE user:123:2024060512 60
```

Race condition: `INCR` succeeds, process crashes before `EXPIRE` — key never expires. Fix: use `SET key 1 EX 60 NX` on the first request.

### Option 2 — Sliding Window with `ZADD` / `ZREMRANGEBYSCORE`

```redis
ZADD user:123 <now_ms> <request_id>
ZREMRANGEBYSCORE user:123 0 <now_ms - window_ms>
ZCARD user:123
EXPIRE user:123 <window_s>
```

These four operations must be wrapped in a Lua script (see below) for atomicity.

### Option 3 — Atomic Lua Script (Token Bucket)

```lua
-- Keys: KEYS[1] = bucket key
-- Args: ARGV[1] = capacity, ARGV[2] = rate (tokens/sec),
--       ARGV[3] = now (unix timestamp float), ARGV[4] = requested tokens

local key      = KEYS[1]
local capacity = tonumber(ARGV[1])
local rate     = tonumber(ARGV[2])
local now      = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens     = tonumber(data[1]) or capacity
local last_refill = tonumber(data[2]) or now

-- Refill
local elapsed = math.max(0, now - last_refill)
tokens = math.min(capacity, tokens + elapsed * rate)

if tokens >= requested then
    tokens = tokens - requested
    redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
    redis.call("EXPIRE", key, math.ceil(capacity / rate) + 1)
    return 1   -- allowed
else
    redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
    redis.call("EXPIRE", key, math.ceil(capacity / rate) + 1)
    return 0   -- rejected
end
```

Because Redis executes Lua scripts atomically, there is no window for a race condition between the read, calculation, and write.

{% mermaid() %}
flowchart TD
    C1([App Server 1]) --> Redis[(Redis\nshared bucket state)]
    C2([App Server 2]) --> Redis
    C3([App Server 3]) --> Redis
    Redis -- allowed/denied --> C1
    Redis -- allowed/denied --> C2
    Redis -- allowed/denied --> C3
{% end %}

{% alert_warning() %}
Every request now makes a round-trip to Redis. Keep Redis co-located (same data centre, low-latency link). For extremely high-throughput APIs, consider a local token bucket with periodic Redis sync (local rate ≈ global_rate / n_instances) to reduce Redis pressure.
{% end %}

---

## Trade-offs and Failure Modes

| Concern | Issue | Mitigation |
|---|---|---|
| Redis unavailable | Rate limiter is non-functional | Fail open (allow all) during outage; alert on Redis health; consider circuit breaker |
| Clock skew | Sliding window timestamps diverge across nodes | Use NTP with sub-millisecond precision; Redis `TIME` command for server-side timestamp |
| Hot keys | High-traffic users hammer a single Redis shard | Partition key: `rl:{user_id % 16}:{user_id}` to spread across slots |
| Race conditions | INCR + EXPIRE non-atomic | Atomic Lua scripts or Redis transactions (`MULTI`/`EXEC`) |
| Over-limiting | Legitimate traffic rejected at peak | Observe p99 request rates, add burst headroom, use sliding window counter not fixed |
| Large burst abuse | Token bucket with high capacity = large burst window | Tune `capacity` separately from `rate`; set capacity ≈ 2–5× expected per-second rate |

- **Fail open vs fail closed.** If Redis is down and you fail closed (deny all), you have a self-inflicted outage. If you fail open (allow all), you are exposed to abuse for the duration of the outage. Neither is universally correct — choose based on your threat model and SLA obligations.

- **Clock skew** in distributed sliding window: NTP keeps most production servers within a few milliseconds. For tighter guarantees, read the timestamp from Redis itself (`TIME` command) rather than each app server's local clock.

- **Hot keys in Redis**: a single high-traffic user's key can become a bottleneck. Prefix-partition by `user_id % N` so requests spread across multiple Redis slots (useful with Redis Cluster).

- **Race conditions**: always use atomic Lua scripts for multi-step read-modify-write operations against Redis. `WATCH`/`MULTI`/`EXEC` transactions work but require retry logic on the client side.

- **Over-limiting**: start by observing your actual traffic distribution (p50/p95/p99 per-user request rates) before setting limits. Add 20–50% burst headroom above the p99 to avoid false positives for legitimate peaks.

---

## Go Implementation: Token Bucket Middleware

The implementation below is self-contained — no Redis dependency. It uses an in-memory map protected by a mutex, suitable for single-instance deployments or local testing.

```go
package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// TokenBucket holds the state for one client's bucket.
type TokenBucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimiter manages per-IP token buckets.
type RateLimiter struct {
	capacity float64       // max tokens (burst size)
	rate     float64       // tokens added per second
	buckets  map[string]*TokenBucket
	mu       sync.Mutex
}

// NewRateLimiter creates a limiter with the given capacity and refill rate.
func NewRateLimiter(capacity, rate float64) *RateLimiter {
	return &RateLimiter{
		capacity: capacity,
		rate:     rate,
		buckets:  make(map[string]*TokenBucket),
	}
}

// bucket retrieves or creates the bucket for a given key.
func (rl *RateLimiter) bucket(key string) *TokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &TokenBucket{
			tokens:     rl.capacity,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
	}
	return b
}

// Allow returns true if the request is within the rate limit.
// It refills the bucket based on elapsed time, then attempts to consume one token.
func (rl *RateLimiter) Allow(key string) bool {
	b := rl.bucket(key)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.lastRefill = now

	// Refill tokens, capped at capacity.
	b.tokens += elapsed * rl.rate
	if b.tokens > rl.capacity {
		b.tokens = rl.capacity
	}

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// Middleware wraps an http.Handler with token-bucket rate limiting per remote IP.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// In production, parse X-Forwarded-For or X-Real-IP instead.
		ip := r.RemoteAddr

		if !rl.Allow(ip) {
			retryAfter := int(rl.capacity/rl.rate) + 1
			w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfter))
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", rl.capacity))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Allow bursts of up to 10 requests; refill at 2 tokens/second.
	limiter := NewRateLimiter(10, 2)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	handler := limiter.Middleware(mux)

	fmt.Println("Listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
```

**Key design decisions:**

- `RateLimiter.mu` guards the `buckets` map (create/lookup). `TokenBucket.mu` guards per-bucket state (refill/consume). Two-level locking avoids a global bottleneck under concurrent load.
- Refill is lazy (computed on each `Allow()` call), not via a background goroutine, which keeps the implementation simple and avoids goroutine leaks.
- In production, replace the `buckets` map with an LRU cache (e.g., `github.com/hashicorp/golang-lru`) to bound memory as the number of distinct clients grows.

### Redis Lua Script for Distributed Token Bucket

```lua
-- KEYS[1] = bucket key (e.g., "rl:user:42")
-- ARGV[1] = capacity    (max tokens)
-- ARGV[2] = rate        (tokens per second, float)
-- ARGV[3] = now         (current unix timestamp, float)
-- ARGV[4] = cost        (tokens to consume, usually 1)
-- Returns: 1 = allowed, 0 = rejected

local key       = KEYS[1]
local capacity  = tonumber(ARGV[1])
local rate      = tonumber(ARGV[2])
local now       = tonumber(ARGV[3])
local cost      = tonumber(ARGV[4])

local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens      = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens      = capacity
    last_refill = now
end

local elapsed = math.max(0, now - last_refill)
tokens = math.min(capacity, tokens + elapsed * rate)

local allowed = 0
if tokens >= cost then
    tokens  = tokens - cost
    allowed = 1
end

local ttl = math.ceil(capacity / rate) + 1
redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
redis.call("EXPIRE", key, ttl)

return allowed
```

Call this script from Go using `go-redis`:

```go
var rateLimitScript = redis.NewScript(`... lua script above ...`)

allowed, err := rateLimitScript.Run(ctx, rdb,
    []string{"rl:user:" + userID},
    capacity, rate, float64(time.Now().UnixMilli())/1000.0, 1,
).Int()
```

---

## Rate Limit Response Design

Consistent headers allow clients to implement back-off logic without guessing.

| Header | Meaning |
|---|---|
| `X-RateLimit-Limit` | Maximum requests allowed in the window |
| `X-RateLimit-Remaining` | Requests remaining in the current window |
| `X-RateLimit-Reset` | Unix timestamp (seconds) when the window resets or the bucket refills to `capacity` |
| `Retry-After` | Seconds to wait before retrying (required on 429 by RFC 6585) |

**Sample HTTP 429 response:**

```http
HTTP/1.1 429 Too Many Requests
Content-Type: text/plain; charset=utf-8
Retry-After: 5
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1717588265

Too Many Requests
```

{% alert_info() %}
Be consistent. Clients and SDKs read these headers to implement automatic retry logic with exponential back-off. If the headers are missing on some endpoints or use different semantics (e.g., `X-RateLimit-Reset` as a relative delta on one service and an absolute timestamp on another), clients cannot be written generically and will break in subtle ways.
{% end %}

The HTTP 429 status code is defined in [RFC 6585](https://tools.ietf.org/html/rfc6585). Always return `Retry-After` alongside it — some proxies and CDNs interpret 429 with `Retry-After` to automatically queue and retry the request on the client's behalf.

---

## How to Practice

- **Build the middleware above**, then use `wrk` or `hey` to drive load and confirm 429s appear at the right threshold. Tune `capacity` and `rate` and observe burst behaviour.
- **Replace the in-memory map** with Redis and the Lua script. Spin up two instances of the service on different ports, share the same Redis, and verify that combined throughput across both instances is still capped at the configured limit.
- **Implement the sliding window counter** algorithm from scratch. Compare its behaviour to token bucket at the window boundary by replaying the same request trace against both.
- **Simulate Redis failure**: stop the Redis container mid-test and observe fail-open vs fail-closed behaviour. Implement a circuit breaker that switches to local in-memory limits when Redis is unreachable.
- **Add observability**: emit a `rate_limit_rejected_total` Prometheus counter per endpoint and per user tier. Build a Grafana dashboard. Set an alert when the rejection rate exceeds 5%.
- **Test clock skew**: run two app servers with intentionally offset system clocks (use `date -s` in Docker containers) and observe how sliding window counters drift.
- **Implement tiered limits**: define multiple `RateLimiter` instances (free, pro, enterprise) and route each authenticated request to the appropriate limiter based on a JWT claim.

---

## References

1. Kleppmann, M. *Designing Data-Intensive Applications*. O'Reilly, 2017. Chapter 8 covers distributed system failure modes relevant to shared rate limit state.
2. Redis. *Rate Limiting*. [https://redis.io/glossary/rate-limiting/](https://redis.io/glossary/rate-limiting/)
3. Martin, D. *The System Design Primer*. [https://github.com/donnemartin/system-design-primer](https://github.com/donnemartin/system-design-primer)
4. Stripe Engineering Blog. *Scaling your API with rate limiters*. Stripe, 2016.
5. IETF RFC 6585. *Additional HTTP Status Codes* (defines HTTP 429). [https://tools.ietf.org/html/rfc6585](https://tools.ietf.org/html/rfc6585)
