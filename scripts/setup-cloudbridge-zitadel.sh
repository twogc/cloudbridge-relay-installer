#!/bin/bash
set -euo pipefail

# CloudBridge ZITADEL Organization and Project Setup
# Скрипт для создания организации cloudbridge и проекта cloudbridge-relay в существующем ZITADEL

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Конфигурация
ZITADEL_DOMAIN="${ZITADEL_DOMAIN:-edge.2gc.ru}"
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"
ORG_NAME="cloudbridge"
PROJECT_NAME="cloudbridge-relay"

# Admin credentials (из ZITADEL_SETUP_COMPLETE.md)
ADMIN_USERNAME="${ADMIN_USERNAME:-zitadel-admin@zitadel.edge.2gc.ru}"
ADMIN_PASSWORD="${ADMIN_PASSWORD:-Password1!}"

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

# Получение admin токена
get_admin_token() {
    log "Получение admin токена..."
    
    # Попытка получить токен с admin credentials
    local token_response
    token_response=$(curl -s -X POST "$ZITADEL_API_URL/oauth/v2/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "username=$ADMIN_USERNAME" \
        -d "password=$ADMIN_PASSWORD" \
        -d "scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud" 2>/dev/null || echo '{}')
    
    local access_token
    access_token=$(echo "$token_response" | jq -r '.access_token // empty' 2>/dev/null)
    
    if [[ -n "$access_token" && "$access_token" != "null" ]]; then
        echo "$access_token"
        return 0
    fi
    
    # Если прямая аутентификация не удалась, пробуем другие методы
    warn "Не удалось получить токен с admin credentials"
    
    # Проверим наличие machine user ключа
    if [[ -f "$HOME/.zitadel/machine-key.json" ]]; then
        log "Найден machine user ключ, пробуем получить токен..."
        
        # TODO: Реализовать получение токена через JWT Bearer с machine key
        warn "Получение токена через machine key пока не реализовано"
    fi
    
    # Если автоматическое получение не удалось
    error "Не удалось автоматически получить admin токен"
    error "Возможные решения:"
    error "1. Убедитесь, что admin credentials корректны"
    error "2. Получите токен вручную через веб-интерфейс"
    error "3. Создайте machine user и его JWT key"
    error ""
    error "Для ручного получения токена:"
    error "1. Войдите в ZITADEL Console: $ZITADEL_API_URL"
    error "2. Создайте Personal Access Token"
    error "3. Установите переменную: export ZITADEL_TOKEN='your_token'"
    error "4. Перезапустите скрипт"
    
    return 1
}

# Проверка существующей организации
check_existing_org() {
    local token="$1"
    
    log "Проверка существующих организаций..."
    
    local orgs_response
    orgs_response=$(curl -s -H "Authorization: Bearer $token" \
        "$ZITADEL_API_URL/management/v1/orgs/_search" \
        -H "Content-Type: application/json" \
        -d '{"queries": []}' 2>/dev/null || echo '{}')
    
    if echo "$orgs_response" | jq -e '.result' > /dev/null 2>&1; then
        log "Найденные организации:"
        echo "$orgs_response" | jq -r '.result[]? | "  - \(.name) (ID: \(.id))"' 2>/dev/null || echo "  Нет организаций"
        
        # Проверим, есть ли уже cloudbridge
        local cloudbridge_org
        cloudbridge_org=$(echo "$orgs_response" | jq -r '.result[]? | select(.name=="cloudbridge") | .id' 2>/dev/null)
        
        if [[ -n "$cloudbridge_org" && "$cloudbridge_org" != "null" ]]; then
            log "Организация 'cloudbridge' уже существует: $cloudbridge_org"
            echo "$cloudbridge_org"
            return 0
        fi
    fi
    
    return 1
}

# Создание организации
create_organization() {
    local token="$1"
    
    log "Создание организации '$ORG_NAME'..."
    
    local org_data
    org_data=$(cat <<EOF
{
  "name": "$ORG_NAME",
  "domain": "$ORG_NAME.cloudbridge.local"
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" \
        -X POST "$ZITADEL_API_URL/management/v1/orgs" \
        -H "Content-Type: application/json" \
        -d "$org_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        local org_id
        org_id=$(echo "$body" | jq -r '.id // empty')
        if [[ -n "$org_id" ]]; then
            log "Организация '$ORG_NAME' создана успешно: $org_id"
            echo "$org_id"
            return 0
        fi
    fi
    
    error "Ошибка создания организации: HTTP $http_code"
    error "Response: $body"
    return 1
}

# Проверка существующих проектов
check_existing_project() {
    local token="$1"
    
    log "Проверка существующих проектов..."
    
    local projects_response
    projects_response=$(curl -s -H "Authorization: Bearer $token" \
        "$ZITADEL_API_URL/management/v1/projects/_search" \
        -H "Content-Type: application/json" \
        -d '{"queries": []}' 2>/dev/null || echo '{}')
    
    if echo "$projects_response" | jq -e '.result' > /dev/null 2>&1; then
        log "Найденные проекты:"
        echo "$projects_response" | jq -r '.result[]? | "  - \(.name) (ID: \(.id))"' 2>/dev/null || echo "  Нет проектов"
        
        # Проверим, есть ли уже cloudbridge-relay
        local relay_project
        relay_project=$(echo "$projects_response" | jq -r '.result[]? | select(.name=="cloudbridge-relay") | .id' 2>/dev/null)
        
        if [[ -n "$relay_project" && "$relay_project" != "null" ]]; then
            log "Проект 'cloudbridge-relay' уже существует: $relay_project"
            echo "$relay_project"
            return 0
        fi
    fi
    
    return 1
}

# Создание проекта
create_project() {
    local token="$1"
    
    log "Создание проекта '$PROJECT_NAME'..."
    
    local project_data
    project_data=$(cat <<EOF
{
  "name": "$PROJECT_NAME",
  "project_role_assertion": true,
  "project_role_check": true,
  "has_project_check": true,
  "private_labeling_setting": "PRIVATE_LABELING_SETTING_ALLOW_LOGIN_USER_RESOURCE_OWNER_POLICY"
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" \
        -X POST "$ZITADEL_API_URL/management/v1/projects" \
        -H "Content-Type: application/json" \
        -d "$project_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        local project_id
        project_id=$(echo "$body" | jq -r '.id // empty')
        if [[ -n "$project_id" ]]; then
            log "Проект '$PROJECT_NAME' создан успешно: $project_id"
            echo "$project_id"
            return 0
        fi
    fi
    
    error "Ошибка создания проекта: HTTP $http_code"
    error "Response: $body"
    return 1
}

# Создание Back-channel OIDC приложения
create_backend_app() {
    local token="$1"
    local project_id="$2"
    
    log "Создание Back-channel OIDC приложения..."
    
    local app_data
    app_data=$(cat <<EOF
{
  "name": "cloudbridge-relay-api",
  "redirect_uris": [
    "https://b1.2gc.space/auth/callback",
    "https://relay.2gc.space/auth/callback"
  ],
  "response_types": [
    "OIDC_RESPONSE_TYPE_CODE"
  ],
  "grant_types": [
    "OIDC_GRANT_TYPE_CLIENT_CREDENTIALS",
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
    "OIDC_GRANT_TYPE_TOKEN_EXCHANGE"
  ],
  "app_type": "OIDC_APP_TYPE_NATIVE",
  "auth_method_type": "OIDC_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT",
  "version": "OIDC_VERSION_1_0",
  "dev_mode": false,
  "access_token_type": "OIDC_TOKEN_TYPE_JWT",
  "access_token_role_assertion": true,
  "id_token_role_assertion": true,
  "id_token_userinfo_assertion": true,
  "clock_skew": "0s",
  "additional_origins": [
    "https://b1.2gc.space",
    "https://*.relay.2gc.space"
  ]
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" \
        -X POST "$ZITADEL_API_URL/management/v1/projects/$project_id/apps/oidc" \
        -H "Content-Type: application/json" \
        -d "$app_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        local app_id
        local client_id
        local client_secret
        
        app_id=$(echo "$body" | jq -r '.app_id // empty')
        client_id=$(echo "$body" | jq -r '.client_id // empty')
        client_secret=$(echo "$body" | jq -r '.client_secret // empty')
        
        log "Back-channel приложение создано:"
        log "   App ID: $app_id"
        log "   Client ID: $client_id"
        log "   Client Secret: $client_secret"
        
        # Сохранение в файл конфигурации
        cat > "$PROJECT_ROOT/configs/backend-app-config.json" <<EOF
{
  "app_id": "$app_id",
  "client_id": "$client_id", 
  "client_secret": "$client_secret",
  "name": "cloudbridge-relay-api",
  "type": "backend",
  "redirect_uris": [
    "https://b1.2gc.space/auth/callback",
    "https://relay.2gc.space/auth/callback"
  ],
  "grant_types": [
    "client_credentials",
    "authorization_code", 
    "token_exchange"
  ]
}
EOF
        
        echo "$app_id"
        return 0
    fi
    
    error "Ошибка создания Back-channel приложения: HTTP $http_code"
    error "Response: $body"
    return 1
}

# Создание Front-channel SPA приложения
create_frontend_app() {
    local token="$1"
    local project_id="$2"
    
    log "Создание Front-channel SPA приложения..."
    
    local app_data
    app_data=$(cat <<EOF
{
  "name": "cloudbridge-relay-ui",
  "redirect_uris": [
    "http://localhost:8080/auth/callback",
    "https://relay.2gc.space/auth/callback",
    "https://b1.2gc.space/auth/callback",
    "urn:ietf:wg:oauth:2.0:oob"
  ],
  "response_types": [
    "OIDC_RESPONSE_TYPE_CODE"
  ],
  "grant_types": [
    "OIDC_GRANT_TYPE_AUTHORIZATION_CODE"
  ],
  "app_type": "OIDC_APP_TYPE_USER_AGENT",
  "auth_method_type": "OIDC_AUTH_METHOD_TYPE_NONE",
  "version": "OIDC_VERSION_1_0",
  "dev_mode": false,
  "access_token_type": "OIDC_TOKEN_TYPE_JWT",
  "access_token_role_assertion": true,
  "id_token_role_assertion": true,
  "id_token_userinfo_assertion": true,
  "skip_native_app_success_page": true,
  "additional_origins": [
    "https://relay.2gc.space",
    "https://b1.2gc.space"
  ]
}
EOF
)
    
    local response
    response=$(curl -s -w "%{http_code}" -H "Authorization: Bearer $token" \
        -X POST "$ZITADEL_API_URL/management/v1/projects/$project_id/apps/oidc" \
        -H "Content-Type: application/json" \
        -d "$app_data")
    
    local http_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        local app_id
        local client_id
        
        app_id=$(echo "$body" | jq -r '.app_id // empty')
        client_id=$(echo "$body" | jq -r '.client_id // empty')
        
        log "Front-channel приложение создано:"
        log "   App ID: $app_id"
        log "   Client ID: $client_id"
        
        # Сохранение в файл конфигурации
        cat > "$PROJECT_ROOT/configs/frontend-app-config.json" <<EOF
{
  "app_id": "$app_id",
  "client_id": "$client_id",
  "name": "cloudbridge-relay-ui",
  "type": "frontend",
  "redirect_uris": [
    "http://localhost:8080/auth/callback",
    "https://relay.2gc.space/auth/callback", 
    "https://b1.2gc.space/auth/callback",
    "urn:ietf:wg:oauth:2.0:oob"
  ],
  "grant_types": [
    "authorization_code"
  ]
}
EOF
        
        echo "$app_id"
        return 0
    fi
    
    error "Ошибка создания Front-channel приложения: HTTP $http_code"
    error "Response: $body"
    return 1
}

# Проверка развертывания
verify_setup() {
    local org_id="$1"
    local project_id="$2"
    
    log "Проверка созданной конфигурации..."
    
    # Обновление OIDC конфигурации
    if curl -k -s "$ZITADEL_API_URL/.well-known/openid-configuration" > "$PROJECT_ROOT/configs/updated-oidc-config.json"; then
        log "OIDC конфигурация обновлена"
    fi
    
    # Создание сводного файла конфигурации
    cat > "$PROJECT_ROOT/configs/cloudbridge-zitadel-setup.json" <<EOF
{
  "zitadel_domain": "$ZITADEL_DOMAIN",
  "issuer": "$ZITADEL_API_URL",
  "organization": {
    "id": "$org_id",
    "name": "$ORG_NAME"
  },
  "project": {
    "id": "$project_id", 
    "name": "$PROJECT_NAME"
  },
  "setup_date": "$(date -Iseconds)",
  "status": "completed"
}
EOF
    
    log "Базовая настройка CloudBridge ZITADEL завершена!"
    log ""
    log "Созданные ресурсы:"
    log "   • Организация: $ORG_NAME ($org_id)"
    log "   • Проект: $PROJECT_NAME ($project_id)"
    log "   • Back-channel приложение: cloudbridge-relay-api"
    log "   • Front-channel приложение: cloudbridge-relay-ui"
    log ""
    log " Конфигурационные файлы:"
    log "   • configs/cloudbridge-zitadel-setup.json - основная конфигурация"
    log "   • configs/backend-app-config.json - настройки API приложения"
    log "   • configs/frontend-app-config.json - настройки UI приложения"
    log "   • configs/updated-oidc-config.json - обновленная OIDC конфигурация"
    log ""
    log "Полезные ссылки:"
    log "   • ZITADEL Console: $ZITADEL_API_URL"
    log "   • OIDC Discovery: $ZITADEL_API_URL/.well-known/openid-configuration"
    log "   • JWKS: $ZITADEL_API_URL/oauth/v2/keys"
    log ""
    log "📝 Следующие шаги:"
    log "   1. Настройка custom scopes: ./scripts/setup-zitadel-scopes.sh"
    log "   2. Создание service account: ./scripts/setup-service-account.sh"
    log "   3. Настройка mTLS: ./scripts/setup-mtls.sh"
}

# Основная функция
main() {
    log "Настройка CloudBridge ZITADEL Organization & Project"
    log "Конфигурация:"
    log "   ZITADEL Domain: $ZITADEL_DOMAIN"
    log "   Organization: $ORG_NAME"
    log "   Project: $PROJECT_NAME"
    log ""
    
    # Получение admin токена
    local token
    if [[ -n "${ZITADEL_TOKEN:-}" ]]; then
        token="$ZITADEL_TOKEN"
        log "Используется предоставленный ZITADEL_TOKEN"
    else
        if ! token=$(get_admin_token); then
            return 1
        fi
    fi
    
    # Проверка/создание организации
    local org_id
    if ! org_id=$(check_existing_org "$token"); then
        if ! org_id=$(create_organization "$token"); then
            error "Не удалось создать организацию"
            return 1
        fi
    fi
    
    # Проверка/создание проекта
    local project_id
    if ! project_id=$(check_existing_project "$token"); then
        if ! project_id=$(create_project "$token"); then
            error "Не удалось создать проект" 
            return 1
        fi
    fi
    
    # Создание приложений
    if ! create_backend_app "$token" "$project_id"; then
        warn "Не удалось создать back-channel приложение"
    fi
    
    if ! create_frontend_app "$token" "$project_id"; then
        warn "Не удалось создать front-channel приложение"
    fi
    
    # Проверка и сохранение конфигурации
    verify_setup "$org_id" "$project_id"
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
