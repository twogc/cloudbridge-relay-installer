# Sentry Integration - Финальная настройка

## 🚨 Проблема

Текущий DSN не работает, так как проект с ID `4510156604702720` не существует или недоступен.

## ✅ Решение

### 1. Создание нового проекта в Sentry

1. **Откройте Sentry Dashboard**: https://sentry.io/
2. **Войдите в аккаунт** или создайте новый
3. **Создайте новую организацию** (если нужно)
4. **Создайте новый проект**:
   - Нажмите "Create Project"
   - Выберите платформу **"Go"**
   - Название: `cloudbridge-relay`
   - Описание: `CloudBridge Relay Error Monitoring`
   - Нажмите "Create Project"

### 2. Получение DSN

1. **После создания проекта** перейдите в настройки
2. **Откройте раздел "Client Keys (DSN)"**
3. **Скопируйте DSN** (формат: `https://KEY@HOST/PROJECT_ID`)

### 3. Обновление конфигурации

```bash
# Обновите .env файл с новым DSN
SENTRY_DSN=https://YOUR_NEW_KEY@YOUR_ORG.ingest.sentry.io/YOUR_NEW_PROJECT_ID
```

### 4. Тестирование

```bash
# Загрузите переменные окружения
source scripts/load-env.sh

# Запустите тест
go run test-sentry-slow.go
```

## 🔧 Альтернативное решение

Если у вас уже есть рабочий проект в Sentry:

1. **Найдите существующий проект Go**
2. **Скопируйте его DSN**
3. **Обновите .env файл**
4. **Протестируйте подключение**

## 📋 Проверка

После настройки правильного DSN вы должны увидеть:

```
✅ События успешно отправляются в Sentry
✅ Нет ошибок "ProjectId" 
✅ Данные появляются в Sentry dashboard
```

## 🚀 Готовые компоненты

Все компоненты Sentry интеграции готовы:

- ✅ **OAuth интеграция** (`internal/sentry/oauth.go`)
- ✅ **Webhook обработка** (`internal/sentry/webhook.go`) 
- ✅ **API клиент** (`internal/sentry/api.go`)
- ✅ **Gin middleware** (`internal/sentry/middleware.go`)
- ✅ **Конфигурация** (`internal/sentry/config.go`)
- ✅ **Тестовые скрипты** (`scripts/`)
- ✅ **Документация** (`docs/`)

## 📊 Следующие шаги

1. **Создайте новый проект в Sentry**
2. **Обновите DSN в .env файле**
3. **Протестируйте отправку событий**
4. **Настройте алерты и уведомления**
5. **Интегрируйте в основное приложение**

## 🔗 Полезные ссылки

- [Sentry Dashboard](https://sentry.io/)
- [Sentry Go SDK Documentation](https://docs.sentry.io/platforms/go/)
- [Sentry OAuth Setup](https://docs.sentry.io/api/auth/)
- [Sentry Webhook Configuration](https://docs.sentry.io/product/integrations/integration-platform/webhooks/)

---

**Статус**: Готово к использованию после настройки правильного DSN 🚀



