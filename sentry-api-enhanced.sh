#!/bin/bash
# Улучшенные скрипты для работы с Sentry API для CloudBridge Relay

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔧 Sentry API Enhanced - CloudBridge Relay"
echo "=========================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Функция для выполнения API запросов с улучшенной обработкой ошибок
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local token="$4"
    
    # Используем переданный токен или токен по умолчанию
    local auth_token="${token:-$SENTRY_API_TOKEN}"
    
    if [ -z "$auth_token" ]; then
        echo "❌ API токен не найден"
        return 1
    fi
    
    if [ -n "$data" ]; then
        curl -s -X "$method" \
            -H "Authorization: Bearer $auth_token" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$endpoint"
    else
        curl -s -X "$method" \
            -H "Authorization: Bearer $auth_token" \
            "$endpoint"
    fi
}

# 1. Проверка токена и прав доступа
check_token_permissions() {
    echo "🔐 Проверка токена и прав доступа..."
    
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
    
    echo ""
}

# 2. Получение информации о проекте
get_project_info() {
    echo "📋 Информация о проекте CloudBridge Relay:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/")
    
    if echo "$response" | jq -e '.slug' > /dev/null 2>&1; then
        echo "   Name: $(echo "$response" | jq -r '.name')"
        echo "   Slug: $(echo "$response" | jq -r '.slug')"
        echo "   Platform: $(echo "$response" | jq -r '.platform')"
        echo "   Date Created: $(echo "$response" | jq -r '.dateCreated')"
        echo "   Status: $(echo "$response" | jq -r '.status')"
    else
        echo "❌ Ошибка получения информации о проекте"
        echo "   Ответ: $response"
    fi
}

# 3. Получение событий с фильтрацией
get_events() {
    echo "📊 Последние события CloudBridge Relay:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/?limit=10&query=environment:$SENTRY_ENVIRONMENT")
    
    if echo "$response" | jq -e '.[0]' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "   \(.id) - \(.title) - \(.timestamp)"'
    else
        echo "   Нет событий или ошибка получения данных"
        echo "   Ответ: $response"
    fi
}

# 4. Получение метрик производительности
get_metrics() {
    echo "📈 Метрики производительности:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/?query=event.type:transaction&limit=10")
    
    if echo "$response" | jq -e '.[0]' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "   \(.id) - \(.title) - \(.timestamp)"'
    else
        echo "   Нет метрик или ошибка получения данных"
        echo "   Ответ: $response"
    fi
}

# 5. Создание алерта для CloudBridge Relay
create_cloudbridge_alert() {
    echo "🚨 Создание алерта для CloudBridge Relay..."
    
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

# 6. Получение релизов
get_releases() {
    echo "🏷️ Релизы CloudBridge Relay:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/")
    
    if echo "$response" | jq -e '.[0]' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "   \(.version) - \(.dateCreated) - \(.status)"'
    else
        echo "   Нет релизов или ошибка получения данных"
        echo "   Ответ: $response"
    fi
}

# 7. Создание релиза
create_release() {
    echo "🏷️ Создание релиза CloudBridge Relay..."
    
    local release_data='{
        "version": "'$SENTRY_RELEASE'",
        "projects": ["'$PROJECT_ID'"],
        "environment": "'$SENTRY_ENVIRONMENT'"
    }'
    
    local response=$(api_request "POST" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/" "$release_data")
    
    if echo "$response" | jq -e '.version' > /dev/null 2>&1; then
        echo "✅ Релиз создан: $(echo "$response" | jq -r '.version')"
    else
        echo "❌ Ошибка создания релиза"
        echo "   Ответ: $response"
    fi
}

# 8. Получение дашбордов
get_dashboards() {
    echo "📊 Дашборды организации:"
    
    local response=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/")
    
    if echo "$response" | jq -e '.[0]' > /dev/null 2>&1; then
        echo "$response" | jq -r '.[] | "   \(.id) - \(.title) - \(.dateCreated)"'
    else
        echo "   Нет дашбордов или ошибка получения данных"
        echo "   Ответ: $response"
    fi
}

# 9. Создание дашборда для CloudBridge Relay
create_cloudbridge_dashboard() {
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
                "title": "Error Rate",
                "display_type": "line",
                "widget_type": "discover",
                "queries": [
                    {
                        "name": "error_rate",
                        "fields": ["count()"],
                        "conditions": "event.type:error AND environment:'$SENTRY_ENVIRONMENT'",
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
                        "conditions": "event.type:transaction AND environment:'$SENTRY_ENVIRONMENT'",
                        "aggregates": ["avg(transaction.duration)"]
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
                        "conditions": "event.type:metric AND metric:p2p.connections AND environment:'$SENTRY_ENVIRONMENT'",
                        "aggregates": ["count()"]
                    }
                ]
            }
        ]
    }'
    
    local response=$(api_request "POST" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" "$dashboard_data")
    
    if echo "$response" | jq -e '.id' > /dev/null 2>&1; then
        echo "✅ Дашборд создан с ID: $(echo "$response" | jq -r '.id')"
        echo "🔗 URL: https://sentry.io/organizations/$ORG_SLUG/dashboards/$(echo "$response" | jq -r '.id')/"
    else
        echo "❌ Ошибка создания дашборда"
        echo "   Ответ: $response"
    fi
}

# 10. Получение статистики
get_stats() {
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
}

# 11. Тестирование интеграции
test_integration() {
    echo "🧪 Тестирование интеграции CloudBridge Relay с Sentry..."
    
    # Проверяем доступ к API
    echo "1. Проверка API доступа..."
    check_token_permissions
    
    # Проверяем данные проекта
    echo "2. Проверка данных проекта..."
    get_project_info
    
    # Проверяем события
    echo "3. Проверка событий..."
    get_events
    
    # Проверяем метрики
    echo "4. Проверка метрик..."
    get_metrics
    
    echo ""
    echo "✅ Тестирование завершено"
}

# Основное меню
case "$1" in
    "check")
        check_token_permissions
        ;;
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
        create_cloudbridge_alert
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
    "create-dashboard")
        create_cloudbridge_dashboard
        ;;
    "stats")
        get_stats
        ;;
    "test")
        test_integration
        ;;
    "all")
        check_token_permissions
        echo ""
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
        echo "   check            - Проверить токен и права доступа"
        echo "   info             - Информация о проекте"
        echo "   events           - Последние события"
        echo "   metrics          - Метрики производительности"
        echo "   alert            - Создать алерт"
        echo "   releases         - Список релизов"
        echo "   create-release   - Создать релиз"
        echo "   dashboards       - Список дашбордов"
        echo "   create-dashboard - Создать дашборд"
        echo "   stats            - Статистика проекта"
        echo "   test             - Тестирование интеграции"
        echo "   all              - Выполнить все команды"
        echo ""
        echo "Примеры:"
        echo "   $0 check"
        echo "   $0 info"
        echo "   $0 events"
        echo "   $0 create-dashboard"
        echo "   $0 test"
        ;;
    *)
        echo "❌ Неизвестная команда: $1"
        echo "Используйте: $0 help"
        ;;
esac
