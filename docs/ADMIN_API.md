# CloudBridge Relay Admin API

## Обзор

Admin API предоставляет административный интерфейс для управления CloudBridge Relay с аутентификацией через Zitadel.

## Аутентификация

Admin API использует JWT токены от Zitadel для аутентификации. Требуются следующие роли:
- `admin-ui` - для доступа к UI администратора
- `admin` - для полного административного доступа

### Заголовки запроса

```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

## Endpoints

### GET /api/status

Проверка статуса Admin API.

**Ответ:**
```json
{
  "status": "ok",
  "message": "Admin API is running"
}
```

### GET /api/audit/logs

Получение логов аудита.

**Ответ:**
```json
{
  "events": [
    {
      "timestamp": "2025-01-05T12:00:00Z",
      "level": "info",
      "message": "User login",
      "user": "admin@example.com",
      "ip": "192.168.1.100"
    }
  ]
}
```

## Конфигурация

Admin API настраивается через конфигурацию CloudBridge Relay:

```json
{
  "zitadel": {
    "enabled": true,
    "domain": "https://zitadel.example.com",
    "project_id": "cloudbridge-project",
    "client_id": "cloudbridge-relay",
    "client_secret": "your-client-secret",
    "fallback_secret": "your-fallback-secret",
    "token_validation": {
      "audience": "cloudbridge-relay",
      "clock_skew": "30s"
    }
  }
}
```

## Использование

### Запуск Admin API

```go
package main

import (
    "github.com/twogc/cloudbridge-relay-installer/internal/adminapi"
    "github.com/twogc/cloudbridge-relay-installer/internal/auth"
    "github.com/twogc/cloudbridge-relay-installer/internal/config"
    "github.com/twogc/cloudbridge-relay-installer/internal/logger"
)

func main() {
    // Создаем логгер
    logger, _ := logger.NewLogger("admin-api", "info")
    
    // Загружаем конфигурацию
    cfg, _ := config.LoadConfig("config.json")
    
    // Создаем auth adapter
    authAdapter, _ := auth.NewAdapter(cfg, logger)
    
    // Запускаем Admin API
    adminapi.RunServer(":8080", authAdapter, logger)
}
```

### Пример запроса

```bash
curl -H "Authorization: Bearer <jwt_token>" \
     -H "Content-Type: application/json" \
     http://localhost:8080/api/status
```

## Безопасность

- Все запросы требуют валидный JWT токен от Zitadel
- Проверяется роль пользователя (admin-ui или admin)
- Используется CORS для ограничения доступа
- Логируются все попытки доступа

## Мониторинг

Admin API предоставляет метрики для мониторинга:
- Количество запросов
- Время ответа
- Ошибки аутентификации
- Активные сессии

## Логирование

Все действия логируются в файл `/var/log/relay/audit.log`:
- Попытки входа
- Доступ к API
- Ошибки аутентификации
- Административные действия
