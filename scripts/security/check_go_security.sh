#!/bin/bash

# CloudBridge Relay - Go Security Analysis Script
# Проверка безопасности Go кода с помощью gosec

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

# Проверка наличия gosec
check_gosec() {
    if ! command -v gosec &> /dev/null; then
        print_error "gosec не установлен. Устанавливаем..."
        sudo snap install gosec
    fi
    print_success "gosec установлен и готов к использованию"
}

# Проверка наличия Go модулей
check_go_modules() {
    if [[ -f "go.mod" ]]; then
        print_status "Найден go.mod файл"
        return 0
    else
        print_warning "go.mod файл не найден в текущей директории"
        return 1
    fi
}

# Запуск gosec анализа
run_gosec_analysis() {
    local output_file=$1
    local format=${2:-json}
    local target=${3:-./...}
    
    print_status "Запускаем gosec анализ для: $target"
    
    # Создаем временный файл для вывода
    local temp_output=$(mktemp)
    
    # Запускаем gosec с различными опциями
    local gosec_cmd="gosec -fmt=$format -out=$temp_output $target"
    
    print_status "Выполняем команду: $gosec_cmd"
    
    # Запускаем gosec и сохраняем код выхода
    if eval "$gosec_cmd" 2>&1; then
        local exit_code=$?
        print_success "Анализ безопасности завершен (код выхода: $exit_code)"
        
        # Проверяем, есть ли проблемы
        if [[ -s "$temp_output" ]]; then
            print_warning "Найдены проблемы безопасности"
            
            # Сохраняем отчет
            if [[ -n "$output_file" ]]; then
                cp "$temp_output" "$output_file"
                print_status "Отчет сохранен в: $output_file"
            fi
            
            # Показываем краткую сводку
            echo "=== СВОДКА ПРОБЛЕМ БЕЗОПАСНОСТИ ==="
            if [[ "$format" == "json" ]]; then
                jq -r '.Issues[] | "\(.severity) - \(.rule_id): \(.details)"' "$temp_output" 2>/dev/null || echo "Не удалось проанализировать JSON"
            else
                cat "$temp_output"
            fi
            
            # Возвращаем количество проблем
            if [[ "$format" == "json" ]]; then
                jq '.Issues | length' "$temp_output" 2>/dev/null || echo "0"
            else
                # Для текстового формата подсчитываем строки с проблемами
                grep -c "G[0-9]*" "$temp_output" 2>/dev/null || echo "0"
            fi
        else
            print_success "Проблем безопасности не найдено!"
            echo "0"
        fi
    else
        print_error "Ошибка при выполнении gosec анализа"
        cat "$temp_output"
        echo "0"
    fi
    
    # Удаляем временный файл
    rm -f "$temp_output"
}

# Анализ конкретных директорий
analyze_directories() {
    local reports_dir=$1
    
    # Список директорий для анализа
    local directories=(
        "cmd"
        "internal"
        "pkg"
    )
    
    local total_issues=0
    
    for dir in "${directories[@]}"; do
        if [[ -d "$dir" ]]; then
            print_status "Анализируем директорию: $dir"
            local output_file="$reports_dir/gosec_${dir}_report.json"
            local issues=$(run_gosec_analysis "$output_file" "json" "$dir/...")
            total_issues=$((total_issues + issues))
        else
            print_warning "Директория не найдена: $dir"
        fi
    done
    
    echo $total_issues
}

# Основная функция проверки
main() {
    echo "=========================================="
    echo "CloudBridge Relay - Go Security Analysis"
    echo "Анализ безопасности Go кода"
    echo "=========================================="
    echo ""
    
    # Проверяем gosec
    check_gosec
    
    # Проверяем Go модули
    if ! check_go_modules; then
        print_error "Не найден go.mod файл. Убедитесь, что вы находитесь в корне Go проекта."
        exit 1
    fi
    
    # Создаем директорию для отчетов
    local reports_dir="security_reports"
    mkdir -p "$reports_dir"
    
    # Полный анализ проекта
    print_status "Выполняем полный анализ проекта..."
    local main_output="$reports_dir/gosec_full_report.json"
    local main_issues=$(run_gosec_analysis "$main_output" "json")
    
    # Анализ по директориям
    print_status "Выполняем анализ по директориям..."
    local dir_issues=$(analyze_directories "$reports_dir")
    
    # Генерируем HTML отчет
    print_status "Генерируем HTML отчет..."
    local html_output="$reports_dir/gosec_report.html"
    if gosec -fmt html -out "$html_output" ./... 2>/dev/null; then
        print_success "HTML отчет сохранен в: $html_output"
    fi
    
    # Генерируем текстовый отчет
    print_status "Генерируем текстовый отчет..."
    local text_output="$reports_dir/gosec_report.txt"
    if gosec -fmt text -out "$text_output" ./... 2>/dev/null; then
        print_success "Текстовый отчет сохранен в: $text_output"
    fi
    
    echo ""
    echo "=========================================="
    echo "ИТОГОВАЯ СВОДКА"
    echo "=========================================="
    
    # Подсчитываем общее количество проблем
    local total_issues=$((main_issues + dir_issues))
    
    if [[ $total_issues -gt 0 ]]; then
        print_warning "Общее количество проблем безопасности: $total_issues"
        print_status "Детальные отчеты сохранены в директории: $reports_dir"
        
        # Показываем краткую сводку из текстового отчета
        echo ""
        echo "=== КРАТКАЯ СВОДКА ПРОБЛЕМ ==="
        if [[ -f "$text_output" ]]; then
            grep -E "G[0-9]+.*\(CWE-[0-9]+\)" "$text_output" | head -10
        fi
    else
        print_success "Проблем безопасности не найдено!"
    fi
    
    echo ""
    echo "=========================================="
    echo "РЕКОМЕНДАЦИИ"
    echo "=========================================="
    echo "1. Регулярно запускайте gosec анализ"
    echo "2. Исправляйте найденные проблемы безопасности"
    echo "3. Настройте gosec в CI/CD пайплайне"
    echo "4. Используйте статический анализ кода"
    echo "5. Следите за обновлениями gosec"
    echo ""
    
    # Показываем статистику по типам проблем
    if [[ -f "$text_output" ]]; then
        echo "=== СТАТИСТИКА ПО ТИПАМ ПРОБЛЕМ ==="
        grep -o "G[0-9]*" "$text_output" | sort | uniq -c | sort -nr || echo "Не удалось получить статистику"
    fi
    
    # Возвращаем код выхода
    if [[ $total_issues -gt 0 ]]; then
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
        echo "  --html         Генерировать только HTML отчет"
        echo "  --text         Генерировать только текстовый отчет"
        echo "  --json         Генерировать только JSON отчет"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Полный анализ"
        echo "  $0 --verbose          # Подробный вывод"
        echo "  $0 --html             # Только HTML отчет"
        exit 0
        ;;
    --verbose|-v)
        set -x
        ;;
    --html)
        # Только HTML отчет
        reports_dir="security_reports"
        mkdir -p "$reports_dir"
        html_output="$reports_dir/gosec_report.html"
        gosec -fmt html -out "$html_output" ./...
        print_success "HTML отчет сохранен в: $html_output"
        exit 0
        ;;
    --text)
        # Только текстовый отчет
        reports_dir="security_reports"
        mkdir -p "$reports_dir"
        text_output="$reports_dir/gosec_report.txt"
        gosec -fmt text -out "$text_output" ./...
        print_success "Текстовый отчет сохранен в: $text_output"
        exit 0
        ;;
    --json)
        # Только JSON отчет
        reports_dir="security_reports"
        mkdir -p "$reports_dir"
        json_output="$reports_dir/gosec_report.json"
        gosec -fmt json -out "$json_output" ./...
        print_success "JSON отчет сохранен в: $json_output"
        exit 0
        ;;
esac

# Запуск основной функции
main "$@" 