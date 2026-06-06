#!/usr/bin/env bash
set -e

echo "=== Hash router (port 8081) — deterministic, even distribution ==="
echo "Writing keys 1–9:"
for key in $(seq 1 9); do
  curl -s "http://localhost:8081/set?key=${key}&val=hash-val-${key}"
done
echo ""
echo "Reading back — same key always routes to same shard:"
for key in 1 5 9; do
  curl -s "http://localhost:8081/get?key=${key}"
done
echo ""
echo "Keys per shard:"
for port in 8091 8092 8093; do
  curl -s "http://localhost:${port}/keys"
done

echo ""
echo "=== Range router (port 8082) — predictable placement by numeric value ==="
echo "key 100 → shard1 (0–332), key 400 → shard2 (333–665), key 700 → shard3 (666+):"
for key in 100 400 700; do
  curl -s "http://localhost:8082/set?key=${key}&val=range-val-${key}"
done
for key in 100 400 700; do
  curl -s "http://localhost:8082/get?key=${key}"
done
