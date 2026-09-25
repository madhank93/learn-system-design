#!/usr/bin/env bash
set -e

echo "=== alice posts tweets ==="
curl -s "http://localhost:8080/tweet?user=alice&text=Hello+world"
curl -s "http://localhost:8080/tweet?user=alice&text=Learning+system+design"
curl -s "http://localhost:8080/tweet?user=alice&text=Distributed+systems+are+fun"

echo ""
echo "=== bob follows alice ==="
curl -s "http://localhost:8080/follow?from=bob&to=alice"

echo ""
echo "=== alice posts after bob follows ==="
curl -s "http://localhost:8080/tweet?user=alice&text=New+tweet+after+bob+followed"

echo ""
echo "=== bob timeline (should include alice's tweets) ==="
curl -s "http://localhost:8080/timeline/bob?limit=5"

echo ""
echo "=== carol timeline (not following alice — empty) ==="
curl -s "http://localhost:8080/timeline/carol?limit=5"

echo ""
echo "=== alice timeline (her own tweets) ==="
curl -s "http://localhost:8080/timeline/alice?limit=5"
