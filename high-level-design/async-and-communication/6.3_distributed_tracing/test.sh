#!/usr/bin/env bash
set -e

echo "=== Sending 3 requests through frontend → order-service → inventory-service ==="
for item in widget gadget gizmo; do
  echo "  buy item=$item:"
  curl -s "http://localhost:8080/buy?item=$item"
done

echo ""
echo "Jaeger UI: http://localhost:16686"
echo "Search for service: 'frontend' to see full distributed traces"
echo "Each /buy request shows 3 spans: frontend → order-service → inventory-service"
