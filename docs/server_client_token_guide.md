# Руководство по получению токена для Server Client

## 🎯 Обзор

Server Client в CloudBridge Relay требует JWT токен для аутентификации при подключении к Relay Server. Этот документ описывает различные способы получения токена в зависимости от вашей конфигурации.

## 🔑 Источники токенов

### 1. **Keycloak** (рекомендуется для production)
- Централизованная система аутентификации
- Поддержка OAuth 2.0 / OpenID Connect
- Автоматическое обновление токенов
- Интеграция с LDAP/AD

### 2. **Django API** (планируется)
- Собственный API для управления токенами
- Интеграция с системой мониторинга
- Детальный аудит использования

### 3. **Локальная генерация** (для разработки)
- Быстрое создание тестовых токенов
- Независимость от внешних сервисов
- Простота настройки

## 🚀 Сценарии получения токена

### Сценарий 1: Через Keycloak Admin Console

#### Шаг 1: Вход в Keycloak Admin Console

```bash
# Откройте браузер и перейдите к
http://your-keycloak-server:8080

# Войдите с учетными данными администратора
Username: admin
Password: [пароль из .env файла]
```

#### Шаг 2: Создание пользователя для Server Client

1. Перейдите в раздел **Users**
2. Нажмите **Add User**
3. Заполните поля:
   - **Username**: `server-client-user`
   - **Email**: `server-client@company.com`
   - **Email Verified**: `ON`
4. Нажмите **Save**

#### Шаг 3: Настройка пароля

1. Перейдите на вкладку **Credentials**
2. Нажмите **Set Password**
3. Введите безопасный пароль
4. Отключите **Temporary** (если включено)
5. Нажмите **Set Password**

#### Шаг 4: Назначение ролей

1. Перейдите на вкладку **Role Mappings**
2. В разделе **Available Roles** найдите `relay-server-client`
3. Нажмите **Add selected** для назначения роли
4. Роль появится в разделе **Assigned Roles**

#### Шаг 5: Получение токена через API

```bash
# Получение access token
curl -X POST http://your-keycloak-server:8080/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=relay-client" \
  -d "client_secret=your-client-secret" \
  -d "username=server-client-user" \
  -d "password=your-password"
```

**Пример ответа:**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "bearer",
  "not-before-policy": 0,
  "session_state": "12345678-1234-1234-1234-123456789012",
  "scope": "openid email profile"
}
```

### Сценарий 2: Через Django API (планируется)

#### Шаг 1: Регистрация в Django Dashboard

```bash
# Откройте Django Dashboard
http://your-django-server:8000/admin

# Войдите с учетными данными администратора
```

#### Шаг 2: Создание API ключа

1. Перейдите в раздел **API Keys**
2. Нажмите **Create New API Key**
3. Заполните поля:
   - **Name**: `Server Client Production`
   - **Permissions**: `tunnel_create, tunnel_read`
   - **Expires**: `2024-12-31`
4. Нажмите **Save**

#### Шаг 3: Получение токена

```bash
# Генерация JWT токена через Django API
curl -X POST http://your-django-server:8000/api/auth/generate-token \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-admin-token" \
  -d '{
    "client_id": "server-client-001",
    "permissions": ["tunnel_create", "tunnel_read"],
    "expires_in": "24h"
  }'
```

### Сценарий 3: Локальная генерация (для разработки)

#### Вариант 1: Использование скрипта генерации

```bash
# Генерация тестового токена
./scripts/generate_test_token.sh \
  --client-id "test-server-client" \
  --secret "your-jwt-secret" \
  --expires "24h"
```

#### Вариант 2: Python скрипт

```python
import jwt
import datetime

# Создание JWT токена
payload = {
    'client_id': 'test-server-client',
    'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=24),
    'iat': datetime.datetime.utcnow(),
    'permissions': ['tunnel_create', 'tunnel_read']
}

token = jwt.encode(payload, 'your-jwt-secret', algorithm='HS256')
print(f"Generated token: {token}")
```

#### Вариант 3: Go скрипт

```go
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v4"
)

func main() {
    claims := jwt.MapClaims{
        "client_id":  "test-server-client",
        "exp":        time.Now().Add(24 * time.Hour).Unix(),
        "iat":        time.Now().Unix(),
        "permissions": []string{"tunnel_create", "tunnel_read"},
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte("your-jwt-secret"))
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Generated token: %s\n", tokenString)
}
```

## 🔧 Конфигурация Server Client с токеном

### 1. Конфигурационный файл

Создайте файл `config.json`:

```json
{
  "relay": {
    "host": "relay.example.com",
    "port": 8080,
    "tls": {
      "enabled": true,
      "cert_file": "/etc/relay/certs/client.crt",
      "key_file": "/etc/relay/certs/client.key"
    }
  },
  "auth": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_source": "keycloak",
    "refresh_enabled": true,
    "refresh_interval": "1h"
  },
  "tunnels": [
    {
      "id": "ldap-company-a",
      "local_port": 389,
      "remote_host": "ldap.company-a.local",
      "remote_port": 389,
      "protocol": "tcp"
    }
  ]
}
```

### 2. Переменные окружения

```bash
# Настройка токена через переменные окружения
export RELAY_AUTH_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
export RELAY_SERVER_HOST="relay.example.com"
export RELAY_SERVER_PORT="8080"
export RELAY_TLS_ENABLED="true"
export RELAY_TOKEN_SOURCE="keycloak"
export RELAY_REFRESH_ENABLED="true"
```

### 3. Запуск Server Client

```bash
# Запуск с конфигурационным файлом
./cloudbridge-client --config config.json

# Запуск с переменными окружения
./cloudbridge-client --token $RELAY_AUTH_TOKEN --host $RELAY_SERVER_HOST

# Запуск с параметрами командной строки
./cloudbridge-client \
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --host "relay.example.com" \
  --port 8080 \
  --tls-enabled
```

## 🔄 Обновление токенов

### Автоматическое обновление

Server Client поддерживает автоматическое обновление токенов:

```go
type TokenManager struct {
    currentToken string
    refreshToken  string
    expiresAt     time.Time
    refreshURL    string
}

func (tm *TokenManager) RefreshToken() error {
    resp, err := http.PostForm(tm.refreshURL, url.Values{
        "grant_type":    {"refresh_token"},
        "refresh_token": {tm.refreshToken},
        "client_id":     {"relay-client"},
    })
    
    if err != nil {
        return err
    }
    
    var tokenResponse struct {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        ExpiresIn    int    `json:"expires_in"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
        return err
    }
    
    tm.currentToken = tokenResponse.AccessToken
    tm.refreshToken = tokenResponse.RefreshToken
    tm.expiresAt = time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
    
    return nil
}
```

### Ручное обновление

```bash
# Обновление токена через Keycloak
curl -X POST http://keycloak:8080/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=refresh_token" \
  -d "client_id=relay-client" \
  -d "refresh_token=your-refresh-token"
```

## 🛡️ Безопасность токенов

### Хранение токенов

#### ✅ Рекомендуемые способы:

1. **Переменные окружения**:
   ```bash
   export RELAY_AUTH_TOKEN="your-token"
   ```

2. **Защищенные файлы конфигурации**:
   ```bash
   # Установите правильные права доступа
   chmod 600 config.json
   chown relay:relay config.json
   ```

3. **Менеджеры секретов**:
   ```bash
   # HashiCorp Vault
   vault kv get secret/relay/token
   
   # AWS Secrets Manager
   aws secretsmanager get-secret-value --secret-id relay/token
   
   # Azure Key Vault
   az keyvault secret show --vault-name myvault --name relay-token
   ```

#### ❌ НЕ рекомендуется:

- Хранить токены в коде
- Передавать токены в логах
- Использовать незащищенные файлы
- Делиться токенами через незащищенные каналы

### Ротация токенов

#### Рекомендации:

1. **Частота обновления**: каждые 24 часа для production
2. **Срок действия**: короткий для production токенов
3. **Автоматическая ротация**: через Keycloak или Django API
4. **Мониторинг**: отслеживание истечения токенов

#### Настройка автоматической ротации в Keycloak:

1. Перейдите в **Clients** → **relay-client**
2. Настройте **Access Token Lifespan**: 5 минут
3. Настройте **Client Session Idle**: 30 минут
4. Включите **Refresh Token Rotation**

### Мониторинг использования

#### Логирование:

```go
// Логирование использования токенов
log.WithFields(log.Fields{
    "client_id": clientID,
    "token_source": tokenSource,
    "expires_at": expiresAt,
}).Info("Token validated successfully")
```

#### Метрики:

```promql
# Количество валидаций токенов
relay_token_validations_total

# Процент успешных валидаций
rate(relay_token_validations_total{result="success"}[5m]) / 
rate(relay_token_validations_total[5m]) * 100

# Время валидации токенов
histogram_quantile(0.95, rate(relay_token_validation_duration_seconds_bucket[5m]))
```

#### Алерты:

```yaml
# prometheus/alerts/relay_alerts.yml
- alert: HighTokenValidationFailureRate
  expr: rate(relay_token_validations_total{result="failure"}[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High token validation failure rate"
    description: "Token validation failure rate is {{ $value }}"
```

## 🔍 Диагностика проблем

### Частые проблемы

#### 1. Неверный токен

**Симптомы:**
```
Error: invalid token
```

**Решение:**
- Проверьте правильность токена
- Убедитесь, что токен не истек
- Проверьте подпись токена

#### 2. Истекший токен

**Симптомы:**
```
Error: token expired
```

**Решение:**
- Обновите токен через refresh token
- Создайте новый токен
- Проверьте настройки времени на сервере

#### 3. Недостаточно прав

**Симптомы:**
```
Error: insufficient permissions
```

**Решение:**
- Проверьте роли пользователя в Keycloak
- Убедитесь, что токен содержит необходимые permissions
- Проверьте настройки клиента в Keycloak

### Команды диагностики

```bash
# Проверка токена через Django API
curl -X POST http://django-server:8000/api/validate-token \
  -H "Content-Type: application/json" \
  -d '{"token": "your-token"}'

# Проверка токена через Keycloak
curl -X GET http://keycloak:8080/realms/cloudbridge/protocol/openid-connect/userinfo \
  -H "Authorization: Bearer your-token"

# Декодирование JWT токена (без проверки подписи)
echo "your-token" | cut -d. -f2 | base64 -d | jq .

# Проверка подключения к Relay Server
telnet relay.example.com 8080
```

## 📋 Чек-лист настройки

### Для Keycloak:

- [ ] Создан пользователь для Server Client
- [ ] Назначена роль `relay-server-client`
- [ ] Установлен пароль
- [ ] Получен access token
- [ ] Сохранен refresh token
- [ ] Настроена автоматическая ротация токенов

### Для Server Client:

- [ ] Токен сохранен в конфигурации
- [ ] Настроен источник токена (keycloak/django/local)
- [ ] Включено автоматическое обновление
- [ ] Настроены туннели
- [ ] Проверено подключение к Relay Server

### Для безопасности:

- [ ] Токен хранится безопасно
- [ ] Настроена ротация токенов
- [ ] Включено логирование
- [ ] Настроены алерты
- [ ] Проведен аудит доступа

## 📞 Поддержка

Если у вас возникли проблемы с получением или использованием токенов:

1. Проверьте логи Server Client
2. Проверьте логи Relay Server
3. Проверьте логи Keycloak/Django
4. Используйте команды диагностики
5. Обратитесь к документации по устранению неполадок 