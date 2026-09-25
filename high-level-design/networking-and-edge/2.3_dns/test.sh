#!/usr/bin/env bash
set -e

echo "=== DNS resolution ==="
echo "'web-service' — 2 IPs (round-robin via Docker network aliases):"
for i in 1 2 3; do
  curl -s "http://localhost:8080/resolve?host=web-service"
done

echo ""
echo "'db' — 1 IP (single container):"
curl -s "http://localhost:8080/resolve?host=db"

echo ""
echo "=== Round-robin in action ==="
echo "Calling 'web-service' 6 times — should alternate between web-1 and web-2:"
for i in $(seq 1 6); do
  curl -s "http://localhost:8080/call?host=web-service"
done
