#!/usr/bin/env bash
# CI-friendly happy-path run of the whole lab. Exits non-zero on first failure.
set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
CHECK="$DIR/lab/check.sh"

cleanup() { (cd "$DIR" && docker compose down --remove-orphans >/dev/null 2>&1) || true; }
trap cleanup EXIT

echo "── lab verify: load-balancing ──"
(cd "$DIR" && docker compose up --build -d)
sleep 3
bash "$CHECK" 1
bash "$CHECK" 2
(cd "$DIR" && docker compose stop server2)
sleep 2
bash "$CHECK" 3
(cd "$DIR" && docker compose start server2)
# nginx passive health (fail_timeout=10s) holds server2 out for ~10s after restart
sleep 11
bash "$CHECK" 4
bash "$CHECK" 5
(cd "$DIR" && docker compose down)
bash "$CHECK" 6
echo "── load-balancing lab: ALL STEPS PASS ──"
