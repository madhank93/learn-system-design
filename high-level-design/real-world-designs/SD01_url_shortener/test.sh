#!/usr/bin/env bash
set -e

echo "=== Shorten URLs ==="
curl -s "http://localhost:8080/shorten?url=https://github.com/golang/go"
echo ""
CODE=$(curl -s "http://localhost:8080/shorten?url=https://go.dev/doc/" | jq -r .code)
echo "Code for https://go.dev/doc/ : $CODE"

echo ""
echo "=== Shorten same URL twice (idempotent) ==="
curl -s "http://localhost:8080/shorten?url=https://github.com/golang/go"

echo ""
echo "=== Redirect (302) ==="
curl -s -I "http://localhost:8080/$CODE" | grep -i location

echo ""
echo "=== Stats ==="
curl -s "http://localhost:8080/$CODE" -L > /dev/null
curl -s "http://localhost:8080/stats/$CODE"
