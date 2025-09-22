#!/bin/bash

# CloudBridge Relay Certificate Update Script
# Автоматическое обновление сертификатов из Let's Encrypt

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

# Проверка root прав
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
}

# Обновление сертификатов
update_certificates() {
    local domain=${1:-auth.2gc.ru}
    
    log_info "Обновление сертификатов для домена: $domain"
    
    # Проверка существования сертификатов Let's Encrypt
    if [ ! -f "/etc/letsencrypt/live/$domain/fullchain.pem" ]; then
        log_error "Сертификаты Let's Encrypt для домена $domain не найдены"
        log_info "Запустите: ./scripts/setup/setup_certs.sh $domain"
        exit 1
    fi
    
    # Создание директории certs если не существует
    mkdir -p certs
    
    # Копирование сертификатов
    log_info "Копирование сертификатов..."
    cp "/etc/letsencrypt/live/$domain/fullchain.pem" "certs/fullchain.pem"
    cp "/etc/letsencrypt/live/$domain/privkey.pem" "certs/privkey.pem"
    
    # Установка правильных прав
    chown ubuntu:ubuntu certs/fullchain.pem certs/privkey.pem
    chmod 600 certs/fullchain.pem certs/privkey.pem
    
    log_success "Сертификаты обновлены"
}

# Проверка сертификатов
verify_certificates() {
    log_info "Проверка сертификатов..."
    
    if [ ! -f "certs/fullchain.pem" ] || [ ! -f "certs/privkey.pem" ]; then
        log_error "Сертификаты не найдены в папке certs/"
        return 1
    fi
    
    # Проверка соответствия сертификата и ключа
    local cert_hash=$(openssl x509 -noout -modulus -in certs/fullchain.pem | openssl md5)
    local key_hash=$(openssl rsa -noout -modulus -in certs/privkey.pem | openssl md5)
    
    if [ "$cert_hash" != "$key_hash" ]; then
        log_error "Сертификат и ключ не соответствуют друг другу"
        return 1
    fi
    
    # Проверка срока действия
    local expiry_date=$(openssl x509 -in certs/fullchain.pem -noout -enddate | cut -d= -f2)
    local expiry_epoch=$(date -d "$expiry_date" +%s)
    local current_epoch=$(date +%s)
    local days_left=$(( (expiry_epoch - current_epoch) / 86400 ))
    
    log_info "Сертификат действителен до: $expiry_date"
    log_info "Осталось дней: $days_left"
    
    if [ $days_left -lt 30 ]; then
        log_warning "Сертификат истекает через $days_left дней"
    fi
    
    # Проверка домена
    local domain=$(openssl x509 -in certs/fullchain.pem -noout -text | grep -A1 "Subject:" | grep "CN = " | sed 's/.*CN = //')
    log_info "Домен сертификата: $domain"
    
    log_success "Сертификаты проверены успешно"
}

# Перезапуск сервисов
restart_services() {
    log_info "Перезапуск сервисов..."
    
    if command -v docker-compose &> /dev/null; then
        docker-compose restart relay
        log_success "Сервис relay перезапущен"
    else
        log_warning "docker-compose не найден, перезапустите сервисы вручную"
    fi
}

# Основная функция
main() {
    local domain=${1:-auth.2gc.ru}
    
    log_info "Начало обновления сертификатов CloudBridge Relay"
    
    # Проверка root прав
    check_root
    
    # Обновление сертификатов
    update_certificates "$domain"
    
    # Проверка сертификатов
    verify_certificates
    
    # Перезапуск сервисов
    restart_services
    
    log_success "Обновление сертификатов завершено!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay Certificate Update Script"
        echo
        echo "Использование: $0 [домен]"
        echo
        echo "Аргументы:"
        echo "  домен    Домен для обновления сертификатов (по умолчанию: auth.2gc.ru)"
        echo
        echo "Примеры:"
        echo "  sudo $0                    Обновить сертификаты для auth.2gc.ru"
        echo "  sudo $0 example.com        Обновить сертификаты для example.com"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --verify       Только проверить сертификаты"
        ;;
    --verify)
        verify_certificates
        ;;
    *)
        main "$@"
        ;;
esac 