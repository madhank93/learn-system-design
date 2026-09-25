#!/usr/bin/env bash
set -e

echo "=== Backend at 60% fail rate — client retries with exponential backoff ==="
echo "Each /call triggers up to 5 retry attempts:"
for i in 1 2 3; do
  echo "  call $i:"
  curl -s http://localhost:8081/call
done

echo ""
echo "=== Setting fail rate to 0% — should succeed on first attempt ==="
curl -s "http://localhost:8080/admin/fail-rate?rate=0"
curl -s http://localhost:8081/call

echo ""
echo "=== Setting fail rate to 100% — all 5 attempts fail ==="
curl -s "http://localhost:8080/admin/fail-rate?rate=100"
curl -s http://localhost:8081/call
