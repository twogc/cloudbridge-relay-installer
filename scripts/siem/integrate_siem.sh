#!/bin/bash

# CloudBridge Relay SIEM Integration Script
# Автоматическая интеграция SIEM с CloudBridge Relay v2.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    # Проверка Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен"
        exit 1
    fi
    
    # Проверка Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose не установлен"
        exit 1
    fi
    
    # Проверка Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python 3 не установлен"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Проверка системных ресурсов
check_resources() {
    log_info "Проверка системных ресурсов..."
    
    # Проверка RAM
    total_ram=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    if [ "$total_ram" -lt 8 ]; then
        log_warning "Рекомендуется минимум 8GB RAM (доступно: ${total_ram}GB)"
    else
        log_success "RAM: ${total_ram}GB"
    fi
    
    # Проверка свободного места
    free_space=$(df -BG . | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$free_space" -lt 50 ]; then
        log_warning "Рекомендуется минимум 50GB свободного места (доступно: ${free_space}GB)"
    else
        log_success "Свободное место: ${free_space}GB"
    fi
}

# Создание конфигурации SIEM
setup_siem_config() {
    log_info "Настройка конфигурации SIEM..."
    
    # Создание директорий
    mkdir -p siem/{config,logs,data,backup}
    
    # Копирование конфигурации
    if [ -f "siem/config/siem_config.yaml" ]; then
        cp siem/config/siem_config.yaml siem/config/siem_config.yaml.backup
    fi
    
    # Создание базовой конфигурации
    cat > siem/config/siem_config.yaml << 'EOF'
siem:
  environment: "production"
  log_level: "INFO"
  
  collection:
    relay_logs_interval: 60
    relay_metrics_interval: 30
    keycloak_events_interval: 120
    django_audit_interval: 60
    
  processing:
    batch_size: 1000
    max_workers: 4
    retry_attempts: 3
    
  storage:
    elasticsearch:
      hosts: ["http://elasticsearch:9200"]
      index_prefix: "siem"
      retention_days: 90
      
  alerting:
    min_severity: "medium"
    channels:
      telegram:
        enabled: true
        bot_token: "${TELEGRAM_BOT_TOKEN}"
        chat_id: "${TELEGRAM_CHAT_ID}"
      email:
        enabled: false
        smtp_server: "${EMAIL_SMTP_SERVER}"
        username: "${EMAIL_USERNAME}"
        password: "${EMAIL_PASSWORD}"
EOF
    
    log_success "Конфигурация SIEM создана"
}

# Настройка multi-tenant конфигурации
setup_multi_tenant() {
    log_info "Настройка multi-tenant конфигурации..."
    
    # Создание конфигурации тенантов
    cat > siem/config/tenants.yaml << 'EOF'
tenants:
  tenant_001:
    name: "Production Company A"
    enabled: true
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
      max_data_transfer_gb: 5000
      rate_limit_per_second: 2000
    allowed_ips:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
    metadata:
      contact: "admin@company-a.com"
      environment: "production"
      billing_tier: "premium"
    features:
      tls_enabled: true
      monitoring_enabled: true
      backup_enabled: true

  tenant_002:
    name: "Development Company B"
    enabled: true
    limits:
      max_tunnels: 5
      max_connections_per_tunnel: 50
      max_bandwidth_mbps: 100
      max_data_transfer_gb: 1000
      rate_limit_per_second: 500
    allowed_ips:
      - "192.168.2.0/24"
    metadata:
      contact: "dev@company-b.com"
      environment: "development"
      billing_tier: "basic"
    features:
      tls_enabled: true
      monitoring_enabled: false
      backup_enabled: false
EOF
    
    log_success "Multi-tenant конфигурация создана"
}

# Настройка правил детекции
setup_detection_rules() {
    log_info "Настройка правил детекции..."
    
    cat > siem/config/detection_rules.yaml << 'EOF'
detection_rules:
  tenant_specific:
    brute_force:
      enabled: true
      conditions:
        event_type: "authentication_failure"
        time_window: 300
        threshold: 5
        group_by: ["source_ip", "tenant_id"]
      actions:
        - type: "alert"
          severity: "high"
          message: "Brute force attack detected for tenant {tenant_id}"
        - type: "block_ip"
          duration: 3600
          
    data_exfiltration:
      enabled: true
      conditions:
        event_type: "data_transfer"
        data_volume_gb: 100
        time_window: 3600
        group_by: ["tenant_id", "user_id"]
      actions:
        - type: "alert"
          severity: "critical"
          message: "Potential data exfiltration detected"
        - type: "quarantine_user"
          duration: 7200
          
  cross_tenant:
    shared_ip_attacks:
      enabled: true
      conditions:
        event_type: "authentication_failure"
        correlation_window: 600
        min_tenants_affected: 2
        group_by: ["source_ip"]
      actions:
        - type: "global_alert"
          severity: "critical"
          message: "Cross-tenant attack detected from IP {source_ip}"
        - type: "global_block_ip"
          duration: 86400
EOF
    
    log_success "Правила детекции настроены"
}

# Настройка алертов
setup_alerting() {
    log_info "Настройка системы алертов..."
    
    cat > siem/config/alerting.yaml << 'EOF'
alerting:
  channels:
    telegram:
      enabled: true
      bot_token: "${TELEGRAM_BOT_TOKEN}"
      chat_id: "${TELEGRAM_CHAT_ID}"
      template: |
        {severity} Alert: {title}
        
        Details:
        • Tenant: {tenant_id}
        • Source: {source_ip}
        • User: {user_id}
        • Time: {timestamp}
        
        📝 Description: {description}
        
         View in SIEM: {siem_url}
        
    email:
      enabled: true
      smtp_server: "${EMAIL_SMTP_SERVER}"
      smtp_port: 587
      username: "${EMAIL_USERNAME}"
      password: "${EMAIL_PASSWORD}"
      from_address: "siem@yourcompany.com"
      to_addresses: ["security@yourcompany.com", "admin@yourcompany.com"]
      subject_template: "[SIEM Alert] {severity}: {title}"
      
  escalation:
    enabled: true
    levels:
      - level: 1
        delay_minutes: 5
        channels: ["telegram"]
      - level: 2
        delay_minutes: 15
        channels: ["telegram", "email"]
      - level: 3
        delay_minutes: 30
        channels: ["telegram", "email", "slack"]
EOF
    
    log_success "Система алертов настроена"
}

# Установка зависимостей Python
install_python_deps() {
    log_info "Установка Python зависимостей..."
    
    if [ -f "siem/requirements.txt" ]; then
        cd siem
        pip3 install -r requirements.txt
        cd ..
        log_success "Python зависимости установлены"
    else
        log_warning "Файл requirements.txt не найден"
    fi
}

# Запуск SIEM сервисов
start_siem_services() {
    log_info "Запуск SIEM сервисов..."
    
    cd siem
    
    # Проверка наличия docker-compose файла
    if [ ! -f "docker-compose.siem.yml" ]; then
        log_error "Файл docker-compose.siem.yml не найден"
        exit 1
    fi
    
    # Запуск сервисов
    docker-compose -f docker-compose.siem.yml up -d
    
    # Ожидание запуска сервисов
    log_info "Ожидание запуска сервисов..."
    sleep 30
    
    # Проверка статуса
    docker-compose -f docker-compose.siem.yml ps
    
    cd ..
    
    log_success "SIEM сервисы запущены"
}

# Настройка интеграции с Relay
setup_relay_integration() {
    log_info "Настройка интеграции с CloudBridge Relay..."
    
    # Создание конфигурации для Relay
    cat > config/siem_integration.json << 'EOF'
{
  "siem": {
    "enabled": true,
    "endpoint": "http://localhost:8081/api/v1/events",
    "timeout": "10s",
    "retry_attempts": 3,
    "batch_size": 100,
    "flush_interval": "5s"
  },
  "events": {
    "tunnel_connection": true,
    "authentication": true,
    "authorization": true,
    "data_transfer": true,
    "errors": true
  },
  "tenant_mapping": {
    "enabled": true,
    "default_tenant": "default"
  }
}
EOF
    
    log_success "Интеграция с Relay настроена"
}

# Настройка мониторинга
setup_monitoring() {
    log_info "Настройка мониторинга..."
    
    # Создание конфигурации Prometheus для SIEM
    cat > prometheus/siem_targets.yml << 'EOF'
scrape_configs:
  - job_name: 'siem-elasticsearch'
    static_configs:
      - targets: ['elasticsearch:9200']
    metrics_path: '/_prometheus/metrics'
    
  - job_name: 'siem-kafka'
    static_configs:
      - targets: ['kafka:9092']
    metrics_path: '/metrics'
    
  - job_name: 'siem-collector'
    static_configs:
      - targets: ['siem-collector:8080']
    metrics_path: '/metrics'
    
  - job_name: 'siem-processor'
    static_configs:
      - targets: ['siem-processor:8080']
    metrics_path: '/metrics'
EOF
    
    log_success "Мониторинг настроен"
}

# Создание дашбордов
setup_dashboards() {
    log_info "Создание дашбордов..."
    
    # Создание директории для дашбордов
    mkdir -p grafana/dashboards
    
    # SIEM Overview дашборд
    cat > grafana/dashboards/siem_overview.json << 'EOF'
{
  "dashboard": {
    "title": "SIEM Overview - Multi-tenant",
    "panels": [
      {
        "title": "Events per Tenant",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(siem_events_total[5m])) by (tenant_id)",
            "legendFormat": "{{tenant_id}}"
          }
        ]
      },
      {
        "title": "Alerts by Severity",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum(siem_alerts_total) by (severity)",
            "legendFormat": "{{severity}}"
          }
        ]
      }
    ]
  }
}
EOF
    
    log_success "Дашборды созданы"
}

# Тестирование интеграции
test_integration() {
    log_info "Тестирование интеграции..."
    
    # Проверка доступности сервисов
    services=("elasticsearch:9200" "kafka:9092" "kibana:5601")
    
    for service in "${services[@]}"; do
        if curl -s "http://$service" > /dev/null 2>&1; then
            log_success "Сервис $service доступен"
        else
            log_warning "Сервис $service недоступен"
        fi
    done
    
    # Тест отправки события
    test_event='{
      "tenant_id": "test_tenant",
      "event_type": "test_event",
      "severity": "medium",
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
      "source_ip": "192.168.1.100",
      "user_id": "test_user",
      "action": "test_action"
    }'
    
    if curl -s -X POST "http://localhost:8081/api/v1/events" \
        -H "Content-Type: application/json" \
        -d "$test_event" > /dev/null 2>&1; then
        log_success "Тест отправки события прошел успешно"
    else
        log_warning "Тест отправки события не прошел"
    fi
    
    log_success "Тестирование завершено"
}

# Создание systemd сервиса
setup_systemd_service() {
    log_info "Настройка systemd сервиса..."
    
    cat > systemd/siem.service << 'EOF'
[Unit]
Description=CloudBridge SIEM System
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/home/ubuntu/cloudbridge-relay-installer/siem
ExecStart=/usr/bin/docker-compose -f docker-compose.siem.yml up -d
ExecStop=/usr/bin/docker-compose -f docker-compose.siem.yml down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
    
    # Копирование сервиса
    sudo cp systemd/siem.service /etc/systemd/system/
    sudo systemctl daemon-reload
    sudo systemctl enable siem.service
    
    log_success "Systemd сервис настроен"
}

# Основная функция
main() {
    echo "=========================================="
    echo "CloudBridge Relay SIEM Integration Script"
    echo "=========================================="
    echo ""
    
    # Проверка аргументов
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --full          Полная интеграция (по умолчанию)"
        echo "  --basic         Базовая интеграция"
        echo "  --multi-tenant  Только multi-tenant настройка"
        echo "  --test          Только тестирование"
        echo "  --help, -h      Показать эту справку"
        echo ""
        exit 0
    fi
    
    # Проверка зависимостей
    check_dependencies
    
    # Проверка ресурсов
    check_resources
    
    # Определение режима интеграции
    mode=${1:-"--full"}
    
    case $mode in
        "--full")
            log_info "Запуск полной интеграции SIEM..."
            setup_siem_config
            setup_multi_tenant
            setup_detection_rules
            setup_alerting
            install_python_deps
            start_siem_services
            setup_relay_integration
            setup_monitoring
            setup_dashboards
            setup_systemd_service
            test_integration
            ;;
        "--basic")
            log_info "Запуск базовой интеграции SIEM..."
            setup_siem_config
            install_python_deps
            start_siem_services
            test_integration
            ;;
        "--multi-tenant")
            log_info "Настройка multi-tenant конфигурации..."
            setup_multi_tenant
            setup_detection_rules
            setup_alerting
            ;;
        "--test")
            log_info "Запуск тестирования..."
            test_integration
            ;;
        *)
            log_error "Неизвестная опция: $mode"
            echo "Используйте --help для получения справки"
            exit 1
            ;;
    esac
    
    echo ""
    echo "=========================================="
    log_success "Интеграция SIEM завершена успешно!"
    echo "=========================================="
    echo ""
    echo "Доступные сервисы:"
    echo "• SIEM Web UI: http://localhost:8081"
    echo "• Kibana: http://localhost:5601"
    echo "• Grafana: http://localhost:3001"
    echo "• Prometheus: http://localhost:9091"
    echo "• Elasticsearch: http://localhost:9200"
    echo ""
    echo "Управление сервисом:"
    echo "• sudo systemctl start siem"
    echo "• sudo systemctl stop siem"
    echo "• sudo systemctl status siem"
    echo ""
    echo "Логи:"
    echo "• sudo journalctl -u siem.service -f"
    echo "• tail -f siem/logs/*.log"
    echo ""
}

# Запуск скрипта
main "$@" 