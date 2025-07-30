#!/bin/bash

# CloudBridge Relay - Enhanced Lynis Security Audit
# Расширенный аудит безопасности с дополнительными проверками

set -e

# Конфигурация
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"
STATE_DIR="/var/lib/lynis-telegram"
LOG_DIR="/var/log/lynis"
REPORT_DIR="/var/log/lynis/reports"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Функция отправки сообщений в Telegram
send_telegram_message() {
    local message="$1"
    if [[ -n "$TELEGRAM_BOT_TOKEN" && -n "$TELEGRAM_CHAT_ID" ]]; then
        curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
            -d "chat_id=${TELEGRAM_CHAT_ID}" \
            -d "text=${message}" \
            -d "parse_mode=HTML" > /dev/null
        log_info "Сообщение отправлено в Telegram"
    else
        log_warn "Telegram не настроен"
    fi
}

# Инициализация директорий
init_directories() {
    mkdir -p "$STATE_DIR" "$LOG_DIR" "$REPORT_DIR"
    chmod 750 "$LOG_DIR" "$REPORT_DIR"
}

# Проверка обновлений Lynis
check_lynis_updates() {
    log_info "Проверка обновлений Lynis..."
    local update_status=$(lynis --check-update 2>/dev/null | grep -c "Outdated" || echo "0")
    
    if [[ $update_status -gt 0 ]]; then
        local message="[LYNIS] <b>Доступно обновление Lynis</b><br><br>[UPDATE] <b>Информация:</b><br>• Рекомендуется обновить Lynis до последней версии<br>• Текущая версия может не содержать последние проверки безопасности<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"
        send_telegram_message "$message"
        return 1
    fi
    
    log_info "Lynis обновлен"
    return 0
}

# Полный аудит системы
run_full_audit() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local log_file="$LOG_DIR/lynis_full_${timestamp}.log"
    local report_file="$REPORT_DIR/lynis_report_${timestamp}.dat"
    
    log_info "Запуск полного аудита безопасности..."
    log_info "Лог файл: $log_file"
    log_info "Отчет: $report_file"
    
    # Запуск Lynis с расширенными опциями
    lynis audit system \
        --quick \
        --auditor "CloudBridge Relay" \
        --logfile "$log_file" \
        --report-file "$report_file" \
        --profile /etc/lynis/custom.prf \
        --nocolors
    
    log_info "Полный аудит завершен"
    return 0
}

# Аудит Docker (если установлен)
run_docker_audit() {
    if command -v docker &> /dev/null; then
        log_info "Запуск аудита Docker..."
        
        # Проверка Docker контейнеров
        local containers=$(docker ps -q 2>/dev/null | wc -l)
        local images=$(docker images -q 2>/dev/null | wc -l)
        
        # Проверка безопасности Docker
        local docker_security_issues=0
        
        # Проверка на запущенные контейнеры с привилегированным режимом
        local privileged_containers=$(docker ps --format "table {{.Names}}\t{{.Status}}" | grep -c "Up" || echo "0")
        
        # Проверка на контейнеры с root пользователем
        local root_containers=$(docker ps --format "table {{.Names}}\t{{.User}}" | grep -c "root" || echo "0")
        
        if [[ $privileged_containers -gt 0 || $root_containers -gt 0 ]]; then
            docker_security_issues=1
        fi
        
        if [[ $docker_security_issues -gt 0 ]]; then
            local message="[DOCKER] <b>Проблемы безопасности Docker</b><br><br>[STATS] <b>Статистика:</b><br>• Контейнеров: $containers<br>• Образов: $images<br>• Привилегированных контейнеров: $privileged_containers<br>• Контейнеров с root: $root_containers<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"
            send_telegram_message "$message"
        fi
        
        log_info "Аудит Docker завершен"
    else
        log_info "Docker не установлен, пропускаем аудит"
    fi
}

# Аудит сетевой безопасности
run_network_audit() {
    log_info "Запуск аудита сетевой безопасности..."
    
    local open_ports=$(ss -tuln | grep -c "LISTEN" || echo "0")
    local ssh_failures=$(grep "Failed password" /var/log/auth.log | grep "$(date '+%b %d')" | wc -l)
    local firewall_status=$(ufw status | grep -c "active" || echo "0")
    
    # Проверка на подозрительную активность
    local suspicious_activity=false
    local network_issues=""
    
    if [[ $ssh_failures -gt 50 ]]; then
        suspicious_activity=true
        network_issues+="• Много неудачных SSH попыток ($ssh_failures)<br>"
    fi
    
    if [[ $firewall_status -eq 0 ]]; then
        suspicious_activity=true
        network_issues+="• UFW не активен<br>"
    fi
    
    if [[ $open_ports -gt 20 ]]; then
        suspicious_activity=true
        network_issues+="• Много открытых портов ($open_ports)<br>"
    fi
    
    if [[ "$suspicious_activity" == "true" ]]; then
        local message="[NETWORK] <b>Проблемы сетевой безопасности</b><br><br>[ISSUES] <b>Обнаружено:</b><br>$network_issues<br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"
        send_telegram_message "$message"
    fi
    
    log_info "Аудит сети завершен"
}

# Анализ результатов
analyze_results() {
    local report_file="$1"
    
    if [[ ! -f "$report_file" ]]; then
        log_error "Файл отчета не найден: $report_file"
        return 1
    fi
    
    # Извлекаем основные метрики
    local score=$(grep "hardening_index" "$report_file" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    local warnings=$(grep "warnings" "$report_file" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    local suggestions=$(grep "suggestions" "$report_file" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    
    # Проверяем критические проблемы
    local critical_issues=$(grep -c "Suggestion.*CRITICAL" "$report_file" || echo "0")
    
    log_info "Результаты анализа:"
    log_info "• Hardening Index: $score"
    log_info "• Предупреждения: $warnings"
    log_info "• Рекомендации: $suggestions"
    log_info "• Критические проблемы: $critical_issues"
    
    # Отправляем уведомление если есть критические проблемы
    if [[ $critical_issues -gt 0 ]]; then
        local critical_details=$(grep "Suggestion.*CRITICAL" "$report_file" | head -3 | sed 's/^/• /')
        local message="[LYNIS] <b>Критические проблемы безопасности</b><br><br>[CRITICAL] <b>Обнаружено:</b><br>$critical_details<br><br>[SCORE] <b>Hardening Index:</b> $score<br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"
        send_telegram_message "$message"
    fi
    
    return 0
}

# Очистка старых логов
cleanup_old_logs() {
    log_info "Очистка старых логов..."
    
    # Удаляем логи старше 30 дней
    find "$LOG_DIR" -name "lynis_*.log" -mtime +30 -delete 2>/dev/null || true
    find "$REPORT_DIR" -name "lynis_report_*.dat" -mtime +30 -delete 2>/dev/null || true
    
    log_info "Очистка завершена"
}

# Основная функция
main() {
    log_info "Запуск расширенного аудита безопасности Lynis..."
    
    # Инициализация
    init_directories
    
    # Проверка обновлений
    check_lynis_updates
    
    # Полный аудит
    run_full_audit
    
    # Дополнительные аудиты
    run_docker_audit
    run_network_audit
    
    # Анализ результатов
    local latest_report=$(ls -t "$REPORT_DIR"/lynis_report_*.dat 2>/dev/null | head -1)
    if [[ -n "$latest_report" ]]; then
        analyze_results "$latest_report"
    fi
    
    # Очистка
    cleanup_old_logs
    
    log_info "Расширенный аудит завершен"
}

# Обработка аргументов
case "${1:-}" in
    --test)
        log_info "Тестовый режим"
        send_telegram_message "[LYNIS] <b>Тест расширенного аудита</b><br><br>[TEST] <b>Информация:</b><br>• Система расширенного аудита работает<br>• Готов к выполнению полного аудита<br><br>[TIME] <b>Время:</b> $(date '+%Y-%m-%d %H:%M:%S')<br>[SERVER] <b>Сервер:</b> $(hostname)"
        ;;
    --help|-h)
        echo "Использование: $0 [опции]"
        echo "Опции:"
        echo "  --test     Тестовый режим"
        echo "  --help     Показать эту справку"
        echo ""
        echo "Без опций запускает полный расширенный аудит"
        ;;
    *)
        main
        ;;
esac 