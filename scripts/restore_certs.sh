#!/bin/bash

# CloudBridge Relay Certificate Restore Script
# Восстановление сертификатов при загрузке системы

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Восстановление сертификатов
restore_certificates() {
    local domain=${1:-edge.2gc.ru}
    
    log_info "Восстановление сертификатов для домена: $domain"
    
    # Проверка существования сертификатов Let's Encrypt
    if ! sudo test -f "/etc/letsencrypt/live/$domain/fullchain.pem"; then
        log_error "Сертификаты Let's Encrypt для домена $domain не найдены"
        return 1
    fi
    
    # Создание директории certs если не существует
    mkdir -p certs
    
    # Копирование сертификатов с правами root
    log_info "Копирование сертификатов..."
    sudo cp "/etc/letsencrypt/live/$domain/fullchain.pem" "certs/fullchain.pem"
    sudo cp "/etc/letsencrypt/live/$domain/privkey.pem" "certs/privkey.pem"
    
    # Установка правильных прав
    sudo chown ubuntu:ubuntu certs/fullchain.pem certs/privkey.pem
    sudo chmod 600 certs/fullchain.pem certs/privkey.pem
    
    log_success "Сертификаты восстановлены"
}

# Основная функция
main() {
    local domain=${1:-edge.2gc.ru}
    
    log_info "Восстановление сертификатов CloudBridge Relay"
    
    # Переход в директорию проекта
    cd /home/ubuntu/cloudbridge-relay-installer
    
    # Восстановление сертификатов
    if restore_certificates "$domain"; then
        log_success "Восстановление сертификатов завершено!"
    else
        log_error "Ошибка при восстановлении сертификатов"
        exit 1
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay Certificate Restore Script"
        echo
        echo "Использование: $0 [домен]"
        echo
        echo "Аргументы:"
        echo "  домен    Домен для восстановления сертификатов (по умолчанию: edge.2gc.ru)"
        echo
        echo "Примеры:"
        echo "  $0                    Восстановить сертификаты для edge.2gc.ru"
        echo "  $0 example.com        Восстановить сертификаты для example.com"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        ;;
    *)
        main "$@"
        ;;
esac 