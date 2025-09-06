# API Documentation

## 📁 Файлы в этой папке

- **[api-reference.md](./api-reference.md)** - Справочник по API

## 🔌 API Endpoints

### CloudBridge Relay API
- **Health Check** - `/health` - Проверка состояния сервиса
- **Metrics** - `/metrics` - Prometheus метрики
- **Status** - `/status` - Статус системы
- **Config** - `/config` - Конфигурация

### QUIC Service API
- **Connection** - `/quic/connect` - Установка QUIC соединения
- **Stream** - `/quic/stream` - Управление потоками
- **Stats** - `/quic/stats` - Статистика QUIC

### WireGuard Service API
- **Tunnel** - `/wireguard/tunnel` - Управление туннелями
- **Peers** - `/wireguard/peers` - Управление пирами
- **Config** - `/wireguard/config` - Конфигурация WireGuard

### AI Service API
- **Predict** - `/ai/predict` - AI предсказания
- **Train** - `/ai/train` - Обучение модели
- **Model** - `/ai/model` - Управление моделями

## 📊 API Examples

### Health Check
```bash
curl http://localhost:9090/health
```

### Metrics
```bash
curl http://localhost:8084/metrics
```

### QUIC Connection
```bash
curl -X POST http://localhost:9090/quic/connect \
  -H "Content-Type: application/json" \
  -d '{"host": "example.com", "port": 443}'
```

## 🔐 Authentication

### API Key
```bash
curl -H "X-API-Key: your-api-key" http://localhost:9090/api/endpoint
```

### JWT Token
```bash
curl -H "Authorization: Bearer your-jwt-token" http://localhost:9090/api/endpoint
```

## 📈 Rate Limiting

- **Default**: 100 requests per minute
- **Burst**: 200 requests per minute
- **Headers**: `X-RateLimit-Limit`, `X-RateLimit-Remaining`

## 🚨 Error Codes

- **200** - Success
- **400** - Bad Request
- **401** - Unauthorized
- **403** - Forbidden
- **404** - Not Found
- **429** - Too Many Requests
- **500** - Internal Server Error



