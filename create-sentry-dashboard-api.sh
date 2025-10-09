#!/bin/bash
# Создание Sentry Dashboard через API с правильной схемой

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 Создание Sentry Dashboard через API"
echo "====================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Проверяем наличие API токена
if [ -z "$SENTRY_API_TOKEN" ]; then
    echo "❌ SENTRY_API_TOKEN не найден в .env файле"
    exit 1
fi

echo "✅ API токен найден"
echo ""

# Создаем дашборд с правильной схемой
echo "📊 Создание дашборда..."

DASHBOARD_DATA='{
  "title": "CloudBridge Relay Dashboard",
  "projects": ['$PROJECT_ID'],
  "environment": ["'$SENTRY_ENVIRONMENT'"],
  "filters": {
    "release": ["'$SENTRY_RELEASE'"]
  },
  "widgets": [
    {
      "title": "Error Rate",
      "display_type": "line",
      "widget_type": "discover",
      "queries": [
        {
          "name": "error_rate",
          "fields": ["count()"],
          "conditions": "event.type:error",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "Response Time",
      "display_type": "line",
      "widget_type": "transaction-like",
      "queries": [
        {
          "name": "response_time",
          "fields": ["avg(transaction.duration)"],
          "conditions": "event.type:transaction",
          "aggregates": ["avg(transaction.duration)"]
        }
      ]
    },
    {
      "title": "Throughput",
      "display_type": "line",
      "widget_type": "transaction-like",
      "queries": [
        {
          "name": "throughput",
          "fields": ["count()"],
          "conditions": "event.type:transaction",
          "aggregates": ["count()"]
        }
      ]
    }
  ]
}'

# Отправляем запрос на создание дашборда
echo "Отправка запроса..."
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
    echo ""
    echo "📊 Созданные виджеты:"
    echo "   - Error Rate (график ошибок)"
    echo "   - Response Time (время ответа)"
    echo "   - Throughput (пропускная способность)"
    echo ""
    echo "🎯 Дашборд готов к использованию!"
else
    echo "❌ Ошибка при создании дашборда"
    echo "Ответ: $RESPONSE"
    echo ""
    echo "🔍 Возможные причины:"
    echo "   - Недостаточно прав доступа"
    echo "   - Неверный формат данных"
    echo "   - Проблемы с API токеном"
fi
