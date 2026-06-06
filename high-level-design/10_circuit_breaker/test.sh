#!/usr/bin/env bash
set -e

probe() {
  local url=$1
  code=$(curl -s -o /dev/null -w "%{http_code}" "$url/")
  status=$(curl -s "$url/cb/status")
  echo "$code | $status"
}

echo "Forcing backend into failing mode..."
curl -s http://localhost:8099/admin/fail

echo ""
echo "=== proxy-sensitive (threshold=2, port 8081) ==="
for i in $(seq 1 6); do echo "  req $i → $(probe http://localhost:8081)"; done

echo ""
echo "=== proxy-standard (threshold=3, port 8082) ==="
for i in $(seq 1 6); do echo "  req $i → $(probe http://localhost:8082)"; done

echo ""
echo "=== proxy-tolerant (threshold=5, port 8083) ==="
for i in $(seq 1 8); do echo "  req $i → $(probe http://localhost:8083)"; done

echo ""
echo "Recovering backend..."
curl -s http://localhost:8099/admin/recover
echo "Waiting 4s for sensitive circuit HALF_OPEN..."
sleep 4
echo "=== After recovery — sensitive proxy ==="
for i in 1 2 3; do echo "  req $i → $(probe http://localhost:8081)"; done
