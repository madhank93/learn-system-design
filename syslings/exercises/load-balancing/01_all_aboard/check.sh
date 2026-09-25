#!/usr/bin/env bash
source "$(dirname "$0")/../common.sh"

require_url "$RR" "start the stack: syslings brings it up automatically — is Docker running?"
ok "round robin       :8081  (requests rotate across the pool)"
require_url "$LC" "the least-conn balancer isn't answering"
ok "least connections :8082  (requests chase the idle server)"
require_url "$GEO" "the geo balancer isn't answering"
ok "geo routing       :8083  (X-Region header picks the pool)"
require_url "$CH" "the consistent-hash balancer isn't answering"
ok "consistent hash   :8084  (same cookie, same server)"
ok "four balancers, three Go servers, one lesson each — onward!"
