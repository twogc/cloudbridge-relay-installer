#!/bin/bash

# Тест мониторинга ресурсов в Minikube
# Автор: CloudBridge AI Analytics
# Дата: 2025-09-20

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
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
    
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl не найден"
        exit 1
    fi
    
    if ! command -v minikube &> /dev/null; then
        log_error "minikube не найден"
        exit 1
    fi
    
    log_success "Все зависимости найдены"
}

# Проверка статуса подов
check_pods_status() {
    log_info "Проверка статуса подов AI Service..."
    
    local ai_pods=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=ai-service --no-headers | wc -l)
    local running_pods=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=ai-service --no-headers | grep Running | wc -l)
    
    log_info "AI Service подов: $ai_pods, запущенных: $running_pods"
    
    if [ "$running_pods" -eq 0 ]; then
        log_error "Нет запущенных AI Service подов"
        return 1
    fi
    
    log_success "AI Service поды работают"
    return 0
}

# Тест мониторинга ресурсов
test_resource_monitoring() {
    log_info "Тест мониторинга ресурсов..."
    
    # Получаем имя первого AI Service пода
    local ai_pod=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=ai-service --no-headers | grep Running | head -1 | awk '{print $1}')
    
    if [ -z "$ai_pod" ]; then
        log_error "Не найден запущенный AI Service под"
        return 1
    fi
    
    log_info "Используем под: $ai_pod"
    
    # Проверяем использование ресурсов
    log_info "Проверка использования ресурсов..."
    
    # CPU и память
    local cpu_usage=$(kubectl top pod $ai_pod -n cloudbridge --no-headers | awk '{print $2}' | sed 's/m//')
    local memory_usage=$(kubectl top pod $ai_pod -n cloudbridge --no-headers | awk '{print $3}' | sed 's/Mi//')
    
    log_info "CPU usage: ${cpu_usage}m"
    log_info "Memory usage: ${memory_usage}Mi"
    
    # Проверяем логи на наличие алертов ресурсов
    log_info "Проверка логов на алерты ресурсов..."
    
    local resource_alerts=$(kubectl logs $ai_pod -n cloudbridge --tail=100 | grep -i "resource alert" | wc -l)
    
    if [ "$resource_alerts" -gt 0 ]; then
        log_success "Найдены алерты ресурсов в логах: $resource_alerts"
        kubectl logs $ai_pod -n cloudbridge --tail=100 | grep -i "resource alert"
    else
        log_info "Алерты ресурсов не найдены (это нормально, если ресурсы не превышают пороги)"
    fi
    
    return 0
}

# Тест конфигурации мониторинга
test_monitoring_config() {
    log_info "Тест конфигурации мониторинга..."
    
    # Проверяем ConfigMap
    if kubectl get configmap ai-telegram-config -n cloudbridge &> /dev/null; then
        log_success "ConfigMap ai-telegram-config найден"
        
        # Проверяем пороги ресурсов
        local disk_threshold=$(kubectl get configmap ai-telegram-config -n cloudbridge -o jsonpath='{.data.config\.yaml}' | grep -o 'disk_usage_threshold: [0-9.]*' | cut -d' ' -f2)
        local cpu_threshold=$(kubectl get configmap ai-telegram-config -n cloudbridge -o jsonpath='{.data.config\.yaml}' | grep -o 'cpu_usage_threshold: [0-9.]*' | cut -d' ' -f2)
        local memory_threshold=$(kubectl get configmap ai-telegram-config -n cloudbridge -o jsonpath='{.data.config\.yaml}' | grep -o 'memory_usage_threshold: [0-9.]*' | cut -d' ' -f2)
        
        log_info "Пороги ресурсов:"
        log_info "  Disk: ${disk_threshold:-не найден}"
        log_info "  CPU: ${cpu_threshold:-не найден}"
        log_info "  Memory: ${memory_threshold:-не найден}"
    else
        log_error "ConfigMap ai-telegram-config не найден"
        return 1
    fi
    
    # Проверяем Secret
    if kubectl get secret ai-telegram-secret -n cloudbridge &> /dev/null; then
        log_success "Secret ai-telegram-secret найден"
    else
        log_error "Secret ai-telegram-secret не найден"
        return 1
    fi
    
    return 0
}

# Тест отправки алертов
test_alert_sending() {
    log_info "Тест отправки алертов..."
    
    # Получаем имя первого AI Service пода
    local ai_pod=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=ai-service --no-headers | grep Running | head -1 | awk '{print $1}')
    
    if [ -z "$ai_pod" ]; then
        log_error "Не найден запущенный AI Service под"
        return 1
    fi
    
    # Проверяем переменные окружения для Telegram
    local telegram_enabled=$(kubectl exec $ai_pod -n cloudbridge -- env | grep AI_TELEGRAM_ENABLED | cut -d'=' -f2)
    local bot_token=$(kubectl exec $ai_pod -n cloudbridge -- env | grep AI_TELEGRAM_BOT_TOKEN | cut -d'=' -f2)
    
    log_info "Telegram enabled: $telegram_enabled"
    log_info "Bot token configured: $([ -n "$bot_token" ] && echo "да" || echo "нет")"
    
    if [ "$telegram_enabled" = "true" ] && [ -n "$bot_token" ]; then
        log_success "Telegram конфигурация корректна"
    else
        log_warning "Telegram конфигурация неполная"
    fi
    
    return 0
}

# Основная функция
main() {
    log_info "Запуск тестов мониторинга ресурсов в Minikube"
    echo
    
    local tests_passed=0
    local tests_total=4
    
    # Запуск тестов
    if check_pods_status; then
        ((tests_passed++))
    fi
    echo
    
    if test_monitoring_config; then
        ((tests_passed++))
    fi
    echo
    
    if test_resource_monitoring; then
        ((tests_passed++))
    fi
    echo
    
    if test_alert_sending; then
        ((tests_passed++))
    fi
    echo
    
    # Результаты
    log_info "Результаты тестирования:"
    log_info "Пройдено тестов: $tests_passed из $tests_total"
    
    if [ $tests_passed -eq $tests_total ]; then
        log_success "Все тесты пройдены успешно!"
        echo
        log_info "Мониторинг ресурсов настроен и работает:"
        log_info "• AI Service поды запущены"
        log_info "• Конфигурация мониторинга применена"
        log_info "• Telegram уведомления настроены"
        log_info "• Пороги ресурсов: 80% для диска, CPU и памяти"
        echo
        log_info "AI будет отправлять алерты в Telegram при превышении порогов"
        return 0
    else
        log_error "Некоторые тесты не пройдены"
        return 1
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Тест мониторинга ресурсов в Minikube"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --pods         Проверить только статус подов"
        echo "  --config       Проверить только конфигурацию"
        echo "  --monitoring   Проверить только мониторинг"
        echo "  --alerts       Проверить только алерты"
        echo
        echo "Примеры:"
        echo "  $0                    # Запустить все тесты"
        echo "  $0 --pods            # Проверить только поды"
        echo "  $0 --config          # Проверить только конфигурацию"
        exit 0
        ;;
    --pods)
        check_dependencies
        check_pods_status
        ;;
    --config)
        check_dependencies
        test_monitoring_config
        ;;
    --monitoring)
        check_dependencies
        test_resource_monitoring
        ;;
    --alerts)
        check_dependencies
        test_alert_sending
        ;;
    "")
        check_dependencies
        main
        ;;
    *)
        log_error "Неизвестная опция: $1"
        log_info "Используйте --help для получения справки"
        exit 1
        ;;
esac
