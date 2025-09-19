# 🔧 **CloudBridge Relay - Исправления для разработчика личного кабинета**

## 🚨 **КРИТИЧЕСКАЯ ПРОБЛЕМА НАЙДЕНА**

Токены, генерируемые личным кабинетом, имеют **неправильный формат** и не проходят валидацию на сервере.

## 🔍 **Анализ проблемы:**

### **Токен из личного кабинета (НЕПРАВИЛЬНЫЙ):**
```
eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0Njc3NDUyMSIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJtZXNoX2pvaW4iLCJtZXNoX21hbmFnZSJdLCJpYXQiOjE3NTgyNDY3NzYsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci0xNzU4MjQ2Nzc0NTIxIiwianRpIjoiand0XzE3NTgyNDY3NzZfbXVlMzZpajcifQ.0qn89U3hgOhs01JY014MDTRQa_9Gm4Xc0epfX8Gb1xk
```

**Проблемы:**
1. ❌ **Отсутствует JWT заголовок** - нет `{"alg":"HS256","typ":"JWT"}`
2. ❌ **Неправильные разрешения** - `["mesh_join","mesh_manage"]` вместо `["p2p_connect","mesh_join"]`
3. ❌ **Отсутствует обязательное разрешение** - нет `p2p_connect`

### **Правильный JWT токен:**
```
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0Njc3NDUyMSIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJwMnBfY29ubmVjdCIsIm1lc2hfam9pbiJdLCJpYXQiOjE3NTgyNDY3NzYsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci0xNzU4MjQ2Nzc0NTIxIiwianRpIjoiand0XzE3NTgyNDY3NzZfbXVlMzZpajcifQ.UpKN07jJE7LhNnPSjdgm39u2LOGnCNMRYIKbr-3N3uU
```

## 🛠️ **Исправления для личного кабинета:**

### **1. Обновить генерацию JWT токенов:**

```python
import jwt
import time
import uuid

def generate_cloudbridge_token(tenant_id, server_id, permissions=None, expires_hours=24):
    """
    Генерирует ПРАВИЛЬНЫЙ JWT токен для CloudBridge Relay
    """
    
    # Секрет для подписи (должен совпадать с сервером)
    secret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
    
    # ИСПРАВЛЕНИЕ: Обязательные разрешения
    if permissions is None:
        permissions = ["p2p_connect", "mesh_join"]  # p2p_connect ОБЯЗАТЕЛЬНО!
    
    # Проверяем наличие обязательного разрешения
    if "p2p_connect" not in permissions:
        permissions.append("p2p_connect")
    
    # Текущее время
    now = int(time.time())
    
    # Уникальный ID токена
    jti = f"jwt_{now}_{uuid.uuid4().hex[:8]}"
    
    # Payload токена
    payload = {
        "protocol_type": "p2p-mesh",
        "scope": "p2p-mesh-claims",
        "tenant_id": tenant_id,
        "server_id": server_id,
        "connection_type": "quic",
        "permissions": permissions,
        "iat": now,
        "iss": "https://auth.2gc.ru/realms/cloudbridge",
        "aud": "account",
        "sub": server_id,
        "jti": jti
    }
    
    # Добавляем время истечения
    if expires_hours > 0:
        payload["exp"] = now + (expires_hours * 3600)
    
    # ИСПРАВЛЕНИЕ: Генерируем ПРАВИЛЬНЫЙ JWT токен с заголовком
    token = jwt.encode(payload, secret, algorithm="HS256")
    
    return token
```

### **2. Обновить UI формы:**

```html
<div class="form-group">
    <label>Разрешения:</label>
    <div class="checkbox-group">
        <label>
            <input type="checkbox" name="permissions" value="p2p_connect" checked disabled>
            P2P подключение (обязательно)
        </label>
        <label>
            <input type="checkbox" name="permissions" value="mesh_join" checked>
            Присоединение к mesh
        </label>
        <label>
            <input type="checkbox" name="permissions" value="relay_access">
            Доступ к relay
        </label>
        <label>
            <input type="checkbox" name="permissions" value="admin_access">
            Административные права
        </label>
    </div>
</div>
```

### **3. Добавить валидацию:**

```javascript
function validateTokenRequest(data) {
    // Проверяем обязательные поля
    const requiredFields = ['tenant_id', 'server_id', 'permissions'];
    
    for (const field of requiredFields) {
        if (!data[field]) {
            throw new Error(`Missing required field: ${field}`);
        }
    }
    
    // Проверяем обязательное разрешение
    if (!data.permissions.includes('p2p_connect')) {
        data.permissions.push('p2p_connect');
    }
    
    return data;
}
```

## 🧪 **Тестирование исправлений:**

### **Правильный токен (работает):**
```bash
# Регистрация пира
curl -X POST "http://localhost:8086/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0Njc3NDUyMSIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJwMnBfY29ubmVjdCIsIm1lc2hfam9pbiJdLCJpYXQiOjE3NTgyNDY3NzYsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci0xNzU4MjQ2Nzc0NTIxIiwianRpIjoiand0XzE3NTgyNDY3NzZfbXVlMzZpajcifQ.UpKN07jJE7LhNnPSjdgm39u2LOGnCNMRYIKbr-3N3uU" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "test-peer", "allowed_ips": ["10.0.0.0/24"]}'

# Ответ: {"success":true,"peer_id":"peer_server-1758246774521",...}
```

### **Неправильный токен (не работает):**
```bash
# Токен из личного кабинета
curl -X POST "http://localhost:8086/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI0Njc3NDUyMSIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJtZXNoX2pvaW4iLCJtZXNoX21hbmFnZSJdLCJpYXQiOjE3NTgyNDY3NzYsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci0xNzU4MjQ2Nzc0NTIxIiwianRpIjoiand0XzE3NTgyNDY3NzZfbXVlMzZpajcifQ.0qn89U3hgOhs01JY014MDTRQa_9Gm4Xc0epfX8Gb1xk" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "test-peer", "allowed_ips": ["10.0.0.0/24"]}'

# Ответ: {"error":"Invalid token","code":401,"message":"token is malformed: token contains an invalid number of segments"}
```

## 📋 **Чек-лист исправлений:**

### **Обязательные изменения:**

- [ ] **Исправить генерацию JWT** - добавить правильный заголовок
- [ ] **Добавить p2p_connect** - обязательное разрешение
- [ ] **Убрать mesh_manage** - заменить на mesh_join
- [ ] **Обновить UI** - сделать p2p_connect обязательным
- [ ] **Добавить валидацию** - проверять формат токена
- [ ] **Протестировать** - убедиться, что токены работают

### **Код для тестирования:**

```python
# Тест правильного токена
def test_correct_token():
    token = generate_cloudbridge_token(
        tenant_id="tenant-216420165",
        server_id="server-1758246774521",
        permissions=["p2p_connect", "mesh_join"]
    )
    
    # Проверяем структуру
    import jwt
    header = jwt.get_unverified_header(token)
    payload = jwt.decode(token, options={'verify_signature': False})
    
    assert header['alg'] == 'HS256'
    assert header['typ'] == 'JWT'
    assert payload['permissions'] == ['p2p_connect', 'mesh_join']
    assert 'p2p_connect' in payload['permissions']
    
    print("✅ Токен корректен!")
    return token

# Запуск теста
test_correct_token()
```

## 🎯 **Результат:**

После исправлений личный кабинет будет генерировать **правильные JWT токены**, которые:
- ✅ Проходят валидацию на сервере
- ✅ Содержат обязательное разрешение `p2p_connect`
- ✅ Имеют правильный формат `header.payload.signature`
- ✅ Работают с API регистрации и discovery

**Исправления критически важны для работы системы! 🚨**
