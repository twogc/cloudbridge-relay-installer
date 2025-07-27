#!/bin/bash

# CloudBridge Relay - Security Monitor Configuration
# Конфигурация фильтров для мониторинга безопасности

# Файл конфигурации
CONFIG_FILE="/etc/cloudbridge/security_monitor.conf"

# Создание директории конфигурации
mkdir -p "$(dirname "$CONFIG_FILE")"

# Функция создания конфигурации по умолчанию
create_default_config() {
    cat > "$CONFIG_FILE" << 'EOF'
# CloudBridge Relay - Security Monitor Configuration
# Конфигурация фильтров для мониторинга безопасности

# Основные настройки
TELEGRAM_BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
TELEGRAM_CHAT_ID="-1002403555405"
CHECK_INTERVAL=300

# Фильтры для исключения некритичных ошибок
# Формат: PATTERN_NAME="regex_pattern"
DOCKER_SHOULDRESTART_FILTER="dockerd.*ShouldRestart failed"
DOCKER_RESTART_CANCELED_FILTER="dockerd.*restart canceled"
DOCKER_PLUGIN_ERROR_FILTER="dockerd.*Error handling plugin refcount operation"
DOCKER_LEGACY_PLUGIN_FILTER="dockerd.*legacy plugin: Unable to find plugin without name"
DOCKER_CONTAINER_RESTART_FILTER="dockerd.*container will not be restarted"
DOCKER_MANUALLY_STOPPED_FILTER="dockerd.*hasBeenManuallyStopped=true"
DOCKER_DAEMON_SHUTDOWN_FILTER="dockerd.*daemonShuttingDown=false"
CLOUDFLARED_FILTER="cloudflared"
FRESHCLAM_429_FILTER="freshclam.*error code 429"
FRESHCLAM_403_FILTER="freshclam.*error code 403"
NETWORKD_DISPATCHER_FILTER="networkd-dispatcher.*Failed to get interface"
NETWORKD_UNKNOWN_INTERFACE_FILTER="networkd-dispatcher.*Unknown interface index"
RELAY_SECURITY_AUDIT_FILTER="relay-security-audit.*Fatal error"
DOCKER_REGISTRY_ERROR_FILTER="dockerd.*Error getting v2 registry"
DOCKER_ENDPOINT_ERROR_FILTER="dockerd.*Attempting next endpoint for pull after error"
DOCKER_FORBIDDEN_FILTER="dockerd.*denied: 403 Forbidden"
DOCKER_NO_HOST_FILTER="dockerd.*no such host"

# Дополнительные фильтры для системных сообщений
EMERGENCY_SERVICE_FILTER="emergency\\.service"
ICMP_IGNORE_FILTER="icmp_ignore_bogus_error_responses"
AIDE_INSTALL_FILTER="Install a file integrity tool"
DHCP_BLOCK_FILTER="SPT=67 DPT=68"
AIDE_DATABASE_FILTER="Check the AIDE database"

# Пороговые значения
DISK_USAGE_THRESHOLD=85
MEMORY_USAGE_THRESHOLD=85
SSH_FAILURES_THRESHOLD=50
ROOT_PROCESSES_THRESHOLD=150
SYSTEM_UPDATES_THRESHOLD=10

# Настройки логирования
LOG_FILE="logs/security-monitor.log"
ALERT_FILE="logs/security-alerts.log"
STATE_DIR="/var/lib/security-monitor"

# Настройки мониторинга
ENABLE_FAIL2BAN_MONITORING=true
ENABLE_CRITICAL_SERVICES_MONITORING=true
ENABLE_DISK_SPACE_MONITORING=true
ENABLE_SYSTEM_UPDATES_MONITORING=true
ENABLE_SUSPICIOUS_ACTIVITY_MONITORING=true
ENABLE_SSL_CERTIFICATE_MONITORING=true
ENABLE_CRITICAL_LOGS_MONITORING=true
ENABLE_FIREWALL_LOGS_MONITORING=true

# Фильтры UFW блокировок
FILTER_SSH_BLOCKS=true
FILTER_DHCP_BLOCKS=true
EOF
}

# Функция загрузки конфигурации
load_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "Конфигурационный файл не найден. Создаю файл по умолчанию..."
        create_default_config
    fi
    
    # Загружаем конфигурацию
    source "$CONFIG_FILE"
}

# Функция показа текущей конфигурации
show_config() {
    if [[ -f "$CONFIG_FILE" ]]; then
        echo "📋 Текущая конфигурация мониторинга безопасности:"
        echo "================================================"
        cat "$CONFIG_FILE"
    else
        echo "❌ Конфигурационный файл не найден"
    fi
}

# Функция редактирования конфигурации
edit_config() {
    if [[ -z "$EDITOR" ]]; then
        EDITOR="nano"
    fi
    
    if [[ ! -f "$CONFIG_FILE" ]]; then
        create_default_config
    fi
    
    $EDITOR "$CONFIG_FILE"
    echo "✅ Конфигурация обновлена"
}

# Функция сброса конфигурации
reset_config() {
    echo "⚠️  Вы уверены, что хотите сбросить конфигурацию? (y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        create_default_config
        echo "✅ Конфигурация сброшена к значениям по умолчанию"
    else
        echo "❌ Сброс отменен"
    fi
}

# Функция проверки конфигурации
validate_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "❌ Конфигурационный файл не найден"
        return 1
    fi
    
    # Проверяем синтаксис конфигурации
    if ! source "$CONFIG_FILE" 2>/dev/null; then
        echo "❌ Ошибка синтаксиса в конфигурационном файле"
        return 1
    fi
    
    # Проверяем обязательные параметры
    if [[ -z "$TELEGRAM_BOT_TOKEN" || -z "$TELEGRAM_CHAT_ID" ]]; then
        echo "❌ Не настроены параметры Telegram"
        return 1
    fi
    
    echo "✅ Конфигурация корректна"
    return 0
}

# Функция получения списка фильтров
get_filters() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "❌ Конфигурационный файл не найден"
        return 1
    fi
    
    echo "🔍 Доступные фильтры:"
    echo "===================="
    grep -E "^[A-Z_]+_FILTER=" "$CONFIG_FILE" | while IFS='=' read -r name value; do
        echo "• ${name%_FILTER}: $value"
    done
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Security Monitor Configuration"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --show         Показать текущую конфигурацию"
        echo "  --edit         Редактировать конфигурацию"
        echo "  --reset        Сбросить конфигурацию к значениям по умолчанию"
        echo "  --validate     Проверить корректность конфигурации"
        echo "  --filters      Показать список фильтров"
        echo "  --create       Создать конфигурацию по умолчанию"
        echo
        echo "Примеры:"
        echo "  $0 --show              Показать конфигурацию"
        echo "  $0 --edit              Редактировать конфигурацию"
        echo "  $0 --reset             Сбросить конфигурацию"
        echo "  $0 --validate          Проверить конфигурацию"
        echo "  $0 --filters           Показать фильтры"
        ;;
    --show)
        show_config
        ;;
    --edit)
        edit_config
        ;;
    --reset)
        reset_config
        ;;
    --validate)
        validate_config
        ;;
    --filters)
        get_filters
        ;;
    --create)
        create_default_config
        echo "✅ Конфигурация создана: $CONFIG_FILE"
        ;;
    *)
        echo "Используйте --help для получения справки"
        exit 1
        ;;
esac 