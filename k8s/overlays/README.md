# Kustomize overlays

| Overlay | Purpose | Status |
|---------|---------|--------|
| [`production/`](production/) | Thin wrapper over `k8s-manifests/00–08` | **Optional** via `deploy.sh --use-overlay` |

## Canonical deploy (default)

```bash
./scripts/deploy/deploy.sh   # file-by-file k8s-manifests/
```

## Overlay deploy (optional)

```bash
./scripts/deploy/deploy.sh --use-overlay
./scripts/deploy/deploy.sh --use-overlay --dry-run
./scripts/deploy/deploy.sh --render-overlay
kubectl apply -k k8s/overlays/production
```

Zitadel (`09–15`) is **not** in the overlay — deploy after core.

## Future

1. Live-cluster e2e validation.
2. Separate Zitadel overlay.
3. Promote overlay to default only after bake time.
