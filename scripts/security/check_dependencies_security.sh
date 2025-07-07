#!/bin/bash

# CloudBridge Relay - Security Dependency Check Script
# Проверка безопасности Python зависимостей с помощью safety

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка наличия safety
check_safety() {
    if ! command -v safety &> /dev/null; then
        print_error "safety не установлен. Устанавливаем..."
        pip3 install safety==3.5.2
        export PATH="$HOME/.local/bin:$PATH"
    fi
    print_success "safety установлен и готов к использованию"
}

# Проверка файла зависимостей
check_requirements_file() {
    local file=$1
    if [[ -f "$file" ]]; then
        print_status "Проверяем файл: $file"
        return 0
    else
        print_warning "Файл не найден: $file"
        return 1
    fi
}

# Запуск safety для файла
run_safety_check() {
    local file=$1
    local output_file=$2
    
    print_status "Запускаем safety scan для $file..."
    
    # Создаем временный файл для вывода
    local temp_output=$(mktemp)
    
    # Запускаем safety scan с подробным выводом
    if safety scan -r "$file" --output json > "$temp_output" 2>&1; then
        print_success "Проверка безопасности завершена для $file"
        
        # Проверяем, есть ли уязвимости
        local vuln_count=$(jq '.vulnerabilities | length' "$temp_output" 2>/dev/null || echo "0")
        
        if [[ "$vuln_count" -gt 0 ]]; then
            print_warning "Найдено $vuln_count уязвимостей в $file"
            
            # Сохраняем детальный отчет
            if [[ -n "$output_file" ]]; then
                safety scan -r "$file" --output json > "$output_file"
                print_status "Детальный отчет сохранен в: $output_file"
            fi
            
            # Показываем краткую сводку
            echo "=== СВОДКА УЯЗВИМОСТЕЙ ==="
            jq -r '.vulnerabilities[] | "\(.package_name) \(.installed_version) -> \(.vulnerable_spec): \(.advisory)"' "$temp_output" 2>/dev/null || echo "Не удалось проанализировать JSON"
        else
            print_success "Уязвимостей не найдено в $file"
        fi
    else
        print_error "Ошибка при проверке $file"
        cat "$temp_output"
    fi
    
    # Удаляем временный файл
    rm -f "$temp_output"
}

# Основная функция проверки
main() {
    echo "=========================================="
    echo "CloudBridge Relay - Security Audit"
    echo "Проверка безопасности зависимостей"
    echo "=========================================="
    echo ""
    
    # Проверяем safety
    check_safety
    
    # Создаем директорию для отчетов
    local reports_dir="security_reports"
    mkdir -p "$reports_dir"
    
    # Список файлов для проверки
    local requirements_files=(
        "ml/requirements.txt"
        "siem/requirements.txt"
    )
    
    local total_vulnerabilities=0
    local checked_files=0
    
    # Проверяем каждый файл
    for file in "${requirements_files[@]}"; do
        if check_requirements_file "$file"; then
            local output_file="$reports_dir/$(basename "$file" .txt)_security_report.json"
            run_safety_check "$file" "$output_file"
            ((checked_files++))
        fi
    done
    
    echo ""
    echo "=========================================="
    echo "ИТОГОВАЯ СВОДКА"
    echo "=========================================="
    print_status "Проверено файлов: $checked_files"
    
    # Проверяем общее количество уязвимостей
    local total_vulns=0
    for report in "$reports_dir"/*_security_report.json; do
        if [[ -f "$report" ]]; then
            local vuln_count=$(jq '.vulnerabilities | length' "$report" 2>/dev/null || echo "0")
            total_vulns=$((total_vulns + vuln_count))
        fi
    done
    
    if [[ $total_vulns -gt 0 ]]; then
        print_warning "Общее количество уязвимостей: $total_vulns"
        print_status "Детальные отчеты сохранены в директории: $reports_dir"
    else
        print_success "Уязвимостей не найдено!"
    fi
    
    echo ""
    echo "=========================================="
    echo "РЕКОМЕНДАЦИИ"
    echo "=========================================="
    echo "1. Регулярно обновляйте зависимости"
    echo "2. Используйте виртуальные окружения"
    echo "3. Настройте автоматические проверки"
    echo "4. Мониторьте CVE базы данных"
    echo "5. Используйте зависимости с фиксированными версиями"
    echo ""
    
    # Возвращаем код выхода
    if [[ $total_vulns -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --verbose, -v  Подробный вывод"
        echo "  --json         Вывод в формате JSON"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Стандартная проверка"
        echo "  $0 --verbose          # Подробный вывод"
        echo "  $0 --json > report.json # JSON отчет"
        exit 0
        ;;
    --verbose|-v)
        set -x
        ;;
esac

# Запуск основной функции
main "$@" 