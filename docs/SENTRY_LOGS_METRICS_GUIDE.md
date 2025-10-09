# Sentry Logs & Metrics Integration Guide

## 🎯 Обзор

Полная интеграция Sentry с логированием и метриками для CloudBridge Relay, включая автоматический сбор логов, метрик производительности и мониторинг ошибок.

## ✅ Реализованные компоненты

### 1. Логирование (internal/sentry/logging.go)
- **SentryZapCore** - интеграция zap logger с Sentry
- **Автоматическая отправка логов** в Sentry по уровням
- **Контекстное логирование** с дополнительными полями
- **Breadcrumbs** для отслеживания последовательности событий

### 2. Метрики (internal/sentry/metrics.go)
- **MetricsCollector** - сбор и отправка метрик
- **Счетчики** (counters) - для подсчета событий
- **Gauges** - для измерения текущих значений
- **Таймеры** - для измерения времени выполнения
- **Автоматическая отправка** метрик каждые 30 секунд

### 3. Middleware (internal/sentry/metrics_middleware.go)
- **MetricsMiddleware** - автоматический сбор метрик HTTP запросов
- **LoggingMiddleware** - структурированное логирование запросов
- **SentryContextMiddleware** - добавление контекста в Sentry
- **HealthCheckMiddleware** - мониторинг здоровья системы
- **ErrorHandlingMiddleware** - обработка и отправка ошибок
- **PerformanceMonitoringMiddleware** - мониторинг производительности

## 🔧 Настройка

### 1. Переменные окружения

```bash
# Основные настройки Sentry
SENTRY_DSN=https://e60ea202d2ee6bc200558dbf6b3cc01c@o4510156466028544.ingest.us.sentry.io/4510156604702720
SENTRY_ENVIRONMENT=production
SENTRY_RELEASE=cloudbridge-relay@1.0.0
SENTRY_DEBUG=true

# Настройки логирования
LOG_LEVEL=info  # debug, info, warn, error, fatal

# Настройки метрик
SENTRY_METRICS_ENABLED=true
SENTRY_METRICS_INTERVAL=30s
```

### 2. Инициализация в main.go

```go
// Инициализация Sentry
sentryConfig := sentry.DefaultConfig()
if err := sentry.Init(sentryConfig); err != nil {
    log.Printf("Failed to initialize Sentry: %v", err)
}
defer sentry.Flush(5 * time.Second)

// Настройка логирования с Sentry
logger, err := sentry.SetupLogging()
if err != nil {
    log.Printf("Failed to setup Sentry logging: %v", err)
} else {
    zap.ReplaceGlobals(logger)
}

// Запуск репортера метрик
sentry.StartGlobalMetricsReporter(30 * time.Second)
```

### 3. Middleware в роутере

```go
// Sentry middleware
router.Use(sentry.NewMiddleware(sentry.DefaultMiddlewareOptions()))
router.Use(sentry.NewCloudBridgeMiddleware())

// Logging and metrics middleware
router.Use(sentry.LoggingMiddleware())
router.Use(sentry.MetricsMiddleware())
router.Use(sentry.SentryContextMiddleware())
router.Use(sentry.HealthCheckMiddleware())
router.Use(sentry.ErrorHandlingMiddleware())
router.Use(sentry.PerformanceMonitoringMiddleware())
```

## 📊 Типы метрик

### 1. HTTP Метрики
- `requests_total` - общее количество запросов
- `requests_{method}_total` - запросы по методам
- `requests_status_{code}_total` - запросы по статус кодам
- `request_duration` - время выполнения запросов
- `request_duration_{method}` - время по методам

### 2. Системные метрики
- `memory_alloc_bytes` - выделенная память
- `memory_total_alloc_bytes` - общая выделенная память
- `memory_sys_bytes` - системная память
- `memory_num_gc` - количество сборок мусора
- `goroutines` - количество горутин

### 3. Метрики ошибок
- `errors_total` - общее количество ошибок
- `errors_{type}_total` - ошибки по типам
- `errors_{component}_total` - ошибки по компонентам

### 4. Метрики производительности
- `performance_{component}` - время выполнения компонентов
- `memory_usage_{component}` - использование памяти компонентами

## 📝 Типы логов

### 1. Уровни логирования
- **DEBUG** - отладочная информация
- **INFO** - информационные сообщения
- **WARN** - предупреждения
- **ERROR** - ошибки
- **FATAL** - критические ошибки

### 2. Контекстное логирование
```go
// Логирование с контекстом
zap.L().Info("Request processed",
    zap.String("method", "GET"),
    zap.String("path", "/api/v1/peers"),
    zap.String("user_agent", "curl/7.68.0"),
    zap.Duration("duration", time.Since(start)),
)
```

### 3. Специальные функции
```go
// Логирование ошибок
sentry.LogError(err, "Database connection failed")

// Логирование с контекстом
sentry.LogWithContext(ctx, zapcore.InfoLevel, "Operation completed", fields...)
```

## 🚀 Использование

### 1. Автоматические метрики
Все HTTP запросы автоматически отслеживаются:
- Время выполнения
- Статус коды
- Размер ответов
- Количество запросов

### 2. Ручные метрики
```go
// Счетчики
sentry.IncrementCounter("api_calls", 1)
sentry.IncrementCounter("errors_total", 1)

// Gauges
sentry.SetGauge("active_connections", 42.5)
sentry.SetGauge("memory_usage", 128.7)

// Таймеры
start := time.Now()
// ... выполнение операции ...
sentry.RecordTimer("operation_duration", time.Since(start))
```

### 3. Метрики запросов
```go
// Автоматические метрики запроса
sentry.RecordRequestMetrics(ctx, "GET", "/api/v1/peers", 200, duration)

// Метрики ошибок
sentry.RecordErrorMetrics(ctx, "database_error", "postgres", 1)

// Метрики производительности
sentry.RecordPerformanceMetrics(ctx, "data_processing", duration, memoryUsage)
```

## 📈 Мониторинг

### 1. Sentry Dashboard
- **Issues** - ошибки и исключения
- **Performance** - метрики производительности
- **Releases** - отслеживание релизов
- **Alerts** - уведомления о проблемах

### 2. Ключевые метрики для мониторинга
- **Response Time** - время ответа API
- **Error Rate** - частота ошибок
- **Throughput** - пропускная способность
- **Memory Usage** - использование памяти
- **CPU Usage** - использование CPU

### 3. Алерты
Настройте алерты в Sentry для:
- Высокий уровень ошибок (>5%)
- Медленные запросы (>1 секунды)
- Высокое использование памяти
- Критические ошибки

## 🧪 Тестирование

### 1. Запуск тестового сервера
```bash
# Загрузка переменных окружения
source scripts/load-env.sh

# Запуск тестового сервера
go run test-sentry-full.go
```

### 2. Тестовые endpoints
```bash
# Тест логирования
curl http://localhost:8080/test-logs

# Тест метрик
curl http://localhost:8080/test-metrics

# Тест ошибок
curl http://localhost:8080/test-errors

# Тест производительности
curl http://localhost:8080/test-performance

# Тест медленных запросов
curl http://localhost:8080/test-slow

# Тест паники
curl http://localhost:8080/test-panic
```

## 🔍 Troubleshooting

### 1. Проблемы с логами
- Проверьте `LOG_LEVEL` в переменных окружения
- Убедитесь, что Sentry DSN корректный
- Проверьте права доступа к файлам логов

### 2. Проблемы с метриками
- Убедитесь, что `SENTRY_METRICS_ENABLED=true`
- Проверьте интервал отправки метрик
- Проверьте доступность Sentry API

### 3. Проблемы с производительностью
- Настройте `SENTRY_SAMPLE_RATE` для уменьшения нагрузки
- Используйте асинхронную отправку метрик
- Мониторьте использование памяти

## 📋 Best Practices

### 1. Логирование
- Используйте структурированное логирование
- Добавляйте контекст к логам
- Не логируйте чувствительные данные
- Используйте соответствующие уровни логирования

### 2. Метрики
- Отслеживайте ключевые бизнес-метрики
- Настройте алерты для критических метрик
- Регулярно анализируйте тренды
- Используйте метрики для оптимизации

### 3. Мониторинг
- Настройте дашборды для ключевых метрик
- Создайте runbooks для типичных проблем
- Регулярно проверяйте алерты
- Документируйте инциденты

## 🎉 Заключение

Интеграция Sentry с логированием и метриками полностью готова и протестирована. Система обеспечивает:

- ✅ **Автоматический сбор логов** всех уровней
- ✅ **Метрики производительности** в реальном времени
- ✅ **Мониторинг ошибок** с контекстом
- ✅ **Алерты и уведомления** о проблемах
- ✅ **Дашборды** для анализа данных

**Готово к использованию в production!** 🚀



