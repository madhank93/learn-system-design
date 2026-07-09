#!/usr/bin/env bash
# Self-test for check-helpers.sh. Run: bash lab-lib/test_helpers.sh
set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
failures=0

expect() { # expect <desc> <want_exit> <cmd...>
  local desc="$1" want="$2"; shift 2
  local out; out="$("$@" 2>&1)"; local got=$?
  if [ "$got" -ne "$want" ]; then
    echo "FAIL: $desc (exit $got, want $want): $out"; failures=$((failures+1))
  else
    echo "ok:   $desc"
  fi
}

# subshell wrappers because fail/bad exit the shell
run_ok()   { bash -c "source '$DIR/check-helpers.sh'; ok 'looks good'"; }
run_bad()  { bash -c "source '$DIR/check-helpers.sh'; bad 'broke' 'try X'"; }
run_req_down() { bash -c "source '$DIR/check-helpers.sh'; require_url http://localhost:1 'start the stack'"; }
run_count() {
  bash -c "source '$DIR/check-helpers.sh'
out=\$(printf 'Server 1 (us)\nServer 2 (eu)\nServer 1 (us)\n' )
[ \"\$(count_server \"\$out\" 1)\" = 2 ] && [ \"\$(count_server \"\$out\" 2)\" = 1 ] && [ \"\$(count_server \"\$out\" 3)\" = 0 ]"
}

run_count_anchor() { # ids 10/11 must not inflate the count for id 1
  bash -c "source '$DIR/check-helpers.sh'
out=\$(printf 'Server 10 (us)\nServer 1 (eu)\nServer 11 (us)\n' )
[ \"\$(count_server \"\$out\" 1)\" = 1 ] && [ \"\$(count_server \"\$out\" 10)\" = 1 ]"
}
run_distinct() {
  bash -c "source '$DIR/check-helpers.sh'
out=\$(printf 'Server 10 (us)\nServer 1 (eu)\nServer 11 (us)\n' )
[ \"\$(distinct_servers \"\$out\")\" = 3 ]"
}
run_distinct_empty() { # bare call asserts the function itself exits 0 under set -e
  bash -c "set -euo pipefail; source '$DIR/check-helpers.sh'
distinct_servers '' >/dev/null
[ \"\$(distinct_servers '')\" = 0 ]"
}
run_get_dead() {
  bash -c "set -euo pipefail; source '$DIR/check-helpers.sh'; out=\$(get http://localhost:1/); [ -z \"\$out\" ]"
}
run_hits_dead() {
  bash -c "set -euo pipefail; source '$DIR/check-helpers.sh'; out=\$(hits http://localhost:1/ 3); [ -z \"\$out\" ]"
}

expect "ok prints and returns 0"           0 run_ok
expect "bad prints hint and exits 1"       1 run_bad
expect "require_url fails on dead port"    1 run_req_down
expect "count_server counts occurrences"   0 run_count
expect "count_server anchors exact id"     0 run_count_anchor
expect "distinct_servers counts ids"       0 run_distinct
expect "distinct_servers empty is 0"       0 run_distinct_empty
expect "get on dead url is safe (set -e)"  0 run_get_dead
expect "hits on dead url is safe (set -e)" 0 run_hits_dead

bash -c "source '$DIR/check-helpers.sh'; bad 'broke' 'try X'" 2>&1 | grep -q "hint: try X" \
  && echo "ok:   bad emits hint text" || { echo "FAIL: bad hint text"; failures=$((failures+1)); }

[ "$failures" -eq 0 ] && echo "ALL PASS" || { echo "$failures FAILURES"; exit 1; }
