#!/bin/bash

# ===========================================
# Sentry Dashboard Management Script
# ===========================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Показать справку
show_help() {
    echo "🔧 Sentry Dashboard Management"
    echo "============================="
    echo ""
    echo "Использование: $0 [команда]"
    echo ""
    echo "Команды:"
    echo "  create     - Создать новый дашборд"
    echo "  list       - Показать все дашборды"
    echo "  delete     - Удалить дашборд"
    echo "  update     - Обновить дашборд"
    echo "  widgets    - Управление виджетами"
    echo "  alerts     - Управление алертами"
    echo "  help       - Показать эту справку"
    echo ""
    echo "Примеры:"
    echo "  $0 create"
    echo "  $0 list"
    echo "  $0 delete dashboard-id"
    echo "  $0 widgets add"
    echo "  $0 alerts create"
}

# Загрузка конфигурации
load_config() {
    if [ -f .env ]; then
        source .env
        success "Конфигурация загружена"
    else
        error ".env файл не найден"
        exit 1
    fi
}

# Получение токена доступа
get_access_token() {
    local token_response=$(curl -s -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials" \
        -d "client_id=$SENTRY_CLIENT_ID" \
        -d "client_secret=$SENTRY_CLIENT_SECRET" \
        -d "scope=project:read,project:write,event:read,event:write,org:read" \
        "$SENTRY_TOKEN_URL")
    
    echo "$token_response" | jq -r '.access_token'
}

# Создание дашборда
create_dashboard() {
    log "Создание дашборда..."
    
    local access_token=$(get_access_token)
    local org_slug=$(echo "$SENTRY_DSN" | sed 's/.*@\([^.]*\)\.ingest\.sentry\.io.*/\1/')
    
    local dashboard_data=$(cat << EOF
{
  "title": "CloudBridge Relay - $(date +%Y%m%d_%H%M%S)",
  "description": "Мониторинг CloudBridge Relay",
  "widgets": [
    {
      "title": "Error Rate",
      "displayType": "line",
      "queries": [
        {
          "name": "error_rate",
          "fields": ["count()"],
          "conditions": "event.type:error",
          "aggregations": [["count", null, "count"]]
        }
      ]
    },
    {
      "title": "Response Time",
      "displayType": "line",
      "queries": [
        {
          "name": "response_time",
          "fields": ["avg(transaction.duration)"],
          "conditions": "event.type:transaction",
          "aggregations": [["avg", "transaction.duration", "avg"]]
        }
      ]
    }
  ]
}
EOF
)
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $access_token" \
        -H "Content-Type: application/json" \
        -d "$dashboard_data" \
        "https://sentry.io/api/0/organizations/$org_slug/dashboards/")
    
    local dashboard_id=$(echo "$response" | jq -r '.id')
    
    if [ "$dashboard_id" != "null" ] && [ -n "$dashboard_id" ]; then
        success "Дашборд создан с ID: $dashboard_id"
        echo "🔗 URL: https://sentry.io/organizations/$org_slug/dashboards/$dashboard_id/"
    else
        error "Ошибка при создании дашборда"
        echo "Ответ: $response"
    fi
}

# Список дашбордов
list_dashboards() {
    log "Получение списка дашбордов..."
    
    local access_token=$(get_access_token)
    local org_slug=$(echo "$SENTRY_DSN" | sed 's/.*@\([^.]*\)\.ingest\.sentry\.io.*/\1/')
    
    local response=$(curl -s -X GET \
        -H "Authorization: Bearer $access_token" \
        "https://sentry.io/api/0/organizations/$org_slug/dashboards/")
    
    echo "$response" | jq -r '.[] | "\(.id) - \(.title) - \(.created)"'
}

# Удаление дашборда
delete_dashboard() {
    local dashboard_id="$1"
    
    if [ -z "$dashboard_id" ]; then
        error "Не указан ID дашборда"
        echo "Использование: $0 delete dashboard-id"
        exit 1
    fi
    
    log "Удаление дашборда $dashboard_id..."
    
    local access_token=$(get_access_token)
    local org_slug=$(echo "$SENTRY_DSN" | sed 's/.*@\([^.]*\)\.ingest\.sentry\.io.*/\1/')
    
    local response=$(curl -s -X DELETE \
        -H "Authorization: Bearer $access_token" \
        "https://sentry.io/api/0/organizations/$org_slug/dashboards/$dashboard_id/")
    
    if [ $? -eq 0 ]; then
        success "Дашборд $dashboard_id удален"
    else
        error "Ошибка при удалении дашборда"
    fi
}

# Управление виджетами
manage_widgets() {
    local action="$1"
    
    case "$action" in
        "add")
            log "Добавление виджета..."
            # Здесь можно добавить логику для добавления виджетов
            ;;
        "remove")
            log "Удаление виджета..."
            # Здесь можно добавить логику для удаления виджетов
            ;;
        "list")
            log "Список виджетов..."
            # Здесь можно добавить логику для показа виджетов
            ;;
        *)
            echo "Доступные действия: add, remove, list"
            ;;
    esac
}

# Управление алертами
manage_alerts() {
    local action="$1"
    
    case "$action" in
        "create")
            log "Создание алерта..."
            # Здесь можно добавить логику для создания алертов
            ;;
        "list")
            log "Список алертов..."
            # Здесь можно добавить логику для показа алертов
            ;;
        "delete")
            log "Удаление алерта..."
            # Здесь можно добавить логику для удаления алертов
            ;;
        *)
            echo "Доступные действия: create, list, delete"
            ;;
    esac
}

# Основная функция
main() {
    local command="$1"
    local arg1="$2"
    
    # Загрузка конфигурации
    load_config
    
    case "$command" in
        "create")
            create_dashboard
            ;;
        "list")
            list_dashboards
            ;;
        "delete")
            delete_dashboard "$arg1"
            ;;
        "update")
            log "Обновление дашборда..."
            # Здесь можно добавить логику для обновления
            ;;
        "widgets")
            manage_widgets "$arg1"
            ;;
        "alerts")
            manage_alerts "$arg1"
            ;;
        "help"|"")
            show_help
            ;;
        *)
            error "Неизвестная команда: $command"
            show_help
            exit 1
            ;;
    esac
}

# Запуск основной функции
main "$@"
