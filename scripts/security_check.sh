#!/bin/bash

# CloudBridge Relay - Security Check Script
# Быстрая проверка основных аспектов безопасности

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

log_header() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Проверка UFW
check_ufw() {
    log_header "Проверка UFW Firewall"
    
    if sudo systemctl is-active --quiet ufw; then
        log_info "UFW активен"
        echo "Статус UFW:"
        sudo ufw status | head -10
    else
        log_error "UFW не активен"
    fi
    echo
}

# Проверка Fail2Ban
check_fail2ban() {
    log_header "Проверка Fail2Ban"
    
    if sudo systemctl is-active --quiet fail2ban; then
        log_info "Fail2Ban активен"
        echo "Активные jail'ы:"
        sudo fail2ban-client status
    else
        log_error "Fail2Ban не активен"
    fi
    echo
}

# Проверка открытых портов
check_open_ports() {
    log_header "Проверка открытых портов"
    
    echo "Критические порты:"
    sudo netstat -tlnp | grep LISTEN | grep -E "(22|8082|80|443)" || log_warn "Критические порты не найдены"
    echo
}

# Проверка SSL сертификатов
check_ssl_certs() {
    log_header "Проверка SSL сертификатов"
    
    if [[ -f "certs/fullchain.pem" && -f "certs/privkey.pem" ]]; then
        log_info "SSL сертификаты найдены"
        ls -la certs/*.pem
    else
        log_warn "SSL сертификаты не найдены"
    fi
    echo
}

# Проверка автоматических обновлений
check_auto_updates() {
    log_header "Проверка автоматических обновлений"
    
    if sudo systemctl is-active --quiet unattended-upgrades; then
        log_info "Автоматические обновления активны"
    else
        log_warn "Автоматические обновления не активны"
    fi
    echo
}

# Проверка логирования
check_logging() {
    log_header "Проверка системы логирования"
    
    if sudo systemctl is-active --quiet rsyslog; then
        log_info "Rsyslog активен"
    else
        log_warn "Rsyslog не активен"
    fi
    echo
}

# Проверка антивируса
check_antivirus() {
    log_header "Проверка антивируса"
    
    if sudo systemctl is-active --quiet clamav-freshclam 2>/dev/null; then
        log_info "ClamAV активен"
    else
        log_warn "ClamAV не установлен или не активен"
    fi
    echo
}

# Проверка системных сервисов
check_system_services() {
    log_header "Проверка системных сервисов"
    
    local failed_services=$(sudo systemctl list-units --failed --no-pager | grep -c "failed")
    if [[ $failed_services -eq 0 ]]; then
        log_info "Нет неудачных сервисов"
    else
        log_warn "Найдено $failed_services неудачных сервисов"
        sudo systemctl list-units --failed --no-pager
    fi
    echo
}

# Проверка прав доступа к файлам
check_file_permissions() {
    log_header "Проверка прав доступа к критическим файлам"
    
    # Проверка SSH конфигурации
    if [[ -f /etc/ssh/sshd_config ]]; then
        local ssh_perms=$(sudo stat -c %a /etc/ssh/sshd_config)
        if [[ $ssh_perms == "644" ]]; then
            log_info "SSH конфигурация: правильные права ($ssh_perms)"
        else
            log_warn "SSH конфигурация: неправильные права ($ssh_perms)"
        fi
    fi
    
    # Проверка сертификатов
    if [[ -f certs/privkey.pem ]]; then
        local cert_perms=$(stat -c %a certs/privkey.pem)
        if [[ $cert_perms == "600" ]]; then
            log_info "Приватный ключ: правильные права ($cert_perms)"
        else
            log_warn "Приватный ключ: неправильные права ($cert_perms)"
        fi
    fi
    echo
}

# Проверка пользователей
check_users() {
    log_header "Проверка пользователей системы"
    
    echo "Пользователи с shell доступом:"
    sudo grep -E ":/bin/(bash|sh)$" /etc/passwd | cut -d: -f1
    echo
    
    echo "Пользователи с sudo правами:"
    sudo grep -Po '^sudo.+:\K.*$' /etc/group | tr ',' '\n' | grep -v "^$" || echo "Нет пользователей с sudo правами"
    echo
}

# Рекомендации по безопасности
show_recommendations() {
    log_header "Рекомендации по безопасности"
    
    echo "1. Настройте регулярные резервные копии"
    echo "2. Установите и настройте ClamAV для антивирусной защиты"
    echo "3. Настройте мониторинг логов"
    echo "4. Регулярно обновляйте систему"
    echo "5. Настройте ротацию логов"
    echo "6. Рассмотрите использование AppArmor или SELinux"
    echo "7. Настройте уведомления о безопасности"
    echo "8. Проводите регулярные аудиты безопасности"
    echo
}

# Основная функция
main() {
    echo "=========================================="
    echo "  CloudBridge Relay - Security Check"
    echo "=========================================="
    echo
    
    check_ufw
    check_fail2ban
    check_open_ports
    check_ssl_certs
    check_auto_updates
    check_logging
    check_antivirus
    check_system_services
    check_file_permissions
    check_users
    
    echo "=========================================="
    show_recommendations
    echo "Проверка завершена!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Security Check Script"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --quick        Быстрая проверка (только критические компоненты)"
        echo
        echo "Примеры:"
        echo "  $0              Полная проверка безопасности"
        echo "  $0 --quick      Быстрая проверка"
        ;;
    --quick)
        echo "Быстрая проверка безопасности..."
        check_ufw
        check_fail2ban
        check_open_ports
        check_ssl_certs
        ;;
    *)
        main
        ;;
esac 