#!/bin/bash

# CloudBridge Relay Dashboard Setup Script
# Автоматическая установка и настройка дашбордов Grafana

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose не установлен"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl не установлен"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Проверка статуса сервисов
check_services() {
    log_info "Проверка статуса сервисов..."
    
    # Проверка Grafana
    if ! docker-compose ps grafana | grep -q "Up"; then
        log_warning "Grafana не запущен. Запускаем сервисы..."
        docker-compose up -d grafana prometheus
        sleep 10
    fi
    
    # Проверка Prometheus
    if ! docker-compose ps prometheus | grep -q "Up"; then
        log_warning "Prometheus не запущен. Запускаем..."
        docker-compose up -d prometheus
        sleep 5
    fi
    
    log_success "Сервисы запущены"
}

# Ожидание готовности Grafana
wait_for_grafana() {
    log_info "Ожидание готовности Grafana..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:3000/api/health | grep -q "ok"; then
            log_success "Grafana готов"
            return 0
        fi
        
        log_info "Попытка $attempt/$max_attempts - Grafana еще не готов..."
        sleep 2
        ((attempt++))
    done
    
    log_error "Grafana не запустился в течение 60 секунд"
    return 1
}

# Получение API ключа Grafana
get_grafana_api_key() {
    log_info "Получение API ключа Grafana..."
    
    # Создание API ключа
    local api_key_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "name": "cloudbridge-relay",
            "role": "Admin"
        }' \
        http://admin:admin@localhost:3000/api/auth/keys)
    
    if echo "$api_key_response" | grep -q "key"; then
        local api_key=$(echo "$api_key_response" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
        echo "$api_key"
        log_success "API ключ получен"
    else
        log_error "Не удалось получить API ключ"
        return 1
    fi
}

# Установка дашбордов
install_dashboards() {
    local api_key=$1
    
    log_info "Установка дашбордов..."
    
    local dashboards=(
        "relay_dashboard.json:CloudBridge Relay Dashboard"
        "tunnel_management_dashboard.json:Tunnel Management Dashboard"
        "security_analytics_dashboard.json:Security Analytics Dashboard"
        "multi_tenant_analytics_dashboard.json:Multi-Tenant Analytics Dashboard"
    )
    
    for dashboard in "${dashboards[@]}"; do
        local file="${dashboard%%:*}"
        local title="${dashboard##*:}"
        
        log_info "Установка дашборда: $title"
        
        if [ -f "deploy/grafana/dashboards/$file" ]; then
            # Чтение и модификация дашборда
            local dashboard_json=$(cat "deploy/grafana/dashboards/$file")
            
            # Установка дашборда через API
            local response=$(curl -s -X POST \
                -H "Content-Type: application/json" \
                -H "Authorization: Bearer $api_key" \
                -d "$dashboard_json" \
                http://localhost:3000/api/dashboards/db)
            
            if echo "$response" | grep -q '"status":"success"'; then
                log_success "Дашборд '$title' установлен"
            else
                log_warning "Не удалось установить дашборд '$title' (возможно, уже существует)"
            fi
        else
            log_error "Файл дашборда не найден: deploy/grafana/dashboards/$file"
        fi
    done
}

# Настройка Prometheus
setup_prometheus() {
    log_info "Настройка Prometheus..."
    
    # Проверка конфигурации Prometheus
    if [ -f "prometheus/prometheus.yml" ]; then
        log_info "Конфигурация Prometheus найдена"
        
        # Проверка наличия job для relay
        if ! grep -q "job_name.*relay" prometheus/prometheus.yml; then
            log_warning "Job для relay не найден в конфигурации Prometheus"
            log_info "Добавьте в prometheus/prometheus.yml:"
            echo "  - job_name: 'relay'"
            echo "    static_configs:"
            echo "      - targets: ['relay:8081']"
            echo "    metrics_path: '/metrics'"
        fi
    else
        log_warning "Файл конфигурации Prometheus не найден"
    fi
}

# Настройка алертов
setup_alerts() {
    log_info "Настройка алертов..."
    
    if [ -f "deploy/prometheus/alerts/relay_alerts.yml" ]; then
        log_info "Файл алертов найден"
        
        # Проверка конфигурации AlertManager
        if [ -f "deploy/alertmanager/alertmanager.yml" ]; then
            log_success "AlertManager настроен"
        else
            log_warning "Файл конфигурации AlertManager не найден"
        fi
    else
        log_warning "Файл алертов не найден"
    fi
}

# Проверка метрик
check_metrics() {
    log_info "Проверка доступности метрик..."
    
    # Проверка метрик relay
    if curl -s http://localhost:8081/metrics | grep -q "relay_"; then
        log_success "Метрики relay доступны"
    else
        log_warning "Метрики relay недоступны (возможно, relay не запущен)"
    fi
    
    # Проверка метрик Prometheus
    if curl -s http://localhost:9090/api/v1/targets | grep -q "relay"; then
        log_success "Prometheus видит relay"
    else
        log_warning "Prometheus не видит relay"
    fi
}

# Создание пользователей Grafana
create_grafana_users() {
    local api_key=$1
    
    log_info "Создание пользователей Grafana..."
    
    # Создание пользователя для мониторинга
    local user_response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $api_key" \
        -d '{
            "name": "monitor",
            "email": "monitor@cloudbridge.local",
            "login": "monitor",
            "password": "monitor123"
        }' \
        http://localhost:3000/api/admin/users)
    
    if echo "$user_response" | grep -q '"id"'; then
        log_success "Пользователь monitor создан"
    else
        log_warning "Не удалось создать пользователя monitor (возможно, уже существует)"
    fi
}

# Основная функция
main() {
    log_info "Начало установки дашбордов CloudBridge Relay"
    
    # Проверка директории
    if [ ! -f "docker-compose.yml" ]; then
        log_error "Скрипт должен быть запущен из корневой директории проекта"
        exit 1
    fi
    
    # Проверка зависимостей
    check_dependencies
    
    # Проверка сервисов
    check_services
    
    # Ожидание готовности Grafana
    if ! wait_for_grafana; then
        exit 1
    fi
    
    # Получение API ключа
    local api_key=$(get_grafana_api_key)
    if [ -z "$api_key" ]; then
        exit 1
    fi
    
    # Установка дашбордов
    install_dashboards "$api_key"
    
    # Настройка Prometheus
    setup_prometheus
    
    # Настройка алертов
    setup_alerts
    
    # Проверка метрик
    check_metrics
    
    # Создание пользователей
    create_grafana_users "$api_key"
    
    log_success "Установка дашбордов завершена!"
    echo
    echo "Доступ к дашбордам:"
    echo "  Grafana: http://localhost:3000"
    echo "  Логин: admin/admin или monitor/monitor123"
    echo "  Prometheus: http://localhost:9090"
    echo
    echo "Доступные дашборды:"
    echo "  - CloudBridge Relay Dashboard"
    echo "  - Tunnel Management Dashboard"
    echo "  - Security Analytics Dashboard"
    echo "  - Multi-Tenant Analytics Dashboard"
    echo
    echo "Документация: docs/dashboard_guide.md"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay Dashboard Setup Script"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --check        Только проверить статус"
        echo
        echo "Примеры:"
        echo "  $0              Установить все дашборды"
        echo "  $0 --check      Проверить статус сервисов"
        ;;
    --check)
        check_dependencies
        check_services
        check_metrics
        log_success "Проверка завершена"
        ;;
    *)
        main
        ;;
esac 