````
/make-plan Refine the learn-system-design astro-site topic pages (`astro-site/src/content/docs/hld/*.mdx`, 6 pages) and landing, based on a Dieter Rams audit (total 20/30, up from 14/30 before the see → play → break redesign).

Verdict paragraph:
> Score 20/30 with no zeros clears the REFINE threshold: the see → play → break structure now holds on all six pages, so the remaining gaps are finish — light-theme rendering, one dishonest prose claim, and colored hint contrast.

Keep (already strong, do NOT touch in this pass):
- Principle #5 (unobtrusive) scored 3 — Evidence: one sim per page, badges = difficulty + "Docker required" + "⚙ Runnable code", no "old site"/"rustlings" strings in `src/content/docs/`. Regression check: `grep -rn "old site\|rustlings" astro-site/src/content/docs/` returns nothing; each page has exactly one `<canvas>` or one `.react-flow`.
- Page skeleton (AUTHORING.md "Canonical page structure"): Break it H2 stays within the first third of every page.
- `src/content/docs/hld/snippets.test.ts` stays green (`npm test` in astro-site).

Fix in priority order:
1. #3 aesthetic: canvas sims (`src/components/*Flow.tsx` via `canvasDraw.ts`, background `#0b0e14`) and mermaid diagrams render dark inside the light theme — make both follow Starlight's theme (read CSS custom properties / `data-theme`, mermaid theme per mode). Also fill or collapse the empty right half of the landing hero at 1440 px. Evidence: light-theme screenshots of circuit-breaker sim and LB mermaid diagram; landing at 1440.
2. #6 honest: `circuit-breaker.mdx:191` says the demo uses "a count-based window … the same state machine the simulations above model", and "How it works" says ">50% errors in the last 10 calls" and "lets one probe through". Code `high-level-design/resilience/5.1_circuit_breaker/proxy/main.go:50-66` trips on N consecutive failures (success resets `failures = 0`) and HALF_OPEN admits every request (`allow()` :44-45). Either rewrite the page copy to describe the code, or change the Go breaker to a sliding window + single probe (then update its test.sh).
3. #8 thorough: light theme `.lsd-hint strong` state words measure 1.43–3.3:1 (green "ok/hit/fresh" 1.81, yellow "stale" 1.43, orange 2.46, blue 2.42, red 3.3); legend glyphs 2.1–2.85:1; dark red hint 3.96:1. Give each state color a per-theme value in `src/styles/custom.css` (text ≥4.5:1, glyphs ≥3:1), the way `.lsd-metric-value` warning already does.
4. #2 useful: `db-replication.mdx` Break it lands at 3.4 screens (1440 px) — trim or move content above it to get ≤3.
5. #4 understandable: landing learning-loop copy says `just syslings` without explaining `just`; add one line (`brew install just`, or `go run ./syslings`).

Out of scope for this pass: new syslings exercises for circuit-breaker / rate-limiting / consistent-hashing / db-replication (next content step, separate plan); new topics; page-structure changes; bundle-size work.

Deliverables for the plan:
- Per-fix: target files, exact change, verification step (contrast re-measured in both themes; screenshots light + dark at 1440 and 375 px).
- All color token changes consolidated in `src/styles/custom.css`.
- Regression checklist for the Keep items above, plus AUTHORING.md "States to check before shipping" for every sim touched.

Anti-patterns to guard against:
- Adding a theming abstraction layer where reading existing Starlight CSS variables suffices.
- Restyling chrome, badges, or page order (#5 scored 3).
- Scope creep into writing new exercises or topics.
- Fixing the circuit-breaker claim by softening words without matching the code.
````
