#!/bin/bash

# CloudBridge Relay Firewall Setup Script
# Настройка firewall для блокировки всех портов, кроме SSH и relay

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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v ufw &> /dev/null; then
        log_info "Установка UFW..."
        apt-get update
        apt-get install -y ufw
    fi
    
    if ! command -v iptables &> /dev/null; then
        log_info "Установка iptables..."
        apt-get update
        apt-get install -y iptables
    fi
    
    log_success "Все зависимости установлены"
}

# Определение портов relay из переменных окружения
get_relay_ports() {
    log_info "Определение портов relay..."
    
    # Загрузка переменных из .env файла
    if [ -f ".env" ]; then
        source .env
    fi
    
    # Порт relay (по умолчанию 8082)
    RELAY_PORT=${RELAY_PORT:-8082}
    
    # Порт метрик (по умолчанию 8081)
    RELAY_METRICS_PORT=${RELAY_METRICS_PORT:-8081}
    
    # Порт Keycloak (по умолчанию 8080)
    KEYCLOAK_PORT=${KEYCLOAK_PORT:-8080}
    
    # Порт Grafana (по умолчанию 3000)
    GRAFANA_PORT=${GRAFANA_PORT:-3000}
    
    # Порт Prometheus (по умолчанию 9090)
    PROMETHEUS_PORT=${PROMETHEUS_PORT:-9090}
    
    # Порт AlertManager (по умолчанию 9093)
    ALERTMANAGER_PORT=${ALERTMANAGER_PORT:-9093}
    
    # Порт Blackbox (по умолчанию 9115)
    BLACKBOX_PORT=${BLACKBOX_PORT:-9115}
    
    log_info "Определенные порты:"
    log_info "  Relay: $RELAY_PORT"
    log_info "  Metrics: $RELAY_METRICS_PORT"
    log_info "  Keycloak: $KEYCLOAK_PORT"
    log_info "  Grafana: $GRAFANA_PORT"
    log_info "  Prometheus: $PROMETHEUS_PORT"
    log_info "  AlertManager: $ALERTMANAGER_PORT"
    log_info "  Blackbox: $BLACKBOX_PORT"
}

# Настройка UFW
setup_ufw() {
    log_info "Настройка UFW firewall..."
    
    # Сброс правил
    ufw --force reset
    
    # Установка политик по умолчанию
    ufw default deny incoming
    ufw default allow outgoing
    
    # Разрешение SSH (порт 22)
    ufw allow 22/tcp comment "SSH"
    log_success "SSH (порт 22) разрешен"
    
    # Разрешение портов relay
    ufw allow $RELAY_PORT/tcp comment "Relay Server"
    log_success "Relay Server (порт $RELAY_PORT) разрешен"
    
    # Разрешение порта метрик (только для локального доступа)
    ufw allow from 127.0.0.1 to any port $RELAY_METRICS_PORT comment "Relay Metrics (localhost only)"
    log_success "Relay Metrics (порт $RELAY_METRICS_PORT) разрешен только для localhost"
    
    # Разрешение портов мониторинга (только для локального доступа)
    ufw allow from 127.0.0.1 to any port $KEYCLOAK_PORT comment "Keycloak (localhost only)"
    ufw allow from 127.0.0.1 to any port $GRAFANA_PORT comment "Grafana (localhost only)"
    ufw allow from 127.0.0.1 to any port $PROMETHEUS_PORT comment "Prometheus (localhost only)"
    ufw allow from 127.0.0.1 to any port $ALERTMANAGER_PORT comment "AlertManager (localhost only)"
    ufw allow from 127.0.0.1 to any port $BLACKBOX_PORT comment "Blackbox (localhost only)"
    
    log_success "Порты мониторинга разрешены только для localhost"
    
    # Включение UFW
    ufw --force enable
    
    log_success "UFW включен и настроен"
}

# Настройка iptables (дополнительная защита)
setup_iptables() {
    log_info "Настройка дополнительных правил iptables..."
    
    # Очистка существующих правил
    iptables -F
    iptables -X
    iptables -t nat -F
    iptables -t nat -X
    iptables -t mangle -F
    iptables -t mangle -X
    
    # Установка политик по умолчанию
    iptables -P INPUT DROP
    iptables -P FORWARD DROP
    iptables -P OUTPUT ACCEPT
    
    # Разрешение loopback
    iptables -A INPUT -i lo -j ACCEPT
    iptables -A OUTPUT -o lo -j ACCEPT
    
    # Разрешение установленных соединений
    iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
    
    # Разрешение SSH
    iptables -A INPUT -p tcp --dport 22 -j ACCEPT
    
    # Разрешение relay порта
    iptables -A INPUT -p tcp --dport $RELAY_PORT -j ACCEPT
    
    # Разрешение портов мониторинга только для localhost
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $RELAY_METRICS_PORT -j ACCEPT
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $KEYCLOAK_PORT -j ACCEPT
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $GRAFANA_PORT -j ACCEPT
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $PROMETHEUS_PORT -j ACCEPT
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $ALERTMANAGER_PORT -j ACCEPT
    iptables -A INPUT -p tcp -s 127.0.0.1 --dport $BLACKBOX_PORT -j ACCEPT
    
    # Защита от DDoS атак
    iptables -A INPUT -p tcp --dport 22 -m limit --limit 6/min --limit-burst 10 -j ACCEPT
    iptables -A INPUT -p tcp --dport $RELAY_PORT -m limit --limit 30/min --limit-burst 60 -j ACCEPT
    
    # Блокировка подозрительных пакетов
    iptables -A INPUT -p tcp --tcp-flags ALL NONE -j DROP
    iptables -A INPUT -p tcp --tcp-flags ALL ALL -j DROP
    
    # Блокировка сканирования портов
    iptables -A INPUT -p tcp --dport 22 -m recent --name SSH --set
    iptables -A INPUT -p tcp --dport 22 -m recent --name SSH --update --seconds 60 --hitcount 4 -j DROP
    
    log_success "Дополнительные правила iptables настроены"
}

# Сохранение правил iptables
save_iptables() {
    log_info "Сохранение правил iptables..."
    
    if command -v iptables-save &> /dev/null; then
        iptables-save > /etc/iptables/rules.v4
        log_success "Правила iptables сохранены"
    else
        log_warning "iptables-save не найден, правила не сохранены"
    fi
}

# Настройка автоматического восстановления правил
setup_iptables_persistent() {
    log_info "Настройка автоматического восстановления правил iptables..."
    
    # Установка пакета для сохранения правил
    if ! dpkg -l | grep -q iptables-persistent; then
        apt-get update
        echo iptables-persistent iptables-persistent/autosave_v4 boolean true | debconf-set-selections
        echo iptables-persistent iptables-persistent/autosave_v6 boolean true | debconf-set-selections
        apt-get install -y iptables-persistent
    fi
    
    # Создание скрипта для восстановления правил
    cat > /etc/network/if-pre-up.d/iptables << 'EOF'
#!/bin/bash
iptables-restore < /etc/iptables/rules.v4
EOF
    
    chmod +x /etc/network/if-pre-up.d/iptables
    
    log_success "Автоматическое восстановление правил настроено"
}

# Настройка fail2ban для дополнительной защиты
setup_fail2ban() {
    log_info "Настройка fail2ban..."
    
    if ! command -v fail2ban-client &> /dev/null; then
        apt-get update
        apt-get install -y fail2ban
    fi
    
    # Создание конфигурации fail2ban
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

[relay]
enabled = true
port = $RELAY_PORT
filter = relay
logpath = /var/log/relay/relay.log
maxretry = 5
EOF
    
    # Создание фильтра для relay
    cat > /etc/fail2ban/filter.d/relay.conf << EOF
[Definition]
failregex = ^.*authentication failed.*from <HOST>.*$
ignoreregex =
EOF
    
    # Перезапуск fail2ban
    systemctl restart fail2ban
    systemctl enable fail2ban
    
    log_success "fail2ban настроен и запущен"
}

# Перезапуск Docker для восстановления цепочек после изменения iptables
restart_docker_if_installed() {
    if command -v docker &> /dev/null; then
        log_info "Перезапуск Docker для восстановления сетевых цепочек..."
        systemctl restart docker
        log_success "Docker перезапущен"
    else
        log_info "Docker не установлен, пропускаю перезапуск"
    fi
}

# Проверка статуса firewall
check_firewall_status() {
    log_info "Проверка статуса firewall..."
    
    echo
    echo "=== Статус UFW ==="
    ufw status verbose
    
    echo
    echo "=== Правила iptables ==="
    iptables -L -n -v
    
    echo
    echo "=== Статус fail2ban ==="
    fail2ban-client status
    
    echo
    echo "=== Открытые порты ==="
    netstat -tlnp | grep LISTEN
}

# Создание скрипта для управления firewall
create_firewall_management_script() {
    log_info "Создание скрипта управления firewall..."
    
    cat > /usr/local/bin/relay-firewall << 'EOF'
#!/bin/bash

# CloudBridge Relay Firewall Management Script

case "$1" in
    status)
        echo "=== UFW Status ==="
        ufw status verbose
        echo
        echo "=== iptables Rules ==="
        iptables -L -n -v
        echo
        echo "=== fail2ban Status ==="
        fail2ban-client status
        ;;
    allow-ip)
        if [ -z "$2" ]; then
            echo "Usage: $0 allow-ip <IP_ADDRESS>"
            exit 1
        fi
        ufw allow from $2
        echo "IP $2 разрешен"
        ;;
    deny-ip)
        if [ -z "$2" ]; then
            echo "Usage: $0 deny-ip <IP_ADDRESS>"
            exit 1
        fi
        ufw deny from $2
        echo "IP $2 заблокирован"
        ;;
    reload)
        ufw reload
        iptables-restore < /etc/iptables/rules.v4
        systemctl reload fail2ban
        echo "Firewall перезагружен"
        ;;
    *)
        echo "Usage: $0 {status|allow-ip|deny-ip|reload}"
        echo
        echo "Commands:"
        echo "  status     - Показать статус firewall"
        echo "  allow-ip   - Разрешить IP адрес"
        echo "  deny-ip    - Заблокировать IP адрес"
        echo "  reload     - Перезагрузить firewall"
        ;;
esac
EOF
    
    chmod +x /usr/local/bin/relay-firewall
    
    log_success "Скрипт управления firewall создан: /usr/local/bin/relay-firewall"
}

# Основная функция
main() {
    log_info "Начало настройки firewall для CloudBridge Relay"
    
    # Проверка root прав
    check_root
    
    # Проверка зависимостей
    check_dependencies
    
    # Определение портов
    get_relay_ports
    
    # Настройка UFW
    setup_ufw
    
    # Настройка iptables
    setup_iptables
    
    # Сохранение правил
    save_iptables
    
    # Перезапуск Docker (новое)
    restart_docker_if_installed
    
    # Настройка автоматического восстановления
    setup_iptables_persistent
    
    # Настройка fail2ban
    setup_fail2ban
    
    # Создание скрипта управления
    create_firewall_management_script
    
    # Проверка статуса
    check_firewall_status
    
    log_success "Настройка firewall завершена!"
    echo
    echo "Доступные команды:"
    echo "  relay-firewall status     - Показать статус"
    echo "  relay-firewall allow-ip   - Разрешить IP"
    echo "  relay-firewall deny-ip    - Заблокировать IP"
    echo "  relay-firewall reload     - Перезагрузить"
    echo
    echo "ВАЖНО: SSH соединение может быть прервано!"
    echo "Убедитесь, что у вас есть альтернативный доступ к серверу."
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay Firewall Setup Script"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --check        Только проверить статус"
        echo "  --dry-run      Показать что будет сделано без применения"
        echo
        echo "Примеры:"
        echo "  sudo $0              Настроить firewall"
        echo "  sudo $0 --check      Проверить статус"
        echo "  sudo $0 --dry-run    Показать план действий"
        ;;
    --check)
        check_firewall_status
        ;;
    --dry-run)
        log_info "Режим dry-run - показать план действий"
        get_relay_ports
        echo
        echo "Будут настроены следующие правила:"
        echo "  - Разрешен SSH (порт 22)"
        echo "  - Разрешен Relay Server (порт $RELAY_PORT)"
        echo "  - Разрешены порты мониторинга только для localhost:"
        echo "    * Metrics: $RELAY_METRICS_PORT"
        echo "    * Keycloak: $KEYCLOAK_PORT"
        echo "    * Grafana: $GRAFANA_PORT"
        echo "    * Prometheus: $PROMETHEUS_PORT"
        echo "    * AlertManager: $ALERTMANAGER_PORT"
        echo "    * Blackbox: $BLACKBOX_PORT"
        echo "  - Все остальные порты заблокированы"
        echo "  - Настроен fail2ban для защиты от атак"
        ;;
    *)
        main
        ;;
esac 