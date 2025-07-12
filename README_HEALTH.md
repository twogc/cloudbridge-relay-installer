# Health Endpoint для CloudBridge Relay

## Быстрый старт

Health endpoint доступен на порту 3456:

```bash
# Проверка health endpoint
curl -s http://localhost:3456/health | jq .

# Ожидаемый ответ:
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

## Тестирование

Запустите автоматизированные тесты:

```bash
./scripts/testing/test-health.sh
```

## Docker

Health endpoint автоматически проверяется Docker healthcheck:

```bash
# Проверка статуса контейнера
docker ps

# Просмотр логов healthcheck
docker inspect cloudbridge-relay | jq '.[0].State.Health'
```

## Kubernetes

Пример конфигурации для Kubernetes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 3456
  initialDelaySeconds: 5
  periodSeconds: 5
```

## Архитектура

Relay-сервер использует мультиплексирование для обработки HTTP и внутреннего протокола на одном порту:

- HTTP-запросы (GET, POST, PUT, DELETE, HEAD, OPTIONS) → HTTP handler
- Остальные запросы → Внутренний протокол CloudBridge

## Подробная документация

См. [docs/guides/health_endpoint.md](docs/guides/health_endpoint.md) для полной документации. 