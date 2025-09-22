#!/bin/bash

# CloudBridge Relay UFW Management Script
# Управление правилами UFW для CloudBridge Relay

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка прав root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "Этот скрипт должен запускаться с правами root (sudo)"
        exit 1
    fi
}

# Показать текущие правила CloudBridge Relay
show_rules() {
    log "Текущие правила UFW для CloudBridge Relay:"
    echo
    ufw status | grep -E "(555[0-5]|909[0-2]|19302|8443)" || echo "Правила не найдены"
}

# Добавить правила CloudBridge Relay
add_rules() {
    log "Добавление правил UFW для CloudBridge Relay..."
    
    # TCP порты
    ufw allow 5550/tcp comment "CloudBridge Relay - Main TCP Server"
    ufw allow 5551/tcp comment "CloudBridge Relay - Metrics (Prometheus)"
    ufw allow 5552/tcp comment "CloudBridge Relay - P2P API"
    ufw allow 5553/tcp comment "CloudBridge Relay - HTTP API"
    ufw allow 5554/tcp comment "CloudBridge Relay - HTTPS API"
    
    # UDP порты
    ufw allow 5555/udp comment "CloudBridge Relay - QUIC Transport"
    ufw allow 9090/udp comment "CloudBridge Relay - QUIC Legacy"
    ufw allow 9092/udp comment "CloudBridge Relay - Enhanced QUIC"
    ufw allow 19302/udp comment "CloudBridge Relay - STUN Server"
    ufw allow 8443/udp comment "CloudBridge Relay - MASQUE Protocol"
    ufw allow 5554/udp comment "CloudBridge Relay - UDP Transport"
    
    log "Правила успешно добавлены!"
}

# Удалить правила CloudBridge Relay
remove_rules() {
    log "Удаление правил UFW для CloudBridge Relay..."
    
    # TCP порты
    ufw delete allow 5550/tcp 2>/dev/null || true
    ufw delete allow 5551/tcp 2>/dev/null || true
    ufw delete allow 5552/tcp 2>/dev/null || true
    ufw delete allow 5553/tcp 2>/dev/null || true
    ufw delete allow 5554/tcp 2>/dev/null || true
    
    # UDP порты
    ufw delete allow 5555/udp 2>/dev/null || true
    ufw delete allow 9090/udp 2>/dev/null || true
    ufw delete allow 9092/udp 2>/dev/null || true
    ufw delete allow 19302/udp 2>/dev/null || true
    ufw delete allow 8443/udp 2>/dev/null || true
    ufw delete allow 5554/udp 2>/dev/null || true
    
    log "Правила успешно удалены!"
}

# Проверить статус UFW
check_ufw() {
    log "Проверка статуса UFW..."
    ufw status verbose
}

# Показать справку
show_help() {
    echo "CloudBridge Relay UFW Management Script"
    echo
    echo "Использование: $0 [команда]"
    echo
    echo "Команды:"
    echo "  add      - Добавить правила UFW для CloudBridge Relay"
    echo "  remove   - Удалить правила UFW для CloudBridge Relay"
    echo "  show     - Показать текущие правила CloudBridge Relay"
    echo "  status   - Показать полный статус UFW"
    echo "  help     - Показать эту справку"
    echo
    echo "Примеры:"
    echo "  sudo $0 add      # Добавить правила"
    echo "  sudo $0 remove   # Удалить правила"
    echo "  sudo $0 show     # Показать правила"
}

# Основная логика
main() {
    case "${1:-help}" in
        "add")
            check_root
            add_rules
            show_rules
            ;;
        "remove")
            check_root
            remove_rules
            ;;
        "show")
            show_rules
            ;;
        "status")
            check_ufw
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# Запуск скрипта
main "$@"
