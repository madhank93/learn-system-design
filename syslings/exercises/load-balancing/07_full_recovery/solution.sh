#!/usr/bin/env bash
# Scripted remediation used by `syslings verify` — the learner does this by hand.
source "$(dirname "$0")/../common.sh"
dc start server2 >/dev/null 2>&1
# nginx passive health (fail_timeout=10s) holds a failed server out for ~10s
sleep 11
