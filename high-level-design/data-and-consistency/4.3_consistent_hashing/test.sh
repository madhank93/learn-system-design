#!/usr/bin/env bash
set -e

echo "=== Writing 12 keys via ring router ==="
for i in $(seq 1 12); do
  node=$(curl -s -D - "http://localhost:8080/set?key=user-${i}&val=v${i}" | grep -i "x-node" | awk '{print $2}' | tr -d '\r')
  echo "  user-$i → $node"
done

echo ""
echo "=== Keys per node (even distribution via vnodes) ==="
for port in 8091 8092 8093; do
  curl -s "http://localhost:${port}/keys"
done

echo ""
echo "=== Removing node3 — only its keys migrate ==="
curl -s "http://localhost:8080/ring/remove?node=node3:8080"
echo "Re-reading all keys (node3 keys now served by node1 or node2):"
for i in $(seq 1 12); do
  out=$(curl -s "http://localhost:8080/get?key=user-${i}" 2>/dev/null || echo "miss")
  echo "  user-$i → $out"
done
