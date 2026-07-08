#!/bin/bash

# CloudBridge Relay - Enhanced Security Monitor
# Улучшенный мониторинг безопасности с конфигурационным файлом

set -e

# Загрузка конфигурации
CONFIG_FILE="/etc/cloudbridge/security_monitor.conf"

# Функция загрузки конфигурации
load_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "Конфигурационный файл не найден. Создаю файл по умолчанию..."
        bash "$(dirname "$0")/security_monitor_config.sh" --create
    fi
    
    # Загружаем конфигурацию
    source "$CONFIG_FILE"
}

# Загружаем конфигурацию
load_config

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

# Улучшенная фильтрация сообщений
should_ignore_alert() {
    local message="$1"
    
    # Создаем список всех фильтров для проверки
    local filters=(
        "$EMERGENCY_SERVICE_FILTER"
        "$ICMP_IGNORE_FILTER"
        "$AIDE_INSTALL_FILTER"
        "$DHCP_BLOCK_FILTER"
        "$AIDE_DATABASE_FILTER"
        "$DOCKER_SHOULDRESTART_FILTER"
        "$DOCKER_RESTART_CANCELED_FILTER"
        "$DOCKER_PLUGIN_ERROR_FILTER"
        "$DOCKER_LEGACY_PLUGIN_FILTER"
        "$DOCKER_CONTAINER_RESTART_FILTER"
        "$DOCKER_MANUALLY_STOPPED_FILTER"
        "$DOCKER_DAEMON_SHUTDOWN_FILTER"
        "$DOCKER_EXEC_ERROR_FILTER"
        "$DOCKER_STREAM_COPY_ERROR_FILTER"
        "$DOCKER_EXECUTABLE_NOT_FOUND_FILTER"
        "$DOCKER_CONTAINER_PROCESS_ERROR_FILTER"
        "$CONTAINERD_CLEANUP_WARNINGS_FILTER"
        "$CONTAINERD_RUNC_REMOVE_ERROR_FILTER"
        "$CONTAINERD_RUNC_TERMINATE_ERROR_FILTER"
    )
    
    # Проверяем каждый фильтр
    for filter in "${filters[@]}"; do
        if [[ -n "$filter" && "$message" =~ $filter ]]; then
            log_info "Сообщение отфильтровано по паттерну: $filter"
            return 0  # игнорировать
        fi
    done
    
    return 1  # не игнорировать
}

# Проверка новых блокировок Fail2Ban
check_fail2ban_blocks() {
    if [[ "$ENABLE_FAIL2BAN_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local current_blocks=$(sudo fail2ban-client status | grep "Total banned" | awk '{print $4}' | tr '\n' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    local last_blocks_file="$STATE_DIR/last_fail2ban_blocks.txt"
    
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
    if [[ "$ENABLE_CRITICAL_SERVICES_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local failed_services=()
    
    # Проверяем критически важные сервисы
    local critical_services=("fail2ban" "sshd" "rsyslog")
    
    for service in "${critical_services[@]}"; do
        if ! systemctl is-active --quiet "$service"; then
            failed_services+=("$service")
        fi
    done
    
    # Специальная проверка UFW
    if ! sudo ufw status | grep -q "Status: active"; then
        failed_services+=("ufw")
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
    if [[ "$ENABLE_DISK_SPACE_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local usage=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [[ $usage -gt $DISK_USAGE_THRESHOLD ]]; then
        # Файл для отслеживания последнего уведомления о диске
        local last_disk_file="$STATE_DIR/last_disk_space_alert.txt"
        local current_time=$(date +%s)
        local last_alert_time=0
        
        # Читаем время последнего уведомления
        if [[ -f "$last_disk_file" ]]; then
            last_alert_time=$(cat "$last_disk_file" 2>/dev/null || echo "0")
        fi
        
        # Проверяем, прошло ли достаточно времени с последнего уведомления
        local time_diff=$((current_time - last_alert_time))
        local min_interval=${DISK_SPACE_ALERT_INTERVAL:-21600}  # 6 часов по умолчанию
        
        if [[ $time_diff -gt $min_interval ]]; then
            local free_space=$(df -h / | awk 'NR==2 {print $4}')
            local message="[DISK] <b>Мало места на диске</b><br><br>[STATS] <b>Использование диска:</b><br>• Использовано: ${usage}%<br>• Свободно: $free_space<br>• Порог: ${DISK_USAGE_THRESHOLD}%<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

            send_telegram_message "$message"
            
            # Сохраняем время отправки уведомления
            echo "$current_time" > "$last_disk_file"
            return 0
        else
            log_info "Уведомление о дисковом пространстве уже отправлено недавно (${time_diff} сек назад)"
            return 1
        fi
    fi
    
    return 1
}

# Проверка обновлений системы
check_system_updates() {
    if [[ "$ENABLE_SYSTEM_UPDATES_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local updates_count=$(apt list --upgradable 2>/dev/null | grep -c upgradable || echo "0")
    updates_count=$(echo "$updates_count" | tr -d '[:space:]' | sed 's/^0*//' || echo "0")
    if [[ -z "$updates_count" ]]; then
        updates_count="0"
    fi
    
    if [[ "$updates_count" =~ ^[0-9]+$ && $updates_count -gt $SYSTEM_UPDATES_THRESHOLD ]]; then
        # Файл для отслеживания последнего уведомления об обновлениях
        local last_update_file="$STATE_DIR/last_system_update_alert.txt"
        local current_time=$(date +%s)
        local last_alert_time=0
        
        # Читаем время последнего уведомления
        if [[ -f "$last_update_file" ]]; then
            last_alert_time=$(cat "$last_update_file" 2>/dev/null || echo "0")
        fi
        
        # Проверяем, прошло ли достаточно времени с последнего уведомления
        local time_diff=$((current_time - last_alert_time))
        local min_interval=${SYSTEM_UPDATES_ALERT_INTERVAL:-86400}  # 24 часа по умолчанию
        
        if [[ $time_diff -gt $min_interval ]]; then
            # Получаем список пакетов для обновления с версиями
            local packages_info=""
            local packages_output=$(apt list --upgradable 2>/dev/null | grep upgradable | head -5)
            
            if [[ -n "$packages_output" ]]; then
                packages_info="<br>• Обновляемые пакеты:<br>"
                while IFS= read -r line; do
                    if [[ -n "$line" ]]; then
                        local package_name=$(echo "$line" | cut -d'/' -f1)
                        # Извлекаем версии более точно
                        local current_version=$(echo "$line" | sed 's/.*\[upgradable from: \(.*\)\]/\1/' | cut -d' ' -f1)
                        local new_version=$(echo "$line" | awk '{print $2}')
                        packages_info="$packages_info  - $package_name: $current_version → $new_version<br>"
                    fi
                done <<< "$packages_output"
                
                if [[ $updates_count -gt 5 ]]; then
                    packages_info="$packages_info  - ... и еще $((updates_count - 5)) пакетов<br>"
                fi
            fi
            
            local message="[UPDATE] <b>Доступны обновления системы</b><br><br>[PACKAGES] <b>Обновления:</b><br>• Количество пакетов: $updates_count$packages_info<br>• Рекомендуется обновить систему<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

            send_telegram_message "$message"
            
            # Сохраняем время отправки уведомления
            echo "$current_time" > "$last_update_file"
            return 0
        else
            log_info "Уведомление об обновлениях уже отправлено недавно (${time_diff} сек назад)"
            return 1
        fi
    fi
    
    return 1
}

# Проверка подозрительной активности
check_suspicious_activity() {
    if [[ "$ENABLE_SUSPICIOUS_ACTIVITY_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local suspicious_found=false
    local current_time=$(date +%s)
    
    # Проверка множественных неудачных SSH попыток
    local ssh_failures=$(grep "Failed password" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
    
    if [[ $ssh_failures -gt $SSH_FAILURES_THRESHOLD ]]; then
        # Файл для отслеживания последнего уведомления о SSH
        local last_ssh_file="$STATE_DIR/last_ssh_activity_alert.txt"
        local last_alert_time=0
        
        # Читаем время последнего уведомления
        if [[ -f "$last_ssh_file" ]]; then
            last_alert_time=$(cat "$last_ssh_file" 2>/dev/null || echo "0")
        fi
        
        # Проверяем, прошло ли достаточно времени с последнего уведомления
        local time_diff=$((current_time - last_alert_time))
        local min_interval=${SSH_ACTIVITY_ALERT_INTERVAL:-3600}  # 1 час по умолчанию
        
        if [[ $time_diff -gt $min_interval ]]; then
            local message="[ALERT] <b>Подозрительная активность SSH</b><br><br>[DETECTED] <b>Обнаружено:</b><br>• Неудачных попыток входа: $ssh_failures<br>• Возможная атака брутфорс<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

            send_telegram_message "$message"
            
            # Сохраняем время отправки уведомления
            echo "$current_time" > "$last_ssh_file"
            suspicious_found=true
        else
            log_info "Уведомление о SSH активности уже отправлено недавно (${time_diff} сек назад)"
        fi
    fi
    
    # Проверка новых процессов от root
    local root_processes=$(ps aux | grep -v grep | grep "^root" | wc -l)
    if [[ $root_processes -gt $ROOT_PROCESSES_THRESHOLD ]]; then
        # Файл для отслеживания последнего уведомления о процессах root
        local last_root_file="$STATE_DIR/last_root_processes_alert.txt"
        local last_alert_time=0
        
        # Читаем время последнего уведомления
        if [[ -f "$last_root_file" ]]; then
            last_alert_time=$(cat "$last_root_file" 2>/dev/null || echo "0")
        fi
        
        # Проверяем, прошло ли достаточно времени с последнего уведомления
        local time_diff=$((current_time - last_alert_time))
        local min_interval=${ROOT_PROCESSES_ALERT_INTERVAL:-7200}  # 2 часа по умолчанию
        
        if [[ $time_diff -gt $min_interval ]]; then
            local message="[WARNING] <b>Много процессов от root</b><br><br>[DETECTED] <b>Обнаружено:</b><br>• Процессов от root: $root_processes<br>• Рекомендуется проверить<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

            send_telegram_message "$message"
            
            # Сохраняем время отправки уведомления
            echo "$current_time" > "$last_root_file"
            suspicious_found=true
        else
            log_info "Уведомление о процессах root уже отправлено недавно (${time_diff} сек назад)"
        fi
    fi
    
    if [[ "$suspicious_found" == "true" ]]; then
        return 0
    fi
    
    return 1
}

# Проверка SSL сертификатов
check_ssl_certificates() {
    if [[ "$ENABLE_SSL_CERTIFICATE_MONITORING" != "true" ]]; then
        return 1
    fi
    
    if [[ -f "certs/fullchain.pem" ]]; then
        local expiry_date=$(openssl x509 -enddate -noout -in certs/fullchain.pem | cut -d= -f2)
        local expiry_timestamp=$(date -d "$expiry_date" +%s)
        local current_timestamp=$(date +%s)
        local days_until_expiry=$(( (expiry_timestamp - current_timestamp) / 86400 ))
        
        if [[ $days_until_expiry -lt 30 ]]; then
            # Файл для отслеживания последнего уведомления об SSL
            local last_ssl_file="$STATE_DIR/last_ssl_cert_alert.txt"
            local current_time=$(date +%s)
            local last_alert_time=0
            
            # Читаем время последнего уведомления
            if [[ -f "$last_ssl_file" ]]; then
                last_alert_time=$(cat "$last_ssl_file" 2>/dev/null || echo "0")
            fi
            
            # Проверяем, прошло ли достаточно времени с последнего уведомления
            local time_diff=$((current_time - last_alert_time))
            local min_interval=${SSL_CERT_ALERT_INTERVAL:-43200}  # 12 часов по умолчанию
            
            if [[ $time_diff -gt $min_interval ]]; then
                local message="[SSL] <b>SSL сертификат скоро истечет</b><br><br>[CERT] <b>Сертификат:</b><br>• Дней до истечения: $days_until_expiry<br>• Дата истечения: $expiry_date<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"

                send_telegram_message "$message"
                
                # Сохраняем время отправки уведомления
                echo "$current_time" > "$last_ssl_file"
                return 0
            else
                log_info "Уведомление об SSL сертификате уже отправлено недавно (${time_diff} сек назад)"
                return 1
            fi
        fi
    fi
    
    return 1
}

# Улучшенная проверка логов на критические ошибки
check_critical_logs() {
    if [[ "$ENABLE_CRITICAL_LOGS_MONITORING" != "true" ]]; then
        return 1
    fi
    
    # Создаем команду grep с динамическими фильтрами
    local grep_command="grep -i \"error\|critical\|emergency\" /var/log/syslog | grep \"$(date '+%b %d')\""
    
    # Добавляем фильтры исключения из конфигурации
    local filters=(
        "$CLOUDFLARED_FILTER"
        "$FRESHCLAM_429_FILTER"
        "$FRESHCLAM_403_FILTER"
        "$RELAY_SECURITY_AUDIT_FILTER"
        "$NETWORKD_DISPATCHER_FILTER"
        "$NETWORKD_UNKNOWN_INTERFACE_FILTER"
        "$DOCKER_REGISTRY_ERROR_FILTER"
        "$DOCKER_ENDPOINT_ERROR_FILTER"
        "$DOCKER_FORBIDDEN_FILTER"
        "$DOCKER_NO_HOST_FILTER"
        "$DOCKER_SHOULDRESTART_FILTER"
        "$DOCKER_RESTART_CANCELED_FILTER"
        "$DOCKER_PLUGIN_ERROR_FILTER"
        "$DOCKER_LEGACY_PLUGIN_FILTER"
        "$DOCKER_CONTAINER_RESTART_FILTER"
        "$DOCKER_MANUALLY_STOPPED_FILTER"
        "$DOCKER_DAEMON_SHUTDOWN_FILTER"
        "$DOCKER_EXEC_ERROR_FILTER"
        "$DOCKER_STREAM_COPY_ERROR_FILTER"
        "$DOCKER_EXECUTABLE_NOT_FOUND_FILTER"
        "$DOCKER_CONTAINER_PROCESS_ERROR_FILTER"
        "$CONTAINERD_CLEANUP_WARNINGS_FILTER"
        "$CONTAINERD_RUNC_REMOVE_ERROR_FILTER"
        "$CONTAINERD_RUNC_TERMINATE_ERROR_FILTER"
    )
    
    # Добавляем каждый фильтр в команду grep
    for filter in "${filters[@]}"; do
        if [[ -n "$filter" ]]; then
            grep_command+=" | grep -v \"$filter\""
        fi
    done
    
    # Выполняем команду и получаем результат
    local critical_errors=$(eval "$grep_command | tail -5")
    
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
    if [[ "$ENABLE_FIREWALL_LOGS_MONITORING" != "true" ]]; then
        return 1
    fi
    
    local ufw_log="/var/log/ufw.log"
    local last_alert_file="$STATE_DIR/last_ufw_block_alert.txt"
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
    log_info "Запуск улучшенной проверки безопасности..."
    
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
    log_info "Запуск непрерывного улучшенного мониторинга безопасности"
    send_telegram_message "<b>Улучшенный мониторинг безопасности запущен</b>
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
 Сервер: $(hostname)
⚙️ Конфигурация: $CONFIG_FILE"
    
    while true; do
        run_security_check
        sleep $CHECK_INTERVAL
    done
}

# Функция разовой проверки
single_check() {
    log_info "Выполнение разовой улучшенной проверки безопасности"
    run_security_check
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Enhanced Security Monitor"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --once         Выполнить разовую проверку"
        echo "  --continuous   Запустить непрерывный мониторинг"
        echo "  --test         Тестовая отправка в Telegram"
        echo "  --config       Показать текущую конфигурацию"
        echo
        echo "Примеры:"
        echo "  $0 --once              Разовая проверка"
        echo "  $0 --continuous        Непрерывный мониторинг"
        echo "  $0 --test              Тестовая отправка"
        echo "  $0 --config            Показать конфигурацию"
        ;;
    --once)
        single_check
        ;;
    --continuous)
        continuous_monitoring
        ;;
    --test)
        send_telegram_message "🧪 <b>Тест улучшенного мониторинга безопасности</b>
Система мониторинга работает
🕐 Время: $(date '+%Y-%m-%d %H:%M:%S')
 Сервер: $(hostname)
⚙️ Конфигурация: $CONFIG_FILE"
        ;;
    --config)
        bash "$(dirname "$0")/security_monitor_config.sh" --show
        ;;
    *)
        single_check
        ;;
esac 