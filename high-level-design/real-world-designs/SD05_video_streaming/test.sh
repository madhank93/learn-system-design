#!/usr/bin/env bash
set -euo pipefail

BASE="http://localhost:8080"
PASS=0; FAIL=0

check() {
  local desc="$1"; local expected="$2"; local actual="$3"
  if echo "$actual" | grep -q "$expected"; then
    echo "  PASS  $desc"
    ((PASS++)) || true
  else
    echo "  FAIL  $desc  (expected '$expected' in: $actual)"
    ((FAIL++)) || true
  fi
}

echo "=== SD05 Video Streaming ==="

# Upload video
R=$(curl -sf -X POST "$BASE/upload?title=BigBuckBunny&owner=alice&size=52428800")
check "POST /upload creates video" "id" "$R"
VIDEO_ID=$(echo "$R" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)

# Get manifest
R=$(curl -sf "$BASE/manifest/$VIDEO_ID?res=720p")
check "GET /manifest returns chunk URLs" "chunk_urls" "$R"

# Get first chunk
R=$(curl -sf "$BASE/chunk/$VIDEO_ID/0?res=720p")
check "GET /chunk/0 returns chunk data" "chunk-0-data" "$R"

# Record a view
R=$(curl -sf -X POST "$BASE/watch/$VIDEO_ID?user=bob")
check "POST /watch increments views" "views" "$R"

# Get video metadata
R=$(curl -sf "$BASE/video/$VIDEO_ID")
check "GET /video returns metadata" "BigBuckBunny" "$R"

# Feed
R=$(curl -sf "$BASE/feed")
check "GET /feed returns videos" "videos" "$R"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ "$FAIL" -eq 0 ]
