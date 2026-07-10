# Authoring guide — migrating & adding topics

This is the gold-standard recipe for a topic page. **Load Balancing (`src/content/docs/hld/load-balancing.mdx`) is the reference implementation** — when in doubt, copy it. Every topic follows the same structure so a migration is "fill in the blanks," not a redesign.

## Canonical page structure (section order)

1. **Frontmatter** — `title`, `description`.
2. **Badges** — difficulty + capabilities (one line of `<Badge>`).
3. **Analogy** — one plain-language paragraph (the "restaurant host" move). No jargon.
4. **"You'll learn"** — 3 bullets.
5. `## What is …?` — one tight definition paragraph.
6. `## Why it matters` — bullets, with numbers where possible.
7. `## How it works` — prose, then **the interactive hero sim** (`See it run.` lead-in), then build-time diagrams.
8. `### …` subsections — mechanisms, with `<Figure>`-wrapped mermaid diagrams + trade-off tables.
9. `## Key concepts` — glossary table.
10. `## When to use it` — two-column "use it when / avoid when" table.
11. `## Trade-offs and failure modes` — concern / issue / mitigation table.
12. `## This implementation` — **`<RunCode path="…" />`** linking the runnable Go/Docker code (makes the "Has code" badge clickable), then the core snippet.
13. `## Further reading` — 3–5 authoritative links.
14. `<details>` **Interview questions** — 5 Q&As with production baselines.
15. `<TopicNav>` — prev/next learning-path footer.

This internal arc *is* the beginner→advanced ramp (analogy → mechanics → trade-offs → failure modes → interview depth). Keep it.

## Frontmatter + badges

```mdx
---
title: Rate Limiting
description: One-sentence summary with the key sub-topics named.
---

import { Badge } from '@astrojs/starlight/components';
import Figure from '../../../components/Figure.astro';
import TopicNav from '../../../components/TopicNav.astro';
// + any interactive components this topic uses

<Badge text="Intermediate" variant="caution" /> <Badge text="Has code" variant="note" /> <Badge text="Docker required" variant="note" />
```

Difficulty → variant: **Beginner** = `success`, **Intermediate** = `caution`, **Advanced** = `danger`. Pull the difficulty from the source topic's `[extra] difficulty` frontmatter in `website/content/hld/**`.

## Component catalog

| Need | Use | Import |
|---|---|---|
| Static diagram (flowchart/sequence/state) | ` ```mermaid ` fenced block wrapped in `<Figure caption="…">` | `Figure` (`.astro`) — renders to inline SVG at build, zero JS |
| Stepped walk-through of a flow | `<SystemDiagram client:only="react" nodes={…} edges={…} steps={…} />` | `components/SystemDiagram` |
| Continuous animated simulation (the hero) | a `*Flow` component (see "Adding a sim") | `components/<Topic>Flow` |
| Callouts | `:::note` / `:::caution` / `:::tip` | Starlight built-in |
| Prev/next path | `<TopicNav prev={…} next={…} />` | `components/TopicNav.astro` |
| Link to runnable code | `<RunCode path="<category>/<dir>" recipe="<topic>" />` | `components/RunCode.astro` — links the GitHub source + shows `cd high-level-design` then `just <recipe>-up / -test / -down` |

**Rules**
- Static, non-interactive diagram → **mermaid fence in `<Figure>`** (never a runtime component). Build-time rendering keeps the page zero-JS.
- One **hero interactive** per topic, placed in `## How it works` under a **`See it run.`** lead-in. Keep the older step/detail widget below it if useful.
- Every diagram gets a `<Figure caption>` so it has consistent framing + a takeaway.

## Adding a new canvas simulation

Three layers (see `interactive-explainer-north-stars` / `canvas-flow-sim-pattern` and the existing 5 flows):

1. **Pure engine** — `src/sim-lib/<topic>-traffic.ts`: a time-stepped `step(dtMs)` model owning all state; seeded via `sim-lib/rng.ts` (`mulberry32`) so it's deterministic. Reuse the per-decision engine if one exists (`load-balancer.ts`, `token-bucket.ts`, `circuit-breaker.ts`). **Add `src/sim-lib/<topic>-traffic.test.ts`** (conservation, stress behavior, determinism).
2. **Renderer** — `src/components/<Topic>Flow.tsx`: maps engine state → canvas via the shared `useCanvasSim` hook and `canvasDraw.ts` helpers (`dot/box/label/roundRect/lerp/laneOf`).
3. **Controls** — reuse `simControls.tsx`: `Slider`, `Metric`, `SimLegend`, `LiveSummary`.

### Required accessibility trio (do not skip)

Every `*Flow` component must:
```tsx
const reduced = usePrefersReducedMotion();
const [playing, setPlaying] = useState(true);
useEffect(() => { if (reduced) setPlaying(false); }, [reduced]);   // 1. no autoplay on reduced-motion

const canvasRef = useCanvasSim({ /* … */ ariaLabel: '…describe the sim…' });  // 2. role=img + label

// 3. after <canvas/>: a screen-reader summary + a glyph-not-colour legend
<LiveSummary text={`…live metrics in words…`} />
<SimLegend items={[{ glyph: '●', color: '#…', label: '…' }, …]} />
```
State must be encoded by **glyph + colour**, never colour alone. Default sliders to a legible steady state (not already saturated). Latencies that aren't literal must be labelled relative (e.g., `p50 (rel) ×`), not `ms`.

## TopicNav wiring

Order the slice as a path; each page links prev/next:
```mdx
<TopicNav prev={{ label: 'Caching', slug: 'hld/caching' }} next={{ label: 'Circuit Breaker', slug: 'hld/circuit-breaker' }} />
```
Slug = Starlight content slug (`hld/<file>`). Also add the page to the sidebar in `astro.config.mjs`.

## Migration checklist (Zola `website/content/hld/**` → Astro)

- [ ] Create `src/content/docs/hld/<topic>.mdx`; copy frontmatter `title`/`description`; map `difficulty` → badge variant.
- [ ] Convert shortcodes: `{% badge_*() %}` → `<Badge>`, `{% note() %}` → `:::note`, alerts → `:::caution/:::tip`.
- [ ] Convert every ` ```mermaid ` block: keep as a fence, wrap in `<Figure caption="…">`. **Remove inline `;`** from sequence-diagram messages (mermaid treats `;` as a separator — it hard-errors at build).
- [ ] Add the hero interactive (reuse an existing `*Flow`, or build one per "Adding a sim").
- [ ] Keep the standard section order; ensure a `## Trade-offs and failure modes` table and an interview `<details>` exist.
- [ ] Add `<RunCode path="…" recipe="…" />` in `## This implementation`. `path` = current `high-level-design/<category>/<dir>` (verify it exists — dirs were renumbered); `recipe` = the justfile recipe base in `high-level-design/justfile` (the `<recipe>-up/-test/-down` set).
- [ ] Add `<TopicNav>` + register the page in `astro.config.mjs` sidebar.
- [ ] `npm test` (sim engines) and `npm run build` clean; spot-check light + dark + mobile.
- [ ] Lab: syslings exercise set under `syslings/exercises/<topic>/` + a `## Lab: fix the broken <thing>` section on the page (see "Adding syslings exercises to a topic")

## Adding syslings exercises to a topic

syslings (`syslings/`) is the repo's rustlings-style trainer: broken configs
mounted into the live Docker stack, file-watch re-checking, progressive hints.

1. Design 5–8 exercises, each teaching ONE failure: mostly `fix` type (ship a
   deliberately broken config the learner edits) plus a couple of `scenario`
   types (a `setup.sh` injects a live fault the learner must remediate).
2. Create `syslings/exercises/<topic>/`: `exercises.yaml` (ordered; see
   load-balancing for the schema), `compose.override.yml` (mounts
   `${SYSLINGS_WORK}/<file>` over the stack's pristine configs — compose merges
   volumes by container target path), and `common.sh` (source
   `lab-lib/check-helpers.sh`, define `dc`/`reload`).
3. Per exercise dir: broken template (`<file>`), `solution/<file>` (or
   `solution.sh` for scenarios — used by `syslings verify` to prove
   solvability in CI), `check.sh`, hints in `exercises.yaml`.
4. Register the topic in `syslings/main.go`'s `topics` map (compose dir +
   probe URL).
5. Verify BOTH states of every exercise: broken template must ✗ with a useful
   story, solution must ✓ (`syslings verify <topic>` covers the ✓ side).
6. Gotchas learned on the pilot:
   - `docker compose stop <svc>` removes its DNS entry — any later
     `nginx -s reload` then FAILS conf validation ("host not found in
     upstream"). Faults that must survive reloads should use a black-hole IP
     in the pool (e.g. `10.255.255.1`) instead of stopping a container.
   - nginx passive health holds a failed server out ~10s (`fail_timeout`);
     recovery checks need to account for it.
   - `reload` must surface `nginx -t` errors, or learners debug stale configs.
7. Add the `## Lab: fix the broken <thing>` section to the topic's mdx page:
   run commands, a short TUI transcript in a `text` fence, and the ordered
   exercise list. Cross-reference the sim where an exercise has a sim twin.
8. Add/extend justfile recipes if the topic needs them (`syslings` and
   `syslings-verify` take a topic argument already).

## Gotchas (learned the hard way)

- **Light-theme text trap:** for canvas/hardcoded-dark chrome use fixed light hex (e.g. `#e6e9ef`), NOT `var(--sl-color-white)` — Starlight inverts it in light theme and the text vanishes. DOM panels on `.lsd-diagram` *do* use the `--sl-color-*` tokens.
- **Underscore files** (`_*.mdx`) are ignored by Astro content collections — fine for drafts, won't build a page.
- **`client:only="react"`** for anything touching `window`/canvas; `<Figure>`/mermaid stay server-rendered.
