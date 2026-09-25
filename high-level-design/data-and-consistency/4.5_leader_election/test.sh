#!/usr/bin/env bash
set -e

echo "=== Initial state — one node is leader ==="
sleep 3
for port in 8081 8082 8083; do
  echo "  node :$port → $(curl -s http://localhost:$port/status)"
done

echo ""
echo "=== Only one node should be LEADER ==="
leaders=$(for port in 8081 8082 8083; do curl -s http://localhost:$port/status; done | grep -c "LEADER")
echo "Leaders found: $leaders (expect: 1)"

echo ""
echo "=== Stop the leader container to trigger re-election ==="
leader_port=""
for port in 8081 8082 8083; do
  if curl -s http://localhost:$port/status | grep -q "LEADER"; then
    leader_port=$port
    break
  fi
done
echo "Current leader is on port $leader_port"
svc="node-$(( (leader_port - 8080) ))"
docker compose stop $svc
echo "Stopped $svc. Waiting 8s for re-election..."
sleep 8
echo "New leader:"
for port in 8081 8082 8083; do
  curl -s http://localhost:$port/status 2>/dev/null || true
done
