# Root justfile — repo-wide commands
# Usage: just <recipe>   e.g.  just diagrams   just diagrams 2.1_load_balancer

list:
    @just --list

# ── syslings ──────────────────────────────────────────────────────────────────
# Rustlings-style hands-on trainer: opens a TUI exercise picker, mounts broken
# configs into the live Docker stack, re-checks on every save.

syslings topic="load-balancing":
    go run ./syslings {{topic}}

# CI: prove every exercise's solution passes (rebuilds the topic stack)
syslings-verify topic="load-balancing":
    go run ./syslings verify {{topic}}

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
