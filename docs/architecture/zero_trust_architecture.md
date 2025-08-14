# Zero Trust Architecture - CloudBridge Relay

## Обзор

CloudBridge Relay теперь поддерживает современную Zero Trust архитектуру с компонентами для обеспечения надежности, безопасности и мониторинга.

## Компоненты Zero Trust

### 1. Circuit Breaker (Автоматический выключатель)

**Назначение:** Защищает систему от каскадных сбоев, автоматически блокируя запросы при превышении порога ошибок.

**Принцип работы:**
- **CLOSED** - нормальная работа, запросы проходят
- **OPEN** - блокировка запросов после превышения порога ошибок
- **HALF-OPEN** - тестовый режим для проверки восстановления

**Конфигурация:**
```json
{
  "zero_trust": {
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 5,
      "success_threshold": 3,
      "timeout": "30s",
      "window_size": "60s"
    }
  }
}
```

**Использование:**
```go
// Создание circuit breaker
cb := circuitbreaker.NewCircuitBreaker(&circuitbreaker.Config{
    Name:             "service-name",
    FailureThreshold: 5,
    SuccessThreshold: 3,
    Timeout:          30 * time.Second,
    WindowSize:       60 * time.Second,
    Logger:           logger,
})

// Использование с защитой
err := cb.Execute(ctx, func() error {
    return riskyOperation()
})
```

### 2. Health Checker (Проверка состояния)

**Назначение:** Мониторинг состояния различных компонентов системы в реальном времени.

**Типы проверок:**
- **HTTP** - проверка HTTP endpoints
- **Ping** - проверка сетевой доступности
- **Custom** - пользовательские проверки

**Конфигурация:**
```json
{
  "zero_trust": {
    "health_check": {
      "enabled": true,
      "interval": "30s",
      "timeout": "10s",
      "checks": [
        {
          "name": "server_health",
          "type": "http",
          "url": "http://localhost:3456/health",
          "description": "Server health check"
        },
        {
          "name": "database_ping",
          "type": "ping",
          "host": "localhost",
          "description": "Database connectivity check"
        }
      ]
    }
  }
}
```

**Использование:**
```go
// Создание health checker
hc := health.NewHealthChecker(&health.Config{
    Interval: 30 * time.Second,
    Timeout:  10 * time.Second,
    Logger:   logger,
})

// Добавление проверок
hc.AddCheck("http_check", health.HTTPHealthCheck("http_check", "https://api.example.com/health"))
hc.AddCheck("ping_check", health.PingHealthCheck("ping_check", "google.com"))

// Запуск мониторинга
hc.Start()

// Получение статуса
status := hc.GetStatus() // "healthy", "unhealthy", "degraded"
```

### 3. Rate Limiting (Ограничение скорости)

**Назначение:** Защита от DDoS атак и злоупотреблений ресурсами.

**Конфигурация:**
```json
{
  "zero_trust": {
    "rate_limit": {
      "enabled": true,
      "requests_per": 100,
      "window_size": 60
    }
  }
}
```

### 4. Authentication (Аутентификация)

**Назначение:** Централизованная аутентификация с поддержкой JWT и Keycloak.

**Конфигурация:**
```json
{
  "zero_trust": {
    "authentication": {
      "enabled": true,
      "type": "jwt",
      "jwt_secret": "${RELAY_JWT_SECRET}",
      "keycloak_url": "${KEYCLOAK_SERVER_URL}"
    }
  }
}
```

## Архитектурные принципы

### 1. Defense in Depth (Защита в глубину)
- Многоуровневая защита на каждом уровне системы
- Независимые механизмы безопасности
- Дублирование критических проверок

### 2. Least Privilege (Минимальные привилегии)
- Каждый компонент имеет только необходимые права
- Изоляция процессов и ресурсов
- Принцип "отказать по умолчанию"

### 3. Continuous Monitoring (Непрерывный мониторинг)
- Постоянная проверка состояния системы
- Автоматическое обнаружение аномалий
- Логирование всех событий безопасности

### 4. Fail-Safe (Безопасный отказ)
- При сбоях система переходит в безопасное состояние
- Circuit breaker предотвращает каскадные сбои
- Graceful degradation при частичных отказах

## Интеграция с существующей архитектурой

### Protocol Engine
```go
// Интеграция с Protocol Engine
protocolEngine, err := protocol.NewEngine(&protocol.Config{
    Default:     cfg.Protocol.Default,
    Fallback:    cfg.Protocol.Fallback,
    AutoSelect:  cfg.Protocol.AutoSelect,
    HealthCheck: cfg.Protocol.HealthCheck,
    Logger:      logger,
})
```

### QUIC Support
```go
// QUIC конфигурация с Zero Trust
quicConfig := &quic.Config{
    MaxIncomingStreams:           cfg.QUIC.MaxIncomingStreams,
    MaxIncomingUniStreams:        cfg.QUIC.MaxIncomingUniStreams,
    HandshakeTimeout:             cfg.QUIC.HandshakeTimeout,
    IdleTimeout:                  cfg.QUIC.IdleTimeout,
    InsecureSkipVerify:          cfg.QUIC.InsecureSkipVerify,
}
```

### Connection Pool
```go
// Пул соединений с health checking
connectionPool, err := connection.NewPool(&connection.Config{
    MaxConnections: cfg.Server.MaxTunnels,
    Logger:         logger,
})
```

## Мониторинг и метрики

### Circuit Breaker Metrics
- Состояние (CLOSED/OPEN/HALF-OPEN)
- Количество запросов/ошибок
- Время последней ошибки/успеха
- Процент успешных запросов

### Health Check Metrics
- Общий статус системы
- Статус каждого компонента
- Время выполнения проверок
- История ошибок

### Security Metrics
- Количество заблокированных запросов
- Попытки аутентификации
- Rate limiting события
- Аудит безопасности

## Демонстрация

Запуск демо Zero Trust компонентов:
```bash
go run cmd/zerotrust-demo/main.go
```

Демо включает:
1. **Circuit Breaker** - симуляция сбоев и восстановления
2. **Health Checker** - мониторинг HTTP, ping и custom проверок

## Конфигурация

Полный пример конфигурации в `config/config.json`:
```json
{
  "zero_trust": {
    "enabled": true,
    "circuit_breaker": {
      "enabled": true,
      "failure_threshold": 5,
      "success_threshold": 3,
      "timeout": "30s",
      "window_size": "60s"
    },
    "health_check": {
      "enabled": true,
      "interval": "30s",
      "timeout": "10s",
      "checks": [...]
    },
    "rate_limit": {
      "enabled": true,
      "requests_per": 100,
      "window_size": 60
    },
    "authentication": {
      "enabled": true,
      "type": "jwt"
    }
  }
}
```

## Безопасность

### Рекомендации по безопасности:
1. **Всегда включайте Zero Trust** в продакшене
2. **Настройте агрессивные пороги** для circuit breaker
3. **Мониторьте health checks** постоянно
4. **Логируйте все события** безопасности
5. **Используйте сильные секреты** для JWT
6. **Регулярно обновляйте** конфигурацию

### Аудит безопасности:
- Все аутентификации логируются
- Попытки обхода rate limiting отслеживаются
- Состояния circuit breaker мониторятся
- Health check результаты сохраняются

## Заключение

Zero Trust архитектура CloudBridge Relay обеспечивает:
- **Надежность** через circuit breaker и health checks
- **Безопасность** через аутентификацию и rate limiting
- **Мониторинг** через метрики и логирование
- **Масштабируемость** через модульную архитектуру

Это делает CloudBridge Relay готовым для использования в критически важных средах с высокими требованиями к безопасности и надежности. 