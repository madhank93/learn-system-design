# Guided Labs + Site Polish — Design

**Date**: 2026-07-08
**Scope**: The 6 topics already migrated to astro-site: load-balancing, caching, rate-limiting, circuit-breaker, db-replication, consistent-hashing.

## Goal

Adopt a rustlings-inspired hands-on learning loop for each topic, using the existing runnable Docker stacks rather than fill-in-the-code exercises:

**Read → Simulate (canvas sim) → Run (`just <topic>-up`) → Lab (guided break/observe/fix steps with scripted checkpoints)**

Alongside labs, complete a polish pass: content depth audit, diagram upgrades, site UX/navigation, and sim enhancements tied to lab steps.

## Decisions made during brainstorming

- **Lab style**: guided labs on running stacks, not broken-code exercises and not a TUI lab runner.
- **Verification**: scripted checkpoints via `just` recipes. Each step has a check script printing ✅/❌ with a diagnosis and hint. Objective, self-paced, offline-friendly.
- **Scope**: 6 migrated topics first; remaining 29 topics follow later using the proven template.
- **Sequencing**: infra first, then load-balancing as a full vertical pilot, then replicate across the other 5 topics.

## Phase 1 — Lab framework (shared infra)

### Check-script infrastructure

- `high-level-design/lab-lib/check-helpers.sh` — shared bash helpers: `probe`, `assert_http`, `assert_contains`, `pass`/`fail` output formatting with hints. Keeps per-topic check scripts tiny.
- Per topic: `high-level-design/<section>/<topic>/lab/check.sh` taking a step number argument.
- Justfile recipes per topic:
  - `<topic>-lab` — print lab steps + current status
  - `<topic>-lab-check <n>` — verify step n, exit non-zero on failure

Example output:

```
$ just load-balancing-lab-check 3

✓ server-2 stopped
✓ nginx evicted it (0/20 requests hit server-2)

$ just load-balancing-lab-check 4
✗ step 4 not done: still round_robin.conf

  hint: edit nginx config → least_conn.conf,
  then: just load-balancing-reload
```

### Site components

- `LabSteps.astro` — renders numbered lab steps: command to run, what to observe, checkpoint command, expected ✅ output. Lab renders as a section within each topic's mdx page (no separate lab page — read→lab stays one flow).
- Landing page becomes a guided learning path: per-topic card shows Read / Sim / Run / Lab affordances; prev/next navigation on every page.

### Cleanup & docs

- Retire Hugo `website/` — delete directory; README notes astro-site is the site. (Migration completeness already verified Jun 24.)
- AUTHORING.md gains a lab-authoring template and checklist.

## Phase 2 — Pilot: load balancing, full vertical

- **Content audit**: deepen failure modes (thundering herd on backend recovery, connection draining), L4 vs L7 trade-offs, sticky sessions, interview-framing sidebar.
- **Diagrams**: D2 statics where the sim doesn't cover — health-check state machine, L4 vs L7 request path. Follow the safe-palette constraint in `docs/diagrams/README.md`; verify dark-mode and mobile rendering.
- **Sim tie-in**: LoadBalancerFlow gains "kill backend" / "revive" controls so the sim mirrors lab step 3.
- **Lab (~6 steps)**:
  1. `just load-balancing-up`
  2. Verify round-robin distribution (checkpoint: even spread across backends)
  3. Kill server-2 (checkpoint: health check evicts it — 0/20 requests hit server-2)
  4. Swap to least_conn with a slow backend (checkpoint: slow backend gets fewer requests)
  5. Switch to consistent-hash config (checkpoint: same key always hits same backend)
  6. `just load-balancing-down`

  Each step: one `check.sh` case with pass/fail + hint.

## Phase 3 — Replicate to 5 topics

Same template, one "break something real" scenario each:

| Topic | Lab scenario |
|-------|-------------|
| Caching | Cold→warm hit ratio, then cache stampede + TTL fix |
| Rate limiting | Burst through limit, watch 429s, tune bucket |
| Circuit breaker | Kill upstream → trip open → half-open probe → recovery |
| DB replication | Write→replica lag observation → kill primary → failover |
| Consistent hashing | Add/remove node, checkpoint: minimal key movement |

Content/diagram/sim audit per topic same as the pilot, scaled to gaps found.

## Verification & quality gates

- Every check script is verified in both states during authoring: fails correctly before the step is done, passes after.
- `just lab-verify-all` — smoke recipe running each lab's happy path end-to-end (CI-friendly).
- Site build green; dark-mode and mobile screenshots for changed pages.

## Error handling

- Check scripts diagnose, not just fail: which sub-assertion failed, plus a hint for the next action. Non-zero exit codes so recipes compose in CI.
- Labs assume the stack is up; check scripts detect a down stack and say so explicitly rather than emitting confusing assertion failures.

## Out of scope (this pass)

- Remaining 29 HLD topics, low-level-design, real-world designs, distributed challenges.
- Site search (revisit after learning-path UX lands).
- A TUI/CLI lab runner.
