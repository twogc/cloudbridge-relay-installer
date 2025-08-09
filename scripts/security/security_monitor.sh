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

# Фильтры для UFW блокировок (можно настроить)
# Чтобы включить/отключить фильтры, измените значения на true/false
FILTER_SSH_BLOCKS=true   # true = фильтровать SSH блокировки (порт 22), false = показывать
FILTER_DHCP_BLOCKS=true  # true = фильтровать DHCP блокировки (порты 67/68), false = показывать

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

    # Заменяем <br> на \n для правильного форматирования
    message=$(echo "$message" | sed 's/<br>/\n/g')
    # Экранируем только амперсанд для HTML-форматирования
    message=$(echo "$message" | sed 's/&/\&amp;/g')
    


    local response=$(curl -s -X POST \
        "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
        -d "chat_id=${TELEGRAM_CHAT_ID}" \
        -d "text=${message}" \
        -d "parse_mode=HTML" \
        2>/dev/null)
    
    if echo "$response" | grep -q '"ok":true'; then
        log_info "Сообщение отправлено в Telegram"
    else
        log_error "Ошибка отправки в Telegram: $response"
    fi
}

# Фильтрация сообщений, которые не нужно отправлять в Telegram
should_ignore_alert() {
    local message="$1"
    # Фильтруем только действительно некритичные сообщения
    if echo "$message" | grep -q -E "emergency\\.service|icmp_ignore_bogus_error_responses|Install a file integrity tool|SPT=67 DPT=68|Check the AIDE database|ShouldRestart failed|restart canceled|Error handling plugin refcount operation|legacy plugin: Unable to find plugin without name|container will not be restarted|hasBeenManuallyStopped=true|daemonShuttingDown=false|Error running exec|stream copy error|executable file not found|unable to start container process|cleanup warnings|failed to remove runc container|runc did not terminate successfully"; then
        return 0  # игнорировать
    fi
    return 1  # не игнорировать
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
                # Получаем детальную информацию о jail'ах
                local jail_info=""
                local jails=$(sudo fail2ban-client status | grep "Jail list" | cut -d: -f2 | tr ',' ' ')
                
                for jail in $jails; do
                    jail=$(echo "$jail" | xargs)  # Убираем пробелы
                    if [[ -n "$jail" ]]; then
                        local jail_status=$(sudo fail2ban-client status "$jail" 2>/dev/null)
                        local banned_count=$(echo "$jail_status" | grep "Currently banned" | awk '{print $4}')
                        local total_banned=$(echo "$jail_status" | grep "Total banned" | awk '{print $4}')
                        
                        if [[ -n "$banned_count" && "$banned_count" != "0" ]]; then
                            jail_info+="<br>• <b>$jail:</b> $banned_count активных, $total_banned всего"
                        fi
                    fi
                done
                
                local message="[ALERT] <b>Новые блокировки Fail2Ban</b><br><br>[STATS] <b>Статистика:</b><br>• Новых блокировок: $diff<br>• Всего заблокированных IP: $new_blocks<br><br>[DETAILS] <b>Детали по jail'ам:</b>$jail_info<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

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
    local critical_services=("fail2ban" "sshd" "rsyslog")
    
    # Проверяем Docker контейнеры
    if command -v docker &> /dev/null; then
        if ! docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "cloudbridge-relay.*Up"; then
            failed_services+=("cloudbridge-relay")
        fi
        if ! docker ps --format "table {{.Names}}\t{{.Status}}" | grep -q "cloudbridge-keycloak.*Up"; then
            failed_services+=("cloudbridge-keycloak")
        fi
    fi
    
    for service in "${critical_services[@]}"; do
        if ! systemctl is-active --quiet "$service"; then
            failed_services+=("$service")
        fi
    done
    
    # Специальная проверка UFW
    if ! sudo ufw status | grep -q "Status: active"; then
        failed_services+=("ufw")
    fi
    
    # Проверка WireGuard порта
    if ! sudo ufw status | grep -q "51820/udp.*ALLOW"; then
        failed_services+=("wireguard-port")
    fi
    
    if [[ ${#failed_services[@]} -gt 0 ]]; then
        local services_list=$(printf '%s\n' "${failed_services[@]}" | tr '\n' ', ' | sed 's/,$//')
        local message="[WARNING] <b>Критические сервисы не работают</b><br><br>[ERROR] <b>Неактивные сервисы:</b><br>• $services_list<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

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
        local message="[DISK] <b>Мало места на диске</b><br><br>[STATS] <b>Использование диска:</b><br>• Использовано: ${usage}%<br>• Свободно: $free_space<br>• Порог: ${threshold}%<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

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
        local message="[UPDATE] <b>Доступны обновления системы</b><br><br>[PACKAGES] <b>Обновления:</b><br>• Количество пакетов: $updates_count<br>• Рекомендуется обновить систему<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

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
        local message="[ALERT] <b>Подозрительная активность SSH</b><br><br>[DETECTED] <b>Обнаружено:</b><br>• Неудачных попыток входа: $ssh_failures<br>• Возможная атака брутфорс<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

        send_telegram_message "$message"
        suspicious_found=true
    fi
    
    # Проверка новых процессов от root
    local root_processes=$(ps aux | grep -v grep | grep "^root" | wc -l)
    if [[ $root_processes -gt 150 ]]; then
        local message="[WARNING] <b>Много процессов от root</b><br><br>[DETECTED] <b>Обнаружено:</b><br>• Процессов от root: $root_processes<br>• Рекомендуется проверить<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

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
            local message="[SSL] <b>SSL сертификат скоро истечет</b><br><br>[CERT] <b>Сертификат:</b><br>• Дней до истечения: $days_until_expiry<br>• Дата истечения: $expiry_date<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

            send_telegram_message "$message"
            return 0
        fi
    fi
    
    return 1
}

# Проверка логов на критические ошибки
check_critical_logs() {
    # Исключаем ошибки от cloudflared и других неактуальных сервисов
    local critical_errors=$(grep -i "error\|critical\|emergency" /var/log/syslog | grep "$(date '+%b %d')" \
    | grep -v "cloudflared" \
    | grep -v "freshclam.*error code 429" \
    | grep -v "freshclam.*error code 403" \
    | grep -v "relay-security-audit.*Fatal error" \
    | grep -v "networkd-dispatcher.*Failed to get interface" \
    | grep -v "networkd-dispatcher.*Unknown interface index" \
    | grep -v "dockerd.*Error getting v2 registry" \
    | grep -v "dockerd.*Attempting next endpoint for pull after error" \
    | grep -v "dockerd.*denied: 403 Forbidden" \
    | grep -v "dockerd.*no such host" \
    | grep -v "dockerd.*ShouldRestart failed" \
    | grep -v "dockerd.*restart canceled" \
    | grep -v "dockerd.*Error handling plugin refcount operation" \
    | grep -v "dockerd.*legacy plugin: Unable to find plugin without name" \
    | grep -v "dockerd.*container will not be restarted" \
    | grep -v "dockerd.*hasBeenManuallyStopped=true" \
    | grep -v "dockerd.*daemonShuttingDown=false" \
    | grep -v "dockerd.*Error running exec" \
    | grep -v "dockerd.*stream copy error" \
    | grep -v "dockerd.*executable file not found" \
    | grep -v "dockerd.*unable to start container process" \
    | grep -v "containerd.*cleanup warnings" \
    | grep -v "containerd.*failed to remove runc container" \
    | grep -v "containerd.*runc did not terminate successfully" \
    | tail -5)
    
    if [[ -n "$critical_errors" ]]; then
        local error_count=$(echo "$critical_errors" | wc -l)
        local message="[ALERT] <b>Критические ошибки в логах</b><br><br>[ERRORS] <b>Ошибки:</b><br>• Количество: $error_count<br>• Последние ошибки:<br>$critical_errors<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

        if ! should_ignore_alert "$message"; then
            send_telegram_message "$message"
        else
            log_info "Сообщение не отправлено в Telegram (игнорируется по фильтру)"
        fi
        return 0
    fi
    
    return 1
}

# Проверка попыток доступа к портам в логах UFW
check_firewall_logs() {
    local ufw_log="/var/log/ufw.log"
    local last_alert_file="logs/last_ufw_block_alert.txt"
    local today=$(date '+%b %e')
    if [[ ! -f "$ufw_log" ]]; then
        log_warn "Лог UFW не найден ($ufw_log)"
        return 1
    fi
    # Найти последние 10 блокировок за сегодня (с применением фильтров)
    local blocked=$(grep "$today" "$ufw_log" | grep -E "UFW BLOCK|UFW LIMIT BLOCK")
    
    # Применяем фильтры
    if [[ "$FILTER_DHCP_BLOCKS" == "true" ]]; then
        blocked=$(echo "$blocked" | grep -v "SPT=67 DPT=68")
    fi
    if [[ "$FILTER_SSH_BLOCKS" == "true" ]]; then
        blocked=$(echo "$blocked" | grep -v "DPT=22")
    fi
    
    blocked=$(echo "$blocked" | tail -10)
    if [[ -z "$blocked" ]]; then
        log_info "Попыток доступа к портам (UFW BLOCK) не обнаружено"
        return 1
    fi
    # Проверка, были ли уже отправлены эти блокировки
    local last_alert=""
    if [[ -f "$last_alert_file" ]]; then
        last_alert=$(cat "$last_alert_file")
    fi
    local blocked_hash=$(echo "$blocked" | md5sum | awk '{print $1}')
    if [[ "$blocked_hash" != "$last_alert" ]]; then
        log_warn "Обнаружены попытки доступа к портам (UFW BLOCK):"
        echo "$blocked"
        
        # Создаем структурированное сообщение
        local message="[ALERT] <b>Обнаружены попытки доступа к портам (UFW BLOCK)</b>"
        message+="<br><br>[STATS] <b>Статистика:</b>"
        message+="<br>• Количество блокировок: $(echo "$blocked" | wc -l)"
        message+="<br>• Временной период: $(echo "$blocked" | head -1 | awk '{print $1, $2, $3}') - $(echo "$blocked" | tail -1 | awk '{print $1, $2, $3}')"
        
        message+="<br><br>[DETAILS] <b>Детали блокировок:</b>"
        
        # Парсим каждую блокировку
        while IFS= read -r line; do
            if [[ -n "$line" ]]; then
                # Извлекаем время
                local time=$(echo "$line" | awk '{print $1, $2, $3}')
                # Извлекаем IP источник
                local src_ip=$(echo "$line" | grep -o 'SRC=[^ ]*' | cut -d'=' -f2)
                # Извлекаем IP назначения
                local dst_ip=$(echo "$line" | grep -o 'DST=[^ ]*' | cut -d'=' -f2)
                # Извлекаем протокол
                local proto=$(echo "$line" | grep -o 'PROTO=[^ ]*' | cut -d'=' -f2)
                # Извлекаем порт источника
                local spt=$(echo "$line" | grep -o 'SPT=[^ ]*' | cut -d'=' -f2)
                # Извлекаем порт назначения
                local dpt=$(echo "$line" | grep -o 'DPT=[^ ]*' | cut -d'=' -f2)
                
                # Определяем тип блокировки
                local block_type=""
                if echo "$line" | grep -q "UFW LIMIT BLOCK"; then
                    block_type=" (Rate Limited)"
                fi
                
                message+="<br><br>[TIME] <b>Время:</b> $time"
                message+="<br>[CONNECTION] <b>Соединение:</b> $src_ip:$spt → $dst_ip:$dpt ($proto)$block_type"
                
                # Добавляем дополнительную информацию если есть
                if [[ -n "$spt" && -n "$dpt" ]]; then
                    local service_info=""
                    case "$dpt" in
                        22) service_info=" (SSH)" ;;
                        80) service_info=" (HTTP)" ;;
                        443) service_info=" (HTTPS)" ;;
                        5432) service_info=" (PostgreSQL)" ;;
                        8081|8082) service_info=" (Relay Server)" ;;
                        51820) service_info=" (WireGuard P2P)" ;;
                    esac
                    message+="$service_info"
                fi
            fi
        done <<< "$blocked"
        
        message+="<br><br>[CHECK_TIME] <b>Время проверки:</b> $(date '+%Y-%m-%d %H:%M:%S')"
        message+="<br>[SERVER] <b>Сервер:</b> $(hostname)"
        
        send_telegram_message "$message"
        echo "$blocked_hash" > "$last_alert_file"
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
    # Проверяем попытки доступа к портам (UFW BLOCK)
    log_info "Проверка логов firewall (UFW BLOCK)..."
    if check_firewall_logs; then
        log_info "Обнаружены попытки доступа к портам (UFW BLOCK)"
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