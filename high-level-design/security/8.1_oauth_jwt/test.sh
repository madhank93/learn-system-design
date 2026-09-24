#!/usr/bin/env bash
set -e

echo "=== Step 1: Get authorization code ==="
CODE=$(curl -s "http://localhost:8080/authorize?user=alice" | jq -r .code)
echo "Authorization code: $CODE"

echo ""
echo "=== Step 2: Exchange code for JWT ==="
TOKEN=$(curl -s "http://localhost:8080/token?code=$CODE" | jq -r .access_token)
echo "JWT: ${TOKEN:0:50}..."

echo ""
echo "=== Step 3: Access protected resource with token ==="
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8081/profile

echo ""
echo "=== Access without token (expect 401) ==="
curl -s http://localhost:8081/profile

echo ""
echo "=== Tampered token (expect 401) ==="
curl -s -H "Authorization: Bearer ${TOKEN}TAMPERED" http://localhost:8081/profile

echo ""
echo "=== Used code again (expect 401 — one-time use) ==="
curl -s "http://localhost:8080/token?code=$CODE"
