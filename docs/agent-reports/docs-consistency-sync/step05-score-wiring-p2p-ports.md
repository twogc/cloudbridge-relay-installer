# step05 — SCORE allocator wiring + p2p-deployment ports

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Цель

1. Подключить canonical `NewSCORETenantNetworkAllocatorProduction` в runtime path.
2. Выровнять `cloudbridge-relay-p2p-deployment.yaml` с port-scheme.

## SCORE wiring (optional, env-gated)

| Env | Role |
|-----|------|
| `SCORE_BASE_URL` | Enable hybrid stack when non-empty |
| `SCORE_API_KEY` | Bearer for SCORE API |

| Component | Path |
|-----------|------|
| Factory | `internal/p2p/score_wiring.go` → `NewRelayManagerHybridFromSCORE` |
| Allocator | `network.NewSCORETenantNetworkAllocatorProduction` |
| Store | `network.NewSCORETenantNetworkStoreFixed` |
| Hybrid | `NewRelayManagerHybrid` (nil-safe repository) |
| Entry | `cmd/relay/p2p_integration.go` when P2P enabled |
| API | `P2PAPIServer.SetRelayManagerHybrid` → WireGuard endpoints |

Default (no env): behaviour unchanged — classic `RelayManager` only.

## p2p-deployment ports

| Port | Was | Now |
|------|-----|-----|
| HTTP API / health | 8080 | **9090** |
| P2P API | 8082 | **5552** |
| Metrics | — | **5551** |
| KEYCLOAK_* env | yes | `ZITADEL_FALLBACK_SECRET` |
| SCORE env | — | optional configMap/secret |

## Tests

```bash
go test -short ./internal/p2p/ -run 'SCORE|NilRepository'  # OK
go build ./cmd/relay/                                       # OK
go vet ./cmd/relay/                                         # OK (cancel leak already fixed)
```

## Не в scope

- Mandatory SCORE in all deploys
- SQLite repository wiring for peer IPAM
- Cadence dual-path merge
