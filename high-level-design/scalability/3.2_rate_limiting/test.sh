#!/usr/bin/env bash
set -e

echo "=== 1. Per-IP rate limit (port 8081) — 5 req/s, no burst ==="
echo "Sending 8 rapid requests — expect 1 OK then 429s (one slot per 200ms):"
for i in $(seq 1 8); do
  code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/)
  echo "  req $i → $code"
done

echo ""
echo "=== 2. Per-API-key rate limit (port 8082) ==="
echo "key-A (1 OK, then 429s):"
for i in $(seq 1 5); do
  code=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Api-Key: key-A" http://localhost:8082/)
  echo "  req $i → $code"
done
echo "key-B (separate quota — first request OK again):"
for i in $(seq 1 3); do
  code=$(curl -s -o /dev/null -w "%{http_code}" -H "X-Api-Key: key-B" http://localhost:8082/)
  echo "  req $i → $code"
done

echo ""
echo "=== 3. Burst mode (port 8083) — 2 req/s sustained, burst=10 ==="
echo "Sending 12 requests at once — 11 pass (1 + burst of 10), rest 429:"
for i in $(seq 1 12); do
  code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8083/)
  echo "  req $i → $code"
done
