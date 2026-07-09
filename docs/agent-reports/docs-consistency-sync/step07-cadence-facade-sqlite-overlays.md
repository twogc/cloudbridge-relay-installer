# step07 — Cadence Client facade, SQLite peer IPAM, k8s overlays skeleton

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## 1. Cadence Client ↔ ProductionClient

`NewClient` (ai-worker) now wraps `NewProductionClient` (relay engine):

| Public API | Behaviour |
|------------|-----------|
| `NewClient` | → `ProductionClient` + identity `cloudbridge-ai-worker` |
| `NewProductionClient` | unchanged for relay; optional `Identity`, ServiceOptions path |
| `Client` methods | delegate Start/Stop/Execute/… to `prod` |

GetDomain returns configured domain (not hardcoded `cloudbridge`).

## 2. Hybrid SQLite peer IPAM

| Env | Role |
|-----|------|
| `SCORE_SQLITE_PATH` | File path; empty/`:memory:` → shared in-memory DSN |

When SCORE hybrid is enabled and no Postgres `Repository` is passed, `OpenSQLitePeerIPAM` creates schema + `Repository` for peer IP allocation.

Files: `internal/database/sqlite_peer_ipam.go`, tests.

## 3. K8s overlays skeleton

```
k8s/overlays/production/kustomization.yaml  # resources → k8s-manifests/00–08
k8s/overlays/README.md
```

Not wired into `deploy.sh` (still CONSOLIDATION_PLAN step 4 future).

## Tests

```bash
go test -short ./internal/cadence/ -run 'TestNewClient|TestClient_|TestCreateWorkflow'  # OK
go test -short ./internal/database/ -run 'SQLite|OpenSQLite|Isolated'                   # OK
go test -short ./internal/p2p/ -run 'SCORE|NilRepository'                               # OK
go build ./cmd/relay/ ./cmd/ai-worker/                                                  # OK
```
