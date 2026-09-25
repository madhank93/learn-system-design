#!/usr/bin/env bash
# Lab checkpoints for 2.1 Load Balancing. Usage: check.sh <step 1-6|steps>
set -uo pipefail
source "$(cd "$(dirname "$0")/../../.." && pwd)/lab-lib/check-helpers.sh"

RR=http://localhost:8081
LC=http://localhost:8082
GEO=http://localhost:8083
CH=http://localhost:8084

steps() {
  cat <<'EOF'
Lab: Load Balancing — check any step with: just load-balancing-lab-check <n>
  1  Start the stack            just load-balancing-up
  2  Round robin spreads load   watch :8081 rotate across servers
  3  Kill server2, no errors    docker compose stop server2  (health check evicts it)
  4  Least-conn avoids the busy docker compose start server2, hold slow reqs on :8082
  5  Consistent hashing sticks  same cookie → same server on :8084
  6  Tear down                  just load-balancing-down
EOF
}

step1() {
  require_url "$RR"
  ok "round-robin LB answering on :8081"
  require_url "$LC" "least-conn LB missing — rebuild: just load-balancing-up"
  require_url "$GEO" "geo LB missing — rebuild: just load-balancing-up"
  require_url "$CH" "consistent-hash LB missing — rebuild: just load-balancing-up"
  ok "all four load balancers up (:8081–:8084)"
}

step2() {
  require_url "$RR"
  local out; out="$(hits "$RR/" 30)"
  local c1 c2 c3
  c1=$(count_server "$out" 1); c2=$(count_server "$out" 2); c3=$(count_server "$out" 3)
  echo "  distribution over 30 requests: S1=$c1 S2=$c2 S3=$c3"
  [ "$c1" -ge 8 ] && [ "$c2" -ge 8 ] && [ "$c3" -ge 8 ] \
    || bad "round robin should spread ~evenly (each ≥ 8 of 30)" \
           "is another lab step still active? restart clean: just load-balancing-down && just load-balancing-up"
  ok "each server took its turn — that's the rotating counter"
}

step3() {
  require_url "$RR"
  local out; out="$(hits "$RR/" 20)"
  local total; total=$(printf '%s\n' "$out" | grep -c 'Server' || true)
  local c2; c2=$(count_server "$out" 2)
  [ "$total" -eq 20 ] || bad "only $total/20 requests succeeded — nginx should retry a dead upstream on the next server" \
      "did you stop server2? from the topic dir: docker compose stop server2"
  [ "$c2" -eq 0 ] || bad "$c2/20 requests still hit Server 2 — it should be out of rotation" \
      "stop it: docker compose stop server2 (then wait ~2s and re-check)"
  ok "server2 stopped, 20/20 requests still succeeded"
  ok "nginx evicted it (0/20 requests hit Server 2)"
}

step4() {
  require_url "$LC"
  local out2; out2="$(get "$RR/")"
  printf '%s' "$out2" | grep -q 'Server' || bad "stack unhealthy" "restart: just load-balancing-up"
  # server2 must be back for a 3-server pool
  local probe; probe="$(hits "$RR/" 6)"
  [ "$(count_server "$probe" 2)" -ge 1 ] \
    || bad "Server 2 still down — this step needs all three servers" \
           "bring it back: docker compose start server2 (wait ~10s — nginx holds a failed server out for 10s)"
  ok "all three servers back in rotation"

  # occupy two servers with slow requests, then fire quick ones
  curl -sf --max-time 15 "$LC/?sleep=6s" >/dev/null 2>&1 &
  curl -sf --max-time 15 "$LC/?sleep=6s" >/dev/null 2>&1 &
  sleep 1
  local out; out="$(hits "$LC/" 9)"
  wait
  local best=0 i c
  for i in 1 2 3; do
    c=$(count_server "$out" "$i"); [ "$c" -gt "$best" ] && best=$c
  done
  echo "  quick-request distribution: S1=$(count_server "$out" 1) S2=$(count_server "$out" 2) S3=$(count_server "$out" 3)"
  [ "$best" -ge 6 ] || bad "least_conn should funnel quick requests to the idle server (max on one server ≥ 6 of 9, got $best)" \
      "make sure you're hitting :8082 (least_conn), not :8081 (round robin)"
  ok "quick requests avoided the two busy servers — that's least connections"
}

step5() {
  require_url "$CH"
  local alice bob
  alice="$(for i in 1 2 3 4 5; do curl -sf --max-time 5 --cookie session_id=alice "$CH/"; done)"
  [ "$(distinct_servers "$alice")" -eq 1 ] \
    || bad "alice's 5 requests hit multiple servers — hashing isn't sticking" \
           "hit :8084 with a session_id cookie: curl --cookie session_id=alice localhost:8084"
  ok "alice pinned to one server across 5 requests"
  bob="$(for i in 1 2 3; do curl -sf --max-time 5 --cookie session_id=bob "$CH/"; done)"
  [ "$(distinct_servers "$bob")" -eq 1 ] || bad "bob's requests bounced between servers"
  ok "bob pinned too (possibly a different server — that's the hash)"
}

step6() {
  if curl -sf --max-time 2 "$RR" >/dev/null 2>&1; then
    bad "stack still up on :8081" "tear it down: just load-balancing-down"
  fi
  ok "stack is down — lab complete 🎉"
}

case "${1:?usage: check.sh <step 1-6|steps>}" in
  steps) steps ;;
  1|2|3|4|5|6) "step$1" ;;
  *) echo "unknown step: $1"; steps; exit 2 ;;
esac
