#!/bin/bash
# Практические скрипты для работы с Sentry API

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔧 Sentry API - Практические скрипты"
echo "==================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
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

# 1. Получение информации о проекте
get_project_info() {
    echo "📋 Информация о проекте:"
    api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/" | jq -r '.name, .slug, .platform'
}

# 2. Получение событий
get_events() {
    echo "📊 Последние события:"
    api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/?limit=5" | jq -r '.[] | "\(.id) - \(.title) - \(.timestamp)"'
}

# 3. Получение метрик
get_metrics() {
    echo "📈 Метрики производительности:"
    api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/?query=event.type:transaction&limit=10" | jq -r '.[] | "\(.id) - \(.title) - \(.timestamp)"'
}

# 4. Создание алерта
create_alert() {
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
        "frequency": 1
    }'
    
    api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/rules/" "$alert_data"
}

# 5. Получение релизов
get_releases() {
    echo "🏷️ Релизы:"
    api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/" | jq -r '.[] | "\(.version) - \(.dateCreated) - \(.status)"'
}

# 6. Создание релиза
create_release() {
    echo "🏷️ Создание нового релиза..."
    local release_data='{
        "version": "'$SENTRY_RELEASE'",
        "projects": ["'$PROJECT_ID'"]
    }'
    
    api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/" "$release_data"
}

# 7. Получение дашбордов
get_dashboards() {
    echo "📊 Дашборды:"
    api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" | jq -r '.[] | "\(.id) - \(.title) - \(.dateCreated)"'
}

# 8. Получение статистики
get_stats() {
    echo "📊 Статистика проекта:"
    echo "   - Всего событий: $(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/" | jq '. | length')"
    echo "   - Активных релизов: $(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/" | jq '. | length')"
    echo "   - Дашбордов: $(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" | jq '. | length')"
}

# Основное меню
case "$1" in
    "info")
        get_project_info
        ;;
    "events")
        get_events
        ;;
    "metrics")
        get_metrics
        ;;
    "alert")
        create_alert
        ;;
    "releases")
        get_releases
        ;;
    "create-release")
        create_release
        ;;
    "dashboards")
        get_dashboards
        ;;
    "stats")
        get_stats
        ;;
    "all")
        get_project_info
        echo ""
        get_events
        echo ""
        get_metrics
        echo ""
        get_releases
        echo ""
        get_dashboards
        echo ""
        get_stats
        ;;
    "help"|"")
        echo "🔧 Доступные команды:"
        echo "   info           - Информация о проекте"
        echo "   events         - Последние события"
        echo "   metrics        - Метрики производительности"
        echo "   alert          - Создать алерт"
        echo "   releases       - Список релизов"
        echo "   create-release - Создать релиз"
        echo "   dashboards     - Список дашбордов"
        echo "   stats          - Статистика проекта"
        echo "   all            - Выполнить все команды"
        echo ""
        echo "Примеры:"
        echo "   $0 info"
        echo "   $0 events"
        echo "   $0 metrics"
        echo "   $0 all"
        ;;
    *)
        echo "❌ Неизвестная команда: $1"
        echo "Используйте: $0 help"
        ;;
esac
