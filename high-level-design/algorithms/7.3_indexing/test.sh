#!/usr/bin/env bash
set -e

echo "=== Insert 10 keys ==="
for i in 10 50 20 80 30 70 40 90 60 100; do
  curl -s "http://localhost:8080/insert?key=$i&val=value-$i"
done

echo ""
echo "=== Point search (O(log n)) ==="
curl -s "http://localhost:8080/search?key=50"
curl -s "http://localhost:8080/search?key=999"

echo ""
echo "=== Range search [30, 70] — returns all keys in range ==="
curl -s "http://localhost:8080/range?lo=30&hi=70"
