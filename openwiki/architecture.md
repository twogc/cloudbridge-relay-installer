# Architecture

## System Overview

CloudBridge Relay is a **multi-service Go monorepo** with 13 cmd/ entrypoints (10 main services + 3 utilities/demos). The core relay server (`cmd/relay`) is a large aggregation point that composes 30+ internal packages into a single process. Auxiliary services (AI, IPAM, Cadence, etc.) run as separate microservices.

### Service Topology

```
                    ┌─────────────────────────────┐
                    │     Kubernetes Cluster       │
                    │       (VK Cloud)             │
                    │                              │
  Internet ────►   │  ┌──── Ingress (NGINX) ──┐  │
                    │  │  edge.2gc.ru          │  │
                    │  │  TLS termination       │  │
                    │  └────────┬───────────────┘  │
                    │           │                   │
                    │  ┌────────▼───────────────┐  │
                    │  │   CloudBridge Relay     │  │
                    │  │   (3 replicas)          │  │
                    │  │   :5550 TCP, :9090 QUIC │  │
                    │  │   see port-scheme.md    │  │
                    │  └──┬─────┬──────┬────────┘  │
                    │     │     │      │            │
                    │  ┌──▼──┐ ┌▼────┐ ┌▼────────┐│
                    │  │ AI  │ │IPAM │ │ Cadence  ││
                    │  │Svc  │ │Svc  │ │ Engine   ││
                    │  │:8080│ │:8083│ │ :7933    ││
                    │  └─────┘ └─────┘ └────┬─────┘│
                    │                        │      │
                    │  ┌─────────────────────▼────┐│
                    │  │  PostgreSQL / Cassandra   ││
                    │  │  (Persistence)            ││
                    │  └──────────────────────────┘│
                    │                              │
                    │  ┌─── Monitoring Stack ──────┐│
                    │  │ Prometheus → VictoriaMetrics│
                    │  │ Grafana / AlertManager     ││
                    │  └──────────────────────────┘│
                    └─────────────────────────────┘
```

### Binary Inventory

| Binary | Source | Purpose |
|---|---|---|
| `cloudbridge-relay` | `cmd/relay/` | Core relay server — P2P, QUIC, WireGuard, NAT, MASQUE |
| `api` | `cmd/api/` | REST API for tunnel management |
| `ai-service` | `cmd/ai-service/` | AI inference (anomaly detection, load prediction) |
| `ai-analytics` | `cmd/ai-analytics/` | Analytics engine with Cloudflare AI integration |
| `ai-worker` | `cmd/ai-worker/` | Cadence worker for async AI tasks |
| `quic-service` | `cmd/quic-service/` | Experimental QUIC metrics endpoints |
| `cadence-service` | `cmd/cadence-service/` | Workflow orchestration HTTP API |
| `ipam-service` | `cmd/ipam-service/` | IP address management REST API |
| `tenant-policy-controller` | `cmd/tenant-policy-controller/` | K8s CRD controller for tenant network policies |
| `cni-plugin` | `cmd/cni-plugin/` | CNI plugin for K8s pod networking |
| `kubectl-zitadel` | `cmd/kubectl-zitadel/` | kubectl plugin for Zitadel IAM management |
| `quantum-service` | `cmd/quantum-service/` | Post-quantum cryptography (currently disabled) |
| `zerotrust-demo` | `cmd/zerotrust-demo/` | Zero-trust connectivity demo client |

---

## Core Relay Server

The relay server (`cmd/relay/server.go`) is the central component. Its `Server` struct composes:

**Transport Layer:**
- `QUICManager` / `EnhancedQUICManager` — QUIC transport with experimental features
- `TCPServer` / `TLSConfig` — TCP fallback
- `WebSocketFallback` — WebSocket transport
- `MASQUEProxy` — RFC 9297 capsule-based proxy
- `gRPCServer` — gRPC transport with protobuf definitions

**Networking:**
- `TUNManager` / `WireGuardManager` — VPN tunnel management
- `STUNServer` / `TURNServer` / `DERPServer` — NAT traversal
- `ICEAgent` — Interactive Connectivity Establishment
- `P2PIntegration` / `P2PManager` — Peer-to-peer orchestration
- `ProtocolEngine` — Auto-switching between QUIC/HTTP2/WebSocket

**Security:**
- `ZitadelIntegration` — Canonical IAM: OIDC/JWT validation via JWKS (Keycloak/Django removed)
- `CircuitBreaker` — Failure detection and isolation
- `RateLimiter` — Request throttling
- `QuantumService` — Post-quantum cryptography

**Orchestration:**
- `CadenceClient` — Workflow engine integration
- `AIClient` — AI/ML inference client
- `HealthChecker` — Comprehensive health probes
- `KubernetesProxy` — K8s API proxy with RBAC

**Observability:**
- Prometheus metrics (connections, tunnels, tokens, bandwidth)
- Structured Zap logging
- QLog tracing for QUIC protocol debugging

---

## Transport Protocols

### Protocol Engine (`internal/protocol/`)

The protocol engine implements **automatic transport switching** based on connection quality:

```
Priority Order: QUIC → HTTP/2 → WebSocket

Each protocol tracks: latency, throughput, error rate, success rate
Switch triggers: error rate > threshold, latency spike, throughput drop
```

- `ProtocolSelector` interface with `Current()`, `Fallback()`, `ShouldSwitch()`, `SelectBest()`
- Per-protocol Prometheus metrics: `protocolLatency`, `protocolErrors`, `protocolSwitches`
- Seamless migration for connected clients

### QUIC Transport (`internal/quic/`)

Enhanced QUIC with experimental features:

| Feature | Description |
|---|---|
| **BBRv2** | Modern congestion control algorithm |
| **ACK Frequency** | Reduced ACK overhead for high-throughput connections |
| **FEC** | Forward Error Correction with configurable redundancy |
| **QUIC Bit Greasing** | RFC 9287 protocol version obfuscation |
| **QLog** | Protocol-level tracing for debugging |
| **Per-tenant isolation** | Bandwidth limits, connection limits per tenant |

### MASQUE Protocol (`internal/masque/`)

HTTP Capsule encapsulation per RFC 9297. Used as a fallback when QUIC DATAGRAM frames are unavailable:

- Capsule types: `Datagram` (0x00), `Close` (0x01)
- Operates over HTTP/3 streams
- Thread-safe bidirectional capsule handling

### TCP Fallback (`cmd/relay/tcp_fallback.go`, `cmd/relay/tls_tunnel.go`)

TLS-encrypted TCP tunnels as last-resort transport when QUIC and WebSocket are unavailable.

---

## Authentication

### Zitadel IAM (`internal/auth/`, `internal/zitadel/`)

**Zitadel** is the sole identity provider. OIDC/JWT tokens are validated via JWKS (RS256).

> **Removed (legacy):** Keycloak and Django auth providers were deleted from the codebase; do not reintroduce them.

### Token Structure (`internal/auth/relay_claims.go`)

`RelayClaims` extends standard JWT claims with relay-specific fields:

```go
type RelayClaims struct {
    jwt.RegisteredClaims
    OrgID          string
    TenantID       string
    ServiceName    string
    Permissions    []string
    Roles          []string
    RelayServerID  string
    ProtocolType   string  // "permanent" or "session"
    NetworkConfig  NetworkConfig  // allowed subnets, max connections, bandwidth
}
```

### Validation Flow

1. JWT token extracted from request (header or query param)
2. `ZitadelValidator` verifies signature via RSA public key (RS256)
3. Claims parsed into `RelayClaims` with tenant/org/permissions
4. `token.Validator` provides fast HMAC-based validation with 5-minute cache
5. Permission checks: `HasPermission()`, `HasRole()`, `IsRelayToken()`

### Key Types

- `ErrExpired`, `ErrNotYetValid` — Token lifetime
- `ErrInvalidIssuer`, `ErrInvalidAudience` — Token origin
- `ErrPermissionDenied`, `ErrRoleDenied` — Authorization
- `ErrNotInitialized` — System startup state

---

## Configuration System

### Main Config (`config/config.go`)

Two-phase JSON parsing:

1. **Raw parse**: JSON → `configRaw` (time.Duration fields as strings)
2. **Type conversion**: `configRaw` → `Config` (strings parsed to `time.Duration`)

Major configuration sections:

| Section | Key Settings |
|---|---|
| `server` | Port, timeouts, request limits |
| `zitadel` | Issuer, client ID, JWKS URL, allowed algorithms |
| `wireguard` | Interface name, listen port, key paths |
| `quic` | Max streams, keep-alive, idle timeout, congestion control |
| `enhancedQuic` | BBRv2, ACK frequency, FEC, greasing |
| `ice` / `stun` / `turn` | Server addresses, credentials, pool sizes |
| `masque` | Listen address, timeout, max connections |
| `protocol` | Switch threshold, quality weights |
| `ai` | Model paths, inference endpoint, cache TTL |
| `cadence` | Host, domain, task list |
| `kubernetes` | Cluster proxy, RBAC policies |
| `grpc` | Max concurrent streams, keepalive |
| `security` | mTLS, cert paths, cipher suites |

### Service Config (`internal/config/config.go`)

Simpler config for auxiliary services (ai-worker, cadence-service):

```go
type Config struct {
    Relay   RelayConfig
    Cadence CadenceConfig
    Tenant  TenantConfig
    Tunnel  TunnelConfig
}
```

Includes path traversal prevention via `isValidConfigPath()`.

---

## Inter-Service Communication

| Pattern | Used Between |
|---|---|
| **REST API** | Relay ↔ AI service, IPAM, Cadence service |
| **gRPC** | Relay ↔ K8s proxy, WireGuard management |
| **Cadence RPC** | Relay ↔ AI worker (async workflows) |
| **HTTP/2** | Client connections (protocol engine) |
| **QUIC** | Client connections (primary transport) |
| **WebSocket** | Client connections (fallback transport) |
| **Prometheus scrape** | All services → monitoring stack |

---

## Key Design Patterns

| Pattern | Where |
|---|---|
| **Composition** | `Server` struct in `cmd/relay/server.go` (30+ components) |
| **Strategy** | `ProtocolSelector` for transport switching |
| **Circuit Breaker** | `internal/circuitbreaker/` — failure isolation |
| **Manager** | P2P, WireGuard, Tunnel, Connection, ICE managers |
| **Workflow** | Cadence activities and workflows for multi-step ops |
| **Repository** | `internal/database/` — data access abstraction |
| **Middleware** | Auth, rate limiting, security headers, CORS |
| **Health Probes** | `/health`, `/ready` on all services |

---

## Source Map

| Component | Key Files |
|---|---|
| Relay server | `/cmd/relay/main.go`, `/cmd/relay/server.go` |
| API router | `/internal/api/router.go`, `/internal/api/server.go` |
| Config system | `/config/config.go`, `/internal/config/config.go` |
| Auth | `/internal/auth/zitadel_validator.go`, `/internal/auth/relay_claims.go` |
| Protocol engine | `/internal/protocol/protocol_engine.go`, `/internal/protocol/selector.go` |
| QUIC | `/internal/quic/config.go`, `/internal/quic/enhanced.go` |
| MASQUE | `/internal/masque/capsule.go` |
| Protobuf | `/proto/wireguard.proto` |
