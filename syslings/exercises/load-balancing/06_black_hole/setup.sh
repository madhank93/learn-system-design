#!/usr/bin/env bash
# Inject the fault: a black-hole upstream entry + a patient-to-a-fault timeout.
source "$(dirname "$0")/../common.sh"

cp "$(dirname "$0")/round_robin.conf" "$SYSLINGS_WORK/round_robin.conf"
dc exec -T nginx-round-robin nginx -s reload >/dev/null 2>&1 || true
sleep 1
