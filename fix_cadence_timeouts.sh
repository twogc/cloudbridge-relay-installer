#!/bin/bash

# Скрипт для исправления таймаутов в Cadence workflows

echo "Исправление таймаутов в Cadence workflows..."

# Функция для добавления ScheduleToStartTimeout
fix_activity_options() {
    local file="$1"
    echo "Исправление файла: $file"
    
    # Заменяем все вхождения ActivityOptions без ScheduleToStartTimeout
    sed -i 's/activityOptions := workflow\.ActivityOptions{/activityOptions := workflow.ActivityOptions{\n\t\tScheduleToStartTimeout: 1 * time.Minute,/' "$file"
    
    # Исправляем отступы
    sed -i 's/^[[:space:]]*StartToCloseTimeout:/&\n\t\tScheduleToStartTimeout: 1 * time.Minute,/' "$file"
}

# Список файлов для исправления
files=(
    "internal/cadence/workflows/ai_optimization.go"
    "internal/cadence/workflows/p2p_connection.go"
    "internal/cadence/workflows/p2p_quic.go"
    "internal/cadence/workflows/wireguard.go"
    "internal/cadence/workflows/security.go"
    "internal/cadence/workflows/quantum.go"
    "internal/cadence/workflows/scaling.go"
)

# Исправляем каждый файл
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        # Создаем резервную копию
        cp "$file" "$file.backup"
        
        # Исправляем файл
        fix_activity_options "$file"
        
        echo "Файл $file исправлен"
    else
        echo "Файл $file не найден"
    fi
done

echo "Исправление завершено!"
