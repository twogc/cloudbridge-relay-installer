# CloudBridge Client-Server Integration Analysis

**Дата**: 2025-10-13
**Версия**: 1.0.0
**Статус**: Анализ завершен, требуется синхронизация

---

## Executive Summary

CloudBridge Client - это полнофункциональный cross-platform клиент с богатой функциональностью:
- **P2P mesh networking** с ICE/STUN/TURN
- **Multi-protocol support** (QUIC, WebSocket, gRPC)
- **L3-overlay network** (WireGuard integration)
- **OIDC authentication** с JWT tokens
- **Connection handover** для seamless migration

**Текущий статус**: Клиент и сервер работают независимо, требуется **синхронизация API и функциональности**.

---

## Архитектура CloudBridge Client

### Компоненты клиента

```
CloudBridge Client Architecture
├── Transport Layer
│ ├── QUIC (primary)
│ ├── WebSocket (fallback)
│ ├── gRPC (API communication)
│ └── WireGuard (L3-overlay)
├── Security Layer
│ ├── OIDC Authentication
│ ├── JWT Token Management
│ ├── TLS 1.3 Encryption
│ └── OS Keyring Integration
├── Network Layer
│ ├── ICE/STUN/TURN (NAT traversal)
│ ├── MASQUE Protocol (tunneling)
│ ├── Connection Handover
│ └── P2P Mesh Networking
└── Management Layer
    ├── Heartbeat Manager
    ├── Metrics & Monitoring
    ├── Performance Optimizer
    └── SLO Controller
```

### Ключевые пакеты клиента

| Пакет | Описание | Зависимость от сервера |
|-------|----------|------------------------|
| `pkg/api/` | HTTP API client для P2P операций | Высокая |
| `pkg/auth/` | Authentication manager (OIDC/JWT) | ысокая |
| `pkg/relay/` | Relay client для сервера | ритическая |
| `pkg/p2p/` | P2P mesh manager | редняя |
| `pkg/quic/` | QUIC connection handler | редняя |
| `pkg/ice/` | ICE agent для NAT traversal | Низкая |
| `pkg/masque/` | MASQUE client | Низкая |
| `pkg/heartbeat/` | Heartbeat manager | ысокая |
| `pkg/metrics/` | Metrics system | Низкая |
| `pkg/handover/` | Connection handover | Средняя |

---

## Интеграционные точки Client-Server

### 1. API Endpoints (КРИТИЧНО)

#### Client ожидает:
```yaml
# pkg/api/client.go
- POST /api/v1/tenants/{tenant_id}/peers/register
- GET /api/v1/tenants/{tenant_id}/peers/discover
- POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat
- POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/status
- POST /api/v1/ice/credentials
```

#### Server предоставляет:
```yaml
# internal/api/ (relay-installer)
- POST /api/v1/peer/register
- GET /api/v1/peer/{peer_id}/peers
- POST /api/v1/peer/{peer_id}/heartbeat
- GET /health
- GET /metrics
```

** ПРОБЛЕМА**: Несовпадение путей API
- Client: `/api/v1/tenants/{tenant_id}/peers/...`
- Server: `/api/v1/peer/...`

** РЕШЕНИЕ**: Добавить alias endpoints на сервере для совместимости

---

### 2. Authentication Flow (КРИТИЧНО)

#### Client реализация:
```go
// pkg/auth/auth.go
type AuthManager struct {
    Type string // "jwt" или "oidc"
    OIDCConfig *OIDCConfig
    JWKSClient *jwks.Client
    TokenValidator func(token string) (*jwt.Token, error)
}
```

**Поддерживает**:
- WT token validation
- IDC discovery
- WKS key retrieval
- oken rotation

#### Server реализация:
```go
// internal/auth/ (relay-installer)
- Zitadel OIDC integration
- JWT validation middleware
- Private-key-JWT for service users
```

** СОВМЕСТИМОСТЬ**: Client и Server используют одинаковую аутентификацию (OIDC/JWT)

**ТРЕБУЕТ ПРОВЕРКИ**:
- Формат JWT claims
- OIDC issuer URL
- JWKS endpoint consistency

---

### 3. P2P Mesh Networking (ВАЖНО)

#### Client P2P Manager:
```go
// pkg/p2p/manager.go
type Manager struct {
    peers map[string]*Peer
    localPeerID string
    tenantID string
    meshEnabled bool
    discoveryTimer *time.Timer
}
```

**Функциональность**:
- eer registration
- eer discovery
- esh configuration
- onnection management

#### Server P2P Support:
```go
// internal/p2p/ (relay-installer)
- Peer registry
- Mesh topology
- Connection tracking
```

** СОВМЕСТИМОСТЬ**: Базовая функциональность совпадает

**ТРЕБУЕТ СИНХРОНИЗАЦИИ**:
- Peer discovery protocol
- Mesh configuration format
- Connection state management

---

### 4. Transport Protocols (ВАЖНО)

#### Client Transport:
```go
// pkg/relay/transport/
- gRPC client (proto definitions)
- QUIC connection
- WebSocket fallback
```

**Proto definitions**:
- `control.proto` - Control plane messages
- `heartbeat.proto` - Heartbeat messages
- `tunnel.proto` - Tunnel management

#### Server Transport:
```go
// internal/quic/ (relay-installer)
- QUIC server with experimental features
- Enhanced connection with BBRv2
- Prometheus metrics
```

** СОВМЕСТИМОСТЬ**: Оба используют QUIC/gRPC

**ТРЕБУЕТ СИНХРОНИЗАЦИИ**:
- Proto definitions (версии)
- QUIC configuration parameters
- Connection handshake protocol

---

### 5. WireGuard L3-Overlay (СРЕДНИЙ ПРИОРИТЕТ)

#### Client WireGuard:
```go
// pkg/relay/wireguard.go
type WireGuardManager struct {
    interfaceName string
    privateKey string
    listenPort int
    peers map[string]*WireGuardPeer
}
```

**Функциональность**:
- ireGuard interface management
- er-peer configuration
- ot-reload support
- ross-platform (Linux/Windows/macOS)

#### Server IPAM:
```go
// internal/ipam/ (relay-installer)
- Deterministic IP allocation
- PostgreSQL storage
- Tenant isolation
```

** СОВМЕСТИМОСТЬ**: IPAM на сервере, WireGuard на клиенте

**ТРЕБУЕТ СИНХРОНИЗАЦИИ**:
- IP allocation API
- WireGuard config format
- Peer update notifications

---

### 6. Heartbeat & Monitoring (ВАЖНО)

#### Client Heartbeat:
```go
// pkg/heartbeat/manager.go
type Manager struct {
    interval time.Duration
    lastHeartbeat time.Time
    failureCount int
    maxFailures int
}
```

**Отправляет**:
```json
{
  "status": "active",
  "relay_session_id": "...",
  "metrics": {
    "rtt": 50,
    "packet_loss": 0.01,
    "bandwidth_up": 100000,
    "bandwidth_down": 100000
  }
}
```

#### Server Heartbeat:
```go
// internal/heartbeat/ (relay-installer)
- Heartbeat handler
- Connection tracking
- Timeout management
```

** СОВМЕСТИМОСТЬ**: Протокол heartbeat совпадает

**ТРЕБУЕТ ПРОВЕРКИ**:
- Heartbeat interval (client: configurable, server: ?)
- Timeout values
- Metrics format

---

### 7. ICE/NAT Traversal (ВАЖНО)

#### Client ICE Agent:
```go
// pkg/ice/agent.go
- ICE candidate gathering
- STUN/TURN support
- Connectivity checks
```

**Использует**:
- STUN servers (configurable)
- TURN servers (configurable)
- ICE credentials exchange

#### Server NAT Traversal:
```go
// internal/nat/ (relay-installer)
- 24 Prometheus metrics
- Enhanced connectivity checker
- Pion/ICE adapter (documentation)
```

** ЧАСТИЧНАЯ СОВМЕСТИМОСТЬ**: Оба используют ICE

** ПРОБЛЕМА**: Server TURN сервер не настроен (coturn)

**ТРЕБУЕТ ДОРАБОТКИ**:
- TURN server deployment (coturn)
- ICE credentials API
- Fallback strategies

---

### 8. MASQUE Protocol (СРЕДНИЙ ПРИОРИТЕТ)

#### Client MASQUE:
```go
// pkg/masque/client.go
- MASQUE client implementation
- HTTP/3-based tunneling
```

#### Server MASQUE:
```go
// internal/masque/ (relay-installer)
- MASQUE proxy (RFC 9298)
- CONNECT-UDP support
- Metrics
```

** СОВМЕСТИМОСТЬ**: Оба реализуют MASQUE

**ТРЕБУЕТ ПРОВЕРКИ**:
- Protocol version compatibility
- Configuration parameters

---

### 9. Metrics & Monitoring (СРЕДНИЙ ПРИОРИТЕТ)

#### Client Metrics:
```go
// pkg/metrics/metrics.go
- Prometheus metrics
- Pushgateway support
- Connection metrics
- Performance metrics
```

**Экспортирует**:
- `cloudbridge_client_connections_total`
- `cloudbridge_client_heartbeat_success`
- `cloudbridge_client_rtt_seconds`
- `cloudbridge_client_bandwidth_bytes`

#### Server Metrics:
```go
// internal/quic/experimental_metrics.go
- 45+ Prometheus metrics
- Grafana dashboard
- HTTP /metrics endpoint
```

** СОВМЕСТИМОСТЬ**: Оба используют Prometheus

**ТРЕБУЕТ СИНХРОНИЗАЦИИ**:
- Metric naming convention
- Label consistency
- Dashboard integration

---

## Детальный анализ интеграционных проблем

### Проблема 1: API Path Mismatch (КРИТИЧНО)

**Что не совпадает**:
```
Client Server
----------------------------------------------
/api/v1/tenants/{tid}/peers/register vs /api/v1/peer/register
/api/v1/tenants/{tid}/peers/discover vs /api/v1/peer/{pid}/peers
/api/v1/tenants/{tid}/peers/{pid}/... vs /api/v1/peer/{pid}/...
```

**Impact**: **БЛОКИРУЮЩАЯ** - Client не может подключиться к серверу

**Решение 1 (Рекомендуемое)**: Добавить alias routes на сервере
```go
// internal/api/routes.go
// Legacy routes (client compatibility)
r.POST("/api/v1/tenants/:tenant_id/peers/register", aliasToNewEndpoint)
r.GET("/api/v1/tenants/:tenant_id/peers/discover", aliasToNewEndpoint)

// New routes (server native)
r.POST("/api/v1/peer/register", peerRegisterHandler)
r.GET("/api/v1/peer/:peer_id/peers", peerDiscoverHandler)
```

**Решение 2 (Альтернативное)**: Обновить client API paths
```go
// pkg/api/client.go
// Update baseURL and paths to match server
const (
    PeerRegisterPath = "/api/v1/peer/register"
    PeerDiscoverPath = "/api/v1/peer/%s/peers"
)
```

**Рекомендация**: Использовать Решение 1 (обратная совместимость)

---

### Проблема 2: Proto Definitions Sync (ВАЖНО)

**Client proto files**:
```
pkg/relay/transport/proto/
├── control.proto
├── heartbeat.proto
└── tunnel.proto
```

**Server proto files**:
```
api/proto/
├── relay.proto
├── p2p.proto
└── control.proto (?)
```

**Impact**: **ВАЖНО** - Возможны проблемы совместимости gRPC

**Решение**: Централизовать proto definitions
```bash
# Создать общий репозиторий proto
mkdir -p api/proto/v1/
cp client/pkg/relay/transport/proto/*.proto api/proto/v1/

# Генерировать код для обоих
make generate-proto-client
make generate-proto-server
```

---

### Проблема 3: TURN Server Missing (КРИТИЧНО для NAT)

**Client configuration**:
```yaml
ice:
  stun_servers:
    - "stun.l.google.com:19302"
  turn_servers: [] # EMPTY!
```

**Server status**:
- TURN server (coturn) не настроен
- TUN support есть
- Fallback strategies не реализованы

**Impact**: **КРИТИЧНО** - P2P connections будут fails в симметричных NAT

**Решение**: Развернуть coturn TURN server
```bash
# День 1-2: Install coturn
apt-get install coturn

# Configure
cat > /etc/turnserver.conf <<EOF
listening-port=3478
tls-listening-port=5349
realm=relay.2gc.ru
server-name=relay.2gc.ru
lt-cred-mech
user=cloudbridge:PASSWORD_HERE
EOF

# Start
systemctl start coturn

# Update client config
ice:
  turn_servers:
    - "turn:relay.2gc.ru:3478?transport=udp"
    - "turns:relay.2gc.ru:5349?transport=tcp"
```

---

### Проблема 4: JWT Claims Format (ВАЖНО)

**Client expects**:
```go
type Claims struct {
    TenantID string `json:"tenant_id"`
    PeerID string `json:"peer_id"`
    Scope string `json:"scope"`
    jwt.RegisteredClaims
}
```

**Server generates** (Zitadel):
```go
// Zitadel OIDC claims
{
  "sub": "user-id",
  "aud": ["cloudbridge-relay"],
  "iss": "https://zitadel.2gc.ru",
  "tenant_id": "...", // Custom claim?
  "scope": "..."
}
```

**Impact**: **ВАЖНО** - Token validation может fails

**Решение**: Проверить и синхронизировать claims
```go
// Добавить middleware для claims mapping
func MapZitadelClaims(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        claims := extractClaims(r)
        // Map Zitadel claims to client expected format
        ctx := context.WithValue(r.Context(), "tenant_id", claims["tenant_id"])
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## Матрица совместимости

| Компонент | Client | Server | Статус | Приоритет | Усилия |
|-----------|--------|--------|--------|-----------|--------|
| API Paths | | Не совпадают | Критично | 2-3 дня |
| Authentication (OIDC) | | овместимо | | 0 ней |
| JWT Claims | | Требует проверки | Важно | 1 день |
| QUIC Transport | | овместимо | | 0 ней |
| gRPC Proto | | Требует sync | Важно | 2-3 дня |
| P2P Discovery | | астично | Важно | 3-4 дня |
| ICE/STUN | | овместимо | | 0 ней |
| TURN Server | | Не настроен | Критично | 2-3 дня |
| WireGuard IPAM | | ребует sync | Средний | 3-4 дня |
| Heartbeat | | овместимо | | 0 ней |
| MASQUE Protocol | | ребует проверки | Средний | 1-2 дня |
| Metrics | | астично | Средний | 2-3 дня |

**Легенда**:
- Критично - блокирует работу
- Важно - ограничивает функциональность
- Средний - желательно для полной функциональности
- отово - не требует действий

---

## Приоритизация работ

### Фаза 1: Критические блокеры (1-2 недели)
**Цель**: Обеспечить базовую connectivity клиента к серверу

1. **API Path Compatibility** (2-3 дня)
   - Добавить alias routes на сервере
   - Протестировать peer registration
   - Протестировать peer discovery

2. **TURN Server Deployment** (2-3 дня)
   - Установить coturn
   - Настроить credentials
   - Обновить client config

3. **JWT Claims Validation** (1 день)
   - Проверить Zitadel claims format
   - Добавить claims mapping если нужно
   - Протестировать token validation

**Результат**: Client может подключиться к серверу и зарегистрироваться

---

### Фаза 2: Важная функциональность (2-3 недели)
**Цель**: Обеспечить полную P2P connectivity

4. **gRPC Proto Sync** (2-3 дня)
   - Централизовать proto definitions
   - Обновить генерацию кода
   - Протестировать gRPC calls

5. **P2P Discovery Protocol** (3-4 дня)
   - Синхронизировать peer discovery format
   - Обновить mesh configuration
   - Протестировать P2P connections

6. **WireGuard IPAM Integration** (3-4 дня)
   - Синхронизировать IP allocation API
   - Обновить WireGuard config format
   - Протестировать L3-overlay

**Результат**: P2P mesh networking полностью работает

---

### Фаза 3: Расширенная функциональность (1-2 недели)
**Цель**: Оптимизация и monitoring

7. **Metrics Integration** (2-3 дня)
   - Синхронизировать metric names
   - Обновить Grafana dashboard
   - Добавить client metrics

8. **MASQUE Protocol Validation** (1-2 дня)
   - Проверить protocol compatibility
   - Протестировать tunneling
   - Документировать использование

**Результат**: Полная observability и advanced features

---

## Следующие шаги

### Немедленные действия (эта неделя):

1. **Создать compatibility layer на сервере**
   ```bash
   mkdir -p internal/api/compatibility/
   touch internal/api/compatibility/client_routes.go
   ```

2. **Развернуть TURN сервер**
   ```bash
   ./scripts/setup/setup-turn-server.sh
   ```

3. **Проверить JWT claims**
   ```bash
   # Получить токен от Zitadel
   # Декодировать и проверить claims
   # Сравнить с client ожиданиями
   ```

4. **Создать integration test**
   ```bash
   mkdir -p tests/integration/client-server/
   # Тесты для peer registration, discovery, heartbeat
   ```

---

## Метрики успеха

### MVP (Minimum Viable Product):
- [ ] Client может зарегистрироваться на сервере
- [ ] Client может discover других peers
- [ ] Client может отправлять heartbeats
- [ ] P2P connection establishes (через relay)

### Full Integration:
- [ ] P2P direct connections работают (NAT traversal)
- [ ] WireGuard L3-overlay работает
- [ ] Metrics собираются на сервере
- [ ] MASQUE tunneling работает
- [ ] Connection handover работает

### Production Ready:
- [ ] 95% P2P connection success rate
- [ ] < 2 секунды connection establishment
- [ ] < 50ms p95 latency
- [ ] Полная observability (logs, metrics, traces)
- [ ] Automated testing (unit, integration, e2e)

---

**Авторы**: CloudBridge Team
**Дата создания**: 2025-10-13
**Последнее обновление**: 2025-10-13
**Версия**: 1.0.0
