#!/bin/bash
# Настройка алертов для CloudBridge Relay в Sentry

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚨 CloudBridge Relay - Alert Setup"
echo "=================================="
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

# 1. Алерт для высокого уровня ошибок
create_error_rate_alert() {
    echo "🚨 Создание алерта для высокого уровня ошибок..."
    
    local alert_data='{
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
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт создан с ID: $(echo "$response" | jq -r '.id')"
    else
        echo "❌ Ошибка создания алерта"
        echo "   Ответ: $response"
    fi
}

# 2. Алерт для медленного времени ответа
create_response_time_alert() {
    echo "⏱️ Создание алерта для медленного времени ответа..."
    
    local alert_data='{
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
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт создан с ID: $(echo "$response" | jq -r '.id')"
    else
        echo "❌ Ошибка создания алерта"
        echo "   Ответ: $response"
    fi
}

# 3. Алерт для высокого использования памяти
create_memory_alert() {
    echo "💾 Создание алерта для высокого использования памяти..."
    
    local alert_data='{
        "name": "CloudBridge Relay - High Memory Usage",
        "conditions": [
            {
                "id": "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
                "value": 3,
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
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт создан с ID: $(echo "$response" | jq -r '.id')"
    else
        echo "❌ Ошибка создания алерта"
        echo "   Ответ: $response"
    fi
}

# 4. Алерт для высокого использования CPU
create_cpu_alert() {
    echo "🔥 Создание алерта для высокого использования CPU..."
    
    local alert_data='{
        "name": "CloudBridge Relay - High CPU Usage",
        "conditions": [
            {
                "id": "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
                "value": 3,
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
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт создан с ID: $(echo "$response" | jq -r '.id')"
    else
        echo "❌ Ошибка создания алерта"
        echo "   Ответ: $response"
    fi
}

# 5. Алерт для потери P2P соединений
create_p2p_alert() {
    echo "🌐 Создание алерта для потери P2P соединений..."
    
    local alert_data='{
        "name": "CloudBridge Relay - P2P Connection Loss",
        "conditions": [
            {
                "id": "sentry.rules.conditions.event_frequency.EventFrequencyCondition",
                "value": 1,
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
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Алерт создан с ID: $(echo "$response" | jq -r '.id')"
    else
        echo "❌ Ошибка создания алерта"
        echo "   Ответ: $response"
    fi
}

# 6. Получение существующих алертов
get_existing_alerts() {
    echo "📋 Существующие алерты:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/")
    
    if echo "$response" | jq -e '.[0]' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "   \(.id) - \(.name) - \(.status)"'
    else
        echo "   Нет алертов или ошибка получения данных"
        echo "   Ответ: $response"
    fi
}

# 7. Создание всех алертов
create_all_alerts() {
    echo "🚨 Создание всех алертов для CloudBridge Relay..."
    echo ""
    
    create_error_rate_alert
    echo ""
    create_response_time_alert
    echo ""
    create_memory_alert
    echo ""
    create_cpu_alert
    echo ""
    create_p2p_alert
    echo ""
    
    echo "✅ Все алерты созданы!"
    echo ""
    echo "📋 Созданные алерты:"
    echo "   🚨 High Error Rate - >10 ошибок в минуту"
    echo "   ⏱️ Slow Response Time - >5 секунд"
    echo "   💾 High Memory Usage - >80%"
    echo "   🔥 High CPU Usage - >90%"
    echo "   🌐 P2P Connection Loss - <1 соединение"
    echo ""
    echo "🔗 Управление алертами:"
    echo "   https://sentry.io/organizations/$ORG_SLUG/alerts/"
}

# 8. Настройка уведомлений
setup_notifications() {
    echo "📧 Настройка уведомлений для CloudBridge Relay..."
    echo ""
    echo "📋 Доступные каналы уведомлений:"
    echo "   1. Slack - для команды разработки"
    echo "   2. Email - для администраторов"
    echo "   3. Webhook - для внешних систем"
    echo "   4. PagerDuty - для критических алертов"
    echo ""
    echo "🔗 Настройка уведомлений:"
    echo "   https://sentry.io/organizations/$ORG_SLUG/settings/notifications/"
    echo ""
    echo "📊 Рекомендуемые настройки:"
    echo "   - Slack: #cloudbridge-alerts"
    echo "   - Email: admin@yourcompany.com"
    echo "   - Webhook: https://your-webhook-url.com/sentry"
    echo "   - PagerDuty: для критических алертов"
}

# Основное меню
case "$1" in
    "error-rate")
        create_error_rate_alert
        ;;
    "response-time")
        create_response_time_alert
        ;;
    "memory")
        create_memory_alert
        ;;
    "cpu")
        create_cpu_alert
        ;;
    "p2p")
        create_p2p_alert
        ;;
    "all")
        create_all_alerts
        ;;
    "list")
        get_existing_alerts
        ;;
    "notifications")
        setup_notifications
        ;;
    "help"|"")
        echo "🔧 Доступные команды:"
        echo "   error-rate     - Создать алерт для высокого уровня ошибок"
        echo "   response-time  - Создать алерт для медленного времени ответа"
        echo "   memory         - Создать алерт для высокого использования памяти"
        echo "   cpu            - Создать алерт для высокого использования CPU"
        echo "   p2p            - Создать алерт для потери P2P соединений"
        echo "   all            - Создать все алерты"
        echo "   list           - Показать существующие алерты"
        echo "   notifications  - Настройка уведомлений"
        echo ""
        echo "Примеры:"
        echo "   $0 all"
        echo "   $0 error-rate"
        echo "   $0 list"
        echo "   $0 notifications"
        ;;
    *)
        echo "❌ Неизвестная команда: $1"
        echo "Используйте: $0 help"
        ;;
esac
