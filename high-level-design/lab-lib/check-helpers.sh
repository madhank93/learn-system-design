#!/usr/bin/env bash
# Shared assertions for topic lab check scripts.
# Source me, then use: ok, bad, require_url, get, hits, count_server, distinct_servers
# Conventions: ok/bad print ✓/✗; bad prints an optional hint and exits 1.

ok()  { printf '  ✓ %s\n' "$1"; }

bad() { # bad <message> [hint]
  printf '  ✗ %s\n' "$1"
  if [ -n "${2:-}" ]; then printf '\n  hint: %s\n' "$2"; fi
  exit 1
}

require_url() { # require_url <url> [hint] — stack-up guard, friendly when down
  curl -sf --max-time 2 "$1" >/dev/null 2>&1 \
    || bad "cannot reach $1 — is the stack up?" "${2:-run: just load-balancing-up}"
}

get() { # get <url> — one request body (empty on failure)
  curl -sf --max-time 5 "$1" 2>/dev/null
}

hits() { # hits <url> <n> — n sequential request bodies, one per line
  local i
  for i in $(seq 1 "$2"); do curl -sf --max-time 5 "$1" 2>/dev/null; done
}

count_server() { # count_server <output> <id> — lines mentioning "Server <id>"
  printf '%s\n' "$1" | grep -c "Server $2" || true
}

distinct_servers() { # distinct_servers <output> — number of distinct server ids seen
  printf '%s\n' "$1" | grep -o 'Server [0-9]*' | sort -u | wc -l | tr -d ' '
}
