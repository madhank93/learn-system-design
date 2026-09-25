# Evidence

## Structure
- "Break it" H2 depth (1440 px): LB 19% / 2.8 screens, caching 27% / 2.5, rate-limiting 22% / 2.3, circuit-breaker 33% / 2.8, consistent-hashing 29% / 2.9, db-replication 33% / **3.4**. At 375 px: 3.4–4.6 screens. Baseline: 89% on 1/6 pages.
- Break it block on **6/6** pages; syslings exercises wired on 2/6 (load-balancing, caching); the other 4 offer the runnable stack only.
- One simulator per page (1 canvas; consistent-hashing uses 1 xyflow widget). Baseline: LB had 2.
- Badges per page: difficulty, "Docker required", "⚙ Runnable code" — no duplicates. Baseline: "Has code" + "Runnable code".
- No "old site" / "rustlings" strings left in `src/content/docs/`.
- LB page: 13,110 px, 2,125 words, 12 H2s (baseline 13,813 px, 3,318 words).

## Honesty
- `snippets.test.ts` verifies every source fence in "This implementation" against tracked repo files: passes 3/3 pages with fences.
- `circuit-breaker.mdx:191`: "count-based window … the same state machine the simulations above model". Code `resilience/5.1_circuit_breaker/proxy/main.go:50-66` trips on N **consecutive** failures (any success resets `failures = 0`) and HALF_OPEN admits all requests (`allow()` :44-45), while "How it works" and the sim describe ">50% errors in the last 10 calls" and "lets one probe through".
- `caching.mdx` RAM claim fixed; LB sim latencies relative; db-replication describes 500 ms polling (`replica/main.go:49`) honestly.
- consistent-hashing 150 vnodes matches `router/main.go:16`.

## Visual
- Canvas sims still hard-coded dark in light theme (screenshot `cb-light.png`).
- Mermaid diagrams render dark node fills with light text on white page in light theme (`lb-mermaid-light.png`).
- Landing hero at 1440 px: text column only, right half empty (`landing-light.png`).
- No horizontal overflow at 375 px on any page.

## Contrast (text, computed)
- Light theme, `.lsd-hint strong` state words: green "ok/hit/fresh" **1.81:1**, yellow "stale" **1.43:1**, orange "miss/HALF_OPEN/Writes" 2.46, blue "Reads" 2.42, red "fail/OPEN" 3.3.
- Light theme legend glyphs (non-text, need 3:1): green 2.1, orange 2.85, blue 2.8.
- Dark theme: red `.lsd-hint strong` 3.96:1 (circuit-breaker); mermaid edge labels 4.43:1.
- `.lsd-hint` body text and `.lsd-metric-value` now ≥4.5:1 (baseline fail fixed).

## Weight
- Landing: 6 requests, 95 KB JS. Topic pages: 15–19 requests, ~230 KB JS (up to ~405 KB on pages loading xyflow/mermaid chunks, cold).
- `useCanvasSim.ts` stops rAF offscreen (IntersectionObserver); reduced-motion defaults to paused; sims autoplay when on screen otherwise.

## Onboarding / syslings
- README Quick Start leads with `just syslings` (`README.md:19-25`).
- `syslings --help` prints usage, topics, Docker requirement; unknown topic error lists topics dynamically.
- Landing learning-loop copy uses `just syslings` without saying what `just` is or linking install.
