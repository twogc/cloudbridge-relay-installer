# Быстрое развертывание SIEM на отдельном сервере

## 🚀 Быстрый старт

### 1. Подготовка сервера

#### Минимальные требования:
- **CPU**: 4 vCPU
- **RAM**: 16 GB
- **Storage**: 500 GB SSD
- **OS**: Ubuntu 20.04+ или 22.04+

#### Рекомендуемые требования:
- **CPU**: 8 vCPU
- **RAM**: 32 GB
- **Storage**: 2 TB NVMe SSD
- **Network**: 10 Gbps

### 2. Автоматическое развертывание

```bash
# Скачивание скрипта
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/deploy_siem_server.sh

# Установка прав на выполнение
chmod +x deploy_siem_server.sh

# Запуск полного развертывания
sudo ./deploy_siem_server.sh --full

# Или минимальное развертывание
sudo ./deploy_siem_server.sh --minimal
```

### 3. Ручное развертывание (пошагово)

#### Шаг 1: Подготовка системы
```bash
# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y docker.io docker-compose nginx git curl wget
```

#### Шаг 2: Установка SIEM
```bash
# Создание рабочей директории
sudo mkdir -p /opt/siem
sudo chown $USER:$USER /opt/siem
cd /opt/siem

# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git .
```

#### Шаг 3: Настройка конфигурации
```bash
# Создание .env файла
cat > .env.siem << 'EOF'
ELASTICSEARCH_PASSWORD=your_secure_password
SIEM_ENVIRONMENT=production
SIEM_HOST=$(hostname -I | awk '{print $1}')
EOF

# Запуск SIEM
cd siem
docker-compose -f docker-compose.siem.yml up -d
```

#### Шаг 4: Проверка работы
```bash
# Проверка сервисов
docker ps

# Проверка доступности
curl http://localhost:8081  # SIEM Web UI
curl http://localhost:5601  # Kibana
curl http://localhost:9200  # Elasticsearch
```

## 🔧 Настройка интеграции

### 1. Настройка сбора данных с CloudBridge Relay

```bash
# Создание SSH ключей для безопасного доступа
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""

# Копирование ключа на сервер Relay
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_relay_server_ip
```

### 2. Настройка конфигурации удаленного сбора

```yaml
# /opt/siem/siem/config/remote_collectors.yaml
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
```

### 3. Настройка правил детекции

```yaml
# /opt/siem/siem/config/detection_rules.yaml
detection_rules:
  brute_force:
    enabled: true
    conditions:
      event_type: "authentication_failure"
      time_window: 300
      threshold: 5
    actions:
      - type: "alert"
        severity: "high"
      - type: "block_ip"
        duration: 3600
```

## 📊 Мониторинг и алерты

### 1. Настройка Prometheus и Grafana

```bash
# Запуск мониторинга
docker run -d --name prometheus -p 9090:9090 prom/prometheus:latest
docker run -d --name grafana -p 3000:3000 -e GF_SECURITY_ADMIN_PASSWORD=admin grafana/grafana:latest
```

### 2. Настройка алертов

```yaml
# /opt/prometheus/siem_alerts.yml
groups:
  - name: siem_alerts
    rules:
      - alert: SIEMHighCPU
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
```

## 🔒 Безопасность

### 1. Настройка firewall

```bash
# Настройка UFW
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 5601/tcp  # Kibana
sudo ufw allow 9200/tcp  # Elasticsearch
sudo ufw enable
```

### 2. Настройка SSL сертификатов

```bash
# Генерация самоподписанных сертификатов
sudo mkdir -p /etc/ssl/siem
cd /etc/ssl/siem

sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout siem.key -out siem.crt \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=YourCompany/OU=IT/CN=siem.yourcompany.com"

sudo chmod 600 siem.key
sudo chmod 644 siem.crt
```

### 3. Настройка Nginx как reverse proxy

```nginx
# /etc/nginx/sites-available/siem
server {
    listen 443 ssl http2;
    server_name siem.yourcompany.com;

    ssl_certificate /etc/ssl/siem/siem.crt;
    ssl_certificate_key /etc/ssl/siem/siem.key;
    
    location / {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🛠️ Управление

### Основные команды

```bash
# Управление сервисом
sudo systemctl start siem
sudo systemctl stop siem
sudo systemctl restart siem
sudo systemctl status siem

# Просмотр логов
sudo journalctl -u siem.service -f
tail -f /opt/siem/logs/*.log

# Мониторинг ресурсов
/opt/siem/scripts/monitor_siem.sh

# Резервное копирование
/opt/siem/scripts/backup_siem.sh

# Обновление
/opt/siem/scripts/update_siem.sh
```

### Проверка работоспособности

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

## 📋 Чек-лист развертывания

### ✅ Подготовка
- [ ] Сервер соответствует требованиям
- [ ] Установлены базовые пакеты
- [ ] Настроена безопасность (firewall, fail2ban)
- [ ] Настроен Docker

### ✅ Установка SIEM
- [ ] Клонирован репозиторий
- [ ] Настроены переменные окружения
- [ ] Сгенерированы SSL сертификаты
- [ ] Запущены сервисы

### ✅ Интеграция
- [ ] Настроен сбор данных с Relay сервера
- [ ] Настроены правила детекции
- [ ] Настроены алерты
- [ ] Настроен мониторинг

### ✅ Безопасность
- [ ] Настроен firewall
- [ ] Настроены SSL сертификаты
- [ ] Настроен Nginx
- [ ] Настроен аудит

### ✅ Тестирование
- [ ] Проверена доступность всех сервисов
- [ ] Протестирована отправка событий
- [ ] Проверены алерты
- [ ] Проверен мониторинг

## 🚨 Устранение неполадок

### Проблемы с запуском

```bash
# Проверка статуса контейнеров
docker ps -a

# Просмотр логов контейнеров
docker-compose -f /opt/siem/siem/docker-compose.siem.yml logs

# Перезапуск сервисов
sudo systemctl restart siem
```

### Проблемы с производительностью

```bash
# Проверка использования ресурсов
htop
df -h
free -h

# Оптимизация Elasticsearch
curl -X PUT "localhost:9200/_cluster/settings" \
    -H "Content-Type: application/json" \
    -d '{
        "persistent": {
            "cluster.routing.allocation.disk.threshold_enabled": false
        }
    }'
```

### Проблемы с сетью

```bash
# Проверка портов
netstat -tlnp | grep -E ':(8081|5601|9200|9090|3000)'

# Проверка firewall
sudo ufw status

# Проверка DNS
nslookup siem.yourcompany.com
```

## 📞 Поддержка

### Полезные ссылки:
- [Полное руководство по развертыванию](siem_separate_server_deployment.md)
- [Руководство по интеграции](siem_integration_guide.md)
- [Анализ SIEM системы](siem_integration_analysis.md)

### Логи для диагностики:
- `/opt/siem/logs/` - логи SIEM
- `/var/log/nginx/` - логи Nginx
- `/var/log/docker/` - логи Docker
- `journalctl -u siem.service` - логи systemd сервиса

### Контакты:
- Создать issue в GitHub репозитории
- Обратиться к команде разработки
- Проверить документацию в папке `docs/` 