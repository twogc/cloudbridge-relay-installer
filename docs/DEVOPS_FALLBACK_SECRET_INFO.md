# 🔐 Fallback Secret для CloudBridge Client

## 📋 Информация для DevOps

**DevOps серверного клиента** запросил fallback секрет для валидации JWT токенов.

### 🎯 Детали запроса:
- **Keycloak сервер**: `https://auth.2gc.ru/realms/cloudbridge`
- **JWT токены содержат**: `kid: "fallback-key"`
- **Алгоритм**: `HS256`
- **Пример токена**: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9...`

### ✅ Fallback Secret предоставлен:

```
85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
```

### 🔧 Конфигурация для CloudBridge Client:

```json
{
  "keycloak": {
    "enabled": true,
    "server_url": "https://auth.2gc.ru",
    "realm": "cloudbridge",
    "fallback_secret": "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=",
    "token_validation": {
      "issuer": "https://auth.2gc.ru/realms/cloudbridge",
      "audience": "account"
    }
  }
}
```

### 📝 Переменные окружения:

```bash
KEYCLOAK_FALLBACK_SECRET=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
```

### 🚀 Как это работает:

1. **JWT токены с `kid: "fallback-key"`** используют алгоритм `HS256`
2. **Fallback секрет** используется для валидации таких токенов
3. **Основные токены** (с RSA ключами) используют JWKS валидацию
4. **Fallback режим** обеспечивает совместимость с legacy токенами

### ⚠️ Важные замечания:

- **Fallback секрет** должен быть одинаковым на всех серверах
- **Не используйте** fallback секрет для продакшена без обновления
- **Рекомендуется** ротация секретов в продакшене
- **Логирование** включено для отслеживания использования fallback режима

### 🔍 Проверка работы:

```bash
# Проверка валидации токена
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     https://your-cloudbridge-server/api/v1/validate
```

### 📞 Контакты:

При возникновении проблем с fallback секретом обращайтесь к команде разработки.

---
**Сгенерировано**: $(date)  
**Версия**: CloudBridge Relay v1.0.0
