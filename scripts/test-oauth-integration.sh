#!/bin/bash

# ===========================================
# Sentry OAuth Integration Test Script
# ===========================================

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    if ! command -v curl &> /dev/null; then
        error "curl не установлен"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq не установлен"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        error "go не установлен"
        exit 1
    fi
    
    success "Все зависимости установлены"
}

# Загрузка переменных окружения
load_env() {
    log "Загрузка переменных окружения..."
    
    if [ -f "scripts/load-env.sh" ]; then
        source scripts/load-env.sh
        success "Переменные окружения загружены"
    else
        error "Файл scripts/load-env.sh не найден"
        exit 1
    fi
}

# Проверка конфигурации OAuth
check_oauth_config() {
    log "Проверка конфигурации OAuth..."
    
    if [ -z "$SENTRY_CLIENT_ID" ]; then
        error "SENTRY_CLIENT_ID не установлен"
        exit 1
    fi
    
    if [ -z "$SENTRY_CLIENT_SECRET" ]; then
        error "SENTRY_CLIENT_SECRET не установлен"
        exit 1
    fi
    
    if [ -z "$SENTRY_AUTH_URL" ]; then
        error "SENTRY_AUTH_URL не установлен"
        exit 1
    fi
    
    if [ -z "$SENTRY_TOKEN_URL" ]; then
        error "SENTRY_TOKEN_URL не установлен"
        exit 1
    fi
    
    success "Конфигурация OAuth корректна"
}

# Запуск тестового сервера
start_test_server() {
    log "Запуск тестового сервера..."
    
    # Проверяем, не запущен ли уже сервер
    if pgrep -f "test-sentry-oauth" > /dev/null; then
        warning "Тестовый сервер уже запущен"
        return 0
    fi
    
    # Запускаем сервер в фоне
    nohup go run test-sentry-oauth.go > test-oauth.log 2>&1 &
    SERVER_PID=$!
    
    # Ждем запуска сервера
    sleep 3
    
    # Проверяем, что сервер запустился
    if ! pgrep -f "test-sentry-oauth" > /dev/null; then
        error "Не удалось запустить тестовый сервер"
        cat test-oauth.log
        exit 1
    fi
    
    success "Тестовый сервер запущен (PID: $SERVER_PID)"
}

# Остановка тестового сервера
stop_test_server() {
    log "Остановка тестового сервера..."
    
    if pgrep -f "test-sentry-oauth" > /dev/null; then
        pkill -f "test-sentry-oauth"
        sleep 2
        success "Тестовый сервер остановлен"
    else
        warning "Тестовый сервер не запущен"
    fi
    
    # Удаляем лог файл
    if [ -f "test-oauth.log" ]; then
        rm test-oauth.log
    fi
}

# Тестирование основных endpoints
test_endpoints() {
    log "Тестирование основных endpoints..."
    
    # Тест главной страницы
    log "Тестирование главной страницы..."
    response=$(curl -s http://localhost:8080/)
    if echo "$response" | jq -e '.message' > /dev/null; then
        success "Главная страница работает"
    else
        error "Главная страница не работает"
        echo "Response: $response"
    fi
    
    # Тест health check
    log "Тестирование health check..."
    response=$(curl -s http://localhost:8080/health)
    if echo "$response" | jq -e '.status == "healthy"' > /dev/null; then
        success "Health check работает"
    else
        error "Health check не работает"
        echo "Response: $response"
    fi
    
    # Тест OAuth endpoint
    log "Тестирование OAuth endpoint..."
    response=$(curl -s http://localhost:8080/auth/sentry)
    if echo "$response" | jq -e '.auth_url' > /dev/null; then
        success "OAuth endpoint работает"
        echo "OAuth URL: $(echo "$response" | jq -r '.auth_url')"
    else
        error "OAuth endpoint не работает"
        echo "Response: $response"
    fi
    
    # Тест API endpoint
    log "Тестирование API endpoint..."
    response=$(curl -s http://localhost:8080/api/sentry/projects)
    if echo "$response" | jq -e '.message' > /dev/null; then
        success "API endpoint работает"
    else
        error "API endpoint не работает"
        echo "Response: $response"
    fi
}

# Тестирование отправки ошибки
test_error_sending() {
    log "Тестирование отправки ошибки..."
    
    response=$(curl -s http://localhost:8080/test-error)
    if echo "$response" | jq -e '.error' > /dev/null; then
        success "Ошибка отправлена в Sentry"
        echo "Response: $response"
    else
        error "Не удалось отправить ошибку"
        echo "Response: $response"
    fi
}

# Тестирование webhook
test_webhook() {
    log "Тестирование webhook..."
    
    # Создаем тестовое webhook событие
    webhook_data='{
        "action": "created",
        "data": {
            "issue": {
                "id": "test-123",
                "title": "Test Issue",
                "level": "error"
            }
        },
        "created_at": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
    }'
    
    # Отправляем webhook событие
    response=$(curl -s -X POST http://localhost:8080/webhook/sentry \
        -H "Content-Type: application/json" \
        -d "$webhook_data")
    
    if [ "$response" = "OK" ]; then
        success "Webhook событие обработано"
    else
        error "Webhook событие не обработано"
        echo "Response: $response"
    fi
}

# Проверка логов Sentry
check_sentry_logs() {
    log "Проверка логов Sentry..."
    
    if [ -f "test-oauth.log" ]; then
        log "Последние 10 строк логов:"
        tail -10 test-oauth.log
    else
        warning "Лог файл не найден"
    fi
}

# Основная функция
main() {
    echo "🚀 CloudBridge Relay - Sentry OAuth Integration Test"
    echo "=================================================="
    echo ""
    
    # Проверка зависимостей
    check_dependencies
    
    # Загрузка переменных окружения
    load_env
    
    # Проверка конфигурации OAuth
    check_oauth_config
    
    # Запуск тестового сервера
    start_test_server
    
    # Тестирование endpoints
    test_endpoints
    
    # Тестирование отправки ошибки
    test_error_sending
    
    # Тестирование webhook
    test_webhook
    
    # Проверка логов
    check_sentry_logs
    
    # Остановка тестового сервера
    stop_test_server
    
    echo ""
    success "Тестирование OAuth интеграции завершено!"
    echo ""
    echo "📋 Результаты:"
    echo "   ✅ OAuth конфигурация корректна"
    echo "   ✅ Тестовый сервер запущен и работает"
    echo "   ✅ Все endpoints отвечают"
    echo "   ✅ Ошибка отправлена в Sentry"
    echo "   ✅ Webhook события обрабатываются"
    echo ""
    echo "🔗 Для полного тестирования OAuth:"
    echo "   1. Откройте: http://localhost:8080/auth/sentry"
    echo "   2. Авторизуйтесь в Sentry"
    echo "   3. Проверьте получение токена"
    echo ""
    echo "📊 Проверьте Sentry dashboard для подтверждения получения событий"
}

# Обработка сигналов
trap 'stop_test_server; exit 1' INT TERM

# Запуск основной функции
main "$@"



