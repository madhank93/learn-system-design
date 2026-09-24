#!/usr/bin/env bash
set -e

echo "=== Geohash encode ==="
curl -s "http://localhost:8080/encode?lat=37.7749&lng=-122.4194&precision=6"
curl -s "http://localhost:8080/encode?lat=48.8566&lng=2.3522&precision=8"

echo ""
echo "=== Geohash decode ==="
curl -s "http://localhost:8080/decode?hash=9q8yy9"

echo ""
echo "=== Add locations to Redis GEO index ==="
curl -s "http://localhost:8080/location?id=driver-1&lat=37.7749&lng=-122.4194"
curl -s "http://localhost:8080/location?id=driver-2&lat=37.7751&lng=-122.4180"
curl -s "http://localhost:8080/location?id=driver-3&lat=37.8044&lng=-122.2711"

echo ""
echo "=== Find nearby within 2km of downtown SF ==="
curl -s "http://localhost:8080/nearby?lat=37.775&lng=-122.419&km=2"

echo ""
echo "=== Only driver-3 (Oakland, 14km away) excluded from 5km radius ==="
curl -s "http://localhost:8080/nearby?lat=37.775&lng=-122.419&km=5"
