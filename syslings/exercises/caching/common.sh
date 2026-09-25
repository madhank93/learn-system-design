#!/usr/bin/env bash
# Shared preamble for caching exercise scripts.
set -uo pipefail

_EXDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"   # exercises/caching
_REPO="$(cd "$_EXDIR/../../.." && pwd)"                  # repo root
source "$_REPO/high-level-design/lab-lib/check-helpers.sh"

RAW=http://localhost:8081     # no cache — every request hits the backend
CACHE=http://localhost:8082   # proxy_cache
STALE=http://localhost:8083   # stale-serving cache

dc() { (cd "$SYSLINGS_TOPIC_DIR" && docker compose -f docker-compose.yml -f "$SYSLINGS_OVERRIDE" "$@"); }

reload() {
  local out
  if ! out="$(dc exec -T "$1" nginx -s reload 2>&1)"; then
    dc exec -T "$1" nginx -t 2>&1 | sed 's/^/  /'
    bad "nginx rejected the config — fix the syntax error above and save again"
  fi
  sleep 0.4
}

# cache_status <headers> — extracts the X-Cache-Status value.
cache_status() { printf '%s\n' "$1" | grep -i '^x-cache-status:' | awk '{print toupper($2)}' | tr -d '\r'; }

# generated <body> — extracts the generated:<ms> stamp.
generated() { printf '%s\n' "$1" | grep -o 'generated:[0-9]*' | head -1; }
