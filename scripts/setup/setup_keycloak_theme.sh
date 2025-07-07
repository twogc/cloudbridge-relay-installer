#!/bin/bash

# Скрипт для настройки кастомной темы Playa в Keycloak
# Автор: CloudBridge Relay Team
# Версия: 1.0.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Проверка переменных окружения
if [ -z "$KEYCLOAK_ADMIN" ] || [ -z "$KEYCLOAK_ADMIN_PASSWORD" ]; then
    error "KEYCLOAK_ADMIN и KEYCLOAK_ADMIN_PASSWORD должны быть установлены"
    exit 1
fi

# URL Keycloak
KEYCLOAK_URL="http://localhost:8080"
REALM_NAME="cloudbridge"

log "Начинаем настройку темы Playa в Keycloak..."

# Функция для ожидания готовности Keycloak
wait_for_keycloak() {
    log "Ожидаем готовности Keycloak..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$KEYCLOAK_URL/health" > /dev/null 2>&1; then
            log "Keycloak готов!"
            return 0
        fi
        
        info "Попытка $attempt/$max_attempts - Keycloak еще не готов..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    error "Keycloak не стал доступен за отведенное время"
    return 1
}

# Функция для получения токена администратора
get_admin_token() {
    log "Получаем токен администратора..."
    
    local token_response=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=$KEYCLOAK_ADMIN" \
        -d "password=$KEYCLOAK_ADMIN_PASSWORD" \
        -d "grant_type=password" \
        -d "client_id=admin-cli")
    
    if echo "$token_response" | grep -q "access_token"; then
        echo "$token_response" | jq -r '.access_token'
    else
        error "Не удалось получить токен администратора"
        echo "$token_response" | jq '.'
        return 1
    fi
}

# Функция для создания realm, если он не существует
create_realm_if_not_exists() {
    local token=$1
    
    log "Проверяем существование realm '$REALM_NAME'..."
    
    local realm_exists=$(curl -s -H "Authorization: Bearer $token" \
        "$KEYCLOAK_URL/admin/realms/$REALM_NAME" | jq -r '.realm // empty')
    
    if [ -z "$realm_exists" ]; then
        log "Создаем realm '$REALM_NAME'..."
        
        curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "{
                \"realm\": \"$REALM_NAME\",
                \"enabled\": true,
                \"displayName\": \"CloudBridge Relay\",
                \"displayNameHtml\": \"<div class=\"kc-logo-text\"><span>CloudBridge</span> Relay</div>\"
            }"
        
        log "Realm '$REALM_NAME' создан"
    else
        log "Realm '$REALM_NAME' уже существует"
    fi
}

# Функция для настройки темы
setup_theme() {
    local token=$1
    
    log "Настраиваем тему 'playa' для realm '$REALM_NAME'..."
    
    # Обновляем настройки realm для использования темы
    curl -s -X PUT "$KEYCLOAK_URL/admin/realms/$REALM_NAME" \
        -H "Authorization: Bearer $token" \
        -H "Content-Type: application/json" \
        -d "{
            \"realm\": \"$REALM_NAME\",
            \"enabled\": true,
            \"displayName\": \"CloudBridge Relay\",
            \"displayNameHtml\": \"<div class=\"kc-logo-text\"><span>CloudBridge</span> Relay</div>\",
            \"loginTheme\": \"playa\",
            \"accountTheme\": \"playa\",
            \"adminTheme\": \"playa\",
            \"emailTheme\": \"playa\"
        }"
    
    log "Тема 'playa' настроена для realm '$REALM_NAME'"
}

# Функция для создания клиента, если он не существует
create_client_if_not_exists() {
    local token=$1
    local client_id="relay-client"
    
    log "Проверяем существование клиента '$client_id'..."
    
    local client_exists=$(curl -s -H "Authorization: Bearer $token" \
        "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$client_id" | jq -r '.[0].id // empty')
    
    if [ -z "$client_exists" ]; then
        log "Создаем клиент '$client_id'..."
        
        local client_response=$(curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients" \
            -H "Authorization: Bearer $token" \
            -H "Content-Type: application/json" \
            -d "{
                \"clientId\": \"$client_id\",
                \"enabled\": true,
                \"publicClient\": false,
                \"standardFlowEnabled\": true,
                \"directAccessGrantsEnabled\": true,
                \"serviceAccountsEnabled\": true,
                \"redirectUris\": [\"http://localhost:8082/*\", \"https://edge.2gc.ru/*\"],
                \"webOrigins\": [\"http://localhost:8082\", \"https://edge.2gc.ru\"],
                \"attributes\": {
                    \"saml.assertion.signature\": \"false\",
                    \"saml.force.post.binding\": \"false\",
                    \"saml.multivalued.roles\": \"false\",
                    \"saml.encrypt\": \"false\",
                    \"saml.server.signature\": \"false\",
                    \"saml.server.signature.keyinfo.ext\": \"false\",
                    \"exclude.session.state.from.auth.response\": \"false\",
                    \"saml_force_name_id_format\": \"false\",
                    \"saml.client.signature\": \"false\",
                    \"tls.client.certificate.bound.access.tokens\": \"false\",
                    \"saml.authnstatement\": \"false\",
                    \"display.on.consent.screen\": \"false\",
                    \"saml.onetimeuse.condition\": \"false\"
                }
            }")
        
        if echo "$client_response" | grep -q "id"; then
            log "Клиент '$client_id' создан"
        else
            warn "Возможно, клиент '$client_id' уже существует или произошла ошибка"
            echo "$client_response" | jq '.'
        fi
    else
        log "Клиент '$client_id' уже существует"
    fi
}

# Основная функция
main() {
    log "=== Настройка темы Playa для Keycloak ==="
    
    # Ожидаем готовности Keycloak
    wait_for_keycloak
    
    # Получаем токен администратора
    local admin_token=$(get_admin_token)
    if [ $? -ne 0 ]; then
        error "Не удалось получить токен администратора"
        exit 1
    fi
    
    # Создаем realm, если не существует
    create_realm_if_not_exists "$admin_token"
    
    # Настраиваем тему
    setup_theme "$admin_token"
    
    # Создаем клиент, если не существует
    create_client_if_not_exists "$admin_token"
    
    log "=== Настройка темы Playa завершена успешно ==="
    log "Тема 'playa' теперь доступна в Keycloak Admin Console"
    log "Для применения изменений может потребоваться перезапуск Keycloak"
}

# Запуск основной функции
main "$@"

# Удаляем строки, связанные с автоматическим перезапуском контейнера Keycloak
grep -vE 'docker (restart|stop|start) keycloak' "$0" > "$0.tmp" && mv "$0.tmp" "$0" 