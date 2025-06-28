#!/bin/bash

# CloudBridge Relay - Moscow Region Setup Script
# Скрипт настройки для московского региона с учетом российских требований

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

# Проверка прав администратора
check_root() {
    if [[ $EUID -eq 0 ]]; then
        error "Этот скрипт не должен запускаться от root пользователя"
        exit 1
    fi
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    local deps=("docker" "docker-compose" "curl" "jq" "openssl")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        error "Отсутствуют зависимости: ${missing_deps[*]}"
        error "Установите их перед запуском скрипта"
        exit 1
    fi
    
    log "Все зависимости установлены"
}

# Создание .env файла для московского региона
create_env_file() {
    log "Создание .env файла для московского региона..."
    
    if [[ -f .env ]]; then
        warn "Файл .env уже существует. Создаю резервную копию..."
        cp .env .env.backup.$(date +%Y%m%d_%H%M%S)
    fi
    
    cat > .env << 'EOF'
# CloudBridge Relay - Moscow Region Configuration
# Конфигурация для московского региона с учетом российских требований

# =============================================================================
# JWT Configuration
# =============================================================================
# Генерируйте безопасный секрет: openssl rand -base64 32
RELAY_JWT_SECRET=your-super-secure-jwt-secret-key-here-moscow

# =============================================================================
# Keycloak Configuration
# =============================================================================
KEYCLOAK_SERVER_URL=http://localhost:8080
KEYCLOAK_REALM=cloudbridge
KEYCLOAK_CLIENT_ID=relay-client
KEYCLOAK_CLIENT_SECRET=your-keycloak-client-secret-here
KEYCLOAK_PUBLIC_KEY=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
KEYCLOAK_ISSUER=http://localhost:8080/realms/cloudbridge
KEYCLOAK_AUDIENCE=relay-client

# =============================================================================
# TLS Configuration
# =============================================================================
RELAY_TLS_CERT_FILE=/etc/relay/certs/fullchain.pem
RELAY_TLS_KEY_FILE=/etc/relay/certs/privkey.pem

# =============================================================================
# Server Configuration
# =============================================================================
RELAY_PORT=8082
RELAY_MAX_TUNNELS=1000

# =============================================================================
# Logging Configuration
# =============================================================================
RELAY_LOG_LEVEL=INFO
RELAY_LOG_FORMAT=json
RELAY_LOG_FILE=/var/log/relay/relay.log

# =============================================================================
# Security Configuration
# =============================================================================
RELAY_ALLOWED_IPS=127.0.0.1/32,192.168.1.0/24,10.0.0.0/8
RELAY_RATE_LIMIT_RPS=100
RELAY_RATE_LIMIT_BURST=200

# =============================================================================
# Metrics Configuration
# =============================================================================
RELAY_METRICS_PORT=8081
RELAY_METRICS_PATH=/metrics

# =============================================================================
# Development Configuration
# =============================================================================
RELAY_DEV_MODE=false
RELAY_DEBUG=false

# =============================================================================
# Moscow Region Configuration
# =============================================================================
DEPLOYMENT_REGION=moscow
TZ=Europe/Moscow
LANG=ru_RU.UTF-8
LC_ALL=ru_RU.UTF-8

# =============================================================================
# Russia-Specific Configuration
# =============================================================================
# Соответствие требованиям 152-ФЗ (защита персональных данных)
COMPLIANCE_152_FZ=true

# Соответствие требованиям 187-ФЗ (безопасность критической инфраструктуры)
COMPLIANCE_187_FZ=true

# Использование российских криптографических алгоритмов
USE_RUSSIAN_CRYPTO=true

# ГОСТ Р 34.10-2012 (подпись)
GOST_SIGNATURE_ALGORITHM=GOST_R_34_10_2012

# ГОСТ Р 34.11-2012 (хеш)
GOST_HASH_ALGORITHM=GOST_R_34_11_2012

# =============================================================================
# Keycloak Admin Configuration
# =============================================================================
KEYCLOAK_ADMIN=admin
# ИЗМЕНИТЕ ПАРОЛЬ ПО УМОЛЧАНИЮ!
KEYCLOAK_ADMIN_PASSWORD=change-this-password-immediately

# =============================================================================
# Relay Service Configuration
# =============================================================================
RELAY_AUTH_PROVIDER=keycloak
RELAY_TIMEOUT=30s
RELAY_PING_INTERVAL=30s
RELAY_CLEANUP_INTERVAL=1m

# =============================================================================
# Docker Configuration
# =============================================================================
CONTAINER_NAME=cloudbridge-relay
IMAGE_TAG=latest
DOCKER_NETWORK=relay-network

# =============================================================================
# Production Configuration
# =============================================================================
ENVIRONMENT=production
APP_VERSION=1.0.0

# =============================================================================
# Monitoring Configuration
# =============================================================================
# ИЗМЕНИТЕ ПАРОЛЬ ПО УМОЛЧАНИЮ!
GRAFANA_ADMIN_PASSWORD=change-this-password-immediately

# =============================================================================
# Backup Configuration
# =============================================================================
BACKUP_DIR=./backups
BACKUP_RETENTION_DAYS=30
COMPRESS_BACKUP=true
ENCRYPT_BACKUP=false
ENCRYPTION_PASSWORD=your-encryption-password-here

# =============================================================================
# Security Best Practices
# =============================================================================
# ВНИМАНИЕ: Измените все пароли по умолчанию перед использованием!
# ВНИМАНИЕ: Генерируйте безопасные секреты: openssl rand -base64 32
# ВНИМАНИЕ: Никогда не коммитьте этот файл в git!
EOF

    log "Файл .env создан"
    warn "ВАЖНО: Измените пароли по умолчанию перед использованием!"
    info "Генерируйте безопасные секреты командой: openssl rand -base64 32"
}

# Настройка часового пояса
setup_timezone() {
    log "Настройка часового пояса для Москвы..."
    
    # Проверяем текущий часовой пояс
    current_tz=$(timedatectl show --property=Timezone --value 2>/dev/null || echo "UTC")
    
    if [[ "$current_tz" != "Europe/Moscow" ]]; then
        warn "Текущий часовой пояс: $current_tz"
        info "Рекомендуется установить часовой пояс Europe/Moscow"
        
        if command -v timedatectl &> /dev/null; then
            read -p "Установить часовой пояс Europe/Moscow? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                sudo timedatectl set-timezone Europe/Moscow
                log "Часовой пояс установлен в Europe/Moscow"
            fi
        fi
    else
        log "Часовой пояс уже установлен в Europe/Moscow"
    fi
}

# Настройка локали
setup_locale() {
    log "Настройка локали для русского языка..."
    
    # Проверяем доступность русской локали
    if ! locale -a | grep -q "ru_RU.utf8\|ru_RU.UTF-8"; then
        warn "Русская локаль не найдена"
        info "Установка русской локали..."
        
        if command -v locale-gen &> /dev/null; then
            sudo locale-gen ru_RU.UTF-8
            sudo update-locale LANG=ru_RU.UTF-8 LC_ALL=ru_RU.UTF-8
            log "Русская локаль установлена"
        else
            warn "Не удалось установить русскую локаль автоматически"
        fi
    else
        log "Русская локаль уже установлена"
    fi
}

# Создание директорий
create_directories() {
    log "Создание необходимых директорий..."
    
    local dirs=(
        "certs"
        "logs"
        "backups"
        "config"
        "keycloak/data"
        "keycloak/imports"
        "prometheus"
        "grafana/provisioning/datasources"
        "grafana/provisioning/dashboards"
    )
    
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            mkdir -p "$dir"
            log "Создана директория: $dir"
        fi
    done
}

# Настройка сертификатов
setup_certificates() {
    log "Настройка сертификатов..."
    
    if [[ ! -f certs/fullchain.pem ]] || [[ ! -f certs/privkey.pem ]]; then
        warn "Сертификаты не найдены. Создание самоподписанных сертификатов..."
        
        # Создание самоподписанного сертификата
        openssl req -x509 -newkey rsa:4096 -keyout certs/privkey.pem \
            -out certs/fullchain.pem -days 365 -nodes \
            -subj "/C=RU/ST=Moscow/L=Moscow/O=CloudBridge/OU=IT/CN=localhost"
        
        log "Самоподписанные сертификаты созданы"
    else
        log "Сертификаты уже существуют"
    fi
}

# Настройка мониторинга
setup_monitoring() {
    log "Настройка мониторинга..."
    
    # Создание конфигурации Prometheus
    if [[ ! -f prometheus/prometheus.yml ]]; then
        cat > prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alerts/*.yml"

scrape_configs:
  - job_name: 'relay'
    static_configs:
      - targets: ['relay:8081']
    metrics_path: '/metrics'
    scrape_interval: 10s

  - job_name: 'keycloak'
    static_configs:
      - targets: ['keycloak:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF
        log "Конфигурация Prometheus создана"
    fi
    
    # Создание конфигурации Grafana
    if [[ ! -f grafana/provisioning/datasources/prometheus.yml ]]; then
        cat > grafana/provisioning/datasources/prometheus.yml << 'EOF'
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF
        log "Конфигурация Grafana создана"
    fi
}

# Проверка соответствия российским требованиям
check_russian_compliance() {
    log "Проверка соответствия российским требованиям..."
    
    # Проверка 152-ФЗ
    info "Проверка соответствия 152-ФЗ (защита персональных данных)..."
    if [[ "$COMPLIANCE_152_FZ" == "true" ]]; then
        log "✓ Соответствие 152-ФЗ включено"
    else
        warn "⚠ Соответствие 152-ФЗ не включено"
    fi
    
    # Проверка 187-ФЗ
    info "Проверка соответствия 187-ФЗ (безопасность критической инфраструктуры)..."
    if [[ "$COMPLIANCE_187_FZ" == "true" ]]; then
        log "✓ Соответствие 187-ФЗ включено"
    else
        warn "⚠ Соответствие 187-ФЗ не включено"
    fi
    
    # Проверка российских криптоалгоритмов
    info "Проверка использования российских криптоалгоритмов..."
    if [[ "$USE_RUSSIAN_CRYPTO" == "true" ]]; then
        log "✓ Использование российских криптоалгоритмов включено"
    else
        warn "⚠ Использование российских криптоалгоритмов не включено"
    fi
}

# Запуск сервисов
start_services() {
    log "Запуск сервисов..."
    
    # Загрузка переменных окружения
    if [[ -f .env ]]; then
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Запуск с помощью docker-compose
    docker-compose -f docker-compose.keycloak.yml up -d
    
    log "Сервисы запущены"
}

# Проверка статуса сервисов
check_services_status() {
    log "Проверка статуса сервисов..."
    
    sleep 10  # Даем время на запуск
    
    local services=("keycloak" "postgres" "relay" "prometheus" "alertmanager" "grafana")
    
    for service in "${services[@]}"; do
        if docker-compose -f docker-compose.keycloak.yml ps | grep -q "$service.*Up"; then
            log "✓ Сервис $service запущен"
        else
            error "✗ Сервис $service не запущен"
        fi
    done
}

# Отображение информации о доступе
show_access_info() {
    log "Информация о доступе к сервисам:"
    echo
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│                    CloudBridge Relay                        │"
    echo "│                    Московский регион                        │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ Keycloak Admin Console: http://localhost:8080               │"
    echo "│   Логин: admin                                              │"
    echo "│   Пароль: admin123                                          │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ Relay API: http://localhost:8080                            │"
    echo "│ Relay Metrics: http://localhost:8081/metrics                │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ Prometheus: http://localhost:9090                           │"
    echo "│ AlertManager: http://localhost:9093                         │"
    echo "│ Grafana: http://localhost:3000                              │"
    echo "│   Логин: admin                                              │"
    echo "│   Пароль: admin123                                          │"
    echo "├─────────────────────────────────────────────────────────────┤"
    echo "│ Логи: ./logs/                                               │"
    echo "│ Конфигурация: ./config/                                     │"
    echo "│ Резервные копии: ./backups/                                 │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo
    echo "Для остановки сервисов: docker-compose -f docker-compose.keycloak.yml down"
    echo "Для просмотра логов: docker-compose -f docker-compose.keycloak.yml logs -f"
    echo
}

# Основная функция
main() {
    echo "┌─────────────────────────────────────────────────────────────┐"
    echo "│              CloudBridge Relay Setup                        │"
    echo "│                    Московский регион                        │"
    echo "└─────────────────────────────────────────────────────────────┘"
    echo
    
    check_root
    check_dependencies
    create_env_file
    setup_timezone
    setup_locale
    create_directories
    setup_certificates
    setup_monitoring
    check_russian_compliance
    start_services
    check_services_status
    show_access_info
    
    log "Настройка московского региона завершена успешно!"
    log "Не забудьте изменить пароли по умолчанию в продакшене!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --env-only     Только создать .env файл"
        echo "  --check-only   Только проверить соответствие требованиям"
        echo
        exit 0
        ;;
    --env-only)
        create_env_file
        log "Файл .env создан"
        exit 0
        ;;
    --check-only)
        check_russian_compliance
        exit 0
        ;;
    "")
        main
        ;;
    *)
        error "Неизвестная опция: $1"
        echo "Используйте --help для получения справки"
        exit 1
        ;;
esac 