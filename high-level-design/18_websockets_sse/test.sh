#!/usr/bin/env bash
set -e

echo "=== SSE — subscribe and receive 3 events ==="
echo "Subscribing to /events for 7 seconds (will receive heartbeats):"
timeout 7 curl -s -N http://localhost:8080/events || true

echo ""
echo "=== HTTP publish → SSE broadcast ==="
echo "Start SSE listener in background:"
curl -s -N http://localhost:8080/events &
SSE_PID=$!
sleep 1

echo "Publishing 3 messages:"
for i in 1 2 3; do
  curl -s "http://localhost:8080/publish?msg=hello-$i"
done
sleep 2
kill $SSE_PID 2>/dev/null || true

echo ""
echo "=== WebSocket echo (requires wscat — shown as curl equivalent) ==="
echo "WebSocket endpoint: ws://localhost:8080/ws"
echo "nginx proxy config ensures Upgrade headers are passed correctly."
echo "Test manually: npx wscat -c ws://localhost:8080/ws"
