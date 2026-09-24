# Diagrams

Architecture diagrams for the learn-system-design repo.

D2 is the source of truth. SVGs and GIFs are derived artifacts — do not edit them manually.

---

## Directory layout

```
docs/diagrams/
  src/           ← D2 source files (manually edited)
    <topic>/
      overview.d2
      request-flow.d2
      algorithms.d2
      ...
  svg/           ← Generated SVG outputs (commit these)
    <topic>/
      <name>.svg           — single-state diagram
      <name>/index.svg     — base state for multi-scenario or multi-step diagrams
      <name>/server_down.svg   — named scenario
      <name>/steps/1.svg   — individual step frames
  gif/           ← Generated animated GIFs (commit these)
    <topic>/
      <name>.gif           — animated steps or scenarios
  assets/        ← Local icon/image files (SVG/PNG) for use in D2 source
```

---

## Render defaults

All diagrams use:
- `--sketch` — hand-drawn aesthetic, easier to read for architecture docs
- `--layout elk` — ELK layered layout, minimises edge crossings in complex graphs
- `--pad 40` — tight margins (default 100 wastes canvas)

Override per-file only when there is a clear readability reason. Document the override in a comment at the top of the `.d2` file.

## Quality gates (enforced by `gen-diagrams.sh` + `check-diagrams.sh`)

The full meta prompt lives in the plan archive; the enforced rules:

| Gate | Rule | Enforcement |
|---|---|---|
| **Size (D1)** | viewBox ≤ 1400×1600 px | render fails on violation |
| **Palette (R5)** | no sketch-hatching fills (`E67E22 F39C12 2ECC71 EC7063 AEB6BF 7FB3D3 F1C40F`) | render fails on violation |
| **Density (R2)** | ≤ 8 shapes + 2 containers per diagram — decompose, never shrink | checker warning |
| **GIF pacing (G2)** | `steps:` blocks, one visual change per frame, 3–6 frames, 4000 ms interval | checker error if interval drifts |

Authoring rules (not machine-checked):
- **One idea per diagram** — split along write/read path, control/data plane, or per-variant grid
- **Labels**: ≤ 3 words/line, ≤ 2 lines per node; ≤ 3 words per edge; detail goes to page captions
- **Numbered edges** ("1. GET", "2. miss") for flows; the page lists the steps in prose
- **Card background** on every file root: `style: { fill: "#FBFAF7"; stroke: "#D5D0C8" }` — the diagram reads as a figure card on both dark and light site themes
- **5 semantic color roles only**: primary `#3498DB`, success `#27AE60`, danger `#E74C3C`, warn `#D35400`, muted `#5D6D7E` — max 3 per diagram; state change = color + label change
- **GIF companions**: every GIF embeds with a numbered step list in prose and the final-state frame (`steps/N.svg`) as a static image
- **Scenarios stay static** — never animate between whole states
- **No tooltips** — D2 tooltip icons render as dead i-buttons inside img embeds; explanation lives in labels and page captions

---

## Regenerate

**All diagrams:**
```sh
# from repo root
just diagrams
```

**One topic:**
```sh
just diagrams-topic 2.1_load_balancer
```

**Direct script:**
```sh
bash scripts/gen-diagrams.sh                  # all
bash scripts/gen-diagrams.sh 2.1_load_balancer  # one topic
```

**Requirements:** `d2 >= 0.6.0`
```sh
brew install d2
d2 --version
```

GIF rendering downloads Chromium + FFMPEG on first run via Playwright (cached in `~/.cache/ms-playwright`).

---

## Adding a new diagram topic

1. Create `docs/diagrams/src/<topic>/` and write `.d2` files.
2. If the topic has animated diagrams (steps or scenarios), add entries to the `ANIMATED` array in `scripts/gen-diagrams.sh`.
3. Create matching output dirs: `docs/diagrams/svg/<topic>/` and `docs/diagrams/gif/<topic>/` (the script creates them automatically).
4. Run `just diagrams-topic <topic>`.
5. Update the topic README to reference the generated SVGs/GIFs.
6. Commit both source and outputs.

---

## D2 feature usage guide

| Feature | When to use | Example files |
|---|---|---|
| `scenarios` | Alternate operating modes: healthy vs degraded, active vs passive | `overview.d2` |
| `steps` | Request lifecycle, startup sequence, troubleshooting walkthrough | `request-flow.d2`, `health-checks.d2` |
| Containers | Logical boundaries: VPC, cluster, app tier, data tier | all files |
| `shape: hexagon` | Load balancer, API gateway | `overview.d2` |
| `shape: cylinder` | Database, cache, queue | future topics |
| `shape: cloud` | DNS, CDN, external SaaS | `overview.d2` |
| `shape: person` | End user, client, browser | `overview.d2` |
| `style.stroke-dash` | Async, health check, fallback paths | all files |
| `style.animated: true` | Highlight active data flow in a step | `request-flow.d2` |

---

## Topics covered

All 35 topics ship `overview.d2` + `request-flow.d2` (2.1 additionally has `algorithms.d2` + `health-checks.d2`). Every diagram has a base `index.svg` plus one SVG per scenario. GIFs exist for topics listed in the `ANIMATED` array of `scripts/gen-diagrams.sh`.

| Topic | Scenarios (overview / request-flow) | GIF |
|---|---|---|
| [1.1 Monoliths vs Microservices](src/1.1_monoliths_vs_microservices/) | deploy_risk / microservice_failure | — |
| [1.2 Networking Fundamentals](src/1.2_networking_fundamentals/) | tcp_handshake, dns_cache_hit / http2_multiplexing | — |
| [1.3 Scaling Fundamentals](src/1.3_scaling_fundamentals/) | horizontal_growth / session_sticky | — |
| [1.4 Latency and Throughput](src/1.4_latency_and_throughput/) | cache_hit, db_slow / thundering_herd | — |
| [1.5 Availability and SLAs](src/1.5_availability_and_slas/) | region_a_down / budget_exhausted | — |
| [1.6 ACID vs BASE](src/1.6_acid_vs_base/) | acid_rollback, base_conflict / network_partition | — |
| [2.1 Load Balancer](src/2.1_load_balancer/) | server_down + algorithms, health-checks / sticky_failover | overview, request-flow, health-checks |
| [2.2 CDN](src/2.2_cdn/) | cache_hit, cache_purge / cache_purge | — |
| [2.3 DNS](src/2.3_dns/) | cache_hit, ttl_expired / dnssec | — |
| [2.4 API Gateway](src/2.4_api_gateway/) | auth_failure, rate_limited / circuit_open | — |
| [2.5 WebSockets and SSE](src/2.5_websockets_sse/) | ws_reconnect, sse_backpressure / server_restart | — |
| [3.1 Caching](src/3.1_caching/) | cache_hit, cache_stampede / write_through | request-flow |
| [3.2 Rate Limiting](src/3.2_rate_limiting/) | burst_allowed, fixed_window_exploit / distributed_race | — |
| [3.3 DB Sharding](src/3.3_db_sharding/) | hotspot, shard_split / resharding | — |
| [3.4 DB Replication](src/3.4_db_replication/) | replica_lag, primary_failure / split_brain | request-flow |
| [4.1 CAP Theorem](src/4.1_cap_theorem/) | network_partition / partition_healed | — |
| [4.2 SQL vs NoSQL](src/4.2_sql_vs_nosql/) | join_heavy, schema_flexible / acid_payment | — |
| [4.3 Consistent Hashing](src/4.3_consistent_hashing/) | node_removed, hotspot / rebalance | — |
| [4.4 CQRS + Event Sourcing](src/4.4_cqrs_event_sourcing/) | projection_rebuild + event-log / temporal_query | — |
| [4.5 Leader Election](src/4.5_leader_election/) | leader_crash / network_partition | request-flow |
| [5.1 Circuit Breaker](src/5.1_circuit_breaker/) | tripped / cascade_prevention | request-flow |
| [5.2 Retry + Backoff](src/5.2_retry_backoff/) | thundering_herd, jitter_spread / non_retryable | overview |
| [5.3 Distributed Locking](src/5.3_distributed_locking/) | lock_acquired, lock_expired / crash_recovery | — |
| [6.1 Message Queues](src/6.1_message_queues/) | consumer_lag / consumer_failure | — |
| [6.2 Service Discovery](src/6.2_service_discovery/) | instance_down / stale_cache | — |
| [6.3 Distributed Tracing](src/6.3_distributed_tracing/) | slow_span + trace-timeline / sampling | — |
| [7.1 Bloom Filter](src/7.1_bloom_filter/) | false_positive_rate / reset | — |
| [7.2 Geohashing](src/7.2_geohashing/) | proximity_search / precision_tradeoff | — |
| [7.3 Indexing](src/7.3_indexing/) | index_hit + index-types / index_bloat | — |
| [8.1 OAuth + JWT](src/8.1_oauth_jwt/) | token_expired, invalid_signature + jwt-structure / key_rotation | — |
| [SD01 URL Shortener](src/SD01_url_shortener/) | cache_hit / collision | — |
| [SD02 Twitter Feed](src/SD02_twitter_feed/) | celebrity_tweet / hot_celebrity | request-flow |
| [SD03 Ride Sharing](src/SD03_ride_sharing/) | surge_pricing / no_drivers | — |
| [SD04 Realtime Chat](src/SD04_realtime_chat/) | offline_delivery / typing_indicator | — |
| [SD05 Video Streaming](src/SD05_video_streaming/) | bandwidth_drop, cdn_miss + upload-pipeline / quality_switch, geo_restriction | — |

### Sketch-mode palette constraint

D2 `--sketch` renders high-luminance fills (`#E67E22`, `#F39C12`, `#2ECC71`, `#F1C40F`, `#AEB6BF`, light pastels used as node fills) as a hatched pattern that destroys label legibility. Use the sketch-safe equivalents for **node fills** (light tints remain fine for container backgrounds):

| Role | Use | Not |
|---|---|---|
| Orange (queue/async/warning) | `#D35400` | `#E67E22`, `#F39C12` |
| Gold (client/person/highlight) | `#9A7D0A` | `#F39C12`, `#F1C40F` |
| Green | `#27AE60` | `#2ECC71` |
| Red | `#E74C3C` | `#EC7063` |
| Gray (inactive) | `#5D6D7E` | `#AEB6BF` |
| Blue | `#3498DB` | `#7FB3D3` |

Also: do not set `style.font-color: "#FFFFFF"` on `shape: person` nodes — the label renders below the icon on the (light) container background, not on the shape.

---

## Git policy

- Commit all `.d2` source files.
- Commit all generated SVGs (`docs/diagrams/svg/`).
- Commit GIFs (`docs/diagrams/gif/`) — they are reasonable in size (~0.5–2 MB) and provide value for docs and demos.
- Never manually edit SVG or GIF files — regenerate instead.
