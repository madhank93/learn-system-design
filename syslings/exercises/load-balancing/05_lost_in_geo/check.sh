#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$GEO"
reload nginx-geo-env

us="$(curl -sf --max-time 5 -H 'X-Region: us' "$GEO/")"
printf '%s' "$us" | grep -q 'Server 1 (us)' || bad "X-Region: us landed on the wrong pool: ${us:-no response}"
ok "us traffic → Server 1 (us)"

eu="$(curl -sf --max-time 5 -H 'X-Region: eu' "$GEO/")"
printf '%s' "$eu" | grep -q 'Server 2 (eu)' || bad "X-Region: eu landed on the wrong pool: ${eu:-no response}"
ok "eu traffic → Server 2 (eu)"

none="$(curl -sf --max-time 5 "$GEO/")"
printf '%s' "$none" | grep -q 'Server 3 (global)' \
  || bad "a request with NO region header landed on: ${none:-no response} — strangers shouldn't be dumped into one region's pool" \
         "check the map block's default in work/load-balancing/geo_env_based.conf"
ok "header-less traffic → Server 3 (global) — the catch-all pool"

jp="$(curl -sf --max-time 5 -H 'X-Region: jp' "$GEO/")"
printf '%s' "$jp" | grep -q 'Server 3 (global)' \
  || bad "an unknown region (jp) landed on: ${jp:-no response} — unknown regions belong on the global pool"
ok "unknown region (jp) → Server 3 (global)"
