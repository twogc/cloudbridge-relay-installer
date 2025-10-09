#!/bin/bash
# Полная настройка Sentry Dashboard через API

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 Sentry Dashboard Setup"
echo "=========================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Проверяем наличие API токена
if [ -z "$SENTRY_API_TOKEN" ]; then
    echo "⚠️  SENTRY_API_TOKEN не найден в .env файле"
    echo ""
    echo "📋 Для получения API токена:"
    echo "1. Откройте https://sentry.io/"
    echo "2. Перейдите в Settings → Auth Tokens"
    echo "3. Создайте новый токен с правами:"
    echo "   - project:read"
    echo "   - project:write"
    echo "   - event:read"
    echo "   - event:write"
    echo "   - org:read"
    echo "4. Добавьте в .env файл:"
    echo "   SENTRY_API_TOKEN=your-token-here"
    echo ""
    echo "🔗 Sentry Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
    exit 1
fi

echo "✅ API токен найден"
echo ""

# Создаем дашборд
echo "📊 Создание дашборда..."

DASHBOARD_DATA='{
  "title": "CloudBridge Relay - '$(date +%Y%m%d_%H%M%S)'",
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
    },
    {
      "title": "Response Time",
      "displayType": "line",
      "queries": [
        {
          "name": "response_time",
          "fields": ["avg(transaction.duration)"],
          "conditions": "event.type:transaction",
          "aggregations": [["avg", "transaction.duration", "avg"]]
        }
      ]
    },
    {
      "title": "Throughput",
      "displayType": "line",
      "queries": [
        {
          "name": "throughput",
          "fields": ["count()"],
          "conditions": "event.type:transaction",
          "aggregations": [["count", null, "count"]]
        }
      ]
    }
  ]
}'

# Отправляем запрос на создание дашборда
RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $SENTRY_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$DASHBOARD_DATA" \
  "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/")

# Проверяем ответ
DASHBOARD_ID=$(echo "$RESPONSE" | jq -r '.id')

if [ "$DASHBOARD_ID" != "null" ] && [ -n "$DASHBOARD_ID" ]; then
    echo "✅ Дашборд создан с ID: $DASHBOARD_ID"
    echo "🔗 URL: https://sentry.io/organizations/$ORG_SLUG/dashboards/$DASHBOARD_ID/"
else
    echo "❌ Ошибка при создании дашборда"
    echo "Ответ: $RESPONSE"
fi

echo ""
echo "📋 Созданные виджеты:"
echo "   - Error Rate (график ошибок)"
echo "   - Response Time (время ответа)"
echo "   - Throughput (пропускная способность)"
echo ""
echo "🎯 Дашборд готов к использованию!"
