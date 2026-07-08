# CloudBridge Service Catalog

**Last Updated**: 2025-10-13
**Version**: MVP 1.0

## Overview

This document catalogs all services in the CloudBridge relay system, their dependencies, APIs, and deployment status.

---

## Production Services

### 1. CloudBridge Relay (Main Server)

**Directory**: `cmd/relay/`
**Binary**: `cloudbridge-relay`
**Port**: 8080 (HTTP), 8443 (HTTPS), 9090 (QUIC UDP)
**Status**: Production Ready
**Lines of Code**: 2391 (main.go)

#### Purpose
Main relay server handling P2P connections, MASQUE proxy, and NAT traversal.

#### Dependencies
- **Zitadel**: Authentication (JWKS)
- **PostgreSQL**: State persistence
- **Redis**: Session cache
- **TURN Server**: NAT traversal fallback
- **IPAM Service**: IP allocation
- **Prometheus**: Metrics collection

#### APIs Exposed
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /peers` - List connected peers
- `POST /connect` - P2P connection initiation
- WebSocket `/ws` - Real-time signaling

#### Configuration
```yaml
# config/relay.yaml
server:
  port: 8080
  quic_port: 9090
enhanced_quic:
  enabled: true
  experimental_features:
    bbrv2: true
    ack_frequency: true
```

#### Deployment
- **File**: `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml`
- **Replicas**: 3 (HA)
- **Resources**: 1 CPU, 2Gi RAM

---

### 2. TURN Server

**Directory**: N/A (coturn binary)
**Port**: 3478 (UDP/TCP)
**Status**: roduction Ready

#### Purpose
TURN relay for NAT traversal when direct P2P connection fails.

#### Dependencies
- None (standalone)

#### Configuration
```yaml
# ConfigMap: turn-server-config
realm: edge.2gc.ru
listening-port: 3478
lt-cred-mech: true
```

#### Deployment
- **File**: `k8s/deployments/turn-server-deployment.yaml` (479 lines)
- **Replicas**: 2 (HA)
- **Resources**: 500m CPU, 512Mi RAM

---

### 3. IPAM Service

**Directory**: `cmd/ipam-service/`
**Port**: 8083
**Status**: roduction Ready
**Lines of Code**: 156

#### Purpose
IP address management and allocation for P2P connections.

#### Dependencies
- **PostgreSQL**: IP pool storage

#### APIs Exposed
- `GET /health` - Health check
- `POST /allocate` - Allocate IP from pool
- `POST /release` - Release IP back to pool
- `GET /pools` - List IP pools

#### Deployment
- **File**: `k8s/deployments/ipam-service-deployment.yaml`
- **Replicas**: 2
- **Resources**: 200m CPU, 256Mi RAM

---

### 4. Tenant Policy Controller

**Directory**: `cmd/tenant-policy-controller/`
**Port**: 8080 (metrics), 8081 (health)
**Status**: roduction Ready
**Lines of Code**: 119

#### Purpose
Kubernetes controller for tenant network isolation using Calico NetworkPolicies.

#### Dependencies
- **Kubernetes API**: Watch pods, manage policies
- **Calico CNI**: Network policy enforcement

#### Deployment
- **File**: `k8s/tenant-policy-controller.yaml`
- **Replicas**: 1 (leader election enabled)
- **Resources**: 100m CPU, 64Mi RAM
- **RBAC**: ClusterRole with Calico CRD permissions

**Documentation**: [cmd/tenant-policy-controller/README.md](cmd/tenant-policy-controller/README.md)

---

### 5. AI Service

**Directory**: `cmd/ai-service/`
**Port**: 8085
**Status**: roduction Ready (Optional)
**Lines of Code**: 83

#### Purpose
AI-powered analytics and predictions for relay optimization.

#### Dependencies
- **Cloudflare AI API**: ML inference
- **PostgreSQL**: Metrics history

#### APIs Exposed
- `GET /health` - Health check
- `POST /analyze` - Analyze connection patterns
- `GET /predictions` - Get predictions
- `GET /anomalies` - Detect anomalies

#### Deployment
- **File**: `k8s/deployments/ai-service-deployment.yaml`
- **Replicas**: 1
- **Resources**: 500m CPU, 512Mi RAM

---

### 6. AI Worker

**Directory**: `cmd/ai-worker/`
**Port**: 8086
**Status**: roduction Ready (Optional)

#### Purpose
Background worker for batch AI processing.

#### Dependencies
- **AI Service**: Task queue
- **PostgreSQL**: Job state

#### Deployment
- **File**: `k8s/deployments/ai-worker-deployment.yaml`
- **Replicas**: 2
- **Resources**: 1 CPU, 1Gi RAM

---

### 7. WireGuard Service

**Directory**: `cmd/wireguard-service/`
**Port**: 51820 (UDP)
**Status**: roduction Ready (Optional)

#### Purpose
WireGuard VPN integration for secure mesh networking.

#### Dependencies
- **Kernel WireGuard**: VPN functionality

#### Deployment
- **File**: `k8s/deployments/wireguard-service-deployment.yaml`
- **Replicas**: 1
- **Resources**: 200m CPU, 256Mi RAM

---

## Optional/Future Services

### 8. Quantum Service

**Directory**: `cmd/quantum-service/`
**Port**: 8084
**Status**: DISABLED (Stub Implementation)
**Lines of Code**: 44

#### Purpose
Post-quantum cryptography and quantum key distribution (future feature).

#### Current State
All endpoints are stubs returning mock data. NOT deployed in MVP.

**Documentation**: [cmd/quantum-service/DISABLED.md](cmd/quantum-service/DISABLED.md)

---

### 9. QUIC Service

**Directory**: `cmd/quic-service/`
**Port**: 8081
**Status**: OPTIONAL (Metrics Sidecar)
**Lines of Code**: 77

#### Purpose
Optional standalone QUIC metrics collector.

#### Current State
Metrics-only service. Connection endpoints are stubs. NOT required for MVP (relay exposes QUIC metrics).

**Documentation**: [cmd/quic-service/README.md](cmd/quic-service/README.md)

---

## Supporting Services

### 10. PostgreSQL

**Port**: 5432
**Status**: equired

#### Purpose
Primary database for state persistence.

#### Stores
- Peer states
- IP allocations
- Metrics history
- AI analytics data

#### Deployment
- **File**: `k8s/deployments/postgres-simple.yaml`
- **Replicas**: 1 (consider HA for production)
- **Storage**: 10Gi PVC

---

### 11. Prometheus

**Port**: 9090
**Status**: equired

#### Purpose
Metrics collection and monitoring.

#### Scrapes
- Relay `/metrics`
- TURN server metrics
- IPAM metrics
- All service endpoints

#### Deployment
- **File**: `k8s/deployments/prometheus-optimized-deployment.yaml`
- **Replicas**: 1
- **Storage**: 50Gi PVC

---

### 12. Redis

**Port**: 6379
**Status**: equired (if using session cache)

#### Purpose
Session cache and pub/sub for distributed relay coordination.

#### Deployment
- **File**: TBD
- **Replicas**: 1 (consider Redis Sentinel for HA)

---

### 13. Cadence

**Ports**: 7933 (frontend), 7934 (matching), 7935 (history), 7939 (worker)
**Status**: ptional

#### Purpose
Workflow orchestration engine for complex relay operations.

#### Deployment
- **File**: `k8s/deployments/cadence-working.yaml`
- **Replicas**: 1
- **Resources**: 1 CPU, 2Gi RAM

---

### 14. Cadence Web

**Port**: 8088
**Status**: ptional

#### Purpose
Web UI for Cadence workflow monitoring.

#### Deployment
- **File**: `k8s/deployments/cadence-web-deployment.yaml`
- **Replicas**: 1

---

## Service Dependencies Matrix

| Service | Zitadel | PostgreSQL | Redis | TURN | IPAM | Prometheus | Calico |
|---------|---------|------------|-------|------|------|------------|--------|
| **Relay** |  |  |  |  |
| **TURN** |  |  |  | - |  |   |
| **IPAM** |  |   |  | - | |  |
| **Tenant Controller** |  |  |  |  |  |  |
| **AI Service** |  |   |  |  | |  |
| **WireGuard** |  |  |  |  |  |   |

---

## Deployment Profiles

### Minimal MVP

**Purpose**: Smallest viable deployment for testing

**Services**:
1. cloudbridge-relay (1 replica)
2. turn-server (1 replica)
3. postgres (1 replica)
4. prometheus (1 replica)

**Resources**: ~2 vCPU, 4Gi RAM

---

### Recommended Production

**Purpose**: Production-ready with HA

**Services**:
1. cloudbridge-relay (3 replicas)
2. turn-server (2 replicas)
3. ipam-service (2 replicas)
4. tenant-policy-controller (1 replica, HA via leader election)
5. postgres (1 replica with backups)
6. prometheus (1 replica with persistent storage)
7. redis (1 replica, consider Sentinel)

**Resources**: ~6 vCPU, 12Gi RAM

---

### Full Production

**Purpose**: All features enabled

**Includes**:
- All from Recommended Production
- ai-service (1 replica)
- ai-worker (2 replicas)
- wireguard-service (1 replica)
- cadence (1 replica)
- cadence-web (1 replica)

**Resources**: ~10 vCPU, 20Gi RAM

---

## API Authentication

### Authentication Methods

1. **JWT Tokens** (Primary)
   - Issued by: Zitadel
   - Validation: JWKS (RS256)
   - Claims: tenant_id, peer_id, permissions

2. **Service Tokens** (Internal)
   - Service-to-service communication
   - Bearer tokens in Authorization header

3. **mTLS** (Optional)
   - Client certificate authentication
   - For high-security deployments

---

## Monitoring & Observability

### Metrics Endpoints

All services expose Prometheus metrics at `/metrics`:
- **Relay**: `:8080/metrics`
- **IPAM**: `:8083/metrics`
- **AI Service**: `:8085/metrics`
- **Tenant Controller**: `:8080/metrics`

### Health Checks

All services expose health checks:
- **Liveness**: `/health` (is service running?)
- **Readiness**: `/ready` (can service accept traffic?)

### Logging

- **Format**: JSON structured logs
- **Library**: zap (Go services)
- **Levels**: debug, info, warn, error, fatal

---

## Network Topology

```
                    Internet
                       │
                       ↓
                 LoadBalancer
                       │
         ┌─────────────┼─────────────┐
         ↓             ↓             ↓
    Relay (8080)  Relay (8080)  Relay (8080)
         │             │             │
         └─────────────┼─────────────┘
                       │
         ┌─────────────┼─────────────┐
         ↓             ↓             ↓
    TURN (3478)   IPAM (8083)   PostgreSQL (5432)
```

---

## Security Considerations

### Service-to-Service Communication

1. **Internal Network**: All services in same Kubernetes namespace
2. **NetworkPolicies**: Tenant isolation via Calico
3. **TLS**: All external endpoints use TLS (cert-manager)
4. **Secrets**: Managed via Kubernetes Secrets or External Secrets Operator

### Access Control

- **Relay**: Requires valid JWT token
- **TURN**: Shared secret authentication
- **IPAM**: Internal-only (not exposed externally)
- **Tenant Controller**: RBAC via ServiceAccount

---

## Troubleshooting

### Service Not Starting

1. **Check logs**: `kubectl logs <pod-name>`
2. **Check dependencies**: Are required services running?
3. **Check config**: Is ConfigMap/Secret correct?
4. **Check resources**: Sufficient CPU/memory?

### Common Issues

**Problem**: Relay can't connect to Zitadel
**Solution**: Check JWKS endpoint, verify network connectivity

**Problem**: TURN server not allocating ports
**Solution**: Check firewall rules, verify UDP port 3478 is open

**Problem**: IPAM out of IPs
**Solution**: Expand IP pool, check for leaks (unreleased IPs)

---

## Future Enhancements

1. **Service Mesh**: Consider Istio/Linkerd for traffic management
2. **Distributed Tracing**: Add OpenTelemetry
3. **API Gateway**: Add Kong/APISIX for unified API
4. **Auto-scaling**: HPA based on connection count
5. **Multi-region**: Deploy relay in multiple regions

---

## Related Documents

- [PROJECT_ARCHITECTURE_ANALYSIS.md](PROJECT_ARCHITECTURE_ANALYSIS.md) - Detailed architecture
- [MVP_IMPLEMENTATION_PLAN.md](MVP_IMPLEMENTATION_PLAN.md) - Implementation roadmap
- [MANUAL_TESTING_PLAN.md](MANUAL_TESTING_PLAN.md) - Testing procedures
- [ARCHITECTURE_CLEANUP_SUMMARY.md](ARCHITECTURE_CLEANUP_SUMMARY.md) - Cleanup report

---

**Maintained by**: CloudBridge Team
**Questions**: Open issue on GitHub
