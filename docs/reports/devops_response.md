# 📋 ОТВЕТ ДЛЯ DEVOPS СЕРВЕРНОГО КЛИЕНТА

## 🎯 КРАТКИЙ ОТВЕТ

**Проблема решена!** Вот все, что вам нужно:

### ✅ ВАЛИДНЫЕ ЗНАЧЕНИЯ connection_type:
- `"wireguard"` - только WireGuard
- `"quic"` - только QUIC  
- `"tls"` - только TLS
- `"hybrid"` - WireGuard + QUIC (рекомендуется)

### 🔑 ГОТОВЫЙ ВАЛИДНЫЙ JWT ТОКЕН:
```
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE3NTc2Mzc4ODQsIm5iZiI6MTc1NzYzNzg4NCwiZXhwIjoxNzU3NjQxNDg0LCJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJqdGkiOiJwMnBfdG9rZW5fMTc1NzYzNzg4NCIsImNvbm5lY3Rpb25fdHlwZSI6Imh5YnJpZCIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjM3ODg0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJhZGRyZXNzIjoiMTAuMC4wLjEvMjQiLCJtdHUiOjE0MjAsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il19fQ.Js8lRTZgTBChJYnkZlcnO-i04tPEgzV6WHspsR-ALvo
```

### 🌐 P2P API ENDPOINT:
```
http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123
```

---

## 📝 ПОДРОБНЫЕ ОТВЕТЫ НА ВСЕ ВОПРОСЫ

### 1. ❌ ПРОБЛЕМА С connection_type

**Вопрос:** Какие именно значения connection_type принимает сервер?

**Ответ:** Сервер принимает ТОЛЬКО эти 4 значения:
- `"wireguard"` - только WireGuard протокол
- `"quic"` - только QUIC протокол  
- `"tls"` - только TLS протокол
- `"hybrid"` - WireGuard + QUIC (рекомендуется)

**❌ НЕПРАВИЛЬНЫЕ значения (которые вы пробовали):**
- `"p2p-mesh"` - НЕ поддерживается
- `"quick"` - НЕ поддерживается (возможно, опечатка в "quic")
- `"p2p"` - НЕ поддерживается

### 2. 🔧 СТРУКТУРА JWT ТОКЕНА

**Вопрос:** Какая точная структура JWT claims ожидается?

**Ответ:** Вот правильная структура:

```json
{
  "iat": 1757637884,                    // Issued at (timestamp)
  "nbf": 1757637884,                    // Not before (timestamp)
  "exp": 1757641484,                    // Expires (timestamp)
  "iss": "https://auth.2gc.ru/realms/relay-servers",  // Issuer
  "sub": "p2p-server-client",           // Subject
  "jti": "p2p_token_1757637884",        // JWT ID
  
  // P2P специфичные claims
  "connection_type": "hybrid",          // ОБЯЗАТЕЛЬНО: wireguard|quic|tls|hybrid
  "tenant_id": "test-tenant-123",       // ОБЯЗАТЕЛЬНО
  "device_id": "device_1757637884",     // ОБЯЗАТЕЛЬНО
  "permissions": ["p2p:connect", "p2p:discover", "p2p:mesh"],
  
  // WireGuard конфигурация (ОБЯЗАТЕЛЬНА для wireguard/hybrid)
  "wireguard_config": {
    "interface_name": "wg0",            // ОБЯЗАТЕЛЬНО
    "listen_port": 51820,               // ОБЯЗАТЕЛЬНО (1-65535)
    "address": "10.0.0.1/24",          // ОБЯЗАТЕЛЬНО
    "mtu": 1420,                        // ОБЯЗАТЕЛЬНО (>0)
    "allowed_ips": ["10.0.0.0/24"]     // ОБЯЗАТЕЛЬНО (не пустой)
  }
}
```

### 3. 🌐 P2P API ENDPOINTS

**Вопрос:** Какие endpoints доступны для P2P функциональности?

**Ответ:** Полный список P2P API endpoints:

#### Peer Management:
- `POST /api/v1/peers/register` - Регистрация пира
- `GET /api/v1/peers/discover?tenant_id=<id>` - Discovery пиров
- `PUT /api/v1/peers/:peer_id/status` - Обновление статуса пира
- `GET /api/v1/peers/:peer_id` - Информация о пире

#### Relay:
- `POST /api/v1/relay/route` - Маршрутизация через relay
- `GET /api/v1/relay/stats` - Статистика relay

#### Health & Monitoring:
- `GET /health` - Health check
- `GET /api/v1/health/` - Health check (API)
- `GET /api/v1/health/metrics` - Системные метрики
- `GET /api/v1/health/peers` - Метрики пиров
- `GET /api/v1/health/connections` - Метрики соединений

#### WebSocket:
- `GET /ws/peers` - WebSocket для real-time соединений

### 4. 🧪 ГОТОВЫЙ ТЕСТОВЫЙ ТОКЕН

**Вопрос:** Можете ли вы предоставить готовый валидный JWT токен для тестирования?

**Ответ:** ✅ ДА! Вот готовый токен:

```
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE3NTc2Mzc4ODQsIm5iZiI6MTc1NzYzNzg4NCwiZXhwIjoxNzU3NjQxNDg0LCJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9yZWxheS1zZXJ2ZXJzIiwic3ViIjoicDJwLXNlcnZlci1jbGllbnQiLCJqdGkiOiJwMnBfdG9rZW5fMTc1NzYzNzg4NCIsImNvbm5lY3Rpb25fdHlwZSI6Imh5YnJpZCIsInRlbmFudF9pZCI6InRlc3QtdGVuYW50LTEyMyIsImRldmljZV9pZCI6ImRldmljZV8xNzU3NjM3ODg0IiwicGVybWlzc2lvbnMiOlsicDJwOmNvbm5lY3QiLCJwMnA6ZGlzY292ZXIiLCJwMnA6bWVzaCJdLCJ3aXJlZ3VhcmRfY29uZmlnIjp7ImludGVyZmFjZV9uYW1lIjoid2cwIiwibGlzdGVuX3BvcnQiOjUxODIwLCJhZGRyZXNzIjoiMTAuMC4wLjEvMjQiLCJtdHUiOjE0MjAsImFsbG93ZWRfaXBzIjpbIjEwLjAuMC4wLzI0Il19fQ.Js8lRTZgTBChJYnkZlcnO-i04tPEgzV6WHspsR-ALvo
```

**Срок действия:** До 12.09.2025 01:44:44 UTC

---

## 🚀 ИНСТРУКЦИЯ ПО ИСПОЛЬЗОВАНИЮ

### 1. Тестирование токена:
```bash
curl -H "Authorization: Bearer <TOKEN>" \
     "http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123"
```

### 2. Регистрация пира:
```bash
curl -X POST \
     -H "Authorization: Bearer <TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
       "peer_id": "your-peer-id",
       "tenant_id": "test-tenant-123",
       "org_id": "test-org-456",
       "device_id": "your-device-id",
       "connection_type": "hybrid",
       "wireguard_config": {
         "interface_name": "wg0",
         "listen_port": 51820,
         "address": "10.0.0.2/24",
         "mtu": 1420,
         "allowed_ips": ["10.0.0.0/24"]
       }
     }' \
     "http://95.163.250.190:30082/api/v1/peers/register"
```

### 3. Health check:
```bash
curl "http://95.163.250.190:30082/health"
```

---

## 🔍 ДИАГНОСТИКА ПРОБЛЕМ

### ❌ Частые ошибки:

1. **"invalid connection_type: p2p-mesh"**
   - **Решение:** Используйте `"hybrid"` вместо `"p2p-mesh"`

2. **"wireguard_config is required"**
   - **Решение:** Добавьте `wireguard_config` для `hybrid`/`wireguard`

3. **"connection_type is required"**
   - **Решение:** Добавьте поле `connection_type`

4. **"token expired"**
   - **Решение:** Создайте новый токен с актуальным `exp`

### ✅ Что работает:
- JWT секрет: `1Pc6bvrTUDtXMKm6mMF6+7UJi/EYAdOcGPkXCRYV4xI=`
- Health endpoint: `http://95.163.250.190:30082/health`
- P2P API: `http://95.163.250.190:30082/api/v1/peers/`

---

## 📞 ПОДДЕРЖКА

Если у вас остались вопросы:
1. Проверьте логи сервера: `kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay`
2. Убедитесь, что используете правильный `connection_type`
3. Проверьте, что `wireguard_config` присутствует для `hybrid`/`wireguard`

**Удачи с интеграцией! 🎯**
