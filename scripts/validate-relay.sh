#!/usr/bin/env bash
set -euo pipefail

HOST="${1:-}"
SNI="${2:-$HOST}"
TIMEOUT="${TIMEOUT:-5}"

if [[ -z "$HOST" ]]; then
	echo "Usage: $0 <host> [sni]" >&2
	exit 2
fi

say() { echo -e "[$(date +%H:%M:%S)] $*"; }
ok() { echo -e "✅ $*"; }
err() { echo -e "❌ $*"; }

# UDP 51820 (WireGuard)
say "Checking UDP 51820 (WireGuard) reachability to $HOST..."
if command -v nc >/dev/null 2>&1; then
	if nc -uzw "$TIMEOUT" "$HOST" 51820; then
		ok "UDP 51820 reachable"
	else
		err "UDP 51820 not reachable"
	fi
else
	echo "netcat not found; skipping UDP 51820 check" >&2
fi

# UDP 9090 (QUIC)
say "Checking UDP 9090 (QUIC) reachability to $HOST..."
if command -v nc >/dev/null 2>&1; then
	if nc -uzw "$TIMEOUT" "$HOST" 9090; then
		ok "UDP 9090 reachable"
	else
		err "UDP 9090 not reachable"
	fi
else
	echo "netcat not found; skipping UDP 9090 check" >&2
fi

# TCP 9090 TLS JSON hello (no HTTP expected)
say "Checking TCP 9090 TLS JSON hello to $HOST with SNI $SNI..."
HELLO_PAYLOAD='{"type":"hello"}\n'
TLS_OUT=$(printf '%b' "$HELLO_PAYLOAD" | openssl s_client -connect "$HOST:9090" -servername "$SNI" -tls1_3 -quiet 2>/dev/null || true)
FIRST_LINE=$(echo "$TLS_OUT" | head -n1)

if [[ -z "$TLS_OUT" ]]; then
	err "No response over TLS on 9090"
	exit 1
fi

if echo "$FIRST_LINE" | grep -qE '^HTTP/'; then
	err "Got HTTP on 9090, expected raw JSON over TLS"
	echo "$FIRST_LINE"
	exit 1
fi

echo "$TLS_OUT" | head -n 5

if echo "$TLS_OUT" | grep -q '"type":"hello"'; then
	ok "Received JSON hello from Relay"
else
	err "Did not detect JSON hello in response"
	exit 1
fi

ok "Validation completed"

