#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080"
PASS=0; FAIL=0

check() {
  local desc="$1"; local expected="$2"; local actual="$3"
  if echo "$actual" | grep -q "$expected"; then
    echo "  PASS  $desc"
    ((PASS++)) || true
  else
    echo "  FAIL  $desc  (expected '$expected' in: $actual)"
    ((FAIL++)) || true
  fi
}

echo "=== SD04 Real-time Chat ==="

# List rooms (empty)
R=$(curl -sf "$BASE/rooms")
check "GET /rooms returns JSON" "rooms" "$R"

# Create a room
R=$(curl -sf -X POST "$BASE/rooms?name=general")
check "POST /rooms creates room" "general" "$R"

# History (empty)
R=$(curl -sf "$BASE/history/general")
check "GET /history/general" "messages" "$R"

# Presence (empty)
R=$(curl -sf "$BASE/presence/general")
check "GET /presence/general" "count" "$R"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
