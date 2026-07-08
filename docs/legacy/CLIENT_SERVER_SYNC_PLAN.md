# CloudBridge Client-Server Synchronization Plan

**Дата**: 2025-10-13
**Версия**: 1.0.0
**Статус**: Готов к реализации
**Время выполнения**: 4-6 недель

---

## Executive Summary

Детальный план синхронизации CloudBridge Client и CloudBridge Relay Server для обеспечения полной совместимости и функциональности P2P mesh networking.

**Текущая готовность**: 30%
**Целевая готовность**: 100%
**Приоритет**: Высокий

---

## Фаза 1: Критические блокеры (Недели 1-2)

### 1.1 API Path Compatibility Layer (День 1-3)

**Цель**: Обеспечить совместимость API endpoints между client и server

#### Задачи:

**День 1: Анализ и дизайн**
- [ ] Создать маппинг всех client API paths → server paths
- [ ] Определить strategy (alias routes vs middleware)
- [ ] Создать документацию API compatibility

**День 2-3: Реализация**
- [ ] Создать `internal/api/compatibility/client_routes.go`
- [ ] Добавить alias routes для client compatibility
- [ ] Создать middleware для path translation
- [ ] Добавить unit tests

**Файлы для создания**:
```bash
internal/api/compatibility/
├── client_routes.go          # Alias routes
├── client_routes_test.go     # Tests
├── middleware.go             # Path translation middleware
└── README.md                 # Documentation
```

**Код (пример)**:
```go
// internal/api/compatibility/client_routes.go
package compatibility

import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/api"
)

// RegisterClientCompatibilityRoutes adds legacy client routes
func RegisterClientCompatibilityRoutes(r *gin.Engine, handlers *api.Handlers) {
    // Client expects: /api/v1/tenants/{tenant_id}/peers/register
    // Server has: /api/v1/peer/register
    r.POST("/api/v1/tenants/:tenant_id/peers/register", func(c *gin.Context) {
        // Extract tenant_id from path, add to context
        tenantID := c.Param("tenant_id")
        c.Set("tenant_id", tenantID)

        // Forward to native handler
        handlers.PeerRegisterHandler(c)
    })

    // Client expects: /api/v1/tenants/{tenant_id}/peers/discover
    // Server has: /api/v1/peer/{peer_id}/peers
    r.GET("/api/v1/tenants/:tenant_id/peers/discover", func(c *gin.Context) {
        tenantID := c.Param("tenant_id")
        c.Set("tenant_id", tenantID)

        // Get peer_id from JWT or request
        peerID := getPeerIDFromContext(c)
        c.Param("peer_id", peerID)

        handlers.PeerDiscoverHandler(c)
    })

    // More routes...
}
```

**Тестирование**:
```bash
# Unit tests
go test ./internal/api/compatibility/... -v

# Integration tests
curl -X POST http://localhost:8080/api/v1/tenants/test-tenant/peers/register \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"peer_id":"test-peer","public_key":"..."}'
```

---

### 1.2 TURN Server Deployment (День 4-6)

**Цель**: Развернуть coturn TURN server для NAT traversal

#### Задачи:

**День 4: Installation**
- [ ] Установить coturn package
- [ ] Создать configuration file
- [ ] Настроить systemd service
- [ ] Открыть firewall ports

**День 5: Configuration**
- [ ] Настроить TLS certificates
- [ ] Создать TURN credentials
- [ ] Настроить relay IP addresses
- [ ] Тестировать connectivity

**День 6: Integration**
- [ ] Обновить server configuration
- [ ] Обновить client default config
- [ ] Создать документацию
- [ ] End-to-end testing

**Setup Script**:
```bash
#!/bin/bash
# scripts/setup/setup-turn-server.sh

set -e

info() { echo "[INFO] $1"; }
success() { echo "[SUCCESS] $1"; }

info "Installing coturn..."
apt-get update && apt-get install -y coturn

info "Configuring coturn..."
cat > /etc/turnserver.conf <<EOF
# CloudBridge TURN Server Configuration

# Listening ports
listening-port=3478
tls-listening-port=5349

# Relay configuration
relay-ip=$(hostname -I | awk '{print $1}')
external-ip=$(curl -s https://api.ipify.org)

# Realm and server name
realm=relay.2gc.ru
server-name=relay.2gc.ru

# Authentication
lt-cred-mech
user=cloudbridge:$(openssl rand -base64 32)

# Security
fingerprint
no-multicast-peers
no-loopback-peers

# TLS certificates
cert=/etc/letsencrypt/live/relay.2gc.ru/fullchain.pem
pkey=/etc/letsencrypt/live/relay.2gc.ru/privkey.pem

# Logging
log-file=/var/log/turnserver.log
verbose

# Performance
max-bps=1000000
total-quota=100
EOF

info "Starting coturn..."
systemctl enable coturn
systemctl start coturn

success "TURN server deployed!"
info "TURN URI: turn:relay.2gc.ru:3478"
info "TURNS URI: turns:relay.2gc.ru:5349"
```

**Тестирование**:
```bash
# Test STUN
turnutils_stunclient relay.2gc.ru

# Test TURN
turnutils_uclient -u cloudbridge -w PASSWORD relay.2gc.ru
```

---

### 1.3 JWT Claims Validation (День 7)

**Цель**: Проверить и синхронизировать JWT claims format

#### Задачи:

- [ ] Получить test JWT token от Zitadel
- [ ] Декодировать и проанализировать claims
- [ ] Сравнить с client expected format
- [ ] Создать claims mapping middleware если нужно
- [ ] Добавить validation tests

**Claims Mapper**:
```go
// internal/auth/claims_mapper.go
package auth

type ZitadelClaims struct {
    Sub       string `json:"sub"`
    Aud       []string `json:"aud"`
    Iss       string `json:"iss"`
    TenantID  string `json:"tenant_id,omitempty"`
    PeerID    string `json:"peer_id,omitempty"`
    Scope     string `json:"scope"`
    jwt.RegisteredClaims
}

type ClientClaims struct {
    TenantID string `json:"tenant_id"`
    PeerID   string `json:"peer_id"`
    Scope    string `json:"scope"`
    jwt.RegisteredClaims
}

// MapZitadelToClientClaims converts Zitadel claims to client format
func MapZitadelToClientClaims(zc *ZitadelClaims) *ClientClaims {
    return &ClientClaims{
        TenantID: zc.TenantID,
        PeerID:   zc.PeerID,
        Scope:    zc.Scope,
        RegisteredClaims: zc.RegisteredClaims,
    }
}
```

---

## Фаза 2: Важная функциональность (Недели 3-4)

### 2.1 gRPC Proto Synchronization (День 8-10)

**Цель**: Централизовать и синхронизировать proto definitions

#### Задачи:

**День 8: Analysis & Design**
- [ ] Сравнить client proto files vs server proto files
- [ ] Определить общую структуру
- [ ] Создать migration plan

**День 9: Implementation**
- [ ] Создать `api/proto/v1/` directory
- [ ] Скопировать и merge proto files
- [ ] Обновить imports в client и server
- [ ] Regenerate proto code

**День 10: Testing**
- [ ] Unit tests для gRPC services
- [ ] Integration tests client → server
- [ ] Backward compatibility tests

**Proto Structure**:
```
api/proto/v1/
├── control.proto         # Control plane messages
├── heartbeat.proto       # Heartbeat messages
├── tunnel.proto          # Tunnel management
├── p2p.proto             # P2P mesh messages
└── common.proto          # Common types
```

**Build Script**:
```bash
#!/bin/bash
# scripts/proto/generate.sh

# Generate for server
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/v1/*.proto

# Generate for client
cd oss-repositories/cloudbridge/client
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ../../../api/proto/v1/*.proto
```

---

### 2.2 P2P Discovery Protocol Sync (День 11-14)

**Цель**: Синхронизировать P2P peer discovery между client и server

#### Задачи:

**День 11-12: Discovery API**
- [ ] Синхронизировать peer registration format
- [ ] Синхронизировать discovery response format
- [ ] Обновить server handlers
- [ ] Обновить client API calls

**День 13: Mesh Configuration**
- [ ] Синхронизировать mesh config format
- [ ] Обновить mesh topology management
- [ ] Тестировать multi-peer scenarios

**День 14: Testing**
- [ ] End-to-end P2P discovery tests
- [ ] Multi-tenant isolation tests
- [ ] Performance testing (discovery latency)

**Discovery Protocol**:
```go
// Unified discovery format
type PeerDiscoveryResponse struct {
    Success  bool   `json:"success"`
    TenantID string `json:"tenant_id"`
    Peers    []PeerInfo `json:"peers"`
}

type PeerInfo struct {
    PeerID         string   `json:"peer_id"`
    PublicKey      string   `json:"public_key"`
    AllowedIPs     []string `json:"allowed_ips"`
    Endpoint       string   `json:"endpoint,omitempty"`
    IsOnline       bool     `json:"is_online"`
    LastSeen       string   `json:"last_seen"`
    QUICConfig     *QUICConfig `json:"quic_config,omitempty"`
}
```

---

### 2.3 WireGuard IPAM Integration (День 15-18)

**Цель**: Синхронизировать IPAM и WireGuard configuration

#### Задачи:

**День 15-16: IPAM API**
- [ ] Создать API для IP allocation
- [ ] Интегрировать с существующим IPAM
- [ ] Обновить client WireGuard manager
- [ ] Тестировать IP allocation

**День 17: Configuration Sync**
- [ ] Синхронизировать WireGuard config format
- [ ] Реализовать hot-reload на client
- [ ] Обновить peer update notifications
- [ ] Тестировать config updates

**День 18: Testing**
- [ ] End-to-end L3-overlay tests
- [ ] Multi-peer WireGuard mesh tests
- [ ] Tenant isolation tests

**IPAM API**:
```go
// POST /api/v1/ipam/allocate
type IPAllocationRequest struct {
    TenantID string `json:"tenant_id"`
    PeerID   string `json:"peer_id"`
    Subnet   string `json:"subnet,omitempty"` // Optional preference
}

type IPAllocationResponse struct {
    Success    bool   `json:"success"`
    IPAddress  string `json:"ip_address"`
    Subnet     string `json:"subnet"`
    Gateway    string `json:"gateway"`
    DNS        []string `json:"dns,omitempty"`
}

// GET /api/v1/wireguard/config
type WireGuardConfigResponse struct {
    Success    bool   `json:"success"`
    Interface  WireGuardInterface `json:"interface"`
    Peers      []WireGuardPeer `json:"peers"`
}
```

---

## Фаза 3: Расширенная функциональность (Недели 5-6)

### 3.1 Metrics Integration (День 19-21)

**Цель**: Синхронизировать metrics между client и server

#### Задачи:

**День 19: Metric Names Sync**
- [ ] Создать unified metric naming convention
- [ ] Обновить client metrics
- [ ] Обновить server metrics
- [ ] Документировать metrics

**День 20: Dashboard Integration**
- [ ] Обновить Grafana dashboard
- [ ] Добавить client metrics panels
- [ ] Создать alerts для client metrics
- [ ] Тестировать visualization

**День 21: Testing**
- [ ] Metrics collection tests
- [ ] Dashboard rendering tests
- [ ] Alert triggering tests

**Metric Convention**:
```
cloudbridge_{component}_{metric}_{unit}

Examples:
- cloudbridge_client_connections_total
- cloudbridge_client_rtt_seconds
- cloudbridge_server_peers_registered
- cloudbridge_p2p_connections_active
```

---

### 3.2 MASQUE Protocol Validation (День 22-23)

**Цель**: Проверить MASQUE protocol compatibility

#### Задачи:

**День 22: Protocol Testing**
- [ ] Тестировать MASQUE client → server
- [ ] Проверить CONNECT-UDP tunneling
- [ ] Валидировать performance
- [ ] Документировать использование

**День 23: Integration**
- [ ] Интегрировать MASQUE с P2P mesh
- [ ] Создать fallback logic
- [ ] End-to-end tests
- [ ] Performance benchmarks

---

### 3.3 Final Integration Testing (День 24-28)

**Цель**: Комплексное тестирование всей системы

#### Задачи:

**День 24-25: End-to-End Tests**
- [ ] Full P2P mesh formation (3+ peers)
- [ ] Multi-tenant isolation
- [ ] NAT traversal scenarios
- [ ] Connection handover

**День 26: Performance Testing**
- [ ] Connection establishment latency
- [ ] P2P throughput
- [ ] Resource usage (CPU/Memory)
- [ ] Scalability tests (10+ peers)

**День 27: Failure Scenarios**
- [ ] Network partition recovery
- [ ] Server restart resilience
- [ ] Client reconnection
- [ ] Token expiration handling

**День 28: Documentation**
- [ ] Update integration-plan.md
- [ ] Create deployment guide
- [ ] Update API documentation
- [ ] Create troubleshooting guide

---

## Критерии успеха

### Phase 1 (Week 1-2): MVP
- [x] API paths compatibility
- [x] TURN server deployed
- [x] JWT claims validated
- [ ] Client can register on server
- [ ] Client can discover peers
- [ ] Heartbeat working

### Phase 2 (Week 3-4): Full P2P
- [ ] gRPC proto synchronized
- [ ] P2P discovery working
- [ ] WireGuard L3-overlay working
- [ ] Direct P2P connections
- [ ] NAT traversal success rate > 80%

### Phase 3 (Week 5-6): Production Ready
- [ ] Metrics integrated
- [ ] MASQUE validated
- [ ] All tests passing
- [ ] Documentation complete
- [ ] Performance targets met

### Performance Targets:
- **Connection establishment**: < 2 seconds
- **P2P discovery**: < 1 second
- **RTT (relay)**: < 100ms
- **RTT (direct)**: < 50ms
- **Throughput**: > 100 Mbps
- **NAT traversal success**: > 90%

---

## Timeline & Milestones

```
Week 1-2: Critical Blockers
├── Day 1-3:   API Compatibility     ████████░░ 80%
├── Day 4-6:   TURN Deployment       ░░░░░░░░░░  0%
└── Day 7:     JWT Claims            ░░░░░░░░░░  0%

Week 3-4: Important Features
├── Day 8-10:  gRPC Proto Sync       ░░░░░░░░░░  0%
├── Day 11-14: P2P Discovery         ░░░░░░░░░░  0%
└── Day 15-18: WireGuard IPAM        ░░░░░░░░░░  0%

Week 5-6: Advanced Features
├── Day 19-21: Metrics Integration   ░░░░░░░░░░  0%
├── Day 22-23: MASQUE Validation     ░░░░░░░░░░  0%
└── Day 24-28: Final Testing         ░░░░░░░░░░  0%
```

**Current Progress**: 10%
**Target Progress**: 100%
**Estimated Completion**: 2025-11-24 (6 weeks)

---

## Next Steps

### This Week (Week 1):
1. Create compatibility layer (`internal/api/compatibility/`)
2. Deploy TURN server (`scripts/setup/setup-turn-server.sh`)
3. Validate JWT claims format
4. Start integration testing

### Next Week (Week 2):
1. Complete API compatibility layer
2. Configure and test TURN server
3. Begin gRPC proto synchronization
4. Document progress

---

**Авторы**: CloudBridge Team
**Дата создания**: 2025-10-13
**Последнее обновление**: 2025-10-13
**Версия**: 1.0.0
