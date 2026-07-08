#!/bin/bash

# CloudBridge Relay - Запуск всех тестов
# Этот скрипт запускает все тесты для генерации данных в дашбордах

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Functions
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

log_demo() {
    echo -e "${PURPLE}[DEMO]${NC} $1"
}

log_header() {
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║                    $1                    ║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
}

# Проверка статуса сервисов
check_services() {
    log_info "Проверка статуса сервисов..."
    
    if ! curl -s -f "http://localhost:8081/metrics" > /dev/null 2>&1; then
        log_error "Relay сервер недоступен"
        return 1
    fi
    
    if ! curl -s -f "http://localhost:9090/api/v1/query?query=up" > /dev/null 2>&1; then
        log_error "Prometheus недоступен"
        return 1
    fi
    
    if ! curl -s -f "http://localhost:3000/api/health" > /dev/null 2>&1; then
        log_error "Grafana недоступна"
        return 1
    fi
    
    log_success "Все сервисы доступны"
    return 0
}

# Показ текущих метрик
show_current_metrics() {
    log_info "Текущие метрики:"
    
    local active_connections=$(curl -s http://localhost:8081/metrics | grep "relay_active_connections" | grep -o '[0-9]*$')
    local total_connections=$(curl -s http://localhost:8081/metrics | grep "relay_connections_total" | grep -o '[0-9]*$')
    local active_tunnels=$(curl -s http://localhost:8081/metrics | grep "relay_active_tunnels" | grep -o '[0-9]*$')
    local active_users=$(curl -s http://localhost:8081/metrics | grep "relay_business_active_users" | grep -o '[0-9]*$')
    local total_revenue=$(curl -s http://localhost:8081/metrics | grep "relay_business_total_revenue" | grep -o '[0-9]*$')
    
    echo "  Активные соединения: $active_connections"
    echo "  Всего соединений: $total_connections"
    echo "  Активные туннели: $active_tunnels"
    echo "  Активные пользователи: $active_users"
    echo "  Общий доход: $total_revenue"
    echo
}

# Запуск тестов соединений
run_connection_tests() {
    log_header "ТЕСТИРОВАНИЕ СОЕДИНЕНИЙ"
    
    log_info "Запуск тестов реальных соединений..."
    ./scripts/utils/create_real_connections.sh &
    local connection_pid=$!
    
    log_info "Запуск тестов VPN клиента..."
    ./scripts/testing/test_vpn_client.sh &
    local vpn_pid=$!
    
    echo "PID тестов соединений: $connection_pid"
    echo "PID тестов VPN: $vpn_pid"
    
    # Ждем немного
    sleep 30
    
    return 0
}

# Запуск тестов бизнес-метрик
run_business_tests() {
    log_header "ТЕСТИРОВАНИЕ БИЗНЕС-МЕТРИК"
    
    log_info "Запуск генерации бизнес-метрик..."
    ./scripts/utils/generate_business_metrics.sh &
    local business_pid=$!
    
    echo "PID бизнес-метрик: $business_pid"
    
    # Ждем немного
    sleep 30
    
    return 0
}

# Запуск тестов метрик
run_metrics_tests() {
    log_header "ТЕСТИРОВАНИЕ МЕТРИК"
    
    log_info "Запуск тестов метрик..."
    ./scripts/test_metrics.sh &
    local metrics_pid=$!
    
    echo "PID тестов метрик: $metrics_pid"
    
    # Ждем немного
    sleep 30
    
    return 0
}

# Запуск демонстрации дашбордов
run_dashboard_demo() {
    log_header "ДЕМОНСТРАЦИЯ ДАШБОРДОВ"
    
    log_info "Запуск демонстрации дашбордов..."
    ../../tests/demo/demo_monitor.sh
    
    return 0
}

# Проверка результатов
check_results() {
    log_header "ПРОВЕРКА РЕЗУЛЬТАТОВ"
    
    log_info "Ожидание 60 секунд для сбора данных..."
    sleep 60
    
    show_current_metrics
    
    # Проверка метрик в Prometheus
    log_info "Проверка метрик в Prometheus..."
    
    local active_connections=$(curl -s "http://localhost:9090/api/v1/query?query=relay_active_connections" | jq -r '.data.result[0].value[1] // "0"')
    local total_connections=$(curl -s "http://localhost:9090/api/v1/query?query=relay_connections_total" | jq -r '.data.result[0].value[1] // "0"')
    local active_tunnels=$(curl -s "http://localhost:9090/api/v1/query?query=relay_active_tunnels" | jq -r '.data.result[0].value[1] // "0"')
    
    echo "Prometheus - Активные соединения: $active_connections"
    echo "Prometheus - Всего соединений: $total_connections"
    echo "Prometheus - Активные туннели: $active_tunnels"
    
    # Проверка дашбордов Grafana
    local password=$(grep "^GRAFANA_ADMIN_PASSWORD=" .env | cut -d'=' -f2 2>/dev/null || echo "change-this-password-immediately")
    local dashboards=$(curl -s -u "admin:$password" "http://localhost:3000/api/search" | jq '.[].title' | wc -l)
    
    echo "Grafana - Количество дашбордов: $dashboards"
    
    if [[ $active_connections -gt 0 ]]; then
        log_success "Тесты успешно создали данные!"
        log_info "Дашборды должны показывать активность"
    else
        log_warning "Активные соединения не обнаружены"
        log_info "Возможно, соединения отклоняются сервером"
    fi
}

# Очистка процессов
cleanup_processes() {
    log_info "Очистка всех процессов..."
    
    pkill -f "create_real_connections.sh" 2>/dev/null || true
    pkill -f "test_vpn_client.sh" 2>/dev/null || true
    pkill -f "generate_business_metrics.sh" 2>/dev/null || true
    pkill -f "test_metrics.sh" 2>/dev/null || true
    pkill -f "vpn_client" 2>/dev/null || true
    pkill -f "timeout.*nc" 2>/dev/null || true
    pkill -f "bash.*tunnel_data" 2>/dev/null || true
    pkill -f "curl.*metrics/simulate" 2>/dev/null || true
    
    log_success "Все процессы очищены"
}

# Показ URL для доступа
show_access_urls() {
    log_header "ДОСТУП К СЕРВИСАМ"
    
    local server_ip=$(hostname -I | awk '{print $1}')
    local password=$(grep "^GRAFANA_ADMIN_PASSWORD=" .env | cut -d'=' -f2 2>/dev/null || echo "change-this-password-immediately")
    
    echo -e "${GREEN}Grafana:${NC} http://$server_ip:3000"
    echo -e "${GREEN}   Логин:${NC} admin"
    echo -e "${GREEN}   Пароль:${NC} $password"
    echo
    echo -e "${GREEN} Prometheus:${NC} http://$server_ip:9090"
    echo
    echo -e "${GREEN} Relay Metrics:${NC} http://$server_ip:8081/metrics"
    echo
}

# Основная функция
main() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║           CloudBridge Relay - Все тесты                     ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    
    # Проверка сервисов
    if ! check_services; then
        exit 1
    fi
    
    # Показываем текущие метрики
    show_current_metrics
    
    # Очистка предыдущих процессов
    cleanup_processes
    
    # Запуск всех тестов
    run_connection_tests
    run_business_tests
    run_metrics_tests
    
    log_info "Все тесты запущены!"
    log_info "Ожидание сбора данных..."
    
    # Проверка результатов
    check_results
    
    # Демонстрация дашбордов
    run_dashboard_demo
    
    # Показ URL для доступа
    show_access_urls
    
    log_success "Все тесты завершены!"
    log_info "Проверьте дашборды в Grafana для просмотра данных"
    
    # Очистка
    cleanup_processes
}

# Обработка Ctrl+C
trap cleanup_processes INT

# Запуск основной функции
main "$@" 