# Руководство по развертыванию CloudBridge Relay v2.0

## Содержание

1. [Требования](#требования)
2. [Подготовка окружения](#подготовка-окружения)
3. [Быстрая установка (рекомендуется)](#быстрая-установка-рекомендуется)
4. [Установка](#установка)
5. [Конфигурация](#конфигурация)
6. [Развертывание](#развертывание)
7. [Автоматизация](#автоматизация)
8. [Мониторинг](#мониторинг)
9. [Обслуживание](#обслуживание)
10. [Устранение неполадок](#устранение-неполадок)
11. [Новые функции v2.0](#новые-функции-v20)

## Требования

### Системные требования

- **ОС:** Linux (Ubuntu 20.04+, CentOS 8+, RHEL 8+)
- **CPU:** 2+ ядра (рекомендуется 4+ для multi-tenancy)
- **RAM:** 4+ GB (рекомендуется 8+ GB для TCP proxy)
- **Диск:** 20+ GB свободного места
- **Сеть:** статический IP-адрес

### Программные требования

- **Go:** 1.21+ (обязательно)
- **Docker:** 20.10+ (опционально)
- **Docker Compose:** 2.0+ (опционально)
- **Systemd:** для установки как сервис
- **Certbot:** для SSL сертификатов

### Сетевые требования

- **Порт 8080:** основной порт для клиентов
- **Порт 8081:** порт для метрик и API
- **Порт 80:** для получения SSL сертификатов
- **Порт 443:** для HTTPS (опционально)
- **Диапазон портов 10000-20000:** для TCP proxy туннелей

## Подготовка окружения

### 1. Обновление системы

```bash
# Ubuntu/Debian
sudo apt update && sudo apt upgrade -y

# CentOS/RHEL
sudo yum update -y
```

### 2. Установка зависимостей

```bash
# Ubuntu/Debian
sudo apt install -y curl wget git build-essential

# CentOS/RHEL
sudo yum install -y curl wget git gcc
```

### 3. Установка Go 1.21+

```bash
# Скачивание Go 1.21.13
wget https://go.dev/dl/go1.21.13.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.13.linux-amd64.tar.gz

# Добавление в PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Проверка установки
go version
# Должно показать: go version go1.21.13 linux/amd64
```

### 4. Установка Docker (опционально)

```bash
# Установка Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Добавление пользователя в группу docker
sudo usermod -aG docker $USER

# Установка Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

## Быстрая установка (рекомендуется)

### Автоматическая установка с Docker Compose

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Генерация production секретов
./scripts/generate_secrets.sh --env .env.production

# Запуск всех сервисов в production режиме
./scripts/start_production.sh --env .env.production

# Проверка статуса
docker-compose ps
```

### Установка как системный сервис (автозапуск)

```bash
# Установка systemd сервиса
sudo ./scripts/install_systemd.sh install

# Проверка статуса
sudo systemctl status cloudbridge-relay

# Управление сервисом
sudo systemctl start cloudbridge-relay
sudo systemctl stop cloudbridge-relay
sudo systemctl restart cloudbridge-relay

# Просмотр логов
sudo journalctl -u cloudbridge-relay -f
```

## Установка

### Способ 1: Установка из исходного кода

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Обновление зависимостей
go mod tidy

# Сборка
go build -o relay cmd/relay/main.go

# Установка
sudo cp relay /usr/local/bin/
sudo chmod +x /usr/local/bin/relay
```

### Способ 2: Установка через Docker

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

# Сборка образа
docker build -t cloudbridge-relay:v2.0 .

# Запуск контейнера
docker run -d \
  --name relay \
  -p 8080:8080 \
  -p 8081:8081 \
  -p 10000-20000:10000-20000 \
  -v /etc/relay:/etc/relay \
  -v /var/log/relay:/var/log/relay \
  -e RELAY_JWT_SECRET=your-secret-key \
  cloudbridge-relay:v2.0 ./relay --config config/config.json
```

### Способ 3: Установка как системный сервис

```bash
# Установка systemd сервиса
sudo ./scripts/utils/install_systemd.sh install

# Проверка статуса
sudo systemctl status cloudbridge-relay

# Управление сервисом
sudo systemctl start cloudbridge-relay
sudo systemctl stop cloudbridge-relay
sudo systemctl restart cloudbridge-relay
```

## Конфигурация

### 1. Создание конфигурационного файла

```bash
# Создание директории конфигурации
sudo mkdir -p /etc/relay

# Копирование конфигурации
sudo cp config/config.json /etc/relay/
```

### 2. Настройка переменных окружения

```bash
# Создание файла с переменными окружения
sudo tee /etc/relay/relay.env << EOF
RELAY_JWT_SECRET=your-secure-jwt-secret
RELAY_DJANGO_API_ENDPOINT=http://your-django-server:8000/api
RELAY_LOG_LEVEL=info
RELAY_TLS_ENABLED=true
RELAY_TENANT_ENABLED=true
RELAY_PROXY_ENABLED=true
EOF
```

### 3. Формат конфигурационного файла v2.0

CloudBridge Relay v2.0 поддерживает расширенную конфигурацию с новыми функциями:

```json
{
    "server": {
        "address": ":8080",
        "port": 8080,
        "timeout": "30s",
        "max_tunnels": 1000,
        "ping_interval": "30s",
        "cleanup_interval": "1m"
    },
    "django": {
        "api_endpoint": "http://localhost:8000/api",
        "jwt_secret": "REPLACE_WITH_SECURE_SECRET_FROM_ENVIRONMENT",
        "jwt_expiry": "24h",
        "rest_timeout": "5s",
        "rest_cache_ttl": "5m",
        "rest_max_retries": 3,
        "rest_retry_delay": "100ms"
    },
    "keycloak": {
        "enabled": true,
        "server_url": "https://keycloak.example.com",
        "realm": "cloudbridge",
        "client_id": "relay-client",
        "client_secret": "REPLACE_WITH_KEYCLOAK_CLIENT_SECRET",
        "public_key": "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...",
        "token_validation": {
            "issuer": "https://keycloak.example.com/realms/cloudbridge",
            "audience": "relay-client",
            "clock_skew": "30s"
        },
        "jwks_cache_ttl": "1h",
        "http_timeout": "30s",
        "fallback_to_django": true
    },
    "tenant": {
        "enabled": true,
        "default_limits": {
            "max_tunnels": 10,
            "max_connections_per_tunnel": 100,
            "max_bandwidth_mbps": 100,
            "max_data_transfer_gb": 1000,
            "rate_limit_per_second": 1000
        },
        "cleanup_interval": "5m",
        "inactivity_timeout": "1h"
    },
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000,
        "port_range": {
            "start": 10000,
            "end": 20000
        },
        "metrics_enabled": true
    },
    "tunnel": {
        "manager_enabled": true,
        "max_tunnels_per_tenant": 50,
        "cleanup_interval": "2m",
        "heartbeat_timeout": "5m"
    },
    "rate_limiting": {
        "enabled": true,
        "requests_per_second": 100,
        "burst": 200,
        "per_tenant": true
    },
    "logging": {
        "level": "info",
        "format": "json",
        "output": "/var/log/relay/relay.log",
        "max_size": "100MB",
        "max_age": "30d",
        "max_backups": 10
    },
    "metrics": {
        "enabled": true,
        "address": ":8081",
        "path": "/metrics",
        "collect_interval": "15s"
    },
    "tls": {
        "enabled": true,
        "cert_file": "/etc/relay/certs/server.crt",
        "key_file": "/etc/relay/certs/server.key",
        "ca_file": "/etc/relay/certs/ca.crt",
        "min_version": "1.3",
        "cipher_suites": [
            "TLS_AES_256_GCM_SHA384",
            "TLS_CHACHA20_POLY1305_SHA256",
            "TLS_AES_128_GCM_SHA256"
        ]
    }
}
```

### 4. Настройка multi-tenancy

```bash
# Создание файла конфигурации тенантов
sudo tee /etc/relay/tenants.yaml << EOF
tenants:
  tenant_001:
    name: "Production Tenant"
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
      max_data_transfer_gb: 5000
      rate_limit_per_second: 2000
    allowed_ips:
      - "192.168.1.0/24"
    metadata:
      contact: "admin@company.com"
      environment: "production"
  
  tenant_002:
    name: "Development Tenant"
    limits:
      max_tunnels: 5
      max_connections_per_tunnel: 50
      max_bandwidth_mbps: 100
      max_data_transfer_gb: 1000
      rate_limit_per_second: 500
    allowed_ips:
      - "192.168.2.0/24"
    metadata:
      contact: "dev@company.com"
      environment: "development"
EOF
```

## Развертывание

### 1. Запуск сервера

```bash
# Запуск с конфигурацией
sudo /usr/local/bin/relay --config /etc/relay/config.json

# Или через systemd
sudo systemctl start relay
sudo systemctl enable relay
```

## Автоматизация

### 1. Скрипты автоматизации

CloudBridge Relay v2.0 включает набор скриптов для автоматизации развертывания и управления:

#### Генерация секретов
```bash
# Генерация production секретов
./scripts/generate_secrets.sh --env .env.production

# Генерация development секретов
./scripts/generate_secrets.sh --env .env.development

# Генерация с кастомными параметрами
./scripts/generate_secrets.sh --env .env.custom --keycloak-admin admin --keycloak-password secure123
```

#### Запуск production стека
```bash
# Полный запуск всех сервисов
./scripts/start_production.sh --env .env.production

# Принудительный перезапуск
./scripts/start_production.sh --env .env.production --force

# Проверка статуса
./scripts/start_production.sh --status

# Остановка всех сервисов
./scripts/start_production.sh --stop
```

#### Установка systemd сервиса
```bash
# Установка автозапуска
sudo ./scripts/install_systemd.sh install

# Удаление сервиса
sudo ./scripts/install_systemd.sh uninstall

# Проверка статуса
sudo ./scripts/install_systemd.sh status

# Перезапуск сервиса
sudo ./scripts/install_systemd.sh restart
```

### 2. Systemd сервис

Автоматический запуск CloudBridge Relay при старте системы:

```bash
# Файл сервиса: /etc/systemd/system/cloudbridge-relay.service
[Unit]
Description=CloudBridge Relay Production Stack
Documentation=https://github.com/twogc/cloudbridge-relay-installer
After=docker.service
Requires=docker.service
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
User=ubuntu
Group=ubuntu
WorkingDirectory=/home/ubuntu/cloudbridge-relay-installer
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ExecStart=/home/ubuntu/cloudbridge-relay-installer/scripts/start_production.sh --env .env.production
ExecStop=/usr/bin/docker-compose down
ExecReload=/usr/bin/docker-compose restart
TimeoutStartSec=300
TimeoutStopSec=60
Restart=on-failure
RestartSec=30

[Install]
WantedBy=multi-user.target
```

### 3. Мониторинг и алерты

#### Health Checks
Все сервисы имеют настроенные health checks:

- **Keycloak**: проверка порта 8080
- **Relay**: проверка метрик на порту 8081
- **PostgreSQL**: встроенный health check
- **Grafana**: проверка API health
- **Prometheus**: встроенный health check
- **AlertManager**: встроенный health check

#### Автоматические алерты
```bash
# Настройка алертов Prometheus
sudo cp deploy/prometheus/alerts/relay_alerts.yml /etc/prometheus/alerts/

# Проверка алертов
curl http://localhost:9090/api/v1/alerts
```

### 4. Резервное копирование

#### Автоматические бэкапы
```bash
# Создание бэкапа Keycloak
./scripts/backup_keycloak.sh

# Создание бэкапа конфигурации
./scripts/backup_keycloak.sh --config-only

# Восстановление из бэкапа
./scripts/backup_keycloak.sh --restore backup_20240628.tar.gz
```

#### Ротация логов
```bash
# Настройка logrotate
sudo tee /etc/logrotate.d/cloudbridge-relay << EOF
/home/ubuntu/cloudbridge-relay-installer/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 ubuntu ubuntu
    postrotate
        docker-compose restart relay
    endscript
}
EOF
```

### 5. Обновления

#### Автоматическое обновление
```bash
# Обновление кода
git pull origin main

# Пересборка образов
docker-compose build --no-cache

# Перезапуск с новыми образами
./scripts/start_production.sh --env .env.production --force
```

#### Откат изменений
```bash
# Откат к предыдущей версии
git checkout v2.0.1

# Пересборка и перезапуск
docker-compose build --no-cache
./scripts/start_production.sh --env .env.production --force
```

### 6. Управление сертификатами

#### Автоматическое обновление SSL
```bash
# Настройка автоматического обновления
./scripts/renew_certs.sh --setup-cron

# Ручное обновление
./scripts/renew_certs.sh

# Проверка сертификатов
./scripts/renew_certs.sh --check
```

### 7. Интеграция с CI/CD

#### GitHub Actions
```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to server
        uses: appleboy/ssh-action@v0.1.5
        with:
          host: ${{ secrets.HOST }}
          username: ${{ secrets.USERNAME }}
          key: ${{ secrets.KEY }}
          script: |
            cd /home/ubuntu/cloudbridge-relay-installer
            git pull origin main
            ./scripts/start_production.sh --env .env.production --force
```

### 8. Мониторинг производительности

#### Метрики в реальном времени
```bash
# Просмотр метрик Relay
curl http://localhost:8081/metrics

# Просмотр метрик Keycloak
curl http://localhost:8080/health

# Просмотр метрик PostgreSQL
docker exec postgres pg_stat_statements
```

#### Дашборды Grafana
- **Relay Dashboard**: мониторинг туннелей и соединений
- **Keycloak Dashboard**: мониторинг аутентификации
- **System Dashboard**: мониторинг ресурсов системы

### 9. Устранение неполадок

#### Автоматическая диагностика
```bash
# Запуск диагностики
./scripts/test_integration.sh

# Проверка логов
./scripts/manage_logs.sh --tail

# Очистка старых логов
./scripts/manage_logs.sh --cleanup
```

#### Восстановление после сбоев
```bash
# Автоматическое восстановление
./scripts/start_production.sh --env .env.production --recover

# Принудительная очистка
docker-compose down --volumes --remove-orphans
./scripts/start_production.sh --env .env.production --force
```

## Мониторинг

### 1. Метрики Prometheus

Сервер экспортирует следующие метрики:

```bash
# Основные метрики
relay_connections_total
relay_tunnels_active
relay_bytes_transferred_total
relay_requests_total
relay_errors_total

# Метрики по тенантам
relay_tenant_tunnels_active
relay_tenant_connections_total
relay_tenant_bytes_transferred_total

# Метрики TCP proxy
relay_proxy_connections_active
relay_proxy_bytes_sent_total
relay_proxy_bytes_received_total
relay_proxy_errors_total
```

### 2. Настройка Grafana

```bash
# Создание дашборда
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @deploy/grafana/dashboards/relay_dashboard.json
```

### 3. Алерты

```bash
# Настройка алертов Prometheus
sudo cp deploy/prometheus/alerts/relay_alerts.yml /etc/prometheus/alerts/
```

## Обслуживание

### 1. Обновление сервера

```bash
# Остановка сервиса
sudo systemctl stop relay

# Создание бэкапа
sudo cp /usr/local/bin/relay /usr/local/bin/relay.backup

# Обновление
git pull
go build -o relay cmd/relay/main.go
sudo cp relay /usr/local/bin/

# Запуск сервиса
sudo systemctl start relay
```

### 2. Ротация логов

```bash
# Настройка logrotate
sudo tee /etc/logrotate.d/relay << EOF
/var/log/relay/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 relay relay
    postrotate
        systemctl reload relay
    endscript
}
EOF
```

### 3. Очистка неактивных туннелей

```bash
# Автоматическая очистка (настроена в конфигурации)
# Ручная очистка через API
curl -X POST http://localhost:8081/api/v1/tunnels/cleanup \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Устранение неполадок

### 1. Проверка логов

```bash
# Основные логи
sudo tail -f /var/log/relay/relay.log

# Systemd логи
sudo journalctl -u relay -f

# Логи TCP proxy
sudo tail -f /var/log/relay/proxy.log
```

### 2. Диагностика соединений

```bash
# Проверка открытых портов
sudo netstat -tlnp | grep relay

# Проверка TCP соединений
sudo ss -tlnp | grep relay

# Проверка метрик
curl http://localhost:8081/metrics | grep relay
```

### 3. Проблемы с multi-tenancy

```bash
# Проверка конфигурации тенантов
sudo relay --config /etc/relay/config.json --validate-tenants

# Просмотр активных тенантов
curl http://localhost:8081/api/v1/tenants \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Новые функции v2.0

### 1. TCP Proxy

**Описание:** Полнофункциональный TCP прокси для двунаправленной передачи данных между клиентами и целевыми сервисами.

**Особенности:**
- Автоматическое управление портами (10000-20000)
- Буферизация данных с настраиваемым размером буфера
- Метрики производительности в реальном времени
- Graceful shutdown соединений

**Конфигурация:**
```json
{
    "proxy": {
        "enabled": true,
        "buffer_size": 8192,
        "timeout": "60s",
        "max_connections": 1000,
        "port_range": {
            "start": 10000,
            "end": 20000
        }
    }
}
```

### 2. Multi-tenancy

**Описание:** Поддержка мультитенантности с изоляцией ресурсов и настраиваемыми лимитами.

**Особенности:**
- Изоляция ресурсов между тенантами
- Настраиваемые лимиты на туннели, соединения, трафик
- IP-фильтрация (whitelist/blacklist)
- Отслеживание активности и автоматическая очистка

**Конфигурация:**
```yaml
tenants:
  tenant_001:
    name: "Production Tenant"
    limits:
      max_tunnels: 20
      max_connections_per_tunnel: 200
      max_bandwidth_mbps: 500
    allowed_ips:
      - "192.168.1.0/24"
```

### 3. Tunnel Manager

**Описание:** Централизованное управление активными туннелями с автоматической очисткой и мониторингом.

**Особенности:**
- Автоматическое создание и удаление туннелей
- Heartbeat мониторинг
- Graceful shutdown
- Метрики по туннелям

### 4. Enhanced Security

**Описание:** Улучшенная безопасность с интеграцией Keycloak и Django fallback.

**Особенности:**
- Интеграция с Keycloak OIDC
- Fallback на Django аутентификацию
- Улучшенная валидация JWT токенов
- Аудит всех операций

### 5. Go 1.21 и современные зависимости

**Описание:** Обновление до Go 1.21.13 с современными зависимостями.

**Улучшения:**
- Повышенная производительность
- Улучшенная безопасность
- Современные библиотеки
- Лучшая поддержка TLS 1.3

### 6. Comprehensive Testing

**Описание:** Полный набор unit и integration тестов.

**Покрытие:**
- TCP proxy функциональность
- Multi-tenancy
- Tunnel management
- Authentication
- Rate limiting
- Error handling

### Миграция с v1.x

Для миграции с предыдущих версий:

1. **Обновление Go до 1.21+**
2. **Обновление конфигурации** (добавление новых секций)
3. **Настройка multi-tenancy** (опционально)
4. **Тестирование** новых функций
5. **Обновление клиентов** для поддержки новых возможностей

## Автоматическое обновление сертификата Let's Encrypt

Certbot автоматически обновляет сертификаты, но для Docker-контейнера relay нужно перезапускать контейнер после обновления, чтобы он подхватил новый сертификат.

### Настройка post-hook для перезапуска relay

1. Создайте post-hook скрипт:
   ```bash
   sudo mkdir -p /etc/letsencrypt/renewal-hooks/post
   sudo nano /etc/letsencrypt/renewal-hooks/post/reload-relay.sh
   ```
   Вставьте в файл:
   ```bash
   #!/bin/bash
   cd /home/ubuntu/cloudbridge-relay-installer
   /usr/bin/docker-compose restart relay
   ```
   Сделайте скрипт исполняемым:
   ```bash
   sudo chmod +x /etc/letsencrypt/renewal-hooks/post/reload-relay.sh
   ```

2. Certbot будет автоматически вызывать этот скрипт после каждого успешного обновления сертификата.

3. Проверить работу можно командой:
   ```bash
   sudo certbot renew --dry-run
   ```
   В конце процесса контейнер relay будет перезапущен.

**Теперь relay всегда будет использовать актуальный сертификат без ручных действий!**

## Интеграция с Keycloak

### 1. Скопируйте сгенерированный конфиг Keycloak

```bash
cp keycloak_config.json config/keycloak_config.json
```

### 2. Обновите основной конфиг relay (config/config.json)

- Откройте файл `config/config.json`.
- В секции `keycloak`:
  - Установите `"enabled": true`.
  - Подставьте значения из `keycloak_config.json`:
    - `server_url`, `realm`, `client_id`, `client_secret`, `public_key`, `token_validation`.

Пример:
```json
"keycloak": {
  "enabled": true,
  "server_url": "http://localhost:8080",
  "realm": "cloudbridge",
  "client_id": "relay-client",
  "client_secret": "<секрет из keycloak_config.json>",
  "public_key": "<публичный ключ из keycloak_config.json>",
  "token_validation": {
    "issuer": "http://localhost:8080/realms/cloudbridge",
    "audience": "relay-client",
    "clock_skew": "30s"
  },
  "jwks_cache_ttl": "1h",
  "http_timeout": "30s"
}
```

### 3. Перезапустите relay

```bash
sudo docker-compose restart relay
```

После этого relay будет использовать Keycloak для аутентификации пользователей.