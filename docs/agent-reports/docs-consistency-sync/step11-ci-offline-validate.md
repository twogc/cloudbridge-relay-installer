# step11 — CI offline-validate (no cluster)

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Deliverables

| Artifact | Role |
|----------|------|
| `.github/workflows/offline-validate.yml` | PR/push CI: overlay validate + kustomize + unit tests |
| `make validate-overlay` | Local offline overlay checks |
| `make test-offline` | Local mirror of CI unit jobs |
| Extra Postgres env unit tests | DATABASE_URL fallback / SCORE URL wins |

## CI jobs

1. **overlay-and-scripts** — `bash -n`, `validate-overlay.sh --all`, `kubectl kustomize` render  
2. **unit-score-ipam** — database/p2p/cadence focused tests + build relay/ai-worker  

No `kubectl cluster-info` / no live apply.
