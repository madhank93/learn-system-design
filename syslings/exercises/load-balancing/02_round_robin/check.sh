#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RR"
reload nginx-round-robin

out="$(hits "$RR/" 30)"
c1=$(count_server "$out" 1); c2=$(count_server "$out" 2); c3=$(count_server "$out" 3)
echo "  distribution over 30 requests: S1=$c1 S2=$c2 S3=$c3"
[ "$c1" -ge 8 ] && [ "$c2" -ge 8 ] && [ "$c3" -ge 8 ] \
  || bad "that's not a rotation — one server is hogging the traffic" \
         "edit work/load-balancing/round_robin.conf and make the pool fair again, then save"
ok "S1=$c1 S2=$c2 S3=$c3 — every server takes its turn"
