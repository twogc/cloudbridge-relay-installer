# 🔌 API документация

Эта папка содержит документацию по API CloudBridge Relay.

## 📋 Содержание

### 🔑 Аутентификация
- **keycloak_tokens_guide.md** - Работа с токенами Keycloak

## 🎯 API эндпоинты

### Auth Service API

#### Аутентификация
```
POST /auth/login          - Вход в систему
POST /auth/logout         - Выход из системы
POST /auth/refresh        - Обновление токена
GET  /auth/profile        - Профиль пользователя
GET  /auth/status         - Статус аутентификации
```

#### Health checks
```
GET  /health              - Проверка здоровья
GET  /ready               - Проверка готовности
GET  /metrics             - Prometheus метрики
```

### Metrics Service API

#### Метрики
```
GET  /metrics/connections - Метрики подключений
GET  /metrics/tunnels     - Метрики туннелей
GET  /metrics/errors      - Метрики ошибок
GET  /metrics/cache       - Метрики кэша
GET  /metrics/security    - Метрики безопасности
GET  /metrics/business    - Бизнес-метрики
```

#### Health checks
```
GET  /health              - Проверка здоровья
GET  /ready               - Проверка готовности
GET  /metrics             - Prometheus метрики
```

## 🔐 Аутентификация

### JWT токены
- **Access Token**: Краткосрочный токен для API
- **Refresh Token**: Долгосрочный токен для обновления
- **ID Token**: Информация о пользователе

### Keycloak интеграция
Подробности в файле `keycloak_tokens_guide.md`

## 📊 Примеры запросов

### Получение токена
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "user", "password": "password"}'
```

### Проверка здоровья
```bash
curl http://localhost:8080/health
```

### Получение метрик
```bash
curl http://localhost:8080/metrics
```

## 🔗 Связанные разделы

- **Разработка**: `../development/`
- **Архитектура**: `../architecture/`
- **Установка**: `../installation/`
