#!/bin/bash

# CloudBridge Relay - Security Monitor with Telegram Notifications
# Мониторинг различных событий безопасности с отправкой уведомлений только о новых событиях

set -e

# Конфигурация Telegram
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"

# Конфигурация мониторинга
STATE_DIR="/var/lib/security-monitor"
FAIL2BAN_LOG="/var/log/fail2ban.log"
AUTH_LOG="/var/log/auth.log"
SYSTEM_LOG="/var/log/syslog"
LAST_CHECK_FILE="$STATE_DIR/last_check.json"
EVENTS_LOG="$STATE_DIR/events.log"

# Интервалы проверки (в секундах)
FAIL2BAN_CHECK_INTERVAL=300  # 5 минут
AUTH_CHECK_INTERVAL=600      # 10 минут
SYSTEM_CHECK_INTERVAL=1800   # 30 минут

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

# Инициализация директории состояния
init_state_directory() {
    mkdir -p "$STATE_DIR"
    if [[ ! -f "$LAST_CHECK_FILE" ]]; then
        cat > "$LAST_CHECK_FILE" << EOF
{
    "fail2ban_last_check": "$(date -Iseconds)",
    "auth_last_check": "$(date -Iseconds)",
    "system_last_check": "$(date -Iseconds)",
    "last_events": {}
}
EOF
    fi
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
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}

# Проверка новых событий fail2ban
check_fail2ban_events() {
    if [[ ! -f "$FAIL2BAN_LOG" ]]; then
        log_warn "Файл fail2ban.log не найден"
        return
    fi
    
    local last_check=$(jq -r '.fail2ban_last_check' "$LAST_CHECK_FILE" 2>/dev/null || echo "$(date -Iseconds)")
    local current_time=$(date -Iseconds)
    
    # Получаем новые события блокировки
    local new_blocks=$(awk -v last="$last_check" '
        BEGIN {
            last_epoch = mktime(gensub(/[-:T]/, " ", "g", last))
        }
        /Ban/ && /sshd/ {
            # Парсим дату из лога
            split($1" "$2, datetime, " ")
            log_time = datetime[1] " " datetime[2]
            log_epoch = mktime(gensub(/[-:]/, " ", "g", log_time))
            
            if (log_epoch > last_epoch) {
                print $0
            }
        }
    ' "$FAIL2BAN_LOG")
    
    if [[ -n "$new_blocks" ]]; then
        local block_count=$(echo "$new_blocks" | wc -l)
        local message="ALERT: Обнаружены новые блокировки fail2ban

Статистика:
• Количество новых блокировок: $block_count
• Временной интервал: $last_check → $current_time

Последние блокировки:
$(echo "$new_blocks" | tail -5 | sed 's/^/• /')

Сервер: $(hostname)
Время проверки: $(date '+%Y-%m-%d %H:%M:%S')"
        
        send_telegram_message "$message"
        
        # Обновляем время последней проверки
        jq --arg time "$current_time" '.fail2ban_last_check = $time' "$LAST_CHECK_FILE" > "$LAST_CHECK_FILE.tmp" && mv "$LAST_CHECK_FILE.tmp" "$LAST_CHECK_FILE"
        
        # Логируем событие
        echo "$(date -Iseconds): Обнаружено $block_count новых блокировок fail2ban" >> "$EVENTS_LOG"
    fi
}

# Проверка новых событий аутентификации
check_auth_events() {
    if [[ ! -f "$AUTH_LOG" ]]; then
        log_warn "Файл auth.log не найден"
        return
    fi
    
    local last_check=$(jq -r '.auth_last_check' "$LAST_CHECK_FILE" 2>/dev/null || echo "$(date -Iseconds)")
    local current_time=$(date -Iseconds)
    
    # Получаем новые неудачные попытки входа
    local new_failed_logins=$(awk -v last="$last_check" '
        BEGIN {
            last_epoch = mktime(gensub(/[-:T]/, " ", "g", last))
        }
        /Failed password/ {
            # Парсим дату из лога
            split($1" "$2" "$3, datetime, " ")
            log_time = datetime[1] " " datetime[2] " " datetime[3]
            log_epoch = mktime(gensub(/[-:]/, " ", "g", log_time))
            
            if (log_epoch > last_epoch) {
                print $0
            }
        }
    ' "$AUTH_LOG")
    
    if [[ -n "$new_failed_logins" ]]; then
        local failed_count=$(echo "$new_failed_logins" | wc -l)
        local unique_ips=$(echo "$new_failed_logins" | grep -oE 'from [0-9]+\.[0-9]+\.[0-9]+\.[0-9]+' | cut -d' ' -f2 | sort | uniq | wc -l)
        
        local message="ALERT: Обнаружены новые неудачные попытки входа

Статистика:
• Количество неудачных попыток: $failed_count
• Уникальных IP адресов: $unique_ips
• Временной интервал: $last_check → $current_time

Последние попытки:
$(echo "$new_failed_logins" | tail -3 | sed 's/^/• /')

Сервер: $(hostname)
Время проверки: $(date '+%Y-%m-%d %H:%M:%S')"
        
        send_telegram_message "$message"
        
        # Обновляем время последней проверки
        jq --arg time "$current_time" '.auth_last_check = $time' "$LAST_CHECK_FILE" > "$LAST_CHECK_FILE.tmp" && mv "$LAST_CHECK_FILE.tmp" "$LAST_CHECK_FILE"
        
        # Логируем событие
        echo "$(date -Iseconds): Обнаружено $failed_count новых неудачных попыток входа" >> "$EVENTS_LOG"
    fi
}

# Проверка системных событий
check_system_events() {
    if [[ ! -f "$SYSTEM_LOG" ]]; then
        log_warn "Файл syslog не найден"
        return
    fi
    
    local last_check=$(jq -r '.system_last_check' "$LAST_CHECK_FILE" 2>/dev/null || echo "$(date -Iseconds)")
    local current_time=$(date -Iseconds)
    
    # Получаем критические системные события
    local critical_events=$(awk -v last="$last_check" '
        BEGIN {
            last_epoch = mktime(gensub(/[-:T]/, " ", "g", last))
        }
        /error|Error|ERROR|critical|Critical|CRITICAL|emergency|Emergency|EMERGENCY/ {
            # Парсим дату из лога
            split($1" "$2" "$3, datetime, " ")
            log_time = datetime[1] " " datetime[2] " " datetime[3]
            log_epoch = mktime(gensub(/[-:]/, " ", "g", log_time))
            
            if (log_epoch > last_epoch) {
                # Фильтруем некритичные ошибки Docker
                if ($0 !~ /dockerd.*ShouldRestart failed/ && 
                    $0 !~ /dockerd.*restart canceled/ && 
                    $0 !~ /dockerd.*Error handling plugin refcount operation/ && 
                    $0 !~ /dockerd.*legacy plugin: Unable to find plugin without name/ && 
                    $0 !~ /dockerd.*container will not be restarted/ && 
                    $0 !~ /dockerd.*hasBeenManuallyStopped=true/ && 
                    $0 !~ /dockerd.*daemonShuttingDown=false/ &&
                    $0 !~ /dockerd.*Error running exec/ &&
                    $0 !~ /dockerd.*stream copy error/ &&
                    $0 !~ /dockerd.*executable file not found/ &&
                    $0 !~ /dockerd.*unable to start container process/ &&
                    $0 !~ /containerd.*cleanup warnings/ &&
                    $0 !~ /containerd.*failed to remove runc container/ &&
                    $0 !~ /containerd.*runc did not terminate successfully/ &&
                    $0 !~ /cloudflared/ &&
                    $0 !~ /freshclam.*error code 429/ &&
                    $0 !~ /freshclam.*error code 403/) {
                    print $0
                }
            }
        }
    ' "$SYSTEM_LOG")
    
    if [[ -n "$critical_events" ]]; then
        local event_count=$(echo "$critical_events" | wc -l)
        local message="ALERT: Обнаружены критические системные события

Статистика:
• Количество критических событий: $event_count
• Временной интервал: $last_check → $current_time

Последние события:
$(echo "$critical_events" | tail -5 | sed 's/^/• /')

Сервер: $(hostname)
Время проверки: $(date '+%Y-%m-%d %H:%M:%S')"
        
        send_telegram_message "$message"
        
        # Обновляем время последней проверки
        jq --arg time "$current_time" '.system_last_check = $time' "$LAST_CHECK_FILE" > "$LAST_CHECK_FILE.tmp" && mv "$LAST_CHECK_FILE.tmp" "$LAST_CHECK_FILE"
        
        # Логируем событие
        echo "$(date -Iseconds): Обнаружено $event_count критических системных событий" >> "$EVENTS_LOG"
    fi
}

# Проверка использования диска
check_disk_usage() {
    local disk_usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    local threshold=80
    
    if [[ $disk_usage -gt $threshold ]]; then
        local message="ALERT: Высокое использование диска

Статистика:
• Использование диска: ${disk_usage}%
• Пороговое значение: ${threshold}%

Детали:
$(df -h / | tail -1)

Сервер: $(hostname)
Время проверки: $(date '+%Y-%m-%d %H:%M:%S')"
        
        send_telegram_message "$message"
        echo "$(date -Iseconds): Высокое использование диска: ${disk_usage}%" >> "$EVENTS_LOG"
    fi
}

# Проверка использования памяти
check_memory_usage() {
    local mem_usage=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
    local threshold=85
    
    if [[ $mem_usage -gt $threshold ]]; then
        local message="ALERT: Высокое использование памяти

Статистика:
• Использование памяти: ${mem_usage}%
• Пороговое значение: ${threshold}%

Детали:
$(free -h | head -2)

Сервер: $(hostname)
Время проверки: $(date '+%Y-%m-%d %H:%M:%S')"
        
        send_telegram_message "$message"
        echo "$(date -Iseconds): Высокое использование памяти: ${mem_usage}%" >> "$EVENTS_LOG"
    fi
}

# Основная функция мониторинга
main_monitoring() {
    log_info "Запуск мониторинга безопасности"
    
    # Инициализация директории состояния
    init_state_directory
    
    # Проверка различных событий
    check_fail2ban_events
    check_auth_events
    check_system_events
    check_disk_usage
    check_memory_usage
    
    log_info "Мониторинг завершен"
}

# Функция непрерывного мониторинга
continuous_monitoring() {
    log_info "Запуск непрерывного мониторинга безопасности"
    
    while true; do
        main_monitoring
        sleep 300  # Проверка каждые 5 минут
    done
}

# Функция сброса состояния
reset_state() {
    log_info "Сброс состояния мониторинга"
    rm -f "$LAST_CHECK_FILE" "$EVENTS_LOG"
    echo "Состояние сброшено. Следующий запуск будет считаться первым."
}

# Функция показа статистики
show_stats() {
    if [[ -f "$EVENTS_LOG" ]]; then
        echo "Статистика событий безопасности:"
        echo "================================"
        echo "Последние 10 событий:"
        tail -10 "$EVENTS_LOG"
        echo
        echo "Общее количество событий: $(wc -l < "$EVENTS_LOG")"
    else
        echo "Файл событий не найден"
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Security Monitor with Telegram Notifications"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --once         Запустить однократную проверку"
        echo "  --continuous   Запустить непрерывный мониторинг"
        echo "  --reset        Сброс состояния мониторинга"
        echo "  --stats        Показать статистику событий"
        echo "  --test         Тестовая отправка в Telegram"
        echo
        echo "Примеры:"
        echo "  sudo $0 --once        Однократная проверка"
        echo "  sudo $0 --continuous  Непрерывный мониторинг"
        echo "  sudo $0 --reset       Сброс состояния"
        echo "  sudo $0 --stats       Статистика"
        echo "  sudo $0 --test        Тестовая отправка"
        ;;
    --once)
        main_monitoring
        ;;
    --continuous)
        continuous_monitoring
        ;;
    --reset)
        reset_state
        ;;
    --stats)
        show_stats
        ;;
    --test)
        if check_telegram_config; then
            send_telegram_message "Тестовое сообщение мониторинга
Система мониторинга безопасности работает
Время: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)"
        fi
        ;;
    *)
        main_monitoring
        ;;
esac 