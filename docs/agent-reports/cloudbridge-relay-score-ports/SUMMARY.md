# SUMMARY — cloudbridge-relay-score-ports

**Дата:** 2026-07-09  
**Миссия:** Выравнивание портовой схемы auxiliary microservices (IPAM, AI Service, AI Worker)  
**Статус:** ✅ Завершено (шаги 2–4); wiring SCORE allocator — **не в scope**, отложено

---

## Результат

| Цель | Статус |
|------|--------|
| Единый источник правды для портов | ✅ `openwiki/port-scheme.md` |
| IPAM / AI / AI Worker docs ↔ runtime | ✅ Синхронизировано (step 2) |
| `go build` ключевых cmd | ✅ OK (step 4) |
| Wiring `score_allocator_production` | ⏸ Не wired (0 prod callers) |

**Решение «одна правда»:** для auxiliary microservices authoritative — **`cmd/*` defaults + `k8s/` ops manifests** (не legacy SERVICE_CATALOG 8083/8085/8086).

---

## Таблица портов (верификация step 4)

### Auxiliary microservices

| Сервис | Порт | Протокол | `cmd/*` | `openwiki/port-scheme.md` | `k8s/` ops | CNI / config | Совпадение |
|--------|-----:|----------|---------|---------------------------|------------|--------------|:----------:|
| IPAM | **8080** | TCP | `cmd/ipam-service/main.go:54` | §Auxiliary | — | `configs/cni/cloudbridge-cni.conf` → `:8080` | ✅ |
| AI Service | **8083** | TCP | `cmd/ai-service/main.go` `AI_PORT` default | §Auxiliary | `k8s/deployments/ai-service-deployment.yaml`, `k8s/services/ai-service.yaml` | `configs/ai-config.yaml` | ✅ |
| AI Worker | **8085** | TCP | `cmd/ai-worker/main.go:46` | §Auxiliary | `k8s/services/ai-service.yaml` (ai-worker) | — | ✅ |

### CloudBridge Relay (canonical, `configs/config.json` + hardcoded)

| Listener | Порт | Протокол | Источник | Совпадение port-scheme |
|----------|-----:|----------|----------|:----------------------:|
| Main TCP | 5550 | TCP | `server.port` | ✅ |
| Prometheus metrics | 5551 | TCP | `metrics.port` | ✅ |
| P2P REST API | 5552 | TCP | `api.port` | ✅ |
| HTTP API | 9090 | TCP | `http_api.port` | ✅ |
| HTTPS API | 5553 | TCP | `https_api.port` | ✅ |
| Health / TLS tunnel | 9091 | TCP | `cmd/relay/tls_tunnel.go` | ✅ |
| QUIC transport | 9090 | UDP | `cmd/relay/main.go` | ✅ |
| Enhanced QUIC | 9092 | UDP | `cmd/relay/main.go` | ✅ |
| P2P QUIC | 5553 | UDP | `cmd/relay/main.go` | ✅ |
| STUN | 19302 | UDP | `cmd/relay/main.go` | ✅ |
| TURN | 3478 | UDP/TCP | `cmd/relay/main.go` | ✅ |
| DERP | 3479 | UDP | `cmd/relay/main.go` | ✅ |
| MASQUE | 8443 | TCP/UDP | `masque_integration.go` | ✅ |
| gRPC | 8444 | TCP | k8s env default | ✅ |
| WireGuard | 51820 | UDP | `wireguard.interface.listen_port` | ✅ |
| P2P API fallback | 8082 | TCP | `main.go` if `api.port` = 0 | ✅ |

---

## Wiring status (SCORE allocator)

| Компонент | Статус | Примечание |
|-----------|--------|------------|
| `score_allocator_production.go` | **keep (canonical)** | Единственная реализация после удаления v1/v2 |
| `relay_manager_hybrid.go` | interface готов | Принимает `network.TenantNetworkAllocator` |
| `NewSCORETenantNetworkAllocatorProduction` | **не wired** | 0 вызовов вне definition (grep step 4) |
| Следующий шаг | отдельная пачка | Подключить в `cmd/relay` → hybrid relay manager |

---

## Оставшийся drift (на момент score-ports)

| Item | Проблема |
|------|----------|
| `k8s-manifests/06-services.yaml` | *Тогда* AI stub **8080/9090** vs canonical **8083** |
| `k8s/services/ai-service.yaml` | metrics **8084** vs runtime `/metrics` **8083** |
| `config/config.json` (dev) | `server.port` **9090** — dev-only |
| `docs/network/PORTS_SUMMARY.md` | Может отставать от port-scheme |

### Superseded by later missions (2026-07-09)

| Open item (score-ports) | Статус |
|-------------------------|--------|
| `k8s-manifests/06` AI stub 8080/9090 | ✅ **k8s-prod-fix** — AI Service **8083**, metrics port **8084** |
| SCORE wiring | **ещё open** (0 prod callers) |
| ops metrics 8084 / dev config 9090 / PORTS_SUMMARY lag | **ещё open** (не P0) |

---

## Шаги миссии

| Шаг | Содержание | Отчёт |
|-----|------------|-------|
| 2 | MiMo analysis + sync `port-scheme.md`, `SERVICE_CATALOG.md` | `step02-write-mimo-port-alignment.md` |
| 4 | `go build`, port grep, audit updates, SUMMARY | `step04-verify-summary.md` |

---

## Проверки (step 4)

```bash
PATH="$HOME/.local/go-install/go/bin:$PATH" go build ./cmd/relay/ ./cmd/ipam-service/ ./cmd/ai-worker/ ./cmd/ai-service/
# exit 0

go version
# go1.25.12 linux/amd64
```

---

*Миссия cloudbridge-relay-score-ports завершена. Commit/push не выполнялись.*
