# Overlay deploy — E2E checklist

Live-cluster validation for optional kustomize path.  
**Default production path remains** `./scripts/deploy/deploy.sh` (file-by-file `k8s-manifests/`).

## Prerequisites

- [ ] `KUBECONFIG` points at target cluster
- [ ] `kubectl cluster-info` succeeds
- [ ] Docker available (for `load_images` in full deploy)
- [ ] Repo root: `cd cloudbridge-relay-installer`

## 0. Offline (no cluster)

```bash
./scripts/deploy/validate-overlay.sh --all
bash -n scripts/deploy/deploy.sh
./scripts/deploy/deploy.sh --help
```

- [ ] All resource paths resolve
- [ ] `kubectl kustomize` / `kustomize build` OK if tool present

## 1. Dry-run (cluster optional for client dry-run)

```bash
./scripts/deploy/deploy.sh --use-overlay --dry-run --skip-mirrors
```

- [ ] No apply errors on dry-run client validation

## 2. Core stack via overlay

```bash
./scripts/deploy/deploy.sh --use-overlay --skip-mirrors
# or step-by-step:
kubectl apply -k k8s/overlays/production
```

- [ ] Namespace `cloudbridge` exists
- [ ] PVC Bound (logs, cadence, registry)
- [ ] `docker-registry` Available
- [ ] Images pushed (`load_images`)
- [ ] Deployments: `cloudbridge-relay`, `ai-service`, `cadence` (cadence may need Cassandra env from ops)
- [ ] Services expose relay **9090**, metrics **5551**, AI **8083**

```bash
kubectl get pods,svc,pvc -n cloudbridge
kubectl get deploy -n cloudbridge
curl -sS http://127.0.0.1:5551/metrics   # after port-forward if needed
```

## 3. Zitadel overlay (after core)

```bash
kubectl apply -k k8s/overlays/zitadel
# or validate first:
./scripts/deploy/validate-overlay.sh --overlay k8s/overlays/zitadel
```

- [ ] Namespace `zitadel` / postgres for Zitadel
- [ ] Zitadel pod Ready (may need storage class / secrets)
- [ ] OIDC discovery reachable
- [ ] Relay `zitadel` config points at correct issuer

See also: [ZITADEL_SETUP.md](ZITADEL_SETUP.md), [openwiki/deploy.md](../../openwiki/deploy.md).

## 4. Live validate script

```bash
./scripts/deploy/validate-overlay.sh --all --live
```

## 5. Rollback notes

```bash
# Remove core overlay resources (careful — deletes apps)
kubectl delete -k k8s/overlays/production
kubectl delete -k k8s/overlays/zitadel
```

Prefer git-based re-apply of known-good manifests over ad-hoc edits.

## Sign-off

| Check | Operator | Date | OK |
|-------|----------|------|-----|
| Offline validate | | | |
| Overlay core deploy | | | |
| Zitadel overlay | | | |
| Relay health + metrics | | | |

**Promote overlay to default** only after repeated green e2e (CONSOLIDATION_PLAN §1).
