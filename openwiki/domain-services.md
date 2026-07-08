# Services Domain

This page covers the AI/ML pipeline, Cadence workflow engine, Kubernetes controllers, and the CNI plugin.

---

## AI/ML Services (`internal/ai/`)

### Overview

The AI subsystem provides intelligent network optimization through anomaly detection, load prediction, and route optimization. It runs as both an in-process library (within the relay) and a standalone HTTP service (`cmd/ai-service/`).

### Service Architecture (`internal/ai/service.go`)

```
AI Service
├── NetworkAnomalyDetector   — Detects unusual network behavior
├── LoadPredictor            — Forecasts load over configurable horizon
├── RouteOptimizer           — ML-based optimal route selection
├── DataPipeline             — Data collection and processing
├── AIMonitoring             — Operational health
└── AIMetrics                — Prometheus instrumentation
```

### Anomaly Detection (`internal/ai/anomaly/`)

| File | Purpose |
|---|---|
| `detector.go` | Base anomaly detection interface |
| `isolation_forest.go` | Isolation Forest algorithm implementation |
| `optimized_detector.go` | Performance-optimized detector |
| `real_anomaly_detector.go` | Production anomaly detector with real metrics |
| `real_isolation_forest.go` | Production Isolation Forest with real data |

**Input:** `NetworkMetrics` (latency, throughput, packet loss, connection count)
**Output:** Anomaly score + anomaly type classification

### Load Prediction (`internal/ai/predictive/`)

| File | Purpose |
|---|---|
| `load_predictor.go` | Base prediction interface |
| `arima.go` | ARIMA time-series model |
| `optimized_predictor.go` | Performance-optimized predictor |
| `real_load_predictor.go` | Production predictor with real data |
| `time_series_buffer.go` | Circular buffer for time-series data |

**Capabilities:**
- Configurable prediction horizon
- Seasonal pattern detection
- Confidence intervals

### Route Optimization (`internal/ai/routing/`)

| File | Purpose |
|---|---|
| `optimizer.go` | Route optimization interface |
| `path_finder.go` | Multi-path route discovery |
| `latency_predictor.go` | Latency prediction for route scoring |
| `bandit.go` | Multi-armed bandit for exploration/exploitation |
| `network_metrics_collector.go` | Real-time network metrics |
| `optimized_optimizer.go` | Production-optimized optimizer |

**Algorithm:** Multi-armed bandit with latency prediction for optimal path selection.

### Data Pipeline (`internal/ai/pipeline/`)

```
Collectors → Processors → Storage → Analysis
```

| Component | Purpose |
|---|---|
| `collectors.go` | Data source abstraction |
| `processors.go` | Transform and enrich data |
| `storage.go` | Persistence layer |
| `data_pipeline.go` | Pipeline orchestration |
| `system_metrics.go` | System-level metric collection |
| `real_data_collector.go` | Production data collection |

### ML Models (`internal/ai/ml/`)

| File | Purpose |
|---|---|
| `integration.go` | Model integration layer |
| `model_implementations.go` | Concrete model implementations |
| `real_models.go` | Production model implementations |

### Cadence Integration (`internal/ai/cadence/`)

AI tasks are orchestrated as Cadence workflows:
- Async anomaly detection jobs
- Batch prediction processing
- Model retraining pipelines

### AI Analytics Engine (`cmd/ai-analytics/`)

Standalone analytics service:

- **Data Collection:** Prometheus collector + AI service collector
- **Analysis:** Cloudflare Workers AI (LLM-based analysis)
- **Reports:** Automated report generation
- **Notifications:** Slack + Telegram alerting

### AI Service API (`cmd/ai-service/`)

| Endpoint | Purpose |
|---|---|
| `GET /health` | Service health |
| `GET /anomalies` | Current anomalies |
| `POST /analyze` | Analyze network data |
| `GET /metrics` | Prometheus metrics |

---

## Cadence Workflow Engine (`internal/cadence/`)

### Overview

Cadence (Uber's workflow engine) provides durable, multi-step orchestration for operations that must survive failures.

### Workflows (`internal/cadence/workflows/`)

| Workflow | Purpose |
|---|---|
| `CloudBridgeMainWorkflow` | Top-level system orchestration (2h timeout) |
| `RelayDeploymentWorkflow` | Multi-step relay deployment pipeline |
| `CertificateRotationWorkflow` | Certificate lifecycle management |
| `P2PConnectionWorkflow` | Durable P2P connection establishment |
| `P2PQUICWorkflow` | QUIC-specific P2P operations |
| `AIOptimizationWorkflow` | AI model execution pipeline |
| `QuantumWorkflow` | Post-quantum crypto operations |
| `ScalingWorkflow` | Auto-scaling orchestration |
| `SecurityWorkflow` | Security audit and compliance |

### Deployment Workflow (`RelayDeploymentWorkflow`)

```
1. Validate Configuration
2. Set Up WireGuard
3. Configure QUIC Server
4. Start Services
5. Health Check (2-minute timeout)
```

Each step is a Cadence activity with typed input/output structs.

### Main Workflow (`CloudBridgeMainWorkflow`)

```go
type CloudBridgeMainWorkflowInput struct {
    Domain  string
    Config  map[string]string
}

type CloudBridgeMainWorkflowResult struct {
    Status    string
    Modules   []ModuleStatus
    Timestamp time.Time
}
```

Uses `ChildWorkflowOptions` with 2-hour execution timeout for module initialization.

### Activities (`internal/cadence/activities/`)

| Activity | Purpose |
|---|---|
| Common activities | Shared utility activities |
| P2P QUIC activities | P2P connection management |
| AI optimization activities | AI task execution |

### Client (`internal/cadence/client.go`)

- Connects to Cadence server via Thrift RPC
- Manages workflow and activity registration
- Supports retry policies and timeouts
- Production client with connection pooling

### Worker (`internal/cadence/worker/worker.go`)

- Registers workflow and activity handlers
- Manages task pollers
- Graceful shutdown

### Health Monitoring

Cadence service exposes:
- `GET /health` — Cadence connectivity check
- `GET /ready` — Readiness probe
- `GET /version` — Version info
- `GET /workflows` — Active workflow list
- `GET /metrics` — Prometheus metrics

---

## Tenant Policy Controller (`internal/controllers/`)

### Kubernetes CRD Controller

A controller-runtime based controller that reconciles `TenantPolicy` custom resources to enforce tenant-level network isolation.

**Source:** `cmd/tenant-policy-controller/main.go`

### Reconciliation Flow

```
TenantPolicy CR Created/Updated
  → Reconciler reads CR
  → Lists pods with tenant.cloudbridge.io/id label
  → Creates/updates Calico GlobalNetworkPolicy
  → Network policy restricts traffic to/from tenant
```

### Controller Pattern

- **Framework:** `sigs.k8s.io/controller-runtime` (Kubebuilder pattern)
- **CRD:** Calico `GlobalNetworkPolicy` integration
- **Leader Election:** For HA deployments
- **Health Probes:** `/healthz` endpoint
- **Profiling:** pprof endpoint for debugging

### Configuration

- Namespace-scoped watching
- Cache resync period
- Leader election namespace
- Calico API registration

---

## CNI Plugin (`cmd/cni-plugin/`)

### Container Network Interface

A CNI plugin for Kubernetes pod networking that integrates with CloudBridge's IPAM and network management.

### Plugin Interface

```go
type CloudBridgeCNI struct {
    ipam    *ipam.Client
    network *network.Manager
}
```

**Standard CNI operations:**
- `Add()` — Allocate IP from IPAM, create network interface, apply network policies
- `Del()` — Remove interface, release IP

### CNI Configuration

```json
{
    "cniVersion": "0.3.1",
    "name": "cloudbridge",
    "type": "cloudbridge",
    "ipam": { "type": "cloudbridge-ipam" },
    "network": { "type": "cloudbridge-network" }
}
```

### Features

- IP allocation from IPAM service
- Network interface creation
- Tenant-aware network policies
- Rollback on failure
- Metrics collection

---

## Identity and Tenant Management

### Zitadel Client (`internal/zitadel/`)

| File | Purpose |
|---|---|
| `client.go` | OIDC client for Zitadel API |
| `jwks_provider.go` | JWKS key provider |
| `k8s_client.go` | K8s-aware Zitadel client |
| `tenant_manager.go` | Tenant lifecycle management |
| `tenant_network_client.go` | Tenant network configuration |
| `validator.go` | Token validation |

### Tenant Model (`internal/tenant/tenant.go`)

```go
type Tenant struct {
    ID          string
    Name        string
    Status      string
    MaxSubnets  int
    Config      map[string]string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Database Models (`internal/database/models.go`)

Core data models using PostgreSQL with JSONB fields for flexible configuration storage and UUID primary keys.

---

## Metrics and Observability

### Prometheus Metrics (`internal/metrics/`)

Comprehensive metrics across all subsystems:

| Category | Metrics |
|---|---|
| **Connections** | `connectionsTotal`, `activeConnections`, `connectionDuration` |
| **Tunnels** | `tunnelsCreatedTotal`, `activeTunnels`, `tunnelBytes{from,to}` |
| **Auth** | `authAttemptsTotal`, `tokenValidations`, `tokenValidationDuration` |
| **QUIC** | `quicConnections`, `quicStreams`, `quicBytesTransferred` |
| **WireGuard** | `wgTunnels`, `wgBytesTransferred`, `wgPeers` |
| **P2P** | `p2pConnections`, `p2pLatency`, `p2pDirectRatio` |
| **AI** | `anomaliesDetected`, `predictionAccuracy`, `routeOptimizations` |
| **Business** | `activeUsers`, `revenuePerHour`, `customerSatisfaction` |
| **System** | `memoryUsage`, `gcPauseDuration`, `networkLatency` |

---

## Compliance (`internal/compliance/`)

### Russian Regulatory Compliance

| Regulation | Code | Purpose |
|---|---|---|
| **152-ФЗ** | Federal Law | Personal Data Protection — data localization in Russia |
| **187-ФЗ** | Federal Law | Critical Information Infrastructure Security |
| **СОРМ** | SORM | Lawful interception requirements |

**`RussianCompliance` features:**
- Configuration validation for compliance requirements
- Connection-level validation (IP ranges, data flows)
- IP-based validation for Russian network ranges
- Consent management (152-ФЗ)
- Certified protection means (187-ФЗ)
- Traffic interception support (SORM)

---

## Source Map

| Component | Key Files |
|---|---|
| AI service | `/internal/ai/service.go`, `/cmd/ai-service/main.go` |
| AI analytics | `/cmd/ai-analytics/main.go` |
| Anomaly detection | `/internal/ai/anomaly/` |
| Load prediction | `/internal/ai/predictive/` |
| Route optimization | `/internal/ai/routing/` |
| Data pipeline | `/internal/ai/pipeline/` |
| ML models | `/internal/ai/ml/` |
| Cadence workflows | `/internal/cadence/workflows/` |
| Cadence client | `/internal/cadence/client.go` |
| Tenant controller | `/internal/controllers/tenant_policy_controller.go` |
| CNI plugin | `/cmd/cni-plugin/main.go` |
| Zitadel client | `/internal/zitadel/client.go` |
| Tenant model | `/internal/tenant/tenant.go` |
| Database models | `/internal/database/models.go` |
| Metrics | `/internal/metrics/metrics.go` |
| Compliance | `/internal/compliance/russian_compliance.go` |
