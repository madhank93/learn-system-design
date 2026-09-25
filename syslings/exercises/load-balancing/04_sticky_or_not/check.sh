#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$CH"
reload nginx-consistent-hash

# 1) a session must stick
alice="$(for i in 1 2 3 4 5; do curl -sf --max-time 5 --cookie session_id=alice "$CH/"; done)"
[ "$(distinct_servers "$alice")" -eq 1 ] \
  || bad "alice's 5 requests bounced between servers — sessions aren't sticking at all" \
         "the hash key must be stable per session: check the hash directive"
ok "alice pinned to one server across 5 requests"

# 2) DIFFERENT sessions must be able to land on different servers
sessions=""
for name in alice bob carol dave erin frank grace henry; do
  sessions+="$(curl -sf --max-time 5 --cookie "session_id=$name" "$CH/")"$'\n'
done
n=$(distinct_servers "$sessions")
echo "  8 distinct sessions spread across $n server(s)"
[ "$n" -ge 2 ] || bad "every session lands on the same server — the hash key is identical for all of them. From one machine, what value never changes between curls?" \
    "hash the session cookie, not the client address: hash \$cookie_session_id consistent;"
ok "different sessions spread across $n servers — keyed by cookie, not by IP"
