#!/usr/bin/env bash
# Start CloudBridge Relay for local client smoke tests.
# Usage: ./scripts/local-smoke/run-relay.sh [start|stop|status]
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
CFG="${ROOT}/configs/config.local-smoke.json"
PID_FILE="${SMOKE_PID_FILE:-/tmp/cloudbridge-smoke/relay.pid}"
LOG_FILE="${SMOKE_LOG_FILE:-/tmp/cloudbridge-smoke/relay.log}"
BIN="${RELAY_BIN:-${ROOT}/relay}"
mkdir -p "$(dirname "$PID_FILE")"

# P2P API JWT validation uses JWT_SECRET or default "test-secret"
export JWT_SECRET="${JWT_SECRET:-test-secret}"

cmd="${1:-start}"

case "$cmd" in
  stop)
    if [[ -f "$PID_FILE" ]]; then
      kill "$(cat "$PID_FILE")" 2>/dev/null || true
      rm -f "$PID_FILE"
      echo "relay stopped"
    else
      echo "no pid file"
    fi
    ;;
  status)
    if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
      echo "running pid=$(cat "$PID_FILE")"
      curl -sS -m 2 "http://127.0.0.1:5552/health" || true
      echo
    else
      echo "not running"
      exit 1
    fi
    ;;
  start)
    if [[ ! -x "$BIN" ]]; then
      echo "binary not found: $BIN (build with: make build && cp bin/cloudbridge-relay ./relay)"
      exit 1
    fi
    if [[ ! -f "$CFG" ]]; then
      echo "missing $CFG"
      exit 1
    fi
    if [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
      echo "already running pid=$(cat "$PID_FILE")"
      exit 0
    fi
    # Avoid host Prometheus TCP :9090 for http_api (config uses 19090).
    # QUIC still binds UDP 9090 (no conflict with smarty-prometheus TCP).
    nohup env JWT_SECRET="$JWT_SECRET" "$BIN" \
      -config "$CFG" -log-level info -tls=false \
      >"$LOG_FILE" 2>&1 &
    echo $! >"$PID_FILE"
    sleep 2
    if ! kill -0 "$(cat "$PID_FILE")" 2>/dev/null; then
      echo "failed to start; see $LOG_FILE"
      tail -40 "$LOG_FILE" || true
      exit 1
    fi
    echo "relay started pid=$(cat "$PID_FILE") log=$LOG_FILE"
    curl -sS -m 3 "http://127.0.0.1:5552/health" || true
    echo
    ;;
  *)
    echo "usage: $0 {start|stop|status}"
    exit 2
    ;;
esac
