# 🚀 Полное руководство по развертыванию SIEM на отдельном сервере

## 📋 Обзор созданных материалов

Вы абсолютно правы - **SIEM систему необходимо развертывать на отдельном сервере**. Мы создали полный набор документов и инструментов для этого.

### 🎯 Ключевые выводы

**Почему нужен отдельный сервер:**
- **Ресурсные требования**: SIEM требует 16-64GB RAM, интенсивные I/O операции
- **Безопасность**: Изоляция security инфраструктуры от production систем  
- **Производительность**: Отсутствие конкуренции ресурсов с основными сервисами
- **Масштабируемость**: Независимое масштабирование SIEM и Relay систем
- **Надежность**: Отказоустойчивость и независимое восстановление

## 📚 Созданная документация

### 1. **Анализ и планирование**
- [📊 **Анализ SIEM системы**](siem_integration_analysis.md) - Комплексный анализ текущей SIEM системы, архитектуры и рекомендаций
- [📋 **Резюме интеграции**](siem_integration_summary.md) - Ключевые выводы и план внедрения

### 2. **Руководства по развертыванию**
- [🖥️ **Развертывание на отдельном сервере**](siem_separate_server_deployment.md) - Подробное пошаговое руководство
- [⚡ **Быстрое развертывание**](siem_quick_deployment.md) - Краткое руководство для быстрого старта
- [🔗 **Руководство по интеграции**](siem_integration_guide.md) - Интеграция с CloudBridge Relay

### 3. **Миграция и управление**
- [🔄 **Руководство по миграции**](siem_migration_guide.md) - Миграция с текущего сервера на отдельный

## 🔧 Автоматизированные инструменты

### 1. **Скрипт развертывания SIEM сервера**
```bash
# scripts/deploy_siem_server.sh
# Полное автоматическое развертывание SIEM на отдельном сервере

# Режимы работы:
sudo ./deploy_siem_server.sh --full      # Полное развертывание
sudo ./deploy_siem_server.sh --minimal   # Минимальное развертывание  
sudo ./deploy_siem_server.sh --test      # Только тестирование
```

**Возможности скрипта:**
- ✅ Автоматическая установка всех зависимостей
- ✅ Настройка безопасности (firewall, fail2ban, auditd)
- ✅ Конфигурация Docker и контейнеров
- ✅ Настройка Nginx как reverse proxy
- ✅ Установка мониторинга (Prometheus, Grafana)
- ✅ Создание systemd сервисов
- ✅ Настройка backup и logrotate
- ✅ Автоматическое тестирование

### 2. **Скрипт интеграции**
```bash
# scripts/integrate_siem.sh
# Интеграция SIEM с существующей инфраструктурой

# Режимы работы:
sudo ./integrate_siem.sh --mode=production  # Production интеграция
sudo ./integrate_siem.sh --mode=test        # Тестовая интеграция
```

## 🏗️ Архитектурное решение

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Relay Server  │    │  Keycloak Server│    │  Django Server  │
│                 │    │                 │    │                 │
│ - CloudBridge   │    │ - Authentication│    │ - Management    │
│ - TCP Proxy     │    │ - User Mgmt     │    │ - API           │
│ - Multi-tenant  │    │ - SSO           │    │ - Dashboard     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   SIEM Server   │
                    │                 │
                    │ - Elasticsearch │
                    │ - Kibana        │
                    │ - Kafka         │
                    │ - Collectors    │
                    │ - Processors    │
                    │ - Alert Manager │
                    │ - Web UI        │
                    └─────────────────┘
```

## 🖥️ Требования к серверу

### Минимальные требования
```yaml
hardware:
  cpu: "4 vCPU"
  ram: "16 GB DDR4"
  storage: "500 GB SSD"
  network: "1 Gbps"
  
software:
  os: "Ubuntu 20.04+ или 22.04+"
  docker: "20.10+"
  docker_compose: "2.0+"
```

### Рекомендуемые требования
```yaml
hardware:
  cpu: "8 vCPU"
  ram: "32 GB DDR4"
  storage: "2 TB NVMe SSD"
  network: "10 Gbps"
```

### Production требования
```yaml
hardware:
  cpu: "16 vCPU"
  ram: "64 GB DDR4 ECC"
  storage: "4 TB NVMe SSD (RAID 1)"
  network: "25 Gbps"
```

## 🚀 Быстрый старт

### 1. **Автоматическое развертывание**
```bash
# Скачивание скрипта
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/deploy_siem_server.sh

# Установка прав на выполнение
chmod +x deploy_siem_server.sh

# Запуск полного развертывания
sudo ./deploy_siem_server.sh --full
```

### 2. **Ручное развертывание**
```bash
# Подготовка системы
sudo apt update && sudo apt upgrade -y
sudo apt install -y docker.io docker-compose nginx git curl wget

# Установка SIEM
sudo mkdir -p /opt/siem
sudo chown $USER:$USER /opt/siem
cd /opt/siem
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git .

# Запуск SIEM
cd siem
docker-compose -f docker-compose.siem.yml up -d
```

### 3. **Проверка работы**
```bash
# Проверка сервисов
docker ps

# Проверка доступности
curl http://localhost:8081  # SIEM Web UI
curl http://localhost:5601  # Kibana
curl http://localhost:9200  # Elasticsearch
```

## 🔧 Настройка интеграции

### 1. **Настройка сбора данных с CloudBridge Relay**
```bash
# Создание SSH ключей
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""

# Копирование ключа на сервер Relay
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@your_relay_server_ip
```

### 2. **Конфигурация удаленного сбора**
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

### 3. **Настройка правил детекции**
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

## 🔒 Безопасность

### 1. **Настройка firewall**
```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 5601/tcp  # Kibana
sudo ufw allow 9200/tcp  # Elasticsearch
sudo ufw enable
```

### 2. **SSL сертификаты**
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

### 3. **Nginx как reverse proxy**
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

## 📊 Мониторинг и алерты

### 1. **Prometheus и Grafana**
```bash
# Запуск мониторинга
docker run -d --name prometheus -p 9090:9090 prom/prometheus:latest
docker run -d --name grafana -p 3000:3000 -e GF_SECURITY_ADMIN_PASSWORD=admin grafana/grafana:latest
```

### 2. **Алерты**
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

## 🔄 Миграция с текущего сервера

### Автоматическая миграция
```bash
# Скачивание скрипта миграции
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/migrate_siem.sh

# Запуск миграции
chmod +x migrate_siem.sh
./migrate_siem.sh OLD_SERVER_IP NEW_SIEM_SERVER_IP
```

### Этапы миграции
1. **Подготовка** (1-2 дня) - Анализ, подготовка нового сервера
2. **Перенос данных** (1 день) - Backup и восстановление
3. **Обновление конфигурации** (2-3 часа) - Настройка интеграции
4. **Тестирование** (2-3 часа) - Проверка функциональности
5. **Переключение** (1 час) - Остановка старого, запуск нового
6. **Очистка** (1 час) - Удаление старой системы

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

## 🚨 Troubleshooting

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

## 📊 Метрики успеха

### Технические метрики
- **Время развертывания**: < 4 часов
- **Время восстановления**: < 30 минут
- **Производительность**: Обработка > 10,000 событий/сек
- **Доступность**: 99.9% uptime

### Бизнес метрики
- **Время обнаружения угроз**: < 5 минут
- **Время реагирования**: < 15 минут
- **Coverage**: 100% критических систем
- **Compliance**: 100% соответствие требованиям

### Операционные метрики
- **False positive rate**: < 5%
- **Alert accuracy**: > 95%
- **Mean time to resolution**: < 2 часа
- **User satisfaction**: > 90%

## 🎯 Следующие шаги

### 1. **Немедленные действия**
1. **Выделить отдельный сервер** для SIEM системы
2. **Провести планирование** сетевой архитектуры
3. **Подготовить SSL сертификаты** для production
4. **Создать backup стратегию** для SIEM данных

### 2. **Краткосрочные задачи (1-2 недели)**
1. **Развернуть SIEM** на отдельном сервере
2. **Настроить интеграцию** с CloudBridge Relay
3. **Протестировать** базовую функциональность
4. **Настроить мониторинг** и алерты

### 3. **Среднесрочные задачи (1-2 месяца)**
1. **Оптимизировать производительность** SIEM
2. **Настроить advanced threat hunting**
3. **Интегрировать** с внешними threat intelligence feeds
4. **Провести security audit** и penetration testing

### 4. **Долгосрочные задачи (3-6 месяцев)**
1. **Масштабировать** SIEM для роста нагрузки
2. **Интегрировать** с SOAR платформами
3. **Настроить automated response** actions
4. **Провести compliance audit** и сертификацию

## 💡 Заключение

**Вы абсолютно правы** - SIEM систему необходимо развертывать на отдельном сервере. Это обеспечивает:

- ✅ **Оптимальную производительность** - выделенные ресурсы для SIEM
- ✅ **Повышенную безопасность** - изоляция security инфраструктуры
- ✅ **Независимое масштабирование** - SIEM и Relay системы масштабируются отдельно
- ✅ **Повышенную надежность** - отказоустойчивость и независимое восстановление
- ✅ **Соответствие стандартам** - правильная архитектура для compliance

**Готовые решения:**
- 📚 **6 подробных руководств** по развертыванию и интеграции
- 🔧 **2 автоматизированных скрипта** для быстрого развертывания
- 🎯 **Пошаговые инструкции** для всех этапов
- 📊 **Готовые конфигурации** для production

**Рекомендуемый подход:**
1. **Начать с быстрого развертывания** на отдельном сервере
2. **Постепенно интегрировать** с существующей инфраструктурой
3. **Оптимизировать** под конкретные требования
4. **Масштабировать** по мере роста нагрузки

Это обеспечит надежную, безопасную и масштабируемую SIEM систему, полностью интегрированную с CloudBridge Relay v2.0. 