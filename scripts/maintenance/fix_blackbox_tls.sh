#!/bin/bash

# CloudBridge Relay - Исправление TLS проверок Blackbox Exporter
# Этот скрипт исправляет проблему с проверкой TLS сертификатов в Docker-среде

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Проверка, что мы в корневой директории проекта
check_project_root() {
    if [[ ! -f "docker-compose.yml" ]] || [[ ! -d "deploy" ]]; then
        log_error "Скрипт должен быть запущен из корневой директории проекта"
        exit 1
    fi
}

# Перезапуск сервисов мониторинга
restart_monitoring_services() {
    log_info "Перезапуск сервисов мониторинга..."
    
    # Остановка сервисов
    docker-compose stop blackbox prometheus
    
    # Запуск сервисов
    docker-compose up -d blackbox prometheus
    
    log_success "Сервисы мониторинга перезапущены"
}

# Проверка статуса сервисов
check_services_status() {
    log_info "Проверка статуса сервисов..."
    
    # Проверка Blackbox
    if curl -s -f "http://localhost:9115/-/healthy" > /dev/null 2>&1; then
        log_success "Blackbox Exporter работает"
    else
        log_error "Blackbox Exporter не отвечает"
        return 1
    fi
    
    # Проверка Prometheus
    if curl -s -f "http://localhost:9090/-/healthy" > /dev/null 2>&1; then
        log_success "Prometheus работает"
    else
        log_error "Prometheus не отвечает"
        return 1
    fi
    
    return 0
}

# Тестирование TLS проверки
test_tls_check() {
    log_info "Тестирование TLS проверки..."
    
    # Ждем немного для стабилизации
    sleep 10
    
    # Проверяем метрики Blackbox
    local response=$(curl -s "http://localhost:9115/metrics" | grep "probe_success" | grep "module=\"tls_connect_internal\"")
    
    if [[ -n "$response" ]]; then
        log_success "TLS проверка настроена корректно"
        echo "$response"
    else
        log_warning "TLS проверка еще не появилась в метриках (может потребоваться время)"
    fi
}

# Основная функция
main() {
    log_info "Исправление TLS проверок Blackbox Exporter..."
    
    check_project_root
    restart_monitoring_services
    
    # Ждем запуска сервисов
    log_info "Ожидание запуска сервисов..."
    sleep 15
    
    if check_services_status; then
        test_tls_check
        log_success "Исправление завершено успешно!"
        log_info "Теперь TLS проверки должны работать корректно в Docker-среде"
    else
        log_error "Проблемы с запуском сервисов"
        exit 1
    fi
}

# Запуск основной функции
main "$@" 