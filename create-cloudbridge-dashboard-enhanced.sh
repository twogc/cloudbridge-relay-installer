#!/bin/bash
# Улучшенное автоматическое создание Sentry Dashboard для CloudBridge Relay

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 CloudBridge Relay - Enhanced Dashboard Creation"
echo "================================================="
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
    echo ""
    echo "🔗 Получить токен:"
    echo "   https://sentry.io/settings/account/api/auth-tokens/"
    echo "   Права: org:read, org:write, project:read, project:write, event:read, event:write"
    exit 1
fi

echo "✅ API токен найден"
echo ""

# Функция для выполнения API запросов
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    
    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "Authorization: Bearer $SENTRY_API_TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $SENTRY_API_TOKEN" \
            "$endpoint"
    fi
}

# Проверяем доступ к API
echo "🔐 Проверка доступа к API..."
ORG_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/")
if ! echo "$ORG_RESPONSE" | jq -e '.slug' > /dev/null 2>&1; then
    echo "❌ Ошибка доступа к API"
    echo "   Ответ: $ORG_RESPONSE"
    echo ""
    echo "💡 Возможные решения:"
    echo "   1. Проверьте правильность API токена"
    echo "   2. Убедитесь, что токен имеет права org:read"
    echo "   3. Получите новый токен с полными правами"
    exit 1
fi

echo "✅ Доступ к API подтвержден"
echo ""

# Создаем дашборд с виджетами для CloudBridge Relay
echo "📊 Создание дашборда для CloudBridge Relay..."

# Основные виджеты для CloudBridge Relay
DASHBOARD_DATA='{
  "title": "CloudBridge Relay - Production Dashboard",
  "projects": ['$PROJECT_ID'],
  "environment": ["'$SENTRY_ENVIRONMENT'"],
  "filters": {
    "release": ["'$SENTRY_RELEASE'"]
  },
  "widgets": [
    {
      "title": "Error Rate - CloudBridge Relay",
      "display_type": "line",
      "widget_type": "discover",
      "queries": [
        {
          "name": "error_rate",
          "fields": ["count()"],
          "conditions": "event.type:error AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "Response Time - API Performance",
      "display_type": "line",
      "widget_type": "transaction-like",
      "queries": [
        {
          "name": "response_time",
          "fields": ["avg(transaction.duration)"],
          "conditions": "event.type:transaction AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["avg(transaction.duration)"]
        }
      ]
    },
    {
      "title": "Throughput - Requests per Minute",
      "display_type": "line",
      "widget_type": "transaction-like",
      "queries": [
        {
          "name": "throughput",
          "fields": ["count()"],
          "conditions": "event.type:transaction AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "Memory Usage - System Resources",
      "display_type": "line",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "memory_usage",
          "fields": ["avg(system.memory.used)"],
          "conditions": "event.type:metric AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["avg(system.memory.used)"]
        }
      ]
    },
    {
      "title": "CPU Usage - System Resources",
      "display_type": "line",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "cpu_usage",
          "fields": ["avg(system.cpu.used)"],
          "conditions": "event.type:metric AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["avg(system.cpu.used)"]
        }
      ]
    },
    {
      "title": "P2P Connections - Active Connections",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "p2p_connections",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:p2p.connections AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "QUIC Connections - QUIC Protocol",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "quic_connections",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:quic.connections AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "WireGuard Tunnels - VPN Connections",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "wireguard_tunnels",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:wireguard.tunnels AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "DERP Server Status - Relay Status",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "derp_status",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:derp.server.status AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "WebSocket Connections - Real-time",
      "display_type": "big_number",
      "widget_type": "metrics",
      "queries": [
        {
          "name": "websocket_connections",
          "fields": ["count()"],
          "conditions": "event.type:metric AND metric:websocket.connections AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    }
  ]
}'

# Отправляем запрос на создание дашборда
echo "📤 Отправка запроса на создание дашборда..."
RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" "$DASHBOARD_DATA")

# Проверяем ответ
DASHBOARD_ID=$(echo "$RESPONSE" | jq -r '.id' 2>/dev/null)

if [ "$DASHBOARD_ID" != "null" ] && [ -n "$DASHBOARD_ID" ]; then
    echo "✅ Дашборд создан успешно!"
    echo "   ID: $DASHBOARD_ID"
    echo "   URL: https://sentry.io/organizations/$ORG_SLUG/dashboards/$DASHBOARD_ID/"
    echo ""
    echo "📊 Созданные виджеты для CloudBridge Relay:"
    echo "   ✅ Error Rate - график ошибок"
    echo "   ✅ Response Time - время ответа API"
    echo "   ✅ Throughput - пропускная способность"
    echo "   ✅ Memory Usage - использование памяти"
    echo "   ✅ CPU Usage - использование CPU"
    echo "   ✅ P2P Connections - активные P2P соединения"
    echo "   ✅ QUIC Connections - QUIC протокол"
    echo "   ✅ WireGuard Tunnels - VPN соединения"
    echo "   ✅ DERP Server Status - статус релея"
    echo "   ✅ WebSocket Connections - реальное время"
    echo ""
    echo "🎯 Дашборд готов к использованию!"
    echo ""
    echo "📋 Следующие шаги:"
    echo "   1. Откройте дашборд по ссылке выше"
    echo "   2. Настройте временной диапазон (Last 24 hours)"
    echo "   3. Включите автообновление (5 minutes)"
    echo "   4. Настройте алерты для критических метрик"
    echo "   5. Добавьте уведомления (Slack/Email)"
    echo ""
    echo "🔗 Полезные ссылки:"
    echo "   Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/$DASHBOARD_ID/"
    echo "   Issues: https://sentry.io/organizations/$ORG_SLUG/issues/"
    echo "   Performance: https://sentry.io/organizations/$ORG_SLUG/performance/"
    echo "   Alerts: https://sentry.io/organizations/$ORG_SLUG/alerts/"
else
    echo "❌ Ошибка при создании дашборда"
    echo "   Ответ: $RESPONSE"
    echo ""
    echo "🔍 Возможные причины:"
    echo "   - Недостаточно прав доступа (нужны org:write, project:write)"
    echo "   - Неверный формат данных"
    echo "   - Проблемы с API токеном"
    echo "   - Ограничения Sentry API"
    echo ""
    echo "💡 Рекомендации:"
    echo "   1. Проверьте права API токена:"
    echo "      https://sentry.io/settings/account/api/auth-tokens/"
    echo "   2. Убедитесь, что токен имеет права:"
    echo "      - org:read, org:write"
    echo "      - project:read, project:write"
    echo "      - event:read, event:write"
    echo "   3. Используйте ручную настройку:"
    echo "      ./setup-cloudbridge-sentry-dashboard.sh"
    echo ""
    echo "🔧 Альтернативное решение:"
    echo "   Используйте скрипт для ручной настройки:"
    echo "   ./setup-cloudbridge-sentry-dashboard.sh"
fi
