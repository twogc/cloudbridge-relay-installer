# Sentry OAuth Integration Guide

## Обзор

Этот документ описывает интеграцию OAuth с Sentry для CloudBridge Relay, включая настройку аутентификации, webhook обработку и API интеграцию.

## Компоненты

### 1. OAuth Configuration (`internal/sentry/oauth.go`)

- **OAuthConfig**: Конфигурация OAuth клиента
- **OAuthToken**: Управление токенами доступа
- **OAuthClient**: Клиент для OAuth операций

### 2. Webhook Handler (`internal/sentry/webhook.go`)

- **WebhookConfig**: Конфигурация webhook
- **WebhookEvent**: Обработка событий от Sentry
- **WebhookHandler**: Обработчик webhook событий

### 3. Sentry API Client (`internal/sentry/api.go`)

- **SentryAPI**: Клиент для работы с Sentry API
- **SentryProject**: Управление проектами
- **SentryIssue**: Управление проблемами
- **SentryEvent**: Управление событиями

## Настройка

### 1. OAuth Credentials

```bash
# В .env файле
SENTRY_CLIENT_ID=eb82ceb348e16911bf6b5d1df08ec91b8ad24567bccf949e3dbd16ad8640ba9e
SENTRY_CLIENT_SECRET=175cdadf71b0827102dad54a7ad9370e08099299af578dd17f60e5133549c91c
SENTRY_AUTH_URL=https://sentry.io/oauth/authorize/
SENTRY_TOKEN_URL=https://sentry.io/oauth/token/
SENTRY_OAUTH_SCOPES=project:read,project:write,event:read,event:write,org:read
SENTRY_REDIRECT_URL=http://localhost:8080/auth/sentry/callback
```

### 2. Webhook Configuration

```bash
# В .env файле
SENTRY_WEBHOOK_URL=https://your-webhook-url.com/sentry
SENTRY_WEBHOOK_SECRET=your-webhook-secret
SENTRY_WEBHOOK_ENABLED=true
```

## Использование

### 1. OAuth Flow

```go
// Создание OAuth клиента
oauthConfig := sentry.DefaultOAuthConfig()
oauthClient := sentry.NewOAuthClient(oauthConfig)

// Получение URL авторизации
authURL := oauthClient.GetAuthURL("state")

// Обмен кода на токен
token, err := oauthClient.ExchangeCodeForToken(ctx, code)
```

### 2. Webhook Handling

```go
// Создание webhook обработчика
webhookConfig := sentry.DefaultWebhookConfig()
webhookHandler := sentry.NewWebhookHandler(webhookConfig)

// Обработка webhook событий
webhookHandler.HandleWebhook(w, r)
```

### 3. Sentry API

```go
// Создание API клиента
api := sentry.NewSentryAPI("https://sentry.io", token)

// Получение проектов
projects, err := api.GetProjects(ctx)

// Получение проблем
issues, err := api.GetIssues(ctx, orgSlug, projectSlug)
```

## Endpoints

### OAuth Endpoints

- `GET /auth/sentry` - Получение URL авторизации
- `GET /auth/sentry/callback` - OAuth callback

### Webhook Endpoints

- `POST /webhook/sentry` - Обработка webhook событий

### API Endpoints

- `GET /api/sentry/projects` - Получение проектов
- `GET /api/sentry/issues` - Получение проблем
- `POST /api/sentry/notes` - Создание заметок

## События Webhook

### Поддерживаемые события

1. **created** - Создание новой проблемы
2. **resolved** - Решение проблемы
3. **assigned** - Назначение проблемы
4. **ignored** - Игнорирование проблемы

### Обработка событий

```go
func (h *WebhookHandler) processEvent(event *WebhookEvent) error {
    switch event.Action {
    case "created":
        return h.handleIssueCreated(event)
    case "resolved":
        return h.handleIssueResolved(event)
    case "assigned":
        return h.handleIssueAssigned(event)
    case "ignored":
        return h.handleIssueIgnored(event)
    default:
        return fmt.Errorf("unknown action: %s", event.Action)
    }
}
```

## Безопасность

### 1. Webhook Signature Verification

```go
func (h *WebhookHandler) VerifySignature(payload []byte, signature string) bool {
    if h.config.Secret == "" {
        return true // Если секрет не настроен, пропускаем проверку
    }

    expectedMAC := h.calculateHMAC(payload)
    expectedSignature := fmt.Sprintf("sha256=%s", expectedMAC)
    
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

### 2. Token Management

```go
// Проверка валидности токена
if !token.IsTokenValid() {
    // Обновление токена
    newToken, err := oauthClient.RefreshToken(ctx, token.RefreshToken)
    if err != nil {
        return err
    }
    token = newToken
}
```

## Тестирование

### 1. Запуск тестового сервера

```bash
# Загрузка переменных окружения
source scripts/load-env.sh

# Запуск тестового сервера
go run test-sentry-oauth.go
```

### 2. Тестирование OAuth

```bash
# Получение URL авторизации
curl http://localhost:8080/auth/sentry

# Тестирование callback (с реальным кодом)
curl "http://localhost:8080/auth/sentry/callback?code=REAL_CODE&state=cloudbridge-123"
```

### 3. Тестирование Webhook

```bash
# Отправка тестового webhook события
curl -X POST http://localhost:8080/webhook/sentry \
  -H "Content-Type: application/json" \
  -H "X-Sentry-Hook-Signature: sha256=..." \
  -d '{"action":"created","data":{"issue":{"id":"123","title":"Test Issue"}}}'
```

## Мониторинг

### 1. Логирование

```go
// Логирование OAuth событий
fmt.Printf("🔐 OAuth token obtained: %s\n", token.AccessToken[:20])

// Логирование webhook событий
fmt.Printf("🚨 New issue created: %+v\n", event.Data)
```

### 2. Метрики

- Количество успешных OAuth авторизаций
- Количество обработанных webhook событий
- Время ответа API запросов
- Количество ошибок аутентификации

## Troubleshooting

### 1. OAuth Issues

- Проверьте правильность `SENTRY_CLIENT_ID` и `SENTRY_CLIENT_SECRET`
- Убедитесь, что `SENTRY_REDIRECT_URL` соответствует настройкам в Sentry
- Проверьте scopes в `SENTRY_OAUTH_SCOPES`

### 2. Webhook Issues

- Проверьте `SENTRY_WEBHOOK_SECRET` для верификации подписи
- Убедитесь, что webhook URL доступен извне
- Проверьте логи на предмет ошибок обработки

### 3. API Issues

- Проверьте валидность токена доступа
- Убедитесь, что токен имеет необходимые права
- Проверьте rate limits Sentry API

## Дальнейшее развитие

### 1. Расширение функциональности

- Добавление поддержки дополнительных Sentry API endpoints
- Реализация кэширования для улучшения производительности
- Добавление метрик и мониторинга

### 2. Интеграция с другими системами

- Slack уведомления для критических ошибок
- Email уведомления для назначенных проблем
- Интеграция с системами управления задачами

### 3. Улучшение безопасности

- Реализация JWT токенов для внутренней аутентификации
- Добавление rate limiting для API endpoints
- Улучшение валидации входных данных



