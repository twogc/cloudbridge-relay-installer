# 🔐 **CloudBridge Relay - Руководство для разработчика личного кабинета**

## 📋 **Обзор изменений**

CloudBridge Relay перешел на новую систему токенов. Удалена поддержка WireGuard, теперь используется единый стандарт JWT токенов для P2P mesh сети с QUIC протоколом.

## 🔑 **Новый формат JWT токенов**

### **Структура токена:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims", 
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169",
  "connection_type": "quic",
  "permissions": ["p2p_connect", "mesh_join"],
  "iat": 1758245498,
  "iss": "https://auth.2gc.ru/realms/cloudbridge",
  "aud": "account",
  "sub": "server-client-server-1758105689169",
  "jti": "jwt_1758245498_unified"
}
```

### **Обязательные поля:**

| Поле | Тип | Обязательное | Описание | Пример |
|------|-----|--------------|----------|---------|
| `protocol_type` | string | ✅ | Тип протокола | `"p2p-mesh"` |
| `scope` | string | ✅ | Область действия | `"p2p-mesh-claims"` |
| `tenant_id` | string | ✅ | ID организации | `"tenant-216420165"` |
| `server_id` | string | ✅ | Уникальный ID сервера | `"server-client-server-1758105689169"` |
| `connection_type` | string | ✅ | Тип соединения | `"quic"` |
| `permissions` | array | ✅ | Разрешения | `["p2p_connect", "mesh_join"]` |
| `iat` | number | ✅ | Время выдачи | `1758245498` |
| `iss` | string | ✅ | Издатель | `"https://auth.2gc.ru/realms/cloudbridge"` |
| `aud` | string | ✅ | Аудитория | `"account"` |
| `sub` | string | ✅ | Субъект | `"server-client-server-1758105689169"` |
| `jti` | string | ✅ | Уникальный ID токена | `"jwt_1758245498_unified"` |

### **Доступные разрешения:**
- `p2p_connect` - Подключение к P2P сети (обязательное)
- `mesh_join` - Присоединение к mesh сети
- `relay_access` - Доступ к relay серверу
- `admin_access` - Административные права

## 🏗️ **Генерация токенов**

### **1. Python пример:**

```python
import jwt
import time
import uuid
from datetime import datetime, timedelta

def generate_cloudbridge_token(tenant_id, server_id, permissions=None, expires_hours=24):
    """
    Генерирует JWT токен для CloudBridge Relay
    
    Args:
        tenant_id (str): ID организации
        server_id (str): Уникальный ID сервера
        permissions (list): Список разрешений
        expires_hours (int): Время жизни токена в часах
    
    Returns:
        str: JWT токен
    """
    
    # Секрет для подписи (должен совпадать с сервером)
    secret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
    
    # Разрешения по умолчанию
    if permissions is None:
        permissions = ["p2p_connect", "mesh_join"]
    
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
    
    # Генерируем токен
    token = jwt.encode(payload, secret, algorithm="HS256")
    
    return token

# Пример использования
tenant_id = "tenant-216420165"
server_id = "server-client-server-1758105689169"
permissions = ["p2p_connect", "mesh_join", "relay_access"]

token = generate_cloudbridge_token(tenant_id, server_id, permissions, expires_hours=24)
print(f"Generated token: {token}")
```

### **2. Node.js пример:**

```javascript
const jwt = require('jsonwebtoken');
const crypto = require('crypto');

function generateCloudBridgeToken(tenantId, serverId, permissions = null, expiresHours = 24) {
    // Секрет для подписи (должен совпадать с сервером)
    const secret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=";
    
    // Разрешения по умолчанию
    if (!permissions) {
        permissions = ["p2p_connect", "mesh_join"];
    }
    
    // Текущее время
    const now = Math.floor(Date.now() / 1000);
    
    // Уникальный ID токена
    const jti = `jwt_${now}_${crypto.randomBytes(4).toString('hex')}`;
    
    // Payload токена
    const payload = {
        protocol_type: "p2p-mesh",
        scope: "p2p-mesh-claims",
        tenant_id: tenantId,
        server_id: serverId,
        connection_type: "quic",
        permissions: permissions,
        iat: now,
        iss: "https://auth.2gc.ru/realms/cloudbridge",
        aud: "account",
        sub: serverId,
        jti: jti
    };
    
    // Добавляем время истечения
    if (expiresHours > 0) {
        payload.exp = now + (expiresHours * 3600);
    }
    
    // Генерируем токен
    const token = jwt.sign(payload, secret, { algorithm: 'HS256' });
    
    return token;
}

// Пример использования
const tenantId = "tenant-216420165";
const serverId = "server-client-server-1758105689169";
const permissions = ["p2p_connect", "mesh_join", "relay_access"];

const token = generateCloudBridgeToken(tenantId, serverId, permissions, 24);
console.log(`Generated token: ${token}`);
```

### **3. PHP пример:**

```php
<?php
require_once 'vendor/autoload.php';

use Firebase\JWT\JWT;
use Firebase\JWT\Key;

function generateCloudBridgeToken($tenantId, $serverId, $permissions = null, $expiresHours = 24) {
    // Секрет для подписи (должен совпадать с сервером)
    $secret = "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=";
    
    // Разрешения по умолчанию
    if ($permissions === null) {
        $permissions = ["p2p_connect", "mesh_join"];
    }
    
    // Текущее время
    $now = time();
    
    // Уникальный ID токена
    $jti = "jwt_" . $now . "_" . bin2hex(random_bytes(4));
    
    // Payload токена
    $payload = [
        "protocol_type" => "p2p-mesh",
        "scope" => "p2p-mesh-claims",
        "tenant_id" => $tenantId,
        "server_id" => $serverId,
        "connection_type" => "quic",
        "permissions" => $permissions,
        "iat" => $now,
        "iss" => "https://auth.2gc.ru/realms/cloudbridge",
        "aud" => "account",
        "sub" => $serverId,
        "jti" => $jti
    ];
    
    // Добавляем время истечения
    if ($expiresHours > 0) {
        $payload["exp"] = $now + ($expiresHours * 3600);
    }
    
    // Генерируем токен
    $token = JWT::encode($payload, $secret, 'HS256');
    
    return $token;
}

// Пример использования
$tenantId = "tenant-216420165";
$serverId = "server-client-server-1758105689169";
$permissions = ["p2p_connect", "mesh_join", "relay_access"];

$token = generateCloudBridgeToken($tenantId, $serverId, $permissions, 24);
echo "Generated token: " . $token . "\n";
?>
```

## 🗄️ **База данных**

### **Структура таблиц для токенов:**

```sql
-- Таблица организаций
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    network_cidr VARCHAR(18) NOT NULL,
    max_peers INT DEFAULT 100,
    max_connections INT DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Таблица серверов
CREATE TABLE servers (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    ip_address VARCHAR(45),
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE
);

-- Таблица токенов
CREATE TABLE tokens (
    id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    server_id VARCHAR(255) NOT NULL,
    jti VARCHAR(255) UNIQUE NOT NULL,
    permissions JSON NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE,
    FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

-- Индексы
CREATE INDEX idx_tokens_tenant_id ON tokens(tenant_id);
CREATE INDEX idx_tokens_server_id ON tokens(server_id);
CREATE INDEX idx_tokens_jti ON tokens(jti);
CREATE INDEX idx_tokens_expires_at ON tokens(expires_at);
```

## 🔧 **API для управления токенами**

### **1. Создание токена:**

```http
POST /api/v1/tokens
Authorization: Bearer <admin_token>
Content-Type: application/json

{
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169",
  "permissions": ["p2p_connect", "mesh_join"],
  "expires_hours": 24
}
```

**Ответ:**
```json
{
  "success": true,
  "token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9...",
  "jti": "jwt_1758245498_abc123",
  "expires_at": "2025-09-20T01:31:43.629456228Z",
  "created_at": "2025-09-19T01:31:43.629456228Z"
}
```

### **2. Список токенов:**

```http
GET /api/v1/tenants/{tenant_id}/tokens
Authorization: Bearer <admin_token>
```

**Ответ:**
```json
{
  "success": true,
  "tokens": [
    {
      "id": "token_123",
      "server_id": "server-client-server-1758105689169",
      "jti": "jwt_1758245498_abc123",
      "permissions": ["p2p_connect", "mesh_join"],
      "expires_at": "2025-09-20T01:31:43.629456228Z",
      "created_at": "2025-09-19T01:31:43.629456228Z",
      "status": "active"
    }
  ]
}
```

### **3. Отзыв токена:**

```http
DELETE /api/v1/tokens/{jti}
Authorization: Bearer <admin_token>
```

**Ответ:**
```json
{
  "success": true,
  "message": "Token revoked successfully"
}
```

## 🎨 **UI компоненты**

### **1. Форма создания токена:**

```html
<form id="token-form">
  <div class="form-group">
    <label for="tenant-select">Организация:</label>
    <select id="tenant-select" name="tenant_id" required>
      <option value="">Выберите организацию</option>
      <option value="tenant-216420165">Организация A</option>
      <option value="tenant-123456789">Организация B</option>
    </select>
  </div>
  
  <div class="form-group">
    <label for="server-select">Сервер:</label>
    <select id="server-select" name="server_id" required>
      <option value="">Выберите сервер</option>
    </select>
  </div>
  
  <div class="form-group">
    <label>Разрешения:</label>
    <div class="checkbox-group">
      <label>
        <input type="checkbox" name="permissions" value="p2p_connect" checked>
        P2P подключение
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
  
  <div class="form-group">
    <label for="expires-hours">Время жизни (часы):</label>
    <input type="number" id="expires-hours" name="expires_hours" value="24" min="1" max="8760">
  </div>
  
  <button type="submit">Создать токен</button>
</form>
```

### **2. JavaScript для формы:**

```javascript
document.getElementById('token-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    const data = {
        tenant_id: formData.get('tenant_id'),
        server_id: formData.get('server_id'),
        permissions: formData.getAll('permissions'),
        expires_hours: parseInt(formData.get('expires_hours'))
    };
    
    try {
        const response = await fetch('/api/v1/tokens', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + getAdminToken()
            },
            body: JSON.stringify(data)
        });
        
        const result = await response.json();
        
        if (result.success) {
            // Показать токен пользователю
            showTokenModal(result.token);
        } else {
            showError('Ошибка создания токена: ' + result.message);
        }
    } catch (error) {
        showError('Ошибка сети: ' + error.message);
    }
});

function showTokenModal(token) {
    const modal = document.createElement('div');
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <h3>Токен создан успешно</h3>
            <p>Скопируйте токен и сохраните его в безопасном месте:</p>
            <textarea readonly class="token-display">${token}</textarea>
            <button onclick="copyToClipboard('${token}')">Копировать</button>
            <button onclick="closeModal()">Закрыть</button>
        </div>
    `;
    document.body.appendChild(modal);
}

function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(() => {
        alert('Токен скопирован в буфер обмена');
    });
}
```

## 🔒 **Безопасность**

### **Рекомендации по безопасности:**

1. **Хранение секрета:**
   ```bash
   # Используйте переменные окружения
   export CLOUDBRIDGE_JWT_SECRET="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
   ```

2. **Время жизни токенов:**
   - **Короткие токены:** 1-24 часа для обычных операций
   - **Долгие токены:** 30-90 дней для автоматизации
   - **Никогда не создавайте бессрочные токены**

3. **Отзыв токенов:**
   - Реализуйте механизм отзыва токенов
   - Ведите лог всех операций с токенами
   - Уведомляйте пользователей об истечении токенов

4. **Валидация:**
   ```python
   def validate_token_request(data):
       required_fields = ['tenant_id', 'server_id', 'permissions']
       
       for field in required_fields:
           if field not in data:
               raise ValueError(f"Missing required field: {field}")
       
       if not isinstance(data['permissions'], list):
           raise ValueError("Permissions must be a list")
       
       if 'p2p_connect' not in data['permissions']:
           raise ValueError("p2p_connect permission is required")
   ```

## 📋 **Чек-лист для миграции**

### **Что нужно изменить в личном кабинете:**

- [ ] **Удалить WireGuard поля** - больше не нужны
- [ ] **Обновить структуру токенов** - использовать новый формат
- [ ] **Изменить генерацию токенов** - использовать новый алгоритм
- [ ] **Обновить UI** - убрать WireGuard настройки
- [ ] **Добавить новые разрешения** - p2p_connect, mesh_join
- [ ] **Обновить валидацию** - проверять новые поля
- [ ] **Изменить API endpoints** - использовать новые маршруты

### **Новые зависимости:**
```json
// package.json
{
  "dependencies": {
    "jsonwebtoken": "^9.0.0"
  }
}
```

## 🚀 **Готово к использованию!**

Новая система токенов готова к использованию. Все компоненты протестированы и работают с CloudBridge Relay. Личный кабинет может генерировать токены в новом формате для P2P mesh сети.

**Для тестирования используйте секрет: `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`**
