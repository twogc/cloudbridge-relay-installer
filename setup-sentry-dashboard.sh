#!/bin/bash
# Настройка Sentry Dashboard через API

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔧 Настройка Sentry Dashboard"
echo "============================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo ""

# Создаем дашборд через API
echo "📊 Создание дашборда..."

# Данные для дашборда
DASHBOARD_DATA='{
  "title": "CloudBridge Relay Dashboard",
  "description": "Мониторинг CloudBridge Relay",
  "widgets": [
    {
      "title": "Error Rate",
      "displayType": "line",
      "queries": [
        {
          "name": "error_rate",
          "fields": ["count()"],
          "conditions": "event.type:error",
          "aggregations": [["count", null, "count"]]
        }
      ]
    }
  ]
}'

# Отправляем запрос (нужен API токен)
echo "⚠️  Для создания дашборда нужен API токен Sentry"
echo "Получите токен в Sentry Dashboard:"
echo "1. Откройте https://sentry.io/"
echo "2. Перейдите в Settings → Auth Tokens"
echo "3. Создайте новый токен с правами: project:read, project:write"
echo "4. Добавьте токен в .env файл как SENTRY_API_TOKEN"
echo ""
echo "🔗 Sentry Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
