#!/bin/bash
set -euo pipefail

# ZITADEL Scopes and Claims Setup Script
# Скрипт для настройки custom scopes и claims в ZITADEL

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Конфигурация
ZITADEL_DOMAIN="${ZITADEL_DOMAIN:-idp.example.com}"
ZITADEL_TOKEN="${ZITADEL_TOKEN:-}"
ORG_ID="${ORG_ID:-}"
PROJECT_ID="${PROJECT_ID:-}"

# API URLs
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"
MGMT_API="$ZITADEL_API_URL/management/v1"
AUTH_API="$ZITADEL_API_URL/auth/v1"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }

# Проверка предварительных требований
check_prerequisites() {
    log "Проверка предварительных требований..."
    
    for tool in curl jq; do
        if ! command -v "$tool" &> /dev/null; then
            error "Необходимый инструмент '$tool' не найден"
            exit 1
        fi
    done
    
    if [[ -z "$ZITADEL_TOKEN" ]]; then
        error "Переменная ZITADEL_TOKEN не установлена"
        error "Получите токен доступа и установите переменную окружения"
        exit 1
    fi
}

# Получение информации об организации
get_organization_info() {
    log "Получение информации об организации..."
    
    local orgs_response
    orgs_response=$(curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
        "$MGMT_API/orgs/_search" \
        -H "Content-Type: application/json" \
        -d '{"queries": [{"name_query": {"name": "cloudbridge"}}]}')
    
    if echo "$orgs_response" | jq -e '.result[0].id' > /dev/null; then
        ORG_ID=$(echo "$orgs_response" | jq -r '.result[0].id')
        log "Найдена организация cloudbridge: $ORG_ID"
    else
        error "Организация 'cloudbridge' не найдена"
        error "Создайте организацию сначала или укажите ORG_ID вручную"
        exit 1
    fi
}

# Получение информации о проекте
get_project_info() {
    log "Получение информации о проекте..."
    
    local projects_response
    projects_response=$(curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
        "$MGMT_API/projects/_search" \
        -H "Content-Type: application/json" \
        -d '{"queries": [{"name_query": {"name": "cloudbridge-relay"}}]}')
    
    if echo "$projects_response" | jq -e '.result[0].id' > /dev/null; then
        PROJECT_ID=$(echo "$projects_response" | jq -r '.result[0].id')
        log "Найден проект cloudbridge-relay: $PROJECT_ID"
    else
        error "Проект 'cloudbridge-relay' не найден"
        error "Создайте проект сначала или укажите PROJECT_ID вручную"
        exit 1
    fi
}

# Создание custom role (scope)
create_project_role() {
    local role_key="$1"
    local display_name="$2"
    local group="$3"
    
    log "Создание роли '$role_key'..."
    
    local role_data
    role_data=$(cat <<EOF
{
  "key": "$role_key",
  "display_name": "$display_name",
  "group": "$group"
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ZITADEL_TOKEN" \
        -X POST "$MGMT_API/projects/$PROJECT_ID/roles" \
        -H "Content-Type: application/json" \
        -d "$role_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        log "✅ Роль '$role_key' создана успешно"
    elif [[ "$http_code" == "409" ]]; then
        warn "Роль '$role_key' уже существует"
    else
        error "❌ Ошибка создания роли '$role_key': HTTP $http_code"
        error "Response: $body"
    fi
}

# Создание всех необходимых ролей
create_all_roles() {
    log "Создание project roles..."
    
    # API Access Roles
    create_project_role "relay.api" "CloudBridge Relay API Access" "api"
    create_project_role "relay.k8s" "Kubernetes Proxy Access" "k8s"
    
    # User Roles
    create_project_role "cluster-admin" "Cluster Administrator" "admin"
    create_project_role "tenant-reader" "Tenant Reader" "reader"
    create_project_role "tenant-admin" "Tenant Administrator" "admin"
    create_project_role "api-consumer" "API Consumer" "api"
    create_project_role "k8s-viewer" "Kubernetes Viewer" "k8s"
    create_project_role "k8s-editor" "Kubernetes Editor" "k8s"
    
    log "Все роли созданы"
}

# Получение списка приложений в проекте
get_project_applications() {
    log "Получение списка приложений проекта..."
    
    local apps_response
    apps_response=$(curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
        "$MGMT_API/projects/$PROJECT_ID/apps/_search")
    
    if echo "$apps_response" | jq -e '.result' > /dev/null; then
        echo "$apps_response" | jq '.result[]'
    else
        error "Ошибка получения приложений: $apps_response"
        return 1
    fi
}

# Настройка JWT токенов для приложения
configure_jwt_template() {
    local app_id="$1"
    local app_name="$2"
    
    log "Настройка JWT шаблона для приложения '$app_name' ($app_id)..."
    
    # Обновление конфигурации приложения с custom claims
    local jwt_config
    jwt_config=$(cat <<EOF
{
  "access_token_type": "OIDC_TOKEN_TYPE_JWT",
  "access_token_role_assertion": true,
  "id_token_role_assertion": true,
  "id_token_userinfo_assertion": true,
  "additional_origins": ["https://b1.2gc.space", "https://*.relay.2gc.space"]
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ZITADEL_TOKEN" \
        -X PUT "$MGMT_API/projects/$PROJECT_ID/apps/oidc/$app_id" \
        -H "Content-Type: application/json" \
        -d "$jwt_config")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]]; then
        log "✅ JWT шаблон для '$app_name' обновлен"
    else
        warn "⚠️ Ошибка обновления JWT шаблона для '$app_name': HTTP $http_code"
        warn "Response: $body"
    fi
}

# Настройка token lifetimes для приложения
configure_token_lifetimes() {
    local app_id="$1"
    local app_name="$2"
    
    log "Настройка token lifetimes для приложения '$app_name'..."
    
    local lifetimes_config
    lifetimes_config=$(cat <<EOF
{
  "access_token_lifetime": "900s",
  "id_token_lifetime": "3600s",
  "refresh_token_lifetime": "28800s",
  "refresh_token_idle_expiration": "86400s"
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ZITADEL_TOKEN" \
        -X PUT "$MGMT_API/projects/$PROJECT_ID/apps/oidc/$app_id" \
        -H "Content-Type: application/json" \
        -d "$lifetimes_config")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]]; then
        log "✅ Token lifetimes для '$app_name' настроены"
    else
        warn "⚠️ Ошибка настройки token lifetimes для '$app_name': HTTP $http_code"
        warn "Response: $body"
    fi
}

# Настройка всех приложений проекта
configure_all_applications() {
    log "Настройка всех приложений проекта..."
    
    local apps
    apps=$(get_project_applications)
    
    if [[ -z "$apps" ]]; then
        warn "Приложения не найдены в проекте"
        return 0
    fi
    
    while IFS= read -r app; do
        local app_id
        local app_name
        
        app_id=$(echo "$app" | jq -r '.id // empty')
        app_name=$(echo "$app" | jq -r '.name // "unknown"')
        
        if [[ -n "$app_id" ]]; then
            configure_jwt_template "$app_id" "$app_name"
            configure_token_lifetimes "$app_id" "$app_name"
        fi
    done <<< "$apps"
}

# Создание пользователя с custom metadata
create_test_user() {
    log "Создание тестового пользователя..."
    
    local user_data
    user_data=$(cat <<EOF
{
  "user_name": "test-user",
  "profile": {
    "first_name": "Test",
    "last_name": "User",
    "display_name": "Test User"
  },
  "email": {
    "email": "test@cloudbridge.example.com",
    "is_email_verified": true
  },
  "password": {
    "password": "TestPassword123!",
    "change_required": false
  },
  "metadata": [
    {
      "key": "tenant_id",
      "value": "test-tenant-001"
    },
    {
      "key": "client_id", 
      "value": "test-client-001"
    }
  ]
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ZITADEL_TOKEN" \
        -X POST "$MGMT_API/users/human" \
        -H "Content-Type: application/json" \
        -d "$user_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        local user_id
        user_id=$(echo "$body" | jq -r '.user_id')
        log "✅ Тестовый пользователь создан: $user_id"
        
        # Назначение ролей пользователю
        assign_user_roles "$user_id"
    elif [[ "$http_code" == "409" ]]; then
        warn "Тестовый пользователь уже существует"
    else
        error "❌ Ошибка создания тестового пользователя: HTTP $http_code"
        error "Response: $body"
    fi
}

# Назначение ролей пользователю
assign_user_roles() {
    local user_id="$1"
    
    log "Назначение ролей пользователю $user_id..."
    
    local grant_data
    grant_data=$(cat <<EOF
{
  "project_id": "$PROJECT_ID",
  "role_keys": ["relay.api", "relay.k8s", "k8s-viewer", "api-consumer"]
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $ZITADEL_TOKEN" \
        -X POST "$MGMT_API/users/$user_id/grants" \
        -H "Content-Type: application/json" \
        -d "$grant_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        log "✅ Роли назначены пользователю"
    else
        warn "⚠️ Ошибка назначения ролей: HTTP $http_code"
        warn "Response: $body"
    fi
}

# Проверка настройки
verify_configuration() {
    log "Проверка настройки scopes и claims..."
    
    # Проверка ролей проекта
    log "Проверка ролей проекта..."
    local roles_response
    roles_response=$(curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
        "$MGMT_API/projects/$PROJECT_ID/roles/_search")
    
    if echo "$roles_response" | jq -e '.result' > /dev/null; then
        local roles_count
        roles_count=$(echo "$roles_response" | jq '.result | length')
        log "✅ Найдено ролей в проекте: $roles_count"
        
        # Вывод всех ролей
        echo "$roles_response" | jq -r '.result[] | "  - \(.key): \(.display_name)"'
    else
        error "❌ Ошибка проверки ролей проекта"
    fi
    
    # Проверка OIDC конфигурации
    log "Проверка OIDC конфигурации..."
    if curl -kLs "$ZITADEL_API_URL/.well-known/openid-configuration" | jq -e '.scopes_supported' > /dev/null; then
        log "✅ OIDC конфигурация доступна"
        
        # Сохранение конфигурации
        curl -kLs "$ZITADEL_API_URL/.well-known/openid-configuration" > "$PROJECT_ROOT/configs/oidc-discovery.json"
        log "OIDC discovery сохранен в configs/oidc-discovery.json"
    else
        error "❌ OIDC конфигурация недоступна"
    fi
}

# Создание примера JWT токена
create_jwt_example() {
    log "Создание примера JWT токена..."
    
    local jwt_example
    jwt_example=$(cat <<EOF
{
  "aud": ["cloudbridge-relay"],
  "iss": "$ZITADEL_API_URL",
  "sub": "test-user-id-123",
  "tenant_id": "test-tenant-001",
  "client_id": "test-client-001",
  "roles": ["relay.api", "relay.k8s", "k8s-viewer", "api-consumer"],
  "scopes": ["openid", "profile", "email", "relay.api", "relay.k8s"],
  "exp": $(date -d '+15 minutes' +%s),
  "iat": $(date +%s),
  "auth_time": $(date +%s),
  "session_id": "session-$(uuidgen 2>/dev/null || openssl rand -hex 16)",
  "audit_session_id": "audit-$(uuidgen 2>/dev/null || openssl rand -hex 16)",
  "nbf": $(date +%s),
  "jti": "jwt-$(uuidgen 2>/dev/null || openssl rand -hex 16)",
  "email": "test@cloudbridge.example.com",
  "email_verified": true,
  "name": "Test User",
  "preferred_username": "test-user"
}
EOF
)
    
    echo "$jwt_example" > "$PROJECT_ROOT/configs/jwt-token-example.json"
    log "Пример JWT токена сохранен в configs/jwt-token-example.json"
}

# Основная функция
main() {
    log "🚀 Настройка ZITADEL Scopes и Claims"
    log "📋 Конфигурация:"
    log "   ZITADEL Domain: $ZITADEL_DOMAIN"
    log "   API URL: $ZITADEL_API_URL"
    log ""
    
    check_prerequisites
    
    # Получение информации о проекте и организации
    if [[ -z "$ORG_ID" ]]; then
        get_organization_info
    fi
    
    if [[ -z "$PROJECT_ID" ]]; then
        get_project_info
    fi
    
    log "Используемые идентификаторы:"
    log "   Organization ID: $ORG_ID"
    log "   Project ID: $PROJECT_ID"
    log ""
    
    # Выполнение настройки
    create_all_roles
    configure_all_applications
    create_test_user
    verify_configuration
    create_jwt_example
    
    log "🎉 Настройка scopes и claims завершена!"
    log ""
    log "📋 Следующие шаги:"
    log "   1. Проверьте созданные роли в ZITADEL Console"
    log "   2. Настройте приложения для использования custom claims"
    log "   3. Протестируйте получение токенов с новыми claims"
    log "   4. Обновите конфигурацию CloudBridge Relay для использования новых scopes"
    log ""
    log "📄 Созданные файлы:"
    log "   - configs/oidc-discovery.json - OIDC Discovery конфигурация"
    log "   - configs/jwt-token-example.json - Пример JWT токена"
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
