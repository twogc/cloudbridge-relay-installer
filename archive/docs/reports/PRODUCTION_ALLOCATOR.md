# Продакшен-готовый SCORE аллокатор

## Обзор

`SCORETenantNetworkAllocatorProduction` - это полностью готовый к продакшену аллокатор подсетей с правильной обработкой ошибок, retry логикой и circuit breaker паттерном.

## Ключевые улучшения

### 1. Правильная retry логика

**Проблема**: Circuit breaker скрывал исходные ошибки.

**Решение**: Сначала выполняем операцию, потом сообщаем CB о результате.

```go
func (a *SCORETenantNetworkAllocatorProduction) executeWithRetryImproved(ctx context.Context, operation func() error) error {
    // Быстрая проверка: открыт ли CB?
    if a.circuitBreaker != nil && a.circuitBreaker.IsOpen() {
        return ErrCircuitOpen
    }

    err := operation()
    if err == nil {
        if a.circuitBreaker != nil {
            a.circuitBreaker.OnSuccess()
        }
        return nil
    }

    // Определяем retryable и обрабатываем Retry-After
    // ...
}
```

### 2. Уважение к Retry-After заголовкам

**Проблема**: Игнорирование серверных указаний о времени ожидания.

**Решение**: Автоматическое использование Retry-After для 429/503 ошибок.

```go
if httpErr.StatusCode == http.StatusTooManyRequests || httpErr.StatusCode == http.StatusServiceUnavailable {
    if ra := httpErr.RetryAfter; ra != "" {
        delay = a.retryConfig.CalculateDelayWithRetryAfter(attempt, ra)
    }
}
```

### 3. Типизированные HTTP ошибки

**Проблема**: Потеря информации об HTTP статусах и Retry-After.

**Решение**: Структурированные HTTPError с дополнительной информацией.

```go
type HTTPError struct {
    StatusCode int    `json:"status_code"`
    Message    string `json:"message"`
    RetryAfter string `json:"retry_after,omitempty"`
    Body       string `json:"body,omitempty"`
}
```

### 4. Убрали глобальный mutex

**Проблема**: `ReleaseSubnet` держал глобальный mutex, блокируя другие операции.

**Решение**: Убрали mutex, добавили retry для GET→PUT операций.

```go
func (a *SCORETenantNetworkAllocatorProduction) ReleaseSubnet(ctx context.Context, tenantID string, expected ResourceVersion) (ResourceVersion, error) {
    // Получаем конфигурацию с retry
    var netCfg *TenantNetwork
    err := a.executeWithRetryImproved(ctx, func() error {
        var err error
        netCfg, err = a.getTenantNetwork(ctx, tenantID)
        return err
    })
    
    // Сохраняем с retry
    var newVer ResourceVersion
    err = a.executeWithRetryImproved(ctx, func() error {
        var err error
        newVer, err = a.saveTenantNetworkWithCAS(ctx, netCfg, expected)
        return err
    })
    
    return newVer, nil
}
```

### 5. O(1) поиск подсетей

**Проблема**: Линейный поиск по списку подсетей.

**Решение**: Map для O(1) поиска.

```go
func (a *SCORETenantNetworkAllocatorProduction) GetNextAvailableSubnet(ctx context.Context) (string, error) {
    allocatedSubnets, err := a.getAllocatedSubnets(ctx)
    if err != nil {
        return "", fmt.Errorf("get allocated subnets: %w", err)
    }

    // Преобразуем в map для O(1) поиска
    used := make(map[string]struct{}, len(allocatedSubnets))
    for _, s := range allocatedSubnets {
        used[s] = struct{}{}
    }

    // Ищем доступную подсеть
    subnetCounter := a.subnetCounter
    for subnetCounter <= 65535 {
        cidr := fmt.Sprintf("10.%d.%d.0/24", subnetCounter/256, subnetCounter%256)
        if _, ok := used[cidr]; !ok {
            return cidr, nil
        }
        subnetCounter++
    }
    
    return "", fmt.Errorf("no available subnets")
}
```

## Использование

### Базовое использование

```go
// Создание аллокатора
allocator := network.NewSCORETenantNetworkAllocatorProduction(
    "https://score-api.example.com",
    "your-api-key",
    logger,
)

// Резервирование подсети
network, err := allocator.ReserveSubnet(ctx, tenantID, expectedVersion)
if err != nil {
    // Обработка ошибок
    handleError(err)
}

// Освобождение подсети
releasedVersion, err := allocator.ReleaseSubnet(ctx, tenantID, expectedVersion)
if err != nil {
    // Обработка ошибок
    handleError(err)
}
```

### Обработка ошибок

```go
func handleError(err error) {
    switch {
    case errors.Is(err, network.ErrCircuitOpen):
        // Circuit breaker открыт - сервис недоступен
        log.Println("Service temporarily unavailable")
        
    case errors.Is(err, network.ErrConflict):
        // Конфликт версий - повторить с обновленной версией
        log.Println("Version conflict - retry with updated version")
        
    case errors.Is(err, network.ErrNotFound):
        // Тенант не найден
        log.Println("Tenant not found")
        
    case errors.Is(err, network.ErrStaleVersion):
        // Устаревшая версия - повторить с текущей версией
        log.Println("Stale version - retry with current version")
        
    default:
        // Проверяем HTTP ошибки
        var httpErr *network.HTTPError
        if errors.As(err, &httpErr) {
            log.Printf("HTTP error %d: %s", httpErr.StatusCode, httpErr.Message)
            
            if httpErr.IsRetryable() {
                log.Println("Error is retryable")
            }
            
            if httpErr.GetRetryAfter() != "" {
                log.Printf("Server suggests retry after: %s", httpErr.GetRetryAfter())
            }
        } else {
            log.Printf("Other error: %v", err)
        }
    }
}
```

### Мониторинг

```go
// Получение состояния circuit breaker
cb := allocator.GetCircuitBreaker()
if cb != nil {
    state := cb.GetState()
    stats := cb.GetStats()
    log.Printf("Circuit Breaker state: %v, stats: %+v", state, stats)
}

// Получение retry конфигурации
retryConfig := allocator.GetRetryConfig()
if retryConfig != nil {
    log.Printf("Retry config: max attempts=%d, initial delay=%v", 
        retryConfig.MaxAttempts, retryConfig.InitialDelay)
}
```

## Конфигурация

### Retry конфигурация

```go
// Кастомная retry конфигурация
retryConfig := &network.RetryConfigImproved{
    MaxAttempts:     5,
    InitialDelay:    500 * time.Millisecond,
    MaxDelay:        60 * time.Second,
    MinDelay:        100 * time.Millisecond,
    BackoffFactor:   1.5,
    Jitter:          true,
    JitterRatio:     0.15,
    RetryableErrors: []int{500, 502, 503, 504, 429, 408},
}

// Нормализация обязательна
retryConfig.Normalize()
```

### Circuit Breaker конфигурация

```go
// Создание circuit breaker
cb := network.NewCircuitBreaker(5, 30*time.Second)

// Мониторинг состояния
if cb.IsOpen() {
    log.Println("Circuit breaker is open")
}

// Ручное управление
cb.OnSuccess()  // Сообщить об успехе
cb.OnFailure()  // Сообщить о неудаче
```

## Метрики

### Prometheus метрики

```go
// Пример метрик для мониторинга
var (
    scoreRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "score_requests_total",
            Help: "Total number of SCORE requests",
        },
        []string{"method", "status"},
    )
    
    scoreRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "score_request_duration_seconds",
            Help: "SCORE request duration",
        },
        []string{"method"},
    )
    
    circuitBreakerState = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
        },
        []string{"service"},
    )
)
```

### Логирование

```go
// Структурированное логирование
logger.Info("SCORE operation completed",
    zap.String("operation", "reserve_subnet"),
    zap.String("tenant_id", tenantID),
    zap.Duration("duration", duration),
    zap.Int("attempts", attempts),
    zap.Error(err))
```

## Тестирование

### Unit тесты

```go
func TestReserveSubnet(t *testing.T) {
    allocator := network.NewSCORETenantNetworkAllocatorProduction(
        "https://test-api.example.com",
        "test-key",
        zap.NewNop(),
    )
    
    ctx := context.Background()
    tenantID := "test-tenant"
    expectedVersion := network.ResourceVersion("1")
    
    network, err := allocator.ReserveSubnet(ctx, tenantID, expectedVersion)
    assert.NoError(t, err)
    assert.NotNil(t, network)
    assert.Equal(t, tenantID, network.TenantID)
}
```

### Интеграционные тесты

```go
func TestRetryWithRetryAfter(t *testing.T) {
    // Мокаем сервер, который возвращает 429 с Retry-After
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Retry-After", "5")
        w.WriteHeader(http.StatusTooManyRequests)
        w.Write([]byte("Too Many Requests"))
    }))
    defer server.Close()
    
    allocator := network.NewSCORETenantNetworkAllocatorProduction(
        server.URL,
        "test-key",
        zap.NewNop(),
    )
    
    // Тест должен уважать Retry-After
    start := time.Now()
    _, err := allocator.ReserveSubnet(context.Background(), "test", "1")
    duration := time.Since(start)
    
    assert.Error(t, err)
    assert.True(t, duration >= 5*time.Second, "Should respect Retry-After")
}
```

## Производительность

### Бенчмарки

```go
func BenchmarkReserveSubnet(b *testing.B) {
    allocator := network.NewSCORETenantNetworkAllocatorProduction(
        "https://test-api.example.com",
        "test-key",
        zap.NewNop(),
    )
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := allocator.ReserveSubnet(context.Background(), "test", "1")
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Оптимизации

1. **Connection pooling** - переиспользование HTTP соединений
2. **SingleFlight** - предотвращение дублирования запросов
3. **O(1) поиск** - map для быстрого поиска подсетей
4. **Параллельные операции** - отсутствие глобальных блокировок

## Заключение

Продакшен-готовый аллокатор обеспечивает:

- **Надежность** - правильная retry логика с уважением к серверным указаниям
- **Производительность** - O(1) поиск, отсутствие глобальных блокировок
- **Мониторинг** - полная видимость состояния circuit breaker и retry
- **Обработка ошибок** - типизированные ошибки с дополнительной информацией
- **Тестируемость** - четкое разделение ответственности и мокабельность

Система готова к продакшену и обеспечивает стабильную работу с SCORE API под нагрузкой.
