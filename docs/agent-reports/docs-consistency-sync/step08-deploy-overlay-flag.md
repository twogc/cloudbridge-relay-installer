# step08 — deploy.sh optional kustomize overlay

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Цель

Подключить `k8s/overlays/production` к `scripts/deploy/deploy.sh` **без** смены default-пути.

## Изменения

### `scripts/deploy/deploy.sh`

| Flag / env | Behaviour |
|------------|-----------|
| _(default)_ | `DEPLOY_MODE=manifests` — file-by-file as before |
| `--use-overlay` | `kubectl apply -k k8s/overlays/production` for core 00–08 |
| `--overlay DIR` | custom overlay path |
| `--mode overlays\|manifests` | explicit mode |
| `--dry-run` | client dry-run; skip waits/load_images where unsafe |
| `--skip-mirrors` | skip Docker daemon mirror setup |
| `--render-overlay` | `kubectl kustomize` only |
| `--help` | usage |

Overlay flow: secrets bootstrap → apply overlay → wait registry → `load_images` → re-apply → wait deployments.

### Overlay fix

`kustomization.yaml` resources: `08-servicemonitor.yaml` (was wrong `08-network-policies.yaml`).

### Docs

`openwiki/deploy.md`, `k8s/overlays/README.md`, `CONSOLIDATION_PLAN` §1 step 4, deploy README, AGENTS.

## Verification

```bash
bash -n scripts/deploy/deploy.sh
./scripts/deploy/deploy.sh --help
# if kubectl present:
# kubectl kustomize k8s/overlays/production | head
```

No live cluster e2e in this step.
