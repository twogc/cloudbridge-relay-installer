#!/usr/bin/env python3

import os
import re
import glob

def fix_activity_options(file_path):
    """Исправляет ActivityOptions в файле, добавляя ScheduleToStartTimeout"""
    print(f"Исправление файла: {file_path}")
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Паттерн для поиска ActivityOptions без ScheduleToStartTimeout
    pattern = r'(activityOptions := workflow\.ActivityOptions\{)\s*\n\s*(StartToCloseTimeout:)'
    
    # Замена с добавлением ScheduleToStartTimeout
    replacement = r'\1\n\t\tScheduleToStartTimeout: 1 * time.Minute,\n\t\t\2'
    
    # Применяем замену
    new_content = re.sub(pattern, replacement, content)
    
    # Если файл изменился, сохраняем его
    if new_content != content:
        # Создаем резервную копию
        backup_path = file_path + '.backup'
        with open(backup_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        # Сохраняем исправленный файл
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        
        print(f"  ✓ Файл исправлен, резервная копия: {backup_path}")
        return True
    else:
        print(f"  - Файл не требует изменений")
        return False

def main():
    """Основная функция"""
    print("Исправление таймаутов в Cadence workflows...")
    
    # Находим все .go файлы в папке workflows
    workflow_dir = "internal/cadence/workflows"
    go_files = glob.glob(os.path.join(workflow_dir, "*.go"))
    
    # Исключаем тестовые файлы
    go_files = [f for f in go_files if not f.endswith('_test.go')]
    
    fixed_count = 0
    for file_path in go_files:
        if fix_activity_options(file_path):
            fixed_count += 1
    
    print(f"\nИсправление завершено! Обработано файлов: {len(go_files)}, исправлено: {fixed_count}")

if __name__ == "__main__":
    main()
