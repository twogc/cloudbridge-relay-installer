#!/bin/bash

# CloudBridge Relay - Production Startup Script
# Автоматический запуск всех сервисов в production режиме

set -e

# Красивая заставка
show_banner() {
    echo -e "\033[1;36m"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                                                                              ║"
    echo "║                       ██████╗ ███████╗██╗      █████╗ ██╗   ██╗              ║"
    echo "║                       ██╔══██╗██╔════╝██║     ██╔══██╗╚██╗ ██╔╝              ║"
    echo "║                       ██████╔╝█████╗  ██║     ███████║ ╚████╔╝               ║"
    echo "║                       ██╔══██╗██╔══╝  ██║     ██╔══██║  ╚██╔╝                ║"
    echo "║                       ██║  ██║███████╗███████╗██║  ██║   ██║                 ║"
    echo "║                       ╚═╝  ╚═╝╚══════╝╚══════╝╚═╝  ╚═╝   ╚═╝                 ║"
    echo "║                                                                              ║"
    echo "║                           Production Environment v2.0                        ║"
    echo "║                                                                              ║"
    echo "║  Multi-Tenant Authentication   |  High-Performance TCP Proxy                 ║"
    echo "║  Enterprise Security           |  Advanced Monitoring & Analytics            ║"
    echo "║  Auto-Scaling & Load Balancing |  Real-time Metrics & Dashboards             ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    echo -e "\033[1;33m🚀 Starting CloudBridge Relay Production Environment...\033[0m"
    echo -e "\033[1;33m⏰ $(date '+%Y-%m-%d %H:%M:%S')\033[0m"
    echo ""
    echo -e "\033[1;33m═══════════════════════════════════════════════════════════════════════════════\033[0m"
    echo -e "\033[1;33m Нажмите Enter для запуска production окружения...\033[0m"
    read -p ""
    echo ""
}

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

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker не установлен"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose не установлен"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        error "Docker не запущен или нет прав доступа"
        exit 1
    fi
    
    log "Все зависимости проверены"
}

# Проверка .env файла
check_env_file() {
    local env_file=${1:-.env}
    
    if [[ ! -f "$env_file" ]]; then
        warn "Файл $env_file не найден"
        log "Генерирую новый .env файл..."
        ./scripts/setup/generate_secrets.sh --env "$env_file"
    fi
    
    # Проверяем критичные переменные
    if ! grep -q "KEYCLOAK_ADMIN_PASSWORD" "$env_file"; then
        error "В файле $env_file отсутствует KEYCLOAK_ADMIN_PASSWORD"
        exit 1
    fi
    
    if ! grep -q "KEYCLOAK_DB_PASSWORD" "$env_file"; then
        error "В файле $env_file отсутствует KEYCLOAK_DB_PASSWORD"
        exit 1
    fi
    
    log "Файл $env_file проверен"
}

# Создание необходимых директорий
create_directories() {
    log "Создание необходимых директорий..."
    
    mkdir -p keycloak/data
    mkdir -p keycloak/imports
    mkdir -p logs
    mkdir -p certs
    mkdir -p backups
    
    # Устанавливаем правильные права
    chmod 755 keycloak/data
    chmod 755 keycloak/imports
    chmod 755 logs
    chmod 700 certs
    chmod 755 backups
    
    log "Директории созданы"

    # --- Новый блок: автоматическое обновление публичного ключа ---
    if [[ -f "config/public_key.pem" ]]; then
        log "Обновляю публичный ключ Keycloak из config/public_key.pem..."
        bash scripts/maintenance/update_keycloak_public_key_from_file.sh config/public_key.pem
    else
        warn "Файл config/public_key.pem не найден. Публичный ключ Keycloak не будет обновлён автоматически."
        warn "Скопируйте публичный ключ из Keycloak UI (Realm Settings → Keys → RS256 → Public Key) в config/public_key.pem для автоматизации!"
    fi

    # --- Новый блок: автоматическая генерация .env ---
    if [[ ! -f ".env" ]]; then
        log "Файл .env не найден. Генерирую новый .env..."
        ./scripts/setup/generate_secrets.sh --env .env
    fi
}

# Остановка существующих контейнеров
stop_existing_containers() {
    log "Остановка существующих контейнеров..."
    
    if docker-compose ps | grep -q "Up"; then
        docker-compose down
        log "Существующие контейнеры остановлены"
    else
        log "Активных контейнеров не найдено"
    fi
}

# Запуск сервисов
start_services() {
    local env_file=${1:-.env}
    
    log "Запуск сервисов с файлом $env_file..."
    
    # Загружаем переменные окружения
    export $(grep -v '^#' "$env_file" | xargs)
    
    # Запускаем сервисы
    docker-compose --env-file "$env_file" up -d
    
    log "Сервисы запущены"
    
    # --- Новый блок: автоматическая установка Grafana дашбордов ---
    log "Установка и проверка Grafana дашбордов..."
    if [ -f "scripts/setup/setup_dashboards.sh" ]; then
    chmod +x scripts/setup/setup_dashboards.sh
    ./scripts/setup/setup_dashboards.sh
    else
        warn "Скрипт setup_dashboards.sh не найден. Дашборды не будут установлены автоматически."
    fi
}

# Проверка здоровья сервисов
check_health() {
    log "Проверка здоровья сервисов..."
    
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        log "Попытка $attempt/$max_attempts..."
        
        # Проверяем Keycloak
        if curl -f http://localhost:8080/ &> /dev/null; then
            log "Keycloak запущен и отвечает"
        else
            warn "⏳ Keycloak еще не готов..."
        fi
        
        # Проверяем Relay
        if curl -f http://localhost:8081/metrics &> /dev/null; then
            log "Relay запущен и отвечает"
        else
            warn "⏳ Relay еще не готов..."
        fi
        
        # Проверяем Grafana
        if curl -f http://localhost:3000/api/health &> /dev/null; then
            log "Grafana запущен и отвечает"
        else
            warn "⏳ Grafana еще не готов..."
        fi
        
        # Если все сервисы готовы, выходим
        if curl -f http://localhost:8080/ &> /dev/null && \
           curl -f http://localhost:8081/metrics &> /dev/null && \
           curl -f http://localhost:3000/api/health &> /dev/null; then
            log "🎉 Все сервисы запущены и готовы к работе!"
            return 0
        fi
        
        sleep 10
        ((attempt++))
    done
    
    error "Не удалось запустить все сервисы за отведенное время"
    return 1
}

# Показать статус
show_status() {
    log "Статус сервисов:"
    echo
    docker-compose ps
    echo
    
    # Красивый финальный экран
    echo -e "\033[1;32m"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                                                                              ║"
    echo "║                    🎉 CloudBridge Relay Successfully Started! 🎉            ║"
    echo "║                                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "\033[0m"
    
    log "URLs сервисов:"
    echo -e "  \033[1;36m🔐 Keycloak:\033[0m     http://localhost:8080"
    echo -e "  \033[1;36m🚀 Relay:\033[0m        http://localhost:8081"
    echo -e "  \033[1;36m📊 Grafana:\033[0m      http://localhost:3000"
    echo -e "  \033[1;36m Prometheus:\033[0m   http://localhost:9090"
    echo -e "  \033[1;36mAlertManager:\033[0m http://localhost:9093"
    echo
    
    log "Логи сервисов:"
    echo -e "  \033[1;33mdocker-compose logs -f keycloak\033[0m"
    echo -e "  \033[1;33mdocker-compose logs -f relay\033[0m"
    echo -e "  \033[1;33mdocker-compose logs -f grafana\033[0m"
    
    echo
    echo -e "\033[1;35mQuick Tips:\033[0m"
    echo -e "  • Use \033[1;33m./scripts/start_production.sh --status\033[0m to check service status"
    echo -e "  • Use \033[1;33mdocker-compose logs -f [service]\033[0m to view real-time logs"
    echo -e "  • Use \033[1;33m./scripts/setup/setup_dashboards.sh\033[0m to update Grafana dashboards"
    echo
}

# Показать справку
show_help() {
    echo "CloudBridge Relay - Production Startup Script"
    echo
    echo "Использование: $0 [опции]"
    echo
    echo "Опции:"
    echo "  -e, --env FILE      Использовать .env файл (по умолчанию .env)"
    echo "  -f, --force         Принудительно остановить существующие контейнеры"
    echo "  -s, --status        Показать статус сервисов"
    echo "  -h, --help          Показать эту справку"
    echo
    echo "Примеры:"
    echo "  $0                    # Запуск с .env"
    echo "  $0 --env .env.production  # Запуск с production .env"
    echo "  $0 --status           # Показать статус"
    echo
}

# Основная функция
main() {
    local env_file=".env"
    local force_stop=false
    local show_status_only=false
    
    # Показываем красивую заставку
    show_banner
    
    # Парсинг аргументов
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--env)
                env_file="$2"
                shift 2
                ;;
            -f|--force)
                force_stop=true
                shift
                ;;
            -s|--status)
                show_status_only=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                error "Неизвестная опция: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    if [[ "$show_status_only" == true ]]; then
        show_status
        exit 0
    fi
    
    log "Запуск CloudBridge Relay в production режиме..."
    echo
    
    # Восстановление сертификатов перед запуском сервисов
    if [ -f ./scripts/maintenance/restore_certs.sh ]; then
    ./scripts/maintenance/restore_certs.sh
    fi
    
    check_dependencies
    check_env_file "$env_file"
    create_directories
    
    if [[ "$force_stop" == true ]]; then
        stop_existing_containers
    fi
    
    start_services "$env_file"
    
    log "Ожидание готовности сервисов..."
    if check_health; then
        echo
        show_status
        log "🎉 CloudBridge Relay успешно запущен в production режиме!"
    else
        error "Не удалось запустить все сервисы"
        exit 1
    fi
}

# Запуск скрипта
main "$@" 