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
    echo "║                           Production Environment v2.1                        ║"
    echo "║                                                                              ║"
    echo "║  Multi-Tenant Authentication   |  High-Performance TCP Proxy                 ║"
    echo "║  Enterprise Security           |  Advanced Monitoring & Analytics            ║"
    echo "║  Auto-Scaling & Load Balancing |  Real-time Metrics & Dashboards             ║"
    echo "║  Custom Keycloak Theme         |  Playa Design System                        ║"
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
    mkdir -p keycloak/themes
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
    
    # --- Новый блок: автоматическая настройка темы Keycloak ---
    log "Настройка кастомной темы Playa для Keycloak..."
    if [ -f "scripts/setup/setup_keycloak_theme.sh" ]; then
        chmod +x scripts/setup/setup_keycloak_theme.sh
        # Проверяем, что переменные окружения установлены
        if [[ -n "$KEYCLOAK_ADMIN" && -n "$KEYCLOAK_ADMIN_PASSWORD" ]]; then
            log "Настраиваю тему Playa для Keycloak..."
            ./scripts/setup/setup_keycloak_theme.sh
        else
            warn "KEYCLOAK_ADMIN или KEYCLOAK_ADMIN_PASSWORD не установлены"
            warn "Тема Keycloak не будет настроена автоматически"
            warn "Запустите вручную: ./scripts/setup/setup_keycloak_theme.sh"
        fi
    else
        warn "Скрипт setup_keycloak_theme.sh не найден. Тема Keycloak не будет настроена автоматически."
    fi

    # --- Новый блок: запуск ML сервиса ---
    log "Запуск ML Anomaly Detection API..."
    if [ -f "ml/anomaly_api.py" ]; then
        # Останавливаем предыдущий процесс ML сервиса, если он запущен
        pkill -f "uvicorn anomaly_api:app" 2>/dev/null || true
        sleep 2
        
        # Запускаем ML сервис в фоне
        cd ml && nohup uvicorn anomaly_api:app --reload --host 0.0.0.0 --port 8000 > ../logs/ml_api.log 2>&1 &
        cd ..
        
        log "ML сервис запущен на порту 8000"
    else
        warn "Файл ml/anomaly_api.py не найден. ML сервис не будет запущен."
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
        
        # Проверяем ML сервис
        if curl -f http://localhost:8000/health &> /dev/null; then
            log "ML сервис запущен и отвечает"
        else
            warn "⏳ ML сервис еще не готов..."
        fi
        
        # Если все сервисы готовы, выходим
        if curl -f http://localhost:8080/ &> /dev/null && \
           curl -f http://localhost:8081/metrics &> /dev/null && \
           curl -f http://localhost:3000/api/health &> /dev/null && \
           curl -f http://localhost:8000/health &> /dev/null; then
            log "🎉 Все сервисы запущены и готовы к работе!"
            
            # Проверяем наличие темы
            if [ -d "keycloak/themes/playa" ]; then
                log "Кастомная тема Playa найдена"
            else
                warn "  Кастомная тема Playa не найдена в keycloak/themes/playa"
            fi
            
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
    echo -e "  \033[1;36m📈 Prometheus:\033[0m   http://localhost:9090"
    echo -e "  \033[1;36mAlertManager:\033[0m http://localhost:9093"
    echo -e "  \033[1;36m ML API:\033[0m       http://localhost:8000"
    echo
    log "Тема Keycloak:"
    echo -e "  \033[1;36m🎨 Кастомная тема:\033[0m Playa (автоматически настроена)"
    echo -e "  \033[1;36m Страница входа:\033[0m http://localhost:8080/realms/cloudbridge/protocol/openid-connect/auth"
    echo
    
    log "Логи сервисов:"
    echo -e "  \033[1;33mdocker-compose logs -f keycloak\033[0m"
    echo -e "  \033[1;33mdocker-compose logs -f relay\033[0m"
    echo -e "  \033[1;33mdocker-compose logs -f grafana\033[0m"
    echo -e "  \033[1;33mtail -f logs/ml_api.log\033[0m"
    
    echo
    echo -e "\033[1;35mQuick Tips:\033[0m"
    echo -e "  • Use \033[1;33m./scripts/start_production.sh --status\033[0m to check service status"
    echo -e "  • Use \033[1;33mdocker-compose logs -f [service]\033[0m to view real-time logs"
    echo -e "  • Use \033[1;33m./scripts/setup/setup_dashboards.sh\033[0m to update Grafana dashboards"
    echo -e "  • Use \033[1;33m./scripts/setup/setup_keycloak_theme.sh\033[0m to reconfigure Keycloak theme"
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
    check_keycloak_theme
    
    if [[ "$force_stop" == true ]]; then
        stop_existing_containers
    fi
    
    # Вызов проверки сертификатов перед запуском сервисов
    check_certificates
    
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

# Добавляем новую функцию для проверки сертификатов
check_certificates() {
    if [[ ! -f "certs/fullchain.pem" || ! -f "certs/privkey.pem" ]]; then
        warn "ВНИМАНИЕ: Для HTTPS Keycloak нужны certs/fullchain.pem и certs/privkey.pem!"
        warn "Keycloak не сможет стартовать по HTTPS (443) без этих файлов."
    else
        log "SSL-сертификаты для Keycloak найдены. HTTPS будет работать."
    fi
}

# Добавляем функцию для проверки темы Keycloak
check_keycloak_theme() {
    log "Проверка кастомной темы Keycloak..."
    
    if [ -d "keycloak/themes/playa" ]; then
        local theme_files=$(find keycloak/themes/playa -type f | wc -l)
        log "Кастомная тема Playa найдена ($theme_files файлов)"
        
        # Проверяем основные файлы темы
        if [ -f "keycloak/themes/playa/theme.properties" ]; then
            log "Основной файл конфигурации темы найден"
        else
            warn "  Файл theme.properties не найден"
        fi
        
        if [ -f "keycloak/themes/playa/login/css/style.css" ]; then
            log "CSS файл темы найден"
        else
            warn "  CSS файл темы не найден"
        fi
        
        if [ -f "keycloak/themes/playa/login/index.ftl" ]; then
            log "Шаблон входа найден"
        else
            warn "  Шаблон входа не найден"
        fi
        
    else
        warn "  Кастомная тема Playa не найдена в keycloak/themes/playa"
        warn "Тема не будет применена автоматически"
    fi
} 