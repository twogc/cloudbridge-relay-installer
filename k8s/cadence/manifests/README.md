# Cadence K8s manifests (ops)

## Canonical

| Artifact | Path |
|----------|------|
| **Deployment** | [`../cadence.yaml`](../cadence.yaml) |
| **Prod stack** | `k8s-manifests/05-deployments.yaml` + `06-services.yaml` (Cadence ports 7933–7939) |
| **Schema / migrate jobs** | files in this directory (`*schema*`, `*migrate*`, `*visibility*`) |

## Archived

Historical `*-fixed*`, `*-simple*`, `*-clean*` deployments live in  
[`archive/k8s/cadence/manifests-variants/`](../../../archive/k8s/cadence/manifests-variants/).
