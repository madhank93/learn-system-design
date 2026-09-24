# Learn System Design — Content / UX / DevEx Audit

**Date:** 2026-06-25
**Scope:** Deep audit of the 6 live interactive Astro topics + beginner→advanced coverage map of the full 35-topic curriculum + DevEx review of the runnable implementations.
**Method:** Drove the live site (`http://localhost:4321`) as four personas with Playwright (desktop 1440, mobile 390, light + dark), read all 6 topic MDX sources, extracted difficulty metadata for all 35 curriculum topics, and ran `go build` smoke checks. Docker was unavailable, so the runnable code was assessed statically + compiled, not run in containers.

---

## 1. Executive summary

**Verdict: the polished slice is excellent; the journey is incomplete.** The 6 interactive topics are, individually, near best-in-class — accurate, well-structured, and the new canvas simulations teach failure modes viscerally (overload, trip, stale reads). The single biggest weakness is **not quality but coverage**: the interactive experience is a 6-of-35 mid-band sample with **no beginner on-ramp and no advanced capstone**, so a learner can't actually travel "beginner → advanced" inside the interactive site.

**Top 5 findings**

1. **No interactive beginner on-ramp.** All 6 *foundational* topics (latency & throughput, scaling fundamentals, networking fundamentals, availability/SLAs, ACID vs BASE, monoliths vs microservices) live only in the old Zola site — none are in the interactive slice. Load Balancing/Caching assume those primitives. *(P0)*
2. **No interactive advanced capstone.** Zero advanced topics (real-world designs, leader election, CQRS, distributed locking) are in the slice. Advanced learners hit a ceiling. *(P1)*
3. **Interactive sims are the standout asset.** The 5 canvas flows + consistent-hashing ring genuinely build intuition; this is the differentiator vs every static competitor. *(strength)*
4. **Accessibility gaps in the sims.** State is encoded by colour alone (green/red/amber), no `prefers-reduced-motion` handling, and the canvases have no text alternative for screen readers. *(P1)*
5. **Two small content/sim inconsistencies.** Cache sim shows latencies in the 800–1600 ms range while the prose says "~1 ms / ~10 ms"; some sims start already at/over capacity. *(P2)*

### Scorecard (1–5)

| Dimension | Score | One-line justification |
|---|---|---|
| Content accuracy | 5 | Correct, deep, well-referenced (Release It!, Resilience4j, RFC 6585); worked examples. |
| Readability & clarity | 4 | Strong analogies + structure; prose is dense/sophisticated for a true beginner. |
| Usability & navigation | 4 | Clean IA, TOC, search, widened layout, good mobile; no cross-topic prereq links; sparse sidebar. |
| Interactivity & pedagogy | 5 | Canvas flows teach failure modes you can *feel*; guided hints, speed control. |
| Visual & accessibility | 3 | Great dark canvas + light/dark prose; colour-only state, no reduced-motion, no SR text alt. |
| DevEx | 5 | Code compiles clean; 130–482-line READMEs; guided `test.sh`; per-topic Docker; 57 unit tests pass. |
| Coverage & progression | 2 | Per-topic ramp is great; cross-curriculum the slice is 6/35 with 0 beginner-basics and 0 advanced. |

**Weighted takeaway:** quality ≈ 4.6/5, coverage ≈ 2/5. Fix coverage and this is a standout resource.

---

## 2. Persona journeys

### Priya — Beginner (CS student / bootcamp grad)
- **Lands well:** The restaurant-host analogy on Load Balancing is immediately graspable; "You'll learn" framing sets expectations; the live flow sim ("push the rate past capacity → drops climb") gives an *aha* without any prior theory. Badges (Beginner / Has code / Docker required) set the right expectation.
- **Frustrates:** She arrives mid-stream. The interactive site has no "what is latency / throughput / a server / horizontal scaling" page, so the first sentence of *Caching* ("a RAM read costs ~1 ms; a DB query ~10 ms") assumes numbers she hasn't met. The prose vocabulary ("idempotent", "quorum", "WAL LSN", "p99") is intermediate-grade. **She needs the 6 foundational basics rendered interactively before these.**
- **Verdict:** Delighted by the demos, occasionally out of her depth in the prose.

### Marcus — Intermediate (mid-level backend, interview prep)
- **Lands well:** This is the bullseye audience. Trade-off tables ("when to use / avoid"), the algorithm switcher on the LB sim (RR → Least-conn and watch p99 recover), and the rate-limiter token tank map directly to interview talking points. The Go implementations let him go from concept → running code.
- **Frustrates:** Wants the *next* topics that aren't here yet — message queues, sharding, API gateway, DNS — all exist in Zola but not interactively. No "prerequisites / next" links to navigate the learning path.
- **Verdict:** Strongest fit; wants more breadth at the same quality bar.

### Lena — Advanced (senior/staff)
- **Lands well:** Pleasantly surprised by the rigor — circuit-breaker tuning baselines (50% / 10-call window / 30 s / 3 probes, slow-call = p99×2–3), rate-limiting's sliding-window-counter formula with a worked example, atomic-Lua race fix, hot-key partitioning, fail-open vs fail-closed threat-model framing. Interview-question accordions are genuinely good.
- **Frustrates:** Hits a ceiling fast — no real-world design problems (Twitter feed, ride-sharing, video streaming), no leader election / consensus, no CQRS, no distributed locking. The sims are single-node mental models; she'd want multi-region/partition scenarios. State encoded only by colour bothers the a11y-minded reviewer.
- **Verdict:** Respects the depth; runs out of advanced material.

### Sam — Hands-on / DevEx
- **Lands well:** `go build` is clean across sampled modules (load-balancer, caching, rate-limiting, db-replication, latency — 5/5 OK on go 1.25.1). READMEs are substantial (130–482 lines). `test.sh` is a guided, press-Enter walkthrough that prints the curl commands it runs. `docker-compose.yml` per topic. The Astro side has 57 passing unit tests and a clean build.
- **Frustrates:** Everything needs Docker (no pure-`go run` quickstart called out for the impatient). `go build ./...` from repo root fails (go.work prefix) — you must `cd` into each module; a root `Makefile`/`just` target per topic would smooth this.
- **Verdict:** Above-average DevEx; a couple of quickstart papercuts.

---

## 3. Per-topic notes (the 6 interactive topics)

Each topic follows a strong, consistent internal arc: **analogy → "You'll learn" → what/why → interactive sim → how it works (diagrams) → trade-off tables → key concepts → when to use → failure modes → implementation → further reading → interview Qs.** That arc *is* a beginner→advanced ramp **within** a topic — the model to keep.

| Topic | Difficulty | Words | Strengths | Issues |
|---|---|---|---|---|
| Load Balancing | beginner | 1810 | Best on-ramp; LB flow sim + 4 algorithms; clean L4/L7 framing | Densest page (6 mermaids); algorithm-decision flowchart could precede the sim |
| Caching | beginner | 1289 | Cache-aside flow + hit-ratio intuition; eviction/observability tables | **Sim latency labelled "ms" (~800–1600) contradicts prose "~1 ms/~10 ms"** (`CacheFlow.tsx`) |
| Rate Limiting | intermediate | 1675 | Token tank sim; sliding-window-counter math; atomic Lua; headers/RFC 6585 | Two widgets (flow + token-bucket) back-to-back is slightly redundant |
| Circuit Breaker | intermediate | 1355 | Breaker flow shows downstream-load → 0 on trip; CB-vs-retry table; interview baselines | State by colour only; "fallback" concept could use one concrete example earlier |
| Database Replication | intermediate | 1257 | WAL-stream sim makes stale reads tangible; failover state machine | Densest concept load; replica-count slider max 4 only |
| Consistent Hashing | intermediate | 1804 | Bespoke interactive ring (add/remove node, vnodes, remap %) — the hero | Uses SVG ring (not canvas); different interaction model from the 5 flows |

No console errors on any page. Heading structure is healthy (7–10 H2, 3–6 H3 per page). All build-time mermaid diagrams render as inline SVG.

---

## 4. Beginner → advanced coverage matrix (all 35 topics)

Difficulty is from each topic's `[extra] difficulty` frontmatter. **"Astro"** = present in the interactive slice.

| Level | Total | In interactive slice | Coverage |
|---|---|---|---|
| Beginner | 9 | 2 (load-balancing, caching) | **22%** |
| Intermediate | 19 | 4 (rate-limiting, circuit-breaker, db-replication, consistent-hashing) | **21%** |
| Advanced | 7 | 0 | **0%** |
| **Total** | **35** | **6** | **17%** |

**Beginner (9)** — `latency-and-throughput`, `scaling-fundamentals`, `networking-fundamentals`, `availability-and-slas`, `monoliths-vs-microservices`, `acid-vs-base`*, `cdn`, `dns`, **load-balancing ✓**, **caching ✓**
*(acid-vs-base is tagged intermediate; grouped here as conceptual basics.)*
→ **Gap: none of the 6 true "fundamentals" are interactive.** These are the prerequisites the slice silently assumes.

**Intermediate (19)** — bloom-filter, geohashing, indexing, distributed-tracing, message-queues, service-discovery, cap-theorem, sql-vs-nosql, api-gateway, websockets-sse, url-shortener, retry-and-backoff, db-sharding, oauth-and-jwt, acid-vs-base, **consistent-hashing ✓**, **circuit-breaker ✓**, **db-replication ✓**, **rate-limiting ✓**
→ **Gap: high-value neighbours of existing topics missing** — message-queues, db-sharding, api-gateway, cap-theorem, retry-and-backoff (the natural "next click" from circuit-breaker).

**Advanced (7)** — `cqrs-and-event-sourcing`, `leader-election`, `distributed-locking`, `realtime-chat`, `ride-sharing`, `twitter-feed`, `video-streaming`
→ **Gap: zero advanced/real-world design content in the interactive site.** No capstone for senior learners.

**Progression assessment:** *Within* a topic the ramp is excellent. *Across* the curriculum the interactive site has no ramp — it starts at two beginner topics, sits in the intermediate band, and stops before advanced. The full Zola corpus does cover all 35 levels; the migration is mid-band first.

---

## 5. DevEx findings

- **Compiles clean:** 5/5 sampled modules build (`go build ./...`) on go 1.25.1 — load-balancer, caching, rate-limiting, db-replication, latency. Astro: 57 unit tests pass, `astro build` clean (8 pages).
- **Docs:** READMEs are substantial and consistent (130–482 lines); each topic ships `docker-compose.yml`, `nginx/`, `server/`, and a guided `test.sh` that echoes the commands it runs (good for learning).
- **Friction:**
  - Everything is Docker-first; no "just `go run` the server" quickstart for the impatient.
  - `go build ./...` from repo root fails (go.work `use` prefixes) — must `cd` per module. A per-topic `just build`/`just up` target would remove this.
  - Docker could not be exercised in this environment, so container-level reproducibility is unverified (code-level reproducibility is good).

---

## 6. Prioritised recommendations

### P0 — fixes the core gap
- **Add an interactive beginner on-ramp.** Migrate the 6 fundamentals (latency-and-throughput, scaling-fundamentals, networking-fundamentals, availability-and-slas, monoliths-vs-microservices, acid-vs-base) into the Astro slice, each with one small sim (e.g., a latency-ladder visual, a vertical-vs-horizontal scaling toggle). This is the prerequisite layer everything else assumes.
- **Add "Prerequisites" + "Next" links** to each topic (frontmatter-driven) so the slice reads as a *path*, not 6 islands.

### P1 — depth + reach
- **Add ≥2 advanced capstones** with sims — best ROI: `leader-election` (node failure → election animation, reuse the xyflow step pattern) and one real-world design (`url-shortener` or `twitter-feed`) that *composes* the primitives already built.
- **Fill the intermediate neighbours** learners click to next: `message-queues`, `db-sharding`, `api-gateway`, `retry-and-backoff`, `cap-theorem`.
- **Accessibility pass on the sims** (shared `useCanvasSim`):
  - Encode state with shape/label, not colour alone (e.g., ✓/✗ glyphs, "OPEN" text — already partly done on the breaker).
  - Respect `prefers-reduced-motion` (pause autorun / offer a static frame).
  - Add an off-screen text summary / `aria-label` describing live metrics so the sim isn't purely visual.

### P2 — polish
- **Fix the cache-sim latency units** — either scale the sim clock to show ~1 ms/~10 ms, or relabel the axis "relative latency" to match the prose (`CacheFlow.tsx`).
- **Tune sim defaults** so each opens in a legible steady state before the user pushes it to failure.
- **De-duplicate** the rate-limiting page's two stacked widgets (lead with the flow; make the token-bucket a "step-by-step" toggle).
- **Sidebar** will look sparse until more topics land — consider grouping or a "more topics on the classic site →" link in the interim.

---

## 7. Quick wins (≤1 hr each)
- Relabel cache-sim latency axis (one-line fix).
- Add `prefers-reduced-motion` guard in `useCanvasSim.ts` (pause autorun).
- Add ✓/✗/▲ glyphs alongside colour in the flow legends.
- Add `Prerequisites:` / `Next:` lines (plain links) to the 6 topic intros.
- Add a one-line "Run it: `docker compose up`, then `./test.sh`" callout atop each impl README's quickstart.

---

## Appendix — evidence
- Screenshots captured this session in the run scratchpad: full-page desktop per topic (`audit/<topic>-full.png`), mobile page + sim (`audit/mobile-*.png`), and sim stress states from the build session (`lbflow-overload.png`, `rlflow-overload.png`, `cbflow.png`, `replflow.png`, `cacheflow.png`, `ch-fixed-step*.png`).
- Content stats (Playwright `innerText`): 0 console errors across all 6 pages; 1257–1810 words/topic; 7–10 H2 / 3–6 H3 per page; 5–7 tables/topic.
- Coverage data: `[extra] difficulty` frontmatter across `website/content/hld/**/index.md` (35 topics).
- Build smoke: `go build ./...` on 5 sampled modules (all OK, go 1.25.1); `npm test` = 57 passing; `astro build` = 8 pages clean.
