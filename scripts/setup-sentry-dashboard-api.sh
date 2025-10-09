#!/bin/bash

# ===========================================
# Sentry Dashboard API Setup Script
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

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    if ! command -v curl &> /dev/null; then
        error "curl не установлен"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq не установлен"
        exit 1
    fi
    
    success "Все зависимости установлены"
}

# Загрузка конфигурации
load_config() {
    log "Загрузка конфигурации..."
    
    # Загружаем переменные из .env файла
    if [ -f .env ]; then
        source .env
        success "Конфигурация загружена из .env файла"
    else
        error ".env файл не найден"
        exit 1
    fi
    
    # Проверяем обязательные переменные
    if [ -z "$SENTRY_DSN" ]; then
        error "SENTRY_DSN не установлен в .env файле"
        exit 1
    fi
    
    if [ -z "$SENTRY_CLIENT_ID" ]; then
        error "SENTRY_CLIENT_ID не установлен в .env файле"
        exit 1
    fi
    
    if [ -z "$SENTRY_CLIENT_SECRET" ]; then
        error "SENTRY_CLIENT_SECRET не установлен в .env файле"
        exit 1
    fi
    
    success "Все обязательные переменные настроены"
}

# Получение токена доступа
get_access_token() {
    log "Получение токена доступа..."
    
    local token_response=$(curl -s -X POST \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials" \
        -d "client_id=$SENTRY_CLIENT_ID" \
        -d "client_secret=$SENTRY_CLIENT_SECRET" \
        -d "scope=project:read,project:write,event:read,event:write,org:read" \
        "$SENTRY_TOKEN_URL")
    
    if [ $? -ne 0 ]; then
        error "Ошибка при получении токена доступа"
        exit 1
    fi
    
    # Извлекаем токен из ответа
    local access_token=$(echo "$token_response" | jq -r '.access_token')
    
    if [ "$access_token" = "null" ] || [ -z "$access_token" ]; then
        error "Не удалось получить токен доступа"
        echo "Ответ: $token_response"
        exit 1
    fi
    
    success "Токен доступа получен"
    echo "$access_token"
}

# Создание дашборда
create_dashboard() {
    local access_token="$1"
    
    log "Создание дашборда..."
    
    # Извлекаем organization slug из DSN
    local org_slug=$(echo "$SENTRY_DSN" | sed 's/.*@\([^.]*\)\.ingest\.sentry\.io.*/\1/')
    local project_id=$(echo "$SENTRY_DSN" | sed 's/.*ingest\.sentry\.io\/\([0-9]*\).*/\1/')
    
    log "Organization: $org_slug"
    log "Project ID: $project_id"
    
    # Создаем дашборд
    local dashboard_data=$(cat << EOF
{
  "title": "CloudBridge Relay Dashboard",
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
    },
    {
      "title": "Throughput",
      "displayType": "line",
      "queries": [
        {
          "name": "throughput",
          "fields": ["count()"],
          "conditions": "event.type:transaction",
          "aggregations": [["count", null, "count"]]
        }
      ]
    }
  ]
}
EOF
)
    
    # Отправляем запрос на создание дашборда
    local dashboard_response=$(curl -s -X POST \
        -H "Authorization: Bearer $access_token" \
        -H "Content-Type: application/json" \
        -d "$dashboard_data" \
        "https://sentry.io/api/0/organizations/$org_slug/dashboards/")
    
    if [ $? -ne 0 ]; then
        error "Ошибка при создании дашборда"
        exit 1
    fi
    
    # Проверяем ответ
    local dashboard_id=$(echo "$dashboard_response" | jq -r '.id')
    
    if [ "$dashboard_id" = "null" ] || [ -z "$dashboard_id" ]; then
        error "Не удалось создать дашборд"
        echo "Ответ: $dashboard_response"
        exit 1
    fi
    
    success "Дашборд создан с ID: $dashboard_id"
    echo "$dashboard_id"
}

# Создание алертов
create_alerts() {
    local access_token="$1"
    local org_slug="$2"
    local project_id="$3"
    
    log "Создание алертов..."
    
    # Алерт на высокий уровень ошибок
    local error_alert=$(cat << EOF
{
  "name": "High Error Rate Alert",
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
  "projects": ["$project_id"]
}
EOF
)
    
    # Создаем алерт
    local alert_response=$(curl -s -X POST \
        -H "Authorization: Bearer $access_token" \
        -H "Content-Type: application/json" \
        -d "$error_alert" \
        "https://sentry.io/api/0/projects/$org_slug/cloudbridge-relay/rules/")
    
    if [ $? -eq 0 ]; then
        success "Алерт создан"
    else
        warning "Не удалось создать алерт (возможно, не настроен Slack)"
    fi
}

# Основная функция
main() {
    echo "🚀 Sentry Dashboard API Setup"
    echo "============================="
    echo ""
    
    # Проверка зависимостей
    check_dependencies
    
    # Загрузка конфигурации
    load_config
    
    # Получение токена доступа
    local access_token=$(get_access_token)
    
    # Создание дашборда
    local dashboard_id=$(create_dashboard "$access_token")
    
    # Извлекаем organization slug
    local org_slug=$(echo "$SENTRY_DSN" | sed 's/.*@\([^.]*\)\.ingest\.sentry\.io.*/\1/')
    local project_id=$(echo "$SENTRY_DSN" | sed 's/.*ingest\.sentry\.io\/\([0-9]*\).*/\1/')
    
    # Создание алертов
    create_alerts "$access_token" "$org_slug" "$project_id"
    
    echo ""
    success "Настройка Sentry Dashboard через API завершена!"
    echo ""
    echo "📋 Создано:"
    echo "   - Дашборд: CloudBridge Relay Dashboard"
    echo "   - Алерты: High Error Rate Alert"
    echo ""
    echo "🔗 Sentry Dashboard: https://sentry.io/organizations/$org_slug/dashboards/"
    echo ""
    echo "📊 Дашборд включает:"
    echo "   - Error Rate (график ошибок)"
    echo "   - Response Time (время ответа)"
    echo "   - Throughput (пропускная способность)"
}

# Запуск основной функции
main "$@"
