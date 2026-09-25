#!/usr/bin/env bash
set -e

echo "=== Writes routed to primary ==="
for i in 1 2 3; do
  curl -s "http://localhost:8080/write?key=user-${i}&val=value-${i}"
done

echo ""
sleep 1
echo "=== Replication lag ==="
curl -s "http://localhost:8080/lag"

echo ""
echo "=== Reads load-balanced across replicas (note X-Source header) ==="
for i in $(seq 1 6); do
  out=$(curl -s -D - "http://localhost:8080/read?key=user-1")
  src=$(echo "$out" | grep -i "x-source" | awk '{print $2}' | tr -d '\r')
  body=$(echo "$out" | tail -1)
  echo "  read $i → source:$src | $body"
done
