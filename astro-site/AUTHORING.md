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
| Guided-lab section | **`<LabSteps steps={[…]} />`** | `components/LabSteps.astro` — Read → Sim → Run → **Lab** breakpoint. Each step: `{ title, run?, observe?, check?, expect? }`. Steps MUST mirror the topic's `high-level-design/<dir>/lab/check.sh` cases 1:1 — the `check` field is `just <topic>-lab-check <n>` and `expect` quotes the ✓ line the script prints. |

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
- [ ] Lab: `lab/check.sh` + `lab/verify.sh` + justfile recipes + `<LabSteps>` section (see "Adding a lab to a topic")

## Adding a lab to a topic

1. Design 4–6 steps: each one is *do something real* (kill a container, flood
   an endpoint) → *observe* → *scripted checkpoint*. First step is always
   `just <topic>-up`, last is `just <topic>-down`.
2. Write `high-level-design/<dir>/lab/check.sh` — source
   `lab-lib/check-helpers.sh`, one `stepN()` function per step, plus a `steps`
   listing. Every failure path calls `bad "<what>" "<hint>"`.
3. Verify BOTH states of every step: run the check before doing the step
   (must ✗ with a useful hint) and after (must ✓). No checkpoint ships
   verified in only one direction.
4. Write `high-level-design/<dir>/lab/verify.sh` — happy-path smoke driving
   up → all checks → down. Wire it into `lab-verify-all` in the justfile.
   Mind timing: nginx passive health checks hold a recovered server out for
   ~10s (`fail_timeout`), so post-recovery checks need an 11s wait.
5. Add justfile recipes `<topic>-lab` and `<topic>-lab-check`.
6. Add the `<LabSteps>` section to the mdx page (import from
   `../../../components/LabSteps.astro`), cross-referencing the sim where the
   lab repeats an experiment the sim previews.

## Gotchas (learned the hard way)

- **Light-theme text trap:** for canvas/hardcoded-dark chrome use fixed light hex (e.g. `#e6e9ef`), NOT `var(--sl-color-white)` — Starlight inverts it in light theme and the text vanishes. DOM panels on `.lsd-diagram` *do* use the `--sl-color-*` tokens.
- **Underscore files** (`_*.mdx`) are ignored by Astro content collections — fine for drafts, won't build a page.
- **`client:only="react"`** for anything touching `window`/canvas; `<Figure>`/mermaid stay server-rendered.
