# Zitadel Relay Integration

## Обзор

CloudBridge Relay теперь поддерживает аутентификацию через Zitadel с использованием JWT токенов для relay серверов. Эта интеграция обеспечивает:

- **Безопасную аутентификацию** relay серверов через JWT токены
- **Поддержку постоянных токенов** без времени истечения для relay серверов
- **Гибкую систему разрешений** и ролей
- **Интеграцию с Zitadel SDK** для валидации токенов

## Архитектура

### Компоненты

1. **RelayClaims** - структура для JWT токенов relay серверов
2. **ZitadelValidator** - валидатор JWT токенов от Zitadel
3. **ZitadelClient** - клиент для работы с Zitadel API
4. **RelayMiddleware** - middleware для аутентификации relay серверов
5. **ZitadelProvider** - провайдер аутентификации для Zitadel

### Структура JWT токена

```json
{
  "iss": "https://edge.2gc.ru/oauth/v2",
  "sub": "relay-server-12345",
  "aud": ["cloudbridge-relay"],
  "iat": 1759528709,
  "jti": "jwt_1759528709721_gch1b06z",
  "org_id": "org-123",
  "tenant_id": "test-tenant-dev",
  "service_name": "cloudbridge-service",
  "permissions": [
    "p2p_connect",
    "mesh_join",
    "relay_access",
    "network:allocate",
    "network:deallocate",
    "tenant:network:read",
    "tenant:network:write"
  ],
  "roles": [
    "service",
    "cloudbridge-service",
    "relay-service",
    "network-manager"
  ],
  "relay_server_id": "relay-server-1759528709721",
  "protocol_type": "quick",
  "token_type": "relay_permanent"
}
```

## Использование

### 1. Конфигурация

```yaml
zitadel:
  enabled: true
  domain: "https://edge.2gc.ru"
  project_id: "project-123"
  client_id: "cloudbridge-relay"
  fallback_secret: "your-fallback-secret"
```

### 2. Инициализация

```go
// Создание Zitadel провайдера
zitadelProvider, err := NewZitadelProvider(config, logger)
if err != nil {
    return fmt.Errorf("failed to create Zitadel provider: %v", err)
}

// Создание middleware
middleware := NewRelayMiddleware(validator, logger)

// Применение middleware
router.Handle("/api/relay", middleware.RequireRelayAuth(requiredPermissions, requiredRoles)(handler))
```

### 3. Валидация токенов

```go
// Валидация relay токена
claims, err := validator.ValidateRelayToken(ctx, tokenString)
if err != nil {
    return fmt.Errorf("token validation failed: %w", err)
}

// Проверка разрешений
if !claims.HasPermission("p2p_connect") {
    return fmt.Errorf("insufficient permissions")
}

// Проверка ролей
if !claims.HasRole("relay-service") {
    return fmt.Errorf("insufficient roles")
}
```

## Разрешения и роли

### Разрешения (Permissions)

- `p2p_connect` - подключение к P2P сети
- `mesh_join` - присоединение к mesh сети
- `relay_access` - доступ к relay функциям
- `network:allocate` - выделение IP адресов
- `network:deallocate` - освобождение IP адресов
- `tenant:network:read` - чтение сетевых конфигураций
- `tenant:network:write` - изменение сетевых конфигураций
- `admin:access` - административный доступ
- `service:access` - доступ к сервисам

### Роли (Roles)

- `service` - базовая роль сервиса
- `cloudbridge-service` - роль CloudBridge сервиса
- `relay-service` - роль relay сервиса
- `network-manager` - роль управления сетью
- `admin` - административная роль
- `system-admin` - системная административная роль

## Типы токенов

### 1. Relay Permanent Token (relay_permanent)

- **Назначение**: Постоянные relay серверы
- **Время жизни**: Без ограничений
- **Использование**: Production relay серверы

```json
{
  "token_type": "relay_permanent",
  "permissions": ["p2p_connect", "mesh_join", "relay_access"],
  "roles": ["relay-service", "network-manager"]
}
```

### 2. Relay Session Token (relay_session)

- **Назначение**: Временные relay серверы
- **Время жизни**: Ограниченное
- **Использование**: Тестовые или временные relay серверы

```json
{
  "token_type": "relay_session",
  "exp": 1759615109,
  "permissions": ["p2p_connect", "mesh_join", "relay_access"],
  "roles": ["relay-service"]
}
```

## Middleware

### RequireRelayAuth

Требует аутентификации relay сервера с определенными разрешениями и ролями.

```go
middleware.RequireRelayAuth(
    []string{"p2p_connect", "relay_access"},
    []string{"relay-service"}
)
```

### RequireRelayPermissions

Требует определенные разрешения.

```go
middleware.RequireRelayPermissions(
    []string{"network:allocate", "network:deallocate"}
)
```

### RequireRelayRoles

Требует определенные роли.

```go
middleware.RequireRelayRoles(
    []string{"network-manager"}
)
```

### RequireTenantIsolation

Проверяет изоляцию тенантов.

```go
middleware.RequireTenantIsolation()
```

## Безопасность

### 1. JWT Валидация

- Проверка подписи токена
- Проверка issuer и audience
- Проверка времени истечения (для сессионных токенов)
- Проверка типа токена

### 2. Разрешения

- Гранулярная система разрешений
- Проверка на уровне middleware
- Автоматическая валидация при каждом запросе

### 3. Роли

- Иерархическая система ролей
- Проверка принадлежности к ролям
- Гибкая настройка доступа

## Тестирование

### Unit тесты

```bash
go test ./internal/auth/... -v
```

### Интеграционные тесты

```bash
go test ./internal/auth/... -tags=integration -v
```

## Примеры использования

### 1. Создание relay сервера

```go
// Создание middleware
middleware := NewRelayMiddleware(validator, logger)

// Защита endpoint для relay серверов
router.Handle("/api/relay/connect", 
    middleware.RequireRelayAuth(
        []string{"p2p_connect", "relay_access"},
        []string{"relay-service"}
    )(relayConnectHandler))
```

### 2. Проверка разрешений

```go
func relayConnectHandler(w http.ResponseWriter, r *http.Request) {
    // Получение claims из контекста
    claims, ok := GetRelayClaimsFromContext(r.Context())
    if !ok {
        http.Error(w, "No relay claims found", http.StatusInternalServerError)
        return
    }
    
    // Проверка разрешений
    if !claims.HasPermission("p2p_connect") {
        http.Error(w, "Insufficient permissions", http.StatusForbidden)
        return
    }
    
    // Обработка запроса
    // ...
}
```

### 3. Работа с тенантами

```go
func allocateIPHandler(w http.ResponseWriter, r *http.Request) {
    // Получение ID тенанта из claims
    tenantID, ok := GetTenantIDFromContext(r.Context())
    if !ok {
        http.Error(w, "No tenant ID found", http.StatusBadRequest)
        return
    }
    
    // Выделение IP для тенанта
    ip, err := allocateIPForTenant(tenantID)
    if err != nil {
        http.Error(w, "Failed to allocate IP", http.StatusInternalServerError)
        return
    }
    
    // Возврат результата
    json.NewEncoder(w).Encode(map[string]string{"ip": ip})
}
```

## Миграция с Keycloak

### 1. Обновление конфигурации

```yaml
# Отключить Keycloak
keycloak:
  enabled: false

# Включить Zitadel
zitadel:
  enabled: true
  domain: "https://edge.2gc.ru"
  project_id: "project-123"
  client_id: "cloudbridge-relay"
  fallback_secret: "your-fallback-secret"
```

### 2. Обновление кода

```go
// Старый код (Keycloak)
if claims.HasRole("relay-service") {
    // ...
}

// Новый код (Zitadel)
if claims.HasRole("relay-service") && claims.HasPermission("p2p_connect") {
    // ...
}
```

## Мониторинг и логирование

### Логирование

```go
logger.Info("Relay authentication successful",
    zap.String("relay_server_id", claims.GetRelayServerID()),
    zap.String("tenant_id", claims.GetTenantID()),
    zap.String("org_id", claims.GetOrgID()),
    zap.Strings("permissions", claims.Permissions),
    zap.Strings("roles", claims.Roles))
```

### Метрики

- Количество успешных аутентификаций
- Количество неудачных аутентификаций
- Время валидации токенов
- Количество активных relay серверов

## Заключение

Интеграция с Zitadel обеспечивает:

- ✅ **Безопасную аутентификацию** relay серверов
- ✅ **Гибкую систему разрешений** и ролей
- ✅ **Поддержку постоянных токенов** для production
- ✅ **Интеграцию с Zitadel SDK** для валидации
- ✅ **Middleware для защиты** API endpoints
- ✅ **Полную совместимость** с существующим кодом

**Готово к использованию в production!** 🚀
