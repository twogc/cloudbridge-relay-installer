# 🔌 **API CloudBridge Relay**

## **Обзор**

Этот раздел содержит документацию по API CloudBridge Relay, включая протоколы, эндпоинты и интеграции.

## 📚 **Документы**

### **API документация:**
- [**MASQUE HTTP3 QUIC Validation Report**](MASQUE_HTTP3_QUIC_VALIDATION_REPORT.md) - Валидация MASQUE HTTP3 QUIC протокола

## 🎯 **API архитектура**

### **Основные протоколы:**
- **QUIC/HTTP-3** - Базовый транспортный протокол
- **MASQUE** - Протокол для туннелей поверх HTTP-3
- **REST API** - HTTP API для управления
- **WebSocket** - Реальное время коммуникации

### **Порты API:**
| Порт | Протокол | Назначение | Описание |
|------|----------|------------|----------|
| 5550 | TCP | Main Server | Основной TCP сервер |
| 5551 | TCP | Metrics | Prometheus метрики |
| 5552 | TCP | P2P API | P2P API для управления |
| 5553 | TCP | HTTP API | HTTP API для клиентов |
| 5554 | TCP | HTTPS API | HTTPS API с TLS |
| 5555 | UDP | QUIC Transport | QUIC транспорт |

##  **REST API**

### **Базовый URL:**
```
http://localhost:5553/api/v1
https://localhost:5554/api/v1
```

### **Основные эндпоинты:**

#### **Health Check:**
```http
GET /health
```
**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-22T21:30:33.099704708Z",
  "uptime": "41.462181905s",
  "version": "1.0.0"
}
```

#### **P2P API Health:**
```http
GET /health
```
**Ответ:**
```json
{
  "status": "healthy",
  "timestamp": "2025-09-22T21:30:44.898147209Z",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "websocket": "healthy"
  },
  "metrics": {
    "active_connections": 0,
    "goroutines": 41,
    "memory_usage_mb": 3,
    "online_peers": 0,
    "total_peers": 0,
    "uptime_seconds": 64.269963502,
    "websocket_connections": 0
  },
  "version": "v1.0.0",
  "uptime": "1m4.269963502s"
}
```

#### **Метрики:**
```http
GET /metrics
```
**Ответ:** Prometheus метрики в текстовом формате

## 🚀 **QUIC API**

### **QUIC Transport:**
- **Порт**: 5555 (UDP)
- **Протокол**: QUIC
- **Функции**: Высокопроизводительный транспорт

### **Enhanced QUIC:**
- **Порт**: 9092 (UDP)
- **Протокол**: QUIC с AI
- **Функции**: Интеллектуальная маршрутизация

##  **MASQUE API**

### **MASQUE Protocol:**
- **Порт**: 8443 (UDP)
- **Протокол**: MASQUE поверх HTTP-3
- **Функции**: Туннелирование UDP/TCP

### **Эндпоинты MASQUE:**
```
/.well-known/masque/udp/{host}:{port}
/.well-known/masque/tcp/{host}:{port}
```

## 📊 **WebSocket API**

### **WebSocket соединения:**
- **Порт**: 5553 (TCP)
- **Протокол**: WebSocket
- **Функции**: Реальное время коммуникация

### **События:**
```javascript
// Подключение
const ws = new WebSocket('ws://localhost:5553/ws');

// События
ws.onopen = () => console.log('Connected');
ws.onmessage = (event) => console.log('Message:', event.data);
ws.onclose = () => console.log('Disconnected');
```

## 🔐 **Аутентификация**

### **JWT токены:**
```http
Authorization: Bearer <jwt-token>
```

### **Keycloak интеграция:**
- **Server URL**: https://auth.2gc.ru
- **Realm**: cloudbridge
- **Client ID**: quic-client

##  **Метрики API**

### **Prometheus метрики:**
- **relay_connections_total** - Общее количество соединений
- **relay_connections_active** - Активные соединения
- **relay_bandwidth_bytes** - Пропускная способность
- **relay_latency_seconds** - Задержка
- **relay_errors_total** - Количество ошибок
- **relay_requests_total** - Количество запросов

### **Системные метрики:**
- **go_gc_duration_seconds** - Время сборки мусора
- **go_memstats_alloc_bytes** - Использование памяти
- **go_goroutines** - Количество горутин

## 🧪 **Тестирование API**

### **cURL примеры:**
```bash
# Health check
curl http://localhost:5553/health

# P2P API health
curl http://localhost:5552/health

# Метрики
curl http://localhost:5551/metrics

# MASQUE (требует QUIC клиент)
# Используйте специализированные клиенты
```

### **Тестирование QUIC:**
```bash
# QUIC клиент (пример)
./quic-client --host=localhost --port=5555
```

## 📝 **OpenAPI спецификация**

*OpenAPI спецификация будет добавлена в будущих версиях*

## 🔍 **Диагностика API**

### **Проверка доступности:**
```bash
# HTTP API
curl -I http://localhost:5553/health

# HTTPS API
curl -k -I https://localhost:5554/health

# Метрики
curl -I http://localhost:5551/metrics
```

### **Проверка портов:**
```bash
# TCP порты
netstat -tlnp | grep -E "555[0-5]"

# UDP порты
netstat -ulnp | grep -E "555[0-5]|909[0-2]|19302|8443"
```

---

**🔌 API CloudBridge Relay готово к использованию!**
