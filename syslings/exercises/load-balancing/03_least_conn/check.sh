#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$LC"
reload nginx-least-conn

# occupy two servers with slow requests, then fire quick ones
curl -sf --max-time 15 "$LC/?sleep=6s" >/dev/null 2>&1 &
curl -sf --max-time 15 "$LC/?sleep=6s" >/dev/null 2>&1 &
sleep 1
out="$(hits "$LC/" 9)"
wait
best=0
for i in 1 2 3; do
  c=$(count_server "$out" "$i"); [ "$c" -gt "$best" ] && best=$c
done
echo "  quick-request distribution: S1=$(count_server "$out" 1) S2=$(count_server "$out" 2) S3=$(count_server "$out" 3)"
[ "$best" -ge 6 ] || bad "two servers are busy, yet the quick requests spread evenly — this balancer isn't counting connections (max on one server = $best/9, want ≥6)" \
    "edit work/load-balancing/least_conn.conf — one directive makes nginx route by active connections"
ok "quick requests piled onto the idle server ($best/9) — that's least connections"
