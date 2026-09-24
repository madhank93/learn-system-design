#!/usr/bin/env bash
set -e

W=http://localhost:8081
R=http://localhost:8082

echo "=== Command side: send events ==="
curl -s -X POST $W/command -H "Content-Type: application/json" \
  -d '{"type":"user.created","user_id":"u1","payload":"alice"}'
curl -s -X POST $W/command -H "Content-Type: application/json" \
  -d '{"type":"user.updated","user_id":"u1","payload":"alice@example.com"}'
curl -s -X POST $W/command -H "Content-Type: application/json" \
  -d '{"type":"user.created","user_id":"u2","payload":"bob"}'

echo ""
echo "=== Event log (full history, append-only) ==="
curl -s $W/events

echo ""
sleep 1
echo "=== Query side: materialized views ==="
echo "User u1:"
curl -s $R/state/u1
echo "User u2:"
curl -s $R/state/u2
