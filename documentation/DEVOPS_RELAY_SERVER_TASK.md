# 🚀 Техническое задание для DevOps Relay сервера

## 📋 Обзор задачи

Необходимо обновить Relay сервер для поддержки P2P функциональности. Сервер должен предоставлять API для peer discovery, управлять P2P соединениями и обеспечивать fallback маршрутизацию для mesh сети.

## 🎯 Цели

1. **Peer Discovery API** - Обнаружение и регистрация серверных клиентов
2. **P2P координация** - Управление P2P соединениями между серверами
3. **Fallback маршрутизация** - Резервные маршруты через relay
4. **Мониторинг P2P** - Отслеживание состояния mesh сети
5. **Безопасность** - Валидация P2P токенов и peer whitelist

## 🔧 Технические требования

### 1. **Peer Discovery API**

#### **A. Регистрация пира**
```http
POST /api/v1/peers/register
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "peer_id": "server-client-789",
  "tenant_id": "test-tenant-456",
  "org_id": "test-org-123",
  "device_id": "device-abc123",
  "connection_type": "p2p-mesh",
  "wireguard_config": {
    "public_key": "F923V0yD5EC58W7OLNc3KOH6gkhrCWAGqU7s7srkZgQ=",
    "endpoint": "server1.example.com:51820",
    "listen_port": 51820,
    "allowed_ips": ["10.0.1.1/32"]
  },
  "mesh_config": {
    "auto_discovery": true,
    "routing": "hybrid"
  },
  "peer_whitelist": {
    "allowed_peers": ["peer1", "peer2"],
    "max_peers": 10
  },
  "network_config": {
    "subnet": "10.0.1.0/24"
  }
}

Response:
{
  "success": true,
  "peer_id": "server-client-789",
  "registered_at": "2024-01-01T00:00:00Z",
  "discovery_endpoint": "wss://relay.2gc.ru/ws/peers"
}
```

#### **B. Discovery пиров**
```http
GET /api/v1/peers/discover?tenant_id=test-tenant-456
Authorization: Bearer <jwt_token>

Response:
{
  "success": true,
  "tenant_id": "test-tenant-456",
  "peers": [
    {
      "peer_id": "server-client-001",
      "device_id": "device-abc123",
      "wireguard_config": {
        "public_key": "F923V0yD5EC58W7OLNc3KOH6gkhrCWAGqU7s7srkZgQ=",
        "endpoint": "server1.example.com:51820",
        "allowed_ips": ["10.0.1.1/32"]
      },
      "mesh_config": {
        "routing": "hybrid"
      },
      "is_online": true,
      "last_seen": "2024-01-01T12:00:00Z",
      "latency_ms": 15
    },
    {
      "peer_id": "server-client-002",
      "device_id": "device-def456",
      "wireguard_config": {
        "public_key": "G834W1zE6FD69X8PMNc4LPH7hlisDXBHrV8t8tstlZgR=",
        "endpoint": "server2.example.com:51820",
        "allowed_ips": ["10.0.1.2/32"]
      },
      "mesh_config": {
        "routing": "hybrid"
      },
      "is_online": true,
      "last_seen": "2024-01-01T12:00:00Z",
      "latency_ms": 22
    }
  ]
}
```

#### **C. Обновление статуса пира**
```http
PUT /api/v1/peers/{peer_id}/status
Authorization: Bearer <jwt_token>

{
  "is_online": true,
  "latency_ms": 15,
  "active_connections": 3,
  "mesh_status": "connected"
}

Response:
{
  "success": true,
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### **D. Получение информации о пире**
```http
GET /api/v1/peers/{peer_id}
Authorization: Bearer <jwt_token>

Response:
{
  "success": true,
  "peer": {
    "peer_id": "server-client-789",
    "tenant_id": "test-tenant-456",
    "device_id": "device-abc123",
    "connection_type": "p2p-mesh",
    "wireguard_config": { ... },
    "mesh_config": { ... },
    "peer_whitelist": { ... },
    "network_config": { ... },
    "is_online": true,
    "last_seen": "2024-01-01T12:00:00Z",
    "registered_at": "2024-01-01T00:00:00Z",
    "active_connections": 3,
    "mesh_status": "connected"
  }
}
```

### 2. **WebSocket API для real-time updates**

#### **A. Подключение к WebSocket**
```javascript
const ws = new WebSocket('wss://relay.2gc.ru/ws/peers', {
  headers: {
    'Authorization': 'Bearer <jwt_token>'
  }
});

// Сообщения от сервера:
{
  "type": "peer_joined",
  "data": {
    "peer_id": "server-client-003",
    "wireguard_config": { ... },
    "timestamp": "2024-01-01T12:00:00Z"
  }
}

{
  "type": "peer_left",
  "data": {
    "peer_id": "server-client-002",
    "timestamp": "2024-01-01T12:00:00Z"
  }
}

{
  "type": "peer_status_update",
  "data": {
    "peer_id": "server-client-001",
    "is_online": true,
    "latency_ms": 15,
    "timestamp": "2024-01-01T12:00:00Z"
  }
}
```

### 3. **Fallback маршрутизация**

#### **A. Маршрутизация через relay**
```http
POST /api/v1/relay/route
Authorization: Bearer <jwt_token>

{
  "from_peer": "server-client-001",
  "to_peer": "server-client-002",
  "data": "base64_encoded_data",
  "message_type": "mesh_packet"
}

Response:
{
  "success": true,
  "routed_at": "2024-01-01T12:00:00Z",
  "latency_ms": 45
}
```

#### **B. Статистика маршрутизации**
```http
GET /api/v1/relay/stats?tenant_id=test-tenant-456
Authorization: Bearer <jwt_token>

Response:
{
  "success": true,
  "tenant_id": "test-tenant-456",
  "stats": {
    "total_peers": 5,
    "online_peers": 4,
    "direct_connections": 8,
    "relay_connections": 2,
    "total_traffic_bytes": 1024000,
    "average_latency_ms": 18
  }
}
```

### 4. **Валидация P2P токенов**

#### **A. Проверка JWT токена**
```go
func validateP2PToken(token string) (*P2PClaims, error) {
    // Парсинг JWT токена
    claims, err := parseJWTToken(token)
    if err != nil {
        return nil, err
    }
    
    // Проверка обязательных полей
    if claims.ConnectionType != "p2p-mesh" {
        return nil, errors.New("invalid connection type")
    }
    
    if claims.DeviceID == "" {
        return nil, errors.New("missing device_id")
    }
    
    if claims.WireGuardConfig == nil {
        return nil, errors.New("missing wireguard_config")
    }
    
    // Проверка отзыва токена
    if isTokenRevoked(claims.JTI) {
        return nil, errors.New("token revoked")
    }
    
    return claims, nil
}
```

#### **B. Проверка peer whitelist**
```go
func validatePeerAccess(peerID string, whitelist *PeerWhitelist) bool {
    if whitelist.AutoApprove {
        return true
    }
    
    for _, allowedPeer := range whitelist.AllowedPeers {
        if allowedPeer == peerID {
            return true
        }
    }
    
    return false
}
```

### 5. **База данных**

#### **A. Таблица пиров**
```sql
CREATE TABLE peers (
  id TEXT PRIMARY KEY,
  peer_id TEXT UNIQUE NOT NULL,
  tenant_id TEXT NOT NULL,
  org_id TEXT NOT NULL,
  device_id TEXT NOT NULL,
  connection_type TEXT NOT NULL,
  wireguard_config JSON NOT NULL,
  mesh_config JSON,
  peer_whitelist JSON,
  network_config JSON,
  is_online BOOLEAN DEFAULT false,
  last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  active_connections INTEGER DEFAULT 0,
  mesh_status TEXT DEFAULT 'disconnected',
  
  INDEX idx_tenant_id (tenant_id),
  INDEX idx_peer_id (peer_id),
  INDEX idx_device_id (device_id),
  INDEX idx_is_online (is_online)
);
```

#### **B. Таблица соединений**
```sql
CREATE TABLE peer_connections (
  id TEXT PRIMARY KEY,
  from_peer_id TEXT NOT NULL,
  to_peer_id TEXT NOT NULL,
  connection_type TEXT NOT NULL, -- 'direct' or 'relay'
  is_active BOOLEAN DEFAULT true,
  established_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  bytes_transferred BIGINT DEFAULT 0,
  latency_ms INTEGER,
  
  UNIQUE(from_peer_id, to_peer_id),
  INDEX idx_from_peer (from_peer_id),
  INDEX idx_to_peer (to_peer_id),
  INDEX idx_connection_type (connection_type)
);
```

#### **C. Таблица метрик**
```sql
CREATE TABLE peer_metrics (
  id TEXT PRIMARY KEY,
  peer_id TEXT NOT NULL,
  metric_type TEXT NOT NULL, -- 'latency', 'traffic', 'connections'
  metric_value DECIMAL(10,2) NOT NULL,
  recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  
  INDEX idx_peer_id (peer_id),
  INDEX idx_metric_type (metric_type),
  INDEX idx_recorded_at (recorded_at)
);
```

### 6. **Мониторинг и метрики**

#### **A. Prometheus метрики**
```go
var (
    peersTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peers_total",
            Help: "Total number of registered peers",
        },
        []string{"tenant_id", "connection_type"},
    )
    
    peersOnline = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peers_online",
            Help: "Number of online peers",
        },
        []string{"tenant_id"},
    )
    
    peerConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cloudbridge_peer_connections",
            Help: "Number of active peer connections",
        },
        []string{"tenant_id", "connection_type"},
    )
    
    peerLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "cloudbridge_peer_latency_seconds",
            Help: "Peer connection latency",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
        },
        []string{"tenant_id", "peer_id"},
    )
    
    relayTraffic = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cloudbridge_relay_traffic_bytes_total",
            Help: "Total traffic routed through relay",
        },
        []string{"tenant_id", "direction"},
    )
)
```

#### **B. Health check endpoint**
```http
GET /api/v1/health

Response:
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "websocket": "healthy"
  },
  "metrics": {
    "total_peers": 25,
    "online_peers": 23,
    "active_connections": 45,
    "uptime_seconds": 86400
  }
}
```

### 7. **Безопасность**

#### **A. Rate limiting**
```go
func setupRateLimiting() {
    // Rate limiting для API endpoints
    limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 requests per second, burst 200
    
    // Rate limiting для WebSocket connections
    wsLimiter := rate.NewLimiter(rate.Limit(10), 20) // 10 connections per second, burst 20
}
```

#### **B. IP whitelist**
```go
func validateIPWhitelist(clientIP string, tenantID string) bool {
    whitelist := getTenantIPWhitelist(tenantID)
    if len(whitelist) == 0 {
        return true // No whitelist means all IPs allowed
    }
    
    for _, allowedIP := range whitelist {
        if isIPInRange(clientIP, allowedIP) {
            return true
        }
    }
    
    return false
}
```

#### **C. Audit logging**
```go
func logAuditEvent(event AuditEvent) {
    logEntry := AuditLog{
        Timestamp:   time.Now(),
        EventType:   event.Type,
        PeerID:      event.PeerID,
        TenantID:    event.TenantID,
        ClientIP:    event.ClientIP,
        UserAgent:   event.UserAgent,
        Details:     event.Details,
    }
    
    // Сохраняем в базу данных
    saveAuditLog(logEntry)
    
    // Отправляем в SIEM систему
    sendToSIEM(logEntry)
}
```

### 8. **Конфигурация**

#### **A. Environment variables**
```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/cloudbridge_relay
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-jwt-secret-key
JWT_ISSUER=https://auth.2gc.ru/realms/cloudbridge

# Server
SERVER_PORT=8080
WEBSOCKET_PORT=8081
TLS_CERT_PATH=/path/to/cert.pem
TLS_KEY_PATH=/path/to/key.pem

# P2P
P2P_DISCOVERY_ENABLED=true
P2P_FALLBACK_ENABLED=true
P2P_MAX_PEERS_PER_TENANT=50

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
LOG_LEVEL=info
```

#### **B. Конфигурационный файл**
```yaml
server:
  port: 8080
  tls:
    enabled: true
    cert_path: "/path/to/cert.pem"
    key_path: "/path/to/key.pem"
  
websocket:
  port: 8081
  max_connections: 1000
  heartbeat_interval: 30s

database:
  url: "postgresql://user:password@localhost:5432/cloudbridge_relay"
  max_connections: 100
  connection_timeout: 30s

redis:
  url: "redis://localhost:6379"
  max_connections: 50
  connection_timeout: 10s

p2p:
  discovery:
    enabled: true
    cleanup_interval: 300s
    peer_timeout: 600s
  
  fallback:
    enabled: true
    max_connections: 1000
    connection_timeout: 30s
  
  limits:
    max_peers_per_tenant: 50
    max_connections_per_peer: 20

security:
  rate_limiting:
    api_requests_per_second: 100
    api_burst: 200
    websocket_connections_per_second: 10
    websocket_burst: 20
  
  ip_whitelist:
    enabled: false
    default_allow: true

monitoring:
  prometheus:
    enabled: true
    port: 9090
    path: "/metrics"
  
  health_check:
    enabled: true
    path: "/health"
    interval: 30s

logging:
  level: "info"
  format: "json"
  audit:
    enabled: true
    siem_endpoint: "https://siem.example.com/api/events"
```

## 📅 План реализации

### **Этап 1: База данных и модели (2-3 дня)**
1. Создание таблиц peers, peer_connections, peer_metrics
2. Создание индексов и внешних ключей
3. Реализация моделей данных в Go

### **Этап 2: API endpoints (3-4 дня)**
1. Реализация peer registration API
2. Реализация peer discovery API
3. Реализация peer status API
4. Реализация fallback routing API

### **Этап 3: WebSocket сервер (2-3 дня)**
1. Реализация WebSocket сервера
2. Обработка real-time updates
3. Управление WebSocket соединениями

### **Этап 4: Валидация и безопасность (2-3 дня)**
1. Реализация JWT валидации
2. Реализация peer whitelist проверки
3. Реализация rate limiting
4. Реализация audit logging

### **Этап 5: Мониторинг и метрики (1-2 дня)**
1. Интеграция с Prometheus
2. Реализация health check
3. Настройка логирования

### **Этап 6: Тестирование (2-3 дня)**
1. Unit тесты для всех компонентов
2. Integration тесты с серверными клиентами
3. Load testing для WebSocket соединений
4. End-to-end тестирование

## 🎯 Критерии готовности

### **Функциональные требования:**
- ✅ Peer registration и discovery API работают
- ✅ WebSocket real-time updates функционируют
- ✅ Fallback маршрутизация работает
- ✅ JWT валидация и peer whitelist работают
- ✅ Мониторинг и метрики собираются

### **Технические требования:**
- ✅ Все API endpoints работают корректно
- ✅ WebSocket сервер стабилен
- ✅ База данных оптимизирована
- ✅ Безопасность обеспечена
- ✅ Мониторинг настроен

### **Производительность:**
- ✅ Поддержка 1000+ одновременных WebSocket соединений
- ✅ Обработка 100+ API запросов в секунду
- ✅ Latency < 50ms для peer discovery
- ✅ 99.9% uptime

### **Тестирование:**
- ✅ Unit тесты покрывают 90%+ кода
- ✅ Integration тесты с клиентами
- ✅ Load testing пройден
- ✅ Security testing пройден

## 📊 Метрики успеха

### **KPI:**
- **Peer Discovery Time**: < 5 секунд
- **WebSocket Connection Time**: < 2 секунды
- **API Response Time**: < 100ms (95th percentile)
- **Fallback Routing Latency**: < 200ms
- **System Uptime**: > 99.9%

### **Мониторинг:**
- Количество зарегистрированных пиров
- Количество активных соединений
- Трафик через fallback маршрутизацию
- Latency между пирами
- Ошибки API и WebSocket

## 📞 Контакты для вопросов

- **Техническая поддержка**: [контакты]
- **Документация**: [ссылки на docs]
- **Примеры кода**: [ссылки на примеры]

---

**Общее время реализации: 12-18 дней**
**Приоритет: Высокий**
**Статус: Готов к началу работ**
