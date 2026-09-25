#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RR"

out="$(hits "$RR/" 12)"
n=$(distinct_servers "$out")
echo "  12 requests reached $n distinct server(s): S1=$(count_server "$out" 1) S2=$(count_server "$out" 2) S3=$(count_server "$out" 3)"
[ "$n" -eq 3 ] \
  || bad "the pool is still short-handed ($n/3 servers taking traffic)" \
         "start server2 (see hint), wait ~10s for nginx's fail_timeout to expire, then re-check (press r)"
ok "all three servers back in rotation — full recovery 🎉"
ok "that ~10s wait you just felt is fail_timeout: eviction has a memory"
