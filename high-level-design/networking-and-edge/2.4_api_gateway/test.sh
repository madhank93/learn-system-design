#!/usr/bin/env bash
set -e

echo "=== 1. Path-based routing (port 8081) ==="
curl -s http://localhost:8081/api/users
curl -s http://localhost:8081/api/products
curl -s http://localhost:8081/api/unknown

echo ""
echo "=== 2. API key auth (port 8082) ==="
echo "No key (expect 401):"
curl -s http://localhost:8082/api/users
echo "Wrong key (expect 401):"
curl -s -H "X-Api-Key: bad-key" http://localhost:8082/api/users
echo "Valid key:"
curl -s -H "X-Api-Key: demo-key-123" http://localhost:8082/api/users

echo ""
echo "=== 3. Versioned routing (port 8083) ==="
echo "v1 (expect Deprecation header):"
curl -s -I http://localhost:8083/v1/users | grep -iE "x-api-version|deprecation|link"
echo "v2 (no deprecation):"
curl -s -I http://localhost:8083/v2/users | grep -i "x-api-version"
