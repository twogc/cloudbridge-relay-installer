# Port Scheme (Canonical)

Single source of truth for CloudBridge Relay ports. Relay ports from `configs/config.json`; auxiliary microservice ports from **`cmd/*` defaults + `k8s/` ops manifests** (`k8s/deployments/`, `k8s/services/`). Runtime-fixed listeners in `cmd/relay/main.go` are noted where they differ from config keys.

**Host note:** Prometheus and other smarty Docker services on the dev host may use **9090** TCP — that is external to this installer; relay QUIC also uses 9090 **UDP** (no TCP conflict).

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
| IPAM | 8080 | TCP | `cmd/ipam-service`, `configs/cni/cloudbridge-cni.conf` | Internal; CNI `serviceUrl` |
| Tenant Policy Controller | 8080 / 8081 | TCP | `cmd/tenant-policy-controller` | metrics / health |
| AI Service | 8083 | TCP | `cmd/ai-service` (`AI_PORT`), `k8s/services/ai-service.yaml` | HTTP + `/metrics` on same port |
| AI Worker | 8085 | TCP | `cmd/ai-worker`, `k8s/services/ai-service.yaml` (`ai-worker`) | Health / metrics |
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
| `ai-service` | 8083 TCP | AI HTTP + `/metrics` (same port; `cmd/ai-service`) |
| `cloudbridge-relay-service` | 5551 TCP | Relay Prometheus metrics (`metrics.port`) |
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

## `k8s-manifests/` vs `k8s/` Manifest Drift

Canonical prod apply: `k8s-manifests/` via `scripts/deploy/deploy.sh`. Ops/component manifests: `k8s/deployments/`, `k8s/services/`.

**P0 stub fix (2026-07-09, k8s-prod-fix steps 2–3):** `05-deployments.yaml` и `06-services.yaml` больше не содержат `nginx:latest` / port 80 / AI **8080** stubs. Образы: `docker-registry:5000/cloudbridge-relay:latest`, `docker-registry:5000/ai-service:latest`, `ubercadence/server:v0.22.4`; `deploy.sh` → `load_images()` собирает relay/ai и пушит в in-cluster registry.

### CloudBridge Relay

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` | Canonical (`configs/config.json` + §Relay) |
|---|---|---|---|
| Image | `docker-registry:5000/cloudbridge-relay:latest` | `cloudbridge-relay:mvp-20251010-110859` | Built image from `docker/Dockerfile` |
| containerPort (declared) | http **9090**, quic **9090** UDP, wg **51820**, grpc **8444**, health **9091** | http **8080** (legacy), quic **9090** UDP, wg **51820**, metrics **9091**, grpc **8444** | HTTP API **9090**, metrics **5551**, P2P API **5552** |
| Health/readiness probe | **9090** `/health` | **9090** `/health` | HTTP API **9090** |
| Service `cloudbridge-relay*` | 9090 TCP/UDP, 51820 UDP, 8444 TCP, **5551** metrics | 9090 TCP/UDP, 51820 UDP, legacy metrics ports, **нет grpc** в части ops YAML | 9090, 51820, 8444, **5551** |
| prometheus.io/port | **5551** | often **8084** (ops residual) | Relay metrics **5551** |

**Вывод:** `k8s-manifests/05–06` prometheus + Service metrics → **5551**. Остаточный drift: ops `k8s/deployments/cloudbridge-relay-deployment.yaml` — legacy `containerPort: 8080` / annotation 8084.

### AI Service

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` | Canonical |
|---|---|---|---|
| Image | `docker-registry:5000/ai-service:latest` | `ai-service:latest` | Built `cmd/ai-service` image |
| containerPort | **8083** | **8083** | `AI_PORT` default **8083** |
| Health probe | **8083** | **8083** | **8083** |
| Service ports | http **8083** only | http **8083** only | HTTP + `/metrics` on **8083** |

**Вывод:** Stub drift устранён (k8s-prod-fix). Metrics port **8084** removed from Services (docs-consistency pass 3 / residual-port-fix) — scrape **8083**.

### Cadence

| Поле | `k8s-manifests/05–06` | `k8s/deployments|services` |
|---|---|---|
| Image | `ubercadence/server:v0.22.4` | `cadence-working.yaml` (отдельный файл) |
| Service ports | 7933, 7934, 7935, **7939** | 7933, 7934, 7935 (без 7939 в `cloudbridge-services.yaml`) |
| Probe | tcpSocket **7933** | varies by deployment variant |
| Runtime env | минимальный | полный (Cassandra/Ringpop) в `k8s/cadence/` |

**Вывод:** Образ и порты согласованы; `k8s-manifests` добавляет **7939** (worker). Cadence pod может потребовать env из ops-дерева `k8s/cadence/`.

---

## Known Documentation Drift (TODO)

| Item | Issue |
|---|---|
| `config/config.json` (dev) | Uses `server.port` **9090** — not the installer canonical `configs/config.json` |
| `docker/Dockerfile` EXPOSE | Lists **8082** (API fallback), not **5552** from installer config |
| ops `k8s/deployments/*relay*` | main + cert-manager aligned to 9090/5551 (pass 4) |
| `mws/*` OVA minimal | metrics **5551**, health **9090** (pass 4) |
| `k8s/deployments/cloudbridge-relay-p2p-deployment.yaml` | Still uses lab ports 8080/8082 — separate P2P profile |
| `k8s/services/ai-service.yaml` metrics **8084** | Declared Service port; `cmd/ai-service` serves `/metrics` on HTTP port **8083** |
| `k8s/deployments/cloudbridge-relay-deployment.yaml` http **8080** | Legacy `containerPort` in ops tree; `k8s-manifests/05` uses **9090** |
| `k8s-manifests/05` prometheus.io/port **8084** | Annotation drift vs relay metrics **5551** |
