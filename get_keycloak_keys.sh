#!/bin/bash

# CloudBridge Relay - Keycloak Keys Retrieval Script
# Этот скрипт получает ключи из Keycloak API

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

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Конфигурация
KEYCLOAK_SERVER="10.0.0.73"
KEYCLOAK_REALM="cloudbridge"
KEYCLOAK_ADMIN="admin"
KEYCLOAK_ADMIN_PASSWORD="XX3pJi714OF9W4b4DtWdndZhy"

# Функция для проверки доступности Keycloak
check_keycloak_availability() {
    local ports=("8080" "8443" "443" "80")
    local protocols=("http" "https")
    
    for protocol in "${protocols[@]}"; do
        for port in "${ports[@]}"; do
            local url="$protocol://$KEYCLOAK_SERVER:$port"
            log "Проверка $url..."
            
            if curl -k -s --connect-timeout 5 "$url/realms/$KEYCLOAK_REALM" > /dev/null 2>&1; then
                success "Keycloak доступен на $url"
                echo "$url"
                return 0
            fi
        done
    done
    
    error "Keycloak недоступен на всех проверенных портах"
    return 1
}

# Функция для получения токена администратора
get_admin_token() {
    local base_url="$1"
    local token_url="$base_url/realms/master/protocol/openid-connect/token"
    
    log "Получение токена администратора..."
    
    local response=$(curl -k -s -X POST "$token_url" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=$KEYCLOAK_ADMIN" \
        -d "password=$KEYCLOAK_ADMIN_PASSWORD" \
        -d "grant_type=password" \
        -d "client_id=admin-cli" 2>/dev/null)
    
    if [ $? -eq 0 ] && echo "$response" | jq -e '.access_token' > /dev/null 2>&1; then
        local token=$(echo "$response" | jq -r '.access_token')
        success "Токен администратора получен"
        echo "$token"
        return 0
    else
        error "Не удалось получить токен администратора"
        echo "$response" | jq . 2>/dev/null || echo "$response"
        return 1
    fi
}

# Функция для получения JWKS
get_jwks() {
    local base_url="$1"
    local jwks_url="$base_url/realms/$KEYCLOAK_REALM/protocol/openid-connect/certs"
    
    log "Получение JWKS из $jwks_url..."
    
    local jwks=$(curl -k -s "$jwks_url" 2>/dev/null)
    
    if [ $? -eq 0 ] && echo "$jwks" | jq -e '.keys' > /dev/null 2>&1; then
        success "JWKS получен"
        echo "$jwks"
        return 0
    else
        error "Не удалось получить JWKS"
        echo "$jwks" | jq . 2>/dev/null || echo "$jwks"
        return 1
    fi
}

# Функция для извлечения Public Key из JWKS
extract_public_key() {
    local jwks="$1"
    
    log "Извлечение Public Key из JWKS..."
    
    # Получаем первый ключ из JWKS
    local key_data=$(echo "$jwks" | jq -r '.keys[0]')
    
    if [ "$key_data" = "null" ] || [ -z "$key_data" ]; then
        error "Ключи не найдены в JWKS"
        return 1
    fi
    
    # Извлекаем x5c (сертификат)
    local x5c=$(echo "$key_data" | jq -r '.x5c[0]')
    
    if [ "$x5c" = "null" ] || [ -z "$x5c" ]; then
        error "Сертификат не найден в ключе"
        return 1
    fi
    
    # Декодируем base64 и извлекаем public key
    local public_key=$(echo "$x5c" | base64 -d 2>/dev/null | openssl x509 -pubkey -noout 2>/dev/null)
    
    if [ $? -eq 0 ] && [ -n "$public_key" ]; then
        success "Public Key извлечен"
        echo "$public_key"
        return 0
    else
        error "Не удалось извлечь Public Key"
        return 1
    fi
}

# Функция для получения информации о клиенте
get_client_info() {
    local base_url="$1"
    local admin_token="$2"
    local client_id="relay-client"
    
    log "Получение информации о клиенте $client_id..."
    
    local clients_url="$base_url/admin/realms/$KEYCLOAK_REALM/clients"
    local response=$(curl -k -s -H "Authorization: Bearer $admin_token" "$clients_url" 2>/dev/null)
    
    if [ $? -eq 0 ] && echo "$response" | jq -e '.[]' > /dev/null 2>&1; then
        local client_info=$(echo "$response" | jq -r ".[] | select(.clientId == \"$client_id\")")
        
        if [ "$client_info" != "null" ] && [ -n "$client_info" ]; then
            success "Информация о клиенте получена"
            echo "$client_info"
            return 0
        else
            warn "Клиент $client_id не найден"
            return 1
        fi
    else
        error "Не удалось получить список клиентов"
        echo "$response" | jq . 2>/dev/null || echo "$response"
        return 1
    fi
}

# Функция для создания .env файла с полученными данными
create_env_file() {
    local base_url="$1"
    local public_key="$2"
    local client_secret="$3"
    
    local env_file="keycloak-keys.env"
    
    log "Создание файла $env_file с полученными данными..."
    
    cat > "$env_file" << EOF
# CloudBridge Relay - Keycloak Keys (получено автоматически)
# Создано скриптом get_keycloak_keys.sh

# Keycloak Configuration
KEYCLOAK_SERVER_URL=$base_url
KEYCLOAK_REALM=$KEYCLOAK_REALM
KEYCLOAK_CLIENT_ID=relay-client
KEYCLOAK_CLIENT_SECRET=$client_secret
KEYCLOAK_ADMIN=$KEYCLOAK_ADMIN
KEYCLOAK_ADMIN_PASSWORD=$KEYCLOAK_ADMIN_PASSWORD
KEYCLOAK_PUBLIC_KEY="$public_key"
KEYCLOAK_ISSUER=$base_url/realms/$KEYCLOAK_REALM
KEYCLOAK_AUDIENCE=relay-client

# Fallback Configuration
KEYCLOAK_FALLBACK_ENABLED=true
KEYCLOAK_FALLBACK_SECRET=85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
EOF
    
    success "Файл $env_file создан"
}

# Основная функция
main() {
    echo
    log "Начало получения ключей из Keycloak API..."
    echo
    
    # Проверка доступности Keycloak
    local base_url=$(check_keycloak_availability)
    if [ $? -ne 0 ]; then
        error "Keycloak недоступен. Проверьте подключение к серверу."
        exit 1
    fi
    
    # Получение JWKS
    local jwks=$(get_jwks "$base_url")
    if [ $? -ne 0 ]; then
        error "Не удалось получить JWKS"
        exit 1
    fi
    
    # Извлечение Public Key
    local public_key=$(extract_public_key "$jwks")
    if [ $? -ne 0 ]; then
        error "Не удалось извлечь Public Key"
        exit 1
    fi
    
    # Получение токена администратора
    local admin_token=$(get_admin_token "$base_url")
    if [ $? -ne 0 ]; then
        warn "Не удалось получить токен администратора, используем fallback client secret"
        local client_secret="YbUqdiZvJU4zNFVd5JRqaTc9tL9iHwAa"
    else
        # Получение информации о клиенте
        local client_info=$(get_client_info "$base_url" "$admin_token")
        if [ $? -eq 0 ]; then
            local client_secret=$(echo "$client_info" | jq -r '.secret // "YbUqdiZvJU4zNFVd5JRqaTc9tL9iHwAa"')
        else
            local client_secret="YbUqdiZvJU4zNFVd5JRqaTc9tL9iHwAa"
        fi
    fi
    
    # Создание .env файла
    create_env_file "$base_url" "$public_key" "$client_secret"
    
    echo
    success "Получение ключей завершено!"
    echo
    log "Для применения изменений:"
    echo "  source keycloak-keys.env"
    echo "  ./cloudbridge-relay --config=configs/config.json"
    echo
    log "Для Minikube:"
    echo "  kubectl create configmap cloudbridge-keycloak-config --from-env-file=keycloak-keys.env -n cloudbridge"
    echo "  kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge"
    echo
}

# Запуск скрипта
main "$@"
