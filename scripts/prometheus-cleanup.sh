#!/bin/bash

# Prometheus Storage Cleanup Script
# Автоматическая очистка старых данных Prometheus для оптимизации дискового пространства

set -e

NAMESPACE="cloudbridge-monitoring"
PROMETHEUS_POD="prometheus-optimized"
CLEANUP_THRESHOLD_GB=35  # Очистка при достижении 35GB
MAX_RETENTION_HOURS=20   # Максимальное время хранения 20 часов

echo "🧹 Starting Prometheus storage cleanup..."

# Функция для получения размера хранилища Prometheus
get_storage_size() {
    kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- du -sh /prometheus/ 2>/dev/null | awk '{print $1}' | sed 's/M//' | sed 's/G//' | sed 's/K//'
}

# Функция для получения размера в GB
get_storage_size_gb() {
    local size=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- du -sb /prometheus/ 2>/dev/null | awk '{print $1}')
    echo "scale=2; $size / 1024 / 1024 / 1024" | bc
}

# Функция для очистки старых блоков
cleanup_old_blocks() {
    echo "🗑️  Cleaning up old blocks..."
    
    # Получаем список старых блоков
    local old_blocks=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- find /prometheus/ -name "*-*" -type d -mtime +1 2>/dev/null | head -5)
    
    if [ -n "$old_blocks" ]; then
        echo "Found old blocks to clean up:"
        echo "$old_blocks"
        
        # Удаляем старые блоки
        for block in $old_blocks; do
            echo "Removing block: $block"
            kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- rm -rf "$block"
        done
        
        echo "✅ Old blocks cleaned up successfully"
    else
        echo "ℹ️  No old blocks found to clean up"
    fi
}

# Функция для очистки WAL файлов
cleanup_wal_files() {
    echo "🗑️  Cleaning up old WAL files..."
    
    # Получаем список старых WAL файлов
    local old_wal_files=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- find /prometheus/wal/ -name "*.log" -mtime +1 2>/dev/null | head -10)
    
    if [ -n "$old_wal_files" ]; then
        echo "Found old WAL files to clean up:"
        echo "$old_wal_files"
        
        # Удаляем старые WAL файлы
        for wal_file in $old_wal_files; do
            echo "Removing WAL file: $wal_file"
            kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- rm -f "$wal_file"
        done
        
        echo "✅ Old WAL files cleaned up successfully"
    else
        echo "ℹ️  No old WAL files found to clean up"
    fi
}

# Функция для принудительной очистки через API
force_cleanup() {
    echo "🔄 Force cleanup via Prometheus API..."
    
    # Отправляем запрос на очистку через API
    kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- curl -X POST http://localhost:9090/api/v1/admin/tsdb/clean_tombstones 2>/dev/null || true
    
    echo "✅ Force cleanup completed"
}

# Функция для мониторинга метрик
monitor_metrics() {
    echo "📊 Current Prometheus storage metrics:"
    
    # Размер хранилища
    local size_gb=$(get_storage_size_gb)
    echo "Storage size: ${size_gb}GB"
    
    # Количество блоков
    local block_count=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- find /prometheus/ -name "*-*" -type d 2>/dev/null | wc -l)
    echo "Number of blocks: $block_count"
    
    # Количество WAL файлов
    local wal_count=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- find /prometheus/wal/ -name "*.log" 2>/dev/null | wc -l)
    echo "Number of WAL files: $wal_count"
    
    # Время последнего блока
    local last_block_time=$(kubectl exec -n $NAMESPACE $PROMETHEUS_POD -- ls -lt /prometheus/ | grep "^-" | head -1 | awk '{print $6, $7, $8}')
    echo "Last block time: $last_block_time"
}

# Основная логика
main() {
    echo "🚀 Prometheus Storage Cleanup Script"
    echo "=================================="
    
    # Проверяем, что под Prometheus запущен
    if ! kubectl get pod -n $NAMESPACE $PROMETHEUS_POD >/dev/null 2>&1; then
        echo "❌ Prometheus pod not found: $PROMETHEUS_POD"
        exit 1
    fi
    
    # Получаем текущий размер хранилища
    local current_size_gb=$(get_storage_size_gb)
    echo "Current storage size: ${current_size_gb}GB"
    
    # Проверяем, нужна ли очистка
    if (( $(echo "$current_size_gb > $CLEANUP_THRESHOLD_GB" | bc -l) )); then
        echo "⚠️  Storage size (${current_size_gb}GB) exceeds threshold (${CLEANUP_THRESHOLD_GB}GB)"
        echo "Starting cleanup process..."
        
        # Выполняем очистку
        cleanup_old_blocks
        cleanup_wal_files
        force_cleanup
        
        # Проверяем результат
        local new_size_gb=$(get_storage_size_gb)
        local saved_gb=$(echo "scale=2; $current_size_gb - $new_size_gb" | bc)
        
        echo "✅ Cleanup completed!"
        echo "Storage size before: ${current_size_gb}GB"
        echo "Storage size after: ${new_size_gb}GB"
        echo "Space saved: ${saved_gb}GB"
        
    else
        echo "ℹ️  Storage size (${current_size_gb}GB) is within acceptable limits (${CLEANUP_THRESHOLD_GB}GB)"
        echo "No cleanup needed at this time"
    fi
    
    # Показываем текущие метрики
    echo ""
    monitor_metrics
    
    echo ""
    echo "🎉 Prometheus storage cleanup script completed!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    "monitor")
        monitor_metrics
        ;;
    "force")
        echo "🔄 Force cleanup mode..."
        cleanup_old_blocks
        cleanup_wal_files
        force_cleanup
        monitor_metrics
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [monitor|force|help]"
        echo "  monitor - Show current storage metrics"
        echo "  force   - Force cleanup regardless of size"
        echo "  help    - Show this help message"
        echo "  (no args) - Run automatic cleanup if needed"
        ;;
    *)
        main
        ;;
esac
