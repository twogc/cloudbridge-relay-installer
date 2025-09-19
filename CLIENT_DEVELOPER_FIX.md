# Исправление для разработчика клиента

## 🚨 **Проблема: HTTP 400 "Invalid request format"**

Клиент получает ошибку при регистрации пира из-за неправильного формата запроса.

## ❌ **Неправильный формат (что отправляет клиент):**
```json
{
  "device_id": "my-device-123",
  "connection_type": "relay-assisted",
  "wireguard_config": {
    "public_key": "your-public-key",
    "endpoint": "your-endpoint",
    "listen_port": 51820,
    "allowed_ips": ["10.100.77.0/24"]
  }
}
```

## ✅ **Правильный формат (что ожидает сервер):**
```json
{
  "public_key": "test-peer-public-key-12345",
  "allowed_ips": ["10.100.77.0/24"]
}
```

## 🔧 **Исправление для клиента:**

### 1. **Обновить структуру запроса регистрации:**
```javascript
// НЕПРАВИЛЬНО:
const registrationData = {
  device_id: "my-device-123",
  connection_type: "relay-assisted",
  wireguard_config: {
    public_key: "your-public-key",
    endpoint: "your-endpoint",
    listen_port: 51820,
    allowed_ips: ["10.100.77.0/24"]
  }
};

// ПРАВИЛЬНО:
const registrationData = {
  public_key: "your-public-key",
  allowed_ips: ["10.100.77.0/24"]
};
```

### 2. **Обновить код регистрации:**
```javascript
async registerPeer(publicKey, allowedIPs) {
  const response = await fetch(`${this.serverUrl}/api/v1/tenants/${this.tenantId}/peers/register`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${this.jwtToken}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      public_key: publicKey,
      allowed_ips: allowedIPs
    })
  });

  const data = await response.json();
  if (data.success) {
    this.peerId = data.peer_id;
    this.relaySessionId = data.relay_session_id;
    return data;
  } else {
    throw new Error('Failed to register peer: ' + data.message);
  }
}
```

### 3. **Пример использования:**
```javascript
// Регистрация пира
const peerInfo = await client.registerPeer(
  'test-peer-public-key-12345',
  ['10.100.77.0/24']
);

console.log('Peer registered:', peerInfo);
// {
//   "success": true,
//   "peer_id": "peer_server-client-server-1758105689169",
//   "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a",
//   "registered_at": "2025-09-19T00:39:32.642169441Z"
// }
```

## 📋 **Полный список изменений:**

### **1. Регистрация пира:**
```javascript
// Endpoint: POST /api/v1/tenants/{tenant_id}/peers/register
// Запрос:
{
  "public_key": "your-public-key",
  "allowed_ips": ["10.100.77.0/24"]
}

// Ответ:
{
  "success": true,
  "peer_id": "peer_abc123def456",
  "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a",
  "registered_at": "2025-09-19T00:39:32.642169441Z"
}
```

### **2. Heartbeat (без изменений):**
```javascript
// Endpoint: POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat
// Запрос:
{
  "status": "active",
  "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a"
}
```

### **3. Обнаружение пиров (без изменений):**
```javascript
// Endpoint: GET /api/v1/tenants/{tenant_id}/peers/discover
// Ответ:
{
  "success": true,
  "tenant_id": "tenant-216420165",
  "peers": [...]
}
```

## 🎯 **Критически важно:**

1. **Убрать** поля `device_id`, `connection_type`, `wireguard_config`
2. **Оставить только** `public_key` и `allowed_ips` в корне запроса
3. **Использовать** `relay_session_id` из ответа для heartbeat

## 🧪 **Тестирование:**

```bash
# Правильный запрос регистрации
curl -k -X POST "https://edge.2gc.ru:9444/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-public-key-12345",
    "allowed_ips": ["10.100.77.0/24"]
  }'
```

**После этого исправления клиент должен успешно регистрироваться!** ✅
