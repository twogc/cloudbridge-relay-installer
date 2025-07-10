#!/bin/bash

# CloudBridge AI Monitoring Agent - Deployment Script
# Развертывание AI агента мониторинга на Cloudflare Workers

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log_step "Проверка зависимостей..."
    
    if ! command -v npm &> /dev/null; then
        log_error "npm не установлен"
        exit 1
    fi
    
    if ! command -v wrangler &> /dev/null; then
        log_warn "wrangler не установлен. Устанавливаем..."
        npm install -g wrangler
    fi
    
    log_info "Зависимости проверены"
}

# Установка зависимостей проекта
install_dependencies() {
    log_step "Установка зависимостей проекта..."
    
    if [ -f "package.json" ]; then
        npm install
        log_info "Зависимости установлены"
    else
        log_error "package.json не найден"
        exit 1
    fi
}

# Настройка переменных окружения
setup_environment() {
    log_step "Настройка переменных окружения..."
    
    # Создание .dev.vars для локальной разработки
    if [ ! -f ".dev.vars" ]; then
        cat > .dev.vars << EOF
# CloudBridge AI Agent Environment Variables
OPENAI_API_KEY=your-openai-api-key-here
RELAY_METRICS_URL=http://localhost:9090/metrics
ML_API_URL=http://localhost:8000/detect
SIEM_API_URL=http://localhost:8081/api/v1/alerts
SIEM_WEBHOOK_URL=http://localhost:8081/api/v1/webhook
CHECK_INTERVAL=30000
ANOMALY_THRESHOLD=0.7
ML_API_KEY=your-ml-api-key-here
SIEM_API_KEY=your-siem-api-key-here
EOF
        log_info "Создан файл .dev.vars"
    else
        log_info "Файл .dev.vars уже существует"
    fi
    
    # Проверка наличия обязательных переменных
    if [ ! -f ".dev.vars" ]; then
        log_error "Файл .dev.vars не найден"
        exit 1
    fi
    
    log_info "Переменные окружения настроены"
}

# Настройка секретов для production
setup_secrets() {
    log_step "Настройка секретов..."
    
    echo "Для production развертывания необходимо настроить секреты:"
    echo "1. OPENAI_API_KEY"
    echo "2. ML_API_KEY" 
    echo "3. SIEM_API_KEY"
    echo
    echo "Команды для настройки секретов:"
    echo "wrangler secret put OPENAI_API_KEY"
    echo "wrangler secret put ML_API_KEY"
    echo "wrangler secret put SIEM_API_KEY"
    echo
    
    read -p "Хотите настроить секреты сейчас? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Введите значения секретов:"
        
        read -s -p "OPENAI_API_KEY: " openai_key
        echo
        if [ ! -z "$openai_key" ]; then
            echo "$openai_key" | wrangler secret put OPENAI_API_KEY
        fi
        
        read -s -p "ML_API_KEY: " ml_key
        echo
        if [ ! -z "$ml_key" ]; then
            echo "$ml_key" | wrangler secret put ML_API_KEY
        fi
        
        read -s -p "SIEM_API_KEY: " siem_key
        echo
        if [ ! -z "$siem_key" ]; then
            echo "$siem_key" | wrangler secret put SIEM_API_KEY
        fi
        
        log_info "Секреты настроены"
    fi
}

# Локальная разработка
dev_mode() {
    log_step "Запуск в режиме разработки..."
    
    if [ ! -f ".dev.vars" ]; then
        log_error "Файл .dev.vars не найден. Запустите setup_environment"
        exit 1
    fi
    
    log_info "Запуск wrangler dev..."
    wrangler dev
}

# Развертывание в staging
deploy_staging() {
    log_step "Развертывание в staging..."
    
    # Проверка конфигурации
    if ! wrangler config --env staging &> /dev/null; then
        log_error "Конфигурация staging не найдена в wrangler.toml"
        exit 1
    fi
    
    # Развертывание
    wrangler deploy --env staging
    
    log_info "Агент развернут в staging"
    log_info "URL: https://cloudbridge-ai-monitoring-agent-staging.your-subdomain.workers.dev"
}

# Развертывание в production
deploy_production() {
    log_step "Развертывание в production..."
    
    # Проверка конфигурации
    if ! wrangler config --env production &> /dev/null; then
        log_error "Конфигурация production не найдена в wrangler.toml"
        exit 1
    fi
    
    # Подтверждение развертывания
    echo "⚠️  Внимание: Вы собираетесь развернуть в PRODUCTION"
    read -p "Продолжить? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Развертывание отменено"
        exit 0
    fi
    
    # Развертывание
    wrangler deploy --env production
    
    log_info "Агент развернут в production"
    log_info "URL: https://cloudbridge-ai-monitoring-agent-prod.your-subdomain.workers.dev"
}

# Тестирование развертывания
test_deployment() {
    log_step "Тестирование развертывания..."
    
    local env=${1:-staging}
    local base_url=""
    
    if [ "$env" = "production" ]; then
        base_url="https://cloudbridge-ai-monitoring-agent-prod.your-subdomain.workers.dev"
    else
        base_url="https://cloudbridge-ai-monitoring-agent-staging.your-subdomain.workers.dev"
    fi
    
    echo "Тестирование endpoints:"
    
    # Health check
    if curl -s "$base_url/health" | grep -q "ok"; then
        log_info "✅ Health check прошел"
    else
        log_error "❌ Health check не прошел"
    fi
    
    # Metrics endpoint
    if curl -s "$base_url/metrics" > /dev/null; then
        log_info "✅ Metrics endpoint доступен"
    else
        log_error "❌ Metrics endpoint недоступен"
    fi
    
    # Check endpoint
    if curl -s -X POST "$base_url/check" > /dev/null; then
        log_info "✅ Check endpoint доступен"
    else
        log_error "❌ Check endpoint недоступен"
    fi
    
    log_info "Тестирование завершено"
}

# Просмотр логов
view_logs() {
    log_step "Просмотр логов..."
    
    local env=${1:-staging}
    
    echo "Логи для окружения: $env"
    wrangler tail --env "$env"
}

# Удаление развертывания
cleanup() {
    log_step "Удаление развертывания..."
    
    local env=${1:-staging}
    
    echo "Удаление агента из окружения: $env"
    wrangler delete --env "$env"
    
    log_info "Развертывание удалено"
}

# Показать справку
show_help() {
    echo "CloudBridge AI Monitoring Agent - Deployment Script"
    echo
    echo "Использование: $0 [команда] [опции]"
    echo
    echo "Команды:"
    echo "  setup           Настройка проекта (зависимости, переменные)"
    echo "  dev             Запуск в режиме разработки"
    echo "  deploy:staging  Развертывание в staging"
    echo "  deploy:prod     Развертывание в production"
    echo "  test [env]      Тестирование развертывания (staging/production)"
    echo "  logs [env]      Просмотр логов (staging/production)"
    echo "  cleanup [env]   Удаление развертывания (staging/production)"
    echo "  secrets         Настройка секретов"
    echo "  help            Показать эту справку"
    echo
    echo "Примеры:"
    echo "  $0 setup"
    echo "  $0 dev"
    echo "  $0 deploy:staging"
    echo "  $0 test production"
    echo "  $0 logs staging"
    echo
}

# Основная логика
main() {
    case "${1:-}" in
        setup)
            check_dependencies
            install_dependencies
            setup_environment
            setup_secrets
            ;;
        dev)
            dev_mode
            ;;
        deploy:staging)
            deploy_staging
            ;;
        deploy:prod)
            deploy_production
            ;;
        test)
            test_deployment "${2:-staging}"
            ;;
        logs)
            view_logs "${2:-staging}"
            ;;
        cleanup)
            cleanup "${2:-staging}"
            ;;
        secrets)
            setup_secrets
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# Запуск основной функции
main "$@" 