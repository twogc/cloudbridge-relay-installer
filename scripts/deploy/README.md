# Deployment scripts

Run from repository root (scripts `cd` to repo root automatically):

```bash
source scripts/deploy/openstack_auth.sh
./scripts/deploy/deploy.sh
sudo ./scripts/deploy/setup-docker-mirrors.sh
./scripts/deploy/setup-cadence.sh
```

| Script | Purpose |
|--------|---------|
| `deploy.sh` | Full VK Cloud K8s deployment |
| `setup-docker-mirrors.sh` | Docker Hub mirrors for RU |
| `setup-cadence.sh` | Cadence + PostgreSQL in K8s |
| `openstack_auth.sh` | OpenStack/VK Cloud env vars |
| `vk-cloud-rc.sh` | VK Cloud RC helper |
| `check-zitadel-access.sh` | Zitadel connectivity check |

Cadence: ops deploy `k8s/cadence/cadence.yaml`; schema jobs in `k8s/cadence/manifests/` (fixed/simple variants archived). Zitadel values: `k8s/zitadel/`.

### deploy.sh modes

```bash
./scripts/deploy/deploy.sh                 # default: k8s-manifests/ file-by-file
./scripts/deploy/deploy.sh --use-overlay   # kustomize k8s/overlays/production
./scripts/deploy/deploy.sh --render-overlay
./scripts/deploy/deploy.sh --help
```
