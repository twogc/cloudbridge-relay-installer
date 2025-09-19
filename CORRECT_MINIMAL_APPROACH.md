# CloudBridge Relay - Правильный минимальный подход к токенам

## 🎯 **Правильная архитектура (согласно DEVELOPER_BRIEFING.md)**

### **Цель проекта:**
> **P2P mesh сеть для создания "почти LAN" соединений между удаленными серверами клиентов через централизованный relay сервер.**

### **Архитектура:**
```
Client A (Office A) ←→ QUIC:9090 ←→ CloudBridge Relay ←→ QUIC:9090 ←→ Client B (Office B)
```

## ✅ **Минимальный токен (только необходимое)**

### **Структура токена:**
```json
{
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169",
  "connection_type": "quic",
  "permissions": ["p2p_connect", "mesh_join"],
  "issued_at": 1758243476
}
```

### **Что убрали (не нужно для P2P mesh):**
- ❌ `network_config` - клиент не настраивает сеть
- ❌ `quic_config` - клиент знает, как подключаться
- ❌ `ice_config` - клиент знает STUN серверы
- ❌ `mesh_config` - не нужен для простого P2P
- ❌ `subnet` - не нужен для P2P mesh
- ❌ `registry_url` - у нас нет отдельного registry

## 🚀 **Что клиент знает (hardcoded)**

### **Endpoints:**
```javascript
// P2P API
const p2pApiUrl = 'https://edge.2gc.ru:9444';

// QUIC Server  
const quicServer = 'edge.2gc.ru:9090';

// STUN Server
const stunServer = 'edge.2gc.ru:19302';
```

### **Конфигурация:**
```javascript
// Heartbeat каждые 30 секунд
const heartbeatInterval = 30000;

// Тип соединения
const connectionType = 'quic';
```

## 🔐 **Безопасность и контроль**

### **1. Keycloak интеграция:**
- ✅ **Relay проверяет токен через Keycloak**
- ✅ **Keycloak контролирует активность пользователя**
- ✅ **Можно отозвать токен через Keycloak**
- ✅ **Контроль одновременных подключений**

### **2. HMAC подпись:**
- ✅ **Проверка целостности токена**
- ✅ **Защита от подделки**

## 📋 **Использование в клиенте**

### **1. Регистрация пира:**
```javascript
const token = "eyJ0ZW5hbnRfaWQiOiJ0ZW5hbnQtMjE2NDIwMTY1...";

const response = await fetch(`${p2pApiUrl}/api/v1/tenants/tenant-216420165/peers/register`, {
    method: 'POST',
    headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        public_key: "peer-public-key-here",
        allowed_ips: ["10.0.0.0/24"]
    })
});
```

### **2. Heartbeat (каждые 30 секунд):**
```javascript
setInterval(async () => {
    await fetch(`${p2pApiUrl}/api/v1/tenants/tenant-216420165/peers/${peerId}/heartbeat`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            status: "active",
            relay_session_id: relaySessionId
        })
    });
}, 30000);
```

### **3. P2P соединение:**
```javascript
// Клиент знает, как подключаться к QUIC серверу
const quicConnection = new QuicConnection('edge.2gc.ru:9090');
```

## 🎯 **Преимущества минимального подхода**

### **1. Простота:**
- ✅ **Минимальная структура токена**
- ✅ **Клиент знает, как подключаться**
- ✅ **Меньше ошибок в реализации**

### **2. Безопасность:**
- ✅ **Постоянный токен (без срока жизни)**
- ✅ **HMAC подпись для проверки**
- ✅ **Контроль через Keycloak**

### **3. Производительность:**
- ✅ **Быстрая валидация токенов**
- ✅ **Меньше данных для передачи**
- ✅ **Простая обработка**

## 🚀 **Итоговая архитектура**

### **1. Генерация токена:**
```bash
go run generate-minimal-token.go
```

### **2. Проверка в relay:**
```go
// 1. Проверить HMAC подпись
// 2. Отправить в Keycloak для валидации
// 3. Проверить, не заблокирован ли пользователь
// 4. Разрешить/запретить подключение
```

### **3. P2P маршрутизация:**
```go
// Relay пересылает QUIC пакеты между клиентами
// Изоляция по tenant_id
// Heartbeat для поддержания соединений
```

## ✅ **Готово к использованию**

**Минимальный токен содержит только необходимую информацию для P2P mesh сети, при этом обеспечивая безопасность через Keycloak и HMAC подпись!**

**Клиент знает, как подключаться к CloudBridge Relay, а токен содержит только идентификацию и права доступа.**
