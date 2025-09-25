#!/bin/bash

#  Скрипт интеграции SIEM с существующим мониторингом CloudBridge Relay
# Автор: CloudBridge Team
# Версия: 2.0.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка прав администратора
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "Этот скрипт должен быть запущен с правами администратора"
        exit 1
    fi
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("docker" "docker-compose" "jq" "curl")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Отсутствуют зависимости: ${missing_deps[*]}"
        log_info "Установите их и повторите попытку"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Получение параметров
parse_arguments() {
    SIEM_SERVER_IP=""
    SIEM_SERVER_PORT="8000"
    BACKUP_EXISTING=true
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --siem-ip)
                SIEM_SERVER_IP="$2"
                shift 2
                ;;
            --siem-port)
                SIEM_SERVER_PORT="$2"
                shift 2
                ;;
            --no-backup)
                BACKUP_EXISTING=false
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Неизвестный параметр: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    if [[ -z "$SIEM_SERVER_IP" ]]; then
        log_error "Необходимо указать IP адрес SIEM сервера (--siem-ip)"
        show_help
        exit 1
    fi
}

# Показать справку
show_help() {
    cat << EOF
 Скрипт интеграции SIEM с существующим мониторингом

Использование:
    sudo $0 --siem-ip <IP_ADDRESS> [опции]

Параметры:
    --siem-ip <IP>        IP адрес SIEM сервера (обязательно)
    --siem-port <PORT>    Порт SIEM сервера (по умолчанию: 8000)
    --no-backup          Не создавать резервные копии существующих конфигураций
    --help               Показать эту справку

Примеры:
    sudo $0 --siem-ip 192.168.1.100
    sudo $0 --siem-ip 10.0.0.50 --siem-port 8080
    sudo $0 --siem-ip 172.16.0.10 --no-backup

EOF
}

# Создание резервных копий
create_backups() {
    if [[ "$BACKUP_EXISTING" == "true" ]]; then
        log_info "Создание резервных копий существующих конфигураций..."
        
        local timestamp=$(date +%Y%m%d_%H%M%S)
        local backup_dir="backups/monitoring_integration_${timestamp}"
        
        mkdir -p "$backup_dir"
        
        # Резервная копия Prometheus
        if [[ -f "prometheus/prometheus.yml" ]]; then
            cp "prometheus/prometheus.yml" "$backup_dir/prometheus.yml.backup"
            log_success "Создана резервная копия prometheus.yml"
        fi
        
        # Резервная копия AlertManager
        if [[ -f "deploy/alertmanager/alertmanager.yml" ]]; then
            cp "deploy/alertmanager/alertmanager.yml" "$backup_dir/alertmanager.yml.backup"
            log_success "Создана резервная копия alertmanager.yml"
        fi
        
        log_success "Резервные копии сохранены в $backup_dir"
    fi
}

# Интеграция с Prometheus
integrate_prometheus() {
    log_info "Интеграция с Prometheus..."
    
    local prometheus_file="prometheus/prometheus.yml"
    
    # Проверка существования файла
    if [[ ! -f "$prometheus_file" ]]; then
        log_error "Файл $prometheus_file не найден"
        return 1
    fi
    
    # Добавление SIEM scrape jobs
    local siem_jobs=(
        "  - job_name: 'siem-backend'"
        "    static_configs:"
        "      - targets: ['${SIEM_SERVER_IP}:${SIEM_SERVER_PORT}']"
        "    metrics_path: '/metrics'"
        "    scrape_interval: 15s"
        "    scrape_timeout: 5s"
        ""
        "  - job_name: 'siem-processor'"
        "    static_configs:"
        "      - targets: ['${SIEM_SERVER_IP}:8001']"
        "    metrics_path: '/metrics'"
        "    scrape_interval: 15s"
        "    scrape_timeout: 5s"
        ""
        "  - job_name: 'siem-alerts'"
        "    static_configs:"
        "      - targets: ['${SIEM_SERVER_IP}:8002']"
        "    metrics_path: '/metrics'"
        "    scrape_interval: 15s"
        "    scrape_timeout: 5s"
    )
    
    # Добавление jobs в конец scrape_configs
    local temp_file=$(mktemp)
    local in_scrape_configs=false
    local added_jobs=false
    
    while IFS= read -r line; do
        echo "$line" >> "$temp_file"
        
        if [[ "$line" == "scrape_configs:" ]]; then
            in_scrape_configs=true
        elif [[ "$in_scrape_configs" == "true" && "$line" =~ ^[[:space:]]*-[[:space:]]*job_name: ]]; then
            if [[ "$added_jobs" == "false" ]]; then
                # Добавляем SIEM jobs перед первым существующим job
                for job in "${siem_jobs[@]}"; do
                    echo "$job" >> "$temp_file"
                done
                added_jobs=true
            fi
        fi
    done < "$prometheus_file"
    
    # Если jobs не были добавлены, добавляем в конец
    if [[ "$added_jobs" == "false" ]]; then
        for job in "${siem_jobs[@]}"; do
            echo "$job" >> "$temp_file"
        done
    fi
    
    mv "$temp_file" "$prometheus_file"
    log_success "Prometheus конфигурация обновлена"
}

# Интеграция с AlertManager
integrate_alertmanager() {
    log_info "Интеграция с AlertManager..."
    
    local alertmanager_file="deploy/alertmanager/alertmanager.yml"
    
    # Проверка существования файла
    if [[ ! -f "$alertmanager_file" ]]; then
        log_error "Файл $alertmanager_file не найден"
        return 1
    fi
    
    # Создание временного файла с новыми receivers
    local temp_file=$(mktemp)
    local in_receivers=false
    local added_siem_receivers=false
    
    while IFS= read -r line; do
        echo "$line" >> "$temp_file"
        
        if [[ "$line" == "receivers:" ]]; then
            in_receivers=true
        elif [[ "$in_receivers" == "true" && "$line" =~ ^[[:space:]]*-[[:space:]]*name: ]]; then
            if [[ "$added_siem_receivers" == "false" ]]; then
                # Добавляем SIEM receivers перед первым существующим receiver
                cat >> "$temp_file" << 'EOF'
- name: 'siem-critical'
  email_configs:
  - to: "security@2gc.ru"
    from: "siem@2gc.ru"
    smarthost: "smtp.rusender.ru:465"
    auth_username: "siem@2gc.ru"
    auth_password: "${SIEM_EMAIL_PASSWORD}"
    send_resolved: true
    require_tls: false
    tls_config:
      insecure_skip_verify: false
      server_name: "smtp.rusender.ru"
  telegram_configs:
  - bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    send_resolved: true

- name: 'siem-warning'
  email_configs:
  - to: "admin@2gc.ru"
    from: "siem@2gc.ru"
    smarthost: "smtp.rusender.ru:465"
    auth_username: "siem@2gc.ru"
    auth_password: "${SIEM_EMAIL_PASSWORD}"
    send_resolved: true
    require_tls: false
    tls_config:
      insecure_skip_verify: false
      server_name: "smtp.rusender.ru"

EOF
                added_siem_receivers=true
            fi
        fi
    done < "$alertmanager_file"
    
    # Если receivers не были добавлены, добавляем в конец
    if [[ "$added_siem_receivers" == "false" ]]; then
        cat >> "$temp_file" << 'EOF'
- name: 'siem-critical'
  email_configs:
  - to: "security@2gc.ru"
    from: "siem@2gc.ru"
    smarthost: "smtp.rusender.ru:465"
    auth_username: "siem@2gc.ru"
    auth_password: "${SIEM_EMAIL_PASSWORD}"
    send_resolved: true
    require_tls: false
    tls_config:
      insecure_skip_verify: false
      server_name: "smtp.rusender.ru"
  telegram_configs:
  - bot_token: "${TELEGRAM_BOT_TOKEN}"
    chat_id: "${TELEGRAM_CHAT_ID}"
    send_resolved: true

- name: 'siem-warning'
  email_configs:
  - to: "admin@2gc.ru"
    from: "siem@2gc.ru"
    smarthost: "smtp.rusender.ru:465"
    auth_username: "siem@2gc.ru"
    auth_password: "${SIEM_EMAIL_PASSWORD}"
    send_resolved: true
    require_tls: false
    tls_config:
      insecure_skip_verify: false
      server_name: "smtp.rusender.ru"

EOF
    fi
    
    mv "$temp_file" "$alertmanager_file"
    log_success "AlertManager конфигурация обновлена"
}

# Создание SIEM алертов для Prometheus
create_siem_alerts() {
    log_info "Создание SIEM алертов для Prometheus..."
    
    local alerts_file="deploy/prometheus/alerts/siem_alerts.yml"
    
    # Создание директории если не существует
    mkdir -p "$(dirname "$alerts_file")"
    
    # Создание файла с алертами
    cat > "$alerts_file" << 'EOF'
groups:
  - name: siem_alerts
    rules:
      # SIEM Backend API Down
      - alert: SIEMBackendDown
        expr: up{job="siem-backend"} == 0
        for: 1m
        labels:
          severity: critical
          service: siem
        annotations:
          summary: "SIEM Backend API недоступен"
          description: "SIEM Backend API на {{ $labels.instance }} недоступен более 1 минуты"

      # SIEM Event Processor Down
      - alert: SIEMProcessorDown
        expr: up{job="siem-processor"} == 0
        for: 1m
        labels:
          severity: critical
          service: siem
        annotations:
          summary: "SIEM Event Processor недоступен"
          description: "SIEM Event Processor на {{ $labels.instance }} недоступен более 1 минуты"

      # SIEM Alert Manager Down
      - alert: SIEMAlertManagerDown
        expr: up{job="siem-alerts"} == 0
        for: 1m
        labels:
          severity: critical
          service: siem
        annotations:
          summary: "SIEM Alert Manager недоступен"
          description: "SIEM Alert Manager на {{ $labels.instance }} недоступен более 1 минуты"

      # High SIEM Event Processing Rate
      - alert: HighSIEMEventRate
        expr: rate(siem_events_processed_total[5m]) > 10000
        for: 2m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокая скорость обработки событий SIEM"
          description: "SIEM обрабатывает {{ $value }} событий в секунду"

      # SIEM Alert Generation Rate
      - alert: HighSIEMAlertRate
        expr: rate(siem_alerts_generated_total[5m]) > 100
        for: 2m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокая скорость генерации алертов SIEM"
          description: "SIEM генерирует {{ $value }} алертов в секунду"

      # SIEM Database Connection Issues
      - alert: SIEMDatabaseConnectionIssues
        expr: siem_database_connection_errors_total > 0
        for: 1m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Проблемы с подключением к базе данных SIEM"
          description: "Обнаружены ошибки подключения к базе данных SIEM"

      # SIEM Elasticsearch Health
      - alert: SIEMElasticsearchUnhealthy
        expr: siem_elasticsearch_health_status != 1
        for: 2m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Elasticsearch SIEM нездоров"
          description: "Elasticsearch кластер SIEM находится в нездоровом состоянии"

      # SIEM Kafka Lag
      - alert: SIEMKafkaLag
        expr: siem_kafka_consumer_lag > 1000
        for: 5m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокий lag в Kafka SIEM"
          description: "Kafka consumer lag составляет {{ $value }} сообщений"

      # SIEM Memory Usage
      - alert: SIEMMemoryUsage
        expr: (siem_memory_usage_bytes / siem_memory_total_bytes) * 100 > 80
        for: 5m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокое потребление памяти SIEM"
          description: "SIEM использует {{ $value }}% памяти"

      # SIEM Disk Usage
      - alert: SIEMDiskUsage
        expr: (siem_disk_usage_bytes / siem_disk_total_bytes) * 100 > 85
        for: 5m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокое использование диска SIEM"
          description: "SIEM использует {{ $value }}% дискового пространства"

      # SIEM Authentication Failures
      - alert: SIEMAuthenticationFailures
        expr: rate(siem_authentication_failures_total[5m]) > 10
        for: 2m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Высокое количество неудачных аутентификаций SIEM"
          description: "{{ $value }} неудачных попыток аутентификации в секунду"

      # SIEM Correlation Rule Performance
      - alert: SIEMCorrelationRuleSlow
        expr: histogram_quantile(0.95, rate(siem_correlation_rule_duration_seconds_bucket[5m])) > 5
        for: 5m
        labels:
          severity: warning
          service: siem
        annotations:
          summary: "Медленная работа правил корреляции SIEM"
          description: "95-й процентиль времени выполнения правил корреляции {{ $value }} секунд"

EOF
    
    log_success "SIEM алерты созданы в $alerts_file"
}

# Добавление SIEM алертов в Prometheus
add_siem_alerts_to_prometheus() {
    log_info "Добавление SIEM алертов в Prometheus..."
    
    local prometheus_file="prometheus/prometheus.yml"
    
    # Проверка существования файла
    if [[ ! -f "$prometheus_file" ]]; then
        log_error "Файл $prometheus_file не найден"
        return 1
    fi
    
    # Добавление rule_files если не существует
    if ! grep -q "rule_files:" "$prometheus_file"; then
        # Добавляем после global секции
        sed -i '/^global:/a rule_files:\n  - "alerts/siem_alerts.yml"' "$prometheus_file"
    else
        # Добавляем в существующую секцию rule_files
        if ! grep -q "siem_alerts.yml" "$prometheus_file"; then
            sed -i '/rule_files:/a \  - "alerts/siem_alerts.yml"' "$prometheus_file"
        fi
    fi
    
    log_success "SIEM алерты добавлены в Prometheus"
}

# Создание SIEM дашбордов для Grafana
create_siem_dashboards() {
    log_info "Создание SIEM дашбордов для Grafana..."
    
    local dashboards_dir="deploy/grafana/dashboards"
    
    # Создание директории если не существует
    mkdir -p "$dashboards_dir"
    
    # SIEM Events Dashboard
    cat > "$dashboards_dir/siem_events_dashboard.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "SIEM Events Dashboard",
    "tags": ["siem", "events", "security"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Events per Second",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(siem_events_processed_total[5m])",
            "legendFormat": "Events/sec"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "custom": {
              "displayMode": "gradient"
            },
            "mappings": [],
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {
                  "color": "green",
                  "value": null
                },
                {
                  "color": "red",
                  "value": 80
                }
              ]
            }
          }
        },
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 0
        }
      },
      {
        "id": 2,
        "title": "Events by Type",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by (event_type) (rate(siem_events_processed_total[5m]))",
            "legendFormat": "{{event_type}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 12,
          "y": 0
        }
      },
      {
        "id": 3,
        "title": "Events Timeline",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(siem_events_processed_total[1m])",
            "legendFormat": "Events/sec"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 24,
          "x": 0,
          "y": 8
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF
    
    # SIEM Alerts Dashboard
    cat > "$dashboards_dir/siem_alerts_dashboard.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "SIEM Alerts Dashboard",
    "tags": ["siem", "alerts", "security"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Alerts per Second",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(siem_alerts_generated_total[5m])",
            "legendFormat": "Alerts/sec"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "custom": {
              "displayMode": "gradient"
            },
            "mappings": [],
            "thresholds": {
              "mode": "absolute",
              "steps": [
                {
                  "color": "green",
                  "value": null
                },
                {
                  "color": "yellow",
                  "value": 10
                },
                {
                  "color": "red",
                  "value": 50
                }
              ]
            }
          }
        },
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 0
        }
      },
      {
        "id": 2,
        "title": "Alerts by Severity",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by (severity) (rate(siem_alerts_generated_total[5m]))",
            "legendFormat": "{{severity}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 12,
          "y": 0
        }
      },
      {
        "id": 3,
        "title": "Alerts Timeline",
        "type": "timeseries",
        "targets": [
          {
            "expr": "rate(siem_alerts_generated_total[1m])",
            "legendFormat": "Alerts/sec"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 24,
          "x": 0,
          "y": 8
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "5s"
  }
}
EOF
    
    # SIEM Security Dashboard
    cat > "$dashboards_dir/siem_security_dashboard.json" << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "SIEM Security Dashboard",
    "tags": ["siem", "security", "overview"],
    "style": "dark",
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Security Events Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(siem_events_processed_total[5m]))",
            "legendFormat": "Total Events/sec"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 0,
          "y": 0
        }
      },
      {
        "id": 2,
        "title": "Active Alerts",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(siem_alerts_generated_total[5m]))",
            "legendFormat": "Active Alerts/sec"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 6,
          "y": 0
        }
      },
      {
        "id": 3,
        "title": "Authentication Failures",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(siem_authentication_failures_total[5m]))",
            "legendFormat": "Auth Failures/sec"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 12,
          "y": 0
        }
      },
      {
        "id": 4,
        "title": "System Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=~\"siem-.*\"}",
            "legendFormat": "{{job}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 18,
          "y": 0
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "10s"
  }
}
EOF
    
    log_success "SIEM дашборды созданы в $dashboards_dir"
}

# Перезапуск сервисов мониторинга
restart_monitoring_services() {
    log_info "Перезапуск сервисов мониторинга..."
    
    # Перезапуск Prometheus
    if docker ps | grep -q prometheus; then
        docker restart prometheus
        log_success "Prometheus перезапущен"
    else
        log_warning "Prometheus контейнер не найден"
    fi
    
    # Перезапуск AlertManager
    if docker ps | grep -q alertmanager; then
        docker restart alertmanager
        log_success "AlertManager перезапущен"
    else
        log_warning "AlertManager контейнер не найден"
    fi
    
    # Перезапуск Grafana
    if docker ps | grep -q grafana; then
        docker restart grafana
        log_success "Grafana перезапущен"
    else
        log_warning "Grafana контейнер не найден"
    fi
}

# Проверка интеграции
verify_integration() {
    log_info "Проверка интеграции..."
    
    # Проверка Prometheus
    if curl -s http://localhost:9090/api/v1/targets | jq -e '.data.activeTargets[] | select(.labels.job == "siem-backend")' > /dev/null; then
        log_success "SIEM Backend target добавлен в Prometheus"
    else
        log_warning "SIEM Backend target не найден в Prometheus"
    fi
    
    # Проверка AlertManager
    if curl -s http://localhost:9093/api/v1/receivers | jq -e '.data[] | select(.name == "siem-critical")' > /dev/null; then
        log_success "SIEM receivers добавлены в AlertManager"
    else
        log_warning "SIEM receivers не найдены в AlertManager"
    fi
    
    # Проверка доступности SIEM сервера
    if curl -s "http://${SIEM_SERVER_IP}:${SIEM_SERVER_PORT}/health" > /dev/null; then
        log_success "SIEM сервер доступен"
    else
        log_warning "SIEM сервер недоступен (проверьте настройки)"
    fi
}

# Основная функция
main() {
    log_info " Начало интеграции SIEM с существующим мониторингом"
    
    # Проверки
    check_root
    check_dependencies
    parse_arguments "$@"
    
    # Интеграция
    create_backups
    integrate_prometheus
    integrate_alertmanager
    create_siem_alerts
    add_siem_alerts_to_prometheus
    create_siem_dashboards
    restart_monitoring_services
    verify_integration
    
    log_success "Интеграция SIEM с мониторингом завершена успешно!"
    
    cat << EOF

🎉 Интеграция завершена!

📊 Доступные дашборды:
   - SIEM Events Dashboard: http://localhost:3000/d/siem-events
   - SIEM Alerts Dashboard: http://localhost:3000/d/siem-alerts  
   - SIEM Security Dashboard: http://localhost:3000/d/siem-security

 Полезные ссылки:
   - Prometheus: http://localhost:9090
   - AlertManager: http://localhost:9093
   - Grafana: http://localhost:3000

  Важные замечания:
   - Убедитесь, что SIEM сервер доступен по адресу ${SIEM_SERVER_IP}:${SIEM_SERVER_PORT}
   - Проверьте настройки уведомлений в AlertManager
   - Настройте переменные окружения для email и Telegram уведомлений

📚 Документация:
   - Техническое задание: siem/TECHNICAL_SPECIFICATION.md
   - Руководство по интеграции: docs/integration/siem_integration_guide.md

EOF
}

# Запуск скрипта
main "$@" 