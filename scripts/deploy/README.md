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

Cadence K8s manifests: `k8s/cadence/manifests/`. Zitadel values: `k8s/zitadel/`.
