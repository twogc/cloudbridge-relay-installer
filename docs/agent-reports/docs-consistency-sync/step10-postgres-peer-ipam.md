# step10 — Postgres peer IPAM for SCORE hybrid (no cluster)

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Goal

Production-friendly peer IPAM backend for SCORE hybrid without requiring a live K8s cluster for the change itself.

## Backend priority

When `NewRelayManagerHybridFromSCORE` is called with `repository == nil`:

1. **Postgres** if `SCORE_DATABASE_URL` / `DATABASE_URL` / `SCORE_PG_HOST` set  
2. **SQLite** (`SCORE_SQLITE_PATH` or shared memory)  
3. Continue without peer IPAM (warn)

## Env

| Variable | Purpose |
|----------|---------|
| `SCORE_DATABASE_URL` | Preferred full DSN/URL |
| `DATABASE_URL` | Fallback DSN/URL |
| `SCORE_PG_HOST` | Discrete host mode |
| `SCORE_PG_PORT` | default 5432 |
| `SCORE_PG_USER` / `SCORE_PG_PASSWORD` / `SCORE_PG_DATABASE` | credentials |
| `SCORE_PG_SSLMODE` | default `disable` |
| `SCORE_SQLITE_PATH` | SQLite file or memory |

## Files

- `internal/database/postgres_peer_ipam.go` (+ tests)
- `internal/p2p/score_wiring.go` — `resolvePeerIPAMRepository`

## Tests (offline)

```bash
go test -short ./internal/database/ -run 'Postgres|SQLite'
go test -short ./internal/p2p/ -run 'SCORE|ResolvePeer'
go build ./cmd/relay/
```

Connect-fail test uses `127.0.0.1:1` (no real Postgres required).
