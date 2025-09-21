#!/usr/bin/env python3

import os
import re

def remove_wireguard_references(file_path):
    """Удаляет все упоминания wireguard из файла"""
    print(f"Обработка файла: {file_path}")
    
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # Удаляем импорты wireguard
    content = re.sub(r'^\s*"[^"]*wireguard[^"]*"\s*\n', '', content, flags=re.MULTILINE)
    
    # Удаляем переменные wireguard
    content = re.sub(r'^\s*wireguard\.[A-Za-z]+\s+\*wireguard\.[A-Za-z]+\s*\n', '', content, flags=re.MULTILINE)
    content = re.sub(r'^\s*var\s+wireguard\.[A-Za-z]+\s+\*wireguard\.[A-Za-z]+\s*\n', '', content, flags=re.MULTILINE)
    
    # Удаляем инициализацию wireguard
    content = re.sub(r'^\s*wireguard\.[A-Za-z]+\s*:=\s*wireguard\.[A-Za-z]+\([^)]*\)\s*\n', '', content, flags=re.MULTILINE)
    
    # Удаляем комментарии о WireGuard
    content = re.sub(r'^\s*//.*[Ww]ire[Gg]uard.*\n', '', content, flags=re.MULTILINE)
    
    # Удаляем пустые строки, оставшиеся после удаления
    content = re.sub(r'\n\s*\n\s*\n', '\n\n', content)
    
    if content != original_content:
        # Создаем резервную копию
        backup_path = file_path + '.backup'
        with open(backup_path, 'w', encoding='utf-8') as f:
            f.write(original_content)
        
        # Сохраняем измененный файл
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        
        print(f"  ✓ Файл обновлен, резервная копия: {backup_path}")
        return True
    else:
        print(f"  - Файл не требует изменений")
        return False

def main():
    """Основная функция"""
    print("Удаление упоминаний WireGuard из кода...")
    
    # Файлы для обработки
    files_to_process = [
        "cmd/relay/main.go",
        "internal/p2p/manager.go",
        "internal/p2p/interfaces.go",
        "internal/p2p/manager_test.go",
        "internal/integration/module_integration_test.go",
    ]
    
    processed_count = 0
    for file_path in files_to_process:
        if os.path.exists(file_path):
            if remove_wireguard_references(file_path):
                processed_count += 1
        else:
            print(f"Файл не найден: {file_path}")
    
    print(f"\nОбработка завершена! Обработано файлов: {len(files_to_process)}, изменено: {processed_count}")

if __name__ == "__main__":
    main()
