#!/bin/bash

# CloudBridge Relay - Тестирование TLS соединений
# Этот скрипт создает правильные TLS соединения к relay серверу

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

# Создание тестового токена
create_test_token() {
    log_info "Создание тестового JWT токена..."
    
    # Простой JWT токен для тестирования
    local header='{"alg":"HS256","typ":"JWT"}'
    local payload='{"sub":"tls-test-client","iat":'$(date +%s)',"exp":'$(($(date +%s) + 3600))',"aud":"relay-client"}'
    
    # Base64 encode (без padding)
    local header_b64=$(echo -n "$header" | base64 | tr -d '=' | tr '/+' '_-')
    local payload_b64=$(echo -n "$payload" | base64 | tr -d '=' | tr '/+' '_-')
    
    # Простая подпись (для демо)
    local signature=$(echo -n "$header_b64.$payload_b64" | openssl dgst -sha256 -hmac "test-secret" -binary | base64 | tr -d '=' | tr '/+' '_-')
    
    echo "$header_b64.$payload_b64.$signature"
}

# Создание TLS соединений с помощью OpenSSL
create_tls_connections() {
    local count=${1:-3}
    local duration=${2:-60}
    
    log_info "Создание $count TLS соединений на $duration секунд..."
    
    local test_token=$(create_test_token)
    
    for i in $(seq 1 $count); do
        (
            log_info "Создание TLS соединения $i..."
            
            # Создаем TLS соединение с помощью OpenSSL
            timeout $duration openssl s_client \
                -connect localhost:8080 \
                -servername localhost \
                -verify_return_error \
                -quiet \
                -ign_eof \
                -no_ign_eof \
                <<< "test_connection_$i" &
            
            local openssl_pid=$!
            log_info "TLS соединение $i создано (PID: $openssl_pid)"
            
            # Ждем завершения
            wait $openssl_pid 2>/dev/null || true
            
        ) &
        
        sleep 2
    done
    
    log_info "Запущено $count TLS соединений"
}

# Создание простых TCP соединений (без TLS)
create_plain_connections() {
    local count=${1:-2}
    local duration=${2:-30}
    
    log_info "Создание $count простых TCP соединений на $duration секунд..."
    
    for i in $(seq 1 $count); do
        (
            log_info "Создание простого соединения $i..."
            
            # Простое TCP соединение (должно вызвать ошибки)
            timeout $duration bash -c "
                exec 3<>/dev/tcp/localhost/8080
                echo 'plain_connection_$i' >&3
                sleep $duration
                exec 3>&-
            " 2>/dev/null &
            
        ) &
        
        sleep 1
    done
    
    log_info "Запущено $count простых соединений"
}

# Проверка метрик после тестирования
check_metrics_after_test() {
    log_info "Проверка метрик после тестирования..."
    
    sleep 10
    
    local active_connections=$(curl -s http://localhost:8081/metrics | grep "relay_active_connections" | grep -o '[0-9]*$')
    local total_connections=$(curl -s http://localhost:8081/metrics | grep "relay_connections_total" | grep -o '[0-9]*$')
    local rejected_connections=$(curl -s http://localhost:8081/metrics | grep "relay_rejected_connections_total" | grep -o '[0-9]*$')
    
    log_info "Активные соединения: $active_connections"
    log_info "Всего соединений: $total_connections"
    log_info "Отклоненные соединения: $rejected_connections"
    
    if [[ $total_connections -gt 0 ]]; then
        log_success "Соединения созданы!"
        log_info "Теперь дашборды должны показывать данные"
    else
        log_warning "Соединения не обнаружены"
    fi
}

# Очистка процессов
cleanup_processes() {
    log_info "Очистка процессов..."
    
    pkill -f "openssl s_client" 2>/dev/null || true
    pkill -f "timeout.*openssl" 2>/dev/null || true
    pkill -f "timeout.*bash.*plain_connection" 2>/dev/null || true
    
    log_success "Процессы очищены"
}

# Основная функция
main() {
    echo -e "${PURPLE}╔══════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${PURPLE}║         CloudBridge Relay - TLS соединения                  ║${NC}"
    echo -e "${PURPLE}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo
    
    # Проверка сервисов
    if ! check_services; then
        exit 1
    fi
    
    # Очистка предыдущих процессов
    cleanup_processes
    
    # Создание соединений
    create_tls_connections 3 90
    create_plain_connections 2 60
    
    log_info "Тестирование запущено!"
    log_info "Дашборды должны начать показывать данные через 10-30 секунд"
    log_info "Для остановки нажмите Ctrl+C"
    
    # Ожидание и проверка
    sleep 15
    check_metrics_after_test
    
    # Ожидание завершения
    log_info "Ожидание завершения тестирования (90 секунд)..."
    sleep 90
    
    # Финальная проверка
    check_metrics_after_test
    
    # Очистка
    cleanup_processes
    
    log_success "Тестирование TLS завершено!"
    log_info "Проверьте дашборды в Grafana: http://$(hostname -I | awk '{print $1}'):3000"
}

# Обработка Ctrl+C
trap cleanup_processes INT

# Запуск основной функции
main "$@" 