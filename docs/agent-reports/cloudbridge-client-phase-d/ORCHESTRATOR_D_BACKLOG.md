# Phase D backlog (orchestrator, fact-based)

**Date:** 2026-07-09  
**MiMo phase-d report:** formal OK but **partial REJECT** — invented `Client.Connect` / ICE chain not present in `pkg/api/client.go` (only `DiscoverPeers`).

## Done

| ID | Status |
|----|--------|
| **D.1** | **ACCEPT** — `api.route_monitoring_enabled=false`; Manager skips `/relay/route` open/close spam; CONTRACT updated |

## Remaining (evidence from docs/code)

| ID | Real state | Next work |
|----|------------|-----------|
| D.2 | Discover LIVE (`GET …/peers`); ICE credentials **missing on relay** (CONTRACT); WG dual-base open | Feature-flag off ICE client probes if 404; doc N/A |
| D.3 | Defaults 30s / smoke 10s in scripts | Grep zero-interval warn paths; clamp interval ≥1s in Manager |
| D.4 | OIDC offline mock already fixed once (STATUS test log) | Document + ensure `go test ./pkg/auth -short` green |
| D.5 | gRPC Authenticate exists; local-smoke may use HMAC | Doc profile: local-smoke JWT HMAC vs OIDC |

## Verification already green (this session)

```
go test ./pkg/pathselect/ ./pkg/types/ ./pkg/config/ -count=1  PASS
go build ./cmd/cloudbridge-client  PASS
```

## Policy notes

- Combat agents: **mimo only**
- Accept mimo only if report file on disk + no invented paths
- Orchestrator implements code after review
