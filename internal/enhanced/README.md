# 🔧 Расширенные компоненты CloudBridge Relay

Этот пакет содержит расширенные компоненты для CloudBridge Relay, которые добавляют новые возможности мониторинга, диагностики и управления.

## 📁 Структура

```
internal/enhanced/
├── metrics/           # Расширенная система метрик
│   ├── enhanced.go    # Основные метрики
│   ├── health_checker.go # Health check система
│   ├── server.go      # HTTP сервер
│   └── example_test.go # Тесты и примеры
└── README.md          # Документация
```

## 🚀 Возможности

### 1. Расширенная система метрик (`metrics/enhanced.go`)

Добавляет новые типы метрик:

- **Build Information** - информация о сборке
- **Protocol Performance** - производительность протоколов
- **Health Check** - статус проверок здоровья
- **Connection Pool** - управление пулом соединений
- **QUIC** - специфичные метрики для QUIC
- **Ingress** - метрики входящего трафика
- **System** - системные метрики
- **Diagnostic** - диагностические метрики

#### Пример использования:

```go
// Создание enhanced metrics
reg := prometheus.NewRegistry()
em := NewEnhancedMetrics(reg)

// Регистрация build info
em.RegisterBuildInfo("release", "2024-01-01", "2.0.0")

// Протокольные метрики
em.ObserveProtocolLatency("http2", "connect", 100*time.Millisecond)
em.IncProtocolErrors("http2", "timeout")
em.IncProtocolSwitches("http2", "quic", "performance")

// Health check метрики
em.SetHealthCheckStatus("database", "component", true)
em.ObserveHealthCheckDuration("database", "component", 50*time.Millisecond)

// QUIC метрики
em.IncQUICStreams("inbound", "data")
em.SetQUICStreamsActive("inbound", "data", 5)
em.ObserveQUICStreamLatency("inbound", "data", 10*time.Millisecond)

// Системные метрики
em.SetSystemMemoryUsage(1024 * 1024 * 100) // 100MB
em.SetSystemCPUUsage(25.5)
em.SetSystemGoroutines(100)
```

### 2. Health Check система (`metrics/health_checker.go`)

Расширенная система проверки здоровья с поддержкой:

- Множественных проверок
- Метрик производительности
- Kubernetes readiness checks
- Автоматического мониторинга

#### Пример использования:

```go
// Создание health checker
logger, _ := zap.NewDevelopment()
em := NewEnhancedMetrics(reg)
checker := NewEnhancedHealthChecker("2.0.0", logger, em)

// Добавление кастомных проверок
checker.AddCheck(&CustomHealthCheck{
    name: "custom_service",
    logger: logger,
})

// Запуск периодических проверок
ctx := context.Background()
go checker.StartPeriodicChecks(ctx, 30*time.Second)

// Проверка статуса
status := checker.RunChecks(ctx)
if status.Ready {
    fmt.Println("System is ready")
}
```

### 3. Расширенный HTTP сервер (`metrics/server.go`)

HTTP сервер с интегрированными возможностями:

- Prometheus метрики
- Health check endpoints
- Диагностические endpoints
- PProf профилирование
- Graceful shutdown

#### Пример использования:

```go
// Создание конфигурации сервера
config := Config{
    Port:            ":8081",
    Path:            "/metrics",
    EnhancedMetrics: em,
    HealthChecker:   checker,
    Logger:          logger,
    Version:         "2.0.0",
    BuildTime:       "2024-01-01",
    BuildType:       "release",
    TunnelID:        uuid.New(),
    ConnectorID:     uuid.New(),
}

// Создание и запуск сервера
server := NewEnhancedServer(config)
ctx := context.Background()
go server.Start(ctx)
```

## 📊 Доступные endpoints

### Метрики
- `GET /metrics` - Prometheus метрики
- `GET /healthcheck` - Простая проверка здоровья
- `GET /ready` - Kubernetes readiness check
- `GET /version` - Информация о версии
- `GET /buildinfo` - Информация о сборке
- `GET /tunnel` - Информация о туннеле

### Диагностика
- `GET /diagnostic/config` - Конфигурация
- `GET /diagnostic/tunnel-state` - Состояние туннеля
- `GET /diagnostic/system` - Системная информация
- `GET /diagnostic/version` - Версия
- `GET /diagnostic/health` - Здоровье
- `GET /diagnostic/metrics` - Перенаправление на метрики
- `GET /diagnostic/pprof/` - PProf профилирование

### Отладка
- `GET /debug/pprof/` - PProf endpoints
- `GET /debug/pprof/heap` - Профиль памяти
- `GET /debug/pprof/goroutine` - Профиль горутин
- `GET /debug/pprof/profile` - CPU профиль

## 🔧 Интеграция с существующей системой

### 1. Обновление существующих метрик

```go
// В существующем коде
import "github.com/twogc/cloudbridge-relay-installer/internal/enhanced/metrics"

// Создание enhanced metrics
em := metrics.NewEnhancedMetrics(prometheus.DefaultRegisterer)

// Регистрация build info
em.RegisterBuildInfo("release", buildTime, version)

// Использование в существующих компонентах
em.ObserveProtocolLatency("http2", "connect", duration)
em.SetHealthCheckStatus("database", "component", healthy)
```

### 2. Интеграция с HTTP сервером

```go
// Создание enhanced server
config := metrics.Config{
    Port:            ":8081",
    Path:            "/metrics",
    EnhancedMetrics: em,
    HealthChecker:   checker,
    Logger:          logger,
    Version:         version,
    BuildTime:       buildTime,
    BuildType:       buildType,
    TunnelID:        tunnelID,
    ConnectorID:     connectorID,
}

server := metrics.NewEnhancedServer(config)

// Запуск в отдельной горутине
go server.Start(ctx)
```

## 🧪 Тестирование

Запуск тестов:

```bash
cd internal/enhanced/metrics
go test -v
```

Запуск бенчмарков:

```bash
go test -bench=.
```

## 📈 Мониторинг

### Новые метрики Prometheus

#### Build Information
```
build_info{goversion="go1.23.0",type="release",revision="2024-01-01",version="2.0.0"} 1
```

#### Protocol Performance
```
protocol_latency_seconds{protocol="http2",operation="connect"} 0.1
protocol_errors_total{protocol="http2",error_type="timeout"} 5
protocol_switches_total{from_protocol="http2",to_protocol="quic",reason="performance"} 2
```

#### Health Check
```
health_check_status{check="database",component="component"} 1
health_check_duration_seconds{check="database",component="component"} 0.05
```

#### Connection Pool
```
connection_pool_size{pool="tcp",protocol="http2"} 100
connection_pool_idle{pool="tcp",protocol="http2"} 20
connection_pool_active{pool="tcp",protocol="http2"} 80
```

#### QUIC
```
quic_streams_total{direction="inbound",type="data"} 1000
quic_streams_active{direction="inbound",type="data"} 5
quic_stream_latency_seconds{direction="inbound",type="data"} 0.01
```

#### System
```
system_memory_usage_bytes 104857600
system_cpu_usage_percent 25.5
system_goroutines 100
system_heap_alloc_bytes 52428800
system_heap_sys_bytes 209715200
```

## 🔒 Безопасность

- Все endpoints защищены таймаутами
- Graceful shutdown для безопасного завершения
- Фильтрация чувствительных переменных окружения
- Логирование всех операций

## 🚀 Производительность

- Оптимизированные метрики с минимальным overhead
- Асинхронный сбор системных метрик
- Эффективное управление памятью
- Поддержка высоких нагрузок

## 📚 Дополнительные ресурсы

- [Prometheus метрики](https://prometheus.io/docs/concepts/metric_types/)
- [Kubernetes health checks](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
- [PProf профилирование](https://golang.org/pkg/net/http/pprof/)
- [Zap logging](https://github.com/uber-go/zap) 