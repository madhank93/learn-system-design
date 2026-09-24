#!/usr/bin/env bash
# gen-diagrams.sh — regenerate all D2 diagrams (SVG + optional GIF)
# Usage:
#   ./scripts/gen-diagrams.sh            # all diagrams
#   ./scripts/gen-diagrams.sh 2.1_load_balancer  # one topic
#
# Requires: d2 >= 0.6.0 on PATH
#   brew install d2
#
# Output:
#   docs/diagrams/svg/<topic>/<name>.svg
#   docs/diagrams/gif/<topic>/<name>.gif  (only for files with a .gif.d2 companion or listed in ANIMATED)
#
# Render defaults: --sketch --layout elk
# Override per-file by adding a header comment like:
#   # d2-flags: --layout dagre

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$REPO_ROOT/docs/diagrams/src"
SVG_DIR="$REPO_ROOT/docs/diagrams/svg"
GIF_DIR="$REPO_ROOT/docs/diagrams/gif"

D2_FLAGS="--sketch --layout elk --pad 40"
ANIMATE_FLAGS="--sketch --layout elk --pad 40 --animate-interval 4000"

# Size gate (meta prompt rule D1): rendered SVGs must fit the page.
MAX_W=${D2_MAX_W:-1400}
MAX_H=${D2_MAX_H:-1600}

# Files to also render as animated GIF (relative to src/<topic>/)
# Format: "<topic>/<file>"
ANIMATED=(
  "2.1_load_balancer/request-flow.d2"
  "2.1_load_balancer/health-checks.d2"
  "3.1_caching/request-flow.d2"
  "3.4_db_replication/request-flow.d2"
  "4.5_leader_election/request-flow.d2"
  "5.1_circuit_breaker/request-flow.d2"
  "5.2_retry_backoff/request-flow.d2"
  "SD02_twitter_feed/request-flow.d2"
)

# ─────────────────────────────────────────────────────────────

check_d2() {
  if ! command -v d2 &>/dev/null; then
    echo "ERROR: d2 not found. Install with: brew install d2"
    exit 1
  fi
  echo "d2 $(d2 --version)"
}

is_animated() {
  local rel="$1"  # e.g. "2.1_load_balancer/overview.d2"
  for entry in "${ANIMATED[@]}"; do
    [[ "$entry" == "$rel" ]] && return 0
  done
  return 1
}

render_topic() {
  local topic="$1"
  local topic_src="$SRC_DIR/$topic"
  local topic_svg="$SVG_DIR/$topic"
  local topic_gif="$GIF_DIR/$topic"

  [[ -d "$topic_src" ]] || { echo "SKIP: $topic_src not found"; return; }
  mkdir -p "$topic_svg" "$topic_gif"

  local pass=0 fail=0

  while IFS= read -r -d '' src_file; do
    local name
    name="$(basename "$src_file" .d2)"
    local rel="${topic}/${name}.d2"

    # Banned-fill gate (meta prompt rule R5): sketch mode hatches these fills
    if grep -qE 'style\.fill: "#(E67E22|F39C12|2ECC71|EC7063|AEB6BF|7FB3D3|F1C40F)"' "$src_file"; then
      echo "  FAIL (banned fill) $rel — see docs/diagrams/README.md palette"
      ((fail++)) || true
      continue
    fi

    # SVG
    local svg_out="$topic_svg/${name}.svg"
    echo "  SVG  $rel → $(basename "$svg_out")"
    if d2 $D2_FLAGS "$src_file" "$svg_out" 2>&1; then
      ((pass++)) || true
    else
      echo "  FAIL $rel"
      ((fail++)) || true
      continue
    fi

    # Size gate (rule D1) on every produced SVG for this diagram
    # (a diagram may emit <name>.svg or <name>/{index,scenario,steps/N}.svg)
    local gate_fail=0
    while IFS= read -r -d '' produced; do
      read -r w h < <(head -c 400 "$produced" \
        | grep -o 'viewBox="[^"]*"' | head -1 \
        | sed 's/viewBox="//;s/"//' | awk '{print $3, $4}')
      w=${w%.*}; h=${h%.*}
      if [[ -n "$w" && -n "$h" ]] && (( w > MAX_W || h > MAX_H )); then
        echo "  FAIL (size ${w}x${h} > ${MAX_W}x${MAX_H}) ${produced#$REPO_ROOT/}"
        gate_fail=1
      fi
    done < <({ [[ -f "$svg_out" ]] && printf '%s\0' "$svg_out"; \
               [[ -d "$topic_svg/$name" ]] && find "$topic_svg/$name" -name "*.svg" -print0; } 2>/dev/null)
    if (( gate_fail )); then
      ((fail++)) || true
    fi

    # GIF (if in ANIMATED list)
    if is_animated "$rel"; then
      local gif_out="$topic_gif/${name}.gif"
      echo "  GIF  $rel → $(basename "$gif_out")"
      if d2 $ANIMATE_FLAGS "$src_file" "$gif_out" 2>&1; then
        ((pass++)) || true
      else
        echo "  FAIL (gif) $rel"
        ((fail++)) || true
      fi
    fi

  done < <(find "$topic_src" -name "*.d2" -print0 | sort -z)

  echo "  $topic: $pass ok, $fail failed"
}

# ─────────────────────────────────────────────────────────────

check_d2

if [[ $# -gt 0 ]]; then
  # Render a specific topic
  for topic in "$@"; do
    echo "── $topic ──"
    render_topic "$topic"
  done
else
  # Render all topics
  echo "Rendering all diagram topics in $SRC_DIR"
  echo ""
  total_pass=0 total_fail=0

  while IFS= read -r -d '' topic_dir; do
    topic="$(basename "$topic_dir")"
    echo "── $topic ──"
    render_topic "$topic"
    echo ""
  done < <(find "$SRC_DIR" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)

  echo "Done."
fi
