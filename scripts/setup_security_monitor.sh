#!/bin/bash

# CloudBridge Relay - Security Monitor Setup
# Установка и настройка системы мониторинга безопасности

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

# Установка systemd сервиса
install_service() {
    log_info "Установка systemd сервиса для мониторинга безопасности..."
    
    # Копируем сервис файл
    cp systemd/security-monitor.service /etc/systemd/system/
    
    # Перезагружаем systemd
    systemctl daemon-reload
    
    # Включаем автозапуск
    systemctl enable security-monitor.service
    
    log_success "Сервис установлен и включен"
}

# Запуск сервиса
start_service() {
    log_info "Запуск сервиса мониторинга безопасности..."
    
    if systemctl start security-monitor.service; then
        log_success "Сервис запущен"
    else
        log_error "Ошибка запуска сервиса"
        systemctl status security-monitor.service --no-pager
        exit 1
    fi
}

# Остановка сервиса
stop_service() {
    log_info "Остановка сервиса мониторинга безопасности..."
    
    if systemctl stop security-monitor.service; then
        log_success "Сервис остановлен"
    else
        log_error "Ошибка остановки сервиса"
    fi
}

# Проверка статуса
check_status() {
    log_info "Проверка статуса сервиса..."
    
    systemctl status security-monitor.service --no-pager
    
    echo
    echo "=== Логи сервиса ==="
    journalctl -u security-monitor.service --no-pager -n 20
}

# Тестовая отправка
test_notification() {
    log_info "Отправка тестового уведомления..."
    
    if ./scripts/security_monitor.sh --test; then
        log_success "Тестовое уведомление отправлено"
    else
        log_error "Ошибка отправки тестового уведомления"
    fi
}

# Удаление сервиса
uninstall_service() {
    log_info "Удаление сервиса мониторинга безопасности..."
    
    # Останавливаем сервис
    systemctl stop security-monitor.service 2>/dev/null || true
    
    # Отключаем автозапуск
    systemctl disable security-monitor.service 2>/dev/null || true
    
    # Удаляем файл сервиса
    rm -f /etc/systemd/system/security-monitor.service
    
    # Перезагружаем systemd
    systemctl daemon-reload
    
    log_success "Сервис удален"
}

# Показать справку
show_help() {
    echo "CloudBridge Relay - Security Monitor Setup"
    echo
    echo "Использование: $0 [команда]"
    echo
    echo "Команды:"
    echo "  install     Установить и запустить сервис"
    echo "  start       Запустить сервис"
    echo "  stop        Остановить сервис"
    echo "  restart     Перезапустить сервис"
    echo "  status      Показать статус сервиса"
    echo "  test        Отправить тестовое уведомление"
    echo "  uninstall   Удалить сервис"
    echo "  help        Показать эту справку"
    echo
    echo "Примеры:"
    echo "  sudo $0 install    Установить и запустить"
    echo "  sudo $0 status     Проверить статус"
    echo "  sudo $0 test       Тест уведомлений"
    echo
    echo "После установки сервис будет:"
    echo "• Автоматически запускаться при загрузке системы"
    echo "• Проверять безопасность каждые 5 минут"
    echo "• Отправлять уведомления в Telegram при обнаружении проблем"
}

# Основная функция
main() {
    case "${1:-}" in
        install)
            check_root
            install_service
            start_service
            test_notification
            log_success "Мониторинг безопасности установлен и запущен!"
            echo
            echo "Сервис будет автоматически:"
            echo "• Запускаться при загрузке системы"
            echo "• Проверять безопасность каждые 5 минут"
            echo "• Отправлять уведомления в Telegram"
            ;;
        start)
            check_root
            start_service
            ;;
        stop)
            check_root
            stop_service
            ;;
        restart)
            check_root
            stop_service
            start_service
            ;;
        status)
            check_status
            ;;
        test)
            test_notification
            ;;
        uninstall)
            check_root
            uninstall_service
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