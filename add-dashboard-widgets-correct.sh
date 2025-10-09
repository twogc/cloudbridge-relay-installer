#!/bin/bash
# Правильное добавление виджетов в дашборд CloudBridge Relay на основе API схемы

source .env

echo "📊 CloudBridge Relay - Правильное добавление виджетов"
echo "====================================================="
echo "Organization: 2gc"
echo "Dashboard ID: 156471"
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

# Получаем текущий дашборд
echo "📋 Получение текущего дашборда..."
DASHBOARD_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/organizations/2gc/dashboards/156471/")
echo "Текущий дашборд: $(echo "$DASHBOARD_RESPONSE" | jq -r '.title')"
echo "Текущих виджетов: $(echo "$DASHBOARD_RESPONSE" | jq '.widgets | length')"
echo ""

# Создаем обновленный дашборд с виджетами
echo "📊 Добавление виджетов в дашборд..."

# Основные виджеты для CloudBridge Relay
DASHBOARD_UPDATE='{
  "title": "CloudBridge Relay - Production Dashboard",
  "projects": [4510156604702720],
  "environment": ["'$SENTRY_ENVIRONMENT'"],
  "period": "24h",
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

# Обновляем дашборд с виджетами
echo "📤 Отправка запроса на обновление дашборда..."
UPDATE_RESPONSE=$(api_request "PUT" "https://sentry.io/api/0/organizations/2gc/dashboards/156471/" "$DASHBOARD_UPDATE")

# Проверяем ответ
if echo "$UPDATE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Дашборд успешно обновлен!"
    echo "   ID: $(echo "$UPDATE_RESPONSE" | jq -r '.id')"
    echo "   Название: $(echo "$UPDATE_RESPONSE" | jq -r '.title')"
    echo "   Виджетов: $(echo "$UPDATE_RESPONSE" | jq '.widgets | length')"
    echo ""
    echo "📊 Добавленные виджеты для CloudBridge Relay:"
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
    echo "🔗 Откройте дашборд:"
    echo "   https://sentry.io/organizations/2gc/dashboards/156471/"
    echo ""
    echo "📋 Следующие шаги:"
    echo "   1. Откройте дашборд по ссылке выше"
    echo "   2. Настройте временной диапазон (Last 24 hours)"
    echo "   3. Включите автообновление (5 minutes)"
    echo "   4. Настройте дополнительные фильтры"
    echo "   5. Добавьте уведомления для критических метрик"
else
    echo "❌ Ошибка при обновлении дашборда"
    echo "   Ответ: $UPDATE_RESPONSE"
    echo ""
    echo "🔍 Возможные причины:"
    echo "   - Неверный формат данных"
    echo "   - Проблемы с API токеном"
    echo "   - Ограничения Sentry API"
    echo ""
    echo "💡 Рекомендации:"
    echo "   1. Проверьте права API токена"
    echo "   2. Убедитесь, что токен имеет права org:write"
    echo "   3. Попробуйте создать виджеты вручную через UI"
fi
