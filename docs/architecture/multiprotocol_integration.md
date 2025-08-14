# 🚀 Мультипротокольная интеграция CloudBridge Relay

## 📋 Обзор

Мультипротокольная интеграция CloudBridge Relay обеспечивает унифицированную систему переключения между P2P и P2C протоколами, автоматически выбирая оптимальный протокол на основе производительности, доступности и требований безопасности.

## 🏗️ Архитектура

### Основные компоненты

1. **UnifiedConnectionHandler** - унифицированный обработчик соединений
2. **P2PManager** - управление P2P соединениями через WireGuard
3. **P2CManager** - управление P2C соединениями через Enhanced QUIC
4. **AutoSwitchManager** - автоматическое переключение протоколов
5. **IntelligentProtocolSelector** - интеллектуальный выбор протокола

### Диаграмма архитектуры

```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │ UnifiedConnection│    │ AutoSwitchManager│                │
│  │    Handler      │    │                 │                │
│  └─────────────────┘    └─────────────────┘                │
│           │                       │                        │
│           ▼                       ▼                        │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │IntelligentProtocol│    │ Protocol Engine │                │
│  │   Selector      │    │                 │                │
│  └─────────────────┘    └─────────────────┘                │
│           │                                               │
│           ▼                                               │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   P2P Manager   │    │   P2C Manager   │                │
│  │                 │    │                 │                │
│  │ • WireGuard     │    │ • Enhanced QUIC │                │
│  │ • Mesh Topology │    │ • Load Balancer │                │
│  │ • Peer Discovery│    │ • Rate Limiting │                │
│  └─────────────────┘    └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

## 🔧 Конфигурация

### Основные настройки

```yaml
protocol:
  auto_select: true
  preferred: "quic"
  fallback: "http2"
  switch_threshold: 0.8
  health_check_interval: "30s"

wireguard:
  enabled: true
  mesh:
    enabled: true
    auto_discovery: true
    max_connections: 10
    routing_strategy: "hybrid"

enhanced_quic:
  enabled: true
  ddos_protection: true
  geographic_routing: true
  keycloak_integration: true
```

### Переменные окружения

```bash
# Мультипротокольные настройки
PROTOCOL_AUTO_SELECT=true
PROTOCOL_PREFERRED=quic
PROTOCOL_FALLBACK=http2
PROTOCOL_SWITCH_THRESHOLD=0.8

# WireGuard P2P настройки
WIREGUARD_ENABLED=true
WIREGUARD_MESH_ENABLED=true
WIREGUARD_AUTO_DISCOVERY=true

# Enhanced QUIC P2C настройки
ENHANCED_QUIC_ENABLED=true
ENHANCED_QUIC_DDOS_PROTECTION=true
ENHANCED_QUIC_GEOGRAPHIC_ROUTING=true
```

## 🚀 Использование

### Инициализация

```go
// Создание мультипротокольных компонентов
protocolSelector := protocol.NewIntelligentProtocolSelector(protocol.ProtocolSelectorConfig{
    Strategy:             protocol.PerformanceBased,
    PreferredOrder:       []protocol.Protocol{protocol.QUIC, protocol.HTTP2, protocol.WebSocket},
    SwitchThreshold:      cfg.Protocol.SwitchThreshold,
    HealthCheckInterval:  cfg.Protocol.HealthCheckInterval,
    Logger:               logger,
})

p2pManager := p2p.NewP2PManager(
    wireGuardMesh,
    wireGuardPeerDiscovery,
    hybridRouter,
    siteToSiteManager,
    logger,
)

p2cManager := p2c.NewP2CManager(
    enhancedQUICManager,
    tenantManager,
    loadBalancer,
    ipRateLimit,
    logger,
)

connectionHandler := protocol.NewUnifiedConnectionHandler(
    protocolSelector,
    p2pManager,
    p2cManager,
    logger,
)

autoSwitchManager := protocol.NewAutoSwitchManager(
    protocolSelector,
    p2pManager,
    p2cManager,
    logger,
)
```

### Обработка соединений

```go
// Унифицированная обработка соединения
err := connectionHandler.HandleConnection(ctx, conn, tenantID)
if err != nil {
    log.Error("Failed to handle connection", zap.Error(err))
}
```

## 📊 Метрики

### P2P метрики

- `p2p_connections_total` - общее количество P2P соединений
- `p2p_latency_seconds` - задержка P2P соединений
- `p2p_throughput_bytes` - пропускная способность P2P
- `p2p_active_peers` - количество активных пиров

### P2C метрики

- `p2c_connections_total` - общее количество P2C соединений
- `p2c_latency_seconds` - задержка P2C соединений
- `p2c_throughput_bytes` - пропускная способность P2C
- `p2c_server_load` - нагрузка на серверы

### Метрики переключения

- `protocol_switches_total` - количество переключений протоколов
- `protocol_switch_duration_seconds` - время переключения протоколов

### Пример запроса метрик

```bash
# Получение метрик P2P
curl http://localhost:9090/metrics | grep p2p

# Получение метрик P2C
curl http://localhost:9090/metrics | grep p2c

# Получение метрик переключения
curl http://localhost:9090/metrics | grep protocol_switch
```

## 🔄 Алгоритм выбора протокола

### 1. Проверка доступности P2P

```go
func (pm *P2PManager) IsAvailable(ctx context.Context, clientIP string, tenantID string) bool {
    // Проверка количества активных пиров
    peers := pm.peerDiscovery.GetActivePeers()
    if len(peers) < 3 {
        return false
    }
    
    // Проверка качества соединения
    latency := pm.measureLatency(ctx, clientIP)
    if latency > 50*time.Millisecond {
        return false
    }
    
    return true
}
```

### 2. Выбор оптимального протокола

```go
func (h *UnifiedConnectionHandler) HandleConnection(ctx context.Context, conn net.Conn, tenantID string) error {
    // 1. Определение оптимального протокола
    protocol := h.protocolSelector.SelectBest()
    
    // 2. Обработка через выбранный протокол
    switch protocol {
    case QUIC:
        return h.handleQUICConnection(ctx, conn, tenantID)
    case HTTP2:
        return h.handleHTTP2Connection(ctx, conn, tenantID)
    case WebSocket:
        return h.handleWebSocketConnection(ctx, conn, tenantID)
    default:
        return h.handleFallbackConnection(ctx, conn, tenantID)
    }
}
```

### 3. Автоматическое переключение

```go
func (asm *AutoSwitchManager) checkAndSwitch(ctx context.Context) {
    currentProtocol := asm.selector.Current()
    optimalProtocol := asm.selector.SelectBest()
    
    if currentProtocol != optimalProtocol {
        asm.performSwitch(ctx, currentProtocol, optimalProtocol)
    }
}
```

## 🧪 Тестирование

### Unit тесты

```bash
# Запуск unit тестов
go test ./internal/protocol/ -v

# Запуск тестов с покрытием
go test ./internal/protocol/ -cover -coverprofile=coverage.out
```

### Интеграционные тесты

```bash
# Запуск интеграционных тестов
go test ./cmd/relay/ -run TestMultiprotocolIntegration -v

# Запуск всех тестов мультипротокольности
go test ./cmd/relay/ -run TestMultiprotocol -v
```

### Пример теста

```go
func TestMultiprotocolIntegration(t *testing.T) {
    cfg := createTestConfig()
    server, err := NewServer(cfg)
    require.NoError(t, err)
    
    // Проверка инициализации компонентов
    assert.NotNil(t, server.protocolSelector)
    assert.NotNil(t, server.connectionHandler)
    assert.NotNil(t, server.autoSwitchManager)
    
    // Проверка выбора протокола
    protocol := server.protocolSelector.SelectBest()
    assert.NotEqual(t, Protocol(999), protocol)
}
```

## 🔒 Безопасность

### P2P безопасность

- **WireGuard шифрование** - все P2P соединения шифруются
- **Peer аутентификация** - проверка подлинности пиров
- **Mesh изоляция** - изоляция между различными mesh сетями

### P2C безопасность

- **Enhanced QUIC** - улучшенная безопасность QUIC
- **DDoS защита** - защита от DDoS атак
- **Rate limiting** - ограничение скорости соединений
- **Tenant изоляция** - изоляция между тенантами

### Общая безопасность

- **Автоматическое переключение** - безопасное переключение протоколов
- **Graceful shutdown** - корректное завершение соединений
- **Audit логирование** - логирование всех операций

## 📈 Производительность

### Оптимизации

1. **Интеллектуальный выбор протокола** - выбор на основе метрик
2. **Автоматическое переключение** - адаптация к изменениям сети
3. **Graceful degradation** - плавное снижение качества при проблемах
4. **Connection pooling** - переиспользование соединений

### Бенчмарки

```bash
# Тест производительности P2P
go test ./internal/p2p/ -bench=BenchmarkP2PConnection -benchmem

# Тест производительности P2C
go test ./internal/p2c/ -bench=BenchmarkP2CConnection -benchmem

# Тест переключения протоколов
go test ./internal/protocol/ -bench=BenchmarkProtocolSwitch -benchmem
```

## 🚨 Мониторинг и алерты

### Ключевые метрики для мониторинга

1. **Процент успешных соединений** - должно быть > 95%
2. **Время переключения протокола** - должно быть < 5 секунд
3. **Задержка соединений** - должна быть < 100ms
4. **Количество активных пиров** - должно быть > 3

### Пример алерта Prometheus

```yaml
groups:
  - name: multiprotocol_alerts
    rules:
      - alert: HighProtocolSwitchRate
        expr: rate(protocol_switches_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High protocol switch rate detected"
          
      - alert: P2PUnavailable
        expr: p2p_active_peers < 3
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "P2P network unavailable"
```

## 🔧 Troubleshooting

### Частые проблемы

1. **P2P недоступен**
   - Проверить количество активных пиров
   - Проверить качество сетевого соединения
   - Проверить конфигурацию WireGuard

2. **Высокая задержка**
   - Проверить метрики задержки
   - Проверить качество интернет-соединения
   - Рассмотреть переключение на P2C

3. **Частые переключения протоколов**
   - Проверить настройки switch_threshold
   - Проверить стабильность сети
   - Увеличить cooldown период

### Логи для диагностики

```bash
# Логи выбора протокола
grep "Selected protocol" /var/log/relay/relay.log

# Логи переключения протоколов
grep "Protocol switch" /var/log/relay/relay.log

# Логи P2P соединений
grep "P2P connection" /var/log/relay/relay.log

# Логи P2C соединений
grep "P2C connection" /var/log/relay/relay.log
```

## 📚 Дополнительные ресурсы

- [Техническое задание](../technical_requirements/MULTIPROTOCOL_INTEGRATION_TASK.md)
- [API документация](../api/multiprotocol_api.md)
- [Примеры конфигурации](../examples/multiprotocol_configs/)
- [Troubleshooting guide](../troubleshooting/multiprotocol_issues.md)

---

**Версия**: 1.0.0  
**Дата**: 2024  
**Автор**: CloudBridge Team
