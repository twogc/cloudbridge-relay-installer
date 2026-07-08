#!/bin/bash

# Скрипт для аккуратной очистки старых k8s конфигураций

set -e

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Переменные
K8S_DIR="/home/ubuntu/cloudbridge-relay-installer/k8s"
BACKUP_DIR="/home/ubuntu/cloudbridge-relay-installer/backup/k8s-deployments"
DRY_RUN=false

# Парсинг аргументов
while getopts "d" opt; do
    case ${opt} in
        d )
            DRY_RUN=true
            ;;
        \? )
            echo "Usage: $0 [-d]"
            echo "  -d  Dry run (показать что будет удалено, но не удалять)"
            exit 1
            ;;
    esac
done

log "🧹 Начинаем очистку старых k8s конфигураций..."

# Создаем резервную копию важных файлов
create_backup() {
    log "Создаем резервную копию важных файлов..."
    
    # Создаем директорию для бэкапа
    mkdir -p "$BACKUP_DIR"
    
    # Копируем актуальные файлы
    cp "$K8S_DIR/deployments/cadence-working.yaml" "$BACKUP_DIR/" 2>/dev/null || true
    cp "$K8S_DIR/deployments/postgres-simple.yaml" "$BACKUP_DIR/" 2>/dev/null || true
    cp "$K8S_DIR/deployments/cloudbridge-relay-deployment.yaml" "$BACKUP_DIR/" 2>/dev/null || true
    cp "$K8S_DIR/services/cadence-nodeport.yaml" "$BACKUP_DIR/" 2>/dev/null || true
    
    success "Резервная копия создана в $BACKUP_DIR"
}

# Функция для безопасного удаления файла
safe_delete() {
    local file="$1"
    local reason="$2"
    
    if [ "$DRY_RUN" = true ]; then
        warning "DRY RUN: Удалили бы $file ($reason)"
    else
        if [ -f "$file" ]; then
            log "Удаляем $file ($reason)"
            rm "$file"
            success "Удален: $file"
        else
            warning "Файл не найден: $file"
        fi
    fi
}

# Удаляем устаревшие Cadence конфигурации
cleanup_cadence_configs() {
    log "🗑️  Очищаем устаревшие Cadence конфигурации..."
    
    # Устаревшие Cadence deployments
    safe_delete "$K8S_DIR/cadence-no-config.yaml" "устаревший, заменен на cadence-working.yaml"
    safe_delete "$K8S_DIR/cadence-schema-init.yaml" "устаревший, схема создается автоматически"
    safe_delete "$K8S_DIR/cadence-schema-setup.yaml" "устаревший, схема создается автоматически"
    safe_delete "$K8S_DIR/cadence-simple-fix.yaml" "устаревший, заменен на cadence-working.yaml"
    safe_delete "$K8S_DIR/cadence-simple-working.yaml" "устаревший, заменен на cadence-working.yaml"
    
    # Устаревшие Cadence deployments в папке deployments
    safe_delete "$K8S_DIR/deployments/cadence-deployment-simple.yaml" "устаревший, заменен на cadence-working.yaml"
    safe_delete "$K8S_DIR/deployments/cadence-postgres.yaml" "устаревший, заменен на cadence-working.yaml"
    safe_delete "$K8S_DIR/deployments/cadence-simple-postgres.yaml" "устаревший, заменен на cadence-working.yaml"
    safe_delete "$K8S_DIR/deployments/cadence-sqlite.yaml" "устаревший, заменен на cadence-working.yaml"
}

# Удаляем устаревшие Cassandra конфигурации
cleanup_cassandra_configs() {
    log "🗑️  Очищаем устаревшие Cassandra конфигурации..."
    
    # Устаревшие Cassandra deployments
    safe_delete "$K8S_DIR/deployments/cassandra-deployment.yaml" "устаревший, Cassandra не используется"
    safe_delete "$K8S_DIR/deployments/cassandra-minimal.yaml" "устаревший, Cassandra не используется"
    safe_delete "$K8S_DIR/deployments/cassandra-simple.yaml" "устаревший, Cassandra не используется"
}

# Удаляем дублирующиеся конфигурации
cleanup_duplicate_configs() {
    log "🗑️  Очищаем дублирующиеся конфигурации..."
    
    # Дублирующиеся файлы в корне k8s
    safe_delete "$K8S_DIR/deployment.yaml" "дублирует deployments/cloudbridge-relay-deployment.yaml"
    safe_delete "$K8S_DIR/simple-deployment.yaml" "дублирует deployments/cloudbridge-relay-deployment.yaml"
    safe_delete "$K8S_DIR/service.yaml" "дублирует services/cloudbridge-services.yaml"
    safe_delete "$K8S_DIR/service-updated.yaml" "дублирует services/cloudbridge-services.yaml"
    safe_delete "$K8S_DIR/configmap.yaml" "дублирует configmaps/cloudbridge-config.yaml"
}

# Удаляем устаревшие мониторинг конфигурации
cleanup_monitoring_configs() {
    log "🗑️  Очищаем устаревшие мониторинг конфигурации..."
    
    # Устаревшие Prometheus конфигурации
    safe_delete "$K8S_DIR/prometheus-config-updated.yaml" "дублирует monitoring/prometheus-config.yaml"
    safe_delete "$K8S_DIR/monitoring-config.yaml" "дублирует monitoring/prometheus-config.yaml"
}

# Удаляем устаревшие AI конфигурации
cleanup_ai_configs() {
    log "🗑️  Очищаем устаревшие AI конфигурации..."
    
    # Устаревшие AI файлы в корне k8s
    safe_delete "$K8S_DIR/ai-analytics-deployment.yaml" "дублирует deployments/ai-service-deployment.yaml"
    safe_delete "$K8S_DIR/ai-analytics-endpoints.yaml" "дублирует services/ai-service.yaml"
    safe_delete "$K8S_DIR/ai-report-cronjob.yaml" "устаревший cronjob"
    safe_delete "$K8S_DIR/ai-service-endpoints.yaml" "дублирует services/ai-service.yaml"
    safe_delete "$K8S_DIR/ai-service-servicemonitor.yaml" "дублирует services/ai-service.yaml"
    safe_delete "$K8S_DIR/ai-telegram-config.yaml" "устаревший конфиг"
    safe_delete "$K8S_DIR/ai-telegram-scheduler.yaml" "устаревший scheduler"
    safe_delete "$K8S_DIR/ai-telegram-secret.yaml" "устаревший secret"
}

# Удаляем устаревшие тестовые файлы
cleanup_test_files() {
    log "🗑️  Очищаем устаревшие тестовые файлы..."
    
    safe_delete "$K8S_DIR/wg-test-pod.yaml" "тестовый файл"
    safe_delete "$K8S_DIR/init-cadence-schema.sh" "устаревший скрипт"
}

# Очищаем пустые директории
cleanup_empty_dirs() {
    log "🗑️  Очищаем пустые директории..."
    
    # Находим и удаляем пустые директории
    find "$K8S_DIR" -type d -empty -delete 2>/dev/null || true
}

# Создаем отчет о том, что осталось
create_report() {
    log "Создаем отчет о оставшихся файлах..."
    
    local report_file="$BACKUP_DIR/cleanup-report-$(date +%Y%m%d-%H%M%S).txt"
    
    echo "=== ОТЧЕТ О ОЧИСТКЕ K8S КОНФИГУРАЦИЙ ===" > "$report_file"
    echo "Дата: $(date)" >> "$report_file"
    echo "Режим: $([ "$DRY_RUN" = true ] && echo "DRY RUN" || echo "РЕАЛЬНОЕ УДАЛЕНИЕ")" >> "$report_file"
    echo "" >> "$report_file"
    
    echo "=== ОСТАВШИЕСЯ ФАЙЛЫ ===" >> "$report_file"
    find "$K8S_DIR" -type f -name "*.yaml" -o -name "*.yml" | sort >> "$report_file"
    
    echo "" >> "$report_file"
    echo "=== СТРУКТУРА ДИРЕКТОРИЙ ===" >> "$report_file"
    tree "$K8S_DIR" >> "$report_file" 2>/dev/null || find "$K8S_DIR" -type d | sort >> "$report_file"
    
    success "Отчет создан: $report_file"
}

# Основная функция
main() {
    log "Начинаем очистку k8s конфигураций..."
    
    if [ "$DRY_RUN" = true ]; then
        warning "РЕЖИМ DRY RUN - файлы не будут удалены"
    fi
    
    # Создаем резервную копию
    create_backup
    
    # Очищаем конфигурации
    cleanup_cadence_configs
    cleanup_cassandra_configs
    cleanup_duplicate_configs
    cleanup_monitoring_configs
    cleanup_ai_configs
    cleanup_test_files
    cleanup_empty_dirs
    
    # Создаем отчет
    create_report
    
    if [ "$DRY_RUN" = true ]; then
        warning "DRY RUN завершен. Для реального удаления запустите без флага -d"
    else
        success "Очистка k8s конфигураций завершена!"
        log "Резервная копия сохранена в: $BACKUP_DIR"
        log "Отчет создан: $BACKUP_DIR/cleanup-report-*.txt"
    fi
}

# Запуск
main "$@"
