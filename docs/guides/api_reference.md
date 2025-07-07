# CloudBridge Relay API Reference

## Обзор

CloudBridge Relay предоставляет REST API для управления туннелями и мониторинга состояния сервера.

## Базовый URL

```
http://localhost:8081/api/v1
```

## Аутентификация

Все API запросы требуют аутентификации через JWT токен в заголовке:

```
Authorization: Bearer <jwt-token>
```

## Endpoints

### 1. Health Check

#### GET /health

Проверяет состояние сервера.

**Ответ:**
```json
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "timestamp": "2024-03-14T12:00:00Z"
}
```

**Коды ответа:**
- `200 OK` - сервер работает нормально
- `503 Service Unavailable` - сервер недоступен

### 2. Metrics

#### GET /metrics

Возвращает метрики в формате Prometheus.

**Ответ:**
```
# HELP relay_active_connections Number of active connections
# TYPE relay_active_connections gauge
relay_active_connections 42

# HELP relay_total_connections Total number of connections
# TYPE relay_total_connections counter
relay_total_connections 1234
```

### 3. Tunnels

#### GET /tunnels

Получает список активных туннелей.

**Параметры запроса:**
- `limit` (int, optional) - максимальное количество туннелей (по умолчанию: 100)
- `offset` (int, optional) - смещение для пагинации (по умолчанию: 0)
- `status` (string, optional) - фильтр по статусу (active, inactive, error)

**Ответ:**
```json
{
  "tunnels": [
    {
      "id": "tunnel-123",
      "client_id": "client-456",
      "local_port": 3389,
      "remote_host": "internal-server",
      "remote_port": 3389,
      "status": "active",
      "created_at": "2024-03-14T12:00:00Z",
      "last_activity": "2024-03-14T12:05:00Z",
      "bytes_sent": 1024,
      "bytes_received": 2048
    }
  ],
  "total": 1,
  "limit": 100,
  "offset": 0
}
```

#### GET /tunnels/{id}

Получает информацию о конкретном туннеле.

**Ответ:**
```json
{
  "id": "tunnel-123",
  "client_id": "client-456",
  "local_port": 3389,
  "remote_host": "internal-server",
  "remote_port": 3389,
  "status": "active",
  "created_at": "2024-03-14T12:00:00Z",
  "last_activity": "2024-03-14T12:05:00Z",
  "bytes_sent": 1024,
  "bytes_received": 2048,
  "options": {
    "compression": "gzip",
    "encryption": "aes256"
  }
}
```

#### DELETE /tunnels/{id}

Закрывает туннель.

**Ответ:**
```json
{
  "success": true,
  "message": "Tunnel closed successfully"
}
```

### 4. Clients

#### GET /clients

Получает список подключенных клиентов.

**Ответ:**
```json
{
  "clients": [
    {
      "id": "client-456",
      "ip_address": "192.168.1.100",
      "connected_at": "2024-03-14T12:00:00Z",
      "last_activity": "2024-03-14T12:05:00Z",
      "active_tunnels": 2,
      "total_tunnels": 5
    }
  ],
  "total": 1
}
```

#### GET /clients/{id}

Получает информацию о конкретном клиенте.

**Ответ:**
```json
{
  "id": "client-456",
  "ip_address": "192.168.1.100",
  "connected_at": "2024-03-14T12:00:00Z",
  "last_activity": "2024-03-14T12:05:00Z",
  "active_tunnels": 2,
  "total_tunnels": 5,
  "tunnels": [
    {
      "id": "tunnel-123",
      "local_port": 3389,
      "remote_host": "internal-server",
      "remote_port": 3389,
      "status": "active"
    }
  ]
}
```

### 5. Statistics

#### GET /stats

Получает общую статистику сервера.

**Ответ:**
```json
{
  "connections": {
    "active": 42,
    "total": 1234,
    "rejected": 56
  },
  "tunnels": {
    "active": 15,
    "total": 89,
    "failed": 3
  },
  "performance": {
    "avg_response_time": 0.15,
    "requests_per_second": 45.2,
    "error_rate": 0.02
  },
  "system": {
    "cpu_usage": 23.5,
    "memory_usage": 67.8,
    "uptime": "2h30m15s"
  }
}
```

#### GET /stats/history

Получает историческую статистику.

**Параметры запроса:**
- `period` (string, required) - период (1h, 6h, 24h, 7d)
- `metric` (string, required) - метрика (connections, tunnels, errors)

**Ответ:**
```json
{
  "period": "1h",
  "metric": "connections",
  "data": [
    {
      "timestamp": "2024-03-14T12:00:00Z",
      "value": 42
    },
    {
      "timestamp": "2024-03-14T12:05:00Z",
      "value": 45
    }
  ]
}
```

### 6. Configuration

#### GET /config

Получает текущую конфигурацию сервера (без секретов).

**Ответ:**
```json
{
  "server": {
    "port": 8080,
    "max_tunnels": 1000,
    "timeout": "30s"
  },
  "security": {
    "tls_enabled": true,
    "rate_limit_enabled": true,
    "rate_limit_requests_per_second": 100
  },
  "metrics": {
    "enabled": true,
    "port": 8081
  }
}
```

#### PUT /config

Обновляет конфигурацию сервера (требует перезапуск).

**Тело запроса:**
```json
{
  "server": {
    "max_tunnels": 1500
  },
  "security": {
    "rate_limit_requests_per_second": 150
  }
}
```

**Ответ:**
```json
{
  "success": true,
  "message": "Configuration updated. Restart required.",
  "restart_required": true
}
```

### 7. Logs

#### GET /logs

Получает логи сервера.

**Параметры запроса:**
- `level` (string, optional) - уровень логирования (debug, info, warning, error)
- `limit` (int, optional) - максимальное количество записей (по умолчанию: 100)
- `since` (string, optional) - фильтр по времени (RFC3339)

**Ответ:**
```json
{
  "logs": [
    {
      "timestamp": "2024-03-14T12:00:00Z",
      "level": "info",
      "message": "Server started",
      "fields": {
        "port": 8080,
        "tls": true
      }
    }
  ],
  "total": 1
}
```

## Коды ошибок

### HTTP Status Codes

- `200 OK` - успешный запрос
- `201 Created` - ресурс создан
- `400 Bad Request` - неверный запрос
- `401 Unauthorized` - не авторизован
- `403 Forbidden` - доступ запрещен
- `404 Not Found` - ресурс не найден
- `429 Too Many Requests` - превышен лимит запросов
- `500 Internal Server Error` - внутренняя ошибка сервера
- `503 Service Unavailable` - сервис недоступен

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid JWT token",
    "details": "Token has expired"
  }
}
```

### Error Codes

- `INVALID_TOKEN` - неверный JWT токен
- `TOKEN_EXPIRED` - токен истек
- `RATE_LIMIT_EXCEEDED` - превышен лимит запросов
- `TUNNEL_NOT_FOUND` - туннель не найден
- `CLIENT_NOT_FOUND` - клиент не найден
- `INVALID_CONFIGURATION` - неверная конфигурация
- `SERVICE_UNAVAILABLE` - сервис недоступен
- `INTERNAL_ERROR` - внутренняя ошибка

## Rate Limiting

API имеет ограничения на количество запросов:

- **Аутентифицированные запросы:** 1000 запросов в минуту
- **Неаутентифицированные запросы:** 10 запросов в минуту

При превышении лимита возвращается статус `429 Too Many Requests`.

## WebSocket API

### WebSocket Endpoint

```
ws://localhost:8081/ws
```

### Подписка на события

```json
{
  "action": "subscribe",
  "events": ["tunnel_created", "tunnel_closed", "client_connected"]
}
```

### События

#### tunnel_created
```json
{
  "event": "tunnel_created",
  "data": {
    "tunnel_id": "tunnel-123",
    "client_id": "client-456",
    "timestamp": "2024-03-14T12:00:00Z"
  }
}
```

#### tunnel_closed
```json
{
  "event": "tunnel_closed",
  "data": {
    "tunnel_id": "tunnel-123",
    "reason": "client_disconnected",
    "timestamp": "2024-03-14T12:05:00Z"
  }
}
```

#### client_connected
```json
{
  "event": "client_connected",
  "data": {
    "client_id": "client-456",
    "ip_address": "192.168.1.100",
    "timestamp": "2024-03-14T12:00:00Z"
  }
}
```

## Примеры использования

### cURL

```bash
# Получить список туннелей
curl -H "Authorization: Bearer <token>" \
     http://localhost:8081/api/v1/tunnels

# Создать туннель
curl -X POST \
     -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"local_port": 3389, "remote_host": "internal-server", "remote_port": 3389}' \
     http://localhost:8081/api/v1/tunnels

# Получить статистику
curl -H "Authorization: Bearer <token>" \
     http://localhost:8081/api/v1/stats
```

### Python

```python
import requests

# Настройка
base_url = "http://localhost:8081/api/v1"
headers = {"Authorization": "Bearer <token>"}

# Получить туннели
response = requests.get(f"{base_url}/tunnels", headers=headers)
tunnels = response.json()

# Получить статистику
response = requests.get(f"{base_url}/stats", headers=headers)
stats = response.json()
```

### JavaScript

```javascript
// Настройка
const baseUrl = "http://localhost:8081/api/v1";
const headers = { "Authorization": "Bearer <token>" };

// Получить туннели
fetch(`${baseUrl}/tunnels`, { headers })
  .then(response => response.json())
  .then(tunnels => console.log(tunnels));

// Получить статистику
fetch(`${baseUrl}/stats`, { headers })
  .then(response => response.json())
  .then(stats => console.log(stats));
```

## Версионирование

API использует семантическое версионирование. Текущая версия: `v1`.

Изменения в API:
- **v1.1** - добавлены WebSocket события
- **v1.0** - первая стабильная версия

## Поддержка

Для получения поддержки:
- Создайте issue в GitHub репозитории
- Обратитесь к документации по отладке
- Проверьте логи сервера 