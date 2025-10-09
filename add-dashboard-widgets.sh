#!/bin/bash
# Добавление виджетов в дашборд CloudBridge Relay

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "📊 CloudBridge Relay - Добавление виджетов в дашборд"
echo "=================================================="
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

# Получаем ID дашборда
DASHBOARD_ID="156471"
echo "📋 Дашборд ID: $DASHBOARD_ID"

# 1. Виджет для Error Rate
echo "📊 Добавление виджета Error Rate..."
ERROR_WIDGET='{
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
}'

ERROR_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$ERROR_WIDGET")
if echo "$ERROR_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет Error Rate добавлен"
else
    echo "❌ Ошибка добавления виджета Error Rate"
    echo "   Ответ: $ERROR_RESPONSE"
fi

# 2. Виджет для Response Time
echo "📊 Добавление виджета Response Time..."
RESPONSE_WIDGET='{
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
}'

RESPONSE_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$RESPONSE_WIDGET")
if echo "$RESPONSE_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет Response Time добавлен"
else
    echo "❌ Ошибка добавления виджета Response Time"
    echo "   Ответ: $RESPONSE_RESPONSE"
fi

# 3. Виджет для Throughput
echo "📊 Добавление виджета Throughput..."
THROUGHPUT_WIDGET='{
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
}'

THROUGHPUT_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$THROUGHPUT_WIDGET")
if echo "$THROUGHPUT_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет Throughput добавлен"
else
    echo "❌ Ошибка добавления виджета Throughput"
    echo "   Ответ: $THROUGHPUT_RESPONSE"
fi

# 4. Виджет для Memory Usage
echo "📊 Добавление виджета Memory Usage..."
MEMORY_WIDGET='{
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
}'

MEMORY_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$MEMORY_WIDGET")
if echo "$MEMORY_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет Memory Usage добавлен"
else
    echo "❌ Ошибка добавления виджета Memory Usage"
    echo "   Ответ: $MEMORY_RESPONSE"
fi

# 5. Виджет для CPU Usage
echo "📊 Добавление виджета CPU Usage..."
CPU_WIDGET='{
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
}'

CPU_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$CPU_WIDGET")
if echo "$CPU_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет CPU Usage добавлен"
else
    echo "❌ Ошибка добавления виджета CPU Usage"
    echo "   Ответ: $CPU_RESPONSE"
fi

# 6. Виджет для P2P Connections
echo "📊 Добавление виджета P2P Connections..."
P2P_WIDGET='{
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
}'

P2P_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$P2P_WIDGET")
if echo "$P2P_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет P2P Connections добавлен"
else
    echo "❌ Ошибка добавления виджета P2P Connections"
    echo "   Ответ: $P2P_RESPONSE"
fi

# 7. Виджет для QUIC Connections
echo "📊 Добавление виджета QUIC Connections..."
QUIC_WIDGET='{
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
}'

QUIC_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$QUIC_WIDGET")
if echo "$QUIC_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет QUIC Connections добавлен"
else
    echo "❌ Ошибка добавления виджета QUIC Connections"
    echo "   Ответ: $QUIC_RESPONSE"
fi

# 8. Виджет для WireGuard Tunnels
echo "📊 Добавление виджета WireGuard Tunnels..."
WIREGUARD_WIDGET='{
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
}'

WIREGUARD_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$WIREGUARD_WIDGET")
if echo "$WIREGUARD_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет WireGuard Tunnels добавлен"
else
    echo "❌ Ошибка добавления виджета WireGuard Tunnels"
    echo "   Ответ: $WIREGUARD_RESPONSE"
fi

# 9. Виджет для DERP Server Status
echo "📊 Добавление виджета DERP Server Status..."
DERP_WIDGET='{
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
}'

DERP_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$DERP_WIDGET")
if echo "$DERP_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет DERP Server Status добавлен"
else
    echo "❌ Ошибка добавления виджета DERP Server Status"
    echo "   Ответ: $DERP_RESPONSE"
fi

# 10. Виджет для WebSocket Connections
echo "📊 Добавление виджета WebSocket Connections..."
WEBSOCKET_WIDGET='{
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
}'

WEBSOCKET_RESPONSE=$(api_request "POST" "https://sentry.io/api/0/organizations/2gc/dashboards/$DASHBOARD_ID/widgets/" "$WEBSOCKET_WIDGET")
if echo "$WEBSOCKET_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ Виджет WebSocket Connections добавлен"
else
    echo "❌ Ошибка добавления виджета WebSocket Connections"
    echo "   Ответ: $WEBSOCKET_RESPONSE"
fi

echo ""
echo "🎯 Добавление виджетов завершено!"
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
echo "   https://sentry.io/organizations/2gc/dashboards/$DASHBOARD_ID/"
echo ""
echo "📋 Следующие шаги:"
echo "   1. Откройте дашборд по ссылке выше"
echo "   2. Настройте временной диапазон (Last 24 hours)"
echo "   3. Включите автообновление (5 minutes)"
echo "   4. Настройте дополнительные фильтры"
echo "   5. Добавьте уведомления для критических метрик"
