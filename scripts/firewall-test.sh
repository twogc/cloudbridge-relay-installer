#!/bin/bash

# 🧪 CloudBridge Relay Firewall Test Script
# Тестирование доступности сервисов после настройки firewall

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Переменные
TARGET_HOST="relaymsk1.2gc.uk"
TIMEOUT=10

# Функция логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Функция тестирования TCP порта
test_tcp_port() {
    local port=$1
    local service=$2
    
    log "Тестируем $service (TCP:$port)..."
    
    if timeout $TIMEOUT bash -c "</dev/tcp/$TARGET_HOST/$port" 2>/dev/null; then
        success "$service (TCP:$port) - ДОСТУПЕН"
        return 0
    else
        error "$service (TCP:$port) - НЕДОСТУПЕН"
        return 1
    fi
}

# Функция тестирования UDP порта
test_udp_port() {
    local port=$1
    local service=$2
    
    log "Тестируем $service (UDP:$port)..."
    
    if timeout $TIMEOUT bash -c "</dev/udp/$TARGET_HOST/$port" 2>/dev/null; then
        success "$service (UDP:$port) - ДОСТУПЕН"
        return 0
    else
        warning "$service (UDP:$port) - Проверьте вручную (UDP сложно тестировать)"
        return 1
    fi
}

# Функция тестирования HTTP endpoint
test_http_endpoint() {
    local port=$1
    local endpoint=$2
    local service=$3
    
    log "Тестируем $service (HTTP:$port$endpoint)..."
    
    response=$(timeout $TIMEOUT curl -s -o /dev/null -w "%{http_code}" "http://$TARGET_HOST:$port$endpoint" 2>/dev/null || echo "000")
    
    if [ "$response" = "200" ] || [ "$response" = "302" ] || [ "$response" = "404" ]; then
        success "$service (HTTP:$port$endpoint) - ДОСТУПЕН (HTTP $response)"
        return 0
    else
        error "$service (HTTP:$port$endpoint) - НЕДОСТУПЕН (HTTP $response)"
        return 1
    fi
}

# Функция тестирования с помощью netcat
test_with_nc() {
    local port=$1
    local protocol=$2
    local service=$3
    
    log "Тестируем $service ($protocol:$port) с помощью netcat..."
    
    if command -v nc &> /dev/null; then
        if [ "$protocol" = "tcp" ]; then
            if timeout $TIMEOUT nc -z "$TARGET_HOST" "$port" 2>/dev/null; then
                success "$service ($protocol:$port) - ДОСТУПЕН (netcat)"
                return 0
            else
                error "$service ($protocol:$port) - НЕДОСТУПЕН (netcat)"
                return 1
            fi
        elif [ "$protocol" = "udp" ]; then
            if timeout $TIMEOUT nc -zu "$TARGET_HOST" "$port" 2>/dev/null; then
                success "$service ($protocol:$port) - ДОСТУПЕН (netcat)"
                return 0
            else
                warning "$service ($protocol:$port) - Проверьте вручную (UDP)"
                return 1
            fi
        fi
    else
        warning "netcat не установлен, пропускаем тест"
        return 1
    fi
}

# Функция тестирования DNS
test_dns() {
    log "Тестируем DNS разрешение..."
    
    if nslookup "$TARGET_HOST" >/dev/null 2>&1; then
        success "DNS разрешение работает"
        return 0
    else
        error "DNS разрешение не работает"
        return 1
    fi
}

# Функция тестирования ping
test_ping() {
    log "Тестируем доступность хоста (ping)..."
    
    if ping -c 3 "$TARGET_HOST" >/dev/null 2>&1; then
        success "Хост доступен (ping)"
        return 0
    else
        warning "Хост недоступен (ping) - возможно, ICMP заблокирован"
        return 1
    fi
}

# Функция тестирования traceroute
test_traceroute() {
    log "Тестируем маршрут до хоста..."
    
    if command -v traceroute &> /dev/null; then
        echo "Маршрут до $TARGET_HOST:"
        traceroute -m 15 "$TARGET_HOST" 2>/dev/null || echo "traceroute недоступен"
    else
        warning "traceroute не установлен"
    fi
}

# Основная функция тестирования
run_tests() {
    echo "🧪 Начинаем тестирование firewall для CloudBridge Relay"
    echo "🎯 Целевой хост: $TARGET_HOST"
    echo "⏱️  Таймаут: ${TIMEOUT}с"
    echo ""
    
    local total_tests=0
    local passed_tests=0
    
    # Базовые тесты
    test_ping && ((passed_tests++)) || true
    ((total_tests++))
    
    test_dns && ((passed_tests++)) || true
    ((total_tests++))
    
    # Критические порты
    test_tcp_port 8080 "Relay API" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_tcp_port 8081 "Relay Metrics" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_udp_port 51820 "WireGuard" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_udp_port 9090 "QUIC" && ((passed_tests++)) || true
    ((total_tests++))
    
    # HTTP endpoints
    test_http_endpoint 8080 "/health" "Relay API Health" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_http_endpoint 8081 "/metrics" "Relay Metrics" && ((passed_tests++)) || true
    ((total_tests++))
    
    # Дополнительные порты
    test_tcp_port 80 "HTTP" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_tcp_port 443 "HTTPS" && ((passed_tests++)) || true
    ((total_tests++))
    
    # Тесты с netcat
    test_with_nc 8080 "tcp" "Relay API (nc)" && ((passed_tests++)) || true
    ((total_tests++))
    
    test_with_nc 51820 "udp" "WireGuard (nc)" && ((passed_tests++)) || true
    ((total_tests++))
    
    echo ""
    echo "📊 Результаты тестирования:"
    echo "  ✅ Пройдено: $passed_tests/$total_tests"
    echo "  ❌ Провалено: $((total_tests - passed_tests))/$total_tests"
    
    if [ $passed_tests -eq $total_tests ]; then
        success "🎉 Все тесты пройдены! Firewall настроен корректно."
        return 0
    else
        warning "⚠️  Некоторые тесты провалены. Проверьте настройки firewall."
        return 1
    fi
}

# Функция для детального тестирования
detailed_test() {
    echo "🔍 Детальное тестирование..."
    echo ""
    
    test_traceroute
    
    echo ""
    echo "📋 Проверка конкретных сервисов:"
    
    # Тест Relay API
    echo "1. Relay API (порт 8080):"
    curl -I "http://$TARGET_HOST:8080/health" 2>/dev/null || echo "  Недоступен"
    
    # Тест Metrics
    echo "2. Relay Metrics (порт 8081):"
    curl -I "http://$TARGET_HOST:8081/metrics" 2>/dev/null || echo "  Недоступен"
    
    # Тест WireGuard
    echo "3. WireGuard (порт 51820):"
    if command -v nc &> /dev/null; then
        nc -zu "$TARGET_HOST" 51820 2>/dev/null && echo "  Доступен" || echo "  Недоступен"
    else
        echo "  netcat не установлен"
    fi
}

# Функция для проверки извне
external_test() {
    echo "🌐 Тестирование из внешней сети..."
    
    # Используем внешний сервис для проверки
    if command -v curl &> /dev/null; then
        echo "Проверяем доступность через внешний сервис..."
        
        # Можно добавить проверку через внешние сервисы
        # Например, через https://portchecker.co/
        echo "Для полной проверки используйте внешние сервисы:"
        echo "  - https://portchecker.co/"
        echo "  - https://www.yougetsignal.com/tools/open-ports/"
        echo "  - https://canyouseeme.org/"
    fi
}

# Главная функция
main() {
    case "${1:-run}" in
        "run")
            run_tests
            ;;
        "detailed")
            run_tests
            detailed_test
            ;;
        "external")
            external_test
            ;;
        "help"|"-h"|"--help")
            echo "Использование: $0 [команда]"
            echo ""
            echo "Команды:"
            echo "  run      - Базовое тестирование (по умолчанию)"
            echo "  detailed - Детальное тестирование"
            echo "  external - Тестирование из внешней сети"
            echo "  help     - Показать эту справку"
            echo ""
            echo "Переменные окружения:"
            echo "  TARGET_HOST - Целевой хост (по умолчанию: relaymsk1.2gc.uk)"
            echo "  TIMEOUT     - Таймаут в секундах (по умолчанию: 10)"
            ;;
        *)
            error "Неизвестная команда: $1"
            echo "Используйте '$0 help' для справки"
            exit 1
            ;;
    esac
}

# Запуск скрипта
main "$@"
