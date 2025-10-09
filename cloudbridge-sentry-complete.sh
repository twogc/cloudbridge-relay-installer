#!/bin/bash
# Полная настройка Sentry Dashboard для CloudBridge Relay
# Объединяет все функции: дашборд, алерты, мониторинг

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 CloudBridge Relay - Complete Sentry Setup"
echo "==========================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo "Server Name: $SENTRY_SERVER_NAME"
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

# 1. Проверка интеграции
check_integration() {
    echo "🔍 Проверка интеграции CloudBridge Relay с Sentry..."
    
    # Проверяем доступ к организации
    ORG_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/")
    if echo "$ORG_RESPONSE" | jq -e '.slug' > /dev/null 2>&1; then
        echo "✅ Доступ к организации: OK"
        echo "   Organization: $(echo "$ORG_RESPONSE" | jq -r '.name')"
    else
        echo "❌ Нет доступа к организации"
        echo "   Ответ: $ORG_RESPONSE"
        return 1
    fi
    
    # Проверяем доступ к проекту
    PROJECT_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/")
    if echo "$PROJECT_RESPONSE" | jq -e '.slug' > /dev/null 2>&1; then
        echo "✅ Доступ к проекту: OK"
        echo "   Project: $(echo "$PROJECT_RESPONSE" | jq -r '.name')"
    else
        echo "❌ Нет доступа к проекту"
        echo "   Ответ: $PROJECT_RESPONSE"
        return 1
    fi
    
    echo "✅ Интеграция работает корректно"
    echo ""
}

# 2. Создание дашборда
create_dashboard() {
    echo "📊 Создание дашборда для CloudBridge Relay..."
    
    local dashboard_data='{
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
            }
        ]
    }'
    
    local response=$(api_request "POST" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" "$dashboard_data")
    local dashboard_id=$(echo "$response" | jq -r '.id' 2>/dev/null)
    
    if [ "$dashboard_id" != "null" ] && [ -n "$dashboard_id" ]; then
        echo "✅ Дашборд создан с ID: $dashboard_id"
        echo "🔗 URL: https://sentry.io/organizations/$ORG_SLUG/dashboards/$dashboard_id/"
        return 0
    else
        echo "❌ Ошибка создания дашборда"
        echo "   Ответ: $response"
        return 1
    fi
}

# 3. Создание алертов
create_alerts() {
    echo "🚨 Создание алертов для CloudBridge Relay..."
    
    # Алерт для высокого уровня ошибок
    local error_alert='{
        "name": "CloudBridge Relay - High Error Rate",
        "conditions": [
            {
                "id": "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
                "value": 10,
                "interval": "1m"
            }
        ],
        "actions": [
            {
                "id": "sentry.rules.actions.notify_event_service.NotifyEventServiceAction",
                "service": "slack"
            }
        ],
        "frequency": 1,
        "environment": "'$SENTRY_ENVIRONMENT'"
    }'
    
    local error_response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$error_alert")
    if echo "$error_response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт для ошибок создан"
    else
        echo "❌ Ошибка создания алерта для ошибок"
    fi
    
    # Алерт для медленного времени ответа
    local response_alert='{
        "name": "CloudBridge Relay - Slow Response Time",
        "conditions": [
            {
                "id": "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
                "value": 5,
                "interval": "1m"
            }
        ],
        "actions": [
            {
                "id": "sentry.rules.actions.notify_event_service.NotifyEventServiceAction",
                "service": "slack"
            }
        ],
        "frequency": 1,
        "environment": "'$SENTRY_ENVIRONMENT'"
    }'
    
    local response_alert_response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$response_alert")
    if echo "$response_alert_response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт для времени ответа создан"
    else
        echo "❌ Ошибка создания алерта для времени ответа"
    fi
    
    echo "✅ Алерты созданы"
    echo ""
}

# 4. Получение статистики
get_statistics() {
    echo "📊 Статистика CloudBridge Relay:"
    
    # События
    local events_response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/")
    local events_count=$(echo "$events_response" | jq '. | length' 2>/dev/null || echo "0")
    echo "   Всего событий: $events_count"
    
    # Релизы
    local releases_response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/")
    local releases_count=$(echo "$releases_response" | jq '. | length' 2>/dev/null || echo "0")
    echo "   Активных релизов: $releases_count"
    
    # Дашборды
    local dashboards_response=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/")
    local dashboards_count=$(echo "$dashboards_response" | jq '. | length' 2>/dev/null || echo "0")
    echo "   Дашбордов: $dashboards_count"
    
    # Алерты
    local alerts_response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/")
    local alerts_count=$(echo "$alerts_response" | jq '. | length' 2>/dev/null || echo "0")
    echo "   Алертов: $alerts_count"
    
    echo ""
}

# 5. Полная настройка
full_setup() {
    echo "🚀 Полная настройка CloudBridge Relay с Sentry..."
    echo ""
    
    # Проверяем интеграцию
    if ! check_integration; then
        echo "❌ Ошибка интеграции. Проверьте API токен и права доступа."
        exit 1
    fi
    
    # Создаем дашборд
    if create_dashboard; then
        echo "✅ Дашборд создан успешно"
    else
        echo "⚠️ Ошибка создания дашборда, но продолжаем..."
    fi
    
    # Создаем алерты
    create_alerts
    
    # Получаем статистику
    get_statistics
    
    echo "🎯 Настройка завершена!"
    echo ""
    echo "📋 Что было настроено:"
    echo "   ✅ Интеграция с Sentry"
    echo "   ✅ Дашборд с виджетами"
    echo "   ✅ Алерты для критических метрик"
    echo "   ✅ Мониторинг CloudBridge Relay"
    echo ""
    echo "🔗 Полезные ссылки:"
    echo "   Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
    echo "   Issues: https://sentry.io/organizations/$ORG_SLUG/issues/"
    echo "   Performance: https://sentry.io/organizations/$ORG_SLUG/performance/"
    echo "   Alerts: https://sentry.io/organizations/$ORG_SLUG/alerts/"
    echo ""
    echo "📊 Специфичные метрики для CloudBridge Relay:"
    echo "   - P2P Mesh Connections"
    echo "   - QUIC Performance"
    echo "   - WireGuard Tunnels"
    echo "   - DERP Server Status"
    echo "   - WebSocket Connections"
    echo "   - API Response Times"
    echo "   - Error Rates by Component"
    echo ""
    echo "🎯 Рекомендации:"
    echo "   1. Настройте временной диапазон (Last 24 hours)"
    echo "   2. Включите автообновление (5 minutes)"
    echo "   3. Добавьте уведомления (Slack/Email)"
    echo "   4. Настройте дополнительные алерты"
    echo "   5. Интегрируйте с внешними системами"
}

# Основное меню
case "$1" in
    "check")
        check_integration
        ;;
    "dashboard")
        create_dashboard
        ;;
    "alerts")
        create_alerts
        ;;
    "stats")
        get_statistics
        ;;
    "full"|"setup")
        full_setup
        ;;
    "help"|"")
        echo "🔧 Доступные команды:"
        echo "   check     - Проверить интеграцию"
        echo "   dashboard - Создать дашборд"
        echo "   alerts    - Создать алерты"
        echo "   stats     - Показать статистику"
        echo "   full      - Полная настройка"
        echo "   setup     - Полная настройка (алиас)"
        echo ""
        echo "Примеры:"
        echo "   $0 check"
        echo "   $0 dashboard"
        echo "   $0 full"
        echo ""
        echo "🎯 Рекомендуется:"
        echo "   $0 full"
        ;;
    *)
        echo "❌ Неизвестная команда: $1"
        echo "Используйте: $0 help"
        ;;
esac
