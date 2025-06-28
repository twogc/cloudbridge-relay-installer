#!/bin/bash

# CloudBridge Relay - Security Monitor
# Автоматический мониторинг безопасности с уведомлениями в Telegram

set -e

# Конфигурация Telegram
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"

# Конфигурация мониторинга
LOG_FILE="logs/security-monitor.log"
ALERT_FILE="logs/security-alerts.log"
CHECK_INTERVAL=300  # 5 минут

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Отправка сообщения в Telegram
send_telegram_message() {
    local message="$1"
    
    if [[ -z "$TELEGRAM_BOT_TOKEN" || -z "$TELEGRAM_CHAT_ID" ]]; then
        log_warn "Telegram конфигурация не настроена"
        return
    fi

    # Удаляем все HTML-теги
    message=$(echo "$message" | sed 's/<[^>]*>//g')
    # Экранируем спецсимволы MarkdownV2 по отдельности
    message=$(echo "$message" | sed 's/[_]/\\_/g' | sed 's/[*]/\\*/g' | sed 's/\[/\\[/g' | sed 's/\]/\\]/g' | sed 's/(/\\(/g' | sed 's/)/\\)/g' | sed 's/~/\\~/g' | sed 's/`/\\`/g' | sed 's/>/\\>/g' | sed 's/#/\\#/g' | sed 's/+/\\+/g' | sed 's/-/\\-/g' | sed 's/=/\\=/g' | sed 's/|/\\|/g' | sed 's/{/\\{/g' | sed 's/}/\\}/g' | sed 's/\./\\./g' | sed 's/!/\\!/g')

    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${message}" \
        -d "parse_mode=MarkdownV2" \
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}

# Проверка новых блокировок Fail2Ban
check_fail2ban_blocks() {
    local current_blocks=$(sudo fail2ban-client status | grep "Total banned" | awk '{print $4}' | tr '\n' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    local last_blocks_file="logs/last_fail2ban_blocks.txt"
    
    # Если current_blocks пустой, устанавливаем 0
    if [[ -z "$current_blocks" ]]; then
        current_blocks="0"
    fi
    
    if [[ ! -f "$last_blocks_file" ]]; then
        echo "$current_blocks" > "$last_blocks_file"
        return 1
    fi
    
    local last_blocks=$(cat "$last_blocks_file" 2>/dev/null || echo "0")
    
    # Если last_blocks пустой, устанавливаем 0
    if [[ -z "$last_blocks" ]]; then
        last_blocks="0"
    fi
    
    if [[ "$current_blocks" != "$last_blocks" ]]; then
        # Безопасное преобразование в числа
        local new_blocks=$(echo "$current_blocks" | tr ' ' '\n' | wc -l)
        local old_blocks=$(echo "$last_blocks" | tr ' ' '\n' | wc -l)
        
        # Проверяем, что переменные являются числами
        if [[ "$new_blocks" =~ ^[0-9]+$ && "$old_blocks" =~ ^[0-9]+$ ]]; then
            local diff=$((new_blocks - old_blocks))
            
            if [[ $diff -gt 0 ]]; then
                local message="🚨 <b>Новые блокировки Fail2Ban</b>
                
📊 <b>Статистика:</b>
• Новых блокировок: $diff
• Всего заблокированных IP: $new_blocks

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

                send_telegram_message "$message"
                echo "$current_blocks" > "$last_blocks_file"
                return 0
            fi
        fi
        
        echo "$current_blocks" > "$last_blocks_file"
    fi
    
    return 1
}

# Проверка критических сервисов
check_critical_services() {
    local failed_services=()
    
    # Проверяем критически важные сервисы
    local critical_services=("ufw" "fail2ban" "sshd" "rsyslog")
    
    for service in "${critical_services[@]}"; do
        if ! systemctl is-active --quiet "$service"; then
            failed_services+=("$service")
        fi
    done
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        local services_list=$(printf '%s\n' "${failed_services[@]}" | tr '\n' ', ' | sed 's/,$//')
        local message="⚠️ <b>Критические сервисы не работают</b>
        
🔴 <b>Неактивные сервисы:</b>
• $services_list

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        return 0
    fi
    
    return 1
}

# Проверка дискового пространства
check_disk_space() {
    local threshold=85  # Процент использования
    local usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [[ $usage -gt $threshold ]]; then
        local free_space=$(df -h / | awk 'NR==2 {print $4}')
        local message="💾 <b>Мало места на диске</b>
        
📊 <b>Использование диска:</b>
• Использовано: ${usage}%
• Свободно: $free_space
• Порог: ${threshold}%

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        return 0
    fi
    
    return 1
}

# Проверка обновлений системы
check_system_updates() {
    local updates_count=$(apt list --upgradable 2>/dev/null | grep -c upgradable || echo "0")
    updates_count=$(echo "$updates_count" | tr -d '[:space:]')
    if [[ "$updates_count" =~ ^[0-9]+$ && $updates_count -gt 10 ]]; then
        local message="🔄 <b>Доступны обновления системы</b>
        
📦 <b>Обновления:</b>
• Количество пакетов: $updates_count
• Рекомендуется обновить систему

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        return 0
    fi
    
    return 1
}

# Проверка подозрительной активности
check_suspicious_activity() {
    local suspicious_found=false
    
    # Проверка множественных неудачных SSH попыток
    local ssh_failures=$(grep "Failed password" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
    
    if [[ $ssh_failures -gt 50 ]]; then
        local message="🚨 <b>Подозрительная активность SSH</b>
        
🔍 <b>Обнаружено:</b>
• Неудачных попыток входа: $ssh_failures
• Возможная атака брутфорс

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        suspicious_found=true
    fi
    
    # Проверка новых процессов от root
    local root_processes=$(ps aux | grep -v grep | grep "^root" | wc -l)
    if [[ $root_processes -gt 20 ]]; then
        local message="⚠️ <b>Много процессов от root</b>
        
🔍 <b>Обнаружено:</b>
• Процессов от root: $root_processes
• Рекомендуется проверить

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        suspicious_found=true
    fi
    
    if [[ "$suspicious_found" == "true" ]]; then
        return 0
    fi
    
    return 1
}

# Проверка SSL сертификатов
check_ssl_certificates() {
    if [[ -f "certs/fullchain.pem" ]]; then
        local expiry_date=$(openssl x509 -enddate -noout -in certs/fullchain.pem | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        if [[ $days_until_expiry -lt 30 ]]; then
            local message="🔐 <b>SSL сертификат скоро истечет</b>
            
📅 <b>Сертификат:</b>
• Дней до истечения: $days_until_expiry
• Дата истечения: $expiry_date

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

            send_telegram_message "$message"
            return 0
        fi
    fi
    
    return 1
}

# Проверка логов на критические ошибки
check_critical_logs() {
    local critical_errors=$(grep -i "error\|critical\|emergency" /var/log/syslog | grep "$(date '+%b %d')" | tail -5)
    
    if [[ -n "$critical_errors" ]]; then
        local error_count=$(echo "$critical_errors" | wc -l)
        local message="🚨 <b>Критические ошибки в логах</b>
        
📋 <b>Ошибки:</b>
• Количество: $error_count
• Последние ошибки:
$critical_errors

🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"

        send_telegram_message "$message"
        return 0
    fi
    
    return 1
}

# Основная функция мониторинга
run_security_check() {
    log_info "Запуск проверки безопасности..."
    
    local problems_found=false
    
    # Проверяем fail2ban блокировки
    log_info "Проверка fail2ban блокировок..."
    if check_fail2ban_blocks; then
        log_info "Найдены новые блокировки fail2ban"
        problems_found=true
    fi
    
    # Проверяем критические сервисы
    log_info "Проверка критических сервисов..."
    if check_critical_services; then
        log_info "Найдены неработающие сервисы"
        problems_found=true
    fi
    
    # Проверяем дисковое пространство
    log_info "Проверка дискового пространства..."
    if check_disk_space; then
        log_info "Найдены проблемы с диском"
        problems_found=true
    fi
    
    # Проверяем обновления системы
    log_info "Проверка обновлений системы..."
    if check_system_updates; then
        log_info "Найдены обновления системы"
        problems_found=true
    fi
    
    # Проверяем подозрительную активность
    log_info "Проверка подозрительной активности..."
    if check_suspicious_activity; then
        log_info "Найдена подозрительная активность"
        problems_found=true
    fi
    
    # Проверяем SSL сертификаты
    log_info "Проверка SSL сертификатов..."
    if check_ssl_certificates; then
        log_info "Найдены проблемы с SSL сертификатами"
        problems_found=true
    fi
    
    # Проверяем критические логи
    log_info "Проверка критических логов..."
    if check_critical_logs; then
        log_info "Найдены критические ошибки в логах"
        problems_found=true
    fi
    
    if [[ "$problems_found" == "false" ]]; then
        log_info "Проблем не обнаружено"
    fi
    
    log_info "Проверка безопасности завершена"
}

# Функция непрерывного мониторинга
continuous_monitoring() {
    log_info "Запуск непрерывного мониторинга безопасности"
    send_telegram_message "🛡️ <b>Мониторинг безопасности запущен</b>
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"
    
    while true; do
        run_security_check
        sleep $CHECK_INTERVAL
    done
}

# Функция разовой проверки
single_check() {
    log_info "Выполнение разовой проверки безопасности"
    run_security_check
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Security Monitor"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --once         Выполнить разовую проверку"
        echo "  --continuous   Запустить непрерывный мониторинг"
        echo "  --test         Тестовая отправка в Telegram"
        echo
        echo "Примеры:"
        echo "  $0 --once              Разовая проверка"
        echo "  $0 --continuous        Непрерывный мониторинг"
        echo "  $0 --test              Тестовая отправка"
        ;;
    --once)
        single_check
        ;;
    --continuous)
        continuous_monitoring
        ;;
    --test)
        send_telegram_message "🧪 <b>Тест мониторинга безопасности</b>
✅ Система мониторинга работает
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
🖥️ Сервер: $(hostname)"
        ;;
    *)
        single_check
        ;;
esac 