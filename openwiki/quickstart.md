# CloudBridge Relay — Quickstart

## What Is This?

CloudBridge Relay is a **Go-based microservices platform** for secure network bridging, tunneling, and P2P connectivity. It provides high-performance relay services with support for WireGuard VPN, QUIC/HTTP3, ICE NAT traversal, and AI-powered network optimization.

The system is designed to run in Russia/CIS regions on **VK Cloud Kubernetes** and is also distributed as **B2B OVA virtual appliances** via the MWS Marketplace.

### Key Capabilities

| Capability | Description |
|---|---|
| **P2P Networking** | Direct peer-to-peer connections with mesh topology, hybrid routing, and site-to-site VPN |
| **QUIC Transport** | Enhanced QUIC with BBRv2 congestion control, FEC, ACK frequency optimization |
| **NAT Traversal** | Full ICE stack (STUN/TURN) with a built-in TURN server for symmetric NAT |
| **WireGuard VPN** | Per-tenant encrypted tunnels with Kubernetes-aware configuration |
| **Protocol Engine** | Auto-switching between QUIC → HTTP/2 → WebSocket based on quality metrics |
| **MASQUE Protocol** | RFC 9297 HTTP capsule encapsulation for relay fallback |
| **AI/ML Optimization** | Anomaly detection, load prediction, and ML-based route optimization |
| **Workflow Engine** | Cadence-based orchestration for deployment, cert rotation, and module init |
| **Tenant Isolation** | Multi-tenant with K8s controller, Calico network policies, per-tenant WireGuard |
| **Russian Compliance** | 152-ФЗ (data localization), 187-ФЗ (critical infra), SORM support |

---

## Repository Structure

```
cloudbridge-relay-installer/
├── cmd/                          # Service entrypoints (10 binaries)
│   ├── relay/                    # Core relay server (main service)
│   ├── api/                      # REST API server
│   ├── ai-service/               # AI inference service
│   ├── ai-analytics/             # AI analytics engine
│   ├── ai-worker/                # Cadence AI task worker
│   ├── quic-service/             # QUIC experimental features
│   ├── cadence-service/          # Workflow orchestration API
│   ├── ipam-service/             # IP address management
│   ├── tenant-policy-controller/ # K8s CRD controller
│   └── cni-plugin/               # CNI plugin for K8s pod networking
│
├── internal/                     # Core library packages
│   ├── p2p/                      # P2P networking (peer discovery, mesh, connections)
│   ├── quic/                     # QUIC transport (enhanced, experimental features)
│   ├── nat/                      # NAT traversal (ICE, STUN, built-in TURN)
│   ├── wireguard/                # WireGuard VPN management
│   ├── tunnel/                   # TCP tunnel proxy
│   ├── protocol/                 # Protocol engine (QUIC/HTTP2/WebSocket switching)
│   ├── masque/                   # MASQUE protocol (RFC 9297 capsules)
│   ├── auth/                     # Authentication (Zitadel, Keycloak, JWT)
│   ├── token/                    # Token validation and claims
│   ├── ai/                       # AI/ML (anomaly detection, prediction, routing)
│   ├── cadence/                  # Cadence workflow definitions
│   ├── ipam/                     # IPAM service logic
│   ├── controllers/              # K8s tenant policy controller
│   ├── api/                      # HTTP API handlers, router, middleware
│   ├── metrics/                  # Prometheus metrics (all subsystems)
│   ├── compliance/               # Russian regulatory compliance
│   ├── database/                 # Database models and repositories
│   ├── transport/grpc/           # gRPC transport layer
│   ├── balancer/                 # Load balancing
│   ├── circuitbreaker/           # Circuit breaker pattern
│   └── ...                       # 20+ more packages
│
├── config/                       # Configuration loading and types
├── proto/                        # Protobuf/gRPC service definitions
├── pkg/                          # Public library packages
│
├── k8s/                          # Kubernetes manifests (organized by domain)
├── k8s-manifests/                # Simplified K8s deployment manifests
├── docker/                       # Dockerfiles
├── mws/                          # OVA marketplace build (Packer, cloud-init)
├── terraform-k8s/                # VK Cloud K8s cluster provisioning
├── terraform/                    # Additional Terraform configs
├── ansible/                      # Ansible playbooks for bare-metal setup
├── scripts/                      # Operational scripts (100+ files)
│
├── test/                         # E2E and integration tests
├── tests/                        # Additional test suites
├── examples/                     # Usage examples
├── docs/                         # Additional documentation
│
├── deploy.sh                     # Automated K8s deployment script
├── Makefile                      # Build/test/deploy targets
├── go.mod                        # Go 1.25.1 module definition
└── README.md                     # Project overview (Russian)
```

---

## Prerequisites

### System Requirements
- **Go 1.25.1+** for building from source
- **Docker 20.10+** for container builds
- **kubectl 1.28+** for K8s operations
- **Helm 3.12+** for chart management
- **Terraform 1.5+** for infrastructure provisioning

### VK Cloud Requirements
- VK Cloud account with Kubernetes access
- OpenStack RC file v3 for authentication
- Sufficient quotas for cluster creation

---

## Quick Start Paths

### Path 1: Deploy to VK Cloud Kubernetes (Production)

```bash
# 1. Clone and configure
source openstack_auth.sh
export KUBECONFIG=~/.kube/config-vkcloud

# 2. Provision K8s cluster (if needed)
cd terraform-k8s && terraform init && terraform apply -auto-approve

# 3. Deploy all services
./deploy.sh

# Or manual:
kubectl apply -f k8s-manifests/
```

See: [Operations Guide](operations.md) for full deployment details.

### Path 2: Build and Run Locally

```bash
# Build the relay binary
make build

# Run with configuration
./cloudbridge-relay -config config.json -log-level info
```

### Path 3: Docker Compose (Development)

```bash
# Build Docker image
make docker-build

# Run the stack
docker-compose up
```

### Path 4: B2B OVA (Marketplace)

```bash
cd mws
# Build OVA with Packer (requires VMware vSphere)
./scripts/build-ova.sh packer

# Or with QEMU
./scripts/build-ova.sh qemu
```

See: [Operations Guide](operations.md#ova-marketplace-build) for OVA details.

---

## Configuration

The main relay server uses a JSON configuration file with 60+ sections. Key configuration areas:

| Section | Purpose |
|---|---|
| `server` | Listen address, timeouts |
| `zitadel` | Identity provider (OIDC, JWKS, issuer) |
| `wireguard` | VPN interface, keys, peer config |
| `quic` / `enhancedQuic` | QUIC transport tuning |
| `ice` / `stun` / `turn` | NAT traversal servers |
| `ai` | AI/ML model paths and endpoints |
| `cadence` | Workflow engine connection |
| `masque` | MASQUE proxy settings |
| `protocol` | Protocol selection and fallback |
| `kubernetes` | K8s cluster proxy config |
| `grpc` | gRPC transport settings |

Services outside the relay use environment variables for configuration.

---

## Building

```bash
# Full build (produces ./cloudbridge-relay)
make build

# Docker image
make docker-build

# All tests
make test

# Critical tests only
make test-critical
```

The Makefile injects version and commit via ldflags. CGO is required (`CGO_ENABLED=1`).

---

## Core Ports

| Port | Protocol | Service |
|---|---|---|
| 9090 | TCP/UDP | Main relay (HTTP + QUIC) |
| 8082 | TCP | AI service API |
| 8443 | TCP | HTTPS API |
| 8085 | TCP | Cadence worker health |
| 5553 | UDP | WireGuard |
| 19302 | UDP | STUN |
| 9090 | UDP | QUIC transport |
| 9092 | UDP | QUIC secondary |

---

## Navigation

| Page | Content |
|---|---|
| [Architecture](architecture.md) | System architecture, transport layers, auth, inter-service patterns |
| [Networking Domain](domain-networking.md) | P2P, NAT traversal, WireGuard, IPAM, protocol engine, MASQUE |
| [Services Domain](domain-services.md) | AI/ML, Cadence workflows, K8s controllers, CNI plugin |
| [Operations](operations.md) | Deployment, CI/CD, monitoring, infrastructure, compliance |
