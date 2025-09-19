# CloudBridge Relay - Руководство для разработчика клиента (Cloudflare Style)

## 🔍 Анализ токенов Cloudflare

После изучения исходного кода cloudflared, мы понимаем, что Cloudflare использует **простую структуру токенов** вместо сложных JWT:

### Структура токена Cloudflare:
```json
{
  "a": "073ae0130b7cee5e55a1ac1a335431a8",  // Account ID
  "t": "36ae01eb-35c0-491c-befe-f2086c5611a0",  // Tunnel ID  
  "s": "HQh1O/moL9w8yHzt9Vp+rO7s9XME6YCdGqoNOgBxUFY="  // Secret/Key
}
```

## 🚀 Рекомендации для CloudBridge Relay

### 1. **Упростить структуру токенов**

Вместо сложных JWT с множеством полей, использовать простую структуру:

```json
{
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169", 
  "secret": "base64-encoded-secret-key"
}
```

### 2. **Безопасность токенов**

- **Короткий срок жизни**: 1-7 дней вместо 1 года
- **Простая структура**: только необходимые поля
- **Base64 кодирование**: как у Cloudflare
- **HMAC подпись**: для проверки целостности

### 3. **Обновленная реализация токенов**

```go
type CloudBridgeToken struct {
    TenantID  string `json:"tenant_id"`
    ServerID  string `json:"server_id"`
    Secret    string `json:"secret"`
    ExpiresAt int64  `json:"expires_at"`
    IssuedAt  int64  `json:"issued_at"`
}
```

## 📋 Обновленное руководство для разработчика

### **Основные endpoints:**

```javascript
// HTTP (разработка)
const serverUrl = 'http://edge.2gc.ru:8083';

// HTTPS (продакшен)  
const serverUrl = 'https://edge.2gc.ru:9444';
```

### **Упрощенная регистрация пира:**

```javascript
// Вместо сложного JWT, используем простой токен
const token = "eyJ0ZW5hbnRfaWQiOiJ0ZW5hbnQtMjE2NDIwMTY1Iiwic2VydmVyX2lkIjoic2VydmVyLWNsaWVudC1zZXJ2ZXItMTc1ODEwNTY4OTE2OSIsInNlY3JldCI6ImFiYzEyMyIsImV4cGlyZXNfYXQiOjE3NTgyNDI3NDB9";

// Регистрация пира
const response = await fetch(`${serverUrl}/api/v1/tenants/tenant-216420165/peers/register`, {
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

### **Heartbeat (обязательно каждые 30 секунд):**

```javascript
setInterval(async () => {
    await fetch(`${serverUrl}/api/v1/tenants/tenant-216420165/peers/${peerId}/heartbeat`, {
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
}, 30000); // 30 секунд
```

## 🔧 Преимущества упрощенного подхода

### **1. Простота реализации:**
- Меньше кода для парсинга
- Быстрее обработка
- Меньше ошибок

### **2. Безопасность:**
- Короткий срок жизни токенов
- Простая структура = меньше уязвимостей
- HMAC подпись для проверки

### **3. Совместимость:**
- Похоже на Cloudflare (знакомо разработчикам)
- Легко интегрировать
- Простое тестирование

## ⚠️ Критически важно

1. **Heartbeat обязателен** - без него пир будет offline
2. **Короткий срок жизни токенов** - 1-7 дней максимум
3. **Tenant изоляция** - каждый пользователь в своем tenant
4. **Простая структура токенов** - только необходимые поля

## 🎯 Следующие шаги

1. **Упростить генерацию токенов** - убрать сложные JWT поля
2. **Сократить срок жизни** - до 7 дней максимум
3. **Добавить HMAC подпись** - для проверки целостности
4. **Обновить документацию** - для разработчиков клиентов

**Вывод**: Cloudflare использует простые токены вместо сложных JWT. Это более безопасно и проще в реализации!
