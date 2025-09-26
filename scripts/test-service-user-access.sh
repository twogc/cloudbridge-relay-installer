#!/bin/bash
set -euo pipefail

# Test Service User Access after ORG_OWNER role assignment
# Тест доступа Service User после назначения роли ORG_OWNER

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')] $*${NC}"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $*${NC}"; }

# Данные Service User
CLIENT_ID="cloudbridge-admin"
CLIENT_SECRET="5abjY0vYqyAS03dZdPsan30ibM5dmiUWW2e8F067FUWv7gODbXv6KRqTmfiqid08"
ZITADEL_DOMAIN="edge.2gc.ru"

# Personal Access Token
PAT="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"

main() {
    log "🧪 Тестирование доступа Service User к Management API"
    log ""
    
    # Тест 1: Personal Access Token
    log "📋 Тест 1: Personal Access Token"
    if test_pat_access; then
        log "✅ PAT работает - у Service User есть Manager роль!"
    else
        warn "❌ PAT не работает - нужно назначить роль ORG_OWNER"
    fi
    echo ""
    
    # Тест 2: Client Credentials
    log "📋 Тест 2: Client Credentials Grant"
    if test_client_credentials; then
        log "✅ Client Credentials работает!"
    else
        warn "❌ Client Credentials не работает"
    fi
    echo ""
    
    # Тест 3: Management API endpoints
    log "📋 Тест 3: Доступ к Management API"
    test_management_endpoints
}

test_pat_access() {
    local response
    response=$(curl -s -H "Authorization: Bearer $PAT" \
        "https://$ZITADEL_DOMAIN/management/v1/orgs/_search" \
        -H "Content-Type: application/json" \
        -d '{"queries": []}')
    
    if echo "$response" | jq -e '.result[]' >/dev/null 2>&1; then
        log "  ✅ Организации найдены: $(echo "$response" | jq '.result | length') шт."
        return 0
    else
        warn "  ❌ Ответ: $(echo "$response" | jq -r '.message // "Unknown error"')"
        return 1
    fi
}

test_client_credentials() {
    local basic_auth access_token
    basic_auth=$(echo -n "$CLIENT_ID:$CLIENT_SECRET" | base64)
    
    local token_response
    token_response=$(curl -s -X POST "https://$ZITADEL_DOMAIN/oauth/v2/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -H "Authorization: Basic $basic_auth" \
        -d "grant_type=client_credentials" \
        -d "scope=openid profile urn:zitadel:iam:org:project:id:zitadel:aud")
    
    access_token=$(echo "$token_response" | jq -r '.access_token // empty')
    
    if [[ -n "$access_token" ]]; then
        log "  ✅ Access token получен (длина: ${#access_token})"
        
        # Тест API вызова с новым токеном
        local api_response
        api_response=$(curl -s -H "Authorization: Bearer $access_token" \
            "https://$ZITADEL_DOMAIN/management/v1/orgs/me")
        
        if echo "$api_response" | jq -e '.org' >/dev/null 2>&1; then
            log "  ✅ API вызов успешен"
            return 0
        else
            warn "  ❌ API вызов не удался: $(echo "$api_response" | jq -r '.message // "Unknown"')"
            return 1
        fi
    else
        warn "  ❌ Не удалось получить токен: $(echo "$token_response" | jq -r '.error_description // .message // "Unknown"')"
        return 1
    fi
}

test_management_endpoints() {
    # Тест различных Management API endpoints
    local endpoints=(
        "/management/v1/orgs/me:Информация об организации"
        "/management/v1/projects:Список проектов"
        "/management/v1/users/me:Информация о Service User"
    )
    
    for endpoint_info in "${endpoints[@]}"; do
        local endpoint="${endpoint_info%%:*}"
        local description="${endpoint_info##*:}"
        
        log "  📡 Тест: $description"
        local response
        response=$(curl -s -H "Authorization: Bearer $PAT" \
            "https://$ZITADEL_DOMAIN$endpoint" 2>/dev/null)
        
        if echo "$response" | jq -e '.' >/dev/null 2>&1; then
            if echo "$response" | jq -e '.code' >/dev/null 2>&1; then
                warn "    ❌ Ошибка: $(echo "$response" | jq -r '.message')"
            else
                log "    ✅ Успешно"
            fi
        else
            warn "    ❌ Некорректный JSON ответ"
        fi
    done
}

# Инструкции после тестирования
show_instructions() {
    echo ""
    log "📋 Если тесты не прошли, выполните следующие шаги:"
    log ""
    log "1. Войдите в ZITADEL Console: https://$ZITADEL_DOMAIN/ui/console/"
    log "2. Логин: admin@cloudbridge.local / CloudBridge123!"
    log "3. Organizations → выберите организацию → '+' (Add Manager)"
    log "4. Найдите: cloudbridge-admin"
    log "5. Назначьте роль: Org Owner"
    log "6. Запустите тест повторно: ./scripts/test-service-user-access.sh"
    log ""
    log "После успешного теста запустите финальную настройку:"
    log "  cd /home/ubuntu/cloudbridge-relay-installer"
    log "  export ZITADEL_TOKEN=\"$PAT\""
    log "  ./scripts/setup-cloudbridge-zitadel.sh"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
    show_instructions
fi
