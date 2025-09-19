# Руководство разработчика клиента CloudBridge Relay

## 🚀 Обновления P2P маршрутизации

CloudBridge Relay теперь поддерживает полноценную P2P маршрутизацию с мультитунельностью и heartbeat системой!

## 📡 API Endpoints

### Основные endpoints:

#### HTTP (для разработки):
```javascript
const serverUrl = 'http://edge.2gc.ru:8083';
```

#### HTTPS (для продакшена):
```javascript
const serverUrl = 'https://edge.2gc.ru:9444';
```

#### Мониторинг:
```javascript
const metricsUrl = 'http://edge.2gc.ru:9091';
```

## 🔐 Аутентификация

### JWT токен обязателен для всех запросов:
```javascript
const headers = {
  'Authorization': `Bearer ${jwtToken}`,
  'Content-Type': 'application/json'
};
```

### JWT токен должен содержать:
- `tenant_id` - для изоляции пользователей
- `exp` - время истечения
- Валидная подпись

## 📋 API Endpoints

### 1. Регистрация пира
```javascript
POST /api/v1/tenants/{tenant_id}/peers/register

// Запрос
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

// Ответ
{
  "success": true,
  "peer_id": "peer_abc123def456",
  "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a",
  "tenant_id": "tenant-216420165",
  "wireguard_config": {
    "public_key": "test-peer-external-dns",
    "endpoint": "",
    "listen_port": 0,
    "allowed_ips": ["10.100.77.12/32"]
  }
}
```

### 2. Обнаружение пиров
```javascript
GET /api/v1/tenants/{tenant_id}/peers/discover

// Ответ
{
  "success": true,
  "tenant_id": "tenant-216420165",
  "peers": [
    {
      "id": "peer_server-client-server-1758105689169",
      "peer_id": "peer_server-client-server-1758105689169",
      "tenant_id": "tenant-216420165",
      "device_id": "server-client-server-1758105689169",
      "connection_type": "relay-assisted",
      "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a",
      "wireguard_config": {
        "public_key": "test-peer-external-dns",
        "endpoint": "",
        "listen_port": 0,
        "allowed_ips": ["10.100.77.12/32"]
      },
      "is_online": true,
      "last_seen": "2025-09-18T23:32:24.44412057Z",
      "registered_at": "2025-09-18T23:51:04.305782317Z",
      "active_connections": 0,
      "mesh_status": "connected",
      "latency_ms": 0
    }
  ]
}
```

### 3. Heartbeat (ОБЯЗАТЕЛЬНО!)
```javascript
POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat

// Запрос
{
  "status": "active",
  "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a"
}

// Ответ
{
  "success": true,
  "peer_id": "peer_abc123def456",
  "status": "active",
  "last_seen": "2025-09-19T00:30:56.616465773Z"
}
```

### 4. Получение информации о пире
```javascript
GET /api/v1/tenants/{tenant_id}/peers/{peer_id}

// Ответ
{
  "success": true,
  "peer": {
    "id": "peer_abc123def456",
    "peer_id": "peer_abc123def456",
    "tenant_id": "tenant-216420165",
    "device_id": "my-device-123",
    "connection_type": "relay-assisted",
    "relay_session_id": "rs_766b63eae33fdb7b97e1ac1ca21ed26a",
    "is_online": true,
    "last_seen": "2025-09-19T00:30:56.616465773Z",
    "registered_at": "2025-09-18T23:51:04.305782317Z",
    "active_connections": 0,
    "mesh_status": "connected",
    "latency_ms": 0
  }
}
```

### 5. Закрытие сессии
```javascript
DELETE /api/v1/tenants/{tenant_id}/peers/{peer_id}

// Ответ
{
  "success": true,
  "message": "Peer session closed successfully"
}
```

## 🔄 Heartbeat система

### ⚠️ КРИТИЧЕСКИ ВАЖНО!
Heartbeat обязателен каждые 30 секунд. Без heartbeat пир будет помечен как offline.

### Пример реализации:
```javascript
class HeartbeatManager {
  constructor(serverUrl, tenantId, peerId, relaySessionId, jwtToken) {
    this.serverUrl = serverUrl;
    this.tenantId = tenantId;
    this.peerId = peerId;
    this.relaySessionId = relaySessionId;
    this.jwtToken = jwtToken;
    this.interval = null;
  }

  start() {
    this.interval = setInterval(() => {
      this.sendHeartbeat();
    }, 30000); // 30 секунд
  }

  stop() {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
  }

  async sendHeartbeat() {
    try {
      const response = await fetch(`${this.serverUrl}/api/v1/tenants/${this.tenantId}/peers/${this.peerId}/heartbeat`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          status: 'active',
          relay_session_id: this.relaySessionId
        })
      });

      const data = await response.json();
      if (data.success) {
        console.log('Heartbeat sent successfully');
      } else {
        console.error('Heartbeat failed:', data);
      }
    } catch (error) {
      console.error('Heartbeat error:', error);
    }
  }
}
```

## 🌐 P2P транспорт

### UDP порты для P2P:
- **9090/UDP** - QUIC Transport (основной)
- **9092/UDP** - Enhanced QUIC (расширенные функции)
- **19302/UDP** - STUN Server (NAT traversal)

### Пример QUIC соединения:
```javascript
// Для QUIC соединений используйте библиотеку quic-go или аналогичную
const quicConfig = {
  server: 'edge.2gc.ru',
  port: 9090,
  protocol: 'udp'
};
```

## 🏢 Мультитунельность

### Tenant изоляция:
- Каждый пользователь работает в своем tenant
- Пиры видят только других пиров в том же tenant
- JWT токен должен содержать правильный tenant_id

### Пример:
```javascript
// JWT токен должен содержать:
{
  "tenant_id": "tenant-216420165",
  "user_id": "user123",
  "exp": 1234567890
}
```

## 📝 Пример полной интеграции

```javascript
class CloudBridgeRelayClient {
  constructor(serverUrl, jwtToken) {
    this.serverUrl = serverUrl;
    this.jwtToken = jwtToken;
    this.tenantId = this.extractTenantId(jwtToken);
    this.peerId = null;
    this.relaySessionId = null;
    this.heartbeatManager = null;
  }

  extractTenantId(jwtToken) {
    // Декодируйте JWT токен и извлеките tenant_id
    const payload = JSON.parse(atob(jwtToken.split('.')[1]));
    return payload.tenant_id;
  }

  async registerPeer(deviceId, wireguardConfig) {
    const response = await fetch(`${this.serverUrl}/api/v1/tenants/${this.tenantId}/peers/register`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.jwtToken}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        device_id: deviceId,
        connection_type: 'relay-assisted',
        wireguard_config: wireguardConfig
      })
    });

    const data = await response.json();
    if (data.success) {
      this.peerId = data.peer_id;
      this.relaySessionId = data.relay_session_id;
      
      // Запускаем heartbeat
      this.heartbeatManager = new HeartbeatManager(
        this.serverUrl,
        this.tenantId,
        this.peerId,
        this.relaySessionId,
        this.jwtToken
      );
      this.heartbeatManager.start();
      
      return data;
    } else {
      throw new Error('Failed to register peer: ' + data.message);
    }
  }

  async discoverPeers() {
    const response = await fetch(`${this.serverUrl}/api/v1/tenants/${this.tenantId}/peers/discover`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${this.jwtToken}`,
        'Content-Type': 'application/json'
      }
    });

    const data = await response.json();
    if (data.success) {
      return data.peers;
    } else {
      throw new Error('Failed to discover peers: ' + data.message);
    }
  }

  async closeSession() {
    if (this.heartbeatManager) {
      this.heartbeatManager.stop();
    }

    if (this.peerId) {
      const response = await fetch(`${this.serverUrl}/api/v1/tenants/${this.tenantId}/peers/${this.peerId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${this.jwtToken}`,
          'Content-Type': 'application/json'
        }
      });

      const data = await response.json();
      return data.success;
    }
  }
}

// Использование:
const client = new CloudBridgeRelayClient('https://edge.2gc.ru:9444', jwtToken);

// Регистрация пира
const peerInfo = await client.registerPeer('my-device-123', {
  public_key: 'your-public-key',
  endpoint: 'your-endpoint',
  listen_port: 51820,
  allowed_ips: ['10.100.77.0/24']
});

// Обнаружение пиров
const peers = await client.discoverPeers();
console.log('Available peers:', peers);

// Закрытие сессии
await client.closeSession();
```

## 🔧 Отладка и мониторинг

### Health check:
```javascript
const healthResponse = await fetch(`${serverUrl}/health`);
const health = await healthResponse.json();
console.log('Server health:', health);
```

### Метрики:
```javascript
const metricsResponse = await fetch('http://edge.2gc.ru:9091/metrics');
const metrics = await metricsResponse.text();
console.log('Server metrics:', metrics);
```

## ⚠️ Важные изменения

### 1. Новый API endpoint
- **Было**: `http://edge.2gc.ru:8080`
- **Стало**: `http://edge.2gc.ru:8083` (HTTP) или `https://edge.2gc.ru:9444` (HTTPS)

### 2. Обязательный heartbeat
- Каждые 30 секунд
- Без heartbeat пир будет offline

### 3. Мультитунельность
- Используйте tenant_id из JWT токена
- Пиры изолированы по tenant

### 4. Новые поля в ответах
- `relay_session_id` - для heartbeat
- `peer_id` - уникальный идентификатор
- `tenant_id` - изоляция пользователей

## 🚀 Готово к использованию!

P2P маршрутизация CloudBridge Relay полностью функциональна и готова для интеграции с вашими клиентами!
