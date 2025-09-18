#!/bin/bash

# Скрипт для запуска AI-аналитика CloudBridge Relay
# Автор: AI Assistant
# Дата: $(date)

set -e

echo "🤖 Запуск AI-аналитика CloudBridge Relay..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
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

# Проверка переменных окружения
check_env() {
    log_info "Проверка переменных окружения..."
    
    if [ -z "$SLACK_WEBHOOK_URL" ]; then
        log_warning "SLACK_WEBHOOK_URL не установлен. Slack уведомления будут отключены."
    else
        log_success "SLACK_WEBHOOK_URL установлен"
    fi
    
    if [ -z "$AI_ANALYTICS_PORT" ]; then
        export AI_ANALYTICS_PORT=8084
        log_info "AI_ANALYTICS_PORT установлен в $AI_ANALYTICS_PORT"
    fi
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go не установлен. Установите Go 1.21 или выше."
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        log_warning "Docker не установлен. Локальный запуск будет невозможен."
    fi
    
    log_success "Зависимости проверены"
}

# Сборка приложения
build_app() {
    log_info "Сборка AI-аналитика..."
    
    cd "$(dirname "$0")/.."
    
    # Скачивание зависимостей
    log_info "Скачивание зависимостей..."
    go mod download
    
    # Сборка
    log_info "Компиляция..."
    CGO_ENABLED=0 go build -o bin/ai-analytics ./cmd/ai-analytics
    
    if [ $? -eq 0 ]; then
        log_success "AI-аналитик успешно собран"
    else
        log_error "Ошибка сборки AI-аналитика"
        exit 1
    fi
}

# Сборка Docker образа
build_docker() {
    if command -v docker &> /dev/null; then
        log_info "Сборка Docker образа..."
        
        docker build -f Dockerfile.ai-analytics -t cloudbridge/ai-analytics:latest .
        
        if [ $? -eq 0 ]; then
            log_success "Docker образ успешно собран"
        else
            log_error "Ошибка сборки Docker образа"
            exit 1
        fi
    else
        log_warning "Docker не доступен, пропускаем сборку образа"
    fi
}

# Запуск в Kubernetes
deploy_k8s() {
    if command -v kubectl &> /dev/null; then
        log_info "Развертывание в Kubernetes..."
        
        # Применение манифестов
        kubectl apply -f k8s/ai-analytics-deployment.yaml
        
        if [ $? -eq 0 ]; then
            log_success "AI-аналитик развернут в Kubernetes"
            
            # Ожидание готовности
            log_info "Ожидание готовности подов..."
            kubectl wait --for=condition=ready pod -l app=ai-analytics -n cloudbridge --timeout=300s
            
            # Получение статуса
            kubectl get pods -l app=ai-analytics -n cloudbridge
        else
            log_error "Ошибка развертывания в Kubernetes"
            exit 1
        fi
    else
        log_warning "kubectl не доступен, пропускаем развертывание в Kubernetes"
    fi
}

# Запуск локально
run_local() {
    log_info "Запуск AI-аналитика локально..."
    
    # Проверка конфигурации
    if [ ! -f "config/ai-analytics.yaml" ]; then
        log_error "Файл конфигурации config/ai-analytics.yaml не найден"
        exit 1
    fi
    
    # Запуск
    log_info "Запуск на порту $AI_ANALYTICS_PORT..."
    ./bin/ai-analytics
}

# Показать статус
show_status() {
    log_info "Статус AI-аналитика:"
    
    if command -v kubectl &> /dev/null; then
        echo "Kubernetes поды:"
        kubectl get pods -l app=ai-analytics -n cloudbridge 2>/dev/null || echo "Поды не найдены"
        
        echo "Сервисы:"
        kubectl get svc -l app=ai-analytics -n cloudbridge 2>/dev/null || echo "Сервисы не найдены"
    fi
    
    echo "Локальные процессы:"
    ps aux | grep ai-analytics | grep -v grep || echo "Локальные процессы не найдены"
}

# Показать логи
show_logs() {
    log_info "Логи AI-аналитика:"
    
    if command -v kubectl &> /dev/null; then
        kubectl logs -l app=ai-analytics -n cloudbridge --tail=50
    else
        log_warning "kubectl не доступен для просмотра логов"
    fi
}

# Показать помощь
show_help() {
    echo "Использование: $0 [команда]"
    echo ""
    echo "Команды:"
    echo "  build     - Собрать AI-аналитик"
    echo "  docker    - Собрать Docker образ"
    echo "  deploy    - Развернуть в Kubernetes"
    echo "  run       - Запустить локально"
    echo "  status    - Показать статус"
    echo "  logs      - Показать логи"
    echo "  help      - Показать эту справку"
    echo ""
    echo "Переменные окружения:"
    echo "  SLACK_WEBHOOK_URL    - URL webhook для Slack уведомлений"
    echo "  AI_ANALYTICS_PORT    - Порт для запуска (по умолчанию: 8084)"
    echo ""
    echo "Примеры:"
    echo "  $0 build"
    echo "  SLACK_WEBHOOK_URL=https://hooks.slack.com/... $0 run"
    echo "  $0 deploy"
}

# Основная логика
main() {
    case "${1:-help}" in
        "build")
            check_dependencies
            build_app
            ;;
        "docker")
            check_dependencies
            build_docker
            ;;
        "deploy")
            check_dependencies
            build_docker
            deploy_k8s
            ;;
        "run")
            check_env
            check_dependencies
            build_app
            run_local
            ;;
        "status")
            show_status
            ;;
        "logs")
            show_logs
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Запуск
main "$@"



