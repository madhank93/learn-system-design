#!/usr/bin/env bash
# Shared preamble for load-balancing exercise scripts. Source from a script
# in an exercise dir: source "$(dirname "$0")/../common.sh"
set -uo pipefail

_EXDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"   # exercises/load-balancing
_REPO="$(cd "$_EXDIR/../../.." && pwd)"                  # repo root
source "$_REPO/high-level-design/lab-lib/check-helpers.sh"

RR=http://localhost:8081
LC=http://localhost:8082
GEO=http://localhost:8083
CH=http://localhost:8084

# dc — compose against the topic stack WITH the exercise override.
dc() { (cd "$SYSLINGS_TOPIC_DIR" && docker compose -f docker-compose.yml -f "$SYSLINGS_OVERRIDE" "$@"); }

# reload <service> — apply the learner's saved conf without a restart.
# nginx keeps the old conf if the new one is invalid; surface that instead of
# checking stale behavior.
reload() {
  local out
  if ! out="$(dc exec -T "$1" nginx -s reload 2>&1)"; then
    dc exec -T "$1" nginx -t 2>&1 | sed 's/^/  /'
    bad "nginx rejected the config — fix the syntax error above and save again"
  fi
  sleep 0.4
}
