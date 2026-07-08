# Networking Domain

This page covers the core networking subsystems: P2P connectivity, NAT traversal, WireGuard VPN, IP address management, protocol selection, and the MASQUE protocol.

**Ports:** canonical service → port mapping is in **[Port Scheme](port-scheme.md)** (`configs/config.json` + k8s-manifests). Relay listeners include **5550** (TCP), **5552** (P2P API), **5553** (P2P QUIC UDP), **9090** (QUIC UDP), **19302** (STUN), **3478** (TURN), **51820** (WireGuard), **8443** (MASQUE).

---

## P2P Networking (`internal/p2p/`)

### Architecture

The P2P system manages direct peer-to-peer connections through mesh topology, hybrid routing, and site-to-site VPN.

```
P2PManager
├── P2PMeshTopologyInterface    — Mesh overlay management
├── PeerDiscoveryInterface      — Service discovery
├── HybridRouterInterface       — Multi-path routing
├── SiteToSiteManagerInterface  — Site-to-site VPN
└── CadenceP2PIntegrationInterface — Workflow-driven P2P ops
```

### Peer Model (`internal/p2p/types.go`)

The `Peer` struct represents a node in the P2P network:

| Field | Purpose |
|---|---|
| `TenantID`, `OrgID`, `DeviceID` | Identity and isolation |
| `ConnectionType` | Direct, relay, or MASQUE |
| `RelaySessionID` | Session tracking |
| `QuicConfig` | Per-peer QUIC settings (endpoint, streams, timeouts) |
| `MeshConfig` | Mesh overlay (auto-discovery, routing mode, encryption) |
| `PeerWhitelist` | Access control (allowed peers, auto-approve, max peers) |
| `NetworkConfig` | Subnet, DNS servers, MTU |
| `ICECandidates` | ICE candidate list for NAT traversal |
| `AllocatedIP` | IPAM-assigned address |
| `LatencyMS` | Connection quality metric |

### Connection States

```
New → Connecting → Connected → Failed
                          ↓
                       Closed
```

### Connection Manager (`internal/p2p/connection_manager.go`)

Manages the full lifecycle:

- **Registration** — Peer announces presence
- **Establishment** — ICE negotiation, QUIC handshake
- **Tracking** — Bytes transferred, latency, state
- **Closure** — Graceful teardown with cleanup

Key types:
- `P2PConnection` — Full connection state including ICE candidate pair, protocol, byte stats
- `RelayConnection` — Relay-specific connection wrapping a `StreamBridge`
- `StreamBridge` — Bidirectional stream bridge (QUIC or WebSocket) through the relay

### StreamBridge (`internal/relay/stream_bridge.go`)

When direct P2P fails (symmetric NAT), the `StreamBridge` connects two streams through the relay server, forwarding data bidirectionally.

### Cadence Integration (`internal/p2p/cadence_integration.go`)

P2P operations can be orchestrated as Cadence workflows for durability:
- Multi-step connection establishment
- Automatic retry and recovery
- Timeout management

---

## NAT Traversal (`internal/nat/`)

### ICE Agent (`internal/nat/ice_agent.go`)

Full ICE (Interactive Connectivity Establishment) implementation:

**Components:**
- `STUNClient` — Server-reflexive candidate gathering
- `TURNPool` / `TURNPoolManager` — Relay candidate allocation
- `ConnectivityChecker` — Candidate pair connectivity checks

**State Machine:**
```
New → Gathering → Gathered → Checking → Connected → Completed
                                                 ↘ Failed
                                              Disconnected
```

**Candidate Types:**
| Type | Code | Source |
|---|---|---|
| host | host | Local network interface |
| srflx | srflx | STUN server reflection |
| prflx | prflx | Peer reflexive (discovered during checks) |
| relay | relay | TURN relay allocation |

### Built-in TURN Server (`internal/nat/turn_server_enhanced.go`)

Extended TURN with:
- Connection pooling
- Rate limiting
- Metrics collection

### STUN Client (`internal/nat/stun_client.go`)

STUN binding request/response handling for server-reflexive candidate discovery.

### DERP Server (`internal/nat/derp_server.go`)

DERP (Detour Encrypted Relay Protocol) style relay for WebRTC-like connections.

### Connectivity Checking (`internal/nat/connectivity_checker_enhanced.go`)

Enhanced connectivity checks with:
- Parallel candidate pair testing
- Priority-based selection
- NAT type detection

### TURN Credentials (`internal/nat/turn_credentials_pion.go`)

Time-limited TURN credentials using HMAC-based short-term credential mechanism. Integrates with the Pion ICE library.

---

## WireGuard VPN (`internal/wireguard/`)

### Config Manager (`internal/wireguard/config_manager.go`)

Generates WireGuard configurations:

- **Server config**: Interface name, private key, address (`10.0.0.1/24`), listen port (`51820`), peer list
- **Per-tenant client config**: Private/public key pair, endpoint, allowed IPs
- Saves configs to disk; supports `wg-quick` startup

### Tenant Manager (`internal/wireguard/tenant_manager.go`)

Manages WireGuard tunnels for VIP tenants:

**`WireGuardTenant` per-tenant state:**
- Public/private key pair (generated with `crypto/rand`)
- Allowed IPs, endpoint, port
- Persistent keepalive setting
- Active status, connection stats, timestamps
- Persisted via `database.Repository`

**`WireGuardPeer` (per-peer):**
- Public key, allowed IPs, endpoint, keepalive

### Integration Flow

```
Tenant Registration
  → WireGuardTenantManager creates keys
  → ConfigManager generates server + client configs
  → Config saved to disk
  → Interface started via wg-quick
  → K8s DaemonSet (k8s/wireguard-daemonset.yaml) distributes configs
```

---

## IP Address Management (`internal/ipam/`)

### Service (`internal/ipam/service.go`)

REST API service (Gin framework) backed by PostgreSQL:

**Tenant Operations:**
- `POST /api/v1/ipam/tenants` — Create tenant with max subnets
- `GET /api/v1/ipam/tenants/:id` — Get tenant info
- `DELETE /api/v1/ipam/tenants/:id` — Remove tenant

**Subnet Operations:**
- `POST /api/v1/ipam/tenants/:id/allocate` — Allocate CIDR subnet (e.g., /24, /25)
- `POST /api/v1/ipam/tenants/:id/deallocate` — Release subnet
- `GET /api/v1/ipam/subnets` — List all allocations
- `GET /api/v1/ipam/stats` — Utilization statistics
- `GET /api/v1/ipam/metrics` — Prometheus metrics

### Database Layer (`internal/database/`)

- `IPAMRepository` — Subnet CRUD with CIDR parsing
- `PeerIPAMRepository` — Per-peer IP allocation tracking
- Models: JSONB fields for flexible config storage, UUID primary keys

---

## Protocol Engine (`internal/protocol/`)

### Auto-Switching

The protocol engine monitors connection quality and automatically switches transports:

```
QUIC (preferred) → HTTP/2 → WebSocket (fallback)
```

**Selection Logic (`ProtocolSelector`):**
- `Current()` — Active transport
- `Fallback()` — Next transport in priority
- `ShouldSwitch()` — Quality below threshold
- `SelectBest()` — Re-evaluate all protocols

**Per-protocol Metrics:**
- Latency (histogram)
- Throughput (gauge)
- Error rate (counter)
- Success rate (counter)

**Switch Triggers:**
- Error rate exceeds configurable threshold
- Latency spike above baseline
- Throughput drop below minimum
- Connection failure

### Prometheus Metrics

```
protocolLatency{protocol="quic|http2|websocket"}  — histogram
protocolErrors{protocol="..."}                     — counter
protocolSwitches{from="...",to="..."}              — counter
protocolSuccess{protocol="..."}                    — counter
```

---

## MASQUE Protocol (`internal/masque/`)

### Capsule Handler (`internal/masque/capsule.go`)

HTTP Capsule encapsulation per RFC 9297. Used when QUIC DATAGRAM frames are unavailable:

**Capsule Types:**
| Type | Value | Purpose |
|---|---|---|
| Datagram | 0x00 | Data encapsulation |
| Close | 0x01 | Stream termination |

**Integration:**
- Operates over HTTP/3 streams (`http3.Stream`)
- Uses `quicvarint` for capsule type/length encoding
- Bidirectional capsule reading with EOF handling
- Thread-safe with mutex protection

### Proxy (`internal/masque/proxy.go`)

Full MASQUE proxy with CONNECT-IP and CONNECT-UDP support:

- `ConnectIP` handler — IP-in-IP encapsulation
- Datagram forwarding between clients
- Metrics collection for capsule operations

### Metrics (`internal/masque/metrics.go`)

Prometheus metrics for MASQUE operations: capsules sent/received, errors, latency.

---

## gRPC Transport (`internal/transport/grpc/`)

### Services

| Service | Proto | Purpose |
|---|---|---|
| `ControlService` | relay/v1/control.proto | Relay control plane |
| `HeartbeatService` | relay/v1/heartbeat.proto | Connection health |
| `TunnelService` | relay/v1/tunnel.proto | Tunnel management |
| `WireGuardService` | wireguard.proto | VPN tunnel CRUD |

### WireGuard gRPC API (`proto/wireguard.proto`)

```protobuf
service WireGuardService {
    rpc CreateTunnel(CreateTunnelRequest) returns (CreateTunnelResponse);
    rpc DeleteTunnel(DeleteTunnelRequest) returns (DeleteTunnelResponse);
    rpc GetTunnel(GetTunnelRequest) returns (GetTunnelResponse);
    rpc ListTunnels(ListTunnelsRequest) returns (ListTunnelsResponse);
    rpc GetStatus(GetStatusRequest) returns (GetStatusResponse);
    rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
    rpc UpdateConfig(UpdateConfigRequest) returns (UpdateConfigResponse);
}
```

All requests include `tenant_id` and `request_id` for multi-tenant isolation and tracing.

### Middleware

- `RateLimiter` — Per-client rate limiting
- `RBACMiddleware` — Role-based access control

---

## Connection Management

### Connection Pool (`internal/connection/`)

- `Pool` — HTTP/2 connection reuse with health checks
- `HTTP2Reuse` — HTTP/2 multiplexing optimization
- `Backoff` — Exponential backoff for reconnection
- `Retry` — Configurable retry policies
- `Performance` — Connection timing and profiling

### Balancer (`internal/balancer/`)

Load balancing across relay instances with:
- P2C (Power of Two Choices) algorithm
- Latency-aware routing
- Health-weighted distribution

### Circuit Breaker (`internal/circuitbreaker/`)

Three-state circuit breaker (Closed → Open → Half-Open):
- Failure threshold triggers open state
- Recovery timeout triggers half-open
- Success in half-open closes the circuit
- Per-subsystem metrics

---

## Source Map

| Component | Key Files |
|---|---|
| P2P types | `/internal/p2p/types.go` |
| P2P manager | `/internal/p2p/manager.go` |
| P2P connections | `/internal/p2p/connection_manager.go` |
| ICE agent | `/internal/nat/ice_agent.go` |
| TURN server | `/internal/nat/turn_server_enhanced.go` |
| STUN client | `/internal/nat/stun_client.go` |
| WireGuard config | `/internal/wireguard/config_manager.go` |
| WireGuard tenants | `/internal/wireguard/tenant_manager.go` |
| IPAM service | `/internal/ipam/service.go` |
| Protocol engine | `/internal/protocol/protocol_engine.go` |
| Protocol selector | `/internal/protocol/selector.go` |
| MASQUE capsules | `/internal/masque/capsule.go` |
| MASQUE proxy | `/internal/masque/proxy.go` |
| gRPC server | `/internal/transport/grpc/server.go` |
| Stream bridge | `/internal/relay/stream_bridge.go` |
| Connection pool | `/internal/connection/pool.go` |
| Circuit breaker | `/internal/circuitbreaker/circuit_breaker.go` |
