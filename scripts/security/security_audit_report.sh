#!/bin/bash

# CloudBridge Relay - Comprehensive Security Audit Report
# Сводный отчет по безопасности проекта

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
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

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

print_section() {
    echo -e "${CYAN}$1${NC}"
}

# Создание отчета
generate_report() {
    local report_file="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    cat > "$report_file" << EOF
# CloudBridge Relay - Security Audit Report
Generated: $timestamp

## Executive Summary

This report contains the results of a comprehensive security audit of the CloudBridge Relay project.

## Table of Contents

1. [Python Dependencies Security](#python-dependencies-security)
2. [Go Code Security Analysis](#go-code-security-analysis)
3. [System Security](#system-security)
4. [Recommendations](#recommendations)
5. [Risk Assessment](#risk-assessment)

---

## Python Dependencies Security

### Safety Analysis Results

EOF

    # Добавляем результаты safety
    if [[ -f "security_reports/ml_requirements_security_report.json" ]]; then
        echo "### ML Requirements Analysis" >> "$report_file"
        local ml_vulns=$(jq '.vulnerabilities | length' "security_reports/ml_requirements_security_report.json" 2>/dev/null || echo "0")
        echo "- **Vulnerabilities found:** $ml_vulns" >> "$report_file"
        
        if [[ $ml_vulns -gt 0 ]]; then
            echo "- **Critical issues:** $(jq '.vulnerabilities[] | select(.severity == "HIGH") | .package_name' "security_reports/ml_requirements_security_report.json" 2>/dev/null | wc -l)" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi
    
    if [[ -f "security_reports/siem_requirements_security_report.json" ]]; then
        echo "### SIEM Requirements Analysis" >> "$report_file"
        local siem_vulns=$(jq '.vulnerabilities | length' "security_reports/siem_requirements_security_report.json" 2>/dev/null || echo "0")
        echo "- **Vulnerabilities found:** $siem_vulns" >> "$report_file"
        
        if [[ $siem_vulns -gt 0 ]]; then
            echo "- **Critical issues:** $(jq '.vulnerabilities[] | select(.severity == "HIGH") | .package_name' "security_reports/siem_requirements_security_report.json" 2>/dev/null | wc -l)" >> "$report_file"
        fi
        echo "" >> "$report_file"
    fi

    cat >> "$report_file" << EOF

---

## Go Code Security Analysis

### Gosec Analysis Results

EOF

    # Добавляем результаты gosec
    if [[ -f "security_reports/gosec_report.txt" ]]; then
        echo "### Security Issues Found" >> "$report_file"
        local gosec_issues=$(grep -c "G[0-9]*" "security_reports/gosec_report.txt" 2>/dev/null || echo "0")
        echo "- **Total issues:** $gosec_issues" >> "$report_file"
        
        if [[ $gosec_issues -gt 0 ]]; then
            echo "- **High severity:** $(grep -c "Severity: HIGH" "security_reports/gosec_report.txt" 2>/dev/null || echo "0")" >> "$report_file"
            echo "- **Medium severity:** $(grep -c "Severity: MEDIUM" "security_reports/gosec_report.txt" 2>/dev/null || echo "0")" >> "$report_file"
            echo "- **Low severity:** $(grep -c "Severity: LOW" "security_reports/gosec_report.txt" 2>/dev/null || echo "0")" >> "$report_file"
            
            echo "" >> "$report_file"
            echo "### Top Issues:" >> "$report_file"
            grep -E "G[0-9]+.*\(CWE-[0-9]+\)" "security_reports/gosec_report.txt" | head -5 | while read -r line; do
                echo "- $line" >> "$report_file"
            done
        fi
        echo "" >> "$report_file"
    fi

    cat >> "$report_file" << EOF

---

## System Security

### Environment Information

- **OS:** $(uname -s -r)
- **Architecture:** $(uname -m)
- **Go Version:** $(go version 2>/dev/null || echo "Not installed")
- **Python Version:** $(python3 --version 2>/dev/null || echo "Not installed")

### Security Tools Status

- **Safety:** $(safety --version 2>/dev/null | head -1 || echo "Not installed")
- **Gosec:** $(gosec --version 2>/dev/null | head -1 || echo "Not installed")
- **Lynis:** $(lynis --version 2>/dev/null | head -1 || echo "Not installed")

---

## Recommendations

### Immediate Actions Required

EOF

    # Генерируем рекомендации на основе найденных проблем
    local total_issues=0
    
    # Подсчитываем проблемы Python
    for report in security_reports/*_requirements_security_report.json; do
        if [[ -f "$report" ]]; then
            local vulns=$(jq '.vulnerabilities | length' "$report" 2>/dev/null || echo "0")
            total_issues=$((total_issues + vulns))
        fi
    done
    
    # Подсчитываем проблемы Go
    if [[ -f "security_reports/gosec_report.txt" ]]; then
        local go_issues=$(grep -c "G[0-9]*" "security_reports/gosec_report.txt" 2>/dev/null || echo "0")
        total_issues=$((total_issues + go_issues))
    fi
    
    if [[ $total_issues -gt 0 ]]; then
        cat >> "$report_file" << EOF
1. **Update vulnerable dependencies** - $total_issues security issues found
2. **Review and fix Go security issues** - Address code-level vulnerabilities
3. **Implement security scanning in CI/CD** - Automate security checks
4. **Regular security audits** - Schedule monthly security reviews

### Medium-term Actions

1. **Implement security monitoring** - Set up alerts for new vulnerabilities
2. **Security training** - Educate team on secure coding practices
3. **Dependency management** - Implement automated dependency updates
4. **Code review process** - Mandate security-focused code reviews

### Long-term Actions

1. **Security framework adoption** - Implement comprehensive security framework
2. **Threat modeling** - Regular threat modeling exercises
3. **Penetration testing** - Annual penetration testing
4. **Security metrics** - Track and improve security metrics
EOF
    else
        cat >> "$report_file" << EOF
1. **Maintain current security posture** - No immediate issues found
2. **Continue regular security monitoring** - Keep up with best practices
3. **Stay updated with dependencies** - Monitor for new vulnerabilities
4. **Document security practices** - Share knowledge with team
EOF
    fi

    cat >> "$report_file" << EOF

---

## Risk Assessment

### Risk Levels

- **Critical (Red):** Issues that could lead to immediate compromise
- **High (Orange):** Issues that could lead to significant security impact
- **Medium (Yellow):** Issues that could lead to moderate security impact
- **Low (Green):** Issues with minimal security impact

### Current Risk Status

EOF

    # Оцениваем общий риск
    if [[ $total_issues -gt 20 ]]; then
        echo "- **Overall Risk Level:** **CRITICAL**" >> "$report_file"
        echo "- **Action Required:** Immediate attention required" >> "$report_file"
    elif [[ $total_issues -gt 10 ]]; then
        echo "- **Overall Risk Level:** 🟠 **HIGH**" >> "$report_file"
        echo "- **Action Required:** Prompt attention required" >> "$report_file"
    elif [[ $total_issues -gt 5 ]]; then
        echo "- **Overall Risk Level:** **MEDIUM**" >> "$report_file"
        echo "- **Action Required:** Attention required" >> "$report_file"
    else
        echo "- **Overall Risk Level:** **LOW**" >> "$report_file"
        echo "- **Action Required:** Monitor and maintain" >> "$report_file"
    fi

    cat >> "$report_file" << EOF

---

## Appendices

### A. Detailed Python Vulnerability Report

See individual JSON reports in security_reports/ directory.

### B. Detailed Go Security Report

See gosec_report.txt and gosec_report.html for detailed analysis.

### C. Tools Used

- **Safety v3.5.2** - Python dependency vulnerability scanner
- **Gosec v2.22.5** - Go security linter
- **Custom scripts** - Automated security analysis

### D. References

- [Safety Documentation](https://pyup.io/safety/)
- [Gosec Documentation](https://github.com/securecodewarrior/gosec)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security)

---

*Report generated by CloudBridge Relay Security Audit System*
EOF

    print_success "Security audit report generated: $report_file"
}

# Основная функция
main() {
    echo "=========================================="
    echo "CloudBridge Relay - Security Audit Report"
    echo "Генерация сводного отчета по безопасности"
    echo "=========================================="
    echo ""
    
    # Создаем директорию для отчетов
    local reports_dir="security_reports"
    mkdir -p "$reports_dir"
    
    # Запускаем все проверки безопасности
    print_header "1. Проверка Python зависимостей..."
    if [[ -f "scripts/security/check_dependencies_security.sh" ]]; then
        ./scripts/security/check_dependencies_security.sh
    else
        print_warning "Скрипт проверки Python зависимостей не найден"
    fi
    
    print_header "2. Проверка Go кода..."
    if [[ -f "scripts/security/check_go_security.sh" ]]; then
        ./scripts/security/check_go_security.sh
    else
        print_warning "Скрипт проверки Go кода не найден"
    fi
    
    print_header "3. Генерация сводного отчета..."
    local report_file="$reports_dir/security_audit_report.md"
    generate_report "$report_file"
    
    echo ""
    echo "=========================================="
    echo "ИТОГОВАЯ СВОДКА"
    echo "=========================================="
    print_success "Сводный отчет по безопасности создан: $report_file"
    
    # Показываем краткую статистику
    echo ""
    print_section "Краткая статистика:"
    
    # Python уязвимости
    local python_vulns=0
    for report in "$reports_dir"/*_requirements_security_report.json; do
        if [[ -f "$report" ]]; then
            local vulns=$(jq '.vulnerabilities | length' "$report" 2>/dev/null || echo "0")
            python_vulns=$((python_vulns + vulns))
        fi
    done
    echo "- Python уязвимости: $python_vulns"
    
    # Go проблемы
    local go_issues=0
    if [[ -f "$reports_dir/gosec_report.txt" ]]; then
        go_issues=$(grep -c "G[0-9]*" "$reports_dir/gosec_report.txt" 2>/dev/null || echo "0")
    fi
    echo "- Go проблемы безопасности: $go_issues"
    
    local total_issues=$((python_vulns + go_issues))
    echo "- Общее количество проблем: $total_issues"
    
    echo ""
    print_section "Доступные отчеты:"
    echo "- Сводный отчет: $report_file"
    echo "- HTML отчет Go: $reports_dir/gosec_report.html"
    echo "- Текстовый отчет Go: $reports_dir/gosec_report.txt"
    echo "- JSON отчеты Python: $reports_dir/*_requirements_security_report.json"
    
    echo ""
    print_section "Следующие шаги:"
    if [[ $total_issues -gt 0 ]]; then
        echo "1. Изучите детальные отчеты"
        echo "2. Исправьте найденные проблемы"
        echo "3. Настройте автоматические проверки"
        echo "4. Запланируйте регулярные аудиты"
    else
        echo "1. Отличная работа! Проблем безопасности не найдено"
        echo "2. Продолжайте регулярные проверки"
        echo "3. Следите за обновлениями зависимостей"
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
        echo "  --quick        Быстрая проверка (без детальных отчетов)"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Полный аудит безопасности"
        echo "  $0 --verbose          # Подробный вывод"
        echo "  $0 --quick            # Быстрая проверка"
        exit 0
        ;;
    --verbose|-v)
        set -x
        ;;
    --quick)
        # Быстрая проверка
        print_header "Быстрая проверка безопасности..."
        ./scripts/security/check_dependencies_security.sh --json > /dev/null 2>&1 || true
        ./scripts/security/check_go_security.sh --text > /dev/null 2>&1 || true
        print_success "Быстрая проверка завершена"
        exit 0
        ;;
esac

# Запуск основной функции
main "$@" 