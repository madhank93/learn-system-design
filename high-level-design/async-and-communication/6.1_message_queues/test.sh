#!/usr/bin/env bash
set -e

echo "=== Work queue — competing consumers ==="
echo "Sending 6 tasks. Expect them split between consumer-work-1 and work-2:"
for i in $(seq 1 6); do
  curl -s "http://localhost:8080/work?msg=task-${i}"
done
echo ""
echo "Check distribution: docker compose logs consumer-work-1 consumer-work-2"

echo ""
echo "=== Fanout — pub/sub ==="
echo "Sending 3 events. Expect BOTH fanout-1 AND fanout-2 receive each:"
for i in $(seq 1 3); do
  curl -s "http://localhost:8080/event?msg=event-${i}"
done
echo ""
echo "Check fan-out: docker compose logs consumer-fanout-1 consumer-fanout-2"

echo ""
echo "RabbitMQ management UI: http://localhost:15672 (guest / guest)"
