# CloudBridge Relay Server - План доработки P2P функциональности

## 📋 Обзор

Данный документ содержит детальный план доработки P2P функциональности в CloudBridge Relay Server для достижения 95% готовности.

---

## 🎯 Цели доработки

### Основные цели
1. **Расширить JWT валидацию** для поддержки P2P-специфичных полей
2. **Реализовать peer whitelist** проверку и валидацию
3. **Добавить P2P-специфичные метрики** для мониторинга
4. **Улучшить конфигурацию** P2P компонентов
5. **Добавить уведомления** о недоступности P2P
6. **Оптимизировать fallback** механизм

### Целевые показатели
- **Готовность P2P**: с 75% до 95%
- **Безопасность**: полная валидация peer whitelist
- **Мониторинг**: 100% покрытие P2P метриками
- **Конфигурация**: гибкая настройка всех P2P параметров

---

## 📅 План выполнения

### Этап 1: Расширение JWT валидации (Высокий приоритет)
**Время**: 2-3 часа
**Файлы**: `internal/token/validator.go`, `internal/token/p2p_claims.go`

#### Задачи:
1. Создать структуры для P2P JWT claims
2. Расширить JWT валидацию для P2P полей
3. Добавить валидацию connection_type
4. Реализовать парсинг wireguard_config
5. Добавить поддержку mesh_config
6. Реализовать peer_whitelist валидацию
7. Добавить network_config поддержку

#### Результат:
```go
type P2PClaims struct {
    jwt.RegisteredClaims
    ConnectionType  string                 `json:"connection_type"`
    WireGuardConfig *WireGuardConfig      `json:"wireguard_config"`
    MeshConfig      *MeshConfig           `json:"mesh_config"`
    PeerWhitelist   []string              `json:"peer_whitelist"`
    NetworkConfig   *NetworkConfig        `json:"network_config"`
}
```

### Этап 2: Реализация peer whitelist (Высокий приоритет)
**Время**: 1-2 часа
**Файлы**: `internal/wireguard/peer_whitelist.go`, `internal/wireguard/wireguard.go`

#### Задачи:
1. Создать PeerWhitelist валидатор
2. Интегрировать с WireGuardInterface
3. Добавить проверку при добавлении пира
4. Реализовать логирование нарушений
5. Добавить метрики для whitelist

#### Результат:
```go
type PeerWhitelist struct {
    AllowedPeers map[string]bool
    mu          sync.RWMutex
    logger      *zap.Logger
}

func (pw *PeerWhitelist) IsPeerAllowed(peerID string) bool
func (pw *PeerWhitelist) AddPeer(peerID string) error
func (pw *PeerWhitelist) RemovePeer(peerID string) error
```

### Этап 3: P2P-специфичные метрики (Средний приоритет)
**Время**: 1-2 часа
**Файлы**: `internal/metrics/p2p_metrics.go`, `internal/metrics/metrics.go`

#### Задачи:
1. Создать P2P метрики структуру
2. Добавить метрики для P2P соединений
3. Реализовать метрики для peer discovery
4. Добавить mesh маршрутизации метрики
5. Интегрировать с основными метриками

#### Результат:
```go
type P2PMetrics struct {
    connectionsTotal      prometheus.Counter
    activePeers          prometheus.Gauge
    trafficBytes         *prometheus.CounterVec
    latencySeconds       *prometheus.HistogramVec
    handshakeDuration    prometheus.Histogram
    meshRoutes           prometheus.Gauge
    peerDiscoveryEvents  *prometheus.CounterVec
}
```

### Этап 4: Улучшение конфигурации (Средний приоритет)
**Время**: 1 час
**Файлы**: `configs/p2p_config.yaml`, `internal/config/p2p_config.go`

#### Задачи:
1. Создать P2P конфигурацию
2. Добавить валидацию конфигурации
3. Интегрировать с основным конфигом
4. Добавить примеры конфигурации

#### Результат:
```yaml
p2p:
  wireguard:
    enabled: true
    interface:
      name: "wg0"
      listen_port: 51820
      address: "10.0.0.1/24"
      mtu: 1420
  mesh:
    enabled: true
    registry_url: "https://registry.cloudbridge.com"
    heartbeat_interval: "30s"
    max_peers: 100
  security:
    peer_whitelist_enabled: true
    key_rotation_interval: "24h"
```

### Этап 5: Уведомления о недоступности P2P (Средний приоритет)
**Время**: 1 час
**Файлы**: `internal/wireguard/notifications.go`, `internal/wireguard/wireguard.go`

#### Задачи:
1. Создать систему уведомлений
2. Добавить callback для недоступности P2P
3. Реализовать уведомления клиента
4. Добавить логирование событий

#### Результат:
```go
type P2PNotificationManager struct {
    callbacks []P2PUnavailableCallback
    logger    *zap.Logger
}

type P2PUnavailableCallback func(reason string, fallback string)
```

### Этап 6: Оптимизация fallback механизма (Низкий приоритет)
**Время**: 30 минут
**Файлы**: `internal/wireguard/hybrid_router.go`

#### Задачи:
1. Улучшить логику fallback
2. Добавить метрики для fallback
3. Оптимизировать выбор маршрута

---

## 🔧 Детальный план реализации

### 1. Расширение JWT валидации

#### Создание P2P claims структуры
```go
// internal/token/p2p_claims.go
package token

import (
    "encoding/json"
    "fmt"
    "github.com/golang-jwt/jwt/v5"
)

type WireGuardConfig struct {
    InterfaceName string   `json:"interface_name"`
    ListenPort    int      `json:"listen_port"`
    Address       string   `json:"address"`
    MTU           int      `json:"mtu"`
    AllowedIPs    []string `json:"allowed_ips"`
}

type MeshConfig struct {
    RegistryURL       string        `json:"registry_url"`
    HeartbeatInterval string        `json:"heartbeat_interval"`
    MaxPeers          int           `json:"max_peers"`
    RoutingStrategy   string        `json:"routing_strategy"`
}

type NetworkConfig struct {
    Subnet           string   `json:"subnet"`
    Gateway          string   `json:"gateway"`
    DNS              []string `json:"dns"`
    FirewallRules    []string `json:"firewall_rules"`
}

type P2PClaims struct {
    jwt.RegisteredClaims
    ConnectionType  string          `json:"connection_type"`
    WireGuardConfig *WireGuardConfig `json:"wireguard_config"`
    MeshConfig      *MeshConfig     `json:"mesh_config"`
    PeerWhitelist   []string        `json:"peer_whitelist"`
    NetworkConfig   *NetworkConfig  `json:"network_config"`
}

func (p *P2PClaims) Validate() error {
    // Валидация P2P полей
    if p.ConnectionType == "" {
        return fmt.Errorf("connection_type is required")
    }
    
    if p.ConnectionType == "wireguard" && p.WireGuardConfig == nil {
        return fmt.Errorf("wireguard_config is required for wireguard connection type")
    }
    
    return nil
}
```

#### Расширение JWT валидатора
```go
// internal/token/validator.go - добавление методов
func (v *Validator) ValidateP2PToken(token string) (*P2PClaims, error) {
    claims := &P2PClaims{}
    
    _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return v.jwtKey, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to parse P2P JWT: %v", err)
    }
    
    if err := claims.Validate(); err != nil {
        return nil, fmt.Errorf("P2P claims validation failed: %v", err)
    }
    
    return claims, nil
}
```

### 2. Реализация peer whitelist

#### Создание PeerWhitelist валидатора
```go
// internal/wireguard/peer_whitelist.go
package wireguard

import (
    "fmt"
    "sync"
    "go.uber.org/zap"
)

type PeerWhitelist struct {
    AllowedPeers map[string]bool
    mu          sync.RWMutex
    logger      *zap.Logger
    metrics     *PeerWhitelistMetrics
}

type PeerWhitelistMetrics struct {
    checksTotal      prometheus.Counter
    allowedTotal     prometheus.Counter
    deniedTotal      prometheus.Counter
    violationsTotal  *prometheus.CounterVec
}

func NewPeerWhitelist(logger *zap.Logger) *PeerWhitelist {
    return &PeerWhitelist{
        AllowedPeers: make(map[string]bool),
        logger:       logger,
        metrics:      NewPeerWhitelistMetrics(),
    }
}

func (pw *PeerWhitelist) IsPeerAllowed(peerID string) bool {
    pw.mu.RLock()
    defer pw.mu.RUnlock()
    
    pw.metrics.checksTotal.Inc()
    
    allowed, exists := pw.AllowedPeers[peerID]
    if !exists || !allowed {
        pw.metrics.deniedTotal.Inc()
        pw.metrics.violationsTotal.WithLabelValues("peer_not_whitelisted").Inc()
        pw.logger.Warn("Peer not in whitelist",
            zap.String("peer_id", peerID),
        )
        return false
    }
    
    pw.metrics.allowedTotal.Inc()
    return true
}

func (pw *PeerWhitelist) AddPeer(peerID string) error {
    pw.mu.Lock()
    defer pw.mu.Unlock()
    
    pw.AllowedPeers[peerID] = true
    pw.logger.Info("Peer added to whitelist",
        zap.String("peer_id", peerID),
    )
    return nil
}

func (pw *PeerWhitelist) RemovePeer(peerID string) error {
    pw.mu.Lock()
    defer pw.mu.Unlock()
    
    delete(pw.AllowedPeers, peerID)
    pw.logger.Info("Peer removed from whitelist",
        zap.String("peer_id", peerID),
    )
    return nil
}

func (pw *PeerWhitelist) LoadFromJWT(peerWhitelist []string) error {
    pw.mu.Lock()
    defer pw.mu.Unlock()
    
    pw.AllowedPeers = make(map[string]bool)
    for _, peerID := range peerWhitelist {
        pw.AllowedPeers[peerID] = true
    }
    
    pw.logger.Info("Peer whitelist loaded from JWT",
        zap.Int("count", len(peerWhitelist)),
    )
    return nil
}
```

### 3. P2P-специфичные метрики

#### Создание P2P метрик
```go
// internal/metrics/p2p_metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

type P2PMetrics struct {
    connectionsTotal      prometheus.Counter
    activePeers          prometheus.Gauge
    trafficBytes         *prometheus.CounterVec
    latencySeconds       *prometheus.HistogramVec
    handshakeDuration    prometheus.Histogram
    meshRoutes           prometheus.Gauge
    peerDiscoveryEvents  *prometheus.CounterVec
    fallbackEvents       *prometheus.CounterVec
    whitelistViolations  *prometheus.CounterVec
}

func NewP2PMetrics(reg prometheus.Registerer) *P2PMetrics {
    m := &P2PMetrics{
        connectionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "cloudbridge_p2p_connections_total",
            Help: "Total number of P2P connections",
        }),
        activePeers: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "cloudbridge_p2p_active_peers",
            Help: "Number of active P2P peers",
        }),
        trafficBytes: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "cloudbridge_p2p_traffic_bytes_total",
            Help: "Total P2P traffic in bytes",
        }, []string{"direction", "peer_id"}),
        latencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
            Name:    "cloudbridge_p2p_latency_seconds",
            Help:    "P2P connection latency in seconds",
            Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0},
        }, []string{"peer_id"}),
        handshakeDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name:    "cloudbridge_p2p_handshake_duration_seconds",
            Help:    "P2P handshake duration in seconds",
            Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0},
        }),
        meshRoutes: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "cloudbridge_p2p_mesh_routes",
            Help: "Number of active mesh routes",
        }),
        peerDiscoveryEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "cloudbridge_p2p_discovery_events_total",
            Help: "Total peer discovery events",
        }, []string{"event_type"}),
        fallbackEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "cloudbridge_p2p_fallback_events_total",
            Help: "Total P2P fallback events",
        }, []string{"from_protocol", "to_protocol", "reason"}),
        whitelistViolations: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "cloudbridge_p2p_whitelist_violations_total",
            Help: "Total peer whitelist violations",
        }, []string{"violation_type"}),
    }
    
    reg.MustRegister(
        m.connectionsTotal,
        m.activePeers,
        m.trafficBytes,
        m.latencySeconds,
        m.handshakeDuration,
        m.meshRoutes,
        m.peerDiscoveryEvents,
        m.fallbackEvents,
        m.whitelistViolations,
    )
    
    return m
}

// Методы для работы с метриками
func (m *P2PMetrics) IncP2PConnections() {
    m.connectionsTotal.Inc()
}

func (m *P2PMetrics) SetActivePeers(count int) {
    m.activePeers.Set(float64(count))
}

func (m *P2PMetrics) IncP2PTraffic(direction, peerID string, bytes int64) {
    m.trafficBytes.WithLabelValues(direction, peerID).Add(float64(bytes))
}

func (m *P2PMetrics) ObserveP2PLatency(peerID string, duration float64) {
    m.latencySeconds.WithLabelValues(peerID).Observe(duration)
}

func (m *P2PMetrics) ObserveHandshakeDuration(duration float64) {
    m.handshakeDuration.Observe(duration)
}

func (m *P2PMetrics) SetMeshRoutes(count int) {
    m.meshRoutes.Set(float64(count))
}

func (m *P2PMetrics) IncPeerDiscoveryEvent(eventType string) {
    m.peerDiscoveryEvents.WithLabelValues(eventType).Inc()
}

func (m *P2PMetrics) IncFallbackEvent(fromProtocol, toProtocol, reason string) {
    m.fallbackEvents.WithLabelValues(fromProtocol, toProtocol, reason).Inc()
}

func (m *P2PMetrics) IncWhitelistViolation(violationType string) {
    m.whitelistViolations.WithLabelValues(violationType).Inc()
}
```

### 4. Улучшение конфигурации

#### Создание P2P конфигурации
```go
// internal/config/p2p_config.go
package config

import (
    "time"
    "gopkg.in/yaml.v2"
)

type P2PConfig struct {
    WireGuard WireGuardConfig `yaml:"wireguard"`
    Mesh      MeshConfig      `yaml:"mesh"`
    Security  SecurityConfig  `yaml:"security"`
}

type WireGuardConfig struct {
    Enabled  bool              `yaml:"enabled"`
    Interface WireGuardInterface `yaml:"interface"`
}

type WireGuardInterface struct {
    Name       string `yaml:"name"`
    ListenPort int    `yaml:"listen_port"`
    Address    string `yaml:"address"`
    MTU        int    `yaml:"mtu"`
}

type MeshConfig struct {
    Enabled           bool          `yaml:"enabled"`
    RegistryURL       string        `yaml:"registry_url"`
    HeartbeatInterval time.Duration `yaml:"heartbeat_interval"`
    MaxPeers          int           `yaml:"max_peers"`
    RoutingStrategy   string        `yaml:"routing_strategy"`
}

type SecurityConfig struct {
    PeerWhitelistEnabled bool          `yaml:"peer_whitelist_enabled"`
    KeyRotationInterval  time.Duration `yaml:"key_rotation_interval"`
    StrictMode           bool          `yaml:"strict_mode"`
}

func (c *P2PConfig) Validate() error {
    if c.WireGuard.Enabled && c.WireGuard.Interface.ListenPort == 0 {
        return fmt.Errorf("wireguard listen_port is required when enabled")
    }
    
    if c.Mesh.Enabled && c.Mesh.RegistryURL == "" {
        return fmt.Errorf("mesh registry_url is required when enabled")
    }
    
    return nil
}

func LoadP2PConfig(filename string) (*P2PConfig, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var config P2PConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    if err := config.Validate(); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### 5. Уведомления о недоступности P2P

#### Создание системы уведомлений
```go
// internal/wireguard/notifications.go
package wireguard

import (
    "sync"
    "go.uber.org/zap"
)

type P2PUnavailableCallback func(reason string, fallback string)

type P2PNotificationManager struct {
    callbacks []P2PUnavailableCallback
    mu        sync.RWMutex
    logger    *zap.Logger
}

func NewP2PNotificationManager(logger *zap.Logger) *P2PNotificationManager {
    return &P2PNotificationManager{
        callbacks: make([]P2PUnavailableCallback, 0),
        logger:    logger,
    }
}

func (nm *P2PNotificationManager) RegisterCallback(callback P2PUnavailableCallback) {
    nm.mu.Lock()
    defer nm.mu.Unlock()
    
    nm.callbacks = append(nm.callbacks, callback)
    nm.logger.Info("P2P notification callback registered")
}

func (nm *P2PNotificationManager) NotifyP2PUnavailable(reason, fallback string) {
    nm.mu.RLock()
    defer nm.mu.RUnlock()
    
    nm.logger.Warn("P2P unavailable, notifying callbacks",
        zap.String("reason", reason),
        zap.String("fallback", fallback),
    )
    
    for _, callback := range nm.callbacks {
        go func(cb P2PUnavailableCallback) {
            defer func() {
                if r := recover(); r != nil {
                    nm.logger.Error("P2P notification callback panicked",
                        zap.Any("panic", r),
                    )
                }
            }()
            cb(reason, fallback)
        }(callback)
    }
}
```

---

## 📊 Ожидаемые результаты

### После завершения доработки:

1. **JWT валидация**: 100% поддержка P2P полей
2. **Peer whitelist**: Полная валидация и логирование
3. **P2P метрики**: 100% покрытие мониторинга
4. **Конфигурация**: Гибкая настройка всех параметров
5. **Уведомления**: Информирование о недоступности P2P
6. **Fallback**: Оптимизированный механизм переключения

### Общая готовность P2P: 95% 🎯

---

## 🚀 Готов к выполнению

План детализирован и готов к реализации. Каждый этап содержит:
- ✅ Конкретные задачи
- ✅ Примеры кода
- ✅ Ожидаемые результаты
- ✅ Временные оценки

**Начинаем выполнение?** 🚀
