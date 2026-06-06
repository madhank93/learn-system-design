#!/usr/bin/env bash
set -e

echo "=== Lock status (should be FREE) ==="
curl -s http://localhost:8081/status

echo ""
echo "=== Trigger worker-1 (acquires lock, works 1s) ==="
curl -s http://localhost:8081/run &
W1_PID=$!
sleep 0.2

echo "=== Lock status while worker-1 holds it ==="
curl -s http://localhost:8081/status

echo "=== Try worker-2 while lock held (expect 409 conflict) ==="
curl -s -w " HTTP:%{http_code}" http://localhost:8082/run
echo ""

echo "=== Try worker-3 while lock held (expect 409 conflict) ==="
curl -s -w " HTTP:%{http_code}" http://localhost:8083/run
echo ""

wait $W1_PID

echo ""
echo "=== Lock released — worker-2 can now acquire ==="
curl -s http://localhost:8082/run
