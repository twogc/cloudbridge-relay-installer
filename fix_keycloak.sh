#!/bin/bash

# CloudBridge Relay - Keycloak Quick Fix Script
# Этот скрипт исправляет проблему с подключением к Keycloak

set -e

echo "🔧 CloudBridge Relay - Keycloak Quick Fix"
echo "=========================================="

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

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Проверка доступности Keycloak сервера
check_keycloak_server() {
    log "Проверка доступности Keycloak сервера на 10.0.0.73:8080..."
    
    if ping -c 1 -W 5 10.0.0.73 > /dev/null 2>&1; then
        success "Сервер 10.0.0.73 доступен"
    else
        warn "Сервер 10.0.0.73 недоступен (ping failed)"
    fi
    
    if curl -s --connect-timeout 5 http://10.0.0.73:8080/realms/cloudbridge > /dev/null 2>&1; then
        success "Keycloak сервис отвечает"
        return 0
    else
        warn "Keycloak сервис недоступен на порту 8080"
        return 1
    fi
}

# Получение Public Key из Keycloak
get_public_key() {
    log "Попытка получить Public Key из Keycloak..."
    
    local jwks_url="http://10.0.0.73:8080/realms/cloudbridge/protocol/openid-connect/certs"
    
    if curl -s --connect-timeout 5 "$jwks_url" > /dev/null 2>&1; then
        log "Получение JWKS из $jwks_url..."
        
        # Получаем JWKS и извлекаем public key
        local public_key=$(curl -s --connect-timeout 5 "$jwks_url" | \
            jq -r '.keys[0].x5c[0]' 2>/dev/null | \
            base64 -d 2>/dev/null | \
            openssl x509 -pubkey -noout 2>/dev/null)
        
        if [ -n "$public_key" ]; then
            success "Public Key получен из Keycloak"
            echo "$public_key"
            return 0
        else
            warn "Не удалось извлечь Public Key из JWKS"
            return 1
        fi
    else
        warn "JWKS endpoint недоступен"
        return 1
    fi
}

# Установка переменных окружения
set_environment_variables() {
    log "Установка переменных окружения Keycloak..."
    
    export KEYCLOAK_SERVER_URL="http://10.0.0.73:8080"
    export KEYCLOAK_REALM="cloudbridge"
    export KEYCLOAK_CLIENT_ID="relay-client"
    export KEYCLOAK_CLIENT_SECRET="YbUqdiZvJU4zNFVd5JRqaTc9tL9iHwAa"
    export KEYCLOAK_ADMIN="admin"
    export KEYCLOAK_ADMIN_PASSWORD="XX3pJi714OF9W4b4DtWdndZhy"
    export KEYCLOAK_ISSUER="http://10.0.0.73:8080/realms/cloudbridge"
    export KEYCLOAK_AUDIENCE="relay-client"
    
    # Попытка получить Public Key
    local public_key=$(get_public_key)
    if [ $? -eq 0 ] && [ -n "$public_key" ]; then
        export KEYCLOAK_PUBLIC_KEY="$public_key"
        success "KEYCLOAK_PUBLIC_KEY установлен из Keycloak"
    else
        # Используем fallback Public Key
        export KEYCLOAK_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwYizA49h313kUlnaCD8u
3oWgVV0o3mEPpuyhfpp+o/PQNvg6oolodiUhum9uHROvXWT/OPlPLy6fsC3hpprl
VII7DiFuRYxtWU24O1Nd66sLt8lNxfTHKH/fwlS4opPHaHZkFzIXrKNeJvTf+kO/h
Hial8Dut+Z/tIBcwBrPWVll6P/4wpruZlK6dWoNtpac/m6RYyrEMnX+nQ9DZ3S/bS
A1SUvRASCWxk/qQQpz/Zp3sgqLNa12h/JztEUuHm1RyshKFvHsTwJuuvmXzzJyAMP
PTHr0M0jcdYH64VJFoDSp6xWgqFpuCn3WqevXCVSMfeI1g4TzLDUEBZdesI4PawID
AQAB
-----END PUBLIC KEY-----"
        warn "Используется fallback KEYCLOAK_PUBLIC_KEY"
    fi
    
    success "Переменные окружения Keycloak установлены"
}

# Проверка переменных окружения
verify_environment() {
    log "Проверка установленных переменных окружения..."
    
    local vars=("KEYCLOAK_SERVER_URL" "KEYCLOAK_REALM" "KEYCLOAK_CLIENT_ID" "KEYCLOAK_PUBLIC_KEY" "KEYCLOAK_ISSUER" "KEYCLOAK_AUDIENCE")
    
    for var in "${vars[@]}"; do
        if [ -n "${!var}" ]; then
            success "$var установлена"
        else
            error "$var не установлена"
            return 1
        fi
    done
    
    return 0
}

# Создание .env файла с переменными Keycloak
create_env_file() {
    log "Создание .env файла с переменными Keycloak..."
    
    local env_file=".env.keycloak"
    
    cat > "$env_file" << EOF
# CloudBridge Relay - Keycloak Environment Variables
# Создано автоматически скриптом fix_keycloak.sh

# Keycloak Configuration
KEYCLOAK_SERVER_URL=$KEYCLOAK_SERVER_URL
KEYCLOAK_REALM=$KEYCLOAK_REALM
KEYCLOAK_CLIENT_ID=$KEYCLOAK_CLIENT_ID
KEYCLOAK_CLIENT_SECRET=$KEYCLOAK_CLIENT_SECRET
KEYCLOAK_ADMIN=$KEYCLOAK_ADMIN
KEYCLOAK_ADMIN_PASSWORD=$KEYCLOAK_ADMIN_PASSWORD
KEYCLOAK_PUBLIC_KEY="$KEYCLOAK_PUBLIC_KEY"
KEYCLOAK_ISSUER=$KEYCLOAK_ISSUER
KEYCLOAK_AUDIENCE=$KEYCLOAK_AUDIENCE

# Fallback Configuration
KEYCLOAK_FALLBACK_ENABLED=true
KEYCLOAK_FALLBACK_SECRET=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
EOF
    
    success "Файл $env_file создан"
}

# Тестирование приложения
test_application() {
    log "Тестирование приложения с новыми переменными..."
    
    if [ -f "./cloudbridge-relay" ]; then
        log "Запуск приложения для проверки..."
        timeout 10s ./cloudbridge-relay --config=configs/config.json --log-level=debug 2>&1 | head -20
    else
        warn "Исполняемый файл cloudbridge-relay не найден"
    fi
}

# Основная функция
main() {
    echo
    log "Начало исправления проблемы с Keycloak..."
    echo
    
    # Проверка доступности Keycloak
    if check_keycloak_server; then
        success "Keycloak сервер доступен"
    else
        warn "Keycloak сервер недоступен, будет использован fallback режим"
    fi
    
    # Установка переменных окружения
    set_environment_variables
    
    # Проверка переменных
    if verify_environment; then
        success "Все переменные окружения установлены корректно"
    else
        error "Ошибка в установке переменных окружения"
        exit 1
    fi
    
    # Создание .env файла
    create_env_file
    
    # Тестирование
    test_application
    
    echo
    success "Исправление завершено!"
    echo
    log "Для применения изменений:"
    echo "  source .env.keycloak"
    echo "  ./cloudbridge-relay --config=configs/config.json"
    echo
    log "Для Minikube:"
    echo "  kubectl create configmap cloudbridge-keycloak-config --from-env-file=.env.keycloak -n cloudbridge"
    echo "  kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge"
    echo
}

# Запуск скрипта
main "$@"
