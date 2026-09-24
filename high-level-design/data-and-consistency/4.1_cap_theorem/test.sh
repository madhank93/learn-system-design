#!/usr/bin/env bash
set -e

echo "=== CP nodes (ports 8081/8082) — Consistency over Availability ==="

echo "Normal operation — write to cp-1, read from cp-2:"
curl -s "http://localhost:8081/set?key=user&val=alice"
sleep 0.3
curl -s "http://localhost:8082/get?key=user"

echo ""
echo "Enabling partition on cp-1..."
curl -s http://localhost:8081/admin/partition

echo "Write during partition (expect 503 — CP refuses):"
curl -s "http://localhost:8081/set?key=user&val=bob"

echo ""
echo "Restoring network on cp-1..."
curl -s http://localhost:8081/admin/partition

echo ""
echo "=== AP nodes (ports 8083/8084) — Availability over Consistency ==="

echo "Normal operation — write to ap-1, read from ap-2:"
curl -s "http://localhost:8083/set?key=user&val=alice"
sleep 0.3
curl -s "http://localhost:8084/get?key=user"

echo ""
echo "Enabling partition on ap-1..."
curl -s http://localhost:8083/admin/partition

echo "Write during partition (expect 200 — AP accepts, stale on ap-2):"
curl -s "http://localhost:8083/set?key=user&val=bob_during_partition"

echo "Read from ap-2 (stale — still shows alice):"
curl -s "http://localhost:8084/get?key=user"

echo ""
echo "Restoring partition..."
curl -s http://localhost:8083/admin/partition
echo "After restore, ap-1 and ap-2 eventually converge on next write."
