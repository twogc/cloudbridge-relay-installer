# Техническое задание: Интеграция Enhanced QUIC в CloudBridge Relay

## 📋 Обзор проекта

### Цель
Интегрировать модифицированный `quic-go` с Enhanced QUIC функциями в существующий проект CloudBridge Relay для обеспечения мультиарендности, DDoS защиты, географической маршрутизации и интеграции с Keycloak.

### Текущее состояние
- ✅ Enhanced QUIC реализован в `internal/quic/enhanced.go`
- ✅ Примеры использования созданы в `internal/quic/enhanced_example.go`
- ✅ Документация подготовлена в `internal/quic/README_ENHANCED.md`
- ✅ Код компилируется без ошибок

## 🎯 Задачи интеграции

### 1. Интеграция с существующим QUIC кодом

#### 1.1 Модификация `internal/quic/connection.go`
**Файл:** `internal/quic/connection.go`
**Задача:** Интегрировать Enhanced QUIC с существующим `QUICConnection`

```go
// Добавить импорт Enhanced QUIC
import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// Модифицировать QUICConnection для поддержки Enhanced features
type QUICConnection struct {
    // Существующие поля...
    conn net.Conn
    
    // Новые поля для Enhanced QUIC
    enhancedConn *quic.EnhancedConnection
    tenantID     string
    config       *quic.EnhancedConfig
}

// Добавить метод для создания Enhanced соединения
func (qc *QUICConnection) CreateEnhancedConnection(tenantID string, config *quic.EnhancedConfig) error {
    enhancedConn, err := quic.DialWithTenant("tcp", qc.conn.RemoteAddr().String(), tenantID, config)
    if err != nil {
        return err
    }
    qc.enhancedConn = enhancedConn
    qc.tenantID = tenantID
    qc.config = config
    return nil
}

// Модифицировать существующие методы для использования Enhanced QUIC
func (qc *QUICConnection) Read(b []byte) (n int, err error) {
    if qc.enhancedConn != nil {
        return qc.enhancedConn.Read(b)
    }
    return qc.conn.Read(b)
}

func (qc *QUICConnection) Write(b []byte) (n int, err error) {
    if qc.enhancedConn != nil {
        return qc.enhancedConn.Write(b)
    }
    return qc.conn.Write(b)
}
```

#### 1.2 Модификация `internal/quic/config.go`
**Файл:** `internal/quic/config.go`
**Задача:** Добавить поддержку Enhanced конфигурации

```go
// Добавить поля для Enhanced QUIC в существующую конфигурацию
type Config struct {
    // Существующие поля...
    MaxStreams        int
    IdleTimeout       time.Duration
    HandshakeTimeout  time.Duration
    
    // Новые поля для Enhanced QUIC
    EnableEnhancedQUIC bool
    EnhancedConfig     *EnhancedConfig
}

// Добавить метод для создания Enhanced конфигурации
func (c *Config) CreateEnhancedConfig() *EnhancedConfig {
    return &EnhancedConfig{
        MaxStreams:        c.MaxStreams,
        IdleTimeout:       c.IdleTimeout,
        HandshakeTimeout:  c.HandshakeTimeout,
        KeepAlive:         true,
        MaxIncomingStreams: c.MaxStreams,
        MaxIncomingUniStreams: c.MaxStreams,
        
        // Enhanced features
        EnableDDoSProtection: true,
        EnableGeographicRouting: true,
        EnableKeycloakIntegration: true,
    }
}
```

### 2. Интеграция с Tenant Manager

#### 2.1 Модификация `internal/tenant/tenant.go`
**Файл:** `internal/tenant/tenant.go`
**Задача:** Интегрировать с Enhanced QUIC TenantManager

```go
// Добавить импорт Enhanced QUIC
import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// Добавить поле для Enhanced TenantManager
type Manager struct {
    // Существующие поля...
    tenants map[string]*Tenant
    
    // Новое поле для Enhanced QUIC
    enhancedTenantManager *quic.TenantManager
}

// Добавить метод инициализации Enhanced TenantManager
func (m *Manager) InitEnhancedTenantManager() {
    m.enhancedTenantManager = quic.NewTenantManager()
    
    // Синхронизировать существующих тенантов с Enhanced
    for id, tenant := range m.tenants {
        m.enhancedTenantManager.AddTenant(id, tenant.Limits.BandwidthLimit, tenant.Limits.ConnectionLimit)
    }
}

// Добавить метод для получения Enhanced TenantManager
func (m *Manager) GetEnhancedTenantManager() *quic.TenantManager {
    return m.enhancedTenantManager
}
```

### 3. Интеграция с Rate Limiting

#### 3.1 Модификация `internal/ratelimit/ratelimit.go`
**Файл:** `internal/ratelimit/ratelimit.go`
**Задача:** Интегрировать с Enhanced QUIC DDoS Protection

```go
// Добавить импорт Enhanced QUIC
import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// Добавить поле для Enhanced DDoS Protection
type Limiter struct {
    // Существующие поля...
    rate    float64
    burst   int
    
    // Новое поле для Enhanced QUIC
    enhancedDDoSProtection *quic.DDoSProtection
}

// Добавить метод инициализации Enhanced DDoS Protection
func (l *Limiter) InitEnhancedDDoSProtection() {
    l.enhancedDDoSProtection = quic.NewDDoSProtection()
}

// Добавить метод для проверки Enhanced DDoS защиты
func (l *Limiter) CheckEnhancedDDoSProtection(clientIP string) error {
    if l.enhancedDDoSProtection != nil {
        return l.enhancedDDoSProtection.CheckRateLimit(clientIP)
    }
    return nil
}
```

### 4. Интеграция с Load Balancer

#### 4.1 Модификация `internal/balancer/balancer.go`
**Файл:** `internal/balancer/balancer.go`
**Задача:** Интегрировать с Enhanced QUIC Geographic Router

```go
// Добавить импорт Enhanced QUIC
import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// Добавить поле для Enhanced Geographic Router
type Balancer struct {
    // Существующие поля...
    servers []*Server
    
    // Новое поле для Enhanced QUIC
    enhancedGeoRouter *quic.GeographicRouter
}

// Добавить метод инициализации Enhanced Geographic Router
func (b *Balancer) InitEnhancedGeoRouter() {
    b.enhancedGeoRouter = quic.NewGeographicRouter()
}

// Добавить метод для географической маршрутизации
func (b *Balancer) RouteByEnhancedGeo(clientIP string) *Server {
    if b.enhancedGeoRouter != nil {
        route := b.enhancedGeoRouter.RouteByLocation(clientIP)
        // Найти сервер по маршруту
        for _, server := range b.servers {
            if server.Address == route {
                return server
            }
        }
    }
    // Fallback к обычной балансировке
    return b.SelectServer()
}
```

### 5. Интеграция с Keycloak

#### 5.1 Создание нового файла `internal/auth/keycloak_enhanced.go`
**Файл:** `internal/auth/keycloak_enhanced.go`
**Задача:** Интегрировать Enhanced QUIC Keycloak с существующей аутентификацией

```go
package auth

import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
)

// EnhancedKeycloakAuth расширяет существующую аутентификацию
type EnhancedKeycloakAuth struct {
    // Существующие поля...
    config *KeycloakConfig
    
    // Новое поле для Enhanced QUIC
    enhancedKeycloakAuth *quic.KeycloakAuth
}

// Создать новый Enhanced Keycloak Auth
func NewEnhancedKeycloakAuth(config *KeycloakConfig) *EnhancedKeycloakAuth {
    enhancedConfig := &quic.KeycloakConfig{
        ServerURL:    config.ServerURL,
        Realm:        config.Realm,
        ClientID:     config.ClientID,
        ClientSecret: config.ClientSecret,
    }
    
    return &EnhancedKeycloakAuth{
        config:              config,
        enhancedKeycloakAuth: quic.NewKeycloakAuth(enhancedConfig),
    }
}

// Аутентификация с Enhanced QUIC
func (eka *EnhancedKeycloakAuth) AuthenticateEnhanced(token string) (bool, error) {
    if eka.enhancedKeycloakAuth != nil {
        return eka.enhancedKeycloakAuth.AuthenticateUser(token)
    }
    // Fallback к обычной аутентификации
    return eka.Authenticate(token)
}
```

### 6. Модификация основного приложения

#### 6.1 Создание файла `cmd/relay/enhanced_quic.go`
**Файл:** `cmd/relay/enhanced_quic.go`
**Задача:** Добавить поддержку Enhanced QUIC в основное приложение

```go
package main

import (
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
    "github.com/twogc/cloudbridge-relay-installer/internal/tenant"
    "github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
    "github.com/twogc/cloudbridge-relay-installer/internal/balancer"
    "github.com/twogc/cloudbridge-relay-installer/internal/auth"
)

// EnhancedQUICManager управляет Enhanced QUIC
type EnhancedQUICManager struct {
    config           *quic.EnhancedConfig
    tenantManager    *quic.TenantManager
    ddosProtection   *quic.DDoSProtection
    geoRouter        *quic.GeographicRouter
    keycloakAuth     *quic.KeycloakAuth
}

// Создать новый Enhanced QUIC Manager
func NewEnhancedQUICManager() *EnhancedQUICManager {
    config := &quic.EnhancedConfig{
        MaxStreams:        100,
        KeepAlive:         true,
        IdleTimeout:       30 * time.Second,
        HandshakeTimeout:  10 * time.Second,
        MaxIncomingStreams: 100,
        MaxIncomingUniStreams: 100,
        
        // Enhanced features
        EnableDDoSProtection: true,
        EnableGeographicRouting: true,
        EnableKeycloakIntegration: true,
    }
    
    return &EnhancedQUICManager{
        config:         config,
        tenantManager:  quic.NewTenantManager(),
        ddosProtection: quic.NewDDoSProtection(),
        geoRouter:      quic.NewGeographicRouter(),
    }
}

// Инициализировать Enhanced QUIC
func (eqm *EnhancedQUICManager) Initialize() error {
    // Настроить тенантов
    eqm.config.TenantManager = eqm.tenantManager
    
    // Настроить Keycloak
    if eqm.config.EnableKeycloakIntegration {
        keycloakConfig := &quic.KeycloakConfig{
            ServerURL:    "https://keycloak.example.com",
            Realm:        "cloudbridge",
            ClientID:     "quic-client",
            ClientSecret: "secret",
        }
        eqm.keycloakAuth = quic.NewKeycloakAuth(keycloakConfig)
    }
    
    return nil
}

// Создать Enhanced соединение
func (eqm *EnhancedQUICManager) CreateConnection(network, address, tenantID string) (*quic.EnhancedConnection, error) {
    return quic.DialWithTenant(network, address, tenantID, eqm.config)
}
```

#### 6.2 Модификация `cmd/relay/main.go`
**Файл:** `cmd/relay/main.go`
**Задача:** Интегрировать Enhanced QUIC в основную функцию

```go
// Добавить импорт Enhanced QUIC
import (
    "github.com/twogc/cloudbridge-relay-installer/cmd/relay"
)

// Добавить поле для Enhanced QUIC Manager
type RelayServer struct {
    // Существующие поля...
    config *Config
    
    // Новое поле для Enhanced QUIC
    enhancedQUICManager *EnhancedQUICManager
}

// Инициализировать Enhanced QUIC в main
func main() {
    // Существующий код...
    
    // Инициализировать Enhanced QUIC
    enhancedQUICManager := NewEnhancedQUICManager()
    if err := enhancedQUICManager.Initialize(); err != nil {
        log.Fatalf("Failed to initialize Enhanced QUIC: %v", err)
    }
    
    // Добавить в RelayServer
    relayServer.enhancedQUICManager = enhancedQUICManager
    
    // Запустить сервер с поддержкой Enhanced QUIC
    if err := relayServer.StartWithEnhancedQUIC(); err != nil {
        log.Fatalf("Failed to start relay server: %v", err)
    }
}

// Добавить метод для запуска с Enhanced QUIC
func (rs *RelayServer) StartWithEnhancedQUIC() error {
    // Существующий код запуска сервера...
    
    // Добавить обработку Enhanced QUIC соединений
    go rs.handleEnhancedQUICConnections()
    
    return nil
}

// Обработка Enhanced QUIC соединений
func (rs *RelayServer) handleEnhancedQUICConnections() {
    for {
        // Принимать соединения
        conn, err := rs.listener.Accept()
        if err != nil {
            log.Printf("Failed to accept connection: %v", err)
            continue
        }
        
        // Создать Enhanced соединение
        enhancedConn, err := rs.enhancedQUICManager.CreateConnection("tcp", conn.RemoteAddr().String(), "default")
        if err != nil {
            log.Printf("Failed to create enhanced connection: %v", err)
            conn.Close()
            continue
        }
        
        // Обработать соединение
        go rs.handleEnhancedConnection(enhancedConn)
    }
}

// Обработка Enhanced соединения
func (rs *RelayServer) handleEnhancedConnection(conn *quic.EnhancedConnection) {
    defer conn.Close()
    
    // Получить статистику
    stats := conn.GetStats()
    log.Printf("Enhanced connection stats: %+v", stats)
    
    // Обработать данные
    buffer := make([]byte, 1024)
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            log.Printf("Failed to read from enhanced connection: %v", err)
            break
        }
        
        // Обработать данные
        log.Printf("Received %d bytes from tenant %s", n, conn.tenantID)
    }
}
```

## 🔧 Конфигурация

### 6.3 Создание конфигурационного файла `configs/enhanced_quic.yaml`
**Файл:** `configs/enhanced_quic.yaml`
**Задача:** Конфигурация Enhanced QUIC

```yaml
enhanced_quic:
  enabled: true
  
  # Базовые настройки
  max_streams: 100
  keep_alive: true
  idle_timeout: 30s
  handshake_timeout: 10s
  
  # Enhanced features
  ddos_protection:
    enabled: true
    rate_limit: 1000
    burst: 100
    blacklist_duration: 5m
  
  geographic_routing:
    enabled: true
    regions:
      us: "us-east.relay.example.com"
      eu: "eu-west.relay.example.com"
      as: "asia.relay.example.com"
  
  keycloak_integration:
    enabled: true
    server_url: "https://keycloak.example.com"
    realm: "cloudbridge"
    client_id: "quic-client"
    client_secret: "secret"
  
  # Tenant configuration
  tenants:
    premium:
      bandwidth_limit: 1073741824  # 1GB
      connection_limit: 100
    standard:
      bandwidth_limit: 104857600   # 100MB
      connection_limit: 10
```

## 🧪 Тестирование

### 7.1 Создание тестов `internal/quic/enhanced_test.go`
**Файл:** `internal/quic/enhanced_test.go`
**Задача:** Тестирование Enhanced QUIC

```go
package quic

import (
    "testing"
    "time"
)

func TestEnhancedQUICIntegration(t *testing.T) {
    // Тест создания Enhanced конфигурации
    config := &EnhancedConfig{
        MaxStreams:        100,
        KeepAlive:         true,
        IdleTimeout:       30 * time.Second,
        EnableDDoSProtection: true,
        EnableGeographicRouting: true,
        EnableKeycloakIntegration: true,
    }
    
    if config.MaxStreams != 100 {
        t.Errorf("Expected MaxStreams 100, got %d", config.MaxStreams)
    }
    
    // Тест создания TenantManager
    tenantManager := NewTenantManager()
    tenantManager.AddTenant("test", 1024*1024*100, 10)
    
    tenant, exists := tenantManager.GetTenant("test")
    if !exists {
        t.Error("Expected tenant to exist")
    }
    
    if tenant.BandwidthLimit != 1024*1024*100 {
        t.Errorf("Expected bandwidth limit %d, got %d", 1024*1024*100, tenant.BandwidthLimit)
    }
    
    // Тест DDoS Protection
    ddosProtection := NewDDoSProtection()
    err := ddosProtection.CheckRateLimit("192.168.1.1")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    // Тест Geographic Router
    geoRouter := NewGeographicRouter()
    route := geoRouter.RouteByLocation("192.168.1.1")
    if route == "" {
        t.Error("Expected route to be returned")
    }
    
    // Тест Keycloak Auth
    keycloakConfig := &KeycloakConfig{
        ServerURL: "https://keycloak.example.com",
        Realm:     "test",
    }
    keycloakAuth := NewKeycloakAuth(keycloakConfig)
    
    authenticated, err := keycloakAuth.AuthenticateUser("test-token")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if !authenticated {
        t.Error("Expected authentication to succeed")
    }
}
```

## 📊 Мониторинг

### 8.1 Создание метрик `internal/metrics/enhanced_quic.go`
**Файл:** `internal/metrics/enhanced_quic.go`
**Задача:** Метрики для Enhanced QUIC

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Enhanced QUIC метрики
    enhancedQuicConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "enhanced_quic_connections_total",
            Help: "Total number of Enhanced QUIC connections",
        },
        []string{"tenant_id", "status"},
    )
    
    enhancedQuicBytesSent = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "enhanced_quic_bytes_sent_total",
            Help: "Total bytes sent via Enhanced QUIC",
        },
        []string{"tenant_id"},
    )
    
    enhancedQuicBytesReceived = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "enhanced_quic_bytes_received_total",
            Help: "Total bytes received via Enhanced QUIC",
        },
        []string{"tenant_id"},
    )
    
    enhancedQuicDDoSBlocked = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "enhanced_quic_ddos_blocked_total",
            Help: "Total DDoS attacks blocked",
        },
        []string{"client_ip"},
    )
    
    enhancedQuicGeoRoutes = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "enhanced_quic_geo_routes_total",
            Help: "Total geographic routes selected",
        },
        []string{"region"},
    )
)

// RecordEnhancedQuicConnection записывает метрику соединения
func RecordEnhancedQuicConnection(tenantID, status string) {
    enhancedQuicConnections.WithLabelValues(tenantID, status).Inc()
}

// RecordEnhancedQuicBytes записывает метрики байтов
func RecordEnhancedQuicBytes(tenantID string, sent, received int64) {
    enhancedQuicBytesSent.WithLabelValues(tenantID).Add(float64(sent))
    enhancedQuicBytesReceived.WithLabelValues(tenantID).Add(float64(received))
}

// RecordEnhancedQuicDDoS записывает метрику DDoS блокировки
func RecordEnhancedQuicDDoS(clientIP string) {
    enhancedQuicDDoSBlocked.WithLabelValues(clientIP).Inc()
}

// RecordEnhancedQuicGeoRoute записывает метрику географической маршрутизации
func RecordEnhancedQuicGeoRoute(region string) {
    enhancedQuicGeoRoutes.WithLabelValues(region).Inc()
}
```

## 🚀 Развертывание

### 9.1 Создание Dockerfile `Dockerfile.enhanced`
**Файл:** `Dockerfile.enhanced`
**Задача:** Docker образ с Enhanced QUIC

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копировать go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копировать исходный код
COPY . .

# Собрать с Enhanced QUIC
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o relay-server ./cmd/relay

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копировать бинарный файл
COPY --from=builder /app/relay-server .

# Копировать конфигурацию
COPY configs/enhanced_quic.yaml ./configs/

# Экспортировать порты
EXPOSE 8080 9090

# Запустить сервер
CMD ["./relay-server", "--config", "./configs/enhanced_quic.yaml"]
```

### 9.2 Создание Kubernetes манифеста `k8s/enhanced-quic.yaml`
**Файл:** `k8s/enhanced-quic.yaml`
**Задача:** Kubernetes развертывание с Enhanced QUIC

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay-enhanced
  labels:
    app: cloudbridge-relay-enhanced
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cloudbridge-relay-enhanced
  template:
    metadata:
      labels:
        app: cloudbridge-relay-enhanced
    spec:
      containers:
      - name: relay-server
        image: cloudbridge-relay-enhanced:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: ENHANCED_QUIC_ENABLED
          value: "true"
        - name: KEYCLOAK_SERVER_URL
          value: "https://keycloak.example.com"
        - name: KEYCLOAK_REALM
          value: "cloudbridge"
        volumeMounts:
        - name: config
          mountPath: /app/configs
      volumes:
      - name: config
        configMap:
          name: enhanced-quic-config

---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-enhanced-service
spec:
  selector:
    app: cloudbridge-relay-enhanced
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: enhanced-quic-config
data:
  enhanced_quic.yaml: |
    enhanced_quic:
      enabled: true
      max_streams: 100
      keep_alive: true
      idle_timeout: 30s
      handshake_timeout: 10s
      ddos_protection:
        enabled: true
        rate_limit: 1000
        burst: 100
      geographic_routing:
        enabled: true
      keycloak_integration:
        enabled: true
```

## 📋 Чек-лист интеграции

### Этап 1: Подготовка (1-2 дня)
- [ ] Создать ветку для интеграции
- [ ] Проанализировать существующий код
- [ ] Создать план миграции
- [ ] Подготовить тестовую среду

### Этап 2: Базовая интеграция (3-5 дней)
- [ ] Модифицировать `internal/quic/connection.go`
- [ ] Модифицировать `internal/quic/config.go`
- [ ] Интегрировать с Tenant Manager
- [ ] Интегрировать с Rate Limiting
- [ ] Интегрировать с Load Balancer

### Этап 3: Расширенная интеграция (2-3 дня)
- [ ] Создать Enhanced Keycloak Auth
- [ ] Модифицировать основное приложение
- [ ] Добавить конфигурацию
- [ ] Создать метрики

### Этап 4: Тестирование (2-3 дня)
- [ ] Создать unit тесты
- [ ] Создать integration тесты
- [ ] Провести нагрузочное тестирование
- [ ] Провести тестирование безопасности

### Этап 5: Развертывание (1-2 дня)
- [ ] Создать Docker образ
- [ ] Создать Kubernetes манифесты
- [ ] Настроить мониторинг
- [ ] Развернуть в staging

### Этап 6: Production (1 день)
- [ ] Развернуть в production
- [ ] Мониторинг работы
- [ ] Документирование изменений

## 🎯 Ожидаемые результаты

### Функциональные улучшения
- ✅ Поддержка мультиарендности в QUIC
- ✅ DDoS защита на уровне QUIC
- ✅ Географическая маршрутизация
- ✅ Интеграция с Keycloak
- ✅ Улучшенная безопасность

### Технические улучшения
- ✅ Лучшая производительность
- ✅ Масштабируемость
- ✅ Мониторинг и метрики
- ✅ Простота развертывания

### Бизнес-улучшения
- ✅ Поддержка множественных клиентов
- ✅ Глобальное развертывание
- ✅ Улучшенная безопасность
- ✅ Снижение операционных расходов

## 📞 Поддержка

### Контакты
- **Разработчик:** [Имя разработчика]
- **Email:** [email]
- **Telegram:** [telegram]

### Документация
- `internal/quic/README_ENHANCED.md` - документация Enhanced QUIC
- `internal/quic/enhanced_example.go` - примеры использования
- `ENHANCED_QUIC_IMPLEMENTATION_COMPLETE.md` - отчет о реализации

### Ресурсы
- [QUIC Protocol RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000)
- [quic-go Documentation](https://github.com/quic-go/quic-go)
- [Keycloak Documentation](https://www.keycloak.org/documentation)

