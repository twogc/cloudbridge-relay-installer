# Руководство по интеграции AI отчетов в Telegram

## 🎯 Обзор

Это руководство описывает развертывание и настройку профессиональной системы генерации отчетов в Telegram с интеграцией Cloudflare AI для анализа метрик CloudBridge.

## 📋 Что реализовано

### ✅ Функциональность
- **Реальная интеграция с Cloudflare AI** - отчеты генерируются через Cloudflare AI API
- **Профессиональный анализ** - отчеты содержат детальный анализ, а не только метрики
- **Интеллектуальные рекомендации** - система предоставляет конкретные рекомендации
- **Надежный fallback** - система работает даже при недоступности Cloudflare AI
- **Профессиональное форматирование** - отчеты выглядят профессионально в Telegram
- **Автоматические отчеты** - CronJob генерирует отчеты каждый час

### ✅ Технические компоненты
- `internal/ai/analytics/ai_report.go` - структура данных AIReport
- `internal/ai/analytics/cloudflare_client.go` - Cloudflare AI клиент
- `cmd/ai-service/handlers.go` - обновленные endpoint'ы
- `k8s/cronjobs/ai-telegram-cronjob.yaml` - CronJob для автоматических отчетов
- `k8s/secrets/cloudflare-ai-secret.yaml` - Kubernetes Secret для Cloudflare AI

## 🚀 Развертывание

### 1. Настройка Cloudflare AI

#### 1.1 Получение токенов
```bash
# Получите токен Cloudflare AI
export CLOUDFLARE_AI_TOKEN="your_cloudflare_ai_token"
export CLOUDFLARE_ACCOUNT_ID="your_account_id"
```

#### 1.2 Создание Kubernetes Secret
```bash
# Создайте secret с вашими токенами
kubectl create secret generic cloudflare-ai-secret \
  --from-literal=CLOUDFLARE_AI_TOKEN="$CLOUDFLARE_AI_TOKEN" \
  --from-literal=CLOUDFLARE_ACCOUNT_ID="$CLOUDFLARE_ACCOUNT_ID" \
  -n cloudbridge
```

### 2. Настройка Telegram

#### 2.1 Создание бота
1. Создайте бота через @BotFather в Telegram
2. Получите токен бота
3. Получите chat_id (можно через @userinfobot)

#### 2.2 Создание Telegram Secret
```bash
# Создайте secret для Telegram
kubectl create secret generic ai-telegram-secret \
  --from-literal=TELEGRAM_BOT_TOKEN="your_bot_token" \
  --from-literal=TELEGRAM_CHAT_ID="your_chat_id" \
  -n cloudbridge
```

### 3. Развертывание компонентов

#### 3.1 Развертывание AI сервиса
```bash
# Примените deployment
kubectl apply -f k8s/ai-service-telegram-deployment.yaml
```

#### 3.2 Развертывание CronJob
```bash
# Примените CronJob для автоматических отчетов
kubectl apply -f k8s/cronjobs/ai-telegram-cronjob.yaml
```

#### 3.3 Проверка развертывания
```bash
# Проверьте статус подов
kubectl get pods -n cloudbridge -l app.kubernetes.io/name=ai-service

# Проверьте CronJob
kubectl get cronjobs -n cloudbridge

# Проверьте логи AI сервиса
kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service -f
```

## 🧪 Тестирование

### 1. Автоматическое тестирование
```bash
# Запустите скрипт тестирования
./scripts/test-ai-telegram-integration.sh
```

### 2. Ручное тестирование

#### 2.1 Генерация отчета с Cloudflare AI
```bash
curl -X POST "http://localhost:8083/api/v1/ai/generate-report" \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "telegram",
    "format": "professional",
    "language": "russian",
    "use_cloudflare_ai": true,
    "include_metrics": true,
    "include_analysis": true,
    "include_recommendations": true,
    "include_forecast": true
  }'
```

#### 2.2 Генерация отчета с fallback
```bash
curl -X POST "http://localhost:8083/api/v1/ai/generate-report" \
  -H "Content-Type: application/json" \
  -d '{
    "report_type": "telegram",
    "format": "professional",
    "language": "russian",
    "use_cloudflare_ai": false,
    "include_metrics": true,
    "include_analysis": true,
    "include_recommendations": true,
    "include_forecast": true
  }'
```

#### 2.3 Отправка произвольного сообщения
```bash
curl -X POST "http://localhost:8083/api/v1/ai/send-telegram-message" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Тестовое сообщение от CloudBridge AI"
  }'
```

## 📊 Мониторинг

### 1. Логи CronJob
```bash
# Просмотр логов последнего запуска CronJob
kubectl logs -n cloudbridge -l job-name=ai-telegram-notifications --tail=100
```

### 2. Логи AI сервиса
```bash
# Просмотр логов AI сервиса
kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service -f
```

### 3. Статус системы
```bash
# Проверка статуса AI сервиса
curl "http://localhost:8083/api/v1/ai/status"

# Проверка статуса Telegram
curl "http://localhost:8083/api/v1/ai/telegram-status"
```

## 🔧 Конфигурация

### 1. Переменные окружения

#### AI сервис
- `CLOUDFLARE_AI_TOKEN` - токен Cloudflare AI
- `CLOUDFLARE_ACCOUNT_ID` - ID аккаунта Cloudflare
- `CLOUDFLARE_AI_MODEL` - модель AI (по умолчанию: @cf/meta/llama-2-7b-chat-int8)
- `TELEGRAM_BOT_TOKEN` - токен Telegram бота
- `TELEGRAM_CHAT_ID` - ID чата для уведомлений

#### CronJob
- Расписание: `0 */1 * * *` (каждый час)
- Политика конкурентности: `Forbid`
- Лимит истории: 3 успешных, 3 неудачных

### 2. Настройка расписания

Для изменения частоты отчетов отредактируйте `k8s/cronjobs/ai-telegram-cronjob.yaml`:

```yaml
spec:
  schedule: "0 */2 * * *"  # Каждые 2 часа
  # или
  schedule: "0 9 * * *"    # Каждый день в 9:00
```

## 🐛 Устранение неполадок

### 1. Cloudflare AI недоступен
- Проверьте токен и account ID
- Убедитесь, что у аккаунта есть доступ к AI Workers
- Проверьте лимиты API

### 2. Telegram не работает
- Проверьте токен бота
- Убедитесь, что бот добавлен в чат
- Проверьте chat_id

### 3. CronJob не запускается
- Проверьте расписание cron
- Убедитесь, что namespace cloudbridge существует
- Проверьте права доступа

### 4. AI сервис недоступен
- Проверьте статус подов
- Убедитесь, что порт 8083 открыт
- Проверьте логи сервиса

## 📈 Примеры отчетов

### Отчет с Cloudflare AI
```
🟢 Отчет системы CloudBridge AI

📊 Статус системы: excellent (Оценка: 95/100)
🕐 Время: 2025-01-27 15:00:00 UTC
🏷️ Namespace: cloudbridge

📈 Ключевые метрики:
• Обнаружено аномалий: 27
• Сделано предсказаний: 27
• Обработано запросов: 27
• Оптимизировано маршрутов: 27
• Точность предсказаний: 85.5%
• Использование CPU: 12.3%
• Использование памяти: 256.7 MB
• Активные алерты: 0

🔍 AI Анализ:
Система CloudBridge AI демонстрирует отличную производительность...

💡 Рекомендации:
• Рассмотреть дополнительное обучение модели для повышения точности
• Продолжить мониторинг аномалий для раннего выявления проблем

🔮 Прогноз:
При текущих показателях система будет стабильно работать в ближайшие 24 часа.
```

## 🔄 Обновление

### 1. Обновление AI сервиса
```bash
# Пересоберите образ
docker build -f Dockerfile.ai-service -t ai-service:latest .

# Обновите deployment
kubectl rollout restart deployment/ai-service -n cloudbridge
```

### 2. Обновление конфигурации
```bash
# Примените изменения
kubectl apply -f k8s/ai-service-telegram-deployment.yaml
kubectl apply -f k8s/cronjobs/ai-telegram-cronjob.yaml
```

## 📞 Поддержка

При возникновении проблем:
1. Проверьте логи компонентов
2. Убедитесь в правильности конфигурации
3. Запустите скрипт тестирования
4. Проверьте доступность внешних сервисов (Cloudflare AI, Telegram)

---

**Примечание**: Это руководство описывает полную интеграцию AI отчетов в Telegram согласно техническому заданию DEVELOPER_TASK_TELEGRAM_AI_REPORTS.md.

