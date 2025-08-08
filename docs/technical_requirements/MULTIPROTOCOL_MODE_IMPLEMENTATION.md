# 🚀 Техническое задание: Мультипротокольный режим (P2P + P2C)

## 📋 Обзор

**Цель**: Доработать CloudBridge Relay сервер для полной поддержки мультипротокольного режима с автоматическим переключением между P2P (Peer-to-Peer) и P2C (Peer-to-Cloud) протоколами.

**Приоритет**: Критический
**Сложность**: Средняя
**Время выполнения**: 1-2 недели

## 🎯 Текущее состояние

### ✅ Готово:
- Базовая архитектура протоколов (HTTP2, QUIC, WebSocket)
- WireGuard P2P компоненты (интерфейсы, mesh, discovery)
- Enhanced QUIC P2C компоненты (manager, tenant, load balancer)
- Конфигурация для включения/выключения модулей

### ❌ Требует доработки:
- Интеграция протокольного селектора в основной поток обработки
- Автоматическое переключение между P2P и P2C
- Унифицированный API для всех протоколов
- Метрики и мониторинг мультипротокольности

## 🔧 Задачи доработки

### 1. **Интеграция протокольного селектора** (День 1-2)

#### 1.1 Обновление Server структуры
```go
// cmd/relay/main.go
type Server struct {
    // ... существующие поля ...
    
    // Мультипротокольные компоненты
    protocolSelector    *protocol.IntelligentProtocolSelector
    protocolEngine      *protocol.ProtocolEngine
    p2pManager          *P2PManager
    p2cManager          *P2CManager
    
    // Унифицированный обработчик
    connectionHandler   *UnifiedConnectionHandler
}
```

#### 1.2 Создание унифицированного обработчика
```go
// internal/protocol/unified_handler.go
type UnifiedConnectionHandler struct {
    protocolSelector *IntelligentProtocolSelector
    p2pManager       *P2PManager
    p2cManager       *P2CManager
    metrics          *ProtocolMetrics
    logger           *zap.Logger
}

func (h *UnifiedConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn) error {
    // 1. Определение оптимального протокола
    protocol := h.protocolSelector.SelectBest()
    
    // 2. Обработка через выбранный протокол
    switch protocol {
    case QUIC:
        return h.handleQUICConnection(ctx, conn)
    case HTTP2:
        return h.handleHTTP2Connection(ctx, conn)
    case WebSocket:
        return h.handleWebSocketConnection(ctx, conn)
    default:
        return h.handleFallbackConnection(ctx, conn)
    }
}
```

### 2. **P2P Manager** (День 3-4)

#### 2.1 Создание P2P Manager
```go
// internal/p2p/manager.go
type P2PManager struct {
    wireGuardMesh      *wireguard.MeshTopology
    peerDiscovery      *wireguard.PeerDiscovery
    hybridRouter       *wireguard.HybridRouter
    siteToSiteManager  *wireguard.SiteToSiteManager
    metrics            *P2PMetrics
    logger             *zap.Logger
}

func (pm *P2PManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    // 1. Поиск оптимального P2P маршрута
    route, err := pm.hybridRouter.FindOptimalRoute(ctx, conn.RemoteAddr().String(), tenantID)
    if err != nil {
        return fmt.Errorf("failed to find P2P route: %w", err)
    }
    
    // 2. Установка P2P соединения
    if route.IsDirectP2P() {
        return pm.handleDirectP2P(ctx, conn, route)
    } else {
        return pm.handleMeshP2P(ctx, conn, route)
    }
}
```

#### 2.2 P2P метрики
```go
// internal/metrics/p2p.go
var (
    P2PConnections = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "p2p_connections_total",
            Help: "Total number of P2P connections",
        },
        []string{"route_type", "status", "tenant_id"},
    )
    
    P2PLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "p2p_latency_seconds",
            Help: "P2P connection latency",
        },
        []string{"route_type", "tenant_id"},
    )
    
    P2PThroughput = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "p2p_throughput_bytes",
            Help: "P2P data throughput",
        },
        []string{"route_type", "tenant_id"},
    )
)
```

### 3. **P2C Manager** (День 5-6)

#### 3.1 Создание P2C Manager
```go
// internal/p2c/manager.go
type P2CManager struct {
    enhancedQUICManager *EnhancedQUICManager
    tenantManager       *tenant.Manager
    loadBalancer        *balancer.Balancer
    ipRateLimit         *ratelimit.IPLimiter
    metrics             *P2CMetrics
    logger              *zap.Logger
}

func (pc *P2CManager) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    // 1. Проверка лимитов тенанта
    if err := pc.tenantManager.CheckTenantLimits(tenantID, 1, 1); err != nil {
        return fmt.Errorf("tenant limits exceeded: %w", err)
    }
    
    // 2. Получение оптимального сервера
    targetServer, err := pc.loadBalancer.GetOptimalServer(ctx, conn.RemoteAddr().String(), tenantID)
    if err != nil {
        return fmt.Errorf("failed to get optimal server: %w", err)
    }
    
    // 3. Обработка через Enhanced QUIC
    return pc.enhancedQUICManager.HandleConnection(ctx, conn, targetServer, tenantID)
}
```

#### 3.2 P2C метрики
```go
// internal/metrics/p2c.go
var (
    P2CConnections = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "p2c_connections_total",
            Help: "Total number of P2C connections",
        },
        []string{"server", "status", "tenant_id"},
    )
    
    P2CLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "p2c_latency_seconds",
            Help: "P2C connection latency",
        },
        []string{"server", "tenant_id"},
    )
    
    P2CThroughput = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "p2c_throughput_bytes",
            Help: "P2C data throughput",
        },
        []string{"server", "tenant_id"},
    )
)
```

### 4. **Интеллектуальное переключение протоколов** (День 7-8)

#### 4.1 Алгоритм выбора протокола
```go
// internal/protocol/selector.go
func (ps *IntelligentProtocolSelector) SelectBest() Protocol {
    ps.mu.RLock()
    defer ps.mu.RUnlock()
    
    // 1. Проверка доступности P2P
    if ps.isP2PAvailable() && ps.shouldUseP2P() {
        return QUIC // P2P через QUIC
    }
    
    // 2. Проверка производительности P2C
    if ps.isP2COptimal() {
        return HTTP2 // P2C через HTTP2
    }
    
    // 3. Fallback на WebSocket
    return WebSocket
}

func (ps *IntelligentProtocolSelector) shouldUseP2P() bool {
    // Критерии для выбора P2P:
    // - Низкая задержка (< 50ms)
    // - Высокая пропускная способность (> 100 Mbps)
    // - Стабильное соединение
    // - Достаточно пиров в сети
    
    latency := ps.getCurrentLatency()
    throughput := ps.getCurrentThroughput()
    peerCount := ps.getActivePeerCount()
    
    return latency < 50*time.Millisecond &&
           throughput > 100*1024*1024 && // 100 Mbps
           peerCount >= 3
}
```

#### 4.2 Автоматическое переключение
```go
// internal/protocol/auto_switch.go
type AutoSwitchManager struct {
    selector    *IntelligentProtocolSelector
    p2pManager  *P2PManager
    p2cManager  *P2CManager
    metrics     *SwitchMetrics
    logger      *zap.Logger
}

func (asm *AutoSwitchManager) MonitorAndSwitch(ctx context.Context) {
    ticker := time.NewTicker(time.Second * 30)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            currentProtocol := asm.selector.Current()
            optimalProtocol := asm.selector.SelectBest()
            
            if currentProtocol != optimalProtocol {
                asm.performSwitch(ctx, currentProtocol, optimalProtocol)
            }
        }
    }
}

func (asm *AutoSwitchManager) performSwitch(ctx context.Context, from, to Protocol) error {
    asm.logger.Info("Switching protocol",
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
    
    // 3. Обновление метрик
    asm.metrics.RecordSwitch(from, to)
    
    return nil
}
```

### 5. **Обновление основного обработчика** (День 9-10)

#### 5.1 Интеграция в processConnection
```go
// cmd/relay/main.go
func (s *Server) processConnection(conn net.Conn) error {
    // ... существующий код аутентификации ...
    
    // После успешной аутентификации:
    tenantID := getTenantIDFromToken(token) // Извлечение tenant_id из токена
    
    // Использование унифицированного обработчика
    return s.connectionHandler.HandleConnection(s.ctx, conn, tenantID)
}
```

#### 5.2 Обработка tunnel_info
```go
func (s *Server) handleTunnelInfo(msg map[string]interface{}, writer *bufio.Writer, tenantID string) error {
    // ... существующий код ...
    
    // Определение оптимального протокола для туннеля
    optimalProtocol := s.protocolSelector.SelectBest()
    
    // Создание туннеля через выбранный протокол
    tunnel, err := s.createTunnel(optimalProtocol, msg, tenantID)
    if err != nil {
        return fmt.Errorf("failed to create tunnel: %w", err)
    }
    
    // Отправка ответа с информацией о протоколе
    respTunnel := map[string]interface{}{
        "type":       "tunnel_response",
        "status":     "ok",
        "protocol":   optimalProtocol.String(),
        "tunnel_id":  tunnel.ID,
        "endpoint":   tunnel.Endpoint,
    }
    
    return writeJSON(writer, respTunnel)
}
```

### 6. **Конфигурация мультипротокольности** (День 11-12)

#### 6.1 Обновленная конфигурация
```yaml
# configs/multiprotocol.yaml
multiprotocol:
  enabled: true
  
  # Настройки автоматического выбора
  auto_selection:
    enabled: true
    check_interval: "30s"
    switch_cooldown: "5m"
    
  # Критерии для P2P
  p2p_criteria:
    min_latency_ms: 50
    min_throughput_mbps: 100
    min_peer_count: 3
    max_peer_distance_km: 1000
    
  # Критерии для P2C
  p2c_criteria:
    max_latency_ms: 200
    min_server_capacity: 0.7
    preferred_regions: ["ru", "eu", "us"]
    
  # Fallback настройки
  fallback:
    enabled: true
    order: ["p2p", "p2c", "websocket"]
    timeout: "10s"
    
  # Мониторинг
  monitoring:
    enabled: true
    metrics_interval: "1m"
    health_check_interval: "30s"
```

#### 6.2 Переменные окружения
```bash
# .env
MULTIPROTOCOL_ENABLED=true
P2P_ENABLED=true
P2C_ENABLED=true
AUTO_SELECTION_ENABLED=true
PROTOCOL_SWITCH_COOLDOWN=5m
P2P_MIN_LATENCY_MS=50
P2P_MIN_THROUGHPUT_MBPS=100
P2C_MAX_LATENCY_MS=200
```

### 7. **Тестирование и валидация** (День 13-14)

#### 7.1 Unit тесты
```go
// internal/protocol/selector_test.go
func TestProtocolSelector(t *testing.T) {
    tests := []struct {
        name           string
        latency        time.Duration
        throughput     float64
        peerCount      int
        expectedProtocol Protocol
    }{
        {
            name:            "P2P optimal conditions",
            latency:         30 * time.Millisecond,
            throughput:      150 * 1024 * 1024, // 150 Mbps
            peerCount:       5,
            expectedProtocol: QUIC,
        },
        {
            name:            "P2C fallback",
            latency:         100 * time.Millisecond,
            throughput:      50 * 1024 * 1024, // 50 Mbps
            peerCount:       1,
            expectedProtocol: HTTP2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            selector := NewIntelligentProtocolSelector(ProtocolSelectorConfig{})
            // ... тест логика
        })
    }
}
```

#### 7.2 Integration тесты
```go
// cmd/relay/multiprotocol_integration_test.go
func TestMultiprotocolIntegration(t *testing.T) {
    // Тест полного цикла мультипротокольности
    server := setupTestServer(t)
    defer server.Stop()
    
    // Тест P2P соединения
    t.Run("P2P Connection", func(t *testing.T) {
        conn := createTestConnection(t)
        defer conn.Close()
        
        // Аутентификация
        authenticateConnection(t, conn)
        
        // Создание туннеля через P2P
        tunnel := createTunnel(t, conn, "p2p")
        assert.Equal(t, "quic", tunnel.Protocol)
    })
    
    // Тест P2C соединения
    t.Run("P2C Connection", func(t *testing.T) {
        conn := createTestConnection(t)
        defer conn.Close()
        
        // Аутентификация
        authenticateConnection(t, conn)
        
        // Создание туннеля через P2C
        tunnel := createTunnel(t, conn, "p2c")
        assert.Equal(t, "http2", tunnel.Protocol)
    })
    
    // Тест автоматического переключения
    t.Run("Auto Switch", func(t *testing.T) {
        // Симуляция изменения условий сети
        // Проверка автоматического переключения
    })
}
```

## 📊 Метрики и мониторинг

### Общие метрики мультипротокольности
```go
// internal/metrics/multiprotocol.go
var (
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
)
```

### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "CloudBridge Relay - Multiprotocol Monitoring",
    "panels": [
      {
        "title": "Protocol Distribution",
        "type": "pie",
        "targets": [
          {
            "expr": "connection_distribution",
            "legendFormat": "{{protocol}}"
          }
        ]
      },
      {
        "title": "Protocol Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "protocol_performance_score",
            "legendFormat": "{{protocol}} - {{metric}}"
          }
        ]
      },
      {
        "title": "Protocol Switches",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(protocol_switches_total[5m])",
            "legendFormat": "{{from}} -> {{to}}"
          }
        ]
      }
    ]
  }
}
```

## ✅ Критерии приемки

### Функциональные требования
- [ ] Автоматический выбор между P2P и P2C протоколами
- [ ] Плавное переключение без прерывания соединений
- [ ] Поддержка всех существующих протоколов (HTTP2, QUIC, WebSocket)
- [ ] Интеграция с существующей аутентификацией и мультитенантностью

### Нефункциональные требования
- [ ] Время переключения протокола < 5 секунд
- [ ] Покрытие тестами > 90%
- [ ] Документация полная и актуальная
- [ ] Метрики и мониторинг работают корректно

### Производительность
- [ ] Overhead мультипротокольности < 5%
- [ ] Правильный выбор протокола в 95% случаев
- [ ] Стабильная работа при высокой нагрузке

## 🎯 Результат

После выполнения задания CloudBridge Relay будет:

1. **Полностью мультипротокольным** - автоматический выбор между P2P и P2C
2. **Интеллектуальным** - выбор протокола на основе производительности
3. **Отказоустойчивым** - автоматическое переключение при проблемах
4. **Мониторируемым** - детальная видимость работы всех протоколов
5. **Масштабируемым** - поддержка роста нагрузки через разные протоколы

---

**Готово к production! 🚀**
