# 🚀 Техническое задание: Интеграция мультипротокольного режима (30% доработка)

## 📋 Обзор

**Цель**: Завершить интеграцию мультипротокольного режима в CloudBridge Relay, интегрировав протокольный селектор в основной поток обработки и создав унифицированную систему переключения между P2P и P2C протоколами.

**Приоритет**: Критический
**Сложность**: Средняя
**Время выполнения**: 1 неделя
**Готовность**: 70% → 100%

## 🎯 Что нужно реализовать (30%)

### 1. **Интеграция протокольного селектора** (День 1-2)

#### 1.1 Обновление Server структуры
```go
// cmd/relay/main.go - обновить структуру Server
type Server struct {
    // ... существующие поля ...
    
    // Мультипротокольные компоненты
    protocolSelector    *protocol.IntelligentProtocolSelector
    protocolEngine      *protocol.ProtocolEngine
    p2pManager          *P2PManager
    p2cManager          *P2CManager
    
    // Унифицированный обработчик
    connectionHandler   *UnifiedConnectionHandler
    
    // Автоматическое переключение
    autoSwitchManager   *AutoSwitchManager
}
```

#### 1.2 Создание UnifiedConnectionHandler
```go
// internal/protocol/unified_handler.go
package protocol

import (
    "context"
    "net"
    "time"
    
    "go.uber.org/zap"
    "github.com/prometheus/client_golang/prometheus"
)

type UnifiedConnectionHandler struct {
    protocolSelector *IntelligentProtocolSelector
    p2pManager       *P2PManager
    p2cManager       *P2CManager
    metrics          *ProtocolMetrics
    logger           *zap.Logger
}

func NewUnifiedConnectionHandler(
    selector *IntelligentProtocolSelector,
    p2pManager *P2PManager,
    p2cManager *P2CManager,
    logger *zap.Logger,
) *UnifiedConnectionHandler {
    return &UnifiedConnectionHandler{
        protocolSelector: selector,
        p2pManager:       p2pManager,
        p2cManager:       p2cManager,
        metrics:          NewProtocolMetrics(),
        logger:           logger,
    }
}

func (h *UnifiedConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    start := time.Now()
    
    // 1. Определение оптимального протокола
    protocol := h.protocolSelector.SelectBest()
    h.logger.Info("Selected protocol for connection",
        zap.String("protocol", protocol.String()),
        zap.String("client_ip", conn.RemoteAddr().String()),
        zap.String("tenant_id", tenantID),
    )
    
    // 2. Обработка через выбранный протокол
    var err error
    switch protocol {
    case QUIC:
        err = h.handleQUICConnection(ctx, conn, tenantID)
    case HTTP2:
        err = h.handleHTTP2Connection(ctx, conn, tenantID)
    case WebSocket:
        err = h.handleWebSocketConnection(ctx, conn, tenantID)
    default:
        err = h.handleFallbackConnection(ctx, conn, tenantID)
    }
    
    // 3. Обновление метрик
    duration := time.Since(start)
    h.metrics.RecordConnection(protocol, tenantID, duration, err == nil)
    
    return err
}

func (h *UnifiedConnectionHandler) handleQUICConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    // Проверяем, доступен ли P2P
    if h.p2pManager.IsAvailable(ctx, conn.RemoteAddr().String(), tenantID) {
        return h.p2pManager.HandleConnection(ctx, conn, tenantID)
    }
    
    // Fallback на P2C через QUIC
    return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

func (h *UnifiedConnectionHandler) handleHTTP2Connection(ctx context.Context, conn net.Conn, tenantID string) error {
    return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

func (h *UnifiedConnectionHandler) handleWebSocketConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}

func (h *UnifiedConnectionHandler) handleFallbackConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    h.logger.Warn("Using fallback connection",
        zap.String("client_ip", conn.RemoteAddr().String()),
        zap.String("tenant_id", tenantID),
    )
    return h.p2cManager.HandleConnection(ctx, conn, tenantID)
}
```

### 2. **P2P Manager** (День 2-3)

#### 2.1 Создание P2P Manager
```go
// internal/p2p/manager.go
package p2p

import (
    "context"
    "fmt"
    "net"
    "time"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
    "go.uber.org/zap"
)

type P2PManager struct {
    wireGuardMesh      *wireguard.MeshTopology
    peerDiscovery      *wireguard.PeerDiscovery
    hybridRouter       *wireguard.HybridRouter
    siteToSiteManager  *wireguard.SiteToSiteManager
    metrics            *P2PMetrics
    logger             *zap.Logger
}

func NewP2PManager(
    mesh *wireguard.MeshTopology,
    discovery *wireguard.PeerDiscovery,
    router *wireguard.HybridRouter,
    siteToSite *wireguard.SiteToSiteManager,
    logger *zap.Logger,
) *P2PManager {
    return &P2PManager{
        wireGuardMesh:     mesh,
        peerDiscovery:     discovery,
        hybridRouter:      router,
        siteToSiteManager: siteToSite,
        metrics:           NewP2PMetrics(),
        logger:            logger,
    }
}

func (pm *P2PManager) IsAvailable(ctx context.Context, clientIP string, tenantID string) bool {
    // Проверяем доступность P2P соединения
    peers := pm.peerDiscovery.GetActivePeers()
    if len(peers) < 3 {
        pm.logger.Debug("P2P not available: insufficient peers",
            zap.Int("peer_count", len(peers)),
            zap.String("tenant_id", tenantID),
        )
        return false
    }
    
    // Проверяем качество соединения
    latency := pm.measureLatency(ctx, clientIP)
    if latency > 50*time.Millisecond {
        pm.logger.Debug("P2P not available: high latency",
            zap.Duration("latency", latency),
            zap.String("tenant_id", tenantID),
        )
        return false
    }
    
    return true
}

func (pm *P2PManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    start := time.Now()
    
    // 1. Поиск оптимального P2P маршрута
    route, err := pm.hybridRouter.FindOptimalRoute(ctx, conn.RemoteAddr().String(), tenantID)
    if err != nil {
        pm.metrics.RecordConnection("p2p", "route_failed", tenantID, time.Since(start))
        return fmt.Errorf("failed to find P2P route: %w", err)
    }
    
    // 2. Установка P2P соединения
    if route.IsDirectP2P() {
        err = pm.handleDirectP2P(ctx, conn, route)
    } else {
        err = pm.handleMeshP2P(ctx, conn, route)
    }
    
    // 3. Обновление метрик
    status := "success"
    if err != nil {
        status = "failed"
    }
    pm.metrics.RecordConnection("p2p", status, tenantID, time.Since(start))
    
    return err
}

func (pm *P2PManager) handleDirectP2P(ctx context.Context, conn net.Conn, route *wireguard.Route) error {
    pm.logger.Info("Establishing direct P2P connection",
        zap.String("route", route.String()),
    )
    
    // Установка прямого P2P соединения
    return pm.wireGuardMesh.EstablishDirectConnection(ctx, conn, route)
}

func (pm *P2PManager) handleMeshP2P(ctx context.Context, conn net.Conn, route *wireguard.Route) error {
    pm.logger.Info("Establishing mesh P2P connection",
        zap.String("route", route.String()),
    )
    
    // Установка mesh P2P соединения
    return pm.wireGuardMesh.EstablishMeshConnection(ctx, conn, route)
}

func (pm *P2PManager) measureLatency(ctx context.Context, clientIP string) time.Duration {
    // Измерение задержки до клиента
    start := time.Now()
    
    // Простой ping для измерения задержки
    // В реальной реализации здесь будет более сложная логика
    time.Sleep(10 * time.Millisecond) // Симуляция
    
    return time.Since(start)
}
```

#### 2.2 P2P метрики
```go
// internal/metrics/p2p.go
package metrics

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
)

type P2PMetrics struct {
    connections *prometheus.CounterVec
    latency     *prometheus.HistogramVec
    throughput  *prometheus.CounterVec
    peerCount   prometheus.Gauge
}

func NewP2PMetrics() *P2PMetrics {
    return &P2PMetrics{
        connections: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "p2p_connections_total",
                Help: "Total number of P2P connections",
            },
            []string{"route_type", "status", "tenant_id"},
        ),
        latency: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "p2p_latency_seconds",
                Help: "P2P connection latency",
                Buckets: prometheus.DefBuckets,
            },
            []string{"route_type", "tenant_id"},
        ),
        throughput: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "p2p_throughput_bytes",
                Help: "P2P data throughput",
            },
            []string{"route_type", "tenant_id"},
        ),
        peerCount: prometheus.NewGauge(
            prometheus.GaugeOpts{
                Name: "p2p_active_peers",
                Help: "Number of active P2P peers",
            },
        ),
    }
}

func (pm *P2PMetrics) RecordConnection(routeType, status, tenantID string, duration time.Duration) {
    pm.connections.WithLabelValues(routeType, status, tenantID).Inc()
    pm.latency.WithLabelValues(routeType, tenantID).Observe(duration.Seconds())
}

func (pm *P2PMetrics) RecordThroughput(routeType, tenantID string, bytes int64) {
    pm.throughput.WithLabelValues(routeType, tenantID).Add(float64(bytes))
}

func (pm *P2PMetrics) SetPeerCount(count int) {
    pm.peerCount.Set(float64(count))
}
```

### 3. **P2C Manager** (День 3-4)

#### 3.1 Создание P2C Manager
```go
// internal/p2c/manager.go
package p2c

import (
    "context"
    "fmt"
    "net"
    "time"
    
    "github.com/twogc/cloudbridge-relay-installer/internal/tenant"
    "github.com/twogc/cloudbridge-relay-installer/internal/balancer"
    "github.com/twogc/cloudbridge-relay-installer/internal/ratelimit"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
    "go.uber.org/zap"
)

type P2CManager struct {
    enhancedQUICManager *EnhancedQUICManager
    tenantManager       *tenant.Manager
    loadBalancer        *balancer.Balancer
    ipRateLimit         *ratelimit.IPLimiter
    metrics             *P2CMetrics
    logger              *zap.Logger
}

func NewP2CManager(
    quicManager *EnhancedQUICManager,
    tenantManager *tenant.Manager,
    loadBalancer *balancer.Balancer,
    ipRateLimit *ratelimit.IPLimiter,
    logger *zap.Logger,
) *P2CManager {
    return &P2CManager{
        enhancedQUICManager: quicManager,
        tenantManager:       tenantManager,
        loadBalancer:        loadBalancer,
        ipRateLimit:         ipRateLimit,
        metrics:             NewP2CMetrics(),
        logger:              logger,
    }
}

func (pc *P2CManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    start := time.Now()
    
    // 1. Проверка лимитов тенанта
    if err := pc.tenantManager.CheckTenantLimits(tenantID, 1, 1); err != nil {
        pc.metrics.RecordConnection("p2c", "tenant_limit_exceeded", tenantID, time.Since(start))
        return fmt.Errorf("tenant limits exceeded: %w", err)
    }
    
    // 2. Проверка rate limiting
    clientIP := conn.RemoteAddr().String()
    if !pc.ipRateLimit.Allow(clientIP) {
        pc.metrics.RecordConnection("p2c", "rate_limited", tenantID, time.Since(start))
        return fmt.Errorf("rate limit exceeded for IP %s", clientIP)
    }
    
    // 3. Получение оптимального сервера
    targetServer, err := pc.loadBalancer.GetOptimalServer(ctx, clientIP, tenantID)
    if err != nil {
        pc.metrics.RecordConnection("p2c", "server_unavailable", tenantID, time.Since(start))
        return fmt.Errorf("failed to get optimal server: %w", err)
    }
    
    // 4. Обработка через Enhanced QUIC
    err = pc.enhancedQUICManager.HandleConnection(ctx, conn, targetServer, tenantID)
    
    // 5. Обновление метрик
    status := "success"
    if err != nil {
        status = "failed"
    }
    pc.metrics.RecordConnection("p2c", status, tenantID, time.Since(start))
    
    // 6. Обновление активности тенанта
    pc.tenantManager.UpdateTenantActivity(tenantID)
    
    return err
}
```

#### 3.2 P2C метрики
```go
// internal/metrics/p2c.go
package metrics

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
)

type P2CMetrics struct {
    connections *prometheus.CounterVec
    latency     *prometheus.HistogramVec
    throughput  *prometheus.CounterVec
    serverLoad  *prometheus.GaugeVec
}

func NewP2CMetrics() *P2CMetrics {
    return &P2CMetrics{
        connections: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "p2c_connections_total",
                Help: "Total number of P2C connections",
            },
            []string{"server", "status", "tenant_id"},
        ),
        latency: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "p2c_latency_seconds",
                Help: "P2C connection latency",
                Buckets: prometheus.DefBuckets,
            },
            []string{"server", "tenant_id"},
        ),
        throughput: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "p2c_throughput_bytes",
                Help: "P2C data throughput",
            },
            []string{"server", "tenant_id"},
        ),
        serverLoad: prometheus.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "p2c_server_load",
                Help: "P2C server load",
            },
            []string{"server"},
        ),
    }
}

func (pc *P2CMetrics) RecordConnection(server, status, tenantID string, duration time.Duration) {
    pc.connections.WithLabelValues(server, status, tenantID).Inc()
    pc.latency.WithLabelValues(server, tenantID).Observe(duration.Seconds())
}

func (pc *P2CMetrics) RecordThroughput(server, tenantID string, bytes int64) {
    pc.throughput.WithLabelValues(server, tenantID).Add(float64(bytes))
}

func (pc *P2CMetrics) SetServerLoad(server string, load float64) {
    pc.serverLoad.WithLabelValues(server).Set(load)
}
```

### 4. **Автоматическое переключение протоколов** (День 4-5)

#### 4.1 AutoSwitchManager
```go
// internal/protocol/auto_switch.go
package protocol

import (
    "context"
    "fmt"
    "time"
    
    "go.uber.org/zap"
)

type AutoSwitchManager struct {
    selector    *IntelligentProtocolSelector
    p2pManager  *P2PManager
    p2cManager  *P2CManager
    metrics     *SwitchMetrics
    logger      *zap.Logger
    
    switchCooldown time.Duration
    lastSwitch     time.Time
}

func NewAutoSwitchManager(
    selector *IntelligentProtocolSelector,
    p2pManager *P2PManager,
    p2cManager *P2CManager,
    logger *zap.Logger,
) *AutoSwitchManager {
    return &AutoSwitchManager{
        selector:       selector,
        p2pManager:     p2pManager,
        p2cManager:     p2cManager,
        metrics:        NewSwitchMetrics(),
        logger:         logger,
        switchCooldown: 5 * time.Minute,
    }
}

func (asm *AutoSwitchManager) Start(ctx context.Context) {
    ticker := time.NewTicker(time.Second * 30)
    defer ticker.Stop()
    
    asm.logger.Info("AutoSwitchManager started")
    
    for {
        select {
        case <-ctx.Done():
            asm.logger.Info("AutoSwitchManager stopped")
            return
        case <-ticker.C:
            asm.checkAndSwitch(ctx)
        }
    }
}

func (asm *AutoSwitchManager) checkAndSwitch(ctx context.Context) {
    // Проверяем cooldown период
    if time.Since(asm.lastSwitch) < asm.switchCooldown {
        return
    }
    
    currentProtocol := asm.selector.Current()
    optimalProtocol := asm.selector.SelectBest()
    
    if currentProtocol != optimalProtocol {
        asm.logger.Info("Protocol switch needed",
            zap.String("from", currentProtocol.String()),
            zap.String("to", optimalProtocol.String()),
        )
        
        if err := asm.performSwitch(ctx, currentProtocol, optimalProtocol); err != nil {
            asm.logger.Error("Failed to switch protocol", zap.Error(err))
            return
        }
        
        asm.lastSwitch = time.Now()
        asm.metrics.RecordSwitch(currentProtocol, optimalProtocol, "auto")
    }
}

func (asm *AutoSwitchManager) performSwitch(ctx context.Context, from, to Protocol) error {
    asm.logger.Info("Performing protocol switch",
        zap.String("from", from.String()),
        zap.String("to", to.String()),
    )
    
    // 1. Graceful shutdown текущего протокола
    if err := asm.gracefulShutdown(ctx, from); err != nil {
        return fmt.Errorf("failed to shutdown %s: %w", from.String(), err)
    }
    
    // 2. Инициализация нового протокола
    if err := asm.initializeProtocol(ctx, to); err != nil {
        return fmt.Errorf("failed to initialize %s: %w", to.String(), err)
    }
    
    // 3. Обновление селектора
    asm.selector.SetCurrent(to)
    
    asm.logger.Info("Protocol switch completed",
        zap.String("from", from.String()),
        zap.String("to", to.String()),
    )
    
    return nil
}

func (asm *AutoSwitchManager) gracefulShutdown(ctx context.Context, protocol Protocol) error {
    switch protocol {
    case QUIC:
        // Graceful shutdown QUIC
        return nil
    case HTTP2:
        // Graceful shutdown HTTP2
        return nil
    case WebSocket:
        // Graceful shutdown WebSocket
        return nil
    default:
        return nil
    }
}

func (asm *AutoSwitchManager) initializeProtocol(ctx context.Context, protocol Protocol) error {
    switch protocol {
    case QUIC:
        // Инициализация QUIC
        return nil
    case HTTP2:
        // Инициализация HTTP2
        return nil
    case WebSocket:
        // Инициализация WebSocket
        return nil
    default:
        return fmt.Errorf("unknown protocol: %s", protocol.String())
    }
}
```

#### 4.2 Switch метрики
```go
// internal/metrics/switch.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/twogc/cloudbridge-relay-installer/internal/protocol"
)

type SwitchMetrics struct {
    switches *prometheus.CounterVec
    duration *prometheus.HistogramVec
}

func NewSwitchMetrics() *SwitchMetrics {
    return &SwitchMetrics{
        switches: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "protocol_switches_total",
                Help: "Total number of protocol switches",
            },
            []string{"from", "to", "reason"},
        ),
        duration: prometheus.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "protocol_switch_duration_seconds",
                Help: "Duration of protocol switches",
                Buckets: prometheus.DefBuckets,
            },
            []string{"from", "to"},
        ),
    }
}

func (sm *SwitchMetrics) RecordSwitch(from, to protocol.Protocol, reason string) {
    sm.switches.WithLabelValues(from.String(), to.String(), reason).Inc()
}
```

### 5. **Интеграция в основной код** (День 5-6)

#### 5.1 Обновление NewServer
```go
// cmd/relay/main.go - обновить функцию NewServer
func NewServer(cfg *config.Config) (*Server, error) {
    // ... существующий код ...
    
    // Инициализация мультипротокольных компонентов
    protocolSelector := protocol.NewIntelligentProtocolSelector(protocol.ProtocolSelectorConfig{
        Strategy:             protocol.PerformanceBased,
        PreferredOrder:       []protocol.Protocol{protocol.QUIC, protocol.HTTP2, protocol.WebSocket},
        SwitchThreshold:      cfg.Protocol.SwitchThreshold,
        HealthCheckInterval:  cfg.Protocol.HealthCheckInterval,
        Logger:               loggerInstance,
    })
    
    p2pManager := p2p.NewP2PManager(
        wireGuardMesh,
        wireGuardPeerDiscovery,
        hybridRouter,
        siteToSiteManager,
        loggerInstance,
    )
    
    p2cManager := p2c.NewP2CManager(
        enhancedQUICManager,
        tenantManager,
        loadBalancer,
        ipRateLimit,
        loggerInstance,
    )
    
    connectionHandler := protocol.NewUnifiedConnectionHandler(
        protocolSelector,
        p2pManager,
        p2cManager,
        loggerInstance,
    )
    
    autoSwitchManager := protocol.NewAutoSwitchManager(
        protocolSelector,
        p2pManager,
        p2cManager,
        loggerInstance,
    )
    
    server := &Server{
        // ... существующие поля ...
        
        // Мультипротокольные компоненты
        protocolSelector:    protocolSelector,
        protocolEngine:      protocol.NewProtocolEngine([]protocol.Protocol{protocol.QUIC, protocol.HTTP2, protocol.WebSocket}, cfg.Protocol.SwitchThreshold, loggerInstance),
        p2pManager:          p2pManager,
        p2cManager:          p2cManager,
        connectionHandler:   connectionHandler,
        autoSwitchManager:   autoSwitchManager,
    }
    
    return server, nil
}
```

#### 5.2 Обновление processConnection
```go
// cmd/relay/main.go - обновить функцию processConnection
func (s *Server) processConnection(conn net.Conn) error {
    // ... существующий код аутентификации до получения tenantID ...
    
    // После успешной аутентификации:
    tenantID := s.extractTenantIDFromToken(token)
    if tenantID == "" {
        tenantID = "default"
    }
    
    // Использование унифицированного обработчика
    return s.connectionHandler.HandleConnection(s.ctx, conn, tenantID)
}

func (s *Server) extractTenantIDFromToken(token string) string {
    // Извлечение tenant_id из JWT токена
    // В реальной реализации здесь будет парсинг JWT
    return "default" // Временная реализация
}
```

#### 5.3 Обновление Start
```go
// cmd/relay/main.go - обновить функцию Start
func (s *Server) Start() error {
    // ... существующий код ...
    
    // Запуск автоматического переключения протоколов
    if s.autoSwitchManager != nil {
        go func() {
            s.autoSwitchManager.Start(s.ctx)
        }()
        s.logger.Info("AutoSwitchManager started")
    }
    
    // ... остальной код ...
}
```

### 6. **Тестирование** (День 6-7)

#### 6.1 Unit тесты
```go
// internal/protocol/unified_handler_test.go
package protocol

import (
    "context"
    "net"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestUnifiedConnectionHandler(t *testing.T) {
    // Создание моков
    mockSelector := &MockIntelligentProtocolSelector{}
    mockP2PManager := &MockP2PManager{}
    mockP2CManager := &MockP2CManager{}
    
    handler := NewUnifiedConnectionHandler(
        mockSelector,
        mockP2PManager,
        mockP2CManager,
        zap.NewNop(),
    )
    
    tests := []struct {
        name           string
        selectedProtocol Protocol
        p2pAvailable   bool
        expectedError  bool
    }{
        {
            name:            "P2P available",
            selectedProtocol: QUIC,
            p2pAvailable:    true,
            expectedError:   false,
        },
        {
            name:            "P2C fallback",
            selectedProtocol: HTTP2,
            p2pAvailable:    false,
            expectedError:   false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Настройка моков
            mockSelector.On("SelectBest").Return(tt.selectedProtocol)
            mockP2PManager.On("IsAvailable", mock.Anything, mock.Anything, mock.Anything).Return(tt.p2pAvailable)
            
            if tt.selectedProtocol == QUIC && tt.p2pAvailable {
                mockP2PManager.On("HandleConnection", mock.Anything, mock.Anything, mock.Anything).Return(nil)
            } else {
                mockP2CManager.On("HandleConnection", mock.Anything, mock.Anything, mock.Anything).Return(nil)
            }
            
            // Создание тестового соединения
            conn := &MockConn{}
            
            // Выполнение теста
            err := handler.HandleConnection(context.Background(), conn, "test-tenant")
            
            // Проверка результатов
            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
            
            // Проверка вызовов моков
            mockSelector.AssertExpectations(t)
            mockP2PManager.AssertExpectations(t)
            mockP2CManager.AssertExpectations(t)
        })
    }
}
```

#### 6.2 Integration тесты
```go
// cmd/relay/multiprotocol_integration_test.go
package main

import (
    "context"
    "net"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMultiprotocolIntegration(t *testing.T) {
    // Создание тестового сервера
    cfg := createTestConfig()
    server, err := NewServer(cfg)
    require.NoError(t, err)
    require.NotNil(t, server)
    
    defer server.Stop()
    
    // Запуск сервера
    go func() {
        err := server.Start()
        require.NoError(t, err)
    }()
    
    // Ожидание запуска
    time.Sleep(100 * time.Millisecond)
    
    t.Run("P2P Connection", func(t *testing.T) {
        conn, err := net.Dial("tcp", "localhost:8080")
        require.NoError(t, err)
        defer conn.Close()
        
        // Аутентификация
        err = authenticateConnection(t, conn)
        require.NoError(t, err)
        
        // Создание туннеля
        tunnel, err := createTunnel(t, conn, "test-tenant")
        require.NoError(t, err)
        
        // Проверка, что используется P2P (если доступен)
        assert.NotEmpty(t, tunnel.Protocol)
    })
    
    t.Run("P2C Connection", func(t *testing.T) {
        conn, err := net.Dial("tcp", "localhost:8080")
        require.NoError(t, err)
        defer conn.Close()
        
        // Аутентификация
        err = authenticateConnection(t, conn)
        require.NoError(t, err)
        
        // Создание туннеля
        tunnel, err := createTunnel(t, conn, "test-tenant")
        require.NoError(t, err)
        
        // Проверка, что используется P2C
        assert.NotEmpty(t, tunnel.Protocol)
    })
    
    t.Run("Auto Switch", func(t *testing.T) {
        // Тест автоматического переключения протоколов
        // Симуляция изменения условий сети
        time.Sleep(2 * time.Second)
        
        // Проверка, что переключение произошло
        // (в реальном тесте здесь будет проверка метрик)
    })
}

func createTestConfig() *config.Config {
    // Создание тестовой конфигурации
    return &config.Config{
        Server: config.ServerConfig{
            Port: 8080,
            Host: "localhost",
        },
        Protocol: config.Protocol{
            AutoSelect:      true,
            Preferred:       "quic",
            Fallback:        "http2",
            SwitchThreshold: 0.8,
            HealthCheckInterval: 30 * time.Second,
        },
        WireGuard: config.WireGuard{
            Enabled: true,
            Interface: struct {
                Name       string `json:"name"`
                ListenPort int    `json:"listen_port"`
                Address    string `json:"address"`
            }{
                Name:       "wg-test",
                ListenPort: 51820,
                Address:    "10.100.1.1/24",
            },
            Mesh: struct {
                Enabled           bool   `json:"enabled"`
                AutoDiscovery     bool   `json:"auto_discovery"`
                MaxConnections    int    `json:"max_connections"`
                RoutingStrategy   string `json:"routing_strategy"`
            }{
                Enabled:        true,
                AutoDiscovery:  true,
                MaxConnections: 10,
                RoutingStrategy: "hybrid",
            },
        },
        EnhancedQUIC: config.EnhancedQUIC{
            Enabled: true,
            MaxStreams: 100,
            KeepAlive: true,
            IdleTimeout: 30 * time.Second,
            HandshakeTimeout: 10 * time.Second,
            DDoSProtection: true,
            GeographicRouting: true,
            KeycloakIntegration: true,
        },
    }
}

func authenticateConnection(t *testing.T, conn net.Conn) error {
    // Реализация аутентификации для тестов
    // Отправка hello и auth сообщений
    return nil
}

func createTunnel(t *testing.T, conn net.Conn, tenantID string) (*TunnelInfo, error) {
    // Реализация создания туннеля для тестов
    // Отправка tunnel_info и получение ответа
    return &TunnelInfo{
        ID:       "test-tunnel",
        Protocol: "quic",
        Endpoint: "localhost:8080",
    }, nil
}

type TunnelInfo struct {
    ID       string
    Protocol string
    Endpoint string
}
```

## 📊 Метрики и мониторинг

### Обновленные метрики
```go
// internal/metrics/multiprotocol.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
)

var (
    // Общие метрики мультипротокольности
    ProtocolSwitches = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "protocol_switches_total",
            Help: "Total number of protocol switches",
        },
        []string{"from", "to", "reason"},
    )
    
    ProtocolPerformance = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "protocol_performance_score",
            Help: "Performance score for each protocol",
        },
        []string{"protocol", "metric"},
    )
    
    ConnectionDistribution = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "connection_distribution",
            Help: "Distribution of connections by protocol",
        },
        []string{"protocol", "tenant_id"},
    )
    
    // Метрики автоматического переключения
    AutoSwitchDecisions = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "auto_switch_decisions_total",
            Help: "Total number of auto-switch decisions",
        },
        []string{"decision", "reason"},
    )
    
    SwitchLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "protocol_switch_latency_seconds",
            Help: "Time taken to switch protocols",
            Buckets: prometheus.DefBuckets,
        },
        []string{"from", "to"},
    )
)
```

## ✅ Критерии приемки

### Функциональные требования
- [ ] Унифицированный обработчик соединений работает корректно
- [ ] P2P Manager правильно обрабатывает P2P соединения
- [ ] P2C Manager правильно обрабатывает P2C соединения
- [ ] Автоматическое переключение протоколов работает
- [ ] Интеграция с существующей аутентификацией

### Нефункциональные требования
- [ ] Время переключения протокола < 5 секунд
- [ ] Покрытие тестами > 90%
- [ ] Метрики собираются корректно
- [ ] Логирование работает правильно

### Производительность
- [ ] Overhead мультипротокольности < 5%
- [ ] Правильный выбор протокола в 95% случаев
- [ ] Стабильная работа при высокой нагрузке

## 🎯 Результат

После выполнения задания CloudBridge Relay будет:

1. **100% мультипротокольным** - полная интеграция P2P и P2C
2. **Интеллектуальным** - автоматический выбор оптимального протокола
3. **Отказоустойчивым** - плавное переключение при проблемах
4. **Мониторируемым** - детальная видимость всех протоколов
5. **Production ready** - готова к развертыванию

---

**Готово к production! 🚀**
