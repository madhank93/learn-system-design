#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RAW/item?id=1" "is the caching stack up?"
reload nginx-proxy-cache

id=$((RANDOM * 32768 + RANDOM))   # fresh key so old cache entries can't help
r1="$(curl -sfi --max-time 5 "$CACHE/item?id=$id")"
s1=$(cache_status "$r1")
[ "$s1" = "MISS" ] || ok "first request came back $s1 (expected MISS on a cold key — did a previous run warm it?)"
ok "first request for item $id filled the cache ($s1)"

r2="$(curl -sfi --max-time 5 "$CACHE/item?id=$id")"
s2=$(cache_status "$r2")
[ "$s2" = "HIT" ] \
  || bad "second request for the SAME item came back $s2 — the cache stored the first response but can never find it again" \
         "two identical requests must produce the same cache key: look at proxy_cache_key in work/caching/proxy_cache.conf"
ok "second request was a HIT — served from cache, backend untouched"

g1=$(generated "$r1"); g2=$(generated "$r2")
[ "$g1" = "$g2" ] || bad "HIT but the body was regenerated ($g1 vs $g2) — that's not a cache, that's a lie"
ok "same $g1 both times — one backend render served twice"
