# Краткое резюме для разработчика клиента

## 🚀 **Главные изменения**

### 1. **Новые API endpoints**
```javascript
// HTTP (для разработки)
const serverUrl = 'http://edge.2gc.ru:8083';

// HTTPS (для продакшена) 
const serverUrl = 'https://edge.2gc.ru:9444';
```

### 2. **Обязательный heartbeat каждые 30 секунд**
```javascript
POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat
{
  "status": "active",
  "relay_session_id": "rs_abc123def456"
}
```

### 3. **Мультитунельность**
- JWT токен должен содержать `tenant_id`
- Пиры видят только других пиров в том же tenant

### 4. **Новые поля в ответах**
- `relay_session_id` - для heartbeat
- `peer_id` - уникальный идентификатор
- `tenant_id` - изоляция пользователей

## ⚠️ **Критически важно обновить:**

1. **URL сервера** - изменить на новый endpoint
2. **Добавить heartbeat** - каждые 30 секунд
3. **Использовать tenant_id** - из JWT токена
4. **Обработать новые поля** - в ответах API

## 🔐 **Безопасность**
- HTTPS доступен на порту 9444
- SSL сертификат действителен до 2026 года
- JWT аутентификация обязательна

## 📡 **P2P транспорт**
- QUIC: `edge.2gc.ru:9090` (UDP)
- STUN: `edge.2gc.ru:19302` (UDP)
- Enhanced QUIC: `edge.2gc.ru:9092` (UDP)

## 🎯 **Минимальные изменения кода:**

```javascript
// 1. Обновить URL
const serverUrl = 'https://edge.2gc.ru:9444';

// 2. Добавить heartbeat
setInterval(() => {
  fetch(`${serverUrl}/api/v1/tenants/${tenantId}/peers/${peerId}/heartbeat`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${jwtToken}` },
    body: JSON.stringify({ status: 'active', relay_session_id })
  });
}, 30000);

// 3. Использовать tenant_id из JWT
const tenantId = jwtPayload.tenant_id;
```

**Готово к использованию!** 🚀
