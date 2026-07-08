#!/bin/bash
set -euo pipefail

# Automatic ZITADEL Token Retrieval
# Попытка автоматического получения токена для ZITADEL API

ZITADEL_DOMAIN="edge.2gc.ru"
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"

# Admin credentials
ADMIN_EMAIL="admin@local.test"
ADMIN_PASSWORD="jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }

# Попытка получения токена через Device Code Flow с известным client_id
try_device_flow_with_console() {
    log "Попытка получения токена через Device Code Flow..."
    
    # Используем известный client_id консоли ZITADEL
    local client_id="zitadel-console"
    
    # Запрос device authorization
    local device_response
    device_response=$(curl -s -k -X POST "$ZITADEL_API_URL/oauth/v2/device_authorization" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "client_id=$client_id" \
        -d "scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud" 2>/dev/null || echo '{}')
    
    local device_code=$(echo "$device_response" | jq -r '.device_code // empty')
    local user_code=$(echo "$device_response" | jq -r '.user_code // empty')
    local verification_uri=$(echo "$device_response" | jq -r '.verification_uri // empty')
    
    if [[ -z "$device_code" ]]; then
        warn "Device Code Flow не удался: $(echo "$device_response" | jq -r '.error_description // "unknown error"')"
        return 1
    fi
    
    info "Device Code Flow инициирован:"
    info "   Код устройства: $user_code"
    info "   URL для авторизации: $verification_uri"
    info ""
    info "Для получения токена:"
    info "1. Откройте: $verification_uri"
    info "2. Войдите как: $ADMIN_EMAIL"
    info "3. Введите код: $user_code"
    info ""
    
    return 0
}

# Попытка найти существующие приложения
discover_applications() {
    log "Поиск существующих OIDC приложений..."
    
    # Попробуем известные client_id
    local known_clients=(
        "zitadel-console"
        "console"  
        "zitadel"
        "cloudbridge-relay-api"
        "cloudbridge-relay-ui"
    )
    
    for client_id in "${known_clients[@]}"; do
        info "Проверка client_id: $client_id"
        
        local test_response
        test_response=$(curl -s -k -X POST "$ZITADEL_API_URL/oauth/v2/device_authorization" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "client_id=$client_id" \
            -d "scope=openid profile email" 2>/dev/null || echo '{}')
        
        local error_type=$(echo "$test_response" | jq -r '.error // empty')
        
        if [[ "$error_type" != "invalid_client" ]]; then
            log "Client ID '$client_id' существует"
            echo "$client_id"
            return 0
        fi
    done
    
    warn "Не найден подходящий client_id"
    return 1
}

# Инструкция для ручного создания токена
show_manual_instructions() {
    info "Ручное создание Personal Access Token:"
    info ""
    info "1. Откройте: https://edge.2gc.ru/ui/console"
    info "2. Войдите:"
    info "   Email: admin@local.test"
    info "   Password: jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"
    info ""
    info "3. Personal Settings (аватар) → Personal Access Tokens → New"
    info "4. Expiration: 1 year → Add"
    info "5. Скопируйте токен"
    info ""
    info "6. Сохраните токен:"
    info "   mkdir -p ~/.zitadel"
    info "   echo 'YOUR_TOKEN' > ~/.zitadel/token"
    info "   chmod 600 ~/.zitadel/token"
    info ""
    info "7. Или экспортируйте:"
    info "   export ZITADEL_TOKEN='YOUR_TOKEN'"
    info ""
}

# Основная функция
main() {
    log "Автоматическое получение ZITADEL токена"
    log "Domain: $ZITADEL_DOMAIN"
    log ""
    
    # Проверка существующего токена
    if [[ -n "${ZITADEL_TOKEN:-}" ]]; then
        log "Токен уже установлен в переменной ZITADEL_TOKEN"
        return 0
    fi
    
    if [[ -f "$HOME/.zitadel/token" ]]; then
        log "Токен найден в ~/.zitadel/token"
        local token=$(cat "$HOME/.zitadel/token" | tr -d '\n\r')
        export ZITADEL_TOKEN="$token"
        return 0
    fi
    
    # Попытка автоматического получения
    if discover_applications; then
        if try_device_flow_with_console; then
            info "Теперь авторизуйтесь в браузере и возвращайтесь сюда"
            read -p "Нажмите Enter после авторизации..."
        fi
    fi
    
    # Если автоматический способ не сработал
    show_manual_instructions
    
    info "После создания токена запустите:"
    info "export ZITADEL_TOKEN='your_token_here'"
    info "./scripts/setup-cloudbridge-zitadel.sh"
}

# Запуск
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
