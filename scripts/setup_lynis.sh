#!/bin/bash

# CloudBridge Relay - Lynis Security Audit Setup Script
# Установка и настройка Lynis для автоматического аудита безопасности

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $1${NC}"
}

# Проверка root прав
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error "Этот скрипт должен быть запущен с правами root (sudo)"
        exit 1
    fi
}

# Определение дистрибутива
detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si)
        VER=$(lsb_release -sr)
    elif [[ -f /etc/lsb-release ]]; then
        . /etc/lsb-release
        OS=$DISTRIB_ID
        VER=$DISTRIB_RELEASE
    elif [[ -f /etc/debian_version ]]; then
        OS=Debian
        VER=$(cat /etc/debian_version)
    elif [[ -f /etc/SuSe-release ]]; then
        OS=SuSE
    elif [[ -f /etc/redhat-release ]]; then
        OS=RedHat
    else
        OS=$(uname -s)
        VER=$(uname -r)
    fi
    
    log "Обнаружена ОС: $OS $VER"
}

# Установка Lynis
install_lynis() {
    log "Установка Lynis..."
    # Всегда устанавливаем из исходников для Ubuntu/Debian
    log "Установка Lynis из исходников..."
    cd /tmp
    rm -rf lynis
    git clone https://github.com/CISOfy/lynis.git
    cd lynis
    mkdir -p /usr/local/lynis
    cp -r * /usr/local/lynis/
    ln -sf /usr/local/lynis/lynis /usr/local/bin/lynis
    log "Lynis установлен успешно из исходников"
}

# Создание конфигурации Lynis
setup_lynis_config() {
    log "Настройка конфигурации Lynis..."
    
    # Создание директории конфигурации
    mkdir -p /etc/lynis
    
    # Основная конфигурация
    cat > /etc/lynis/lynis.conf << 'EOF'
# CloudBridge Relay - Lynis Configuration
# Конфигурация для аудита безопасности

# Основные настройки
PROFILE=relay-server
SKIP_TESTS=test_ssl_certificate_expiry
SKIP_TESTS=test_ssl_certificate_validity

# Настройки отчета
REPORT_FILE=/var/log/lynis/lynis-report.dat
LOG_FILE=/var/log/lynis/lynis.log
REPORT_FORMAT=text

# Настройки безопасности
QUIET=0
VERBOSE=1
WARNINGS=1

# Настройки для CloudBridge Relay
SKIP_TESTS=test_ssl_certificate_expiry
SKIP_TESTS=test_ssl_certificate_validity
SKIP_TESTS=test_ssl_certificate_trust

# Исключения для Docker контейнеров
SKIP_TESTS=test_ssl_certificate_expiry
SKIP_TESTS=test_ssl_certificate_validity
SKIP_TESTS=test_ssl_certificate_trust

# Настройки для российских стандартов
COMPLIANCE_152_FZ=true
COMPLIANCE_187_FZ=true
USE_RUSSIAN_CRYPTO=true
EOF
    
    # Создание профиля для CloudBridge Relay
    cat > /etc/lynis/custom.prf << 'EOF'
# CloudBridge Relay Security Profile
# Профиль безопасности для CloudBridge Relay

# Настройки для relay сервера
RELAY_SERVER=true
DOCKER_ENVIRONMENT=true
CONTAINERIZED_APPLICATION=true

# Настройки для multi-tenancy
MULTI_TENANT=true
TENANT_ISOLATION=true

# Настройки для TCP proxy
TCP_PROXY=true
HIGH_PERFORMANCE=true

# Настройки для мониторинга
MONITORING_ENABLED=true
METRICS_COLLECTION=true
PROMETHEUS_INTEGRATION=true

# Настройки для аутентификации
KEYCLOAK_INTEGRATION=true
JWT_AUTHENTICATION=true
MTLS_ENABLED=true

# Настройки для соответствия российским стандартам
RUSSIAN_COMPLIANCE=true
GOST_CRYPTO=true
FZ_152_COMPLIANCE=true
FZ_187_COMPLIANCE=true
EOF
    
    # Создание директорий для логов
    mkdir -p /var/log/lynis
    chmod 750 /var/log/lynis
    
    log "Конфигурация Lynis создана"
}

# Создание скрипта для автоматического аудита
create_audit_script() {
    log "Создание скрипта автоматического аудита..."
    
    cat > /usr/local/bin/relay-security-audit << 'EOF'
#!/bin/bash

# CloudBridge Relay - Security Audit Script
# Автоматический аудит безопасности с Lynis

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
AUDIT_TYPE=${1:-full}
OUTPUT_DIR=${2:-/var/log/lynis}
REPORT_NAME="relay-security-audit-$(date +%Y%m%d-%H%M%S)"

# Создание директории для отчетов
mkdir -p "$OUTPUT_DIR"

log "Запуск аудита безопасности CloudBridge Relay..."
log "Тип аудита: $AUDIT_TYPE"
log "Директория отчетов: $OUTPUT_DIR"

# Запуск Lynis
case $AUDIT_TYPE in
    "full")
        log "Выполнение полного аудита..."
        lynis audit system --profile /etc/lynis/custom.prf --quick
        ;;
    "quick")
        log "Выполнение быстрого аудита..."
        lynis audit system --profile /etc/lynis/custom.prf --quick
        ;;
    "docker")
        log "Аудит Docker окружения..."
        lynis audit system --profile /etc/lynis/custom.prf --tests-from-group docker
        ;;
    "network")
        log "Аудит сетевой безопасности..."
        lynis audit system --profile /etc/lynis/custom.prf --tests-from-group network
        ;;
    "files")
        log "Аудит файловой системы..."
        lynis audit system --profile /etc/lynis/custom.prf --tests-from-group files
        ;;
    *)
        error "Неизвестный тип аудита: $AUDIT_TYPE"
        echo "Доступные типы: full, quick, docker, network, files"
        exit 1
        ;;
esac

# Копирование отчета
if [[ -f /var/log/lynis/lynis-report.dat ]]; then
    cp /var/log/lynis/lynis-report.dat "$OUTPUT_DIR/$REPORT_NAME.dat"
    log "Отчет сохранен: $OUTPUT_DIR/$REPORT_NAME.dat"
fi

if [[ -f /var/log/lynis/lynis.log ]]; then
    cp /var/log/lynis/lynis.log "$OUTPUT_DIR/$REPORT_NAME.log"
    log "Лог сохранен: $OUTPUT_DIR/$REPORT_NAME.log"
fi

# Генерация HTML отчета
if command -v lynis-report >/dev/null 2>&1; then
    lynis-report --output html --file "$OUTPUT_DIR/$REPORT_NAME.html" /var/log/lynis/lynis-report.dat
    log "HTML отчет создан: $OUTPUT_DIR/$REPORT_NAME.html"
fi

# Анализ результатов
log "Анализ результатов аудита..."

# Извлечение оценок
SCORE=$(grep "Hardening index" /var/log/lynis/lynis-report.dat | cut -d'=' -f2 | tr -d ' ')
WARNINGS=$(grep "warnings" /var/log/lynis/lynis-report.dat | cut -d'=' -f2 | tr -d ' ')
SUGGESTIONS=$(grep "suggestions" /var/log/lynis/lynis-report.dat | cut -d'=' -f2 | tr -d ' ')

log "Результаты аудита:"
log "  Оценка безопасности: $SCORE/100"
log "  Предупреждения: $WARNINGS"
log "  Рекомендации: $SUGGESTIONS"

# Проверка критических проблем
CRITICAL_ISSUES=$(grep -c "Suggestion.*CRITICAL" /var/log/lynis/lynis-report.dat || echo "0")
if [[ $CRITICAL_ISSUES -gt 0 ]]; then
    warn "Обнаружено $CRITICAL_ISSUES критических проблем!"
    echo "Критические проблемы:"
    grep "Suggestion.*CRITICAL" /var/log/lynis/lynis-report.dat
fi

# Отправка уведомления в Prometheus (если доступен)
if command -v curl >/dev/null 2>&1; then
    curl -X POST -H "Content-Type: application/json" \
         -d "{\"score\":$SCORE,\"warnings\":$WARNINGS,\"suggestions\":$SUGGESTIONS,\"critical_issues\":$CRITICAL_ISSUES}" \
         http://localhost:8081/metrics/security-audit 2>/dev/null || true
fi

log "Аудит безопасности завершен"
EOF
    
    chmod +x /usr/local/bin/relay-security-audit
    
    log "Скрипт автоматического аудита создан"
}

# Создание systemd сервиса для регулярного аудита
create_audit_service() {
    log "Создание systemd сервиса для регулярного аудита..."
    
    cat > /etc/systemd/system/relay-security-audit.service << 'EOF'
[Unit]
Description=CloudBridge Relay Security Audit Service
Documentation=https://github.com/CISOfy/lynis
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/relay-security-audit full
User=root
Group=root
StandardOutput=journal
StandardError=journal

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/lynis

[Install]
WantedBy=multi-user.target
EOF

    cat > /etc/systemd/system/relay-security-audit.timer << 'EOF'
[Unit]
Description=Run CloudBridge Relay Security Audit daily
Documentation=https://github.com/CISOfy/lynis

[Timer]
OnCalendar=daily
Persistent=true
RandomizedDelaySec=3600

[Install]
WantedBy=timers.target
EOF
    
    # Включение и запуск сервиса
    systemctl daemon-reload
    systemctl enable relay-security-audit.timer
    systemctl start relay-security-audit.timer
    
    log "Systemd сервис для регулярного аудита создан"
}

# Создание скрипта для интеграции с Prometheus
create_prometheus_integration() {
    log "Создание интеграции с Prometheus..."
    
    cat > /usr/local/bin/lynis-metrics-exporter << 'EOF'
#!/bin/bash

# Lynis Metrics Exporter для Prometheus
# Экспорт метрик аудита безопасности

set -e

METRICS_FILE="/var/log/lynis/lynis-metrics.prom"
REPORT_FILE="/var/log/lynis/lynis-report.dat"

# Функция для извлечения метрик
extract_metrics() {
    if [[ ! -f "$REPORT_FILE" ]]; then
        return 1
    fi
    
    # Извлечение оценок
    SCORE=$(grep "Hardening index" "$REPORT_FILE" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    WARNINGS=$(grep "warnings" "$REPORT_FILE" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    SUGGESTIONS=$(grep "suggestions" "$REPORT_FILE" | cut -d'=' -f2 | tr -d ' ' || echo "0")
    CRITICAL_ISSUES=$(grep -c "Suggestion.*CRITICAL" "$REPORT_FILE" || echo "0")
    
    # Создание метрик в формате Prometheus
    cat > "$METRICS_FILE" << METRICS_EOF
# HELP lynis_security_score Security hardening score (0-100)
# TYPE lynis_security_score gauge
lynis_security_score $SCORE

# HELP lynis_warnings_total Total number of warnings
# TYPE lynis_warnings_total gauge
lynis_warnings_total $WARNINGS

# HELP lynis_suggestions_total Total number of suggestions
# TYPE lynis_suggestions_total gauge
lynis_suggestions_total $SUGGESTIONS

# HELP lynis_critical_issues_total Total number of critical issues
# TYPE lynis_critical_issues_total gauge
lynis_critical_issues_total $CRITICAL_ISSUES

# HELP lynis_audit_timestamp Last audit timestamp
# TYPE lynis_audit_timestamp gauge
lynis_audit_timestamp $(date +%s)
METRICS_EOF
}

# Основной цикл
while true; do
    extract_metrics
    sleep 300  # Обновление каждые 5 минут
done
EOF
    
    chmod +x /usr/local/bin/lynis-metrics-exporter
    
    # Создание systemd сервиса для экспорта метрик
    cat > /etc/systemd/system/lynis-metrics-exporter.service << 'EOF'
[Unit]
Description=Lynis Metrics Exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/lynis-metrics-exporter
User=root
Group=root
Restart=always
RestartSec=10

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/log/lynis

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable lynis-metrics-exporter
    systemctl start lynis-metrics-exporter
    
    log "Интеграция с Prometheus создана"
}

# Создание Grafana дашборда
create_grafana_dashboard() {
    log "Создание Grafana дашборда для аудита безопасности..."
    
    mkdir -p /etc/grafana/provisioning/dashboards
    
    cat > /etc/grafana/provisioning/dashboards/lynis-security-audit.yaml << 'EOF'
apiVersion: 1

providers:
  - name: 'Lynis Security Audit'
    orgId: 1
    folder: 'Security'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards/lynis
EOF
    
    mkdir -p /var/lib/grafana/dashboards/lynis
    
    cat > /var/lib/grafana/dashboards/lynis/lynis-security-audit.json << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "Lynis Security Audit",
    "tags": ["security", "lynis", "audit"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Security Score",
        "type": "stat",
        "targets": [
          {
            "expr": "lynis_security_score",
            "legendFormat": "Security Score"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 50},
                {"color": "green", "value": 80}
              ]
            },
            "unit": "percent"
          }
        }
      },
      {
        "id": 2,
        "title": "Security Issues",
        "type": "stat",
        "targets": [
          {
            "expr": "lynis_warnings_total",
            "legendFormat": "Warnings"
          },
          {
            "expr": "lynis_suggestions_total",
            "legendFormat": "Suggestions"
          },
          {
            "expr": "lynis_critical_issues_total",
            "legendFormat": "Critical Issues"
          }
        ]
      }
    ]
  }
}
EOF
    
    log "Grafana дашборд создан"
}

# Создание скрипта для управления аудитом
create_management_script() {
    log "Создание скрипта управления аудитом..."
    
    cat > /usr/local/bin/relay-security-manager << 'EOF'
#!/bin/bash

# CloudBridge Relay - Security Audit Manager
# Управление аудитом безопасности

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

# Функция показа справки
show_help() {
    echo "CloudBridge Relay - Security Audit Manager"
    echo
    echo "Использование: $0 [команда] [опции]"
    echo
    echo "Команды:"
    echo "  audit [тип]     - Запуск аудита (full|quick|docker|network|files)"
    echo "  status          - Показать статус сервисов аудита"
    echo "  report [файл]   - Показать последний отчет"
    echo "  fix [проблема]  - Автоматическое исправление проблем"
    echo "  schedule        - Управление расписанием аудита"
    echo "  install         - Установка Lynis"
    echo "  uninstall       - Удаление Lynis"
    echo
    echo "Примеры:"
    echo "  $0 audit full"
    echo "  $0 status"
    echo "  $0 report"
    echo "  $0 fix all"
}

# Функция показа статуса
show_status() {
    log "Статус сервисов аудита безопасности:"
    echo
    
    echo "=== Lynis ==="
    if command -v lynis >/dev/null 2>&1; then
        lynis --version
    else
        error "Lynis не установлен"
    fi
    
    echo
    echo "=== Systemd сервисы ==="
    systemctl status relay-security-audit.timer --no-pager || true
    systemctl status lynis-metrics-exporter --no-pager || true
    
    echo
    echo "=== Последние отчеты ==="
    ls -la /var/log/lynis/ | head -10
}

# Функция показа отчета
show_report() {
    local report_file=${1:-/var/log/lynis/lynis-report.dat}
    
    if [[ ! -f "$report_file" ]]; then
        error "Отчет не найден: $report_file"
        return 1
    fi
    
    log "Отчет аудита безопасности: $report_file"
    echo
    
    # Показать основные метрики
    echo "=== Основные метрики ==="
    grep -E "(Hardening index|warnings|suggestions)" "$report_file" || true
    
    echo
    echo "=== Критические проблемы ==="
    grep "Suggestion.*CRITICAL" "$report_file" || echo "Критических проблем не найдено"
    
    echo
    echo "=== Рекомендации ==="
    grep "Suggestion" "$report_file" | head -10 || true
}

# Функция автоматического исправления
fix_issues() {
    local issue_type=${1:-all}
    
    log "Автоматическое исправление проблем: $issue_type"
    
    case $issue_type in
        "all")
            # Обновление системы
            log "Обновление системы..."
            apt-get update && apt-get upgrade -y || yum update -y || true
            
            # Настройка firewall
            log "Настройка firewall..."
            /home/ubuntu/cloudbridge-relay-installer/scripts/setup_firewall.sh || true
            
            # Настройка fail2ban
            log "Настройка fail2ban..."
            systemctl enable fail2ban || true
            systemctl start fail2ban || true
            ;;
        "firewall")
            /home/ubuntu/cloudbridge-relay-installer/scripts/setup_firewall.sh
            ;;
        "updates")
            apt-get update && apt-get upgrade -y || yum update -y
            ;;
        *)
            error "Неизвестный тип проблемы: $issue_type"
            echo "Доступные типы: all, firewall, updates"
            exit 1
            ;;
    esac
    
    log "Исправление завершено"
}

# Основная логика
case ${1:-help} in
    "audit")
        /usr/local/bin/relay-security-audit "${2:-full}"
        ;;
    "status")
        show_status
        ;;
    "report")
        show_report "$2"
        ;;
    "fix")
        fix_issues "$2"
        ;;
    "schedule")
        case ${2:-status} in
            "status")
                systemctl status relay-security-audit.timer
                ;;
            "enable")
                systemctl enable relay-security-audit.timer
                systemctl start relay-security-audit.timer
                ;;
            "disable")
                systemctl stop relay-security-audit.timer
                systemctl disable relay-security-audit.timer
                ;;
            *)
                echo "Доступные команды: status, enable, disable"
                ;;
        esac
        ;;
    "install")
        /home/ubuntu/cloudbridge-relay-installer/scripts/setup_lynis.sh
        ;;
    "uninstall")
        apt-get remove -y lynis || yum remove -y lynis || true
        rm -rf /etc/lynis /var/log/lynis
        systemctl disable relay-security-audit.timer || true
        systemctl disable lynis-metrics-exporter || true
        rm -f /usr/local/bin/relay-security-audit
        rm -f /usr/local/bin/relay-security-manager
        rm -f /usr/local/bin/lynis-metrics-exporter
        ;;
    "help"|*)
        show_help
        ;;
esac
EOF
    
    chmod +x /usr/local/bin/relay-security-manager
    
    log "Скрипт управления аудитом создан"
}

# Основная функция
main() {
    log "Настройка Lynis для CloudBridge Relay v2.0"
    
    check_root
    detect_distro
    install_lynis
    setup_lynis_config
    create_audit_script
    create_audit_service
    create_prometheus_integration
    create_grafana_dashboard
    create_management_script
    
    log "Настройка Lynis завершена успешно!"
    log ""
    log "Доступные команды:"
    log "  relay-security-audit [тип]  - Запуск аудита"
    log "  relay-security-manager      - Управление аудитом"
    log "  lynis audit system          - Прямой запуск Lynis"
    log ""
    log "Примеры использования:"
    log "  relay-security-audit full   - Полный аудит"
    log "  relay-security-audit quick  - Быстрый аудит"
    log "  relay-security-manager status - Статус сервисов"
    log "  relay-security-manager report - Последний отчет"
}

# Запуск основной функции
main "$@" 