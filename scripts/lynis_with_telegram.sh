#!/bin/bash

# CloudBridge Relay - Lynis Audit with Telegram Notifications
# Запускает полный аудит безопасности и отправляет результаты в Telegram

set -e

# Конфигурация Telegram
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"

# Конфигурация таймаутов и проверок
LYNIS_TIMEOUT=1800  # 30 минут максимум
PID_FILE="/var/run/lynis_telegram.pid"
LOCK_FILE="/var/run/lynis_telegram.lock"

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

# Проверка на уже запущенные процессы
check_running_processes() {
    if [[ -f "$PID_FILE" ]]; then
        local pid=$(cat "$PID_FILE" 2>/dev/null)
        if [[ -n "$pid" && -d "/proc/$pid" ]]; then
            log_warn "Lynis уже запущен (PID: $pid)"
            return 1
        else
            # Удаляем устаревший PID файл
            rm -f "$PID_FILE"
        fi
    fi
    
    # Проверяем процессы lynis (более точная проверка)
    local lynis_processes=$(ps aux | grep -E "lynis.*audit" | grep -v grep | grep -v "systemctl\|sudo\|xargs\|ps" | wc -l)
    if [[ $lynis_processes -gt 0 ]]; then
        log_warn "Обнаружены запущенные процессы Lynis"
        ps aux | grep -E "lynis.*audit" | grep -v grep | grep -v "systemctl\|sudo\|xargs\|ps"
        return 1
    fi
    
    return 0
}

# Создание PID файла
create_pid_file() {
    echo $$ > "$PID_FILE"
    trap 'cleanup_pid_file' EXIT
}

# Очистка PID файла
cleanup_pid_file() {
    rm -f "$PID_FILE"
    rm -f "$LOCK_FILE"
}

# Проверка блокировки
check_lock() {
    if [[ -f "$LOCK_FILE" ]]; then
        local lock_time=$(stat -c %Y "$LOCK_FILE" 2>/dev/null || echo 0)
        local current_time=$(date +%s)
        local lock_age=$((current_time - lock_time))
        
        if [[ $lock_age -gt $LYNIS_TIMEOUT ]]; then
            log_warn "Удаляю устаревшую блокировку (возраст: ${lock_age}с)"
            rm -f "$LOCK_FILE"
        else
            log_error "Lynis заблокирован (возраст блокировки: ${lock_age}с)"
            return 1
        fi
    fi
    
    touch "$LOCK_FILE"
    return 0
}

# Проверка конфигурации Telegram
check_telegram_config() {
    if [[ -z "$TELEGRAM_BOT_TOKEN" || -z "$TELEGRAM_CHAT_ID" ]]; then
        log_warn "Telegram конфигурация не настроена"
        log_info "Для настройки укажите TELEGRAM_BOT_TOKEN и TELEGRAM_CHAT_ID в скрипте"
        return 1
    fi
    return 0
}

# Отправка сообщения в Telegram
send_telegram_message() {
    local message="$1"
    
    if ! check_telegram_config; then
        log_warn "Пропускаю отправку в Telegram"
        return
    fi
    
    log_info "Отправка уведомления в Telegram..."
    
    # Экранируем специальные символы для URL
    local escaped_message=$(echo "$message" | sed 's/&/\&amp;/g; s/</\&lt;/g; s/>/\&gt;/g; s/"/\&quot;/g')
    
    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${escaped_message}" \
        -d "parse_mode=HTML" \
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}

# Анализ результатов Lynis
analyze_lynis_results() {
    local log_file="$1"
    
    if [[ ! -f "$log_file" ]]; then
        log_error "Файл лога не найден: $log_file"
        return 1
    fi
    
    log_info "Анализ результатов Lynis..."
    
    # Извлекаем основные метрики
    local warnings=$(grep -c "Warning:" "$log_file" || echo "0")
    local suggestions=$(grep -c "Suggestion:" "$log_file" || echo "0")
    local security_issues=$(grep -c "Security issue:" "$log_file" || echo "0")
    local score=$(grep "Hardening index" "$log_file" | tail -1 | grep -o '[0-9]\+' | head -1 || echo "0")
    
    # Формируем сообщение
    local message="🔒 <b>Lynis Security Audit завершен</b>
    
📊 <b>Результаты:</b>
• Hardening Index: ${score}/100
• Предупреждений: ${warnings}
• Рекомендаций: ${suggestions}
• Проблем безопасности: ${security_issues}

📁 Лог файл: $log_file
🕐 Время завершения: $(date '+%Y-%m-%d %H:%M:%S')

"
    
    # Добавляем критические проблемы, если есть
    local critical_issues=$(grep -A 2 "Security issue:" "$log_file" | head -10)
    if [[ -n "$critical_issues" ]]; then
        message+="⚠️ <b>Критические проблемы:</b>
$critical_issues

"
    fi
    
    # Добавляем рекомендации
    local top_suggestions=$(grep -A 1 "Suggestion:" "$log_file" | head -6)
    if [[ -n "$top_suggestions" ]]; then
        message+="💡 <b>Топ рекомендаций:</b>
$top_suggestions

"
    fi
    
    # Оценка безопасности
    if [[ $score -ge 80 ]]; then
        message+="✅ <b>Отличная безопасность!</b>"
    elif [[ $score -ge 60 ]]; then
        message+="⚠️ <b>Хорошая безопасность, но есть возможности для улучшения</b>"
    else
        message+="🚨 <b>Требуется внимание к безопасности!</b>"
    fi
    
    echo "$message"
}

# Основная функция
main() {
    log_info "Запуск Lynis Security Audit с уведомлениями в Telegram"
    
    # Проверка root прав
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root"
        exit 1
    fi
    
    # Проверка на уже запущенные процессы
    if ! check_running_processes; then
        log_error "Lynis уже запущен. Завершите предыдущий процесс или подождите."
        exit 1
    fi
    
    # Проверка блокировки
    if ! check_lock; then
        log_error "Lynis заблокирован. Попробуйте позже."
        exit 1
    fi
    
    # Создание PID файла
    create_pid_file
    
    # Создание директории для логов
    mkdir -p logs
    
    # Генерация имени файла лога
    local log_file="logs/lynis-$(date +%F-%H%M).log"
    
    log_info "Начало аудита безопасности..."
    log_info "Лог файл: $log_file"
    log_info "Таймаут: ${LYNIS_TIMEOUT} секунд"
    
    # Отправка уведомления о начале
    send_telegram_message "🔍 <b>Lynis Security Audit начат</b>
🕐 Время начала: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"
    
    # Запуск Lynis с таймаутом
    local start_time=$(date +%s)
    local lynis_exit_code=0
    
    # Запускаем Lynis с таймаутом
    if timeout $LYNIS_TIMEOUT lynis audit system > "$log_file" 2>&1; then
        lynis_exit_code=$?
    else
        lynis_exit_code=$?
        if [[ $lynis_exit_code -eq 124 ]]; then
            log_error "Lynis превысил таймаут (${LYNIS_TIMEOUT} секунд)"
            send_telegram_message "⏰ <b>Lynis превысил таймаут</b>
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)
⏱️ Таймаут: ${LYNIS_TIMEOUT} секунд"
            exit 1
        fi
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_info "Аудит завершен за ${duration} секунд (exit code: $lynis_exit_code)"
    
    # Проверяем успешность выполнения
    if [[ $lynis_exit_code -eq 0 ]]; then
        # Анализ результатов
        local report=$(analyze_lynis_results "$log_file")
        
        # Отправка результатов в Telegram
        send_telegram_message "$report"
        
        log_info "Результаты отправлены в Telegram"
    else
        log_error "Lynis завершился с ошибкой (exit code: $lynis_exit_code)"
        
        # Отправляем информацию об ошибке
        local error_info=$(tail -10 "$log_file" | head -5)
        send_telegram_message "❌ <b>Lynis завершился с ошибкой</b>
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)
📊 Exit code: $lynis_exit_code
📋 Последние строки лога:
$error_info"
        exit 1
    fi
}

# Функция настройки Telegram
setup_telegram() {
    log_info "Настройка Telegram уведомлений"
    
    echo "Для настройки Telegram уведомлений:"
    echo "1. Создайте бота через @BotFather"
    echo "2. Получите токен бота"
    echo "3. Получите Chat ID (отправьте сообщение боту и проверьте https://api.telegram.org/bot<TOKEN>/getUpdates)"
    echo
    echo "Затем отредактируйте скрипт и укажите:"
    echo "TELEGRAM_BOT_TOKEN=\"ваш_токен_бота\""
    echo "TELEGRAM_CHAT_ID=\"ваш_chat_id\""
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Lynis Audit with Telegram Notifications"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --setup        Настройка Telegram"
        echo "  --test         Тестовая отправка в Telegram"
        echo
        echo "Примеры:"
        echo "  sudo $0              Запустить аудит с уведомлениями"
        echo "  sudo $0 --setup      Настройка Telegram"
        echo "  sudo $0 --test       Тестовая отправка"
        ;;
    --setup)
        setup_telegram
        ;;
    --test)
        if check_telegram_config; then
            send_telegram_message "🧪 <b>Тестовое сообщение</b>
✅ Telegram уведомления настроены правильно
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"
        fi
        ;;
    *)
        main
        ;;
esac 