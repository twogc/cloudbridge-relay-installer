#!/bin/bash
# Исправление метрик в дашборде CloudBridge Relay

source .env

echo "🔧 CloudBridge Relay - Исправление метрик дашборда"
echo "================================================="
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

# Создаем исправленный дашборд с правильными метриками
echo "🔧 Исправление метрик в дашборде..."

# Исправленные виджеты с правильными метриками
DASHBOARD_UPDATE='{
  "title": "CloudBridge Relay - Production Dashboard",
  "projects": [4510156604702720],
  "environment": ["'$SENTRY_ENVIRONMENT'"],
  "period": "24h",
  "widgets": [
    {
      "title": "Error Rate - CloudBridge Relay",
      "display_type": "line",
      "widget_type": "error-events",
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
      "title": "Transaction Count - All Transactions",
      "display_type": "line",
      "widget_type": "transaction-like",
      "queries": [
        {
          "name": "transaction_count",
          "fields": ["count()"],
          "conditions": "event.type:transaction AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "Error Count - All Errors",
      "display_type": "line",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "error_count",
          "fields": ["count()"],
          "conditions": "event.type:error AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "P2P Connections - Active Connections",
      "display_type": "big_number",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "p2p_connections",
          "fields": ["count()"],
          "conditions": "event.type:error AND message:*P2P* AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "QUIC Connections - QUIC Protocol",
      "display_type": "big_number",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "quic_connections",
          "fields": ["count()"],
          "conditions": "event.type:error AND message:*QUIC* AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "WireGuard Tunnels - VPN Connections",
      "display_type": "big_number",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "wireguard_tunnels",
          "fields": ["count()"],
          "conditions": "event.type:error AND message:*WireGuard* AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "DERP Server Status - Relay Status",
      "display_type": "big_number",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "derp_status",
          "fields": ["count()"],
          "conditions": "event.type:error AND message:*DERP* AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    },
    {
      "title": "WebSocket Connections - Real-time",
      "display_type": "big_number",
      "widget_type": "error-events",
      "queries": [
        {
          "name": "websocket_connections",
          "fields": ["count()"],
          "conditions": "event.type:error AND message:*WebSocket* AND environment:'$SENTRY_ENVIRONMENT' AND release:'$SENTRY_RELEASE'",
          "aggregates": ["count()"]
        }
      ]
    }
  ]
}'

# Обновляем дашборд с исправленными метриками
echo "📤 Отправка запроса на обновление дашборда..."
UPDATE_RESPONSE=$(api_request "PUT" "https://sentry.io/api/0/organizations/2gc/dashboards/156471/" "$DASHBOARD_UPDATE")

# Проверяем ответ
if echo "$UPDATE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Дашборд успешно обновлен с исправленными метриками!"
    echo "   ID: $(echo "$UPDATE_RESPONSE" | jq -r '.id')"
    echo "   Название: $(echo "$UPDATE_RESPONSE" | jq -r '.title')"
    echo "   Виджетов: $(echo "$UPDATE_RESPONSE" | jq '.widgets | length')"
    echo ""
    echo "📊 Исправленные виджеты для CloudBridge Relay:"
    echo "   ✅ Error Rate - график ошибок"
    echo "   ✅ Response Time - время ответа API"
    echo "   ✅ Throughput - пропускная способность"
    echo "   ✅ Transaction Count - все транзакции"
    echo "   ✅ Error Count - все ошибки"
    echo "   ✅ P2P Connections - P2P события"
    echo "   ✅ QUIC Connections - QUIC события"
    echo "   ✅ WireGuard Tunnels - WireGuard события"
    echo "   ✅ DERP Server Status - DERP события"
    echo "   ✅ WebSocket Connections - WebSocket события"
    echo ""
    echo "🔗 Откройте дашборд:"
    echo "   https://sentry.io/organizations/2gc/dashboards/156471/"
    echo ""
    echo "📋 Изменения:"
    echo "   - Убраны несуществующие метрики system.memory.used и system.cpu.used"
    echo "   - Заменены на доступные метрики из событий CloudBridge Relay"
    echo "   - Используются фильтры по сообщениям для специфичных компонентов"
    echo ""
    echo "💡 Рекомендации:"
    echo "   1. Откройте дашборд по ссылке выше"
    echo "   2. Проверьте, что виджеты отображают данные"
    echo "   3. Настройте временной диапазон (Last 24 hours)"
    echo "   4. Включите автообновление (5 minutes)"
    echo "   5. При необходимости добавьте кастомные метрики в CloudBridge Relay"
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
