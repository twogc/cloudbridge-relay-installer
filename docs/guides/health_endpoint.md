# Health Endpoint для CloudBridge Relay

## Обзор

CloudBridge Relay предоставляет HTTP endpoint `/health` для проверки работоспособности сервиса. Этот endpoint доступен на основном порту relay-сервера (3456) и предназначен для использования системами мониторинга, балансировщиками нагрузки, Kubernetes и другими внешними системами.

## Архитектура

Relay-сервер использует мультиплексирование для обработки как внутреннего протокола CloudBridge, так и HTTP-запросов на одном порту:

1. **Определение типа запроса**: Сервер анализирует первые байты входящего соединения
2. **HTTP-запросы**: Если запрос начинается с HTTP-метода (GET, POST, PUT, DELETE, HEAD, OPTIONS), он обрабатывается как HTTP
3. **Внутренний протокол**: Все остальные запросы обрабатываются как внутренний протокол CloudBridge

## Endpoint

### GET /health

Возвращает информацию о состоянии сервиса.

**URL**: `http://localhost:3456/health`

**Метод**: `GET`

**Заголовки ответа**:
- `Content-Type: application/json`
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

**Пример запроса**:
```bash
curl -s http://localhost:3456/health | jq .
```

**Пример ответа**:
```json
{
  "status": "ok",
  "version": "1.0.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "service": "cloudbridge-relay",
  "uptime": "2h30m15s",
  "components": {
    "tcp_server": "running",
    "metrics": true,
    "health_checker": true,
    "circuit_breaker": true
  }
}
```

### Поля ответа

| Поле | Тип | Описание |
|------|-----|----------|
| `status` | string | Статус сервиса (`ok` - сервис работает) |
| `version` | string | Версия сервиса из конфигурации |
| `timestamp` | string | Время ответа в формате RFC3339 |
| `service` | string | Название сервиса |
| `uptime` | string | Время работы сервиса |
| `components` | object | Статус компонентов системы |

### Компоненты

| Компонент | Тип | Описание |
|-----------|-----|----------|
| `tcp_server` | string | Статус TCP-сервера (`running`) |
| `metrics` | boolean | Включены ли метрики |
| `health_checker` | boolean | Включен ли health checker |
| `circuit_breaker` | boolean | Включен ли circuit breaker |

## Обработка ошибок

### 404 Not Found

Для неизвестных путей возвращается HTTP 404:

```json
{
  "error": "Not Found"
}
```

### CORS поддержка

Endpoint поддерживает CORS для веб-приложений:

- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

## Мониторинг

### Docker Healthcheck

В docker-compose.yml настроен healthcheck:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:3456/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Kubernetes Liveness/Readiness Probes

Пример для Kubernetes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### Prometheus Monitoring

Health endpoint можно использовать для создания алертов в Prometheus:

```yaml
groups:
  - name: relay-health
    rules:
      - alert: RelayHealthCheckFailed
        expr: up{job="cloudbridge-relay"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "CloudBridge Relay health check failed"
          description: "Relay service is not responding to health checks"
```

## Тестирование

### Локальное тестирование

```bash
# Проверка health endpoint
curl -s http://localhost:3456/health | jq .

# Проверка неизвестного пути
curl -s http://localhost:3456/unknown | jq .

# Проверка CORS
curl -H "Origin: http://example.com" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Content-Type" \
     -X OPTIONS \
     http://localhost:3456/health
```

### Автоматизированное тестирование

```bash
#!/bin/bash
# test-health.sh

RELAY_HOST="localhost"
RELAY_PORT="3456"

echo "Testing CloudBridge Relay health endpoint..."

# Test health endpoint
echo "1. Testing /health endpoint..."
response=$(curl -s -w "%{http_code}" http://$RELAY_HOST:$RELAY_PORT/health)
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 200 ]; then
    echo "✅ Health endpoint returned 200"
    echo "Response: $body" | jq .
else
    echo "❌ Health endpoint returned $http_code"
    echo "Response: $body"
    exit 1
fi

# Test unknown path
echo "2. Testing unknown path..."
response=$(curl -s -w "%{http_code}" http://$RELAY_HOST:$RELAY_PORT/unknown)
http_code="${response: -3}"
body="${response%???}"

if [ "$http_code" -eq 404 ]; then
    echo "✅ Unknown path correctly returned 404"
else
    echo "❌ Unknown path returned $http_code instead of 404"
    exit 1
fi

echo "✅ All health endpoint tests passed!"
```

## Безопасность

- Health endpoint не требует аутентификации
- Возвращает только публичную информацию о состоянии сервиса
- Не раскрывает конфиденциальные данные
- Поддерживает CORS для веб-приложений

## Устранение неполадок

### Endpoint недоступен

1. Проверьте, что relay-сервер запущен:
   ```bash
   docker ps | grep cloudbridge-relay
   ```

2. Проверьте логи сервера:
   ```bash
   docker logs cloudbridge-relay
   ```

3. Проверьте, что порт 3456 проброшен:
   ```bash
   netstat -tlnp | grep 3456
   ```

### Неправильный ответ

1. Проверьте конфигурацию в `config/config.json`
2. Убедитесь, что версия в конфигурации корректна
3. Проверьте логи на наличие ошибок

### Проблемы с CORS

1. Убедитесь, что запросы отправляются с правильными заголовками
2. Проверьте, что OPTIONS запросы обрабатываются корректно
3. Проверьте логи на наличие ошибок CORS 