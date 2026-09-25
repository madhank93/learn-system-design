#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RAW/item?id=1" "is the caching stack up?"
reload nginx-stale

id=$((RANDOM * 32768 + RANDOM))
prime="$(curl -sfi --max-time 5 "$STALE/item?id=$id")"
ok "primed the cache for item $id ($(cache_status "$prime"))"

sleep 4   # let the 3s validity expire

r="$(curl -sfi --max-time 5 "$STALE/item?id=$id")"
s=$(cache_status "$r")
case "$s" in
  STALE|UPDATING)
    ok "expired entry served as $s — the visitor got the old copy instantly"
    g0=$(generated "$prime"); g=$(generated "$r")
    [ "$g0" = "$g" ] && ok "same $g as the primed copy — refresh happened in the background, nobody waited" \
                     || ok "content already refreshed in the background ($g)"
    ;;
  *)
    bad "post-expiry request came back $s — that visitor waited for a synchronous backend refresh" \
        "serve the stale copy while refreshing behind the scenes: proxy_cache_use_stale + proxy_cache_background_update in work/caching/stale.conf"
    ;;
esac
