````
/make-plan Redesign the learn-system-design topic page (astro-site `src/content/docs/hld/*.mdx`, 6 pages) plus onboarding. Current design failed a Dieter Rams audit at 14/30 with critical gaps in principles #2 useful, #6 honest, #10 as little design as possible.

Verdict paragraph:
> Score 14/30 is below the REFINE threshold, and the failures cluster in useful (#2) and honest (#6): the page structure hides the practice step, and two pages describe code that doesn't exist. The visual system, sim engines and syslings stay; the per-topic page template gets rebuilt around see → play → break → fix.

Why redesign and not refine: total 14/30 < 20, and the load-bearing #2 failure is structural (practice step at 89% page depth, present on 1/6 pages).

Preserve:
- Starlight theme + `src/styles/custom.css` tokens.
- Canvas sim engines `src/sim-lib/*-traffic.ts` + `src/components/useCanvasSim.ts` + `*Flow.tsx` a11y trio (reduced-motion pause, canvas aria-label, LiveSummary/SimLegend).
- syslings TUI (`syslings/`), exercises for load-balancing (7) and caching (2), `TopicNav.astro`.

Discard:
- 11-section page template with Lab as last optional section. Evidence: `load-balancing.mdx` Lab at y=12,222/13,813. Caused #2.
- Hand-written "This implementation" code not taken from the repo. Evidence: `caching.mdx` shows Redis cache-aside, repo `scalability/3.1_caching` is nginx proxy_cache; `rate-limiting.mdx` shows Go token bucket, repo uses nginx limit_req. Caused #6.
- Second LB simulator (card sim, y=3322) duplicating the canvas flow (y=1717); "old site" notes (`load-balancing.mdx:88`, `consistent-hashing.mdx:120`); duplicate "Has code"/"Runnable code" badges (`RunCode.astro:19`). Caused #10.

Top moves:
1. #6 honest: rewrite caching + rate-limiting "This implementation" from real confs; fix `caching.mdx:28` "RAM ~1 ms"; `LoadBalancerFlow.tsx:265-267` latency → `(rel) ×`.
2. #2 useful: page skeleton problem → sim → Break it (syslings block, first third) → go deeper; add caching's syslings block; landing cards show true Lab status.
3. #10 little design: one simulator per concept; remove history notes; one code badge.
4. #4 understandable: label sim counters (S0 31/41); README Quick Start → `just syslings`; `syslings --help`; fix `syslings/main.go:123` hard-coded topic list.
5. #9/#8: pause sims offscreen (IntersectionObserver in `useCanvasSim.ts`); `.lsd-hint` 4.49:1 and `.lsd-metric-value` 3.96:1 → ≥4.5:1 in dark theme.

Redesign principles in priority order:
1. #6 honest — every code sample on a page is copied from (or checked against) the repo path in its RunCode block.
2. #2 useful — learner reaches "break it" within 3 screens on every topic that has exercises.
3. #10 as little design as possible — one sim, one badge set, no process history in reader copy.

Deliverables:
- New page skeleton in `astro-site/AUTHORING.md`, applied to all 6 pages.
- States checklist per sim (paused, running, reduced-motion, offscreen).
- Check script: fail build if a "This implementation" snippet's identifiers are absent from its RunCode path.
- Cutover: all 6 pages migrated in one branch (`learning-platform`) before merge.

Anti-patterns: porting the old 11-section order under new headings; keeping both LB sims; adding topics before the 6 pages pass.
````
