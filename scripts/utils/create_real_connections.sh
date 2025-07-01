#!/bin/bash

# CloudBridge Relay - Создание реальных соединений
# Этот скрипт создает реальные TCP соединения к relay серверу

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
    
    if ! curl -s -f "http://localhost:9090/api/v1/query?query=up" > /dev/null 2>&1; then
        log_error "Prometheus недоступен"
        return 1
    fi
    
    log_success "Все сервисы доступны"
    return 0
}

# Создание тестового токена
create_test_token() {
    log_info "Создание тестового JWT токена..."
    
    # Простой JWT токен для тестирования (не для продакшена)
    local header='{"alg":"HS256","typ":"JWT"}'
    local payload='{"sub":"test-client-1","iat":'$(date +%s)',"exp":'$(($(date +%s) + 3600))',"aud":"relay-client"}'
    
    # Base64 encode (без padding)
    local header_b64=$(echo -n "$header" | base64 | tr -d '=' | tr '/+' '_-')
    local payload_b64=$(echo -n "$payload" | base64 | tr -d '=' | tr '/+' '_-')
    
    # Простая подпись (для демо)
    local signature=$(echo -n "$header_b64.$payload_b64" | openssl dgst -sha256 -hmac "test-secret" -binary | base64 | tr -d '=' | tr '/+' '_-')
    
    echo "$header_b64.$payload_b64.$signature"
}

# Создание реальных TCP соединений
create_real_connections() {
    local count=${1:-5}
    local duration=${2:-60}
    
    log_info "Создание $count реальных TCP соединений на $duration секунд..."
    
    local test_token=$(create_test_token)
    
    for i in $(seq 1 $count); do
        (
            log_info "Создание соединения $i..."
            
            # Создаем TCP соединение к relay серверу
            exec 3<>/dev/tcp/localhost/8080
            
            # Отправляем hello сообщение
            echo '{"type":"hello","version":"1.0","capabilities":["tls","heartbeat","tunnel_info"]}' >&3
            
            # Отправляем auth сообщение
            echo "{\"type\":\"auth\",\"token\":\"$test_token\"}" >&3
            
            # Отправляем heartbeat сообщения
            timeout $duration bash -c "
                while true; do
                    echo '{\"type\":\"heartbeat\",\"timestamp\":'$(date +%s)'}' >&3
                    sleep 10
                done
            " 2>/dev/null &
            
            # Отправляем tunnel_info сообщения
            timeout $duration bash -c "
                while true; do
                    echo '{\"type\":\"tunnel_info\",\"tunnels\":[{\"id\":\"tunnel-'$i'\",\"remote_host\":\"192.168.1.'$i'\",\"remote_port\":8080,\"bytes_sent\":1024,\"bytes_received\":2048,\"active_connections\":1}]}' >&3
                    sleep 15
                done
            " 2>/dev/null &
            
            sleep 2
        ) &
        
        sleep 1
    done
    
    log_info "Запущено $count фоновых соединений"
}

# Создание тестовых туннелей
create_test_tunnels() {
    local count=${1:-3}
    local duration=${2:-90}
    
    log_info "Создание $count тестовых туннелей на $duration секунд..."
    
    for i in $(seq 1 $count); do
        (
            log_info "Создание туннеля $i..."
            
            # Создаем UDP соединение (имитация туннеля)
            timeout $duration bash -c "
                while true; do
                    echo 'tunnel_data_$i' | nc -u localhost 8080 2>/dev/null || true
                    sleep 5
                done
            " 2>/dev/null &
        ) &
        
        sleep 2
    done
    
    log_info "Запущено $count фоновых туннелей"
}

# Генерация бизнес-метрик
generate_business_metrics() {
    log_info "Генерация бизнес-метрик..."
    
    # Симуляция активных пользователей
    local active_users=$((RANDOM % 50 + 10))
    local revenue=$((RANDOM % 1000 + 100))
    local satisfaction=$((RANDOM % 30 + 70))
    
    log_info "Симулируем: $active_users активных пользователей, $revenue дохода, $satisfaction% удовлетворенности"
    
    # Отправляем метрики через curl (если есть API endpoint)
    curl -s -X POST "http://localhost:8081/api/v1/metrics/business" \
        -H "Content-Type: application/json" \
        -d "{\"active_users\":$active_users,\"revenue\":$revenue,\"satisfaction\":$satisfaction}" \
        2>/dev/null || log_warning "API endpoint для бизнес-метрик недоступен"
}

# Проверка метрик после создания соединений
check_metrics_after_connections() {
    log_info "Проверка метрик после создания соединений..."
    
    sleep 10
    
    local active_connections=$(curl -s http://localhost:8081/metrics | grep "relay_active_connections" | grep -o '[0-9]*$')
    local total_connections=$(curl -s http://localhost:8081/metrics | grep "relay_connections_total" | grep -o '[0-9]*$')
    local active_tunnels=$(curl -s http://localhost:8081/metrics | grep "relay_active_tunnels" | grep -o '[0-9]*$')
    
    log_info "Активные соединения: $active_connections"
    log_info "Всего соединений: $total_connections"
    log_info "Активные туннели: $active_tunnels"
    
    if [[ $active_connections -gt 0 ]]; then
        log_success "Реальные соединения успешно созданы!"
        log_info "Теперь дашборды должны показывать данные"
    else
        log_warning "Активные соединения не обнаружены"
        log_info "Возможно, соединения отклоняются сервером"
    fi
}

# Очистка процессов
cleanup_processes() {
    log_info "Очистка процессов..."
    
    pkill -f "timeout.*nc" 2>/dev/null || true
    pkill -f "bash.*tunnel_data" 2>/dev/null || true
    
    log_success "Процессы очищены"
}

# Основная функция
main() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║           CloudBridge Relay - Реальные соединения            ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    
    # Проверка сервисов
    if ! check_services; then
        exit 1
    fi
    
    # Очистка предыдущих процессов
    cleanup_processes
    
    # Создание реальных соединений
    create_real_connections 8 120
    create_test_tunnels 5 120
    generate_business_metrics
    
    log_info "Соединения созданы!"
    log_info "Дашборды должны начать показывать данные через 10-30 секунд"
    log_info "Для остановки нажмите Ctrl+C"
    
    # Ожидание и проверка
    sleep 15
    check_metrics_after_connections
    
    # Ожидание завершения
    log_info "Ожидание завершения соединений (120 секунд)..."
    sleep 120
    
    # Финальная проверка
    check_metrics_after_connections
    
    # Очистка
    cleanup_processes
    
    log_success "Демонстрация завершена!"
    log_info "Проверьте дашборды в Grafana: http://$(hostname -I | awk '{print $1}'):3000"
}

# Обработка Ctrl+C
trap cleanup_processes INT

# Запуск основной функции
main "$@" 