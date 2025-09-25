# Миграция с Keycloak на Zitadel

## Обзор

Этот документ описывает процесс миграции системы аутентификации CloudBridge Relay с Keycloak на Zitadel.

## 🎯 Цели миграции

1. **Улучшенная производительность**: Zitadel обеспечивает лучшую производительность и масштабируемость
2. **Современная архитектура**: Event sourcing, CQRS, cloud-native подход
3. **Лучший DX**: Более простой API и интеграция
4. **Встроенное управление тенантами**: Нативная поддержка мульти-тенантности
5. **Безопасность**: Современные стандарты безопасности из коробки

## 🏗️ Архитектура интеграции

### Новые компоненты

```
internal/zitadel/
├── validator.go          # Валидатор JWT токенов через Zitadel API
├── tenant_manager.go     # Управление тенантами через Zitadel организации
└── types.go              # Типы данных (совместимые с Keycloak)

cmd/relay/
├── zitadel_integration.go # Интеграция Zitadel в основной сервер
└── migration_helper.go    # Помощник для плавной миграции

configs/
├── zitadel_config.json   # Конфигурация Zitadel
└── config.json           # Обновленная основная конфигурация
```

### Совместимость

- **Обратная совместимость**: Поддержка старых Keycloak токенов через fallback
- **Постепенная миграция**: Возможность работы в гибридном режиме
- **Сохранение API**: Существующие endpoints остаются без изменений

## 🔧 Конфигурация

### 1. Переменные окружения

```bash
# Zitadel настройки
export ZITADEL_DOMAIN="https://zitadel.local"
export ZITADEL_PROJECT_ID="your-project-id"
export ZITADEL_CLIENT_ID="your-client-id"
export ZITADEL_CLIENT_SECRET="your-client-secret"

# Отключение Keycloak
export KEYCLOAK_ENABLED="false"
```

### 2. Конфигурационный файл

```json
{
  "keycloak": {
    "enabled": false
  },
  "zitadel": {
    "enabled": true,
    "domain": "https://zitadel.local",
    "project_id": "${ZITADEL_PROJECT_ID}",
    "client_id": "${ZITADEL_CLIENT_ID}",
    "client_secret": "${ZITADEL_CLIENT_SECRET}",
    "tenant_management": {
      "enabled": true,
      "auto_create": false
    }
  }
}
```

## 🚀 Процесс миграции

### Этап 1: Подготовка Zitadel

1. **Развертывание Zitadel в Minikube**:
   ```bash
   kubectl apply -f k8s/zitadel/
   ```

2. **Создание проекта и приложения**:
   ```bash
   # Использовать Zitadel Console или API
   # Получить project_id, client_id, client_secret
   ```

3. **Настройка организаций (тенантов)**:
   ```bash
   # Создать организации для каждого существующего тенанта
   ```

### Этап 2: Обновление кода

1. **Установка зависимостей**:
   ```bash
   go mod tidy
   ```

2. **Обновление конфигурации**:
   ```bash
   # Обновить configs/config.json
   # Установить переменные окружения
   ```

### Этап 3: Тестирование

1. **Модульные тесты**:
   ```bash
   go test ./internal/zitadel/...
   ```

2. **Интеграционные тесты**:
   ```bash
   # Тестирование с реальным Zitadel
   ```

### Этап 4: Развертывание

1. **Гибридный режим** (рекомендуется):
   - Включить Zitadel для новых клиентов
   - Сохранить Keycloak fallback для существующих

2. **Полная миграция**:
   - Отключить Keycloak
   - Перевести всех клиентов на Zitadel

## 🔄 Mapping между системами

### Keycloak → Zitadel

| Keycloak | Zitadel | Описание |
|----------|---------|----------|
| Realm | Organization | Изоляция тенантов |
| Client | Application | Приложения для аутентификации |
| User | User | Пользователи |
| Role | Role | Роли и разрешения |
| Group | - | Используются метаданные |

### Токены и Claims

```json
// Keycloak JWT
{
  "iss": "https://auth.2gc.ru/realms/cloudbridge",
  "aud": "relay-client",
  "tenant_id": "tenant-123",
  "server_id": "server-456",
  "protocol_type": "p2p-mesh"
}

// Zitadel JWT (совместимый)
{
  "iss": "https://zitadel.local",
  "aud": "cloudbridge-relay",
  "tenant_id": "org-789",        // из метаданных
  "server_id": "server-456",     // из метаданных
  "protocol_type": "p2p-mesh"    // из метаданных
}
```

## 🔐 Управление тенантами

### Создание тенанта

```go
// Через Zitadel API
tenantManager := zitadel.NewTenantManager(validator, logger)

tenant, err := tenantManager.CreateTenant(ctx, "CloudBridge Corp", "corp.example.com", map[string]interface{}{
    "max_peers": 50,
    "max_connections": 500,
    "max_tunnels": 100,
})
```

### Квоты и лимиты

```go
// Обновление квот
quotas := map[string]interface{}{
    "max_peers": 100,
    "bandwidth_limit": "10GB",
}

err := tenantManager.UpdateTenantQuotas(ctx, tenantID, quotas)
```

## 📊 Мониторинг миграции

### Метрики

- `zitadel_token_validations_total` - Количество валидаций токенов
- `zitadel_tenant_operations_total` - Операции с тенантами
- `keycloak_fallback_validations_total` - Fallback валидации

### Логирование

```json
{
  "level": "INFO",
  "msg": "Token validated via Zitadel",
  "tenant_id": "org-123",
  "validation_method": "zitadel"
}
```

## Troubleshooting

### Распространенные проблемы

1. **Ошибка подключения к Zitadel**:
   ```
   failed to create Zitadel client: connection refused
   ```
   - Проверить доступность Zitadel
   - Проверить сетевые политики

2. **Неверные credentials**:
   ```
   failed to authenticate: invalid client credentials
   ```
   - Проверить ZITADEL_CLIENT_ID и ZITADEL_CLIENT_SECRET
   - Убедиться что приложение активно

3. **Токен не валидируется**:
   ```
   token validation failed: audience mismatch
   ```
   - Проверить audience в конфигурации
   - Убедиться что токен выдан для правильного приложения

### Откат

В случае проблем можно быстро вернуться к Keycloak:

```json
{
  "keycloak": {
    "enabled": true
  },
  "zitadel": {
    "enabled": false
  }
}
```

## 🎯 Roadmap

### Фаза 1: Базовая интеграция ✅
- [x] Zitadel валидатор токенов
- [x] Управление тенантами
- [x] Конфигурация и интеграция

### Фаза 2: Расширенные функции
- [ ] Миграция пользователей из Keycloak
- [ ] Автоматическое создание тенантов
- [ ] Расширенные метаданные

### Фаза 3: Оптимизация
- [ ] Кэширование токенов
- [ ] Batch операции
- [ ] Мониторинг и алерты

## 📚 Ресурсы

- [Zitadel Documentation](https://zitadel.com/docs)
- [Zitadel Go SDK](https://github.com/zitadel/zitadel-go)
- [Миграция с Keycloak](https://zitadel.com/docs/guides/migrate/keycloak)
- [OIDC/OAuth2 в Zitadel](https://zitadel.com/docs/apis/openidoauth)

## 🤝 Поддержка

Для вопросов по миграции:
- Создать issue в репозитории
- Обратиться к команде DevOps
- Проверить логи и метрики

---

**Примечание**: Миграция требует тщательного планирования и тестирования. Рекомендуется начать с development/staging окружения.
