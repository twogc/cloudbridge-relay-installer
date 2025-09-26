#!/bin/bash
set -euo pipefail

# ZITADEL Admin Password Reset Script
# Скрипт для сброса пароля администратора ZITADEL

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Конфигурация
ZITADEL_DOMAIN="edge.2gc.ru"
ZITADEL_API_URL="https://$ZITADEL_DOMAIN"
ADMIN_USER_ID="339340397199818863"
NEW_PASSWORD="${1:-CloudBridge123!}"

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

# Метод 1: Прямое обновление в базе данных PostgreSQL
reset_password_via_database() {
    log "Попытка сброса пароля через базу данных..."
    
    local new_password="$1"
    local password_hash
    
    # Создаем bcrypt hash пароля
    password_hash=$(python3 -c "
import bcrypt
password = '$new_password'.encode('utf-8')
hashed = bcrypt.hashpw(password, bcrypt.gensalt(rounds=10))
print(hashed.decode('utf-8'))
" 2>/dev/null || echo "")
    
    if [[ -z "$password_hash" ]]; then
        error "Не удалось создать hash пароля. Установите: pip3 install bcrypt"
        return 1
    fi
    
    log "Hash пароля создан: ${password_hash:0:20}..."
    
    # Обновляем пароль в базе данных
    local pg_password
    pg_password=$(kubectl get secret -n zitadel pg-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d)
    
    local update_result
    update_result=$(kubectl exec -n zitadel pg-postgresql-0 -- env PGPASSWORD="$pg_password" psql -U postgres -d zitadel -c "
        UPDATE projections.users14_humans 
        SET password_encoded = '$password_hash',
            password_change_required = false,
            password_changed = NOW()
        WHERE user_id = '$ADMIN_USER_ID';
    " 2>&1)
    
    if echo "$update_result" | grep -q "UPDATE 1"; then
        log "✅ Пароль обновлен в базе данных"
        return 0
    else
        warn "⚠️ Обновление в базе данных не удалось: $update_result"
        return 1
    fi
}

# Метод 2: Создание нового admin пользователя
create_new_admin_user() {
    log "Создание нового admin пользователя..."
    
    local login_token
    login_token=$(kubectl get secret -n zitadel login-client -o jsonpath='{.data.pat}' | base64 -d)
    
    local new_user_data
    new_user_data=$(cat <<EOF
{
  "userName": "cloudbridge-admin",
  "profile": {
    "firstName": "CloudBridge", 
    "lastName": "Administrator",
    "displayName": "CloudBridge Admin"
  },
  "email": {
    "email": "admin@cloudbridge.local",
    "isEmailVerified": true
  },
  "password": {
    "password": "$NEW_PASSWORD",
    "changeRequired": false
  }
}
EOF
)
    
    local response
    response=$(curl -s -X POST "$ZITADEL_API_URL/management/v1/users/human" \
        -H "Authorization: Bearer $login_token" \
        -H "Content-Type: application/json" \
        -d "$new_user_data")
    
    local user_id
    user_id=$(echo "$response" | jq -r '.userId // empty' 2>/dev/null)
    
    if [[ -n "$user_id" ]]; then
        log "✅ Новый admin пользователь создан: $user_id"
        log "   Username: cloudbridge-admin"
        log "   Email: admin@cloudbridge.local"  
        log "   Password: $NEW_PASSWORD"
        return 0
    else
        warn "⚠️ Не удалось создать нового пользователя"
        warn "Response: $response"
        return 1
    fi
}

# Метод 3: Сброс через ZITADEL CLI в поде
reset_via_zitadel_cli() {
    log "Попытка сброса через ZITADEL CLI..."
    
    # Запускаем временный под с ZITADEL CLI
    kubectl run zitadel-admin-reset --rm -i --tty \
        --image=ghcr.io/zitadel/zitadel:latest \
        --env="ZITADEL_DATABASE_POSTGRES_HOST=pg-postgresql.zitadel.svc.cluster.local" \
        --env="ZITADEL_DATABASE_POSTGRES_PORT=5432" \
        --env="ZITADEL_DATABASE_POSTGRES_DATABASE=zitadel" \
        --env="ZITADEL_DATABASE_POSTGRES_USER_USERNAME=zitadel" \
        --env="ZITADEL_DATABASE_POSTGRES_USER_PASSWORD=$(kubectl get secret -n zitadel pg-postgresql -o jsonpath='{.data.password}' | base64 -d)" \
        --env="ZITADEL_DATABASE_POSTGRES_USER_SSL_MODE=disable" \
        --env="ZITADEL_EXTERNALDOMAIN=$ZITADEL_DOMAIN" \
        --env="ZITADEL_EXTERNALSECURE=true" \
        -- /bin/bash -c "
            zitadel start-from-init \
                --config /zitadel-config.yaml \
                --steps /zitadel-init-steps.yaml \
                --masterkeyFromEnv
        " || true
    
    warn "CLI метод требует дополнительной настройки"
    return 1
}

# Проверка доступа
test_login() {
    local username="$1" 
    local password="$2"
    
    log "Тестирование входа: $username"
    
    # Попытка получить токен через password flow (если поддерживается)
    local test_response
    test_response=$(curl -s -X POST "$ZITADEL_API_URL/oauth/v2/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=password" \
        -d "username=$username" \
        -d "password=$password" \
        -d "scope=openid profile email" 2>/dev/null || echo '{}')
    
    local access_token
    access_token=$(echo "$test_response" | jq -r '.access_token // empty' 2>/dev/null)
    
    if [[ -n "$access_token" ]]; then
        log "✅ Вход успешен! Токен получен"
        return 0
    else
        warn "⚠️ Автоматический тест входа не удался"
        warn "Попробуйте войти вручную: https://$ZITADEL_DOMAIN/ui/console/"
        return 1
    fi
}

# Основная функция
main() {
    log "🔐 Сброс пароля ZITADEL Admin"
    log "Пользователь: zitadel-admin@zitadel.edge.2gc.ru"
    log "Новый пароль: $NEW_PASSWORD"
    log ""
    
    # Проверяем подключение к базе данных
    local pg_password
    pg_password=$(kubectl get secret -n zitadel pg-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d 2>/dev/null)
    
    if [[ -z "$pg_password" ]]; then
        error "Не удалось получить пароль PostgreSQL"
        exit 1
    fi
    
    log "Доступные методы сброса:"
    log "1. Обновление hash пароля в базе данных (прямой метод)"
    log "2. Создание нового admin пользователя"
    log "3. Использование ZITADEL CLI (продвинутый)"
    log ""
    
    # Пробуем методы по очереди
    if reset_password_via_database "$NEW_PASSWORD"; then
        log "🎉 Пароль сброшен через базу данных!"
        log ""
        log "Попробуйте войти:"
        log "  URL: https://$ZITADEL_DOMAIN/ui/console/"
        log "  Username: zitadel-admin@zitadel.edge.2gc.ru"
        log "  Password: $NEW_PASSWORD"
        
    elif create_new_admin_user; then
        log "🎉 Создан новый admin пользователь!"
        log ""
        log "Попробуйте войти:"
        log "  URL: https://$ZITADEL_DOMAIN/ui/console/"
        log "  Username: cloudbridge-admin"
        log "  Email: admin@cloudbridge.local"
        log "  Password: $NEW_PASSWORD"
        
    else
        error "Не удалось сбросить пароль автоматически"
        error ""
        error "Ручные варианты:"
        error "1. Используйте port-forward: kubectl port-forward -n zitadel svc/zitadel 8080:8080"
        error "2. Затем откройте: http://localhost:8080"
        error "3. Или обратитесь к системному администратору"
    fi
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
