# Root justfile — repo-wide commands
# Usage: just <recipe>   e.g.  just diagrams   just diagrams 2.1_load_balancer

list:
    @just --list

# ── Diagrams ──────────────────────────────────────────────────────────────────
# Regenerate all D2 diagrams → SVG + GIF
# Requires: d2  (brew install d2)
# Output:   docs/diagrams/svg/   docs/diagrams/gif/
# Source:   docs/diagrams/src/

diagrams:
    @bash scripts/gen-diagrams.sh

# Regenerate one topic only: just diagrams-topic 2.1_load_balancer
diagrams-topic topic:
    @bash scripts/gen-diagrams.sh {{topic}}
