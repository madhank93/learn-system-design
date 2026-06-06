#!/usr/bin/env bash
set -e

echo "=== 6 requests — round-robin across 3 healthy instances ==="
for i in $(seq 1 6); do
  curl -s http://localhost:8080/call
done

echo ""
echo "Consul UI: http://localhost:8500"

echo ""
echo "=== Stopping svc-3 to trigger health-check failure ==="
docker compose stop svc-3
echo "Waiting 20s for Consul to deregister svc-3..."
sleep 20
echo "6 more requests — only svc-1 and svc-2 should appear:"
for i in $(seq 1 6); do
  curl -s http://localhost:8080/call
done
