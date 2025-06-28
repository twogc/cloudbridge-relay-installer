# Техническое задание: Настройка Keycloak для CloudBridge Relay

## 1. Общие сведения

### 1.1 Назначение документа
Данное техническое задание описывает требования к настройке Keycloak для интеграции с CloudBridge Relay, включая создание realm, клиентов, ролей, пользователей и конфигурации безопасности. Документ учитывает существующую реализацию в `internal/keycloak/client.go` и тесты в `tests/integration/keycloak_integration_test.go`.

### 1.2 Область применения
Настройка Keycloak предназначена для:
- Централизованной аутентификации пользователей CloudBridge Relay
- Управления ролями и разрешениями
- Интеграции с OpenID Connect протоколом
- Обеспечения безопасности JWT токенов
- Поддержки rate limiting на основе пользователей
- Интеграции с существующей системой мониторинга и логирования

## 2. Архитектура интеграции

### 2.1 Компоненты системы
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

### 2.2 Поток аутентификации
1. **Client Login**: Пользователь аутентифицируется в Keycloak
2. **Token Issuance**: Keycloak выдает JWT токен
3. **Relay Validation**: Relay валидирует токен через JWKS (с кэшированием)
4. **Permission Check**: Проверка ролей и разрешений через TokenInfo
5. **Access Grant**: Предоставление доступа к ресурсам

### 2.3 Интеграция с существующим кодом
Проект уже содержит реализацию Keycloak клиента в `internal/keycloak/client.go`:
- JWKS кэширование с TTL
- Валидация JWT токенов
- Структура TokenInfo для извлечения пользовательских данных
- HTTP клиент с настраиваемыми таймаутами

## 3. Требования к Keycloak

### 3.1 Версия и окружение
- **Keycloak версия**: 21.0.1 (фиксированная версия для стабильности)
- **Java**: OpenJDK 17+ или Oracle JDK 17+
- **База данных**: PostgreSQL 13+ (рекомендуется) или MySQL 8.0+
- **Память**: Минимум 2GB RAM, рекомендуется 4GB+
- **Диск**: 10GB+ свободного места

### 3.2 Сетевая конфигурация
- **Порт**: 8080 (HTTP) / 8443 (HTTPS)
- **Внешний доступ**: Только для административных операций
- **Внутренний доступ**: Для relay-серверов
- **Firewall**: Ограничить доступ только необходимыми IP

### 3.3 Docker конфигурация
```yaml
# docker-compose.keycloak.yml
keycloak:
  image: quay.io/keycloak/keycloak:21.0.1  # Фиксированная версия
  environment:
    KEYCLOAK_ADMIN: admin
    KEYCLOAK_ADMIN_PASSWORD: admin123
    KC_DB: postgres
    KC_DB_URL: jdbc:postgresql://postgres:5432/keycloak
    KC_DB_USERNAME: keycloak
    KC_DB_PASSWORD: keycloak123
    KC_HOSTNAME_STRICT: false
    KC_HOSTNAME_STRICT_HTTPS: false
    KC_HTTP_ENABLED: true
    KC_HEALTH_ENABLED: true
    KC_METRICS_ENABLED: true
  ports:
    - "8080:8080"
```

## 4. Конфигурация Realm

### 4.1 Создание Realm
```json
{
  "realm": "cloudbridge",
  "displayName": "CloudBridge",
  "enabled": true,
  "displayNameHtml": "<div class=\"kc-logo-text\"><span>CloudBridge</span></div>",
  "attributes": {
    "relay.enabled": "true",
    "relay.version": "1.1.1"
  }
}
```

### 4.2 Настройки Realm
- **Realm Name**: `cloudbridge`
- **Display Name**: `CloudBridge`
- **Enabled**: `true`
- **User Registration**: `false` (только административное создание)
- **Email Verification**: `true`
- **Reset Password**: `true`
- **Remember Me**: `true`
- **Login with Email**: `false`

### 4.3 Настройки токенов
- **Access Token Lifespan**: 15 минут
- **Client Session Idle**: 30 минут
- **Client Session Max**: 60 минут
- **Client Offline Session Idle**: 30 минут
- **Client Offline Session Max**: 60 минут
- **SSO Session Idle**: 30 минут
- **SSO Session Max**: 10 часов
- **SSO Session Idle Remember Me**: 30 минут
- **SSO Session Max Remember Me**: 10 часов
- **Offline Session Idle**: 30 минут
- **Offline Session Max**: 60 минут

## 5. Конфигурация клиентов

### 5.1 Relay Client (Основной)
```json
{
  "clientId": "relay-client",
  "enabled": true,
  "protocol": "openid-connect",
  "publicClient": false,
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": true,
  "authorizationServicesEnabled": false,
  "attributes": {
    "saml.assertion.signature": "false",
    "saml.force.post.binding": "false",
    "saml.multivalued.roles": "false",
    "saml.encrypt": "false",
    "saml.server.signature": "false",
    "saml.server.signature.keyinfo.ext": "false",
    "exclude.session.state.from.auth.response": "false",
    "saml_force_name_id_format": "false",
    "saml.client.signature": "false",
    "tls.client.certificate.bound.access.tokens": "false",
    "saml.authnstatement": "false",
    "display.on.consent.screen": "false",
    "saml.onetimeuse.condition": "false"
  },
  "redirectUris": [
    "http://localhost:8080/*",
    "https://localhost:8080/*",
    "http://relay-server:8080/*",
    "https://relay-server:8080/*"
  ],
  "webOrigins": [
    "http://localhost:8080",
    "https://localhost:8080",
    "http://relay-server:8080",
    "https://relay-server:8080"
  ]
}
```

### 5.2 Admin Client (для управления)
```json
{
  "clientId": "relay-admin",
  "enabled": true,
  "protocol": "openid-connect",
  "publicClient": false,
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": true,
  "authorizationServicesEnabled": true,
  "attributes": {
    "display.on.consent.screen": "false"
  },
  "redirectUris": [
    "http://localhost:3000/*",
    "https://localhost:3000/*"
  ]
}
```

### 5.3 Настройки клиентов
- **Access Type**: `confidential`
- **Valid Redirect URIs**: Настроить для всех relay-серверов (порт 8080)
- **Web Origins**: Настроить CORS для веб-интерфейсов
- **Service Account Roles**: Настроить для автоматизации

## 6. Роли и разрешения

### 6.1 Realm Roles
```json
[
  {
    "name": "relay-admin",
    "description": "Full administrative access to CloudBridge Relay",
    "composite": false,
    "clientRole": false,
    "attributes": {
      "relay.permissions": ["*"]
    }
  },
  {
    "name": "relay-user",
    "description": "Standard user access to CloudBridge Relay",
    "composite": false,
    "clientRole": false,
    "attributes": {
      "relay.permissions": ["tunnel:create", "tunnel:read", "tunnel:delete"]
    }
  },
  {
    "name": "relay-readonly",
    "description": "Read-only access to CloudBridge Relay",
    "composite": false,
    "clientRole": false,
    "attributes": {
      "relay.permissions": ["tunnel:read"]
    }
  },
  {
    "name": "relay-monitor",
    "description": "Monitoring access to CloudBridge Relay",
    "composite": false,
    "clientRole": false,
    "attributes": {
      "relay.permissions": ["metrics:read", "logs:read"]
    }
  }
]
```

### 6.2 Client Roles
```json
[
  {
    "name": "relay-client-admin",
    "description": "Admin role for relay-client",
    "composite": false,
    "clientRole": true,
    "attributes": {
      "relay.client.permissions": ["*"]
    }
  },
  {
    "name": "relay-client-user",
    "description": "User role for relay-client",
    "composite": false,
    "clientRole": true,
    "attributes": {
      "relay.client.permissions": ["tunnel:create", "tunnel:read"]
    }
  }
]
```

### 6.3 Иерархия ролей
```
relay-admin
├── relay-user
│   └── relay-readonly
└── relay-monitor
```

### 6.4 Реализация проверки разрешений
```go
// Пример реализации проверки разрешений в internal/keycloak/client.go
func (c *Client) CheckPermission(token *TokenInfo, permission string) bool {
    // Проверка realm ролей
    if roles, ok := token.RealmRoles["roles"]; ok {
        if strings.Contains(roles, "relay-admin") {
            return true // Админ имеет все права
        }
    }
    
    // Проверка конкретных разрешений
    switch permission {
    case "tunnel:create", "tunnel:read", "tunnel:delete":
        return c.hasRole(token, "relay-user") || c.hasRole(token, "relay-admin")
    case "metrics:read", "logs:read":
        return c.hasRole(token, "relay-monitor") || c.hasRole(token, "relay-admin")
    default:
        return false
    }
}

func (c *Client) hasRole(token *TokenInfo, role string) bool {
    if roles, ok := token.RealmRoles["roles"]; ok {
        return strings.Contains(roles, role)
    }
    return false
}
```

## 7. Группы пользователей

### 7.1 Структура групп
```json
[
  {
    "name": "relay-admins",
    "attributes": {
      "description": ["Administrators with full access"],
      "relay.max_tunnels": ["100"],
      "relay.rate_limit": ["1000"]
    }
  },
  {
    "name": "relay-users",
    "attributes": {
      "description": ["Standard users"],
      "relay.max_tunnels": ["10"],
      "relay.rate_limit": ["100"]
    }
  },
  {
    "name": "relay-monitors",
    "attributes": {
      "description": ["Monitoring users"],
      "relay.max_tunnels": ["0"],
      "relay.rate_limit": ["50"]
    }
  }
]
```

### 7.2 Наследование ролей
- **relay-admins** → `relay-admin` role
- **relay-users** → `relay-user` role
- **relay-monitors** → `relay-monitor` role

## 8. Пользователи

### 8.1 Администратор
```json
{
  "username": "admin",
  "email": "admin@cloudbridge.local",
  "firstName": "Admin",
  "lastName": "User",
  "enabled": true,
  "emailVerified": true,
  "credentials": [
    {
      "type": "password",
      "value": "SecurePassword123!",
      "temporary": false
    }
  ],
  "groups": ["relay-admins"],
  "realmRoles": ["relay-admin"],
  "attributes": {
    "relay.max_tunnels": ["100"],
    "relay.rate_limit": ["1000"],
    "relay.allowed_ips": ["*"]
  }
}
```

### 8.2 Обычный пользователь
```json
{
  "username": "user",
  "email": "user@cloudbridge.local",
  "firstName": "Regular",
  "lastName": "User",
  "enabled": true,
  "emailVerified": true,
  "credentials": [
    {
      "type": "password",
      "value": "UserPassword123!",
      "temporary": false
    }
  ],
  "groups": ["relay-users"],
  "realmRoles": ["relay-user"],
  "attributes": {
    "relay.max_tunnels": ["10"],
    "relay.rate_limit": ["100"],
    "relay.allowed_ips": ["192.168.1.0/24"]
  }
}
```

### 8.3 Мониторинг пользователь
```json
{
  "username": "monitor",
  "email": "monitor@cloudbridge.local",
  "firstName": "Monitor",
  "lastName": "User",
  "enabled": true,
  "emailVerified": true,
  "credentials": [
    {
      "type": "password",
      "value": "MonitorPassword123!",
      "temporary": false
    }
  ],
  "groups": ["relay-monitors"],
  "realmRoles": ["relay-monitor"],
  "attributes": {
    "relay.max_tunnels": ["0"],
    "relay.rate_limit": ["50"],
    "relay.allowed_ips": ["10.0.0.0/8"]
  }
}
```

## 9. Безопасность

### 9.1 Политики паролей
- **Minimum Length**: 12 символов
- **Uppercase**: Требуется
- **Lowercase**: Требуется
- **Digits**: Требуется
- **Special Characters**: Требуется
- **Not Username**: Да
- **Not Email**: Да
- **Not First Name**: Да
- **Not Last Name**: Да
- **Password History**: 5 паролей
- **Brute Force Detection**: Включено
- **Permanent Lockout**: После 5 неудачных попыток

### 9.2 Настройки сессий
- **SSO Session Idle**: 30 минут
- **SSO Session Max**: 10 часов
- **SSO Session Idle Remember Me**: 30 минут
- **SSO Session Max Remember Me**: 10 часов
- **Offline Session Idle**: 30 минут
- **Offline Session Max**: 60 минут

### 9.3 Настройки токенов
- **Access Token Lifespan**: 15 минут
- **Client Session Idle**: 30 минут
- **Client Session Max**: 60 минут
- **Client Offline Session Idle**: 30 минут
- **Client Offline Session Max**: 60 минут

### 9.4 Двухфакторная аутентификация
- **Required Actions**: `CONFIGURE_TOTP`
- **OTP Policy**: Google Authenticator
- **Algorithm**: `HmacSHA1`
- **Digits**: 6
- **Look Ahead Window**: 1
- **Period**: 30 секунд

## 10. Интеграция с Relay

### 10.1 JWKS Endpoint
```
GET /realms/cloudbridge/protocol/openid-connect/certs
```

### 10.2 Token Endpoint
```
POST /realms/cloudbridge/protocol/openid-connect/token
```

### 10.3 User Info Endpoint
```
GET /realms/cloudbridge/protocol/openid-connect/userinfo
```

### 10.4 Introspection Endpoint
```
POST /realms/cloudbridge/protocol/openid-connect/token/introspect
```

### 10.5 Конфигурация для Relay
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

### 10.6 Переменные окружения
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

### 10.7 Использование существующего кода
```go
// Пример использования internal/keycloak/client.go
import (
    "github.com/twogc/cloudbridge-relay-installer/internal/keycloak"
    "github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func setupKeycloakClient(cfg *config.Config) (*keycloak.Client, error) {
    keycloakConfig := &keycloak.Config{
        ServerURL:    cfg.Keycloak.ServerURL,
        Realm:        cfg.Keycloak.Realm,
        ClientID:     cfg.Keycloak.ClientID,
        ClientSecret: cfg.Keycloak.ClientSecret,
        TokenValidation: struct {
            Issuer    string        `json:"issuer"`
            Audience  string        `json:"audience"`
            ClockSkew time.Duration `json:"clock_skew"`
        }{
            Issuer:    cfg.Keycloak.TokenValidation.Issuer,
            Audience:  cfg.Keycloak.TokenValidation.Audience,
            ClockSkew: cfg.Keycloak.TokenValidation.ClockSkew,
        },
        Logger: logger,
    }

    return keycloak.NewClient(keycloakConfig)
}
```

## 11. Мониторинг и логирование

### 11.1 Логирование событий
- **Admin Events**: Включено
- **User Events**: Включено
- **Login Events**: Включено
- **Logout Events**: Включено
- **Token Events**: Включено

### 11.2 Экспорт логов
- **Format**: JSON
- **Rotation**: Ежедневно
- **Retention**: 90 дней
- **Compression**: Gzip

### 11.3 Метрики
- **Prometheus**: Включено
- **Endpoint**: `/metrics`
- **Authentication**: Требуется

### 11.4 Интеграция с существующими метриками
```prometheus
# Метрики Keycloak аутентификации (уже реализованы в internal/metrics/)
relay_keycloak_token_validations_total{result="success"}
relay_keycloak_token_validations_total{result="failure"}
relay_keycloak_jwks_cache_hits_total
relay_keycloak_jwks_cache_misses_total
relay_keycloak_token_validation_duration_seconds
```

## 12. Тестирование

### 12.1 Интеграционные тесты
Проект содержит готовые тесты в `tests/integration/keycloak_integration_test.go`:

```go
// Запуск тестов
go test -v ./tests/integration/ -run TestKeycloakIntegration

// Запуск с Docker
docker-compose -f docker-compose.keycloak.yml up -d
go test -v ./tests/integration/ -run TestKeycloakIntegration
```

### 12.2 Функциональное тестирование
- [ ] Создание realm
- [ ] Создание клиентов
- [ ] Создание ролей и групп
- [ ] Создание пользователей
- [ ] Аутентификация пользователей
- [ ] Выдача JWT токенов
- [ ] Валидация токенов
- [ ] Проверка ролей и разрешений

### 12.3 Интеграционное тестирование
- [ ] Подключение relay-сервера
- [ ] Аутентификация через relay
- [ ] Создание туннелей
- [ ] Rate limiting
- [ ] Аудит событий

### 12.4 Тестирование безопасности
- [ ] Валидация JWT токенов
- [ ] Проверка истечения токенов
- [ ] Тестирование ролей
- [ ] Тестирование rate limiting
- [ ] Penetration testing

## 13. Резервное копирование

### 13.1 Автоматическое резервное копирование
- **Frequency**: Ежедневно
- **Retention**: 30 дней
- **Compression**: Да
- **Encryption**: Да

### 13.2 Ручное резервное копирование
```bash
# Экспорт realm
./kc.sh export --realm cloudbridge --file cloudbridge-realm.json

# Экспорт пользователей
./kc.sh export --realm cloudbridge --users same_file --file cloudbridge-users.json

# Резервное копирование базы данных
pg_dump keycloak > keycloak_backup.sql
```

## 14. Автоматизация

### 14.1 Скрипт настройки
Проект содержит готовый скрипт `scripts/setup_keycloak.sh`:

```bash
# Запуск автоматической настройки
./scripts/setup_keycloak.sh

# Настройка с кастомными параметрами
KEYCLOAK_URL=https://keycloak.company.com \
REALM_NAME=cloudbridge-prod \
CLIENT_ID=relay-client-prod \
./scripts/setup_keycloak.sh
```

### 14.2 Docker Compose
```bash
# Запуск Keycloak с Relay
docker-compose -f docker-compose.keycloak.yml up -d

# Проверка статуса
docker-compose -f docker-compose.keycloak.yml ps

# Просмотр логов
docker-compose -f docker-compose.keycloak.yml logs -f keycloak
```

## 15. Документация

### 15.1 Обязательная документация
- README с инструкциями по установке
- Конфигурационные файлы
- Политики безопасности
- Процедуры резервного копирования
- Troubleshooting guide

### 15.2 Рекомендуемая документация
- Architecture overview
- Security considerations
- Performance tuning guide
- Disaster recovery plan
- User management guide

### 15.3 Интеграция с существующей документацией
- Обновить `docs/keycloak_integration.md` с учетом ТЗ
- Синхронизировать примеры конфигураций
- Добавить ссылки на тесты и скрипты

## 16. Критерии приемки

### 16.1 Функциональные требования
- [ ] Realm создан и настроен
- [ ] Клиенты созданы и настроены
- [ ] Роли и группы созданы
- [ ] Пользователи созданы
- [ ] JWT токены выдаются корректно
- [ ] Интеграция с relay работает
- [ ] Все интеграционные тесты проходят
- [ ] Скрипт автоматической настройки работает

### 16.2 Нефункциональные требования
- [ ] Время ответа < 1 секунды
- [ ] Доступность 99.9%
- [ ] Безопасность соответствует требованиям
- [ ] Логирование работает корректно
- [ ] Резервное копирование настроено
- [ ] Метрики собираются корректно

### 16.3 Требования безопасности
- [ ] TLS 1.3 включен
- [ ] Политики паролей настроены
- [ ] Двухфакторная аутентификация работает
- [ ] Аудит событий включен
- [ ] Rate limiting настроен
- [ ] JWKS кэширование работает

## 17. Сроки и этапы

### 17.1 Этап 1 (1-2 дня)
- Установка и базовая настройка Keycloak
- Создание realm и клиентов
- Настройка безопасности
- Интеграция с существующим кодом

### 17.2 Этап 2 (1 день)
- Создание ролей и групп
- Создание пользователей
- Настройка интеграции
- Запуск интеграционных тестов

### 17.3 Этап 3 (1 день)
- Тестирование
- Документация
- Передача в эксплуатацию
- Обучение команды

## 18. Контакты

- **Технический контакт:** [Указать контакт]
- **Security контакт:** [Указать контакт]
- **Репозиторий:** [Указать ссылку]
- **Issue tracker:** [Указать ссылку]

## 19. Приложения

### 19.1 Скрипты автоматизации
- `scripts/setup_keycloak.sh` - основной скрипт настройки
- `scripts/create_users.sh` - создание пользователей
- `scripts/backup_keycloak.sh` - резервное копирование
- `scripts/test_integration.sh` - тестирование интеграции

### 19.2 Конфигурационные файлы
- `config/config.json` - конфигурация для relay
- `docker-compose.keycloak.yml` - Docker Compose для Keycloak
- `keycloak_config.json` - экспорт конфигурации Keycloak

### 19.3 Тесты
- `tests/integration/keycloak_integration_test.go` - интеграционные тесты
- `tests/e2e/relay_e2e_test.go` - end-to-end тесты

### 19.4 Документация
- `docs/keycloak_integration.md` - руководство по интеграции
- `docs/api_reference.md` - API документация
- `docs/deployment_guide.md` - руководство по развертыванию
- `docs/troubleshooting.md` - устранение неполадок 