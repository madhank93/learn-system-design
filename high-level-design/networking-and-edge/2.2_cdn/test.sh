#!/usr/bin/env bash
set -e

hit() {
  local headers
  headers=$(mktemp)
  t=$(curl -s -D "$headers" -o /dev/null -w "%{time_total}" "$1")
  status=$(grep -i "x-cache-status" "$headers" | awk '{print $2}' | tr -d '\r' || true)
  rm -f "$headers"
  echo "${t}s [${status:--}]"
}

echo "=== 1. No cache (port 8081) — every request hits origin ==="
for i in 1 2 3; do echo "  static req $i → $(hit http://localhost:8081/static/logo.png)"; done

echo ""
echo "=== 2. Edge cache (port 8082) — first MISS then HITs ==="
for i in 1 2 3; do echo "  static req $i → $(hit http://localhost:8082/static/logo.png)"; done
echo "Dynamic path (Cache-Control: no-store — never cached):"
for i in 1 2 3; do echo "  api req $i → $(hit http://localhost:8082/api/feed)"; done

echo ""
echo "=== 3. Selective cache (port 8083) — /static cached, /api bypassed ==="
echo "Static:"
for i in 1 2 3; do echo "  static req $i → $(hit http://localhost:8083/static/logo.png)"; done
echo "Dynamic:"
for i in 1 2 3; do echo "  api req $i → $(hit http://localhost:8083/api/feed)"; done
