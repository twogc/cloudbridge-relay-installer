# step06 — Cadence shared factory + Cadence K8s archive

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Cadence dual-path (safe consolidation)

**Kept (intentional dual public API):**

| Entrypoint | API | Role |
|------------|-----|------|
| `cmd/relay` | `NewProductionClient` | Circuit breaker / domain-rich production wrapper |
| `cmd/ai-worker` | `NewClient` | Lighter worker client |

**Unified (this step):**

| Before | After |
|--------|--------|
| `CreateWorkflowServiceClient` stub-only | → delegates to `CreateWorkflowServiceClientImproved` |
| `CreateWorkflowServiceClientWithOptions` stub-only | → same improved factory (+ optional timeout probe) |
| `HealthCheck` separate stub | → `HealthCheckImproved` |

Files: `internal/cadence/service_client.go`, `service_client_improved.go`, `service_client_test.go`.

## K8s consolidation (incremental)

Archived non-canonical Cadence deployment variants:

```
archive/k8s/cadence/manifests-variants/
  cadence-clean-deployment.yaml
  cadence-fixed-*.yaml
  cadence-simple-*.yaml
  cadence-migrate-simple.yaml
```

**Kept:** `k8s/cadence/cadence.yaml`, schema/migrate/visibility jobs, README pointers.

## Tests

```bash
go test -short ./internal/cadence/ -run 'CreateWorkflow|HealthCheck'  # OK
go build ./cmd/relay/ ./cmd/ai-worker/                               # OK
```

## Not done

- Full merge of Client + ProductionClient types
- Full merge `k8s/` tree into overlays
- TChannel real transport implementation
