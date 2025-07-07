#!/bin/bash

# CloudBridge Relay - Генерация бизнес-метрик
# Этот скрипт генерирует тестовые бизнес-метрики для дашбордов

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

# Проверка статуса сервисов
check_services() {
    log_info "Проверка статуса сервисов..."
    
    if ! curl -s -f "http://localhost:8081/metrics" > /dev/null 2>&1; then
        log_error "Relay сервер недоступен"
        return 1
    fi
    
    log_success "Relay сервер доступен"
    return 0
}

# Генерация случайных бизнес-метрик
generate_random_business_metrics() {
    local active_users=$((RANDOM % 100 + 20))
    local revenue=$((RANDOM % 5000 + 500))
    local satisfaction=$((RANDOM % 40 + 60))
    local total_revenue=$((RANDOM % 50000 + 10000))
    
    echo "active_users=$active_users"
    echo "revenue=$revenue"
    echo "satisfaction=$satisfaction"
    echo "total_revenue=$total_revenue"
}

# Симуляция активных пользователей
simulate_active_users() {
    local duration=${1:-300}
    
    log_info "Симуляция активных пользователей на $duration секунд..."
    
    (
        timeout $duration bash -c "
            while true; do
                # Генерируем случайное количество активных пользователей
                local users=\$((RANDOM % 100 + 20))
                
                # Отправляем метрику через curl (имитация)
                curl -s -X POST http://localhost:8081/metrics/simulate \
                    -H 'Content-Type: application/json' \
                    -d '{\"active_users\":'\$users'}' \
                    2>/dev/null || true
                
                sleep 30
            done
        " 2>/dev/null &
    ) &
    
    log_info "Симуляция активных пользователей запущена"
}

# Симуляция дохода
simulate_revenue() {
    local duration=${1:-300}
    
    log_info "Симуляция дохода на $duration секунд..."
    
    (
        timeout $duration bash -c "
            while true; do
                # Генерируем случайный доход
                local revenue=\$((RANDOM % 1000 + 100))
                local total_revenue=\$((RANDOM % 10000 + 5000))
                
                # Отправляем метрики через curl (имитация)
                curl -s -X POST http://localhost:8081/metrics/simulate \
                    -H 'Content-Type: application/json' \
                    -d '{\"revenue_per_hour\":'\$revenue',\"total_revenue\":'\$total_revenue'}' \
                    2>/dev/null || true
                
                sleep 60
            done
        " 2>/dev/null &
    ) &
    
    log_info "Симуляция дохода запущена"
}

# Симуляция удовлетворенности клиентов
simulate_customer_satisfaction() {
    local duration=${1:-300}
    
    log_info "Симуляция удовлетворенности клиентов на $duration секунд..."
    
    (
        timeout $duration bash -c "
            while true; do
                # Генерируем случайную удовлетворенность (60-95%)
                local satisfaction=\$((RANDOM % 35 + 60))
                
                # Отправляем метрику через curl (имитация)
                curl -s -X POST http://localhost:8081/metrics/simulate \
                    -H 'Content-Type: application/json' \
                    -d '{\"customer_satisfaction\":'\$satisfaction'}' \
                    2>/dev/null || true
                
                sleep 120
            done
        " 2>/dev/null &
    ) &
    
    log_info "Симуляция удовлетворенности клиентов запущена"
}

# Создание тестовых данных для мультитенантности
create_multi_tenant_data() {
    local duration=${1:-300}
    
    log_info "Создание тестовых данных мультитенантности на $duration секунд..."
    
    local tenants=("company-a" "company-b" "company-c" "startup-x" "enterprise-y")
    
    (
        timeout $duration bash -c "
            while true; do
                for tenant in \"\${tenants[@]}\"; do
                    # Генерируем случайные метрики для каждого тенанта
                    local tunnels=\$((RANDOM % 10 + 1))
                    local connections=\$((RANDOM % 50 + 5))
                    local bandwidth=\$((RANDOM % 1000 + 100))
                    local revenue=\$((RANDOM % 500 + 50))
                    
                    # Отправляем метрики через curl (имитация)
                    curl -s -X POST http://localhost:8081/metrics/tenant \
                        -H 'Content-Type: application/json' \
                        -d '{\"tenant_id\":\"'\$tenant'\",\"active_tunnels\":'\$tunnels',\"active_connections\":'\$connections',\"bandwidth_bytes\":'\$bandwidth',\"revenue\":'\$revenue'}' \
                        2>/dev/null || true
                done
                
                sleep 45
            done
        " 2>/dev/null &
    ) &
    
    log_info "Тестовые данные мультитенантности созданы"
}

# Проверка текущих бизнес-метрик
check_business_metrics() {
    log_info "Проверка текущих бизнес-метрик..."
    
    local active_users=$(curl -s http://localhost:8081/metrics | grep "relay_business_active_users" | grep -o '[0-9]*$')
    local total_revenue=$(curl -s http://localhost:8081/metrics | grep "relay_business_total_revenue" | grep -o '[0-9]*$')
    local satisfaction=$(curl -s http://localhost:8081/metrics | grep "relay_business_customer_satisfaction" | grep -o '[0-9]*$')
    
    echo "  Активные пользователи: $active_users"
    echo "  Общий доход: $total_revenue"
    echo "  Удовлетворенность: $satisfaction%"
    
    if [[ $active_users -gt 0 ]]; then
        log_success "Бизнес-метрики генерируются!"
    else
        log_warning "Бизнес-метрики не обнаружены"
    fi
}

# Очистка процессов
cleanup_processes() {
    log_info "Очистка процессов..."
    
    pkill -f "timeout.*bash.*simulate" 2>/dev/null || true
    pkill -f "curl.*metrics/simulate" 2>/dev/null || true
    pkill -f "curl.*metrics/tenant" 2>/dev/null || true
    
    log_success "Процессы очищены"
}

# Основная функция
main() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║         CloudBridge Relay - Бизнес-метрики                   ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    
    # Проверка сервисов
    if ! check_services; then
        exit 1
    fi
    
    # Очистка предыдущих процессов
    cleanup_processes
    
    # Показываем текущие метрики
    check_business_metrics
    
    # Запуск симуляции
    simulate_active_users 600
    simulate_revenue 600
    simulate_customer_satisfaction 600
    create_multi_tenant_data 600
    
    log_info "Бизнес-метрики запущены!"
    log_info "Дашборды должны начать показывать данные через 30-60 секунд"
    log_info "Для остановки нажмите Ctrl+C"
    
    # Ожидание и проверка
    sleep 30
    check_business_metrics
    
    # Ожидание завершения
    log_info "Ожидание завершения симуляции (600 секунд)..."
    sleep 600
    
    # Финальная проверка
    check_business_metrics
    
    # Очистка
    cleanup_processes
    
    log_success "Генерация бизнес-метрик завершена!"
    log_info "Проверьте дашборды в Grafana: http://$(hostname -I | awk '{print $1}'):3000"
}

# Обработка Ctrl+C
trap cleanup_processes INT

# Запуск основной функции
main "$@" 