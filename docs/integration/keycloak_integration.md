# Внедрение Multi-Tenancy (2025): что сделано и как интегрировано

В рамках обновления CloudBridge Relay:
- Собрано и интегрировано расширение [Keycloak Multi-Tenancy](https://github.com/anarsultanov/keycloak-multi-tenancy) (v26.1.0):
  - Клонирование репозитория, сборка JAR через Maven под JDK 17
  - Монтирование JAR в Keycloak через docker-compose (volume ./keycloak_providers)
- Настроено использование HTTPS с реальными сертификатами (монтирование /etc/relay/certs в контейнер)
- Keycloak теперь работает на 443 порту (внешний) с внутренним 8443
- Переменные окружения и команда запуска Keycloak обновлены для поддержки TLS
- После перезапуска Keycloak расширение Multi-Tenancy доступно в админке
- Включены Required Actions: Review tenant invitations, Create tenant, Select active tenant
- Проверена работа REST API Multi-Tenancy (создание, просмотр, управление арендаторами, приглашениями, членствами)
- Документация и практические шаги по установке добавлены в [tech_task_keycloak_multitenancy.md](../reports/tech_task_keycloak_multitenancy.md)

---

# Интеграция CloudBridge Relay с Keycloak

## Обзор

CloudBridge Relay поддерживает интеграцию с Keycloak для аутентификации и авторизации пользователей. Эта интеграция обеспечивает:

- **Централизованную аутентификацию** через Keycloak
- **Поддержку OAuth 2.0 / OpenID Connect**
- **Управление ролями и группами**
- **Единый вход (SSO)**
- **Безопасную валидацию JWT токенов**

## Архитектура интеграции

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │   Keycloak  │    │   Relay     │
│             │    │             │    │             │
│ 1. Login    │───▶│ 2. Auth     │    │ 3. Validate │
│             │    │ 3. Issue    │───▶│ 4. Extract  │
│ 5. Use JWT  │◀───│    JWT      │    │ 5. Check    │
│             │    │             │    │ Permissions │
└─────────────┘    └─────────────┘    └─────────────┘
```

## Настройка Keycloak

### 1. Создание Realm

1. Войдите в Keycloak Admin Console
2. Создайте новый realm `cloudbridge`
3. Настройте realm settings:
   - **Realm name**: `cloudbridge`
   - **Display name**: `CloudBridge`
   - **Enabled**: `ON`

### 2. Создание Client

1. Перейдите в **Clients** → **Create**
2. Настройте client:
   ```
   Client ID: relay-client
   Client Protocol: openid-connect
   Root URL: https://your-relay-server.com
   ```

3. Настройте **Settings**:
   ```
   Access Type: confidential
   Valid Redirect URIs: https://your-relay-server.com/*
   Web Origins: https://your-relay-server.com
   ```

4. Настройте **Credentials**:
   - Скопируйте **Client Secret** для конфигурации

### 3. Создание Roles

1. Перейдите в **Roles** → **Add Role**
2. Создайте роли:
   ```
   relay-admin: Полный доступ к relay
   relay-user: Базовый доступ к relay
   relay-readonly: Только чтение
   ```

### 4. Создание Groups

1. Перейдите в **Groups** → **New**
2. Создайте группы:
   ```
   relay-admins: Администраторы
   relay-users: Обычные пользователи
   relay-monitors: Мониторинг
   ```

### 5. Настройка Users

1. Создайте пользователей в **Users** → **Add User**
2. Назначьте роли и группы
3. Настройте **Credentials** (пароли)

## Конфигурация Relay

### 1. Включение Keycloak

Обновите `config/config.json`:

```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "https://keycloak.example.com",
    "realm": "cloudbridge",
    "client_id": "relay-client",
    "client_secret": "your-client-secret",
    "public_key": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
    "token_validation": {
      "issuer": "https://keycloak.example.com/realms/cloudbridge",
      "audience": "relay-client",
      "clock_skew": "30s"
    },
    "jwks_cache_ttl": "1h",
    "http_timeout": "30s"
  }
}
```

### 2. Переменные окружения

```bash
# Keycloak configuration
export KEYCLOAK_SERVER_URL="https://keycloak.example.com"
export KEYCLOAK_REALM="cloudbridge"
export KEYCLOAK_CLIENT_ID="relay-client"
export KEYCLOAK_CLIENT_SECRET="your-client-secret"
export KEYCLOAK_PUBLIC_KEY="MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA..."

# Enable Keycloak
export RELAY_AUTH_PROVIDER="keycloak"
```

### 3. Получение Public Key

Получите public key из Keycloak:

```bash
curl -s https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/certs | jq -r '.keys[0].n'
```

## Использование

### 1. Получение токена

```bash
# Получение access token
curl -X POST https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=relay-client" \
  -d "client_secret=your-client-secret" \
  -d "username=user@example.com" \
  -d "password=password"
```

### 2. Подключение к Relay

```python
import requests

# Получение токена
token_response = requests.post(
    "https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/token",
    data={
        "grant_type": "password",
        "client_id": "relay-client",
        "client_secret": "your-client-secret",
        "username": "user@example.com",
        "password": "password"
    }
)

access_token = token_response.json()["access_token"]

# Подключение к relay с токеном
headers = {"Authorization": f"Bearer {access_token}"}
response = requests.get("https://relay.example.com/api/v1/health", headers=headers)
```

### 3. Проверка разрешений

```python
# Проверка ролей пользователя
def check_user_permissions(token):
    # Декодирование токена (без валидации)
    import jwt
    decoded = jwt.decode(token, options={"verify_signature": False})
    
    # Проверка ролей
    roles = decoded.get("realm_access", {}).get("roles", [])
    if "relay-admin" in roles:
        return "admin"
    elif "relay-user" in roles:
        return "user"
    else:
        return "none"
```

## Безопасность

### 1. Валидация токенов

Relay автоматически валидирует JWT токены:

- **Подпись**: Проверяется с помощью public key из Keycloak
- **Время жизни**: Проверяется exp, iat, nbf claims
- **Issuer**: Проверяется соответствие realm
- **Audience**: Проверяется client_id

### 2. Кэширование JWKS

- JWKS (JSON Web Key Set) кэшируется на 1 час
- Автоматическое обновление при истечении
- Fallback на статический public key

### 3. Rate Limiting

```json
{
  "security": {
    "rate_limit": {
      "enabled": true,
      "requests_per_second": 100,
      "burst": 200
    }
  }
}
```

## Мониторинг

### 1. Метрики аутентификации

```prometheus
# Метрики Keycloak аутентификации
relay_keycloak_token_validations_total{result="success"}
relay_keycloak_token_validations_total{result="failure"}
relay_keycloak_jwks_cache_hits_total
relay_keycloak_jwks_cache_misses_total
relay_keycloak_token_validation_duration_seconds
```

### 2. Логирование

```json
{
  "timestamp": "2024-03-14T12:00:00Z",
  "level": "INFO",
  "message": "Token validated successfully",
  "fields": {
    "provider": "keycloak",
    "user_id": "user-123",
    "username": "john.doe",
    "roles": ["relay-user"],
    "groups": ["relay-users"]
  }
}
```

## Troubleshooting

### 1. Ошибки валидации токенов

```bash
# Проверка конфигурации Keycloak
curl -s https://keycloak.example.com/realms/cloudbridge/.well-known/openid_configuration

# Проверка JWKS
curl -s https://keycloak.example.com/realms/cloudbridge/protocol/openid-connect/certs
```

### 2. Проблемы с подключением

```bash
# Проверка доступности Keycloak
curl -I https://keycloak.example.com/realms/cloudbridge

# Проверка SSL сертификатов
openssl s_client -connect keycloak.example.com:443 -servername keycloak.example.com
```

### 3. Отладка аутентификации

```bash
# Включение debug логирования
./relay --log-level DEBUG --config config/config.json

# Просмотр логов аутентификации
tail -f /var/log/relay/relay.log | grep -i keycloak
```

## Миграция с Django

### 1. Поэтапная миграция

1. **Этап 1**: Настройка Keycloak параллельно с Django
2. **Этап 2**: Тестирование с ограниченной группой пользователей
3. **Этап 3**: Полная миграция на Keycloak
4. **Этап 4**: Отключение Django аутентификации

### 2. Совместимость

```json
{
  "keycloak": {
    "enabled": true
  },
  "django": {
    "enabled": false,
    "fallback": true
  }
}
```

### 3. Маппинг ролей

```yaml
# Маппинг ролей Django → Keycloak
role_mapping:
  django_admin: relay-admin
  django_user: relay-user
  django_readonly: relay-readonly
```

## Примеры конфигураций

### 1. Development

```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "http://localhost:8080",
    "realm": "cloudbridge-dev",
    "client_id": "relay-client-dev",
    "client_secret": "dev-secret",
    "token_validation": {
      "issuer": "http://localhost:8080/realms/cloudbridge-dev",
      "audience": "relay-client-dev",
      "clock_skew": "60s"
    }
  }
}
```

### 2. Production

```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "https://keycloak.company.com",
    "realm": "cloudbridge-prod",
    "client_id": "relay-client-prod",
    "client_secret": "REPLACE_WITH_SECURE_SECRET",
    "token_validation": {
      "issuer": "https://keycloak.company.com/realms/cloudbridge-prod",
      "audience": "relay-client-prod",
      "clock_skew": "30s"
    },
    "jwks_cache_ttl": "1h",
    "http_timeout": "30s"
  }
}
```

### 3. High Availability

```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "https://keycloak-lb.company.com",
    "realm": "cloudbridge-ha",
    "client_id": "relay-client-ha",
    "client_secret": "REPLACE_WITH_SECURE_SECRET",
    "token_validation": {
      "issuer": "https://keycloak-lb.company.com/realms/cloudbridge-ha",
      "audience": "relay-client-ha",
      "clock_skew": "30s"
    },
    "jwks_cache_ttl": "2h",
    "http_timeout": "60s"
  }
}
```

## Заключение

Интеграция с Keycloak обеспечивает:

- **Централизованное управление пользователями**
- **Высокий уровень безопасности**
- **Масштабируемость**
- **Стандартизированную аутентификацию**

Для получения дополнительной помощи обратитесь к:
- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [OpenID Connect Specification](https://openid.net/specs/openid-connect-core-1_0.html)
- [JWT RFC](https://tools.ietf.org/html/rfc7519) 