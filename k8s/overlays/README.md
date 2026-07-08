# Kustomize overlays

| Overlay | Purpose | Status |
|---------|---------|--------|
| [`production/`](production/) | Core stack `k8s-manifests/00–08` | Optional: `deploy.sh --use-overlay` |
| [`zitadel/`](zitadel/) | IAM: 09 + 11 + 15 | Optional: `kubectl apply -k` after core |

## Default deploy

```bash
./scripts/deploy/deploy.sh
```

## Overlay deploy

```bash
# Offline checks (no cluster required)
./scripts/deploy/validate-overlay.sh --all

# Core
./scripts/deploy/deploy.sh --use-overlay
./scripts/deploy/deploy.sh --render-overlay

# IAM (after core)
kubectl apply -k k8s/overlays/zitadel
```

## E2E checklist

See [docs/guides/OVERLAY_E2E_CHECKLIST.md](../../docs/guides/OVERLAY_E2E_CHECKLIST.md).

## Future

1. Live-cluster sign-off → consider overlay as default.
2. Compose `production` + `zitadel` multi-base overlay if needed.
3. Drop duplicate numbered manifests after bake time.
