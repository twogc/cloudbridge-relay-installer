#!/bin/bash
# Автоматическое создание Sentry Dashboard для CloudBridge Relay через API

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 CloudBridge Relay - Автоматическое создание Dashboard"
echo "======================================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Проверяем наличие API токена
if [ -z "$SENTRY_API_TOKEN" ]; then
    echo "❌ SENTRY_API_TOKEN не найден в .env файле"
    echo "Добавьте токен в .env файл:"
    echo "SENTRY_API_TOKEN=your-token-here"
    exit 1
fi

echo "✅ API токен найден"
echo ""

# Создаем дашборд с виджетами для CloudBridge Relay
echo "📊 Создание дашборда для CloudBridge Relay..."

DASHBOARD_DATA='{
  "title": "CloudBridge Relay - Production Dashboard",
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
    },
    {
      "title": "Memory Usage",
      "display_type": "line",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "memory_usage",
          "fields": ["avg(system.memory.used)"],
          "conditions": "event.type:metric",
          "aggregates": ["avg(system.memory.used)"]
        }
      ]
    },
    {
      "title": "CPU Usage",
      "display_type": "line",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "cpu_usage",
          "fields": ["avg(system.cpu.used)"],
          "conditions": "event.type:metric",
          "aggregates": ["avg(system.cpu.used)"]
        }
      ]
    },
    {
      "title": "P2P Connections",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "p2p_connections",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:p2p.connections",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "QUIC Connections",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "quic_connections",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:quic.connections",
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
    echo "📊 Созданные виджеты для CloudBridge Relay:"
    echo "   - Error Rate (график ошибок)"
    echo "   - Response Time (время ответа)"
    echo "   - Throughput (пропускная способность)"
    echo "   - Memory Usage (использование памяти)"
    echo "   - CPU Usage (использование CPU)"
    echo "   - P2P Connections (P2P соединения)"
    echo "   - QUIC Connections (QUIC соединения)"
    echo ""
    echo "🎯 Дашборд готов к использованию!"
    echo ""
    echo "📋 Следующие шаги:"
    echo "   1. Откройте дашборд по ссылке выше"
    echo "   2. Настройте алерты для критических метрик"
    echo "   3. Добавьте уведомления (Slack/Email)"
    echo "   4. Настройте автообновление каждые 5 минут"
else
    echo "❌ Ошибка при создании дашборда"
    echo "Ответ: $RESPONSE"
    echo ""
    echo "🔍 Возможные причины:"
    echo "   - Недостаточно прав доступа"
    echo "   - Неверный формат данных"
    echo "   - Проблемы с API токеном"
    echo ""
    echo "💡 Рекомендация:"
    echo "   Используйте ручную настройку:"
    echo "   ./setup-cloudbridge-sentry-dashboard.sh"
fi
