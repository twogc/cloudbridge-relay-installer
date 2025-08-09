#!/bin/bash

# CloudBridge Relay - Quick Lynis Test
# Быстрый тест Lynis для проверки функциональности

set -e

# Конфигурация Telegram
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"

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

# Проверка конфигурации Telegram
check_telegram_config() {
    if [[ -z "$TELEGRAM_BOT_TOKEN" || -z "$TELEGRAM_CHAT_ID" ]]; then
        log_warn "Telegram конфигурация не настроена"
        return 1
    fi
    return 0
}

# Отправка сообщения в Telegram
send_telegram_message() {
    local message="$1"
    local telegram_log="logs/telegram-error.log"
    
    if ! check_telegram_config; then
        log_warn "Пропускаю отправку в Telegram"
        return
    fi
    
    log_info "Отправка уведомления в Telegram..."
    
    # Экранируем только амперсанд для HTML-форматирования
    local escaped_message=$(echo "$message" | sed 's/&/\&amp;/g')
    
    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${escaped_message}" \
        -d "parse_mode=HTML" \
        2>&1)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] $response" >> "$telegram_log"
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] MESSAGE: $message" >> "$telegram_log"
    fi
}

# Быстрый тест Lynis
quick_lynis_test() {
    log_info "Запуск быстрого теста Lynis"
    
    # Создание директории для логов
    mkdir -p logs
    
    # Генерация имени файла лога
    local log_file="logs/lynis-quick-$(date +%F-%H%M).log"
    
    log_info "Лог файл: $log_file"
    
    # Отправка уведомления о начале
    send_telegram_message "Быстрый тест Lynis начат
Время начала: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)"
    
    # Запуск Lynis в быстром режиме с таймаутом 5 минут
    local start_time=$(date +%s)
    local lynis_exit_code=0
    
    log_info "Запуск Lynis в быстром режиме..."
    
    if timeout 300 lynis audit system --quick > "$log_file" 2>&1; then
        lynis_exit_code=$?
    else
        lynis_exit_code=$?
        if [[ $lynis_exit_code -eq 124 ]]; then
            log_error "Lynis превысил таймаут (5 минут)"
            send_telegram_message "Lynis превысил таймаут
Время: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)
Таймаут: 5 минут"
            return 1
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_info "Тест завершен за ${duration} секунд (exit code: $lynis_exit_code)"
    
    # Проверяем успешность выполнения
    if [[ $lynis_exit_code -eq 0 ]]; then
        # Извлекаем основные метрики
        local warnings=$(grep -c "Warning:" "$log_file" || echo "0")
        local suggestions=$(grep -c "Suggestion:" "$log_file" || echo "0")
        local security_issues=$(grep -c "Security issue:" "$log_file" || echo "0")
        local score=$(grep "Hardening index" "$log_file" | tail -1 | grep -o '[0-9]\+' | head -1 || echo "0")
        
        # Проверка WireGuard порта
        local wireguard_port_status=""
        if sudo ufw status | grep -q "51820/udp.*ALLOW"; then
            wireguard_port_status="✅ WireGuard порт 51820/udp открыт"
        else
            wireguard_port_status="❌ WireGuard порт 51820/udp НЕ открыт"
        fi
        
        # Формируем сообщение
        local message="Быстрый тест Lynis завершен

Результаты:
• Hardening Index: ${score}/100
• Предупреждений: ${warnings}
• Рекомендаций: ${suggestions}
• Проблем безопасности: ${security_issues}
• $wireguard_port_status

Лог файл: $log_file
Время выполнения: ${duration} секунд
Время завершения: $(date '+%Y-%m-%d %H:%M:%S')

"
        
        # Добавляем критические проблемы, если есть
        local critical_issues=$(grep -A 2 "Security issue:" "$log_file" | head -10)
        if [[ -n "$critical_issues" ]]; then
            message+="Критические проблемы:
$critical_issues

"
        fi
        
        # Добавляем рекомендации
        local top_suggestions=$(grep -A 1 "Suggestion:" "$log_file" | head -6)
        if [[ -n "$top_suggestions" ]]; then
            message+="Топ рекомендаций:
$top_suggestions

"
        fi
        
        # Оценка безопасности
        if [[ $score -ge 80 ]]; then
            message+="Отличная безопасность!"
        elif [[ $score -ge 60 ]]; then
            message+="Хорошая безопасность, но есть возможности для улучшения"
        else
            message+="Требуется внимание к безопасности!"
        fi
        
        # Отправка результатов в Telegram
        send_telegram_message "$message"
        
        log_info "Результаты отправлены в Telegram"
    else
        log_error "Lynis завершился с ошибкой (exit code: $lynis_exit_code)"
        
        # Отправляем информацию об ошибке
        local error_info=$(tail -10 "$log_file" | head -5)
        send_telegram_message "Lynis завершился с ошибкой
Время: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)
Exit code: $lynis_exit_code
Последние строки лога:
$error_info"
        return 1
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Quick Lynis Test"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --test         Тестовая отправка в Telegram"
        echo
        echo "Примеры:"
        echo "  sudo $0              Запустить быстрый тест"
        echo "  sudo $0 --test       Тестовая отправка"
        ;;
    --test)
        if check_telegram_config; then
            send_telegram_message "Тестовое сообщение быстрого теста
Быстрый тест Lynis настроен правильно
Время: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)"
        fi
        ;;
    *)
        quick_lynis_test
        ;;
esac 