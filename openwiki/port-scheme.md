# Port Scheme (Canonical)

Single source of truth for CloudBridge Relay ports. Derived from `configs/config.json`, `docs/operations/SERVICE_CATALOG.md`, and `k8s-manifests/`. Runtime-fixed listeners in `cmd/relay/main.go` are noted where they differ from config keys.

**Config note:** `configs/config.json` contains duplicate `http_api` / `https_api` keys; JSON parsers keep the **last** value (`http_api.port` → 9090, `https_api.port` → 5553).

---

## CloudBridge Relay (`cmd/relay`)

| Listener | Port | Protocol | Config / code | Public |
|---|---:|---|---|---|
| Main TCP server | 5550 | TCP | `server.port` | Internal |
| Prometheus metrics | 5551 | TCP | `metrics.port` | Internal |
| P2P REST API | 5552 | TCP | `api.port` | Yes |
| HTTP API (health, REST) | 9090 | TCP | `http_api.port` (last key) | Via ingress |
| HTTPS API | 5553 | TCP | `https_api.port` (last key) | Via ingress |
| Health / TLS tunnel | 9091 | TCP | hardcoded (`cmd/relay/https.go`) | Internal |
| QUIC transport | 9090 | UDP | hardcoded (`main.go`) | Yes |
| Enhanced QUIC | 9092 | UDP | hardcoded (`main.go`) | Optional |
| P2P QUIC | 5553 | UDP | hardcoded (`main.go`) | Yes |
| STUN | 19302 | UDP | hardcoded (`main.go`) | Yes |
| Built-in TURN | 3478 | UDP/TCP | hardcoded (`main.go`) | Yes |
| DERP relay | 3479 | UDP | hardcoded (`main.go`) | Internal |
| MASQUE HTTP/3 | 8443 | TCP/UDP | `masque.proxy_url` / `masque_integration.go` | Yes |
| gRPC control plane | 8444 | TCP | `grpc.port` (k8s env default) | Internal |
| WireGuard | 51820 | UDP | `wireguard.interface.listen_port` | Yes |

**P2P API fallback:** if `api.port` is 0, relay uses **8082** (`main.go`).

---

## Auxiliary Microservices

| Service | Port(s) | Protocol | Primary source | Notes |
|---|---:|---|---|---|
| TURN (coturn, external pool) | 3478 | UDP/TCP | SERVICE_CATALOG, `configs/config.json` `turn_pool` | Standalone coturn; relay also embeds TURN on 3478 |
| IPAM | 8083 | TCP | SERVICE_CATALOG | `cmd/ipam-service` listens on **8080** — drift |
| Tenant Policy Controller | 8080 / 8081 | TCP | SERVICE_CATALOG | metrics / health |
| AI Service | 8085 | TCP | SERVICE_CATALOG | k8s Service **8080**; code default **8083** via `AI_PORT` |
| AI Worker | 8086 | TCP | SERVICE_CATALOG | `cmd/ai-worker` health on **8085** |
| QUIC Service (optional) | 8081 | TCP | SERVICE_CATALOG | metrics sidecar |
| Quantum Service | 8084 | TCP | SERVICE_CATALOG | disabled stub |
| WireGuard Service | 51820 | UDP | SERVICE_CATALOG | same as relay WireGuard |
| PostgreSQL | 5432 | TCP | SERVICE_CATALOG, k8s-manifests | Zitadel + app DB |
| Redis | 6379 | TCP | SERVICE_CATALOG | session cache |
| Prometheus | 9090 | TCP | SERVICE_CATALOG | scrapes relay; conflicts in name only with relay QUIC UDP |
| Cadence frontend | 7933 | TCP | SERVICE_CATALOG, k8s-manifests | |
| Cadence matching | 7934 | TCP | SERVICE_CATALOG, k8s-manifests | |
| Cadence history (svc) | 7935 | TCP | k8s-manifests | catalog lists 7935 as history |
| Cadence worker | 7939 | TCP | SERVICE_CATALOG, k8s-manifests | |
| Cadence Web | 8088 | TCP | SERVICE_CATALOG | optional UI |
| Cassandra | 9042 | TCP | k8s-manifests | Cadence backend |
| Zitadel | 8080 / 8081 | TCP | k8s-manifests/09-zitadel.yaml | API / gRPC-health |
| Container registry | 5000 | TCP | k8s-manifests/04-registry.yaml | in-cluster only |

---

## Kubernetes Service Exposure (`k8s-manifests/06-services.yaml`)

| K8s Service | Cluster Port | Maps to |
|---|---:|---|
| `cloudbridge-relay-service` | 9090 TCP/UDP | relay HTTP + QUIC |
| `cloudbridge-relay-service` | 51820 UDP | WireGuard |
| `cloudbridge-relay-service` | 8444 TCP | gRPC |
| `ai-service` | 8080 TCP | AI HTTP |
| `ai-service` | 9090 TCP | AI metrics |
| `cadence-service` | 7933–7939 TCP | Cadence gRPC |
| `cassandra-service` | 9042 TCP | Cassandra CQL |

---

## Internet / Firewall Minimum

Public-facing relay ports (see also `docs/network/PORTS_SUMMARY.md`):

| Port | Protocol | Purpose |
|---:|---|---|
| 5552 | TCP | P2P REST API |
| 5553 | UDP | P2P QUIC |
| 19302 | UDP | STUN |
| 3478 | UDP/TCP | TURN (external coturn or built-in) |
| 51820 | UDP | WireGuard (when enabled) |
| 8443 | TCP | MASQUE (when enabled) |
| 9090 | UDP | QUIC transport (recommended) |

---

## Known Documentation Drift (TODO)

| Item | Issue |
|---|---|
| `docs/operations/SERVICE_CATALOG.md` | Lists relay HTTP **8080** / HTTPS **8443** — legacy; installer uses `configs/config.json` scheme above |
| `config/config.json` (dev) | Uses `server.port` **9090** — not the installer canonical `configs/config.json` |
| `docker/Dockerfile` EXPOSE | Lists **8082** (API fallback), not **5552** from installer config |
| AI / IPAM ports | SERVICE_CATALOG vs k8s-manifests vs `cmd/*` defaults differ — table notes list actuals |
