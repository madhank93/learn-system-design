# Verdict: REDESIGN (topic-page structure only)

Score 14/30 is below the REFINE threshold, and the failures cluster in useful (#2) and honest (#6): the page structure hides the practice step, and two pages describe code that doesn't exist. The visual system, sim engines and syslings stay; the per-topic page template gets rebuilt around see → play → break → fix.

Top moves:
1. **#6 honest** — rewrite `caching.mdx` and `rate-limiting.mdx` "This implementation" to show the real nginx `proxy_cache` / `limit_req` confs; fix "RAM ~1 ms"; LB sim latency → `(rel) ×`.
2. **#2 useful** — new page skeleton: problem → sim → "Break it" (syslings block, within first third) → go deeper; wire caching's existing exercises; landing cards reflect real Lab status.
3. **#10 little design** — one simulator per concept on LB (drop or merge the card sim at y=3322); remove "old site" notes; merge "Has code"/"Runnable code" badges.
4. **#4 understandable** — label sim counters; README Quick Start → `just syslings`; add `syslings --help`; fix `main.go:123`.
5. **#9 / #8** — pause sims offscreen (IntersectionObserver in `useCanvasSim.ts`); lift `.lsd-hint`/`.lsd-metric-value` to ≥4.5:1.
