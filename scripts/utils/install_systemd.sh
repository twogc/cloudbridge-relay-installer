#!/bin/bash

# CloudBridge Relay - Systemd Service Installer
# Установка systemd сервиса для автозапуска

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Проверка прав root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
}

# Получение пользователя
get_user() {
    if [[ -n "$SUDO_USER" ]]; then
        echo "$SUDO_USER"
    else
        echo "$USER"
    fi
}

# Установка systemd сервиса
install_service() {
    local user=$(get_user)
    local project_dir="/home/$user/cloudbridge-relay-installer"
    
    log "Установка systemd сервиса для пользователя $user..."
    
    # Проверяем существование директории проекта
    if [[ ! -d "$project_dir" ]]; then
        error "Директория проекта не найдена: $project_dir"
        exit 1
    fi
    
    # Копируем unit-файл
    local service_file="/etc/systemd/system/cloudbridge-relay.service"
    local source_file="$project_dir/systemd/cloudbridge-relay.service"
    
    if [[ ! -f "$source_file" ]]; then
        error "Файл сервиса не найден: $source_file"
        exit 1
    fi
    
    # Создаем копию с правильным пользователем
    sed "s/User=ubuntu/User=$user/g; s/Group=ubuntu/Group=$user/g; s|WorkingDirectory=/home/ubuntu/cloudbridge-relay-installer|WorkingDirectory=$project_dir|g; s|ExecStart=/home/ubuntu/cloudbridge-relay-installer/scripts/start_production.sh|ExecStart=$project_dir/scripts/utils/start_production.sh|g; s|ReadWritePaths=/home/ubuntu/cloudbridge-relay-installer|ReadWritePaths=$project_dir|g" "$source_file" > "$service_file"
    
    # Устанавливаем правильные права
    chmod 644 "$service_file"
    
    # Перезагружаем systemd
    systemctl daemon-reload
    
    log "Systemd сервис установлен: $service_file"
}

# Включение автозапуска
enable_service() {
    log "Включение автозапуска сервиса..."
    
    systemctl enable cloudbridge-relay.service
    
    log "Сервис включен для автозапуска"
}

# Проверка установки
verify_installation() {
    log "Проверка установки..."
    
    if systemctl is-enabled cloudbridge-relay.service &> /dev/null; then
        log "Сервис включен для автозапуска"
    else
        warn "Сервис не включен для автозапуска"
    fi
    
    if systemctl is-active cloudbridge-relay.service &> /dev/null; then
        log "Сервис активен"
    else
        log "Сервис не запущен (будет запущен при перезагрузке)"
    fi
    
    log "Статус сервиса:"
    systemctl status cloudbridge-relay.service --no-pager -l
}

# Показать справку по управлению
show_management_help() {
    echo
    log "Управление сервисом:"
    echo "  Запуск:           sudo systemctl start cloudbridge-relay"
    echo "  Остановка:        sudo systemctl stop cloudbridge-relay"
    echo "  Перезапуск:       sudo systemctl restart cloudbridge-relay"
    echo "  Статус:           sudo systemctl status cloudbridge-relay"
    echo "  Логи:             sudo journalctl -u cloudbridge-relay -f"
    echo "  Отключить автозапуск: sudo systemctl disable cloudbridge-relay"
    echo
}

# Удаление сервиса
uninstall_service() {
    log "Удаление systemd сервиса..."
    
    # Останавливаем сервис
    if systemctl is-active cloudbridge-relay.service &> /dev/null; then
        systemctl stop cloudbridge-relay.service
    fi
    
    # Отключаем автозапуск
    if systemctl is-enabled cloudbridge-relay.service &> /dev/null; then
        systemctl disable cloudbridge-relay.service
    fi
    
    # Удаляем файл сервиса
    if [[ -f "/etc/systemd/system/cloudbridge-relay.service" ]]; then
        rm -f "/etc/systemd/system/cloudbridge-relay.service"
    fi
    
    # Перезагружаем systemd
    systemctl daemon-reload
    
    log "Сервис удален"
}

# Показать справку
show_help() {
    echo "CloudBridge Relay - Systemd Service Installer"
    echo
    echo "Использование: $0 [опции]"
    echo
    echo "Опции:"
    echo "  install              Установить systemd сервис"
    echo "  uninstall            Удалить systemd сервис"
    echo "  status               Показать статус сервиса"
    echo "  enable               Включить автозапуск"
    echo "  disable              Отключить автозапуск"
    echo "  start                Запустить сервис"
    echo "  stop                 Остановить сервис"
    echo "  restart              Перезапустить сервис"
    echo "  logs                 Показать логи"
    echo "  -h, --help           Показать эту справку"
    echo
    echo "Примеры:"
    echo "  sudo $0 install      # Установить сервис"
    echo "  sudo $0 uninstall    # Удалить сервис"
    echo "  sudo $0 status       # Показать статус"
    echo
}

# Основная функция
main() {
    case "${1:-}" in
        install)
            check_root
            install_service
            enable_service
            verify_installation
            show_management_help
            log "Systemd сервис успешно установлен!"
            ;;
        uninstall)
            check_root
            uninstall_service
            log "🗑️ Systemd сервис удален"
            ;;
        status)
            if [[ $EUID -eq 0 ]]; then
                systemctl status cloudbridge-relay.service --no-pager -l
            else
                sudo systemctl status cloudbridge-relay.service --no-pager -l
            fi
            ;;
        enable)
            check_root
            enable_service
            ;;
        disable)
            check_root
            systemctl disable cloudbridge-relay.service
            log "Автозапуск отключен"
            ;;
        start)
            check_root
            systemctl start cloudbridge-relay.service
            log "Сервис запущен"
            ;;
        stop)
            check_root
            systemctl stop cloudbridge-relay.service
            log "Сервис остановлен"
            ;;
        restart)
            check_root
            systemctl restart cloudbridge-relay.service
            log "Сервис перезапущен"
            ;;
        logs)
            if [[ $EUID -eq 0 ]]; then
                journalctl -u cloudbridge-relay.service -f
            else
                sudo journalctl -u cloudbridge-relay.service -f
            fi
            ;;
        -h|--help|"")
            show_help
            ;;
        *)
            error "Неизвестная команда: $1"
            show_help
            exit 1
            ;;
    esac
}

# Запуск скрипта
main "$@" 