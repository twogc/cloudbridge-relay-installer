#!/bin/bash

# CloudBridge Relay - Lynis Audit with Telegram Notifications
# Запускает полный аудит безопасности и отправляет результаты в Telegram
# Отправляет уведомления только о новых событиях

set -e

# Конфигурация Telegram
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"

# Конфигурация таймаутов и проверок
# LYNIS_TIMEOUT=1200   # Убираем таймаут - пусть работает столько, сколько нужно
PID_FILE="/var/run/lynis_telegram.pid"
LOCK_FILE="/var/run/lynis_telegram.lock"

# Файлы для отслеживания изменений
STATE_DIR="/var/lib/lynis-telegram"
PREVIOUS_RESULTS="$STATE_DIR/previous_results.json"
CURRENT_RESULTS="$STATE_DIR/current_results.json"
CHANGES_LOG="$STATE_DIR/changes.log"

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
    if [[ ! -f "$PREVIOUS_RESULTS" ]]; then
        echo '{"warnings": 0, "suggestions": 0, "security_issues": 0, "score": 0, "issues": [], "suggestions": []}' > "$PREVIOUS_RESULTS"
    fi
}

# Извлечение результатов из лога Lynis
extract_results() {
    local log_file="$1"
    local output_file="$2"
    
    if [[ ! -f "$log_file" ]]; then
        log_error "Файл лога не найден: $log_file"
        return 1
    fi
    
    # Извлекаем основные метрики
    local warnings=$(grep -c "Warning:" "$log_file" || echo "0")
    local suggestions=$(grep -c "Suggestion:" "$log_file" || echo "0")
    local security_issues=$(grep -c "Security issue:" "$log_file" || echo "0")
    local score=$(grep "Hardening index" "$log_file" | tail -1 | grep -o '[0-9]\+' | head -1 || echo "0")
    
    # Извлекаем список проблем безопасности
    local issues=$(grep -A 1 "Security issue:" "$log_file" | grep -v "^--$" | head -20 | sed 's/^[[:space:]]*//' | jq -R -s -c 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')
    
    # Извлекаем список рекомендаций
    local suggestions_list=$(grep -A 1 "Suggestion:" "$log_file" | grep -v "^--$" | head -20 | sed 's/^[[:space:]]*//' | jq -R -s -c 'split("\n") | map(select(length > 0))' 2>/dev/null || echo '[]')
    
    # Создаем JSON с результатами
    cat > "$output_file" << EOF
{
    "warnings": $warnings,
    "suggestions": $suggestions,
    "security_issues": $security_issues,
    "score": $score,
    "issues": $issues,
    "suggestions": $suggestions_list,
    "timestamp": "$(date -Iseconds)"
}
EOF
}

# Сравнение результатов и определение изменений
detect_changes() {
    if [[ ! -f "$PREVIOUS_RESULTS" || ! -f "$CURRENT_RESULTS" ]]; then
        log_warn "Файлы результатов не найдены, считаем это первым запуском"
        return 1
    fi
    
    # Проверяем, есть ли jq для JSON сравнения
    if ! command -v jq &> /dev/null; then
        log_warn "jq не установлен, используем простое сравнение"
        return 1
    fi
    
    # Сравниваем основные метрики
    local prev_score=$(jq -r '.score' "$PREVIOUS_RESULTS" 2>/dev/null || echo "0")
    local curr_score=$(jq -r '.score' "$CURRENT_RESULTS" 2>/dev/null || echo "0")
    
    local prev_warnings=$(jq -r '.warnings' "$PREVIOUS_RESULTS" 2>/dev/null || echo "0")
    local curr_warnings=$(jq -r '.warnings' "$CURRENT_RESULTS" 2>/dev/null || echo "0")
    
    local prev_security_issues=$(jq -r '.security_issues' "$PREVIOUS_RESULTS" 2>/dev/null || echo "0")
    local curr_security_issues=$(jq -r '.security_issues' "$CURRENT_RESULTS" 2>/dev/null || echo "0")
    
    # Проверяем изменения
    local has_changes=false
    local changes_summary=""
    
    if [[ "$prev_score" != "$curr_score" ]]; then
        has_changes=true
        changes_summary+="• Hardening Index: $prev_score → $curr_score\n"
    fi
    
    if [[ "$prev_warnings" != "$curr_warnings" ]]; then
        has_changes=true
        changes_summary+="• Предупреждения: $prev_warnings → $curr_warnings\n"
    fi
    
    if [[ "$prev_security_issues" != "$curr_security_issues" ]]; then
        has_changes=true
        changes_summary+="• Проблемы безопасности: $prev_security_issues → $curr_security_issues\n"
    fi
    
    # Проверяем новые проблемы безопасности
    local new_issues=$(jq -r '.issues[]' "$CURRENT_RESULTS" 2>/dev/null | while read -r issue; do
        if ! jq -r '.issues[]' "$PREVIOUS_RESULTS" 2>/dev/null | grep -q "$issue"; then
            echo "$issue"
        fi
    done)
    
    if [[ -n "$new_issues" ]]; then
        has_changes=true
        changes_summary+="• Новые проблемы безопасности обнаружены\n"
    fi
    
    if [[ "$has_changes" == "true" ]]; then
        echo -e "$changes_summary"
        return 0
    else
        return 1
    fi
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

# Анализ результатов Lynis с учетом изменений
analyze_lynis_results() {
    local log_file="$1"
    local changes_summary="$2"
    
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
    local message="Lynis Security Audit завершен

Результаты:
• Hardening Index: ${score}/100
• Предупреждений: ${warnings}
• Рекомендаций: ${suggestions}
• Проблем безопасности: ${security_issues}

Лог файл: $log_file
Время завершения: $(date '+%Y-%m-%d %H:%M:%S')

"
    
    # Добавляем информацию об изменениях, если есть
    if [[ -n "$changes_summary" ]]; then
        message+="Обнаружены изменения:
$changes_summary

"
    fi
    
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
    
    # --- Автоматическая очистка lock-файла, если процесс не существует ---
    if [[ -f "$LOCK_FILE" ]]; then
        # Получаем PID процесса, который держит lock-файл (если flock записал PID)
        LOCK_PID=$(lsof -t "$LOCK_FILE" 2>/dev/null | head -1)
        if [[ -n "$LOCK_PID" ]]; then
            if ! ps -p "$LOCK_PID" > /dev/null 2>&1; then
                echo "[INFO] Lock-файл устарел (PID $LOCK_PID не существует), удаляю..."
                rm -f "$LOCK_FILE"
            fi
        else
            # Если lsof не находит процесс, просто удаляем lock-файл
            echo "[INFO] Lock-файл найден, но процесс не обнаружен, удаляю..."
            rm -f "$LOCK_FILE"
        fi
    fi
    
    # Инициализация директории состояния
    init_state_directory
    
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
    
    # Отправка уведомления о начале
    send_telegram_message "Lynis Security Audit начат
Время начала: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)"
    
    # Запускаем Lynis без таймаута
    if lynis audit system --quick > "$log_file" 2>&1; then
        lynis_exit_code=$?
    else
        lynis_exit_code=$?
    fi
    
    log_info "Аудит завершен (exit code: $lynis_exit_code)"
    
    # Проверяем успешность выполнения
    if [[ $lynis_exit_code -eq 0 ]]; then
        # Извлекаем текущие результаты
        extract_results "$log_file" "$CURRENT_RESULTS"
        
        # Определяем изменения
        local changes_summary=$(detect_changes)
        local has_changes=$?
        
        # Отправляем уведомление только если есть изменения или это первый запуск
        if [[ $has_changes -eq 0 || ! -f "$PREVIOUS_RESULTS" ]]; then
            # Анализ результатов
            local report=$(analyze_lynis_results "$log_file" "$changes_summary")
            
            # Отправка результатов в Telegram
            send_telegram_message "$report"
            
            log_info "Результаты отправлены в Telegram (обнаружены изменения)"
            
            # Сохраняем текущие результаты как предыдущие
            cp "$CURRENT_RESULTS" "$PREVIOUS_RESULTS"
            
            # Логируем изменения
            echo "$(date -Iseconds): Обнаружены изменения в результатах Lynis" >> "$CHANGES_LOG"
        else
            log_info "Изменений не обнаружено, уведомление не отправляется"
            
            # Обновляем предыдущие результаты для следующего сравнения
            cp "$CURRENT_RESULTS" "$PREVIOUS_RESULTS"
        fi
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
    echo
    echo "Для корректной работы отслеживания изменений рекомендуется установить jq:"
    echo "sudo apt-get install jq"
}

# Функция сброса состояния
reset_state() {
    log_info "Сброс состояния отслеживания изменений"
    rm -f "$PREVIOUS_RESULTS" "$CURRENT_RESULTS"
    echo "Состояние сброшено. Следующий запуск будет считаться первым."
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
        echo "  --reset        Сброс состояния отслеживания изменений"
        echo "  --force        Принудительная отправка (игнорируя изменения)"
        echo
        echo "Примеры:"
        echo "  sudo $0              Запустить аудит с уведомлениями"
        echo "  sudo $0 --setup      Настройка Telegram"
        echo "  sudo $0 --test       Тестовая отправка"
        echo "  sudo $0 --reset      Сброс состояния"
        echo "  sudo $0 --force      Принудительная отправка"
        ;;
    --setup)
        setup_telegram
        ;;
    --test)
        if check_telegram_config; then
            send_telegram_message "Тестовое сообщение
Telegram уведомления настроены правильно
Время: $(date '+%Y-%m-%d %H:%M:%S')
Сервер: $(hostname)"
        fi
        ;;
    --reset)
        reset_state
        ;;
    --force)
        # Принудительная отправка - копируем текущие результаты как предыдущие
        # чтобы следующее сравнение показало изменения
        if [[ -f "$CURRENT_RESULTS" ]]; then
            cp "$CURRENT_RESULTS" "$PREVIOUS_RESULTS"
        fi
        main
        ;;
    *)
        main
        ;;
esac 