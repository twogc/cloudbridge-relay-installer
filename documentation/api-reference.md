# CloudBridge Relay Server - API Reference

## Обзор

CloudBridge Relay Server предоставляет REST API для управления туннелями, мониторинга и конфигурации. API доступен на порту 9090 и поддерживает JSON формат данных.

## Базовые URL

### Основной API (TCP Server)
```
http://localhost:9090/api/v1
```

### HTTP API Server (Workflow Management)
```
http://localhost:9091/api/v1
```

## Аутентификация

API использует JWT токены для аутентификации. Токен должен быть передан в заголовке `Authorization`:

```
Authorization: Bearer <jwt-token>
```

## Формат ответов

Все ответы возвращаются в JSON формате:

### Успешный ответ
```json
{
  "status": "success",
  "data": { ... },
  "timestamp": "2025-09-04T11:30:00Z"
}
```

### Ошибка
```json
{
  "status": "error",
  "error": {
    "code": "INVALID_TOKEN",
    "message": "JWT token is invalid or expired",
    "details": "Token expired at 2025-09-04T10:00:00Z"
  },
  "timestamp": "2025-09-04T11:30:00Z"
}
```

## Endpoints

### 1. Health Check

#### GET /health

Проверка состояния сервера.

**Параметры:** Нет

**Ответ:**
```json
{
  "status": "ok",
  "service": "cloudbridge-relay",
  "version": "v1",
  "uptime": "1h23m45.678s",
  "timestamp": "2025-09-04T11:30:00Z",
  "components": {
    "tcp_server": "running",
    "wireguard_server": "running",
    "auth_adapter": "running",
    "tunnel_manager": "running",
    "metrics": "running"
  },
  "tenant_stats": {
    "total_tenants": 5,
    "active_tenants": 3,
    "enhanced_mode": true
  },
  "enhanced_quic_stats": {
    "enabled": true,
    "ddos_protection_enabled": true,
    "geographic_routing_enabled": true,
    "keycloak_integration_enabled": false,
    "tenants": {}
  }
}
```

### 2. Аутентификация

#### POST /auth/validate

Валидация JWT токена.

**Тело запроса:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tenant_id": "relay-test-tenant"
}
```

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "valid": true,
    "client_id": "relay-client-001",
    "tenant_id": "relay-test-tenant",
    "expires_at": "2025-09-04T12:00:00Z",
    "permissions": ["tunnel:create", "tunnel:read"]
  }
}
```

### 3. Управление туннелями

#### POST /tunnels

Создание нового туннеля.

**Тело запроса:**
```json
{
  "mode": "proxy",
  "remote_host": "192.168.1.100",
  "remote_port": 3389,
  "protocol": "tcp",
  "tenant_id": "relay-test-tenant"
}
```

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tunnel_id": "tunnel-12345",
    "status": "created",
    "local_port": 54321,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "protocol": "tcp",
    "tenant_id": "relay-test-tenant",
    "created_at": "2025-09-04T11:30:00Z",
    "expires_at": "2025-09-04T12:30:00Z"
  }
}
```

#### GET /tunnels

Получение списка активных туннелей.

**Параметры:**
- `tenant_id` (опционально) - фильтр по tenant
- `status` (опционально) - фильтр по статусу
- `limit` (опционально) - количество записей (по умолчанию 100)
- `offset` (опционально) - смещение (по умолчанию 0)

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tunnels": [
      {
        "tunnel_id": "tunnel-12345",
        "status": "active",
        "local_port": 54321,
        "remote_host": "192.168.1.100",
        "remote_port": 3389,
        "protocol": "tcp",
        "tenant_id": "relay-test-tenant",
        "created_at": "2025-09-04T11:30:00Z",
        "last_heartbeat": "2025-09-04T11:35:00Z",
        "bytes_sent": 1048576,
        "bytes_received": 2097152
      }
    ],
    "total": 1,
    "limit": 100,
    "offset": 0
  }
}
```

#### GET /tunnels/{tunnel_id}

Получение информации о конкретном туннеле.

**Параметры:**
- `tunnel_id` - ID туннеля

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tunnel_id": "tunnel-12345",
    "status": "active",
    "local_port": 54321,
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "protocol": "tcp",
    "tenant_id": "relay-test-tenant",
    "created_at": "2025-09-04T11:30:00Z",
    "last_heartbeat": "2025-09-04T11:35:00Z",
    "bytes_sent": 1048576,
    "bytes_received": 2097152,
    "connection_count": 1,
    "error_count": 0
  }
}
```

#### DELETE /tunnels/{tunnel_id}

Закрытие туннеля.

**Параметры:**
- `tunnel_id` - ID туннеля

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tunnel_id": "tunnel-12345",
    "status": "closed",
    "closed_at": "2025-09-04T11:40:00Z"
  }
}
```

#### POST /tunnels/{tunnel_id}/heartbeat

Отправка heartbeat для туннеля.

**Параметры:**
- `tunnel_id` - ID туннеля

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tunnel_id": "tunnel-12345",
    "heartbeat_received": true,
    "timestamp": "2025-09-04T11:35:00Z"
  }
}
```

### 4. Управление tenant'ами

#### GET /tenants

Получение списка tenant'ов.

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tenants": [
      {
        "tenant_id": "relay-test-tenant",
        "bandwidth_limit": 104857600,
        "connection_limit": 20,
        "rate_limit": 500,
        "burst": 100,
        "active_connections": 5,
        "active_tunnels": 3,
        "bytes_sent": 52428800,
        "bytes_received": 104857600
      }
    ],
    "total": 1
  }
}
```

#### GET /tenants/{tenant_id}

Получение информации о tenant'е.

**Параметры:**
- `tenant_id` - ID tenant'а

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tenant_id": "relay-test-tenant",
    "bandwidth_limit": 104857600,
    "connection_limit": 20,
    "rate_limit": 500,
    "burst": 100,
    "active_connections": 5,
    "active_tunnels": 3,
    "bytes_sent": 52428800,
    "bytes_received": 104857600,
    "created_at": "2025-09-04T10:00:00Z",
    "last_activity": "2025-09-04T11:35:00Z"
  }
}
```

#### PUT /tenants/{tenant_id}

Обновление настроек tenant'а.

**Тело запроса:**
```json
{
  "bandwidth_limit": 209715200,
  "connection_limit": 50,
  "rate_limit": 1000,
  "burst": 200
}
```

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "tenant_id": "relay-test-tenant",
    "bandwidth_limit": 209715200,
    "connection_limit": 50,
    "rate_limit": 1000,
    "burst": 200,
    "updated_at": "2025-09-04T11:40:00Z"
  }
}
```

### 5. Метрики

#### GET /metrics

Получение метрик в формате Prometheus.

**Параметры:** Нет

**Ответ:**
```
# HELP cloudbridge_connections_total Total number of connections
# TYPE cloudbridge_connections_total counter
cloudbridge_connections_total 1234

# HELP cloudbridge_tunnels_created_total Total number of tunnels created
# TYPE cloudbridge_tunnels_created_total counter
cloudbridge_tunnels_created_total 567

# HELP cloudbridge_active_connections Current number of active connections
# TYPE cloudbridge_active_connections gauge
cloudbridge_active_connections 89

# HELP cloudbridge_request_duration_seconds Request duration in seconds
# TYPE cloudbridge_request_duration_seconds histogram
cloudbridge_request_duration_seconds_bucket{le="0.1"} 100
cloudbridge_request_duration_seconds_bucket{le="0.5"} 200
cloudbridge_request_duration_seconds_bucket{le="1.0"} 250
cloudbridge_request_duration_seconds_bucket{le="+Inf"} 300
cloudbridge_request_duration_seconds_sum 150.5
cloudbridge_request_duration_seconds_count 300
```

#### GET /stats

Получение статистики в JSON формате.

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "connections": {
      "total": 1234,
      "active": 89,
      "failed": 12
    },
    "tunnels": {
      "total_created": 567,
      "active": 45,
      "closed": 522
    },
    "tenants": {
      "total": 5,
      "active": 3
    },
    "performance": {
      "avg_response_time_ms": 25.5,
      "requests_per_second": 150.2,
      "bytes_per_second": 1048576
    },
    "errors": {
      "total": 12,
      "auth_failures": 5,
      "tunnel_failures": 3,
      "network_errors": 4
    }
  }
}
```

### 6. WireGuard P2P

#### GET /wireguard/peers

Получение списка WireGuard пиров.

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "peers": [
      {
        "peer_id": "peer-001",
        "public_key": "base64-encoded-public-key",
        "endpoint": "192.168.1.100:51820",
        "last_handshake": "2025-09-04T11:30:00Z",
        "bytes_sent": 1048576,
        "bytes_received": 2097152,
        "status": "connected"
      }
    ],
    "total": 1
  }
}
```

#### POST /wireguard/peers

Добавление нового WireGuard пира.

**Тело запроса:**
```json
{
  "public_key": "base64-encoded-public-key",
  "endpoint": "192.168.1.100:51820",
  "allowed_ips": ["10.0.0.0/24"]
}
```

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "peer_id": "peer-002",
    "public_key": "base64-encoded-public-key",
    "endpoint": "192.168.1.100:51820",
    "allowed_ips": ["10.0.0.0/24"],
    "created_at": "2025-09-04T11:40:00Z"
  }
}
```

#### DELETE /wireguard/peers/{peer_id}

Удаление WireGuard пира.

**Параметры:**
- `peer_id` - ID пира

**Ответ:**
```json
{
  "status": "success",
  "data": {
    "peer_id": "peer-002",
    "status": "removed",
    "removed_at": "2025-09-04T11:45:00Z"
  }
}
```

## Коды ошибок

### HTTP статус коды
- `200 OK` - Успешный запрос
- `201 Created` - Ресурс создан
- `400 Bad Request` - Неверный запрос
- `401 Unauthorized` - Не авторизован
- `403 Forbidden` - Доступ запрещен
- `404 Not Found` - Ресурс не найден
- `409 Conflict` - Конфликт ресурсов
- `429 Too Many Requests` - Превышен лимит запросов
- `500 Internal Server Error` - Внутренняя ошибка сервера
- `503 Service Unavailable` - Сервис недоступен

### Коды ошибок API
- `INVALID_TOKEN` - Неверный JWT токен
- `TOKEN_EXPIRED` - Токен истек
- `INSUFFICIENT_PERMISSIONS` - Недостаточно прав
- `TENANT_NOT_FOUND` - Tenant не найден
- `TUNNEL_NOT_FOUND` - Туннель не найден
- `TUNNEL_LIMIT_EXCEEDED` - Превышен лимит туннелей
- `BANDWIDTH_LIMIT_EXCEEDED` - Превышен лимит пропускной способности
- `RATE_LIMIT_EXCEEDED` - Превышен лимит запросов
- `INVALID_CONFIGURATION` - Неверная конфигурация
- `SERVICE_UNAVAILABLE` - Сервис недоступен

## Примеры использования

### Создание туннеля с cURL

```bash
# Аутентификация
curl -X POST http://localhost:9090/api/v1/auth/validate \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "tenant_id": "relay-test-tenant"
  }'

# Создание туннеля
curl -X POST http://localhost:9090/api/v1/tunnels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "mode": "proxy",
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "protocol": "tcp",
    "tenant_id": "relay-test-tenant"
  }'

# Отправка heartbeat
curl -X POST http://localhost:9090/api/v1/tunnels/tunnel-12345/heartbeat \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Закрытие туннеля
curl -X DELETE http://localhost:9090/api/v1/tunnels/tunnel-12345 \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### Создание туннеля с Python

```python
import requests
import json

# Базовый URL
base_url = "http://localhost:9090/api/v1"

# JWT токен
token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Заголовки
headers = {
    "Content-Type": "application/json",
    "Authorization": f"Bearer {token}"
}

# Создание туннеля
tunnel_data = {
    "mode": "proxy",
    "remote_host": "192.168.1.100",
    "remote_port": 3389,
    "protocol": "tcp",
    "tenant_id": "relay-test-tenant"
}

response = requests.post(
    f"{base_url}/tunnels",
    headers=headers,
    json=tunnel_data
)

if response.status_code == 200:
    tunnel = response.json()["data"]
    print(f"Туннель создан: {tunnel['tunnel_id']}")
    print(f"Локальный порт: {tunnel['local_port']}")
else:
    print(f"Ошибка: {response.json()}")
```

### Мониторинг с JavaScript

```javascript
// Функция для получения статистики
async function getStats() {
    try {
        const response = await fetch('http://localhost:9090/api/v1/stats');
        const data = await response.json();
        
        if (data.status === 'success') {
            const stats = data.data;
            console.log(`Активных соединений: ${stats.connections.active}`);
            console.log(`Активных туннелей: ${stats.tunnels.active}`);
            console.log(`Среднее время отклика: ${stats.performance.avg_response_time_ms}ms`);
        }
    } catch (error) {
        console.error('Ошибка получения статистики:', error);
    }
}

// Обновление статистики каждые 30 секунд
setInterval(getStats, 30000);
```

## Rate Limiting

API использует rate limiting для защиты от злоупотреблений:

- **По умолчанию**: 100 запросов в минуту на IP
- **Для аутентифицированных пользователей**: 1000 запросов в минуту
- **Для tenant'ов**: Настраивается индивидуально

При превышении лимита возвращается HTTP 429 с заголовками:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1691234567
```

## WebSocket API

Для real-time уведомлений доступен WebSocket API:

### Подключение
```javascript
const ws = new WebSocket('ws://localhost:9090/ws');

ws.onopen = function() {
    console.log('WebSocket подключен');
    
    // Подписка на события туннелей
    ws.send(JSON.stringify({
        type: 'subscribe',
        channel: 'tunnels',
        tenant_id: 'relay-test-tenant'
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Получено событие:', data);
};
```

### События
- `tunnel_created` - Туннель создан
- `tunnel_closed` - Туннель закрыт
- `tunnel_error` - Ошибка туннеля
- `heartbeat_received` - Получен heartbeat
- `tenant_limits_exceeded` - Превышены лимиты tenant'а

---

## Cadence Workflow Engine API

### Базовый URL
```
http://localhost:9091/api/v1/cadence
```

### 1. Health Check

**GET** `/health`

Проверка состояния Cadence Workflow Engine.

#### Ответ
```json
{
  "status": "healthy",
  "domain": "cloudbridge"
}
```

#### Пример
```bash
curl -s http://localhost:9091/api/v1/cadence/health | jq
```

### 2. Relay Deployment Workflow

**POST** `/workflows/relay-deployment`

Запуск workflow развертывания Relay сервера.

#### Запрос
```json
{
  "config": {
    "test": true,
    "environment": "production"
  },
  "wireguard_config": {
    "enabled": true,
    "port": 51820
  },
  "quic_config": {
    "enabled": true,
    "port": 9090
  },
  "services": ["relay", "wireguard", "quic"],
  "timeout": "5m"
}
```

#### Ответ
```json
{
  "status": "started",
  "workflow_id": "RelayDeploymentWorkflow-1757145801968487275"
}
```

#### Пример
```bash
curl -X POST http://localhost:9091/api/v1/cadence/workflows/relay-deployment \
  -H "Content-Type: application/json" \
  -d '{
    "config": {"test": true},
    "services": ["relay", "wireguard", "quic"]
  }' | jq
```

### 3. Backup Workflow

**POST** `/workflows/backup`

Запуск workflow резервного копирования.

#### Запрос
```json
{
  "backup_config": {
    "type": "full",
    "compression": "gzip"
  },
  "config_paths": ["/etc/relay", "/var/lib/relay"],
  "data_paths": ["/var/log/relay"],
  "cleanup_config": {
    "retention_days": 7,
    "max_backups": 10
  }
}
```

#### Ответ
```json
{
  "status": "started",
  "workflow_id": "BackupWorkflow-1757145807687688437"
}
```

#### Пример
```bash
curl -X POST http://localhost:9091/api/v1/cadence/workflows/backup \
  -H "Content-Type: application/json" \
  -d '{
    "backup_config": {"type": "full"},
    "config_paths": ["/etc/relay"]
  }' | jq
```

### 4. Workflow Status

**GET** `/workflows/status?workflow_id={workflow_id}`

Получение статуса выполнения workflow.

#### Параметры
- `workflow_id` (string, required) - ID workflow

#### Ответ
```json
{
  "status": "RUNNING",
  "workflow_id": "RelayDeploymentWorkflow-1757145801968487275"
}
```

#### Пример
```bash
curl -s "http://localhost:9091/api/v1/cadence/workflows/status?workflow_id=RelayDeploymentWorkflow-1757145801968487275" | jq
```

### 5. JavaScript SDK для Cadence API

```javascript
class CadenceClient {
    constructor(baseUrl = 'http://localhost:9091/api/v1/cadence') {
        this.baseUrl = baseUrl;
    }

    async health() {
        const response = await fetch(`${this.baseUrl}/health`);
        return await response.json();
    }

    async startRelayDeployment(config) {
        const response = await fetch(`${this.baseUrl}/workflows/relay-deployment`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });
        return await response.json();
    }

    async startBackup(config) {
        const response = await fetch(`${this.baseUrl}/workflows/backup`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(config)
        });
        return await response.json();
    }

    async getWorkflowStatus(workflowId) {
        const response = await fetch(`${this.baseUrl}/workflows/status?workflow_id=${workflowId}`);
        return await response.json();
    }
}

// Использование
const cadence = new CadenceClient();

// Проверка здоровья
const health = await cadence.health();
console.log('Cadence status:', health.status);

// Запуск развертывания
const deployment = await cadence.startRelayDeployment({
    config: { test: true },
    services: ['relay', 'wireguard']
});
console.log('Deployment started:', deployment.workflow_id);

// Проверка статуса
const status = await cadence.getWorkflowStatus(deployment.workflow_id);
console.log('Workflow status:', status.status);
```

---

*API Reference для CloudBridge Relay Server v1.0.0*
*Обновлено: 2025-09-06*
