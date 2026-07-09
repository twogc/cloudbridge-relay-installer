# step09 — overlay e2e tooling + Zitadel overlay

**Дата:** 2026-07-09  
**Ветка:** `docs/consistency-sync-reports-ports`

## Deliverables

| Artifact | Role |
|----------|------|
| `scripts/deploy/validate-overlay.sh` | Offline path checks + optional `--live` |
| `deploy.sh --validate-overlay` | Wrapper to validator |
| `k8s/overlays/zitadel/` | 09 + 11 + 15 (minimal IAM path) |
| `docs/guides/OVERLAY_E2E_CHECKLIST.md` | Operator e2e sign-off |

## Verification (this environment)

```text
./scripts/deploy/validate-overlay.sh --all  → All offline checks passed
production: 9 resources OK
zitadel: 3 resources OK
kubectl not installed → render skipped (warn only)
```

## Not done

- Live cluster apply/e2e (no KUBECONFIG here)
- Including 10/12/13/14 alternate manifests in zitadel overlay
