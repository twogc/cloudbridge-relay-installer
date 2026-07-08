#!/bin/bash
set -euo pipefail

# ZITADEL Token Helper Script
# Скрипт для получения токена доступа к ZITADEL API

ZITADEL_DOMAIN="${ZITADEL_DOMAIN:-edge.2gc.ru}"
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"

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

# Получение токена через Device Code Flow
get_token_device_flow() {
    log "Инициация Device Code Flow..."
    
    # ZITADEL system application client ID (обычно для console)
    local client_id="zitadel-console"
    
    # Запрос device authorization
    local device_response
    device_response=$(curl -s -X POST "$ZITADEL_API_URL/oauth/v2/device_authorization" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "client_id=$client_id" \
        -d "scope=openid profile email urn:zitadel:iam:org:project:id:zitadel:aud" 2>/dev/null || echo '{}')
    
    local device_code
    local user_code
    local verification_uri
    local verification_uri_complete
    local expires_in
    local interval
    
    device_code=$(echo "$device_response" | jq -r '.device_code // empty')
    user_code=$(echo "$device_response" | jq -r '.user_code // empty')
    verification_uri=$(echo "$device_response" | jq -r '.verification_uri // empty')
    verification_uri_complete=$(echo "$device_response" | jq -r '.verification_uri_complete // empty')
    expires_in=$(echo "$device_response" | jq -r '.expires_in // 300')
    interval=$(echo "$device_response" | jq -r '.interval // 5')
    
    if [[ -z "$device_code" || "$device_code" == "null" ]]; then
        error "Не удалось инициировать Device Code Flow"
        error "Response: $device_response"
        return 1
    fi
    
    info "Для аутентификации:"
    info "   1. Откройте браузер по ссылке: $verification_uri"
    info "   2. Введите код: $user_code"
    if [[ -n "$verification_uri_complete" && "$verification_uri_complete" != "null" ]]; then
        info "   3. Или откройте прямую ссылку: $verification_uri_complete"
    fi
    info ""
    info "У вас есть $expires_in секунд для завершения аутентификации"
    info "🔄 Ожидание завершения аутентификации..."
    
    # Polling для получения токена
    local end_time=$(($(date +%s) + expires_in))
    
    while [[ $(date +%s) -lt $end_time ]]; do
        sleep "$interval"
        
        local token_response
        token_response=$(curl -s -X POST "$ZITADEL_API_URL/oauth/v2/token" \
            -H "Content-Type: application/x-www-form-urlencoded" \
            -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
            -d "client_id=$client_id" \
            -d "device_code=$device_code" 2>/dev/null || echo '{}')
        
        local error_type
        error_type=$(echo "$token_response" | jq -r '.error // empty')
        
        case "$error_type" in
            "")
                # Успех - токен получен
                local access_token
                access_token=$(echo "$token_response" | jq -r '.access_token // empty')
                if [[ -n "$access_token" && "$access_token" != "null" ]]; then
                    log "Токен доступа получен успешно!"
                    echo "$access_token"
                    return 0
                fi
                ;;
            "authorization_pending")
                info "Ожидание аутентификации пользователя..."
                ;;
            "slow_down")
                info "🐌 Замедление запросов..."
                interval=$((interval + 2))
                ;;
            "access_denied")
                error "Аутентификация отклонена пользователем"
                return 1
                ;;
            "expired_token")
                error "⏰ Время ожидания аутентификации истекло"
                return 1
                ;;
            *)
                error "Неизвестная ошибка: $error_type"
                error "Response: $token_response"
                return 1
                ;;
        esac
    done
    
    error "⏰ Время ожидания аутентификации истекло"
    return 1
}

# Инструкция по созданию Personal Access Token
show_pat_instructions() {
    info "Инструкция по созданию Personal Access Token:"
    info ""
    info "1. Откройте ZITADEL Console: $ZITADEL_API_URL"
    info "2. Войдите под admin учетной записью:"
    info "   Username: zitadel-admin@zitadel.edge.2gc.ru"
    info "   Password: Password1!"
    info ""
    info "3. Перейдите в Personal Settings (аватар в правом верхнем углу)"
    info "4. Выберите \"Personal Access Tokens\""
    info "5. Нажмите \"New\""
    info "6. Укажите срок действия (например, 1 год)"
    info "7. Скопируйте созданный токен"
    info ""
    info "8. Используйте токен одним из способов:"
    info "   export ZITADEL_TOKEN='ваш_токен'"
    info "   или"
    info "   echo 'ваш_токен' > ~/.zitadel/token"
    info ""
}

# Создание Machine User через console
show_machine_user_instructions() {
    info "Инструкция по созданию Machine User:"
    info ""
    info "1. Откройте ZITADEL Console: $ZITADEL_API_URL"
    info "2. Войдите под admin учетной записью"
    info "3. Перейдите в Organization → Users"
    info "4. Нажмите \"New\" → \"Machine\""
    info "5. Заполните:"
    info "   Name: cloudbridge-admin"
    info "   Username: cloudbridge-admin"
    info "   Description: CloudBridge Relay Admin Machine User"
    info ""
    info "6. После создания:"
    info "   - Перейдите в созданного пользователя"
    info "   - Вкладка \"Keys\""
    info "   - \"New Key\""
    info "   - Тип: JSON"
    info "   - Срок действия: 1 год"
    info "   - Сохраните JSON ключ в ~/.zitadel/machine-key.json"
    info ""
    info "7. Назначьте права:"
    info "   - Вкладка \"Authorizations\""
    info "   - \"New\" → Organization Manager или Project Manager"
    info ""
}

# Проверка существующего токена
check_existing_token() {
    # Проверка переменной окружения
    if [[ -n "${ZITADEL_TOKEN:-}" ]]; then
        log "Найден токен в переменной ZITADEL_TOKEN"
        echo "$ZITADEL_TOKEN"
        return 0
    fi
    
    # Проверка файла токена
    if [[ -f "$HOME/.zitadel/token" ]]; then
        log "Найден токен в ~/.zitadel/token"
        cat "$HOME/.zitadel/token" | tr -d '\n\r'
        return 0
    fi
    
    return 1
}

# Валидация токена
validate_token() {
    local token="$1"
    
    log "Проверка валидности токена..."
    
    local response
    response=$(curl -s -H "Authorization: Bearer $token" \
        "$ZITADEL_API_URL/auth/v1/users/me" 2>/dev/null || echo '{}')
    
    local user_id
    user_id=$(echo "$response" | jq -r '.user.id // empty' 2>/dev/null)
    
    if [[ -n "$user_id" && "$user_id" != "null" ]]; then
        local display_name
        display_name=$(echo "$response" | jq -r '.user.human.profile.display_name // .user.machine.name // "Unknown"' 2>/dev/null)
        log "Токен валиден, пользователь: $display_name ($user_id)"
        return 0
    fi
    
    error "Токен недействителен или истек"
    return 1
}

# Основная функция
main() {
    log "ZITADEL Token Helper"
    log "Domain: $ZITADEL_DOMAIN"
    log ""
    
    # Проверка существующего токена
    local token
    if token=$(check_existing_token); then
        if validate_token "$token"; then
            info "Токен готов к использованию:"
            info "export ZITADEL_TOKEN='$token'"
            return 0
        else
            warn "Существующий токен недействителен"
        fi
    fi
    
    # Если существующего токена нет, предложить варианты
    info "Выберите способ получения токена:"
    info "1. Device Code Flow (интерактивная аутентификация)"
    info "2. Personal Access Token (ручное создание)"
    info "3. Machine User (для автоматизации)"
    info ""
    
    read -p "Выберите вариант (1/2/3): " choice
    
    case "$choice" in
        1)
            if token=$(get_token_device_flow); then
                # Сохранение токена
                mkdir -p "$HOME/.zitadel"
                echo "$token" > "$HOME/.zitadel/token"
                chmod 600 "$HOME/.zitadel/token"
                
                log "Токен сохранен в ~/.zitadel/token"
                info "Для использования в других скриптах:"
                info "export ZITADEL_TOKEN='$token'"
            fi
            ;;
        2)
            show_pat_instructions
            ;;
        3)
            show_machine_user_instructions
            ;;
        *)
            error "Неверный выбор"
            return 1
            ;;
    esac
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
