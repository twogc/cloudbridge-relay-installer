# 🚀 **CloudBridge Relay - Брифинг для разработчика**

## **Краткое резюме проекта**

**CloudBridge Relay** - это QUIC-first + MASQUE сеть для создания "почти LAN" соединений между удаленными серверами клиентов с поддержкой CONNECT-UDP/CONNECT-IP и ICE/STUN/TURN слоя.

### **🎯 Цель проекта:**
Реализовать QUIC-first + MASQUE архитектуру с тремя ключевыми компонентами:

1. **QUIC-first + MASQUE**: реализовать CONNECT-UDP (далее CONNECT-IP), использовать HTTP Datagrams; поддержать downgrade на WS/TCP.
2. **ICE/STUN/TURN слой**: свой STUN, управляемый TURN-пул (в т.ч. TCP-alloc при необходимости), метрики NAT-типов/успеха.
3. **WG-профиль для site-to-site/legacy**: с чёткими ограничениями и сценариями роуминга; держать собственные релеи как «последний шанс».

### **🏗️ Архитектура:**
```
┌─────────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay Network                    │
├─────────────────────────────────────────────────────────────────┤
│  QUIC-first + MASQUE (Port 443/HTTP-3)                        │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   CONNECT-UDP   │    │   CONNECT-IP    │                    │
│  │   HTTP Datagrams│    │   Full IP VPN   │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  ICE/STUN/TURN Layer (NAT Traversal)                          │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Direct P2P    │    │   TURN Relay    │                    │
│  │   STUN/TURN     │    │   TCP-alloc     │                    │
│  └─────────────────┘    └─────────────────┘                    │
├─────────────────────────────────────────────────────────────────┤
│  WG-профиль (Site-to-Site/Legacy)                              │
│  ┌─────────────────┐    ┌─────────────────┐                    │
│  │   Legacy VPN    │    │   Last Resort   │                    │
│  │   Roaming       │    │   Relay         │                    │
│  └─────────────────┘    └─────────────────┘                    │
└─────────────────────────────────────────────────────────────────┘
```

## 📊 **Текущее состояние (по результатам тестирования DevOps)**

### **ЧТО УЖЕ РАБОТАЕТ:**
- **gRPC сервер (TLS 1.3):** порт 8444 (NodePort 30844), Control/Tunnel/Heartbeat — OK
- **JWT аутентификация:** проверка подписи и авторизация — OK
- **Метрики/health:** порт 5551, endpoints `/metrics` и `/health` — OK
- **Kubernetes:** Deployment/Service/NetworkPolicy обновлены под gRPC 8444 — OK
- **Туннели через gRPC:** создание и heartbeat — OK
- **MASQUE/Datagrams/STUN/TURN:** не включены (в roadmap)

### **🆕 Актуальный Roadmap (Q4 2025)**
| Приоритет | Feature | Статус | Следующие шаги |
|-----------|---------|--------|---------------|
| 🔥 | MASQUE / CONNECT-UDP & CONNECT-IP | **Design** готов | PoC на базе `quic-go` + unit-benchmarks |
| 🔥 | QUIC Datagrams (HTTP Datagrams) | **0%** | Включить `EnableDatagrams=true`, реализовать mux-layer |
| 🔥 | NAT Telemetry (ICE/STUN/TURN) | **10%** | Scaffolding; добавить Prometheus-метрики success-rate и типов NAT |
| ⚡ | WireGuard profile (legacy site-to-site) | **In Progress** | CLI опция `--wg-profile`, Helm secret c конфигом |
| ⚡ | TCP/WebSocket fallback | **TCP готов / WS stub** | Feature-flag `--tcp-fallback`, e2e/CI тесты |
| 💤 | QUIC 0-RTT optimisation | **Backlog** | Анализ CVE + optional enable per-tenant |

> Обновлено: 26 сен 2025 после интеграции gRPC + TLS 1.3, полноценного клиента и mesh-routing.

## 🎯 **Основные задачи для разработчика**

### **1. QUIC-first + MASQUE (Приоритет: ВЫСОКИЙ)**
Реализовать CONNECT-UDP и CONNECT-IP с HTTP Datagrams:
```go
// Целевая функциональность
func (s *Server) handleMASQUERelay(conn *quic.Conn, peerID string, data []byte) error {
    // 1. Определить тип туннеля (CONNECT-UDP/CONNECT-IP)
    // 2. Обработать HTTP Datagrams
    // 3. Проверить изоляцию tenant
    // 4. Переслать через соответствующий протокол
}
```

### **2. ICE/STUN/TURN слой (Приоритет: ВЫСОКИЙ)**
Реализовать NAT traversal с метриками:
```go
// ICE/STUN/TURN endpoints
POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/ice-candidates
GET /api/v1/tenants/{tenant_id}/peers/{peer_id}/nat-metrics
// Свой STUN сервер + управляемый TURN-пул
```

### **3. WG-профиль для site-to-site/legacy (Приоритет: СРЕДНИЙ)**
Поддержка legacy VPN с ограничениями:
```go
// WG-профиль с чёткими ограничениями
type WGProfile struct {
    Enabled bool
    MaxConnections int
    RoamingScenarios []string
    LastResortRelay bool
}
```

### **4. Downgrade на WS/TCP (Приоритет: СРЕДНИЙ)**
Поддержка fallback для блокирующих сетей:
```go
// Автоматический downgrade
func (s *Server) handleProtocolDowngrade(conn *quic.Conn) error {
    // 1. Попробовать QUIC
    // 2. Fallback на WebSocket
    // 3. Fallback на TCP
}
```

## 🏗️ **Архитектура мультитунельности**

### **Изоляция по уровням:**
1. **JWT токены** содержат `tenant_id` для идентификации
2. **Сетевые пространства** изолированы по tenant (10.100.77.0/24, 10.100.78.0/24)
3. **P2P соединения** только внутри одного tenant
4. **Маршрутизация** на основе tenant_id

### **Пример использования:**
```bash
# Компания A (tenant-216420165)
Server A1 (10.100.77.10) ←→ Server A2 (10.100.77.11)

# Компания B (tenant-123456789) 
Server B1 (10.100.78.10) ←→ Server B2 (10.100.78.11)

# Изоляция: A и B не могут общаться друг с другом
```

## 🐳 **Инструкция по сборке Minikube**

### **Быстрый старт:**
```bash
# 1. Установить Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# 2. Запустить Minikube
minikube start --driver=docker --memory=4096 --cpus=2

# 3. Собрать и загрузить образ
docker build -t cloudbridge-relay:quic-ice-v2 .
minikube image load cloudbridge-relay:quic-ice-v2
    
# 4. Применить конфигурации
kubectl create namespace cloudbridge
kubectl apply -f k8s/deployment.yaml -n cloudbridge
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# 5. Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

### **Тестирование:**
```bash
# Port forwarding для тестирования
kubectl port-forward service/cloudbridge-relay-external 8080:8080 -n cloudbridge &
kubectl port-forward service/cloudbridge-relay-external 8082:8082 -n cloudbridge &

# Тестировать API
curl http://localhost:8080/health
curl http://localhost:8082/health
```

## **План разработки**

### **Этап 1: QUIC-first + MASQUE (1-2 дня)**
- [ ] Реализовать CONNECT-UDP протокол
- [ ] Добавить HTTP Datagrams поддержку
- [ ] Создать MASQUE Manager
- [ ] Протестировать базовую MASQUE связность

### **Этап 2: ICE/STUN/TURN слой (1-2 дня)**
- [ ] Реализовать свой STUN сервер
- [ ] Создать управляемый TURN-пул
- [ ] Добавить TCP-alloc поддержку
- [ ] Собирать метрики NAT-типов/успеха

### **Этап 3: WG-профиль для site-to-site/legacy (2-3 дня)**
- [ ] Реализовать WG-профиль с ограничениями
- [ ] Добавить сценарии роуминга
- [ ] Создать last resort relay
- [ ] Протестировать legacy VPN совместимость

### **Этап 4: Downgrade на WS/TCP (1-2 дня)**
- [ ] Реализовать WebSocket fallback
- [ ] Добавить TCP fallback
- [ ] Протестировать блокирующие сети
- [ ] Оптимизировать производительность

## 🧪 **Критерии успеха**

### **Функциональность:**
- CONNECT-UDP и CONNECT-IP работают через MASQUE
- HTTP Datagrams поддерживаются
- ICE/STUN/TURN слой с метриками NAT
- WG-профиль для legacy VPN
- Downgrade на WS/TCP при блокировке

### **Производительность:**
- Поддержка 100+ одновременных MASQUE соединений
- Memory usage < 500MB
- CPU usage < 50%
- Поддержка 10+ tenant одновременно

### **Стабильность:**
- Автоматический fallback на WS/TCP
- Graceful shutdown MASQUE соединений
- Восстановление соединений после сбоев

## 📚 **Документация**

### **Основные документы:**
- **Техническое задание:** `DEVELOPER_TECHNICAL_SPECIFICATION.md`
- **Архитектура мультитунельности:** `MULTI_TENANT_ARCHITECTURE.md`
- **Инструкция по сборке:** `MINIKUBE_BUILD_GUIDE.md`
- **План доработки:** `P2P_ROUTING_DEVELOPMENT_PLAN.md`

### **Текущие конфигурации:**
- **Kubernetes:** `k8s/deployment.yaml`, `k8s/cloudbridge-relay-external-service.yaml`
- **Конфигурация:** `configs/config.json`
- **JWT токены:** `official-token.txt`

## 🚀 **Следующие шаги**

### **Немедленно (сегодня):**
1. Изучить текущий код в `cmd/relay/`
2. Понять структуру JWT токенов
3. Проанализировать существующие API endpoints
4. Создать базовые структуры данных для MASQUE

### **Этап 1 (1-2 дня):**
1. Реализовать MASQUE Manager
2. Добавить CONNECT-UDP поддержку
3. Реализовать HTTP Datagrams
4. Протестировать базовую MASQUE связность

### **Этап 2 (1-2 дня):**
1. Реализовать ICE/STUN/TURN слой
2. Добавить метрики NAT-типов
3. Протестировать NAT traversal

### **Этап 3 (2-3 дня):**
1. Реализовать WG-профиль для legacy
2. Добавить сценарии роуминга
3. Протестировать site-to-site соединения

### **Этап 4 (1 день):**
1. Реализовать downgrade на WS/TCP
2. Добавление метрик и мониторинга
3. Финальное тестирование

## 📞 **Поддержка**

- **Текущий статус:** Все порты работают, нужна доработка QUIC-first + MASQUE архитектуры
- **Тестирование:** DevOps уже протестировал базовую инфраструктуру
- **Результаты тестирования:** Клиенты регистрируются, но нужна реализация CONNECT-UDP/CONNECT-IP и ICE/STUN/TURN слоя

---

**🎯 Цель: Создать QUIC-first + MASQUE сеть с CONNECT-UDP/CONNECT-IP, ICE/STUN/TURN слоем и WG-профилем для legacy VPN!**

**Задача: Реализовать QUIC-first + MASQUE архитектуру с тремя ключевыми компонентами: CONNECT-UDP/CONNECT-IP, ICE/STUN/TURN слой с метриками NAT, и WG-профиль для site-to-site/legacy с fallback на WS/TCP.**

## 🔒 Future Risks & Mitigations (Mesh / QUIC)

| # | Risk | Evidence | Technical Mitigations |
|---|------|-----------|-----------------------|
| 1 | High CPU load with QUIC/TLS (user-space crypto) | Fastly 2023 «QUIC vs TCP»; arXiv 2310.09423; SIGCOMM’20 | AES-NI/ChaCha offload, UDP GSO/GRO, XDP/DPDK fast-path, NUMA pinning, per-tenant rate-limit |
| 2 | Mesh over-connect (O(N²) links) | Princeton «Scalable Routing Overlay Networks» | Sparse-mesh pruning, `max_peers` в JWT, latency-aware peer selection |
| 3 | Troubleshooting complexity (route loops, flaps) | Overlay/Underlay study (DIV-2020) | Control-plane telemetry, OpenTelemetry tracing, alerts on asymmetric routes |
| 4 | UDP blocks / Sym-NAT / DPI | RFC 9000; enterprise firewall tests | TCP-fallback (443), STUN/ICE, QUIC disguise, disable 0-RTT by default |
| 5 | Overhead of many concurrent streams | Google QUIC scalability papers | Stream pool & priority, BBR v2, adaptive windows, hard caps per peer |

Backlog items (Q4 2025):
1. eBPF CPU profiling & syscall heat-maps.
2. XDP fast-path PoC on ≥10 Gbps relays.
3. Sparse-mesh scheduler in `internal/p2p/relay_manager.go`.
4. OpenTelemetry exporter in `internal/metrics`.
5. Feature-flag `relay.tcp_fallback.enabled` (TLS-over-TCP mode).
