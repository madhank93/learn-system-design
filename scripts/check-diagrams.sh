#!/usr/bin/env bash
# check-diagrams.sh — validate rendered diagrams against the meta prompt gates
# Usage:
#   ./scripts/check-diagrams.sh                 # all topics
#   ./scripts/check-diagrams.sh 3.1_caching ... # specific topics
#
# Gates (docs/diagrams/README.md → meta prompt):
#   D1  size:        SVG viewBox width ≤ 1400, height ≤ 1600
#   R5  palette:     no sketch-hatching fills in .d2 sources
#   R2  density:     ≤ 8 leaf shapes per diagram (warning only)
#   G2  gif pacing:  ANIMATE_FLAGS interval = 4000

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC_DIR="$REPO_ROOT/docs/diagrams/src"
SVG_DIR="$REPO_ROOT/docs/diagrams/svg"

MAX_W=1400
MAX_H=1600
MAX_SHAPES=8
BANNED='style\.fill: "#(E67E22|F39C12|2ECC71|EC7063|AEB6BF|7FB3D3|F1C40F)"'

topics=("$@")
if (( ${#topics[@]} == 0 )); then
  while IFS= read -r -d '' d; do topics+=("$(basename "$d")"); done \
    < <(find "$SRC_DIR" -mindepth 1 -maxdepth 1 -type d -print0 | sort -z)
fi

errors=0
warnings=0

for topic in "${topics[@]}"; do
  echo "── $topic ──"

  # R5: banned fills in sources
  for src in "$SRC_DIR/$topic"/*.d2; do
    [[ -f "$src" ]] || continue
    if grep -qE "$BANNED" "$src"; then
      echo "  ERROR R5 banned fill: ${src#$REPO_ROOT/}"
      ((errors++))
    fi

    # R2: leaf shape count (shapes = lines declaring a node with a label,
    # excluding style/edge/keyword lines). Heuristic; warning only.
    local_count=$(grep -cE '^\s{0,4}[a-z_][a-z_0-9]*: ' "$src" | head -1)
    kw_count=$(grep -cE '^\s{0,4}(vars|direction|style|scenarios|steps|classes|grid-rows|grid-columns):' "$src")
    shapes=$(( local_count - kw_count ))
    if (( shapes > MAX_SHAPES + 2 )); then
      echo "  WARN  R2 density: ~${shapes} top-level decls in ${src#$REPO_ROOT/}"
      ((warnings++))
    fi
  done

  # D1: size gate on every rendered SVG
  while IFS= read -r -d '' svg; do
    read -r w h < <(head -c 400 "$svg" \
      | grep -o 'viewBox="[^"]*"' | head -1 \
      | sed 's/viewBox="//;s/"//' | awk '{print $3, $4}')
    w=${w%.*}; h=${h%.*}
    if [[ -n "${w:-}" && -n "${h:-}" ]] && (( w > MAX_W || h > MAX_H )); then
      echo "  ERROR D1 size ${w}x${h}: ${svg#$REPO_ROOT/}"
      ((errors++))
    fi
  done < <(find "$SVG_DIR/$topic" -name "*.svg" -print0 2>/dev/null)
done

# G2: animation interval
if ! grep -q 'animate-interval 4000' "$REPO_ROOT/scripts/gen-diagrams.sh"; then
  echo "ERROR G2: ANIMATE_FLAGS interval is not 4000"
  ((errors++))
fi

echo ""
echo "check-diagrams: $errors error(s), $warnings warning(s)"
(( errors == 0 ))
