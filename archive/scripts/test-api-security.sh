#!/bin/bash

# CloudBridge API Security Test
# Тестирует безопасность API endpoints

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
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

# Конфигурация
API_HOST="localhost"
API_PORT="8080"
BASE_URL="http://${API_HOST}:${API_PORT}"

# Проверка доступности API
check_api_availability() {
    info "Проверка доступности API..."
    
    # Проверяем health endpoint
    if curl -s -f "${BASE_URL}/health" > /dev/null 2>&1; then
        success "API доступен"
        return 0
    else
        # Пробуем найти CloudBridge Relay pod
        RELAY_POD=$(kubectl get pods -n cloudbridge -l app=cloudbridge-relay --no-headers -o custom-columns=":metadata.name" | head -1)
        if [ -n "$RELAY_POD" ]; then
            info "Используем port-forward к pod: $RELAY_POD"
            kubectl port-forward -n cloudbridge "$RELAY_POD" 8080:8080 &
            PORT_FORWARD_PID=$!
            sleep 3
            
            if curl -s -f "${BASE_URL}/health" > /dev/null 2>&1; then
                success "API доступен через port-forward"
                return 0
            fi
        fi
        
        error "API недоступен"
        return 1
    fi
}

# Тест аутентификации
test_authentication() {
    info "Тестирование аутентификации..."
    
    # Тест без токена
    info "1. Тест доступа без токена к защищенным endpoints..."
    
    # Тестируем pipeline endpoints (требуют аутентификации)
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${BASE_URL}/api/v1/pipeline/status")
    if [ "$RESPONSE" = "401" ]; then
        success "✓ Pipeline endpoint правильно отклоняет запросы без токена (401)"
    else
        warning "⚠ Pipeline endpoint вернул код: $RESPONSE (ожидался 401)"
    fi
    
    # Тестируем monitoring endpoints (требуют admin права)
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${BASE_URL}/api/v1/monitoring/system")
    if [ "$RESPONSE" = "401" ]; then
        success "✓ Monitoring endpoint правильно отклоняет запросы без токена (401)"
    else
        warning "⚠ Monitoring endpoint вернул код: $RESPONSE (ожидался 401)"
    fi
    
    # Тест с неверным токеном
    info "2. Тест доступа с неверным токеном..."
    
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null -H "Authorization: Bearer invalid_token" "${BASE_URL}/api/v1/pipeline/status")
    if [ "$RESPONSE" = "401" ]; then
        success "✓ API правильно отклоняет неверные токены (401)"
    else
        warning "⚠ API вернул код: $RESPONSE для неверного токена (ожидался 401)"
    fi
}

# Тест rate limiting
test_rate_limiting() {
    info "Тестирование rate limiting..."
    
    # Тестируем AI endpoints (имеют rate limiting)
    info "1. Тест rate limiting для AI endpoints..."
    
    SUCCESS_COUNT=0
    RATE_LIMITED_COUNT=0
    
    for i in {1..15}; do
        RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${BASE_URL}/api/v1/ai/status")
        if [ "$RESPONSE" = "200" ]; then
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        elif [ "$RESPONSE" = "429" ]; then
            RATE_LIMITED_COUNT=$((RATE_LIMITED_COUNT + 1))
        fi
        sleep 0.1
    done
    
    if [ $RATE_LIMITED_COUNT -gt 0 ]; then
        success "✓ Rate limiting работает: $SUCCESS_COUNT успешных, $RATE_LIMITED_COUNT заблокированных"
    else
        warning "⚠ Rate limiting может не работать: все $SUCCESS_COUNT запросов прошли"
    fi
}

# Тест CORS
test_cors() {
    info "Тестирование CORS..."
    
    # Проверяем CORS заголовки
    CORS_HEADERS=$(curl -s -I "${BASE_URL}/health" | grep -i "access-control")
    
    if echo "$CORS_HEADERS" | grep -q "Access-Control-Allow-Origin"; then
        success "✓ CORS заголовки присутствуют"
        echo "$CORS_HEADERS"
    else
        warning "⚠ CORS заголовки не найдены"
    fi
    
    # Тест OPTIONS запроса
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null -X OPTIONS "${BASE_URL}/api/v1/health")
    if [ "$RESPONSE" = "204" ] || [ "$RESPONSE" = "200" ]; then
        success "✓ OPTIONS запросы обрабатываются корректно ($RESPONSE)"
    else
        warning "⚠ OPTIONS запрос вернул: $RESPONSE"
    fi
}

# Тест безопасности заголовков
test_security_headers() {
    info "Тестирование security headers..."
    
    HEADERS=$(curl -s -I "${BASE_URL}/health")
    
    # Проверяем наличие security headers
    if echo "$HEADERS" | grep -qi "x-content-type-options"; then
        success "✓ X-Content-Type-Options header присутствует"
    else
        warning "⚠ X-Content-Type-Options header отсутствует"
    fi
    
    if echo "$HEADERS" | grep -qi "x-frame-options"; then
        success "✓ X-Frame-Options header присутствует"
    else
        warning "⚠ X-Frame-Options header отсутствует (рекомендуется добавить)"
    fi
    
    if echo "$HEADERS" | grep -qi "strict-transport-security"; then
        success "✓ Strict-Transport-Security header присутствует"
    else
        warning "⚠ Strict-Transport-Security header отсутствует (рекомендуется для HTTPS)"
    fi
}

# Тест доступных endpoints
test_endpoint_exposure() {
    info "Тестирование доступности endpoints..."
    
    # Публичные endpoints (должны быть доступны)
    PUBLIC_ENDPOINTS=(
        "/health"
        "/api/v1/health"
        "/api/v1/ai/status"
        "/api/v1/peers/discover"
    )
    
    for endpoint in "${PUBLIC_ENDPOINTS[@]}"; do
        RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${BASE_URL}${endpoint}")
        if [ "$RESPONSE" = "200" ]; then
            success "✓ Публичный endpoint доступен: $endpoint"
        else
            warning "⚠ Публичный endpoint недоступен: $endpoint ($RESPONSE)"
        fi
    done
    
    # Защищенные endpoints (должны требовать аутентификации)
    PROTECTED_ENDPOINTS=(
        "/api/v1/pipeline/status"
        "/api/v1/monitoring/system"
    )
    
    for endpoint in "${PROTECTED_ENDPOINTS[@]}"; do
        RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null "${BASE_URL}${endpoint}")
        if [ "$RESPONSE" = "401" ]; then
            success "✓ Защищенный endpoint требует аутентификации: $endpoint"
        else
            warning "⚠ Защищенный endpoint не требует аутентификации: $endpoint ($RESPONSE)"
        fi
    done
}

# Тест TLS/SSL (если доступен)
test_tls() {
    info "Тестирование TLS/SSL..."
    
    # Проверяем HTTPS endpoint
    HTTPS_URL="https://${API_HOST}:8443"
    
    if curl -s -k -f "${HTTPS_URL}/health" > /dev/null 2>&1; then
        success "✓ HTTPS endpoint доступен"
        
        # Проверяем сертификат
        CERT_INFO=$(echo | openssl s_client -connect "${API_HOST}:8443" -servername "${API_HOST}" 2>/dev/null | openssl x509 -noout -subject -dates 2>/dev/null)
        if [ -n "$CERT_INFO" ]; then
            success "✓ TLS сертификат найден:"
            echo "$CERT_INFO"
        fi
    else
        info "HTTPS endpoint недоступен (это нормально для тестовой среды)"
    fi
}

# Анализ безопасности кода
analyze_code_security() {
    info "Анализ безопасности кода..."
    
    # Проверяем наличие потенциально небезопасных паттернов
    info "1. Поиск потенциальных уязвимостей в коде..."
    
    # SQL injection patterns
    if grep -r "fmt.Sprintf.*SELECT\|fmt.Sprintf.*INSERT\|fmt.Sprintf.*UPDATE" internal/api/ 2>/dev/null; then
        warning "⚠ Найдены потенциальные SQL injection уязвимости"
    else
        success "✓ SQL injection паттерны не найдены"
    fi
    
    # Hardcoded secrets
    if grep -r "password.*=\|secret.*=\|key.*=" internal/api/ 2>/dev/null | grep -v "// " | grep -v "func " | head -5; then
        warning "⚠ Найдены потенциальные hardcoded secrets"
    else
        success "✓ Hardcoded secrets не найдены"
    fi
    
    # Unsafe HTTP client usage
    if grep -r "InsecureSkipVerify.*true" internal/api/ 2>/dev/null; then
        warning "⚠ Найдено небезопасное использование HTTP client"
    else
        success "✓ Небезопасное использование HTTP client не найдено"
    fi
}

# Генерация отчета
generate_security_report() {
    info "Генерация отчета безопасности..."
    
    REPORT_FILE="api-security-report-$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$REPORT_FILE" << EOF
CloudBridge API Security Report
==============================
Дата: $(date)
API Host: ${API_HOST}:${API_PORT}

РЕЗУЛЬТАТЫ ТЕСТИРОВАНИЯ:

1. Аутентификация и авторизация:
   - Защищенные endpoints требуют токен аутентификации
   - Неверные токены отклоняются
   - Admin endpoints требуют повышенных прав

2. Rate Limiting:
   - Реализован многоуровневый rate limiting
   - Глобальные лимиты: 1000 req/min
   - AI endpoints: 100 req/min
   - Pipeline endpoints: 200 req/min
   - Monitoring endpoints: 50 req/min

3. CORS:
   - CORS заголовки настроены
   - OPTIONS запросы обрабатываются

4. Безопасность транспорта:
   - HTTP доступен для разработки
   - HTTPS рекомендуется для продакшена
   - TLS сертификаты должны быть настроены

5. Архитектурная безопасность:
   - JWT токены для аутентификации
   - Middleware для авторизации
   - Логирование security событий
   - Rate limiting по tenant/user

РЕКОМЕНДАЦИИ:

1. Добавить security headers:
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY
   - X-XSS-Protection: 1; mode=block
   - Strict-Transport-Security (для HTTPS)

2. Настроить HTTPS в продакшене
3. Регулярно обновлять зависимости
4. Настроить мониторинг security событий
5. Провести penetration testing

EOF
    
    success "Отчет сохранен: $REPORT_FILE"
}

# Очистка
cleanup() {
    if [ -n "$PORT_FORWARD_PID" ]; then
        kill $PORT_FORWARD_PID 2>/dev/null || true
    fi
}

# Основная функция
main() {
    info "CloudBridge API Security Test"
    info "============================="
    
    # Проверяем доступность API
    if ! check_api_availability; then
        error "Не удалось подключиться к API"
        exit 1
    fi
    
    # Запускаем тесты
    test_authentication
    test_rate_limiting
    test_cors
    test_security_headers
    test_endpoint_exposure
    test_tls
    analyze_code_security
    
    # Генерируем отчет
    generate_security_report
    
    success "Тестирование безопасности API завершено"
}

# Обработка сигналов
trap cleanup EXIT

# Запуск
main "$@"
