#!/bin/bash

# CloudBridge Relay - Security Monitor Setup
# Установка и настройка системы мониторинга безопасности с Telegram уведомлениями

set -e

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

# Проверка root прав
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root"
        exit 1
    fi
}

# Установка зависимостей
install_dependencies() {
    log_info "Установка зависимостей..."
    
    # Обновление пакетов
    apt-get update
    
    # Установка необходимых пакетов
    apt-get install -y jq curl fail2ban
    
    log_info "Зависимости установлены"
}

# Настройка fail2ban
setup_fail2ban() {
    log_info "Настройка fail2ban..."
    
    # Проверяем, установлен ли fail2ban
    if ! command -v fail2ban-client &> /dev/null; then
        log_warn "fail2ban не установлен, пропускаем настройку"
        return
    fi
    
    # Создаем конфигурацию для SSH
    cat > /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log
maxretry = 3
bantime = 3600
EOF
    
    # Перезапускаем fail2ban
    systemctl restart fail2ban
    systemctl enable fail2ban
    
    log_info "fail2ban настроен и запущен"
}

# Создание директорий
create_directories() {
    log_info "Создание директорий..."
    
    mkdir -p /var/lib/security-monitor
    mkdir -p /var/lib/lynis-telegram
    mkdir -p /home/ubuntu/cloudbridge-relay-installer/logs
    
    # Устанавливаем права доступа
    chown root:root /var/lib/security-monitor
    chmod 700 /var/lib/security-monitor
    
    chown root:root /var/lib/lynis-telegram
    chmod 700 /var/lib/lynis-telegram
    
    log_info "Директории созданы"
}

# Установка systemd сервисов
install_services() {
    log_info "Установка systemd сервисов..."
    
    # Копируем файлы сервисов
    cp systemd/security-monitor.service /etc/systemd/system/
    cp systemd/lynis-security-audit.service /etc/systemd/system/
    cp systemd/lynis-security-audit.timer /etc/systemd/system/
    
    # Перезагружаем systemd
    systemctl daemon-reload
    
    # Включаем сервисы
    systemctl enable security-monitor
    systemctl enable lynis-security-audit.timer
    
    log_info "Systemd сервисы установлены"
}

# Настройка Telegram
setup_telegram() {
    log_info "Настройка Telegram уведомлений..."
    
    echo "Для настройки Telegram уведомлений:"
    echo "1. Создайте бота через @BotFather"
    echo "2. Получите токен бота"
    echo "3. Получите Chat ID"
    echo
    echo "Затем отредактируйте файлы:"
    echo "- scripts/lynis_with_telegram.sh"
    echo "- scripts/security_monitor_telegram.sh"
    echo
    echo "И укажите в них:"
    echo "TELEGRAM_BOT_TOKEN=\"ваш_токен_бота\""
    echo "TELEGRAM_CHAT_ID=\"ваш_chat_id\""
    echo
    read -p "Нажмите Enter для продолжения..."
}

# Тестирование системы
test_system() {
    log_info "Тестирование системы мониторинга..."
    
    # Тест Lynis
    if [[ -f "scripts/lynis_with_telegram.sh" ]]; then
        log_info "Тестирование Lynis мониторинга..."
        ./scripts/lynis_with_telegram.sh --test
    fi
    
    # Тест Security Monitor
    if [[ -f "scripts/security_monitor_telegram.sh" ]]; then
        log_info "Тестирование Security Monitor..."
        ./scripts/security_monitor_telegram.sh --test
    fi
    
    log_info "Тестирование завершено"
}

# Показать статус
show_status() {
    log_info "Статус системы мониторинга:"
    echo "================================"
    
    # Статус сервисов
    echo "Systemd сервисы:"
    systemctl is-active security-monitor 2>/dev/null && echo "✅ security-monitor: активен" || echo "❌ security-monitor: неактивен"
    systemctl is-active lynis-security-audit.timer 2>/dev/null && echo "✅ lynis-security-audit.timer: активен" || echo "❌ lynis-security-audit.timer: неактивен"
    
    echo
    echo "Зависимости:"
    command -v jq &> /dev/null && echo "✅ jq: установлен" || echo "❌ jq: не установлен"
    command -v curl &> /dev/null && echo "✅ curl: установлен" || echo "❌ curl: не установлен"
    command -v fail2ban-client &> /dev/null && echo "✅ fail2ban: установлен" || echo "❌ fail2ban: не установлен"
    
    echo
    echo "Файлы состояния:"
    [[ -d "/var/lib/security-monitor" ]] && echo "✅ /var/lib/security-monitor: существует" || echo "❌ /var/lib/security-monitor: не существует"
    [[ -d "/var/lib/lynis-telegram" ]] && echo "✅ /var/lib/lynis-telegram: существует" || echo "❌ /var/lib/lynis-telegram: не существует"
    
    echo
    echo "Логи:"
    [[ -f "/var/lib/security-monitor/events.log" ]] && echo "✅ events.log: существует" || echo "❌ events.log: не существует"
    [[ -f "/var/lib/lynis-telegram/changes.log" ]] && echo "✅ changes.log: существует" || echo "❌ changes.log: не существует"
}

# Запуск сервисов
start_services() {
    log_info "Запуск сервисов..."
    
    systemctl start security-monitor
    systemctl start lynis-security-audit.timer
    
    log_info "Сервисы запущены"
}

# Основная функция
main() {
    log_info "Установка системы мониторинга безопасности"
    
    check_root
    install_dependencies
    setup_fail2ban
    create_directories
    install_services
    setup_telegram
    test_system
    start_services
    show_status
    
    log_info "Установка завершена!"
    echo
    echo "Для управления сервисами используйте:"
    echo "sudo systemctl start/stop/restart security-monitor"
    echo "sudo systemctl start/stop/restart lynis-security-audit.timer"
    echo
    echo "Для просмотра логов:"
    echo "sudo journalctl -u security-monitor -f"
    echo "sudo journalctl -u lynis-security-audit -f"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Security Monitor Setup"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --status       Показать статус системы"
        echo "  --test         Только тестирование"
        echo "  --start        Только запуск сервисов"
        echo
        echo "Примеры:"
        echo "  sudo $0              Полная установка"
        echo "  sudo $0 --status     Показать статус"
        echo "  sudo $0 --test       Только тестирование"
        echo "  sudo $0 --start      Запуск сервисов"
        ;;
    --status)
        show_status
        ;;
    --test)
        test_system
        ;;
    --start)
        start_services
        show_status
        ;;
    *)
        main
        ;;
esac 