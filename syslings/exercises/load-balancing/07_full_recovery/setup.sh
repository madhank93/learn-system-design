#!/usr/bin/env bash
# The fault carries over: server2 is down. (Idempotent if it already is.)
source "$(dirname "$0")/../common.sh"
dc stop server2 >/dev/null 2>&1 || true
sleep 1
