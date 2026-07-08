# Operations

This page covers deployment, CI/CD, monitoring, infrastructure provisioning, and regulatory compliance operations.

---

## Kubernetes Deployment

### Manifest Organization

Kubernetes manifests are organized in two locations:

**`k8s/`** — Detailed, domain-organized manifests:

```
k8s/
├── deployments/          # Application deployments
├── services/             # Service definitions
├── namespaces/           # Namespace isolation
├── configmaps/           # Configuration
├── rbac/                 # Role-based access control
├── network-policies/     # Network segmentation
├── security/             # Security policies (cert-manager, mTLS, external secrets)
├── persistentvolumeclaims/ # Storage
├── storageclasses/       # Storage class definitions
├── monitoring/           # Prometheus/Grafana
├── monitoring-stack/     # Production monitoring (VictoriaMetrics, AlertManager, SLO)
├── cadence/              # Cadence workflow engine (manifests/ subdir for detailed YAMLs)
├── zitadel/              # Zitadel IAM (Helm values, nginx configs, ingress)
├── bgp-monitoring/       # BGP routing monitoring
├── policies/             # Kyverno pod security policies
└── cronjobs/             # Scheduled tasks
```

**`k8s-manifests/`** — Simplified deployment (numbered for ordering):

| File | Purpose |
|---|---|
| `00-namespace.yaml` | Namespace creation |
| `01-pvc.yaml` | PersistentVolumeClaim for logs |
| `02-secrets.yaml` | Secret management |
| `03-configmap.yaml` | Configuration |
| `04-registry.yaml` | Docker registry |
| `05-deployments.yaml` | Relay deployment (3 replicas), AI service, Cadence, Cassandra |
| `06-services.yaml` | ClusterIP services |
| `07-ingress.yaml` | NGINX Ingress with TLS |
| `08-servicemonitor.yaml` | Prometheus ServiceMonitor |
| `09-zitadel.yaml` | Zitadel IAM |
| `10-zitadel-complete.yaml` | Zitadel complete deployment |
| `11-postgres-zitadel.yaml` | PostgreSQL for Zitadel |
| `12-default-backend*.yaml` | NGINX default backend (multiple variants) |
| `13-beautiful-homepage.yaml` | Landing page |
| `14-default-backend-with-config.yaml` | Default backend with ConfigMap |
| `15-zitadel-root-ingress.yaml` | Zitadel root domain ingress |

### Core Deployment (`k8s-manifests/05-deployments.yaml`)

The main relay deployment:
- 3 replicas with rolling update strategy
- WireGuard integration (host network, CAP_NET_ADMIN)
- Health/readiness probes on port 9091
- TLS volume mounts
- Resource limits and requests

### Ingress (`k8s-manifests/07-ingress.yaml`)

NGINX Ingress controller with TLS termination:
- Host: `edge.2gc.ru`
- Routes: `/api` → relay, `/ai` → AI service, `/cadence` → Cadence
- TLS via cert-manager

### Namespaces

| Namespace | Purpose |
|---|---|
| `cloudbridge` | Application workloads |
| `cloudbridge-monitoring` | Observability stack |
| `cloudbridge-staging` | Staging environment |

---

## Docker Build

### Dockerfile (`docker/Dockerfile`)

Multi-stage build:

```dockerfile
# Stage 1: Builder
FROM golang:1.25.12-alpine AS builder
# CGO_ENABLED=1, cross-compilation support

# Stage 2: Runtime
FROM alpine:3.18
# Non-root user (appuser:1001)
# WireGuard tools included
# Health check on port 9091
```

**Exposed ports (Dockerfile):** 9091, 8082, 8443, 5553/udp, 19302/udp, 9090/udp, 9092/udp — see [Port Scheme](port-scheme.md) for installer config (`configs/config.json`: main **5550**, API **5552**, metrics **5551**)

### Container Registries

| Registry | Usage |
|---|---|
| `ghcr.io` | GitHub Container Registry (primary) |
| `docker.io/mlanies/cloudbridge-relay` | Docker Hub |
| VK Cloud Registry | Russian mirror for faster pulls |

---

## OVA Marketplace Build (`mws/`)

### Overview

Pre-built virtual appliances for the MWS B2B Store marketplace. Each OVA contains:

- Ubuntu 22.04 LTS Server
- Docker CE + Compose plugin
- PostgreSQL 15 (Zitadel database)
- Zitadel v2.64.1 (OIDC/control plane)
- CloudBridge Relay (latest)

### Build Options

```bash
cd mws
./scripts/build-ova.sh <method>
# Methods: packer | qemu | manual | upload
```

**Packer** (`mws/packer/cloudbridge-relay.pkr.hcl`):
- VMware vSphere builder
- 4 vCPU, 8GB RAM, 80GB disk
- Ubuntu 22.04 ISO autoinstall

**QEMU** (`mws/scripts/build-ova-qemu.sh`):
- QEMU/KVM-based build for non-vCenter environments
- Creates OVA without VMware dependency

### Docker Compose Stack (`mws/templates/docker-compose.minimal.yml`)

```yaml
services:
  postgres:   # PostgreSQL 15 Alpine
  zitadel:    # Zitadel IAM v2.64.1
  cloudbridge-relay:  # Main relay service
# Ports: 8080 (Zitadel), 5550/9090 (relay TCP/QUIC), 51820 (WireGuard) — see port-scheme.md
```

### Cloud-Init (`mws/cloud-init/cloudbridge-relay-minimal.yaml`)

Automated initial setup for OVA instances.

---

## CI/CD Pipeline (`.github/workflows/`)

### Release Pipeline (`release.yml`)

**Trigger:** Tag push (`v*`) or manual dispatch

**Steps:**
1. Build Go binaries (linux/amd64, linux/arm64, windows/amd64, macOS/amd64+arm64)
2. Build & push multi-arch Docker image to GHCR + Docker Hub
3. Create GitHub Release with changelog
4. Update Docker Hub description

### Staging Pipeline (`staging-deployment.yml`)

**Trigger:** Push to `develop` or `staging` branches

**Stages:**
1. **Pre-deployment:** K8s cluster connectivity check
2. **Build:** Docker Buildx with GHCR push
3. **Deploy:** K8s rolling update to `cloudbridge-staging` namespace

**Features:** Configurable image tag, skip-tests option, force-deploy

---

## Infrastructure Provisioning

### VK Cloud K8s (`terraform-k8s/`)

**Provider:** VK Cloud (`vk-cs/vkcs ~> 0.8.0`)

**Resources:**
- Kubernetes cluster (1.28)
- Node group (2 worker nodes, STD3-4-8 flavor)
- Region: ME1 (Moscow)

```bash
cd terraform-k8s
terraform init
terraform plan
terraform apply -auto-approve
```

### Ansible (`ansible/`)

Bare-metal Docker deployment playbook:
- Install Docker + docker-compose
- Deploy CloudBridge Relay stack
- Automated relay host setup

### Deployment Script (`scripts/deploy/deploy.sh`)

Full automated K8s deployment:
1. Dependencies check (kubectl, helm, docker, terraform)
2. Docker Hub mirrors setup (Russian/Chinese mirrors)
3. Cluster connectivity verification
4. Namespace-based deployment

---

## Monitoring Stack

### Production Monitoring (`k8s/monitoring-stack/`)

**Architecture:**
```
Prometheus (HA, 2 replicas) → VictoriaMetrics (1-year retention)
       ↓                              ↓
  AlertManager (3 replicas)      Grafana (dashboards)
       ↓
  SLO Monitoring (Sloth, burn-rate alerts)
```

**Components:**

| Component | Config | Retention |
|---|---|---|
| Prometheus | HA (2 replicas), remote write | 15 days local |
| VictoriaMetrics | Cluster (vmstorage/vminsert/vmselect) | 1 year |
| AlertManager | HA (3 replicas), advanced routing | — |
| Grafana | Provisioned datasources + dashboards | — |

**SLO Monitoring:**
- `cloudbridge-slo.yaml` — Service Level Objectives
- Burn-rate alerts for error budget consumption
- Multi-window alerting (5m, 30m, 1h, 6h)

### BGP Monitoring (`k8s/bgp-monitoring/`)

Network routing monitoring for anycast deployments:

| Exporter | Port | BGP Daemon |
|---|---|---|
| BIRD Exporter | 9324 | BIRD |
| FRR Exporter | 9342 | FRR |
| GoBGP Exporter | 9474 | GoBGP |

**Dashboards:**
- BGP session status and prefix counts
- Anycast health monitoring
- Route origin validation

### Operational Scripts (`scripts/`)

| Script | Purpose |
|---|---|
| `deploy-monitoring-stack.sh` | Deploy full monitoring stack |
| `deploy-bgp-monitoring.sh` | Deploy BGP monitoring |
| `test-quick.sh` | Fast E2E testing |
| `health-check.sh` | System diagnostics |
| `update-production.sh` | Rolling production updates |
| `zitadel-setup.sh` | Zitadel IAM configuration |
| `optimize-prometheus.sh` | Prometheus performance tuning |

---

## Compliance Operations

### Russian Regulatory Requirements

| Regulation | Requirement | Implementation |
|---|---|---|
| **152-ФЗ** | Data localization in Russia | All data stored in VK Cloud Moscow region |
| **152-ФЗ** | Consent management | Token-based consent tracking |
| **187-ФЗ** | Critical infrastructure security | Certified protection, audit logging |
| **СОРМ** | Lawful interception | Traffic interception support |

**Validation:**
- `internal/compliance/russian_compliance.go` — Configuration and connection validation
- IP range validation for Russian networks
- Data residency enforcement

---

## Security

### Container Security

- Non-root containers (UID 1001)
- Read-only root filesystem where possible
- Resource limits enforced
- Network policies for pod-to-pod traffic
- Secrets via K8s Secrets (not environment variables)

### Network Security

- TLS/gTLS with cert-manager
- WireGuard encrypted tunnels
- mTLS for inter-service communication
- Network policies (Calico) for tenant isolation
- RBAC for K8s API access

### Authentication

- Zitadel OIDC (canonical IAM; Keycloak removed)
- JWT tokens with short TTL
- JWKS-based key rotation

### Monitoring Security

- Trivy vulnerability scanning (`.trivyignore` for known exceptions)
- Security audit reports (`security_reports/`)
- Prometheus alerting on security events

---

## Development Workflow

### Building

```bash
make build          # Build relay binary
make docker-build   # Build Docker image
make test           # Run all tests
make test-critical  # Run critical tests only
```

### Testing

```
test/
├── e2e/
│   ├── auth_test.go
│   ├── integration_test.go
│   ├── kube_proxy_test.go
│   └── metrics_test.go
tests/
├── vpn_integration_test.go
cmd/relay/
├── comprehensive_integration_test.go
├── integration_test.go
├── masque_*_test.go
└── p2p_api_test.go
```

### Code Quality

- **Linting:** `.golangci.yml` with comprehensive rules
- **Testing:** Unit, integration, and E2E tests
- **Security:** Trivy scanning, `.trivyignore`
- **Formatting:** `.gitattributes` for consistent line endings

---

## Source Map

| Area | Key Files |
|---|---|
| K8s deployments | `/k8s-manifests/05-deployments.yaml` |
| K8s services | `/k8s-manifests/06-services.yaml` |
| K8s ingress | `/k8s-manifests/07-ingress.yaml` |
| Docker build | `/docker/Dockerfile` |
| CI/CD release | `/.github/workflows/release.yml` |
| CI/CD staging | `/.github/workflows/staging-deployment.yml` |
| Terraform | `/terraform-k8s/k8s-cluster.tf` |
| Monitoring stack | `/k8s/monitoring-stack/` |
| BGP monitoring | `/k8s/bgp-monitoring/` |
| OVA build | `/mws/scripts/build-ova.sh` |
| Packer template | `/mws/packer/cloudbridge-relay.pkr.hcl` |
| Deploy script | `/scripts/deploy/deploy.sh` |
| Makefile | `/Makefile` |
| Compliance | `/internal/compliance/russian_compliance.go` |
