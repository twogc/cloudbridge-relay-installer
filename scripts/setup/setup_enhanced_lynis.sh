#!/bin/bash

# CloudBridge Relay - Enhanced Lynis Setup Script
# Установка и настройка расширенного аудита безопасности

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("lynis" "jq" "curl" "docker")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_warn "Отсутствуют зависимости: ${missing_deps[*]}"
        log_info "Установите их вручную или используйте базовый Lynis"
    else
        log_info "Все зависимости установлены"
    fi
}

# Установка расширенного профиля
install_enhanced_profile() {
    log_info "Установка расширенного профиля Lynis..."
    
    # Создаем директорию если не существует
    sudo mkdir -p /etc/lynis
    
    # Копируем расширенный профиль
    sudo cp etc/lynis/enhanced_custom.prf /etc/lynis/
    sudo chmod 644 /etc/lynis/enhanced_custom.prf
    
    log_info "Расширенный профиль установлен"
}

# Установка systemd сервисов
install_systemd_services() {
    log_info "Установка systemd сервисов..."
    
    # Копируем файлы systemd
    sudo cp systemd/enhanced-lynis-audit.service /etc/systemd/system/
    sudo cp systemd/enhanced-lynis-audit.timer /etc/systemd/system/
    
    # Перезагружаем systemd
    sudo systemctl daemon-reload
    
    # Включаем таймер
    sudo systemctl enable enhanced-lynis-audit.timer
    
    log_info "Systemd сервисы установлены"
}

# Создание директорий
create_directories() {
    log_info "Создание директорий..."
    
    sudo mkdir -p /var/log/lynis/reports
    sudo mkdir -p /var/lib/lynis-telegram
    sudo mkdir -p /home/ubuntu/cloudbridge-relay-installer/logs
    
    # Устанавливаем права доступа
    sudo chmod 750 /var/log/lynis /var/log/lynis/reports
    sudo chown root:root /var/log/lynis /var/log/lynis/reports
    
    log_info "Директории созданы"
}

# Настройка ротации логов
setup_log_rotation() {
    log_info "Настройка ротации логов..."
    
    cat > /tmp/lynis-logrotate << 'EOF'
/var/log/lynis/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        systemctl reload enhanced-lynis-audit.service > /dev/null 2>&1 || true
    endscript
}
EOF
    
    sudo cp /tmp/lynis-logrotate /etc/logrotate.d/lynis
    sudo chmod 644 /etc/logrotate.d/lynis
    
    log_info "Ротация логов настроена"
}

# Тестирование системы
test_system() {
    log_info "Тестирование расширенного аудита..."
    
    # Тест расширенного скрипта
    if [[ -f "scripts/security/enhanced_lynis_audit.sh" ]]; then
        log_info "Тестирование расширенного скрипта..."
        sudo bash scripts/security/enhanced_lynis_audit.sh --test
    fi
    
    # Тест systemd сервиса
    log_info "Тестирование systemd сервиса..."
    sudo systemctl start enhanced-lynis-audit.service
    
    # Ждем завершения
    sleep 5
    
    # Проверяем статус
    if sudo systemctl is-active enhanced-lynis-audit.service &> /dev/null; then
        log_info "Сервис работает"
    else
        log_warn "Сервис не активен, проверьте логи"
    fi
    
    log_info "Тестирование завершено"
}

# Показ статуса
show_status() {
    log_info "Статус расширенного аудита Lynis:"
    echo "=================================="
    
    # Systemd сервисы
    echo "Systemd сервисы:"
    sudo systemctl is-active enhanced-lynis-audit.timer 2>/dev/null && echo "enhanced-lynis-audit.timer: активен" || echo "❌ enhanced-lynis-audit.timer: неактивен"
    sudo systemctl is-enabled enhanced-lynis-audit.timer 2>/dev/null && echo "enhanced-lynis-audit.timer: включен" || echo "❌ enhanced-lynis-audit.timer: не включен"
    
    # Зависимости
    echo ""
    echo "Зависимости:"
    command -v lynis &> /dev/null && echo "lynis: установлен" || echo "❌ lynis: не установлен"
    command -v jq &> /dev/null && echo "jq: установлен" || echo "❌ jq: не установлен"
    command -v curl &> /dev/null && echo "curl: установлен" || echo "❌ curl: не установлен"
    command -v docker &> /dev/null && echo "docker: установлен" || echo "❌ docker: не установлен"
    
    # Файлы
    echo ""
    echo "Файлы:"
    [[ -f "/etc/lynis/enhanced_custom.prf" ]] && echo "enhanced_custom.prf: существует" || echo "❌ enhanced_custom.prf: не существует"
    [[ -f "/etc/systemd/system/enhanced-lynis-audit.service" ]] && echo "enhanced-lynis-audit.service: существует" || echo "❌ enhanced-lynis-audit.service: не существует"
    [[ -f "/etc/systemd/system/enhanced-lynis-audit.timer" ]] && echo "enhanced-lynis-audit.timer: существует" || echo "❌ enhanced-lynis-audit.timer: не существует"
    
    # Директории
    echo ""
    echo "Директории:"
    [[ -d "/var/log/lynis/reports" ]] && echo "/var/log/lynis/reports: существует" || echo "❌ /var/log/lynis/reports: не существует"
    [[ -d "/var/lib/lynis-telegram" ]] && echo "/var/lib/lynis-telegram: существует" || echo "❌ /var/lib/lynis-telegram: не существует"
    
    # Логи
    echo ""
    echo "Логи:"
    [[ -f "/var/log/lynis/lynis.log" ]] && echo "lynis.log: существует" || echo "❌ lynis.log: не существует"
    [[ -f "/etc/logrotate.d/lynis" ]] && echo "logrotate: настроен" || echo "❌ logrotate: не настроен"
}

# Показ справки
show_help() {
    echo "Использование: $0 [опции]"
    echo ""
    echo "Опции:"
    echo "  install     Установить расширенный аудит"
    echo "  test        Протестировать систему"
    echo "  status      Показать статус"
    echo "  help        Показать эту справку"
    echo ""
    echo "Примеры:"
    echo "  $0 install    # Полная установка"
    echo "  $0 test       # Тестирование"
    echo "  $0 status     # Статус системы"
}

# Основная функция
main() {
    case "${1:-}" in
        install)
            log_info "Установка расширенного аудита Lynis..."
            check_dependencies
            create_directories
            install_enhanced_profile
            install_systemd_services
            setup_log_rotation
            test_system
            show_status
            log_info "Установка завершена!"
            ;;
        test)
            test_system
            ;;
        status)
            show_status
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

# Запуск
main "$@" 