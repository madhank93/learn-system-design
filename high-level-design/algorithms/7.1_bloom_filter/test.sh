#!/usr/bin/env bash
set -e

echo "=== Bloom filter stats (empty) ==="
curl -s http://localhost:8080/stats

echo ""
echo "=== Adding 5 known items ==="
for item in apple banana cherry date elderberry; do
  curl -s "http://localhost:8080/add?item=$item"
done

echo ""
echo "=== Checking known items (expect: probably_yes) ==="
for item in apple banana cherry date elderberry; do
  curl -s "http://localhost:8080/check?item=$item"
done

echo ""
echo "=== Checking unknown items (expect: definitely_no or rare false positive) ==="
for item in fig grape honeydew jackfruit kiwi; do
  curl -s "http://localhost:8080/check?item=$item"
done

echo ""
echo "=== Stats after 5 inserts ==="
curl -s http://localhost:8080/stats
