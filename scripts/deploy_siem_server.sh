#!/bin/bash

# CloudBridge SIEM Server Deployment Script
# Автоматическое развертывание SIEM на отдельном сервере

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

# Проверка root прав
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "Этот скрипт должен выполняться с правами root"
        log_info "Используйте: sudo $0"
        exit 1
    fi
}

# Проверка системных требований
check_requirements() {
    log_info "Проверка системных требований..."
    
    # Проверка ОС
    if ! grep -q "Ubuntu" /etc/os-release; then
        log_warning "Рекомендуется Ubuntu 20.04+ или 22.04+"
    fi
    
    # Проверка RAM
    total_ram=$(free -m | awk 'NR==2{printf "%.0f", $2/1024}')
    if [ "$total_ram" -lt 16 ]; then
        log_error "Требуется минимум 16GB RAM (доступно: ${total_ram}GB)"
        exit 1
    else
        log_success "RAM: ${total_ram}GB"
    fi
    
    # Проверка свободного места
    free_space=$(df -BG / | awk 'NR==2{print $4}' | sed 's/G//')
    if [ "$free_space" -lt 500 ]; then
        log_error "Требуется минимум 500GB свободного места (доступно: ${free_space}GB)"
        exit 1
    else
        log_success "Свободное место: ${free_space}GB"
    fi
    
    # Проверка архитектуры
    if [ "$(uname -m)" != "x86_64" ]; then
        log_error "Поддерживается только x86_64 архитектура"
        exit 1
    fi
}

# Установка базовых пакетов
install_base_packages() {
    log_info "Установка базовых пакетов..."
    
    # Обновление системы
    apt update && apt upgrade -y
    
    # Установка необходимых пакетов
    apt install -y \
        curl \
        wget \
        git \
        htop \
        iotop \
        nethogs \
        ufw \
        fail2ban \
        logrotate \
        rsyslog \
        python3 \
        python3-pip \
        python3-venv \
        docker.io \
        docker-compose \
        nginx \
        certbot \
        python3-certbot-nginx \
        auditd \
        unzip \
        jq \
        tree
    
    log_success "Базовые пакеты установлены"
}

# Настройка безопасности
setup_security() {
    log_info "Настройка безопасности..."
    
    # Настройка firewall
    ufw default deny incoming
    ufw default allow outgoing
    ufw allow ssh
    ufw allow 80/tcp
    ufw allow 443/tcp
    ufw allow 5601/tcp  # Kibana
    ufw allow 9200/tcp  # Elasticsearch
    ufw allow 9092/tcp  # Kafka
    ufw allow 8081/tcp  # SIEM Web UI
    ufw allow 9090/tcp  # Prometheus
    ufw allow 3000/tcp  # Grafana
    ufw enable
    
    # Настройка fail2ban
    systemctl enable fail2ban
    systemctl start fail2ban
    
    # Настройка auditd
    systemctl enable auditd
    systemctl start auditd
    
    log_success "Безопасность настроена"
}

# Настройка Docker
setup_docker() {
    log_info "Настройка Docker..."
    
    # Добавление пользователя в группу docker
    usermod -aG docker ubuntu
    
    # Настройка Docker daemon
    tee /etc/docker/daemon.json << EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ],
  "default-ulimits": {
    "nofile": {
      "Hard": 64000,
      "Name": "nofile",
      "Soft": 64000
    }
  }
}
EOF
    
    systemctl restart docker
    
    log_success "Docker настроен"
}

# Создание пользователя SIEM
create_siem_user() {
    log_info "Создание пользователя SIEM..."
    
    # Создание пользователя siem
    useradd -m -s /bin/bash siem
    usermod -aG docker siem
    
    # Создание директорий
    mkdir -p /opt/siem
    chown siem:siem /opt/siem
    
    # Создание SSH директории
    mkdir -p /opt/siem/.ssh
    chown siem:siem /opt/siem/.ssh
    chmod 700 /opt/siem/.ssh
    
    log_success "Пользователь SIEM создан"
}

# Установка SIEM
install_siem() {
    log_info "Установка SIEM..."
    
    cd /opt/siem
    
    # Клонирование репозитория
    sudo -u siem git clone https://github.com/maxlanies/cloudbridge-relay-installer.git .
    
    # Создание конфигурационных директорий
    sudo -u siem mkdir -p {config,logs,data,backup,scripts}
    
    log_success "SIEM установлен"
}

# Настройка переменных окружения
setup_environment() {
    log_info "Настройка переменных окружения..."
    
    # Получение IP адреса сервера
    SERVER_IP=$(hostname -I | awk '{print $1}')
    
    # Создание .env файла
    sudo -u siem tee /opt/siem/.env.siem << EOF
# Elasticsearch
ELASTICSEARCH_PASSWORD=changeme_secure_password_here
ELASTICSEARCH_USERNAME=elastic

# SIEM Configuration
SIEM_ENVIRONMENT=production
SIEM_LOG_LEVEL=INFO

# Network Configuration
SIEM_HOST=${SERVER_IP}
RELAY_HOST=your_relay_server_ip
KEYCLOAK_HOST=your_keycloak_server_ip
DJANGO_HOST=your_django_server_ip

# Telegram Notifications (опционально)
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here

# Email Notifications (опционально)
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=your_email@gmail.com
EMAIL_PASSWORD=your_app_password_here
EMAIL_FROM=siem@yourcompany.com
EMAIL_TO=security@yourcompany.com,admin@yourcompany.com

# Security
SSL_CERT_PATH=/etc/ssl/siem/siem.crt
SSL_KEY_PATH=/etc/ssl/siem/siem.key
EOF
    
    log_success "Переменные окружения настроены"
}

# Генерация SSL сертификатов
generate_ssl_certs() {
    log_info "Генерация SSL сертификатов..."
    
    # Создание директории для сертификатов
    mkdir -p /etc/ssl/siem
    cd /etc/ssl/siem
    
    # Генерация самоподписанных сертификатов
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout siem.key -out siem.crt \
        -subj "/C=RU/ST=Moscow/L=Moscow/O=YourCompany/OU=IT/CN=siem.yourcompany.com"
    
    chmod 600 siem.key
    chmod 644 siem.crt
    
    log_success "SSL сертификаты сгенерированы"
}

# Настройка конфигурации SIEM
setup_siem_config() {
    log_info "Настройка конфигурации SIEM..."
    
    # Создание конфигурации тенантов
    sudo -u siem tee /opt/siem/siem/config/tenants.yaml << 'EOF'
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
    
    # Создание правил детекции
    sudo -u siem tee /opt/siem/siem/config/detection_rules.yaml << 'EOF'
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
    
    log_success "Конфигурация SIEM настроена"
}

# Настройка Nginx
setup_nginx() {
    log_info "Настройка Nginx..."
    
    # Создание конфигурации Nginx
    tee /etc/nginx/sites-available/siem << 'EOF'
server {
    listen 80;
    server_name siem.yourcompany.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name siem.yourcompany.com;

    ssl_certificate /etc/ssl/siem/siem.crt;
    ssl_certificate_key /etc/ssl/siem/siem.key;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # Security headers
    add_header X-Frame-Options DENY;
    add_header X-Content-Type-Options nosniff;
    add_header X-XSS-Protection "1; mode=block";
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # SIEM Web UI
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Kibana
    location /kibana/ {
        proxy_pass http://localhost:5601/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Grafana
    location /grafana/ {
        proxy_pass http://localhost:3001/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Elasticsearch API (только для внутреннего доступа)
    location /elasticsearch/ {
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        allow 192.168.0.0/16;
        deny all;
        
        proxy_pass http://localhost:9200/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
    
    # Активация конфигурации
    ln -sf /etc/nginx/sites-available/siem /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    
    # Проверка конфигурации
    nginx -t
    
    # Перезапуск Nginx
    systemctl restart nginx
    
    log_success "Nginx настроен"
}

# Настройка мониторинга
setup_monitoring() {
    log_info "Настройка мониторинга..."
    
    # Создание директории для Prometheus
    mkdir -p /opt/prometheus
    
    # Создание конфигурации Prometheus
    tee /opt/prometheus/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "siem_alerts.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['localhost:9100']

  - job_name: 'siem-elasticsearch'
    static_configs:
      - targets: ['localhost:9200']
    metrics_path: '/_prometheus/metrics'

  - job_name: 'siem-kafka'
    static_configs:
      - targets: ['localhost:9092']
    metrics_path: '/metrics'

  - job_name: 'siem-collector'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'

  - job_name: 'siem-processor'
    static_configs:
      - targets: ['localhost:8081']
    metrics_path: '/metrics'
EOF
    
    # Создание правил алертов
    tee /opt/prometheus/siem_alerts.yml << 'EOF'
groups:
  - name: siem_alerts
    rules:
      - alert: SIEMHighCPU
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage on SIEM server"
          
      - alert: SIEMHighMemory
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on SIEM server"
          
      - alert: SIEMHighDisk
        expr: (node_filesystem_size_bytes - node_filesystem_free_bytes) / node_filesystem_size_bytes * 100 > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High disk usage on SIEM server"
          
      - alert: SIEMServiceDown
        expr: up{job=~"siem.*"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "SIEM service is down"
EOF
    
    # Запуск Prometheus Node Exporter
    docker run -d \
        --name prometheus-node-exporter \
        --restart unless-stopped \
        -p 9100:9100 \
        -v "/:/host:ro,rslave" \
        prom/node-exporter:latest \
        --path.rootfs=/host
    
    # Запуск Prometheus
    docker run -d \
        --name prometheus \
        --restart unless-stopped \
        -p 9090:9090 \
        -v /opt/prometheus:/etc/prometheus \
        prom/prometheus:latest \
        --config.file=/etc/prometheus/prometheus.yml
    
    # Запуск Grafana
    docker run -d \
        --name grafana \
        --restart unless-stopped \
        -p 3000:3000 \
        -e GF_SECURITY_ADMIN_PASSWORD=admin \
        grafana/grafana:latest
    
    log_success "Мониторинг настроен"
}

# Создание systemd сервиса
create_systemd_service() {
    log_info "Создание systemd сервиса..."
    
    # Создание systemd сервиса
    tee /etc/systemd/system/siem.service << 'EOF'
[Unit]
Description=CloudBridge SIEM System
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/siem/siem
EnvironmentFile=/opt/siem/.env.siem
ExecStart=/usr/bin/docker-compose -f docker-compose.siem.yml up -d
ExecStop=/usr/bin/docker-compose -f docker-compose.siem.yml down
TimeoutStartSec=0
User=siem
Group=siem

[Install]
WantedBy=multi-user.target
EOF
    
    # Активация сервиса
    systemctl daemon-reload
    systemctl enable siem.service
    
    log_success "Systemd сервис создан"
}

# Создание скриптов управления
create_management_scripts() {
    log_info "Создание скриптов управления..."
    
    # Скрипт мониторинга
    sudo -u siem tee /opt/siem/scripts/monitor_siem.sh << 'EOF'
#!/bin/bash

echo "=== Использование ресурсов ==="
echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
echo "RAM: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')"
echo "Disk: $(df -h / | awk 'NR==2 {print $5}')"

echo -e "\n=== Статус контейнеров ==="
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo -e "\n=== Последние ошибки ==="
docker-compose -f /opt/siem/siem/docker-compose.siem.yml logs --tail=10 | grep -i error
EOF
    
    # Скрипт резервного копирования
    sudo -u siem tee /opt/siem/scripts/backup_siem.sh << 'EOF'
#!/bin/bash

BACKUP_DIR="/opt/siem/backups"
DATE=$(date +%Y%m%d_%H%M%S)

mkdir -p "$BACKUP_DIR"

# Резервное копирование конфигурации
tar -czf "$BACKUP_DIR/config_$DATE.tar.gz" \
    /opt/siem/siem/config \
    /opt/siem/.env.siem

# Резервное копирование данных Elasticsearch
curl -X PUT "localhost:9200/_snapshot/siem_backup/snapshot_$DATE" \
    -H "Content-Type: application/json" \
    -d '{
        "indices": "siem-*",
        "ignore_unavailable": true,
        "include_global_state": false
    }'

# Очистка старых резервных копий (оставить последние 7 дней)
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +7 -delete
EOF
    
    # Скрипт обновления
    sudo -u siem tee /opt/siem/scripts/update_siem.sh << 'EOF'
#!/bin/bash

cd /opt/siem

# Остановка сервисов
systemctl stop siem.service

# Обновление кода
git pull origin main

# Перезапуск сервисов
systemctl start siem.service

echo "SIEM обновлен успешно"
EOF
    
    # Установка прав на выполнение
    chmod +x /opt/siem/scripts/*.sh
    chown -R siem:siem /opt/siem/scripts
    
    # Добавление резервного копирования в cron
    (crontab -u siem -l 2>/dev/null; echo "0 2 * * * /opt/siem/scripts/backup_siem.sh") | crontab -u siem -
    
    log_success "Скрипты управления созданы"
}

# Настройка аудита
setup_audit() {
    log_info "Настройка аудита..."
    
    # Настройка аудита для SIEM директорий
    tee /etc/audit/rules.d/siem.rules << 'EOF'
-w /opt/siem -p wa -k siem_config
-w /opt/siem/siem/config -p wa -k siem_config
-w /opt/siem/.env.siem -p wa -k siem_secrets
-w /etc/ssl/siem -p wa -k siem_ssl
EOF
    
    # Настройка logrotate для SIEM логов
    tee /etc/logrotate.d/siem << 'EOF'
/opt/siem/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 siem siem
    postrotate
        systemctl reload siem
    endscript
}
EOF
    
    # Перезапуск auditd
    systemctl restart auditd
    
    log_success "Аудит настроен"
}

# Запуск SIEM
start_siem() {
    log_info "Запуск SIEM..."
    
    # Запуск сервиса
    systemctl start siem.service
    
    # Ожидание запуска
    sleep 30
    
    # Проверка статуса
    systemctl status siem.service --no-pager
    
    log_success "SIEM запущен"
}

# Тестирование
test_deployment() {
    log_info "Тестирование развертывания..."
    
    # Проверка сервисов
    services=(
        "http://localhost:8081"    # SIEM Web UI
        "http://localhost:5601"    # Kibana
        "http://localhost:3000"    # Grafana
        "http://localhost:9090"    # Prometheus
        "http://localhost:9200"    # Elasticsearch
    )
    
    for service in "${services[@]}"; do
        if curl -s "$service" > /dev/null 2>&1; then
            log_success "$service доступен"
        else
            log_warning "$service недоступен"
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

# Основная функция
main() {
    echo "=========================================="
    echo "CloudBridge SIEM Server Deployment Script"
    echo "=========================================="
    echo ""
    
    # Проверка аргументов
    if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --full          Полное развертывание (по умолчанию)"
        echo "  --minimal       Минимальное развертывание"
        echo "  --test          Только тестирование"
        echo "  --help, -h      Показать эту справку"
        echo ""
        exit 0
    fi
    
    # Проверка root прав
    check_root
    
    # Проверка требований
    check_requirements
    
    # Определение режима развертывания
    mode=${1:-"--full"}
    
    case $mode in
        "--full")
            log_info "Запуск полного развертывания SIEM..."
            install_base_packages
            setup_security
            setup_docker
            create_siem_user
            install_siem
            setup_environment
            generate_ssl_certs
            setup_siem_config
            setup_nginx
            setup_monitoring
            create_systemd_service
            create_management_scripts
            setup_audit
            start_siem
            test_deployment
            ;;
        "--minimal")
            log_info "Запуск минимального развертывания SIEM..."
            install_base_packages
            setup_security
            setup_docker
            create_siem_user
            install_siem
            setup_environment
            setup_siem_config
            create_systemd_service
            start_siem
            test_deployment
            ;;
        "--test")
            log_info "Запуск тестирования..."
            test_deployment
            ;;
        *)
            log_error "Неизвестная опция: $mode"
            echo "Используйте --help для получения справки"
            exit 1
            ;;
    esac
    
    echo ""
    echo "=========================================="
    log_success "Развертывание SIEM завершено успешно!"
    echo "=========================================="
    echo ""
    echo "Доступные сервисы:"
    echo "• SIEM Web UI: http://localhost:8081"
    echo "• Kibana: http://localhost:5601"
    echo "• Grafana: http://localhost:3000"
    echo "• Prometheus: http://localhost:9090"
    echo "• Elasticsearch: http://localhost:9200"
    echo ""
    echo "Управление сервисом:"
    echo "• sudo systemctl start siem"
    echo "• sudo systemctl stop siem"
    echo "• sudo systemctl status siem"
    echo "• sudo systemctl restart siem"
    echo ""
    echo "Мониторинг:"
    echo "• /opt/siem/scripts/monitor_siem.sh"
    echo "• /opt/siem/scripts/backup_siem.sh"
    echo "• /opt/siem/scripts/update_siem.sh"
    echo ""
    echo "Логи:"
    echo "• sudo journalctl -u siem.service -f"
    echo "• tail -f /opt/siem/logs/*.log"
    echo ""
    echo "⚠️  ВАЖНО: Не забудьте настроить переменные окружения в /opt/siem/.env.siem"
    echo "⚠️  ВАЖНО: Обновите IP адреса удаленных серверов в конфигурации"
    echo "⚠️  ВАЖНО: Настройте SSL сертификаты для production"
    echo ""
}

# Запуск скрипта
main "$@" 