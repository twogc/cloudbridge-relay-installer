# Развертывание SIEM на отдельном сервере

## 🎯 Обоснование отдельного сервера

### Почему SIEM нужен отдельный сервер

#### 1. **Ресурсные требования**
- **Высокое потребление RAM**: Elasticsearch требует 8-16GB RAM
- **Интенсивные I/O операции**: Постоянная запись и чтение данных
- **CPU нагрузки**: Обработка событий в реальном времени
- **Дисковое пространство**: Хранение больших объемов логов (100GB+)

#### 2. **Безопасность**
- **Изоляция данных**: Разделение production и security инфраструктуры
- **Контроль доступа**: Ограничение доступа к SIEM только для security team
- **Аудит**: Независимое логирование всех операций
- **Compliance**: Соответствие требованиям регуляторов

#### 3. **Производительность**
- **Отсутствие конкуренции ресурсов** с основными сервисами
- **Оптимизация под SIEM нагрузки**
- **Масштабируемость**: Возможность независимого масштабирования

#### 4. **Надежность**
- **Отказоустойчивость**: Сбой SIEM не влияет на основную инфраструктуру
- **Резервное копирование**: Независимые стратегии backup
- **Восстановление**: Изолированные процедуры восстановления

## 🖥️ Требования к серверу SIEM

### Минимальные требования
```yaml
hardware:
  cpu: "4 vCPU (Intel Xeon или AMD EPYC)"
  ram: "16 GB DDR4"
  storage: "500 GB SSD (NVMe предпочтительно)"
  network: "1 Gbps"
  
software:
  os: "Ubuntu 20.04 LTS или 22.04 LTS"
  docker: "20.10+"
  docker_compose: "2.0+"
  python: "3.8+"
```

### Рекомендуемые требования
```yaml
hardware:
  cpu: "8 vCPU (Intel Xeon или AMD EPYC)"
  ram: "32 GB DDR4"
  storage: "2 TB NVMe SSD"
  network: "10 Gbps"
  
software:
  os: "Ubuntu 22.04 LTS"
  docker: "24.0+"
  docker_compose: "2.20+"
  python: "3.11+"
```

### Production требования
```yaml
hardware:
  cpu: "16 vCPU (Intel Xeon или AMD EPYC)"
  ram: "64 GB DDR4 ECC"
  storage: "4 TB NVMe SSD (RAID 1)"
  network: "25 Gbps"
  
software:
  os: "Ubuntu 22.04 LTS"
  docker: "24.0+"
  docker_compose: "2.20+"
  python: "3.11+"
  monitoring: "Prometheus + Grafana"
```

## 🚀 Пошаговое развертывание

### Этап 1: Подготовка сервера

#### 1.1. Установка базовой системы
```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y \
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
    python3-certbot-nginx
```

#### 1.2. Настройка безопасности
```bash
# Настройка firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 5601/tcp  # Kibana
sudo ufw allow 9200/tcp  # Elasticsearch
sudo ufw allow 9092/tcp  # Kafka
sudo ufw allow 8081/tcp  # SIEM Web UI
sudo ufw enable

# Настройка fail2ban
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

#### 1.3. Настройка Docker
```bash
# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Настройка Docker daemon
sudo tee /etc/docker/daemon.json << EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m",
    "max-file": "3"
  },
  "storage-driver": "overlay2",
  "storage-opts": [
    "overlay2.override_kernel_check=true"
  ]
}
EOF

sudo systemctl restart docker
```

### Этап 2: Установка SIEM

#### 2.1. Клонирование репозитория
```bash
# Создание рабочей директории
sudo mkdir -p /opt/siem
sudo chown $USER:$USER /opt/siem
cd /opt/siem

# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git .
```

#### 2.2. Настройка переменных окружения
```bash
# Создание .env файла
cat > .env.siem << 'EOF'
# Elasticsearch
ELASTICSEARCH_PASSWORD=changeme_secure_password_here
ELASTICSEARCH_USERNAME=elastic

# SIEM Configuration
SIEM_ENVIRONMENT=production
SIEM_LOG_LEVEL=INFO

# Telegram Notifications
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here

# Email Notifications
EMAIL_SMTP_SERVER=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_USERNAME=your_email@gmail.com
EMAIL_PASSWORD=your_app_password_here
EMAIL_FROM=siem@yourcompany.com
EMAIL_TO=security@yourcompany.com,admin@yourcompany.com

# Network Configuration
SIEM_HOST=your_siem_server_ip
RELAY_HOST=your_relay_server_ip
KEYCLOAK_HOST=your_keycloak_server_ip
DJANGO_HOST=your_django_server_ip

# Security
SSL_CERT_PATH=/etc/ssl/certs/siem.crt
SSL_KEY_PATH=/etc/ssl/private/siem.key
EOF

# Загрузка переменных
source .env.siem
```

#### 2.3. Генерация SSL сертификатов
```bash
# Создание самоподписанных сертификатов (для тестирования)
sudo mkdir -p /etc/ssl/siem
cd /etc/ssl/siem

sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout siem.key -out siem.crt \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=YourCompany/OU=IT/CN=siem.yourcompany.com"

sudo chmod 600 siem.key
sudo chmod 644 siem.crt

# Для production используйте Let's Encrypt
# sudo certbot --nginx -d siem.yourcompany.com
```

### Этап 3: Конфигурация SIEM

#### 3.1. Настройка multi-tenant конфигурации
```bash
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
```

#### 3.2. Настройка правил детекции
```bash
# Создание правил детекции
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
```

### Этап 4: Запуск SIEM

#### 4.1. Запуск сервисов
```bash
cd /opt/siem/siem

# Запуск всех сервисов
docker-compose -f docker-compose.siem.yml up -d

# Проверка статуса
docker-compose -f docker-compose.siem.yml ps

# Просмотр логов
docker-compose -f docker-compose.siem.yml logs -f
```

#### 4.2. Настройка systemd сервиса
```bash
# Создание systemd сервиса
sudo tee /etc/systemd/system/siem.service << 'EOF'
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

[Install]
WantedBy=multi-user.target
EOF

# Активация сервиса
sudo systemctl daemon-reload
sudo systemctl enable siem.service
sudo systemctl start siem.service
```

### Этап 5: Настройка мониторинга

#### 5.1. Настройка Nginx как reverse proxy
```bash
# Создание конфигурации Nginx
sudo tee /etc/nginx/sites-available/siem << 'EOF'
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
sudo ln -s /etc/nginx/sites-available/siem /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

#### 5.2. Настройка мониторинга сервера
```bash
# Установка Prometheus Node Exporter
sudo docker run -d \
    --name prometheus-node-exporter \
    --restart unless-stopped \
    -p 9100:9100 \
    -v "/:/host:ro,rslave" \
    prom/node-exporter:latest \
    --path.rootfs=/host

# Создание конфигурации Prometheus
sudo mkdir -p /opt/prometheus
sudo tee /opt/prometheus/prometheus.yml << 'EOF'
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

# Запуск Prometheus
sudo docker run -d \
    --name prometheus \
    --restart unless-stopped \
    -p 9090:9090 \
    -v /opt/prometheus:/etc/prometheus \
    prom/prometheus:latest \
    --config.file=/etc/prometheus/prometheus.yml

# Запуск Grafana
sudo docker run -d \
    --name grafana \
    --restart unless-stopped \
    -p 3000:3000 \
    -e GF_SECURITY_ADMIN_PASSWORD=admin \
    grafana/grafana:latest
```

### Этап 6: Интеграция с CloudBridge Relay

#### 6.1. Настройка сбора данных с удаленных серверов
```bash
# Создание конфигурации для удаленного сбора
cat > siem/config/remote_collectors.yaml << 'EOF'
remote_collectors:
  relay_server:
    host: "your_relay_server_ip"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "/var/log/relay/relay.log"
        type: "relay_logs"
        interval: 60
      - path: "http://localhost:8081/metrics"
        type: "relay_metrics"
        interval: 30

  keycloak_server:
    host: "your_keycloak_server_ip"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "http://localhost:8080/admin/events"
        type: "keycloak_events"
        interval: 120

  django_server:
    host: "your_django_server_ip"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "http://localhost:8000/api/audit/"
        type: "django_audit"
        interval: 60
EOF
```

#### 6.2. Настройка SSH ключей для безопасного доступа
```bash
# Генерация SSH ключей
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""

# Копирование публичного ключа на удаленные серверы
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_relay_server_ip
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_keycloak_server_ip
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_django_server_ip

# Настройка прав доступа
chmod 600 /opt/siem/.ssh/id_rsa
chmod 644 /opt/siem/.ssh/id_rsa.pub
```

### Этап 7: Тестирование и валидация

#### 7.1. Проверка доступности сервисов
```bash
# Проверка всех сервисов
services=(
    "http://localhost:8081"    # SIEM Web UI
    "http://localhost:5601"    # Kibana
    "http://localhost:3000"    # Grafana
    "http://localhost:9090"    # Prometheus
    "http://localhost:9200"    # Elasticsearch
)

for service in "${services[@]}"; do
    if curl -s "$service" > /dev/null 2>&1; then
        echo "✅ $service доступен"
    else
        echo "❌ $service недоступен"
    fi
done
```

#### 7.2. Тест отправки событий
```bash
# Тест отправки события в SIEM
test_event='{
  "tenant_id": "test_tenant",
  "event_type": "test_event",
  "severity": "medium",
  "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "source_ip": "192.168.1.100",
  "user_id": "test_user",
  "action": "test_action"
}'

curl -X POST "http://localhost:8081/api/v1/events" \
    -H "Content-Type: application/json" \
    -d "$test_event"
```

## 🔧 Управление и обслуживание

### Мониторинг производительности
```bash
# Создание скрипта мониторинга
cat > /opt/siem/scripts/monitor_siem.sh << 'EOF'
#!/bin/bash

# Проверка использования ресурсов
echo "=== Использование ресурсов ==="
echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)%"
echo "RAM: $(free | grep Mem | awk '{printf "%.1f%%", $3/$2 * 100.0}')"
echo "Disk: $(df -h / | awk 'NR==2 {print $5}')"

# Проверка Docker контейнеров
echo -e "\n=== Статус контейнеров ==="
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Проверка логов
echo -e "\n=== Последние ошибки ==="
docker-compose -f /opt/siem/siem/docker-compose.siem.yml logs --tail=10 | grep -i error
EOF

chmod +x /opt/siem/scripts/monitor_siem.sh
```

### Резервное копирование
```bash
# Создание скрипта резервного копирования
cat > /opt/siem/scripts/backup_siem.sh << 'EOF'
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

chmod +x /opt/siem/scripts/backup_siem.sh

# Добавление в cron
echo "0 2 * * * /opt/siem/scripts/backup_siem.sh" | sudo crontab -
```

## 🔒 Безопасность

### Дополнительные меры безопасности
```bash
# Настройка auditd
sudo apt install auditd
sudo systemctl enable auditd
sudo systemctl start auditd

# Настройка аудита для SIEM директорий
sudo tee /etc/audit/rules.d/siem.rules << 'EOF'
-w /opt/siem -p wa -k siem_config
-w /opt/siem/siem/config -p wa -k siem_config
-w /opt/siem/.env.siem -p wa -k siem_secrets
EOF

sudo systemctl restart auditd

# Настройка logrotate для SIEM логов
sudo tee /etc/logrotate.d/siem << 'EOF'
/opt/siem/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 ubuntu ubuntu
    postrotate
        systemctl reload siem
    endscript
}
EOF
```

## 📊 Мониторинг и алерты

### Настройка алертов
```bash
# Создание правил алертов Prometheus
sudo tee /opt/prometheus/siem_alerts.yml << 'EOF'
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
```

## 🔗 Заключение

Развертывание SIEM на отдельном сервере обеспечивает:

1. **Оптимальную производительность** - выделенные ресурсы для SIEM
2. **Повышенную безопасность** - изоляция security инфраструктуры
3. **Масштабируемость** - независимое масштабирование SIEM
4. **Надежность** - отказоустойчивость и независимое восстановление
5. **Соответствие стандартам** - правильная архитектура для compliance

### Следующие шаги:
1. Настройка интеграции с внешними системами (SOAR, DLP)
2. Настройка advanced threat hunting
3. Интеграция с threat intelligence feeds
4. Настройка automated response actions
5. Проведение penetration testing
6. Настройка disaster recovery procedures 