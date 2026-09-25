#!/usr/bin/env bash
set -e

echo "=== Register 5 drivers around San Francisco ==="
curl -s "http://localhost:8080/driver/location?id=driver-1&lat=37.7749&lng=-122.4194&status=available"
curl -s "http://localhost:8080/driver/location?id=driver-2&lat=37.7751&lng=-122.4180&status=available"
curl -s "http://localhost:8080/driver/location?id=driver-3&lat=37.7740&lng=-122.4210&status=busy"
curl -s "http://localhost:8080/driver/location?id=driver-4&lat=37.7800&lng=-122.4100&status=available"
curl -s "http://localhost:8080/driver/location?id=driver-5&lat=37.8044&lng=-122.2711&status=available"

echo ""
echo "=== Find drivers within 2km of downtown SF ==="
curl -s "http://localhost:8080/riders/nearby?lat=37.7749&lng=-122.4194&km=2"

echo ""
echo "=== Find drivers within 20km (includes Oakland driver) ==="
curl -s "http://localhost:8080/riders/nearby?lat=37.7749&lng=-122.4194&km=20"

echo ""
echo "=== Driver status ==="
curl -s "http://localhost:8080/driver/driver-1"

echo ""
echo "=== Stats ==="
curl -s "http://localhost:8080/stats"
