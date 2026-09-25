# Evidence

## Structure
- LB page: 13,813 px tall, 3,318 words, 11 H2s, 31 buttons, 3 range inputs.
- LB page has **two load-balancer simulators** with the same algorithm tabs: canvas flow at y=1717, card sim at y=3322.
- Lab section sits at y=12,222 of 13,813 (89% down the page) — `load-balancing.mdx` "Lab: fix the broken load balancer".
- Lab/syslings present on **1 of 6** pages. Caching has 2 syslings exercises (`syslings/exercises/caching/`) but `caching.mdx` never mentions them.
- Landing Caching card: "Read → Sim → Run" (no Lab) — understates.
- Badges: "Has code" and "⚙ Runnable code" both on every page (`RunCode.astro:19`). Circuit breaker lacks "Docker required" although its stack is Docker Compose.
- Internal history in reader copy: `load-balancing.mdx:88` and `consistent-hashing.mdx:120` ("replaces … from the old site").

## Honesty (claims vs code)
- `caching.mdx` "This implementation": Go cache-aside against **Redis** with `rdb.Get`/`mockDB`. Repo `scalability/3.1_caching/`: nginx `proxy_cache`/`stale` confs + 747-byte Go server, **no Redis**.
- `rate-limiting.mdx`: Go `func (rl *RateLimiter) Allow` token bucket. Repo `scalability/3.2_rate_limiting/`: nginx `limit_req` confs (`burst/per_ip/per_key.conf`); no such Go func.
- `caching.mdx:28`: "A RAM read costs ~1 ms" — RAM is ~100 ns; ~1 ms is a Redis network round trip.
- `LoadBalancerFlow.tsx:265-267` reports p50/p95/p99 in `ms` though sim time is not real time; `CacheFlow.tsx:161` correctly uses `(rel) ×`.
- `db-replication.mdx`: "streaming replicas" — replica polls primary (`replica/main.go:49`). Minor.
- Circuit breaker, consistent hashing, load balancing samples match code.

## Visual
- Starlight tokens throughout; single type scale; consistent.
- Canvas sims hard-coded dark (`#0b0e14`) inside light theme; xyflow diagram nodes dark on light theme.
- Landing hero: text column only, right half empty at 1440 px.
- Dark-theme contrast: `.lsd-hint` and sim `<strong>` 4.49:1, `.lsd-metric-value` 3.96:1 (need 4.5).
- No horizontal overflow at 606 px.

## A11y
- Canvas has `aria-label`; range inputs labelled; skip link present; focus-visible outline (browser default) on sim buttons via Tab.
- Reduced-motion default-pause present (per `useCanvasSim.ts`).

## Weight
- `useCanvasSim.ts`: no IntersectionObserver / visibilitychange → rAF sim keeps running offscreen.
- LB page: 17 requests, 14 scripts; total dist JS ≈ 536 KB raw (React + xyflow chunks 177 KB / 131 KB) — INFERRED per-page share.

## Onboarding / syslings
- `README.md` Quick Start lists retired `just <topic>-lab` / `-lab-check N`; no mention of syslings.
- `syslings --help` → `unknown topic "--help"` (no usage text).
- `syslings/main.go:123` hard-codes `(available: load-balancing)`; caching also exists.
- Boot failure shown in TUI: "stack failed: … — is Docker running?" (`internal/tui/view.go:288`). Good.
