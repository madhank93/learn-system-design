#!/usr/bin/env bash
set -e

BASE=http://localhost
ITEM=42

hit() {
  local url=$1
  local headers
  headers=$(mktemp)
  t=$(curl -s -D "$headers" -o /dev/null -w "%{time_total}" "$url")
  status=$(grep -i "x-cache-status" "$headers" | awk '{print $2}' | tr -d '\r' || true)
  rm -f "$headers"
  echo "${t}s [${status:--}]"
}

echo "=== 1. No cache (port 8081) — every request hits backend (~80ms) ==="
for i in $(seq 1 4); do
  echo "  req $i → $(hit "${BASE}:8081/item?id=${ITEM}")"
done

echo ""
echo "=== 2. Proxy cache (port 8082) — first MISS then HITs ==="
for i in $(seq 1 4); do
  echo "  req $i → $(hit "${BASE}:8082/item?id=${ITEM}")"
done

echo ""
echo "=== 3. Stale-while-revalidate (port 8083) — TTL=10s, background refresh ==="
for i in $(seq 1 3); do
  echo "  req $i → $(hit "${BASE}:8083/item?id=${ITEM}")"
done
echo "Waiting 11s for TTL to expire..."
sleep 11
echo "Post-expiry — stale served instantly, background revalidation:"
for i in 1 2; do
  echo "  req $i → $(hit "${BASE}:8083/item?id=${ITEM}")"
done
