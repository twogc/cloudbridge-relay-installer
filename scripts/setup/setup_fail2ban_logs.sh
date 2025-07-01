#!/bin/bash

# CloudBridge Relay - Fail2Ban Logs Setup Script
# Этот скрипт настраивает символические ссылки для лог-файлов fail2ban

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Проверка root прав
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root"
        exit 1
    fi
}

# Создание директории для логов relay
create_log_directory() {
    log_info "Создание директории для логов relay..."
    
    if [[ ! -d /var/log/relay ]]; then
        mkdir -p /var/log/relay
        log_info "Директория /var/log/relay создана"
    else
        log_info "Директория /var/log/relay уже существует"
    fi
}

# Создание символической ссылки
create_symlink() {
    log_info "Создание символической ссылки для лог-файла relay..."
    
    local source_log="/home/ubuntu/cloudbridge-relay-installer/logs/relay.log"
    local target_log="/var/log/relay/relay.log"
    
    # Проверяем, существует ли исходный файл
    if [[ ! -f "$source_log" ]]; then
        log_warn "Исходный лог-файл $source_log не найден"
        log_info "Создаем пустой файл..."
        touch "$source_log"
        chown ubuntu:ubuntu "$source_log"
    fi
    
    # Создаем символическую ссылку
    if [[ -L "$target_log" ]]; then
        log_info "Удаляем существующую символическую ссылку..."
        rm "$target_log"
    fi
    
    ln -sf "$source_log" "$target_log"
    log_info "Символическая ссылка создана: $target_log -> $source_log"
}

# Настройка прав доступа
setup_permissions() {
    log_info "Настройка прав доступа..."
    
    # Убеждаемся, что fail2ban может читать лог-файл
    chmod 644 /home/ubuntu/cloudbridge-relay-installer/logs/relay.log
    chown ubuntu:ubuntu /home/ubuntu/cloudbridge-relay-installer/logs/relay.log
    
    # Права на директорию
    chmod 755 /var/log/relay
    chown root:root /var/log/relay
}

# Перезапуск fail2ban
restart_fail2ban() {
    log_info "Перезапуск fail2ban..."
    
    if systemctl is-active --quiet fail2ban; then
        systemctl restart fail2ban
        log_info "fail2ban перезапущен"
    else
        log_warn "fail2ban не запущен, запускаем..."
        systemctl start fail2ban
        systemctl enable fail2ban
        log_info "fail2ban запущен и включен"
    fi
}

# Проверка статуса
check_status() {
    log_info "Проверка статуса fail2ban..."
    
    if systemctl is-active --quiet fail2ban; then
        log_info "fail2ban активен"
        echo
        echo "=== Статус fail2ban ==="
        fail2ban-client status
        echo
        echo "=== Статус relay jail ==="
        fail2ban-client status relay
    else
        log_error "fail2ban не активен"
        exit 1
    fi
}

# Основная функция
main() {
    log_info "Начало настройки лог-файлов fail2ban для CloudBridge Relay"
    
    # Проверка root прав
    check_root
    
    # Создание директории
    create_log_directory
    
    # Создание символической ссылки
    create_symlink
    
    # Настройка прав доступа
    setup_permissions
    
    # Перезапуск fail2ban
    restart_fail2ban
    
    # Проверка статуса
    check_status
    
    log_info "Настройка лог-файлов fail2ban завершена!"
    echo
    echo "Полезные команды:"
    echo "  sudo fail2ban-client status          - Статус всех jail'ов"
    echo "  sudo fail2ban-client status relay    - Статус relay jail"
    echo "  sudo fail2ban-client set relay unbanip <IP> - Разблокировать IP"
    echo "  sudo tail -f /var/log/fail2ban.log   - Логи fail2ban"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Fail2Ban Logs Setup Script"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --check        Только проверить статус"
        echo
        echo "Примеры:"
        echo "  sudo $0              Настроить лог-файлы"
        echo "  sudo $0 --check      Проверить статус"
        ;;
    --check)
        check_status
        ;;
    *)
        main
        ;;
esac 