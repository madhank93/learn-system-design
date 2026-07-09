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

expect "ok prints and returns 0"          0 run_ok
expect "bad prints hint and exits 1"      1 run_bad
expect "require_url fails on dead port"   1 run_req_down
expect "count_server counts occurrences"  0 run_count

bash -c "source '$DIR/check-helpers.sh'; bad 'broke' 'try X'" 2>&1 | grep -q "hint: try X" \
  && echo "ok:   bad emits hint text" || { echo "FAIL: bad hint text"; failures=$((failures+1)); }

[ "$failures" -eq 0 ] && echo "ALL PASS" || { echo "$failures FAILURES"; exit 1; }
