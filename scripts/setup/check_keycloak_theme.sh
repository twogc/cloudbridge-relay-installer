#!/bin/bash

# Скрипт для проверки и применения темы в Keycloak
# Проверяет, что тема собралась и применяет её к realm

set -e

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

THEME_DIR="keycloak/themes/playa"
REALM_NAME="cloudbridge"

log "=== Проверка темы Keycloak ==="

# Проверяем наличие файлов темы
log "Проверяем файлы темы..."

if [ ! -f "$THEME_DIR/theme.properties" ]; then
    error "Файл theme.properties не найден"
    exit 1
fi

if [ ! -f "$THEME_DIR/login/login.ftl" ]; then
    error "Шаблон входа login.ftl не найден"
    exit 1
fi

if [ ! -f "$THEME_DIR/login/resources/js/login.js" ]; then
    error "JavaScript файл login.js не найден"
    exit 1
fi

log "✅ Все основные файлы темы найдены"

# Проверяем, что Keycloak запущен
log "Проверяем, что Keycloak запущен..."
if ! curl -s -f http://localhost:8080/ > /dev/null; then
    error "Keycloak не отвечает на http://localhost:8080"
    exit 1
fi

log "✅ Keycloak запущен и отвечает"

# Проверяем, что тема доступна в Keycloak
log "Проверяем доступность темы в Keycloak..."

# Получаем токен администратора
if [ -z "$KEYCLOAK_ADMIN" ] || [ -z "$KEYCLOAK_ADMIN_PASSWORD" ]; then
    warn "KEYCLOAK_ADMIN или KEYCLOAK_ADMIN_PASSWORD не установлены"
    warn "Не могу проверить настройки темы автоматически"
    echo
    log "Ручная проверка:"
    echo "1. Откройте http://localhost:8080"
    echo "2. Войдите как администратор"
    echo "3. Перейдите в Realm Settings > Themes"
    echo "4. Убедитесь, что тема 'playa' доступна"
    echo "5. Выберите 'playa' для Login Theme"
    echo "6. Сохраните настройки"
    exit 0
fi

# Получаем токен
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=$KEYCLOAK_ADMIN" \
    -d "password=$KEYCLOAK_ADMIN_PASSWORD" \
    -d "grant_type=password" \
    -d "client_id=admin-cli")

if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
    log "✅ Токен администратора получен"
else
    error "Не удалось получить токен администратора"
    echo "$TOKEN_RESPONSE"
    exit 1
fi

# Проверяем настройки realm
REALM_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" \
    http://localhost:8080/admin/realms/$REALM_NAME)

if echo "$REALM_RESPONSE" | grep -q "realm"; then
    CURRENT_THEME=$(echo "$REALM_RESPONSE" | jq -r '.loginTheme // empty')
    log "✅ Realm '$REALM_NAME' найден"
    
    if [ "$CURRENT_THEME" = "playa" ]; then
        log "✅ Тема 'playa' уже применена к realm"
    else
        warn "Тема 'playa' не применена к realm (текущая тема: $CURRENT_THEME)"
        
        # Применяем тему
        log "Применяю тему 'playa' к realm..."
        curl -s -X PUT http://localhost:8080/admin/realms/$REALM_NAME \
            -H "Authorization: Bearer $ACCESS_TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"realm\": \"$REALM_NAME\",
                \"loginTheme\": \"playa\",
                \"accountTheme\": \"playa\",
                \"adminTheme\": \"playa\",
                \"emailTheme\": \"playa\"
            }"
        
        log "✅ Тема 'playa' применена к realm"
    fi
else
    error "Realm '$REALM_NAME' не найден"
    exit 1
fi

echo
log "=== Проверка завершена ==="
log "Тема готова к использованию"
log "Страница входа: http://localhost:8080/realms/$REALM_NAME/protocol/openid-connect/auth" 