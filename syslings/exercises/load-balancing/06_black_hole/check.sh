#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RR"
reload nginx-round-robin

out="$(hits "$RR/" 20)"
total=$(printf '%s\n' "$out" | grep -c 'Server' || true)
echo "  $total/20 requests answered within the 5s budget"
[ "$total" -eq 20 ] \
  || bad "$((20 - total))/20 requests hung — they're stuck connecting to the black-hole entry, and nginx is willing to wait a full minute before failing over" \
         "make nginx give up fast: proxy_connect_timeout in work/load-balancing/round_robin.conf (removing the dead entry also works — but timeouts protect you from the NEXT dead entry too)"
ok "20/20 requests succeeded with a black hole in the pool"
ok "first request paid the short timeout, then nginx evicted the dead entry — clients never noticed"
