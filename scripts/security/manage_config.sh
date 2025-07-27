#!/bin/bash

# CloudBridge Relay - Configuration Management Script
# Управление конфигурацией мониторинга безопасности

set -e

# Пути к файлам конфигурации
SYSTEM_CONFIG="/etc/cloudbridge/security_monitor.conf"
USER_CONFIG="$HOME/.cloudbridge/security_monitor.conf"
BACKUP_CONFIG="$HOME/.cloudbridge/security_monitor.conf.backup"

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

# Создание директории пользовательской конфигурации
create_user_config_dir() {
    mkdir -p "$(dirname "$USER_CONFIG")"
}

# Синхронизация конфигурации с системной
sync_to_system() {
    if [[ ! -f "$USER_CONFIG" ]]; then
        log_error "Пользовательская конфигурация не найдена: $USER_CONFIG"
        return 1
    fi
    
    log_info "Синхронизация конфигурации с системной..."
    sudo cp "$USER_CONFIG" "$SYSTEM_CONFIG"
    sudo chown root:root "$SYSTEM_CONFIG"
    sudo chmod 644 "$SYSTEM_CONFIG"
    log_info "Конфигурация синхронизирована"
}

# Синхронизация конфигурации с пользовательской
sync_from_system() {
    if [[ ! -f "$SYSTEM_CONFIG" ]]; then
        log_error "Системная конфигурация не найдена: $SYSTEM_CONFIG"
        return 1
    fi
    
    log_info "Синхронизация конфигурации с пользовательской..."
    create_user_config_dir
    cp "$SYSTEM_CONFIG" "$USER_CONFIG"
    chmod 644 "$USER_CONFIG"
    log_info "Конфигурация синхронизирована"
}

# Создание резервной копии
create_backup() {
    if [[ -f "$USER_CONFIG" ]]; then
        log_info "Создание резервной копии..."
        cp "$USER_CONFIG" "$BACKUP_CONFIG"
        log_info "Резервная копия создана: $BACKUP_CONFIG"
    else
        log_warn "Нет конфигурации для резервного копирования"
    fi
}

# Восстановление из резервной копии
restore_backup() {
    if [[ -f "$BACKUP_CONFIG" ]]; then
        log_info "Восстановление из резервной копии..."
        cp "$BACKUP_CONFIG" "$USER_CONFIG"
        log_info "Конфигурация восстановлена"
    else
        log_error "Резервная копия не найдена: $BACKUP_CONFIG"
        return 1
    fi
}

# Сравнение конфигураций
compare_configs() {
    if [[ ! -f "$SYSTEM_CONFIG" ]]; then
        log_error "Системная конфигурация не найдена"
        return 1
    fi
    
    if [[ ! -f "$USER_CONFIG" ]]; then
        log_error "Пользовательская конфигурация не найдена"
        return 1
    fi
    
    log_info "Сравнение конфигураций..."
    if diff "$SYSTEM_CONFIG" "$USER_CONFIG" > /dev/null; then
        log_info "Конфигурации идентичны"
    else
        log_warn "Конфигурации различаются:"
        diff "$SYSTEM_CONFIG" "$USER_CONFIG" || true
    fi
}

# Редактирование конфигурации
edit_config() {
    create_user_config_dir
    
    # Если пользовательская конфигурация не существует, копируем системную
    if [[ ! -f "$USER_CONFIG" ]]; then
        if [[ -f "$SYSTEM_CONFIG" ]]; then
            log_info "Копирование системной конфигурации..."
            cp "$SYSTEM_CONFIG" "$USER_CONFIG"
        else
            log_warn "Системная конфигурация не найдена, создаю новую..."
            bash "$(dirname "$0")/security_monitor_config.sh" --create
            cp "$SYSTEM_CONFIG" "$USER_CONFIG"
        fi
    fi
    
    # Создаем резервную копию перед редактированием
    create_backup
    
    # Определяем редактор
    if [[ -z "$EDITOR" ]]; then
        if command -v nano &> /dev/null; then
            EDITOR="nano"
        elif command -v vim &> /dev/null; then
            EDITOR="vim"
        elif command -v vi &> /dev/null; then
            EDITOR="vi"
        else
            log_error "Не найден подходящий редактор. Установите nano, vim или vi"
            return 1
        fi
    fi
    
    log_info "Открытие конфигурации в редакторе: $EDITOR"
    log_info "Файл: $USER_CONFIG"
    $EDITOR "$USER_CONFIG"
    
    log_info "Редактирование завершено"
    log_info "Используйте '$0 --sync' для применения изменений"
}

# Валидация конфигурации
validate_config() {
    local config_file="$1"
    
    if [[ ! -f "$config_file" ]]; then
        log_error "Файл конфигурации не найден: $config_file"
        return 1
    fi
    
    log_info "Валидация конфигурации: $config_file"
    
    # Проверяем синтаксис
    if ! source "$config_file" 2>/dev/null; then
        log_error "Ошибка синтаксиса в конфигурационном файле"
        return 1
    fi
    
    # Проверяем обязательные параметры
    local required_vars=("TELEGRAM_BOT_TOKEN" "TELEGRAM_CHAT_ID")
    local missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "Отсутствуют обязательные параметры: ${missing_vars[*]}"
        return 1
    fi
    
    log_info "✅ Конфигурация корректна"
    return 0
}

# Показать статус конфигурации
show_status() {
    echo "📋 Статус конфигурации мониторинга безопасности"
    echo "=============================================="
    
    if [[ -f "$SYSTEM_CONFIG" ]]; then
        echo "✅ Системная конфигурация: $SYSTEM_CONFIG"
        echo "   Размер: $(du -h "$SYSTEM_CONFIG" | cut -f1)"
        echo "   Владелец: $(stat -c '%U:%G' "$SYSTEM_CONFIG")"
        echo "   Права: $(stat -c '%a' "$SYSTEM_CONFIG")"
    else
        echo "❌ Системная конфигурация не найдена"
    fi
    
    if [[ -f "$USER_CONFIG" ]]; then
        echo "✅ Пользовательская конфигурация: $USER_CONFIG"
        echo "   Размер: $(du -h "$USER_CONFIG" | cut -f1)"
        echo "   Владелец: $(stat -c '%U:%G' "$USER_CONFIG")"
        echo "   Права: $(stat -c '%a' "$USER_CONFIG")"
    else
        echo "❌ Пользовательская конфигурация не найдена"
    fi
    
    if [[ -f "$BACKUP_CONFIG" ]]; then
        echo "✅ Резервная копия: $BACKUP_CONFIG"
        echo "   Размер: $(du -h "$BACKUP_CONFIG" | cut -f1)"
        echo "   Дата: $(stat -c '%y' "$BACKUP_CONFIG")"
    else
        echo "❌ Резервная копия не найдена"
    fi
    
    echo
    echo "🔄 Синхронизация:"
    if [[ -f "$SYSTEM_CONFIG" && -f "$USER_CONFIG" ]]; then
        if diff "$SYSTEM_CONFIG" "$USER_CONFIG" > /dev/null; then
            echo "✅ Конфигурации синхронизированы"
        else
            echo "⚠️  Конфигурации различаются"
        fi
    else
        echo "❌ Невозможно сравнить конфигурации"
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "CloudBridge Relay - Configuration Management"
        echo
        echo "Использование: $0 [опции]"
        echo
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --edit         Редактировать конфигурацию"
        echo "  --sync         Синхронизировать с системной конфигурацией"
        echo "  --sync-from    Синхронизировать с пользовательской конфигурацией"
        echo "  --backup       Создать резервную копию"
        echo "  --restore      Восстановить из резервной копии"
        echo "  --compare      Сравнить конфигурации"
        echo "  --validate     Проверить конфигурацию"
        echo "  --status       Показать статус конфигурации"
        echo
        echo "Примеры:"
        echo "  $0 --edit              Редактировать конфигурацию"
        echo "  $0 --sync              Применить изменения"
        echo "  $0 --status            Показать статус"
        echo "  $0 --validate          Проверить конфигурацию"
        ;;
    --edit)
        edit_config
        ;;
    --sync)
        sync_to_system
        ;;
    --sync-from)
        sync_from_system
        ;;
    --backup)
        create_backup
        ;;
    --restore)
        restore_backup
        ;;
    --compare)
        compare_configs
        ;;
    --validate)
        if [[ -f "$USER_CONFIG" ]]; then
            validate_config "$USER_CONFIG"
        elif [[ -f "$SYSTEM_CONFIG" ]]; then
            validate_config "$SYSTEM_CONFIG"
        else
            log_error "Конфигурация не найдена"
            exit 1
        fi
        ;;
    --status)
        show_status
        ;;
    *)
        echo "Используйте --help для получения справки"
        exit 1
        ;;
esac 