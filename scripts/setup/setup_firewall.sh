#!/bin/bash

# CloudBridge Relay - Настройка файрвола
# Этот скрипт настраивает UFW для защиты сервисов CloudBridge Relay

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
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
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
}

# Проверка UFW
check_ufw() {
    if ! command -v ufw &> /dev/null; then
        log_error "UFW не установлен. Установите его: sudo apt install ufw"
        exit 1
    fi
}

# Настройка базовых правил
setup_basic_rules() {
    log_info "Настройка базовых правил файрвола..."
    
    # Сброс правил
    ufw --force reset
    
    # Базовые политики
    ufw default deny incoming
    ufw default allow outgoing
    
    log_success "Базовые правила настроены"
}

# Настройка SSH
setup_ssh() {
    log_info "Настройка SSH доступа..."
    
    # SSH порт
    ufw allow 22/tcp comment "SSH"
    
    # Rate limiting для SSH
    ufw limit 22/tcp comment "SSH rate limiting"
    
    log_success "SSH доступ настроен"
}

# Настройка внешних сервисов
setup_external_services() {
    log_info "Настройка внешних сервисов..."
    
    # Keycloak Admin Console (внешний доступ)
    ufw allow 8080/tcp comment "Keycloak Admin Console"
    
    # Relay API (внешний доступ)
    ufw allow 8082/tcp comment "Relay API"
    
    # Relay TLS (внешний доступ)
    ufw allow 8080/tcp comment "Relay TLS"
    
    log_success "Внешние сервисы настроены"
}

# Настройка внутренних сервисов (только localhost)
setup_internal_services() {
    log_info "Настройка внутренних сервисов..."
    
    # Relay Metrics (только localhost)
    ufw allow from 127.0.0.1 to any port 8081 comment "Relay Metrics (localhost only)"
    
    # Prometheus (только localhost)
    ufw allow from 127.0.0.1 to any port 9090 comment "Prometheus (localhost only)"
    
    # AlertManager (только localhost)
    ufw allow from 127.0.0.1 to any port 9093 comment "AlertManager (localhost only)"
    
    # Blackbox Exporter (только localhost)
    ufw allow from 127.0.0.1 to any port 9115 comment "Blackbox Exporter (localhost only)"
    
    # Grafana (только localhost)
    ufw allow from 127.0.0.1 to any port 3000 comment "Grafana (localhost only)"
    
    log_success "Внутренние сервисы настроены"
}

# Настройка Docker сети
setup_docker_network() {
    log_info "Настройка Docker сети..."
    
    # Разрешить трафик внутри Docker сети
    ufw allow from 172.16.0.0/12 comment "Docker internal network"
    ufw allow from 192.168.0.0/16 comment "Docker internal network"
    ufw allow from 10.0.0.0/8 comment "Docker internal network"
    
    log_success "Docker сеть настроена"
}

# Дополнительные правила безопасности
setup_security_rules() {
    log_info "Настройка дополнительных правил безопасности..."
    
    # Защита от сканирования
    ufw deny in on eth0 from 0.0.0.0/0 to any port 1433 comment "Block SQL Server"
    ufw deny in on eth0 from 0.0.0.0/0 to any port 3389 comment "Block RDP"
    
    log_success "Дополнительные правила безопасности настроены"
}

# Проверка конфигурации
verify_configuration() {
    log_info "Проверка конфигурации файрвола..."
    
    echo ""
    log_info "Текущие правила UFW:"
    ufw status numbered
    
    echo ""
    log_info "Проверка доступности сервисов:"
    
    # Проверка внешних сервисов
    local services=(
        "8080:Keycloak"
        "8082:Relay API"
    )
    
    for service in "${services[@]}"; do
        IFS=':' read -r port name <<< "$service"
        if ufw status | grep -q "$port/tcp.*ALLOW"; then
            log_success "$name (порт $port) - разрешен"
        else
            log_error "$name (порт $port) - НЕ разрешен"
        fi
    done
    
    # Проверка внутренних сервисов
    local internal_services=(
        "8081:Relay Metrics"
        "9090:Prometheus"
        "9093:AlertManager"
        "9115:Blackbox"
        "3000:Grafana"
    )
    
    for service in "${internal_services[@]}"; do
        IFS=':' read -r port name <<< "$service"
        if ufw status | grep -q "$port.*127.0.0.1"; then
            log_success "$name (порт $port) - разрешен только localhost"
        else
            log_warning "$name (порт $port) - проверьте настройки"
        fi
    done
}

# Активация файрвола
activate_firewall() {
    log_info "Активация файрвола..."
    
    # Включение UFW
    ufw --force enable
    
    log_success "Файрвол активирован"
}

# Основная функция
main() {
    log_info "Настройка файрвола для CloudBridge Relay..."
    
    check_root
    check_ufw
    
    setup_basic_rules
    setup_ssh
    setup_external_services
    setup_internal_services
    setup_docker_network
    setup_security_rules
    
    echo ""
    log_warning "ВНИМАНИЕ: Файрвол будет активирован!"
    log_warning "Убедитесь, что у вас есть доступ к SSH, иначе вы можете потерять доступ к серверу."
    echo ""
    
    read -p "Продолжить активацию файрвола? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        activate_firewall
        verify_configuration
        
        log_success "Файрвол настроен и активирован!"
        log_info "Внешние порты (8080, 8082): открыты для всех"
        log_info "Внутренние порты (8081, 9090, 9093, 9115, 3000): только localhost"
        log_info "SSH (22): с rate limiting"
    else
        log_warning "Активация файрвола отменена"
        log_info "Для активации запустите: sudo ufw enable"
    fi
}

# Обработка аргументов
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --status       Показать статус файрвола"
        echo "  --disable      Отключить файрвол"
        echo "  --reset        Сбросить правила файрвола"
        echo ""
        echo "Примеры:"
        echo "  sudo $0                    # Полная настройка"
        echo "  sudo $0 --status          # Показать статус"
        echo "  sudo $0 --disable         # Отключить файрвол"
        ;;
    --status)
        ufw status verbose
        ;;
    --disable)
        ufw disable
        log_success "Файрвол отключен"
        ;;
    --reset)
        ufw --force reset
        log_success "Правила файрвола сброшены"
        ;;
    "")
        main "$@"
        ;;
    *)
        log_error "Неизвестная опция: $1"
        echo "Используйте --help для справки"
        exit 1
        ;;
esac 