# Архитектура и использование метрик Prometheus

## Общие принципы

Вся работа с метриками Prometheus инкапсулирована в структуре `Metrics` (internal/metrics/metrics.go). Это обеспечивает:
- отсутствие конфликтов имен метрик
- удобство тестирования (можно использовать отдельный реестр)
- чистоту и расширяемость кода

## Создание и регистрация метрик

Для создания метрик используйте конструктор:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/twogc/cloudbridge-relay-installer/internal/metrics"
)

// Для production-кода:
m := metrics.NewMetrics(prometheus.DefaultRegisterer)

// Для тестов:
reg := prometheus.NewRegistry()
m := metrics.NewMetrics(reg)
```

Все метрики регистрируются только через переданный реестр (`prometheus.Registerer`).

## Использование метрик

Вместо глобальных переменных используйте методы структуры:

```go
m.IncConnections()
m.DecConnections()
m.IncErrors("code")
m.IncHeartbeats()
m.IncHeartbeatErrors()
m.SetActiveTunnels(count)
m.SetCacheSize(count)
m.IncCacheHits()
m.IncCacheMisses()
m.ObserveConnectionDuration(duration)
m.ObserveHeartbeatLatency(duration)
m.ObserveTokenValidationDuration("method", duration)
m.IncTokenValidations("method", "result")
m.ObserveMessageLatency("type", duration)
m.IncMessageErrors("type", "error")
```

## Новые метрики для Keycloak

### Метрики аутентификации
```go
// Валидация токенов Keycloak
m.IncKeycloakTokenValidations("success", "jwt")
m.IncKeycloakTokenValidations("failure", "jwt")

// Длительность валидации токенов
m.ObserveKeycloakTokenValidationDuration("jwt", duration)
m.ObserveKeycloakTokenValidationDuration("jwks", duration)

// Длительность извлечения информации пользователя
m.ObserveKeycloakUserInfoDuration(duration)
```

### Метрики JWKS кэша
```go
// Производительность кэша JWKS
m.IncKeycloakJWKSCacheHits()
m.IncKeycloakJWKSCacheMisses()
m.SetKeycloakJWKSCacheSize(count)

// Длительность получения JWKS
m.ObserveKeycloakJWKSFetchDuration(duration)
```

### Метрики проверки разрешений
```go
// Проверка разрешений
m.IncKeycloakPermissionChecks("tunnel", "create", "granted")
m.IncKeycloakPermissionChecks("user", "read", "denied")
```

## Бизнес-метрики

### Метрики пользователей
```go
// Активные пользователи
m.SetActiveUsers(count)

// Доход в час
m.IncRevenuePerHour(amount)

// Удовлетворенность клиентов
m.SetCustomerSatisfaction(score)
```

## Технические метрики

### Системные метрики
```go
// Использование памяти
m.SetMemoryUsage(bytes)

// Длительность пауз GC
m.ObserveGCPauseDuration(duration)

// Сетевая задержка
m.ObserveNetworkLatency(duration)

// Операции ввода-вывода
m.IncDiskIO()
```

## Доступные метрики

### Основные метрики
- `relay_connections_total` - общее количество соединений
- `relay_rejected_connections_total` - отклоненные соединения
- `relay_errors_total{code="error_code"}` - ошибки по кодам
- `relay_heartbeats_total` - количество heartbeat
- `relay_heartbeat_errors_total` - ошибки heartbeat
- `relay_active_connections` - активные соединения
- `relay_active_tunnels` - активные туннели
- `relay_connection_duration_seconds` - время соединения
- `relay_heartbeat_latency_seconds` - задержка heartbeat

### Keycloak метрики
- `relay_keycloak_token_validations_total{result="success|failure",provider="jwt|django"}` - валидация токенов
- `relay_keycloak_jwks_cache_hits_total` - попадания в кэш JWKS
- `relay_keycloak_jwks_cache_misses_total` - промахи кэша JWKS
- `relay_keycloak_jwks_cache_size` - размер кэша JWKS
- `relay_keycloak_jwks_fetch_duration_seconds` - время получения JWKS
- `relay_keycloak_token_validation_duration_seconds{method="jwt|jwks"}` - время валидации токенов
- `relay_keycloak_user_info_duration_seconds` - время извлечения информации пользователя
- `relay_keycloak_permission_checks_total{resource="tunnel|user",action="create|read|update|delete",result="granted|denied"}` - проверки разрешений

### Бизнес-метрики
- `relay_business_active_users` - активные пользователи
- `relay_business_revenue_per_hour_total` - доход в час
- `relay_business_customer_satisfaction` - удовлетворенность клиентов (0-100)

### Технические метрики
- `relay_technical_memory_usage_bytes` - использование памяти
- `relay_technical_gc_pause_duration_seconds` - длительность пауз GC
- `relay_technical_network_latency_seconds` - сетевая задержка
- `relay_technical_disk_io_total` - операции ввода-вывода

## Мониторинг и алерты

### Grafana дашборд

Обновленный дашборд включает:
- Панели для Keycloak метрик
- Бизнес-метрики
- Технические метрики
- Системное здоровье

### Алерты

Новые алерты для Keycloak:
- Высокий процент неудачных валидаций токенов
- Медленные операции JWKS
- Низкая производительность кэша
- Отказы в разрешениях

### Примеры запросов Prometheus

```promql
# Процент успешных валидаций токенов
rate(relay_keycloak_token_validations_total{result="success"}[5m]) / 
rate(relay_keycloak_token_validations_total[5m]) * 100

# Среднее время валидации токенов
histogram_quantile(0.95, rate(relay_keycloak_token_validation_duration_seconds_bucket[5m]))

# Процент попаданий в кэш JWKS
rate(relay_keycloak_jwks_cache_hits_total[5m]) / 
(rate(relay_keycloak_jwks_cache_hits_total[5m]) + rate(relay_keycloak_jwks_cache_misses_total[5m])) * 100

# Активные пользователи
relay_business_active_users

# Доход в час
rate(relay_business_revenue_per_hour_total[1h])

# Использование памяти
relay_technical_memory_usage_bytes / 1024 / 1024  # в MB
```

## Тестирование

В тестах всегда используйте отдельный реестр:

```go
func TestKeycloakMetrics(t *testing.T) {
    reg := prometheus.NewRegistry()
    m := metrics.NewMetrics(reg)

    // Тестируем метрики Keycloak
    m.IncKeycloakTokenValidations("success", "jwt")
    m.IncKeycloakJWKSCacheHits()
    m.SetKeycloakJWKSCacheSize(5)
    m.ObserveKeycloakJWKSFetchDuration(100 * time.Millisecond)
    m.IncKeycloakPermissionChecks("tunnel", "create", "granted")

    // Тестируем бизнес-метрики
    m.SetActiveUsers(100)
    m.IncRevenuePerHour(50.0)
    m.SetCustomerSatisfaction(85.5)

    // Тестируем технические метрики
    m.SetMemoryUsage(1024 * 1024 * 100) // 100MB
    m.ObserveGCPauseDuration(10 * time.Millisecond)
    m.ObserveNetworkLatency(5 * time.Millisecond)
    m.IncDiskIO()
}
```

## Интеграция с логированием

Метрики интегрированы с системой логирования для создания структурированных логов:

```go
// Логирование с метриками
logger.LogPerformance("token_validation", duration, map[string]interface{}{
    "method": "jwt",
    "success": true,
    "cache_hit": false,
})

// Логирование бизнес-событий
logger.LogBusinessEvent("user_login", userID, map[string]interface{}{
    "active_users": activeUsers,
    "revenue_per_hour": revenue,
})
```

## Мониторинг в production

### Настройка Prometheus

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'relay'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
    scrape_interval: 10s
```

### Настройка алертов

```yaml
# alerts.yml
groups:
  - name: relay_alerts
    rules:
      - alert: HighTokenValidationFailureRate
        expr: rate(relay_keycloak_token_validations_total{result="failure"}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High token validation failure rate"
          
      - alert: LowJWKSCacheHitRate
        expr: rate(relay_keycloak_jwks_cache_hits_total[5m]) / 
              (rate(relay_keycloak_jwks_cache_hits_total[5m]) + rate(relay_keycloak_jwks_cache_misses_total[5m])) < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Low JWKS cache hit rate"
          
      - alert: HighMemoryUsage
        expr: relay_technical_memory_usage_bytes / 1024 / 1024 / 1024 > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage"
```

### Дашборд Grafana

Основные панели дашборда:
1. **Обзор системы** - активные соединения, туннели, ошибки
2. **Keycloak метрики** - валидация токенов, производительность кэша
3. **Бизнес-метрики** - активные пользователи, доход, удовлетворенность
4. **Технические метрики** - память, GC, сеть, диск
5. **Алерты** - текущие алерты и их статус

## Устранение неполадок

### Проверка метрик

```bash
# Проверка доступности метрик
curl -s http://localhost:8081/metrics | head -20

# Поиск конкретных метрик
curl -s http://localhost:8081/metrics | grep relay_keycloak

# Проверка через Prometheus
curl -s "http://localhost:9090/api/v1/query?query=relay_active_connections"
```

### Отладка метрик

```bash
# Проверка логов с метриками
tail -f /var/log/relay/relay.log | jq 'select(.msg | contains("metrics"))'

# Проверка производительности
curl -s http://localhost:8081/metrics | grep -E "(go_goroutines|go_memstats)"
```

### Частые проблемы

1. **Метрики не обновляются** - проверьте, что метрики вызываются в коде
2. **Неправильные значения** - проверьте логику подсчета метрик
3. **Высокая нагрузка** - увеличьте интервал сбора метрик в Prometheus
4. **Отсутствие метрик** - проверьте регистрацию метрик в NewMetrics()

## Производительность

- Все метрики используют атомарные операции
- Минимальное влияние на производительность
- Кэширование для часто используемых значений
- Асинхронное обновление для тяжелых операций

## Расширение метрик

Для добавления новых метрик:

1. Добавьте поле в структуру `Metrics`
2. Инициализируйте метрику в `NewMetrics()`
3. Зарегистрируйте в реестре
4. Добавьте методы для работы с метрикой
5. Напишите тесты
6. Обновите документацию 