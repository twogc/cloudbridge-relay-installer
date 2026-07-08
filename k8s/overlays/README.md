# Kustomize overlays

| Overlay | Purpose | Status |
|---------|---------|--------|
| [`production/`](production/) | Thin wrapper over `k8s-manifests/00–08` | **Skeleton** — not wired into `deploy.sh` yet |

## Canonical deploy (today)

```bash
./scripts/deploy/deploy.sh   # applies k8s-manifests/00–15
```

## Future

1. Validate `kubectl apply -k k8s/overlays/production` on a test cluster.
2. Add optional flag to `deploy.sh` (CONSOLIDATION_PLAN §1 step 4).
3. Include Zitadel (09–15) only after ops/Helm path is decided.
