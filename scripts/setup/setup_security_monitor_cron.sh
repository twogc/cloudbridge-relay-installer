#!/bin/bash

# CloudBridge Relay - Security Monitor Cron Setup
# Альтернативная установка мониторинга безопасности через cron

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

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Проверка root прав
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root"
        exit 1
    fi
}

# Установка cron задачи
install_cron() {
    log_info "Установка cron задачи для мониторинга безопасности..."
    
    # Создаем временный файл с cron задачей
    local cron_job="*/5 * * * * cd /home/ubuntu/cloudbridge-relay-installer && /usr/bin/bash scripts/security/security_monitor.sh --once >> logs/security-monitor-cron.log 2>&1"
    
    # Добавляем задачу в crontab пользователя ubuntu
    (crontab -u ubuntu -l 2>/dev/null; echo "$cron_job") | crontab -u ubuntu -
    
    log_success "Cron задача установлена"
}

# Удаление cron задачи
uninstall_cron() {
    log_info "Удаление cron задачи..."
    
    # Удаляем задачу из crontab
    crontab -u ubuntu -l 2>/dev/null | grep -v "security_monitor.sh" | crontab -u ubuntu -
    
    log_success "Cron задача удалена"
}

# Проверка статуса
check_status() {
    log_info "Проверка статуса cron задач..."
    
    echo "=== Cron задачи пользователя ubuntu ==="
    crontab -u ubuntu -l 2>/dev/null || echo "Нет cron задач"
    
    echo
    echo "=== Последние логи мониторинга ==="
    if [[ -f "logs/security-monitor-cron.log" ]]; then
        tail -20 logs/security-monitor-cron.log
    else
        echo "Лог файл не найден"
    fi
}

# Тестовая отправка
test_notification() {
    log_info "Отправка тестового уведомления..."
    
    if sudo -u ubuntu /usr/bin/bash scripts/security/security_monitor.sh --test; then
        log_success "Тестовое уведомление отправлено"
    else
        log_error "Ошибка отправки тестового уведомления"
    fi
}

# Показать справку
show_help() {
    echo "CloudBridge Relay - Security Monitor Cron Setup"
    echo
    echo "Использование: $0 [команда]"
    echo
    echo "Команды:"
    echo "  install     Установить cron задачу"
    echo "  uninstall   Удалить cron задачу"
    echo "  status      Показать статус"
    echo "  test        Отправить тестовое уведомление"
    echo "  help        Показать эту справку"
    echo
    echo "Примеры:"
    echo "  sudo $0 install    Установить cron задачу"
    echo "  sudo $0 status     Проверить статус"
    echo "  sudo $0 test       Тест уведомлений"
    echo
    echo "После установки:"
    echo "• Проверка безопасности будет выполняться каждые 5 минут"
    echo "• Уведомления будут отправляться в Telegram при обнаружении проблем"
    echo "• Логи будут сохраняться в logs/security-monitor-cron.log"
}

# Основная функция
main() {
    case "${1:-}" in
        install)
            check_root
            install_cron
            test_notification
            log_success "Мониторинг безопасности установлен!"
            echo
            echo "Cron задача настроена для:"
            echo "• Проверки безопасности каждые 5 минут"
            echo "• Отправки уведомлений в Telegram"
            echo "• Логирования в logs/security-monitor-cron.log"
            ;;
        uninstall)
            check_root
            uninstall_cron
            ;;
        status)
            check_status
            ;;
        test)
            test_notification
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            show_help
            exit 1
            ;;
    esac
}

# Запуск основной функции
main "$@" 