# Multi-tenancy Guide - CloudBridge Relay v2.0

## Обзор

Multi-tenancy в CloudBridge Relay v2.0 обеспечивает полную изоляцию ресурсов между различными клиентами (тенантами) с настраиваемыми лимитами, IP-фильтрацией и мониторингом активности. Эта функциональность позволяет безопасно обслуживать множество клиентов на одном relay-сервере.

## Архитектура Multi-tenancy

### Основные принципы

```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay v2.0                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Tenant A      │  │   Tenant B      │  │   Tenant C   │ │
│  │                 │  │                 │  │              │ │
│  │ • Tunnels: 5/10 │  │ • Tunnels: 2/20 │  │ • Tunnels: 0/5│ │
│  │ • Connections:  │  │ • Connections:  │  │ • Connections:│ │
│  │   50/100        │  │   150/200       │  │   0/50       │ │
│  │ • Bandwidth:    │  │ • Bandwidth:    │  │ • Bandwidth: │ │
│  │   50/100 Mbps   │  │   300/500 Mbps  │  │   0/100 Mbps │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                    Shared Resources                         │
│  • CPU, Memory, Network                                     │
│  • TLS Certificates                                         │
│  • Monitoring & Metrics                                     │
└─────────────────────────────────────────────────────────────┘
```

### Компоненты системы

1. **Tenant Manager** - управление жизненным циклом тенантов
2. **Resource Allocator** - выделение и контроль ресурсов
3. **IP Filter** - фильтрация по IP-адресам
4. **Activity Tracker** - отслеживание активности тенантов
5. **Cleanup Manager** - автоматическая очистка неактивных ресурсов

## Конфигурация Multi-tenancy

### Основные настройки

```json
{
    "tenant": {
        "enabled": true,
        "default_limits": {
            "max_tunnels": 10,
            "max_connections_per_tunnel": 100,
            "max_bandwidth_mbps": 100,
            "max_data_transfer_gb": 1000,
            "rate_limit_per_second": 1000
        },
        "cleanup_interval": "5m",
        "inactivity_timeout": "1h",
        "enforce_limits": true,
        "auto_cleanup": true
    }
}
```

### Конфигурация тенантов

```yaml
# /etc/relay/tenants.yaml
tenants:
  tenant_001:
    name: "Production Company A"
    enabled: true
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
      max_data_transfer_gb: 5000
      rate_limit_per_second: 2000
    allowed_ips:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    blocked_ips:
      - "192.168.1.100"
      - "10.0.0.50"
    metadata:
      contact: "admin@company-a.com"
      environment: "production"
      billing_tier: "premium"
    features:
      tls_enabled: true
      monitoring_enabled: true
      backup_enabled: true

  tenant_002:
    name: "Development Company B"
    enabled: true
    limits:
      max_tunnels: 5
      max_connections_per_tunnel: 50
      max_bandwidth_mbps: 100
      max_data_transfer_gb: 1000
      rate_limit_per_second: 500
    allowed_ips:
      - "192.168.2.0/24"
    metadata:
      contact: "dev@company-b.com"
      environment: "development"
      billing_tier: "basic"
    features:
      tls_enabled: true
      monitoring_enabled: false
      backup_enabled: false

  tenant_003:
    name: "Testing Company C"
    enabled: false  # Временно отключен
    limits:
      max_tunnels: 2
      max_connections_per_tunnel: 10
      max_bandwidth_mbps: 50
      max_data_transfer_gb: 500
      rate_limit_per_second: 100
    allowed_ips:
      - "192.168.3.0/24"
    metadata:
      contact: "test@company-c.com"
      environment: "testing"
      billing_tier: "trial"
```

## Управление ресурсами

### Лимиты ресурсов

| Ресурс | Описание | Единица измерения |
|--------|----------|-------------------|
| `max_tunnels` | Максимальное количество активных туннелей | количество |
| `max_connections_per_tunnel` | Максимальное количество соединений на туннель | количество |
| `max_bandwidth_mbps` | Максимальная пропускная способность | Мбит/с |
| `max_data_transfer_gb` | Максимальный объем переданных данных | ГБ |
| `rate_limit_per_second` | Лимит запросов в секунду | запросы/с |

### Мониторинг использования

```go
// Получение статистики тенанта
type TenantStats struct {
    TenantID           string    `json:"tenant_id"`
    ActiveTunnels      int       `json:"active_tunnels"`
    TotalConnections   int       `json:"total_connections"`
    BytesSent          int64     `json:"bytes_sent"`
    BytesReceived      int64     `json:"bytes_received"`
    DataTransferGB     float64   `json:"data_transfer_gb"`
    RequestsPerSecond  float64   `json:"requests_per_second"`
    LastActivity       time.Time `json:"last_activity"`
    Errors             int       `json:"errors"`
}

// Проверка лимитов
func checkTenantLimits(tenantID string, stats TenantStats) error {
    tenant := getTenant(tenantID)
    limits := tenant.Limits
    
    if stats.ActiveTunnels >= limits.MaxTunnels {
        return fmt.Errorf("tunnel limit exceeded: %d/%d", 
            stats.ActiveTunnels, limits.MaxTunnels)
    }
    
    if stats.TotalConnections >= limits.MaxConnectionsPerTunnel {
        return fmt.Errorf("connection limit exceeded: %d/%d", 
            stats.TotalConnections, limits.MaxConnectionsPerTunnel)
    }
    
    if stats.DataTransferGB >= float64(limits.MaxDataTransferGB) {
        return fmt.Errorf("data transfer limit exceeded: %.2f/%.2f GB", 
            stats.DataTransferGB, float64(limits.MaxDataTransferGB))
    }
    
    return nil
}
```

## IP-фильтрация

### Настройка IP-фильтров

```yaml
tenants:
  tenant_001:
    allowed_ips:
      - "192.168.1.0/24"      # Разрешить всю подсеть
      - "10.0.0.0/8"          # Разрешить большую подсеть
      - "172.16.1.100"        # Разрешить конкретный IP
    blocked_ips:
      - "192.168.1.100"       # Заблокировать конкретный IP
      - "10.0.0.50"           # Заблокировать другой IP
```

### Проверка IP-адресов

```go
func validateClientIP(tenantID string, clientIP string) error {
    tenant := getTenant(tenantID)
    
    // Проверка заблокированных IP
    for _, blockedIP := range tenant.BlockedIPs {
        if isIPInRange(clientIP, blockedIP) {
            return fmt.Errorf("client IP %s is blocked", clientIP)
        }
    }
    
    // Проверка разрешенных IP
    if len(tenant.AllowedIPs) > 0 {
        allowed := false
        for _, allowedIP := range tenant.AllowedIPs {
            if isIPInRange(clientIP, allowedIP) {
                allowed = true
                break
            }
        }
        if !allowed {
            return fmt.Errorf("client IP %s is not in allowed range", clientIP)
        }
    }
    
    return nil
}

func isIPInRange(ip, cidr string) bool {
    _, ipNet, err := net.ParseCIDR(cidr)
    if err != nil {
        // Если не CIDR, сравниваем как обычный IP
        return ip == cidr
    }
    
    clientIP := net.ParseIP(ip)
    return ipNet.Contains(clientIP)
}
```

## Аутентификация и авторизация

### JWT с поддержкой тенантов

```go
type JWTPayload struct {
    Subject   string `json:"sub"`
    TenantID  string `json:"tenant_id"`
    Issuer    string `json:"iss"`
    Audience  string `json:"aud"`
    ExpiresAt int64  `json:"exp"`
    IssuedAt  int64  `json:"iat"`
}

func validateTokenWithTenant(tokenString string, expectedTenantID string) (*JWTPayload, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTPayload{}, func(token *jwt.Token) (interface{}, error) {
        return jwtSecret, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("invalid token: %v", err)
    }
    
    if claims, ok := token.Claims.(*JWTPayload); ok && token.Valid {
        // Проверка тенанта
        if claims.TenantID != expectedTenantID {
            return nil, fmt.Errorf("tenant mismatch: expected %s, got %s", 
                expectedTenantID, claims.TenantID)
        }
        
        // Проверка активности тенанта
        tenant := getTenant(claims.TenantID)
        if !tenant.Enabled {
            return nil, fmt.Errorf("tenant %s is disabled", claims.TenantID)
        }
        
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token claims")
}
```

### Keycloak интеграция

```go
type KeycloakClaims struct {
    Subject   string   `json:"sub"`
    TenantID  string   `json:"tenant_id"`
    Roles     []string `json:"realm_access.roles"`
    Issuer    string   `json:"iss"`
    Audience  string   `json:"aud"`
    ExpiresAt int64    `json:"exp"`
}

func validateKeycloakToken(tokenString string, expectedTenantID string) (*KeycloakClaims, error) {
    // Получение публичного ключа из JWKS
    publicKey, err := getPublicKeyFromJWKS(tokenString)
    if err != nil {
        return nil, err
    }
    
    token, err := jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
        return publicKey, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("invalid keycloak token: %v", err)
    }
    
    if claims, ok := token.Claims.(*KeycloakClaims); ok && token.Valid {
        // Проверка тенанта
        if claims.TenantID != expectedTenantID {
            return nil, fmt.Errorf("tenant mismatch in keycloak token")
        }
        
        // Проверка ролей
        if !hasRequiredRole(claims.Roles, "relay-user") {
            return nil, fmt.Errorf("insufficient permissions")
        }
        
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid keycloak token claims")
}
```

## Мониторинг и метрики

### Prometheus метрики

```bash
# Метрики по тенантам
relay_tenant_tunnels_active{tenant_id="tenant_001"} 5
relay_tenant_connections_total{tenant_id="tenant_001"} 100
relay_tenant_bytes_sent_total{tenant_id="tenant_001"} 1048576
relay_tenant_bytes_received_total{tenant_id="tenant_001"} 2097152
relay_tenant_errors_total{tenant_id="tenant_001",error_type="limit_exceeded"} 2

# Метрики лимитов
relay_tenant_limit_usage{tenant_id="tenant_001",resource="tunnels"} 0.5
relay_tenant_limit_usage{tenant_id="tenant_001",resource="connections"} 0.8
relay_tenant_limit_usage{tenant_id="tenant_001",resource="bandwidth"} 0.3

# Метрики активности
relay_tenant_last_activity_seconds{tenant_id="tenant_001"} 1640995200
relay_tenant_uptime_seconds{tenant_id="tenant_001"} 86400
```

### Grafana дашборд

```sql
-- Использование ресурсов по тенантам
sum(relay_tenant_limit_usage) by (tenant_id, resource)

-- Трафик по тенантам
rate(relay_tenant_bytes_sent_total[5m]) + rate(relay_tenant_bytes_received_total[5m])

-- Ошибки по тенантам
sum(rate(relay_tenant_errors_total[5m])) by (tenant_id, error_type)

-- Активность тенантов
relay_tenant_last_activity_seconds

-- Топ тенантов по использованию ресурсов
topk(10, sum(relay_tenant_tunnels_active) by (tenant_id))
```

## Автоматическая очистка

### Политики очистки

```go
type CleanupPolicy struct {
    InactivityTimeout time.Duration `json:"inactivity_timeout"`
    MaxIdleTunnels    int           `json:"max_idle_tunnels"`
    CleanupInterval   time.Duration `json:"cleanup_interval"`
    ForceCleanup      bool          `json:"force_cleanup"`
}

func cleanupInactiveTenants() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            tenants := getAllTenants()
            
            for _, tenant := range tenants {
                if shouldCleanupTenant(tenant) {
                    cleanupTenant(tenant.ID)
                }
            }
        }
    }
}

func shouldCleanupTenant(tenant *Tenant) bool {
    // Проверка неактивности
    if time.Since(tenant.LastActivity) > tenant.CleanupPolicy.InactivityTimeout {
        return true
    }
    
    // Проверка превышения лимитов
    if tenant.Stats.ActiveTunnels > tenant.Limits.MaxTunnels {
        return true
    }
    
    // Принудительная очистка
    if tenant.CleanupPolicy.ForceCleanup {
        return true
    }
    
    return false
}

func cleanupTenant(tenantID string) {
    log.Infof("Cleaning up tenant: %s", tenantID)
    
    // Закрытие всех туннелей
    tunnels := getTenantTunnels(tenantID)
    for _, tunnel := range tunnels {
        closeTunnel(tunnel.ID)
    }
    
    // Сброс статистики
    resetTenantStats(tenantID)
    
    // Логирование
    log.Infof("Tenant %s cleaned up successfully", tenantID)
}
```

## API для управления тенантами

### Endpoints

```bash
# Получение списка тенантов
GET /api/v1/tenants
Authorization: Bearer YOUR_TOKEN

# Получение информации о тенанте
GET /api/v1/tenants/{tenant_id}
Authorization: Bearer YOUR_TOKEN

# Создание нового тенанта
POST /api/v1/tenants
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

# Обновление тенанта
PUT /api/v1/tenants/{tenant_id}
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

# Удаление тенанта
DELETE /api/v1/tenants/{tenant_id}
Authorization: Bearer YOUR_TOKEN

# Получение статистики тенанта
GET /api/v1/tenants/{tenant_id}/stats
Authorization: Bearer YOUR_TOKEN

# Очистка тенанта
POST /api/v1/tenants/{tenant_id}/cleanup
Authorization: Bearer YOUR_TOKEN
```

### Примеры запросов

```bash
# Создание тенанта
curl -X POST http://localhost:8081/api/v1/tenants \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant_004",
    "name": "New Company D",
    "limits": {
      "max_tunnels": 15,
      "max_connections_per_tunnel": 150,
      "max_bandwidth_mbps": 300,
      "max_data_transfer_gb": 3000,
      "rate_limit_per_second": 1500
    },
    "allowed_ips": ["192.168.4.0/24"],
    "metadata": {
      "contact": "admin@company-d.com",
      "environment": "production"
    }
  }'

# Получение статистики
curl http://localhost:8081/api/v1/tenants/tenant_001/stats \
  -H "Authorization: Bearer YOUR_TOKEN"

# Очистка тенанта
curl -X POST http://localhost:8081/api/v1/tenants/tenant_001/cleanup \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Безопасность

### Изоляция ресурсов

```go
// Изоляция туннелей по тенантам
func getTenantTunnels(tenantID string) []*Tunnel {
    tunnels := make([]*Tunnel, 0)
    
    for _, tunnel := range getAllTunnels() {
        if tunnel.TenantID == tenantID {
            tunnels = append(tunnels, tunnel)
        }
    }
    
    return tunnels
}

// Изоляция метрик по тенантам
func getTenantMetrics(tenantID string) *TenantMetrics {
    metrics := &TenantMetrics{
        TenantID: tenantID,
    }
    
    // Сбор метрик только для указанного тенанта
    metrics.ActiveTunnels = len(getTenantTunnels(tenantID))
    metrics.TotalConnections = getTenantConnections(tenantID)
    metrics.BytesSent = getTenantBytesSent(tenantID)
    metrics.BytesReceived = getTenantBytesReceived(tenantID)
    
    return metrics
}
```

### Аудит операций

```go
type AuditLog struct {
    Timestamp   time.Time `json:"timestamp"`
    TenantID    string    `json:"tenant_id"`
    UserID      string    `json:"user_id"`
    Action      string    `json:"action"`
    Resource    string    `json:"resource"`
    ResourceID  string    `json:"resource_id"`
    IPAddress   string    `json:"ip_address"`
    UserAgent   string    `json:"user_agent"`
    Success     bool      `json:"success"`
    Error       string    `json:"error,omitempty"`
}

func logAuditEvent(tenantID, userID, action, resource, resourceID, ipAddress, userAgent string, success bool, err error) {
    auditLog := &AuditLog{
        Timestamp:  time.Now(),
        TenantID:   tenantID,
        UserID:     userID,
        Action:     action,
        Resource:   resource,
        ResourceID: resourceID,
        IPAddress:  ipAddress,
        UserAgent:  userAgent,
        Success:    success,
    }
    
    if err != nil {
        auditLog.Error = err.Error()
    }
    
    // Сохранение в лог
    auditLogger.Info("audit_event", zap.Any("event", auditLog))
    
    // Отправка в систему мониторинга
    sendAuditEvent(auditLog)
}
```

## Troubleshooting

### Диагностика проблем

```bash
# Проверка активных тенантов
curl http://localhost:8081/api/v1/tenants \
  -H "Authorization: Bearer YOUR_TOKEN"

# Проверка лимитов тенанта
curl http://localhost:8081/api/v1/tenants/tenant_001/limits \
  -H "Authorization: Bearer YOUR_TOKEN"

# Проверка логов тенанта
grep "tenant_001" /var/log/relay/relay.log

# Проверка метрик тенанта
curl http://localhost:8081/metrics | grep "tenant_001"
```

### Частые проблемы

1. **Превышение лимитов**
   - Проверьте текущее использование ресурсов
   - Увеличьте лимиты при необходимости
   - Очистите неиспользуемые ресурсы

2. **IP-фильтрация**
   - Проверьте настройки allowed_ips и blocked_ips
   - Убедитесь в правильности CIDR нотации
   - Проверьте IP-адрес клиента

3. **Неактивные тенанты**
   - Проверьте настройки inactivity_timeout
   - Убедитесь в правильности last_activity
   - Настройте автоматическую очистку

4. **Проблемы с аутентификацией**
   - Проверьте JWT токен и tenant_id
   - Убедитесь в активности тенанта
   - Проверьте настройки Keycloak 