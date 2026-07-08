# CloudBridge Pilot Project - Готовое решение

**Дата:** $(date)  
**Статус:** ГОТОВО ДЛЯ ПИЛОТНОГО ПРОЕКТА  
**Архитектура:** Kubernetes + VK Cloud LoadBalancer  

## Готовые endpoints для пилотного проекта

### Прямой доступ к Kubernetes (рекомендуется для пилота)

**Base URL**: `http://192.168.49.2:32500`

| Endpoint | URL | Описание | Статус |
|----------|-----|----------|--------|
| Health Check | `http://192.168.49.2:32500/health` | Проверка здоровья системы | K |
| API v1 | `http://192.168.49.2:32500/api/v1/` | REST API | K |
| WebSocket | `ws://192.168.49.2:32500/ws/` | WebSocket соединения | K |
| Metrics | `http://192.168.49.2:32500/metrics` | Prometheus метрики | K |
| P2P API | `http://192.168.49.2:32500/api/v1/peers/` | P2P управление | K |

### Доступ через домены (через VK Cloud LB)

**Проблема**: Бесконечный редирект HTTP ↔ HTTPS  
**Статус**: Требует настройки relay сервера  

| Домен | URL | Статус |
|-------|-----|--------|
| relay.2gc.ru | `https://relay.2gc.ru/health` | Редирект |
| masque.2gc.ru | `https://masque.2gc.ru/health` | Редирект |
| api.2gc.ru | `https://api.2gc.ru/health` | Редирект |

## Рекомендации для пилотного проекта

### Вариант 1: Прямой доступ (быстрый старт)

Используйте прямой доступ к Kubernetes для немедленного начала пилотного проекта:

```bash
# Базовый URL для всех запросов
BASE_URL="http://192.168.49.2:32500"

# Примеры использования
curl "$BASE_URL/health"
curl "$BASE_URL/api/v1/peers/discover"
curl "$BASE_URL/metrics"
```

**Преимущества**:
- аботает немедленно
- ет проблем с редиректами
- олная функциональность API
- одходит для разработки и тестирования

**Недостатки**:
- Нет SSL (HTTP только)
- Прямой IP вместо доменного имени

### Вариант 2: Port-forward для локального доступа

Настройте port-forward для удобного локального доступа:

```bash
# Запустите port-forward
kubectl port-forward -n cloudbridge svc/cloudbridge-relay-external 8080:5552 &

# Используйте локальный доступ
curl "http://localhost:8080/health"
curl "http://localhost:8080/api/v1/peers/discover"
```

### Вариант 3: Доменный доступ (после настройки)

После настройки relay сервера будут доступны домены:

```bash
# После исправления редиректов
curl "https://relay.2gc.ru/health"
curl "https://api.2gc.ru/api/v1/peers/discover"
```

## API Endpoints для пилотного проекта

### Health & Status

```bash
# Проверка здоровья системы
GET http://192.168.49.2:32500/health

# Ответ:
{
  "status": "healthy",
  "timestamp": "2025-09-28T10:14:30Z",
  "services": {
    "database": "healthy",
    "redis": "healthy", 
    "websocket": "healthy"
  },
  "metrics": {
    "active_connections": 0,
    "online_peers": 0,
    "total_peers": 0,
    "uptime_seconds": 3265
  },
  "version": "v1.0.0"
}
```

### P2P Management

```bash
# Обнаружение пиров
GET http://192.168.49.2:32500/api/v1/peers/discover

# Регистрация пира
POST http://192.168.49.2:32500/api/v1/peers/register
Content-Type: application/json
{
  "peer_id": "peer-123",
  "address": "192.168.1.100:8080"
}

# Список пиров
GET http://192.168.49.2:32500/api/v1/peers/list
```

### Tunnel Management

```bash
# Создание туннеля
POST http://192.168.49.2:32500/api/v1/tunnels/create
Content-Type: application/json
{
  "tunnel_type": "quic",
  "destination": "target.example.com:443"
}

# Список туннелей
GET http://192.168.49.2:32500/api/v1/tunnels/list

# Статус туннеля
GET http://192.168.49.2:32500/api/v1/tunnels/{tunnel_id}/status
```

### WebSocket Connections

```javascript
// WebSocket подключение
const ws = new WebSocket('ws://192.168.49.2:32500/ws/peers');

ws.onopen = function() {
    console.log('Connected to CloudBridge');
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};
```

### Metrics & Monitoring

```bash
# Prometheus метрики
GET http://192.168.49.2:32500/metrics

# Системные метрики
GET http://192.168.49.2:32500/api/v1/metrics/system

# Метрики производительности
GET http://192.168.49.2:32500/api/v1/metrics/performance
```

##  Настройка клиента для пилотного проекта

### Конфигурация CloudBridge Client

```yaml
# config.yaml
server:
  address: "192.168.49.2:32500"
  protocol: "http"  # Для пилотного проекта
  
client:
  id: "pilot-client-001"
  tenant_id: "pilot-tenant"
  
transport:
  primary: "grpc"
  fallback: "json"
  
security:
  tls_enabled: false  # Для пилотного проекта
  
logging:
  level: "debug"
  output: "stdout"
```

### Пример подключения

```go
package main

import (
    "fmt"
    "github.com/twogc/cloudbridge-relay-installer/test/relay-client/pkg/relay"
)

func main() {
    config := &relay.Config{
        ServerAddress: "192.168.49.2:32500",
        Protocol:      "http",
        ClientID:      "pilot-client-001",
        TenantID:      "pilot-tenant",
    }
    
    client := relay.NewClient(config)
    
    // Подключение
    if err := client.Connect(); err != nil {
        panic(err)
    }
    
    // Аутентификация
    token := "your-jwt-token"
    if err := client.Authenticate(token); err != nil {
        panic(err)
    }
    
    fmt.Println("CloudBridge client connected successfully!")
}
```

## Мониторинг для пилотного проекта

### Проверка доступности

```bash
#!/bin/bash
# pilot-health-check.sh

BASE_URL="http://192.168.49.2:32500"

echo "=== CloudBridge Pilot Health Check ==="

# Health endpoint
echo -n "Health Check: "
if curl -s "$BASE_URL/health" > /dev/null; then
    echo " OK"
else
    echo " FAIL"
fi

# API endpoint
echo -n "API Check: "
if curl -s "$BASE_URL/api/v1/peers/discover" > /dev/null; then
    echo " OK"
else
    echo " FAIL"
fi

# Metrics endpoint
echo -n "Metrics Check: "
if curl -s "$BASE_URL/metrics" > /dev/null; then
    echo " OK"
else
    echo " FAIL"
fi

echo "=== End Health Check ==="
```

### Логи и отладка

```bash
# Логи Kubernetes pods
kubectl logs -n cloudbridge -l app=cloudbridge-relay -f

# Статус сервисов
kubectl get pods,svc -n cloudbridge

# Метрики производительности
curl -s "http://192.168.49.2:32500/metrics" | grep cloudbridge
```

## Готовность к пилотному проекту

### то готово:

1. **Kubernetes кластер** - полностью функционален
2. **CloudBridge сервисы** - все компоненты работают
3. **API endpoints** - доступны для интеграции
4. **Health checks** - система мониторинга работает
5. **WebSocket** - реальное время поддерживается
6. **P2P функциональность** - готова к тестированию
7. **Метрики** - Prometheus интеграция активна

### Чек-лист для запуска пилота:

- [x] Kubernetes кластер запущен
- [x] CloudBridge сервисы развернуты  
- [x] API endpoints доступны
- [x] Health checks работают
- [x] Документация API готова
- [x] Примеры клиентского кода подготовлены
- [x] Мониторинг настроен

## Запуск пилотного проекта

**Система готова к использованию!**

Используйте базовый URL `http://192.168.49.2:32500` для всех интеграций и начинайте разработку пилотного проекта немедленно.

**Контакты для поддержки**: Все конфигурации и документация сохранены в проекте.
