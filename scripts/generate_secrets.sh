#!/bin/bash

# CloudBridge Relay - Secret Generator
# Генератор безопасных секретов и паролей

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Проверка наличия openssl
check_openssl() {
    if ! command -v openssl &> /dev/null; then
        error "OpenSSL не установлен. Установите его для генерации секретов."
        exit 1
    fi
}

# Генерация JWT секрета
generate_jwt_secret() {
    log "Генерация JWT секрета..."
    local secret=$(openssl rand -base64 32)
    echo "RELAY_JWT_SECRET=$secret"
    echo
}

# Генерация пароля
generate_password() {
    local length=${1:-16}
    local password=$(openssl rand -base64 $length | tr -d "=+/" | cut -c1-$length)
    echo "$password"
}

# Генерация всех секретов
generate_all_secrets() {
    log "Генерация всех секретов для CloudBridge Relay..."
    echo
    
    # JWT секрет
    echo "# JWT Configuration"
    echo "RELAY_JWT_SECRET=$(openssl rand -base64 32)"
    echo
    
    # Пароли администраторов
    echo "# Admin Passwords"
    echo "KEYCLOAK_ADMIN_PASSWORD=$(generate_password 16)"
    echo "GRAFANA_ADMIN_PASSWORD=$(generate_password 16)"
    echo
    
    # Пароль базы данных
    echo "# Database Password"
    echo "KEYCLOAK_DB_PASSWORD=$(generate_password 16)"
    echo
    
    # Секрет клиента Keycloak
    echo "# Keycloak Client Secret"
    echo "KEYCLOAK_CLIENT_SECRET=$(openssl rand -base64 24)"
    echo
    
    # Пароль для шифрования резервных копий
    echo "# Backup Encryption Password"
    echo "ENCRYPTION_PASSWORD=$(generate_password 20)"
    echo
}

# Генерация .env файла с секретами
generate_env_file() {
    local output_file=${1:-.env}
    
    log "Генерация файла $output_file с безопасными секретами..."
    
    if [[ -f "$output_file" ]]; then
        warn "Файл $output_file уже существует. Создаю резервную копию..."
        cp "$output_file" "$output_file.backup.$(date +%Y%m%d_%H%M%S)"
    fi
    
    # Создаем новый .env файл
    cat > "$output_file" << 'EOF'
# CloudBridge Relay - Environment Variables
# Автоматически сгенерированные секреты
# ВНИМАНИЕ: Никогда не коммитьте этот файл в git!

# =============================================================================
# JWT Configuration
# =============================================================================
EOF
    
    echo "RELAY_JWT_SECRET=$(openssl rand -base64 32)" >> "$output_file"
    
    cat >> "$output_file" << 'EOF'

# =============================================================================
# Keycloak Configuration
# =============================================================================
KEYCLOAK_SERVER_URL=http://localhost:8080
KEYCLOAK_REALM=cloudbridge
KEYCLOAK_CLIENT_ID=relay-client
EOF
    
    echo "KEYCLOAK_CLIENT_SECRET=$(openssl rand -base64 24)" >> "$output_file"
    
    cat >> "$output_file" << 'EOF'
KEYCLOAK_PUBLIC_KEY=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
KEYCLOAK_ISSUER=http://localhost:8080/realms/cloudbridge
KEYCLOAK_AUDIENCE=relay-client

# =============================================================================
# Admin Passwords
# =============================================================================
KEYCLOAK_ADMIN=admin
EOF
    
    echo "KEYCLOAK_ADMIN_PASSWORD=$(generate_password 16)" >> "$output_file"
    echo "GRAFANA_ADMIN_PASSWORD=$(generate_password 16)" >> "$output_file"
    
    cat >> "$output_file" << 'EOF'

# =============================================================================
# Database Configuration
# =============================================================================
EOF
    
    echo "KEYCLOAK_DB_PASSWORD=$(generate_password 16)" >> "$output_file"
    
    cat >> "$output_file" << 'EOF'

# =============================================================================
# Backup Configuration
# =============================================================================
BACKUP_DIR=./backups
BACKUP_RETENTION_DAYS=30
COMPRESS_BACKUP=true
ENCRYPT_BACKUP=false
EOF
    
    echo "ENCRYPTION_PASSWORD=$(generate_password 20)" >> "$output_file"
    
    cat >> "$output_file" << 'EOF'

# =============================================================================
# Security Best Practices
# =============================================================================
# ВНИМАНИЕ: Эти секреты сгенерированы автоматически
# ВНИМАНИЕ: Храните их в безопасном месте
# ВНИМАНИЕ: Регулярно ротируйте секреты
# ВНИМАНИЕ: Используйте менеджер секретов в продакшене
EOF
    
    log "Файл $output_file создан с безопасными секретами"
    warn "ВАЖНО: Измените URL и другие настройки под ваше окружение!"
}

# Показать справку
show_help() {
    echo "CloudBridge Relay - Secret Generator"
    echo
    echo "Использование: $0 [опции]"
    echo
    echo "Опции:"
    echo "  -a, --all           Генерировать все секреты"
    echo "  -j, --jwt           Генерировать только JWT секрет"
    echo "  -p, --password      Генерировать пароль указанной длины"
    echo "  -e, --env FILE      Генерировать .env файл (по умолчанию .env)"
    echo "  -h, --help          Показать эту справку"
    echo
    echo "Примеры:"
    echo "  $0 --all                    # Генерировать все секреты"
    echo "  $0 --jwt                    # Только JWT секрет"
    echo "  $0 --password 20            # Пароль длиной 20 символов"
    echo "  $0 --env .env.production    # .env файл для продакшена"
    echo
}

# Основная функция
main() {
    check_openssl
    
    case "${1:-}" in
        -a|--all)
            generate_all_secrets
            ;;
        -j|--jwt)
            generate_jwt_secret
            ;;
        -p|--password)
            local length=${2:-16}
            if [[ ! "$length" =~ ^[0-9]+$ ]]; then
                error "Длина пароля должна быть числом"
                exit 1
            fi
            echo "Сгенерированный пароль: $(generate_password $length)"
            ;;
        -e|--env)
            local file=${2:-.env}
            generate_env_file "$file"
            ;;
        -h|--help|"")
            show_help
            ;;
        *)
            error "Неизвестная опция: $1"
            show_help
            exit 1
            ;;
    esac
}

# Запуск скрипта
main "$@" 