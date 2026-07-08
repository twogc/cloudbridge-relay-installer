#!/bin/bash

# Скрипт для проверки структуры документации
# Проверяет наличие всех необходимых файлов и папок

set -e

echo "Проверка структуры документации CloudBridge Relay"
echo "=================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для проверки файла
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}${NC} $1"
        return 0
    else
        echo -e "${RED}${NC} $1 - НЕ НАЙДЕН"
        return 1
    fi
}

# Функция для проверки папки
check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}📁${NC} $1"
        return 0
    else
        echo -e "${RED}${NC} $1 - НЕ НАЙДЕНА"
        return 1
    fi
}

# Функция для подсчета файлов
count_files() {
    local count=$(find "$1" -name "*.md" | wc -l)
    echo -e "${YELLOW}📊${NC} $1: $count файлов"
    return $count
}

errors=0

echo ""
echo "📁 Проверка основных папок:"
echo "---------------------------"

# Проверка основных папок
check_dir "docs" || ((errors++))
check_dir "docs/guides" || ((errors++))
check_dir "docs/deployment" || ((errors++))
check_dir "docs/security" || ((errors++))
check_dir "docs/integration" || ((errors++))
check_dir "docs/troubleshooting" || ((errors++))
check_dir "docs/reports" || ((errors++))
check_dir "docs/changelog" || ((errors++))

echo ""
echo "📖 Проверка основных файлов:"
echo "---------------------------"

# Проверка основных файлов
check_file "docs/README.md" || ((errors++))
check_file "DOCUMENTATION.md" || ((errors++))
check_file "docs/DOCUMENTATION_REORGANIZATION_REPORT.md" || ((errors++))

echo ""
echo "Проверка README файлов в папках:"
echo "----------------------------------"

# Проверка README файлов в папках
check_file "docs/guides/README.md" || ((errors++))
check_file "docs/deployment/README.md" || ((errors++))
check_file "docs/security/README.md" || ((errors++))
check_file "docs/integration/README.md" || ((errors++))
check_file "docs/troubleshooting/README.md" || ((errors++))
check_file "docs/reports/README.md" || ((errors++))
check_file "docs/changelog/README.md" || ((errors++))

echo ""
echo "Статистика файлов по категориям:"
echo "-----------------------------------"

# Подсчет файлов в каждой папке
count_files "docs/guides"
count_files "docs/deployment"
count_files "docs/security"
count_files "docs/integration"
count_files "docs/troubleshooting"
count_files "docs/reports"
count_files "docs/changelog"

echo ""
echo "Проверка перемещенных файлов:"
echo "-------------------------------"

# Проверка перемещенных файлов
check_file "docs/guides/VUE_THEME_USAGE.md" || ((errors++))
check_file "docs/changelog/CHANGELOG_START_PRODUCTION.md" || ((errors++))
check_file "docs/deployment/START_PRODUCTION_UPDATE.md" || ((errors++))
check_file "docs/integration/KEYCLOAK_THEME_ACTIVATION.md" || ((errors++))
check_file "docs/guides/DASHBOARD_SUMMARY.md" || ((errors++))
check_file "docs/security/gosec_checklist.md" || ((errors++))
check_file "docs/reports/gosec_report.html" || ((errors++))
check_file "docs/reports/gosec_report.txt" || ((errors++))
check_file "docs/integration/ARCHITECTURE.md" || ((errors++))
check_file "docs/deployment/QUICKSTART.md" || ((errors++))
check_file "docs/integration/siem_readme.md" || ((errors++))

echo ""
echo " Проверка ключевых файлов документации:"
echo "----------------------------------------"

# Проверка ключевых файлов
check_file "docs/guides/quick_start_telegram.md" || ((errors++))
check_file "docs/deployment/deployment_guide.md" || ((errors++))
check_file "docs/security/security.md" || ((errors++))
check_file "docs/integration/keycloak_integration.md" || ((errors++))
check_file "docs/troubleshooting/debugging.md" || ((errors++))

echo ""
echo " Общая статистика:"
echo "-------------------"

total_files=$(find docs -name "*.md" | wc -l)
echo -e "${YELLOW}📊${NC} Всего файлов документации: $total_files"

echo ""
echo "=================================================="

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}Все проверки пройдены успешно!${NC}"
    echo -e "${GREEN}📚 Документация полностью структурирована${NC}"
    exit 0
else
    echo -e "${RED}Найдено ошибок: $errors${NC}"
    echo -e "${YELLOW}  Проверьте отсутствующие файлы и папки${NC}"
    exit 1
fi 