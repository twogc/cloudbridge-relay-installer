#!/bin/bash

# CloudBridge Relay - Quick Security Audit
# Быстрый запуск аудита безопасности

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

# Проверка аргументов
AUDIT_TYPE=${1:-quick}
OUTPUT_DIR=${2:-logs}

# Создание директории для логов
mkdir -p "$OUTPUT_DIR"

log "Запуск быстрого аудита безопасности CloudBridge Relay..."
log "Тип аудита: $AUDIT_TYPE"
log "Директория логов: $OUTPUT_DIR"

# Очистка старых lock-файлов
if [[ -f /var/run/lynis_telegram.lock ]]; then
    rm -f /var/run/lynis_telegram.lock
    log "Удалён старый lock-файл"
fi

# Проверка, что нет запущенных процессов Lynis
if pgrep -f "lynis.*audit" > /dev/null; then
    warn "Обнаружены запущенные процессы Lynis, останавливаю..."
    sudo pkill -f "lynis.*audit" || true
    sleep 2
fi

# Генерация имени файла лога
LOG_FILE="$OUTPUT_DIR/lynis-$(date +%F-%H%M).log"

log "Начало аудита..."
log "Лог файл: $LOG_FILE"

# Запуск Lynis без таймаута
case $AUDIT_TYPE in
    "quick")
        log "Выполнение быстрого аудита..."
        lynis audit system --quick > "$LOG_FILE" 2>&1
        ;;
    "full")
        log "Выполнение полного аудита..."
        lynis audit system > "$LOG_FILE" 2>&1
        ;;
    "docker")
        log "Аудит Docker окружения..."
        lynis audit system --tests-from-group docker > "$LOG_FILE" 2>&1
        ;;
    *)
        error "Неизвестный тип аудита: $AUDIT_TYPE"
        echo "Доступные типы: quick, full, docker"
        exit 1
        ;;
esac

AUDIT_EXIT_CODE=$?

if [[ $AUDIT_EXIT_CODE -ne 0 ]]; then
    error "Аудит завершился с ошибкой (exit code: $AUDIT_EXIT_CODE)"
    exit 1
fi

log "Аудит успешно завершён!"

# Анализ результатов
log "Анализ результатов..."

# Извлечение оценок
SCORE=$(grep "Hardening index" "$LOG_FILE" | tail -1 | grep -o '[0-9]\+' | head -1 || echo "0")
WARNINGS=$(grep -c "Warning:" "$LOG_FILE" || echo "0")
SUGGESTIONS=$(grep -c "Suggestion:" "$LOG_FILE" || echo "0")

log "Результаты аудита:"
log "  Оценка безопасности: $SCORE/100"
log "  Предупреждения: $WARNINGS"
log "  Рекомендации: $SUGGESTIONS"

# Оценка безопасности
if [[ $SCORE -ge 80 ]]; then
    log "✅ Отличная безопасность!"
elif [[ $SCORE -ge 60 ]]; then
    log "⚠️  Хорошая безопасность, но есть возможности для улучшения"
else
    log "🚨 Требуется внимание к безопасности!"
fi

# Показываем топ рекомендаций
echo
log "Топ рекомендаций для улучшения:"
grep -A 1 "Suggestion:" "$LOG_FILE" | head -10 | sed 's/^/  /'

log "Аудит завершён. Лог сохранён в: $LOG_FILE" 