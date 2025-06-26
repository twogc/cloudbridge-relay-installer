# Руководство по развертыванию CloudBridge Relay

## Содержание

1. [Требования](#требования)
2. [Подготовка окружения](#подготовка-окружения)
3. [Установка](#установка)
4. [Конфигурация](#конфигурация)
5. [Развертывание](#развертывание)
6. [Мониторинг](#мониторинг)
7. [Обслуживание](#обслуживание)
8. [Устранение неполадок](#устранение-неполадок)

## Требования

### Системные требования

- **ОС:** Linux (Ubuntu 20.04+, CentOS 8+, RHEL 8+)
- **CPU:** 2+ ядра
- **RAM:** 4+ GB
- **Диск:** 20+ GB свободного места
- **Сеть:** статический IP-адрес

### Программные требования

- **Go:** 1.18+
- **Docker:** 20.10+ (опционально)
- **Docker Compose:** 2.0+ (опционально)
- **Systemd:** для установки как сервис
- **Certbot:** для SSL сертификатов

### Сетевые требования

- **Порт 8080:** основной порт для клиентов
- **Порт 8081:** порт для метрик и API
- **Порт 80:** для получения SSL сертификатов
- **Порт 443:** для HTTPS (опционально)

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

### 3. Установка Go

```bash
# Скачивание Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Добавление в PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Проверка установки
go version
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

## Установка

### Способ 1: Установка из исходного кода

```bash
# Клонирование репозитория
git clone https://github.com/maxlanies/cloudbridge-relay-installer.git
cd cloudbridge-relay-installer

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
docker build -t cloudbridge-relay .

# Запуск контейнера
docker run -d \
  --name relay \
  -p 8080:8080 \
  -p 8081:8081 \
  -v /etc/relay:/etc/relay \
  -v /var/log/relay:/var/log/relay \
  -e RELAY_JWT_SECRET=your-secret-key \
  cloudbridge-relay ./relay --config config/config.json
```

### Способ 3: Установка как системный сервис

```bash
# Запуск скрипта установки
sudo ./install.sh

# Проверка статуса
sudo systemctl status relay
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
EOF
```

### 3. Формат конфигурационного файла

CloudBridge Relay поддерживает конфигурацию в формате JSON с поддержкой строковых значений для time.Duration полей:

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
        "enabled": false,
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
        "http_timeout": "30s"
    },
    "security": {
        "tls_enabled": true,
        "cert_file": "/etc/relay/certs/fullchain.pem",
        "key_file": "/etc/relay/certs/privkey.pem",
        "allowed_ips": "",
        "rate_limit": {
            "enabled": true,
            "requests_per_second": 100,
            "burst": 200
        }
    },
    "metrics": {
        "enabled": true,
        "port": ":8081",
        "path": "/metrics"
    },
    "logging": {
        "level": "INFO",
        "format": "json",
        "output": "/var/log/relay/relay.log"
    }
}
```

**Важно:** Все поля типа `time.Duration` должны быть указаны в виде строк с суффиксами времени:
- `s` - секунды
- `m` - минуты  
- `h` - часы
- `ms` - миллисекунды

### 4. Настройка SSL сертификатов

```bash
# Установка Certbot
sudo apt install -y certbot

# Получение сертификата
sudo certbot certonly --standalone -d your-domain.com

# Копирование сертификатов
sudo mkdir -p /etc/relay/certs
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem /etc/relay/certs/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem /etc/relay/certs/

# Установка прав доступа
sudo chown -R relay:relay /etc/relay/certs
sudo chmod 600 /etc/relay/certs/*.pem
```

### 5. Настройка файрвола

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 8080/tcp
sudo ufw allow 8081/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=8081/tcp
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --reload
```

## Развертывание

### Запуск сервера

```bash
# Запуск с конфигурационным файлом
./relay --config config/config.json --log-level DEBUG

# Запуск с переменными окружения
RELAY_JWT_SECRET=your-secret ./relay --config config/config.json

# Запуск в Docker
docker run -d \
  --name relay \
  -p 8080:8080 \
  -p 8081:8081 \
  -e RELAY_JWT_SECRET=your-secret \
  -v /etc/relay:/etc/relay \
  cloudbridge-relay ./relay --config config/config.json
```

### Проверка работоспособности

```bash
# Проверка метрик
curl http://localhost:8081/metrics

# Проверка здоровья сервера
curl http://localhost:8081/health

# Проверка логов
tail -f /var/log/relay/relay.log
```

## Мониторинг

### Prometheus метрики

Сервер автоматически экспортирует метрики на порту 8081:

```bash
# Получение метрик
curl http://localhost:8081/metrics

# Основные метрики
relay_connections_total          # Общее количество соединений
relay_active_connections         # Активные соединения
relay_active_tunnels            # Активные туннели
relay_errors_total              # Количество ошибок
relay_heartbeats_total          # Количество heartbeat
```

### Логирование

CloudBridge Relay использует структурированное логирование с поддержкой:

- **Форматы:** JSON, Console
- **Уровни:** DEBUG, INFO, WARN, ERROR
- **Ротация:** Автоматическая с настройкой размера и количества файлов
- **Поля:** timestamp, level, message, user_id, request_id, connection_id, duration, error

Пример лога:
```json
{
  "level": "INFO",
  "timestamp": "2024-03-14T12:00:00Z",
  "caller": "main.go:123",
  "msg": "Server started",
  "port": 8080,
  "tls": true,
  "workers": 4
}
```

## Обслуживание

### Обновление сертификатов

```bash
# Автоматическое обновление через Certbot
sudo certbot renew

# Копирование обновленных сертификатов
sudo cp /etc/letsencrypt/live/your-domain.com/fullchain.pem /etc/relay/certs/
sudo cp /etc/letsencrypt/live/your-domain.com/privkey.pem /etc/relay/certs/

# Перезапуск сервера
sudo systemctl restart relay
```

### Ротация логов

Логи автоматически ротируются с настройками:
- Максимальный размер файла: 100MB
- Количество файлов: 10
- Срок хранения: 30 дней
- Сжатие: включено

### Мониторинг производительности

```bash
# Проверка использования ресурсов
top -p $(pgrep relay)

# Проверка сетевых соединений
netstat -tlnp | grep relay

# Проверка логов в реальном времени
tail -f /var/log/relay/relay.log | jq .
```

## Устранение неполадок

### Частые проблемы

#### 1. Ошибка парсинга конфигурации
```
Failed to load configuration: failed to parse config file: json: cannot unmarshal string into Go struct field .server.timeout of type time.Duration
```

**Решение:** Убедитесь, что все поля time.Duration указаны в виде строк с суффиксами времени (например, "30s", "5m", "1h").

#### 2. Порт уже занят
```
Failed to start server: failed to create listener: listen tcp :8080: bind: address already in use
```

**Решение:** Остановите процесс, использующий порт 8080, или измените порт в конфигурации.

#### 3. Ошибка JWT секрета
```
JWT secret must be set via RELAY_JWT_SECRET environment variable
```

**Решение:** Установите переменную окружения RELAY_JWT_SECRET или замените placeholder в конфигурации.

#### 4. Ошибка Docker контейнера
```
docker: Error response from daemon: failed to create task for container: failed to create shim task: OCI runtime create failed: runc create failed: unable to start container process: error during container init: exec: "--config": executable file not found in $PATH
```

**Решение:** Укажите правильную команду запуска: `cloudbridge-relay ./relay --config config/config.json`

### Диагностика

```bash
# Проверка статуса сервиса
sudo systemctl status relay

# Просмотр логов systemd
sudo journalctl -u relay -f

# Проверка конфигурации
./relay --config config/config.json --help

# Тестирование подключения
telnet localhost 8080
telnet localhost 8081

# Проверка метрик
curl -s http://localhost:8081/metrics | grep relay_
```

### Получение поддержки

При возникновении проблем:

1. Соберите логи: `sudo journalctl -u relay > relay-logs.txt`
2. Соберите метрики: `curl http://localhost:8081/metrics > metrics.txt`
3. Проверьте конфигурацию: `cat /etc/relay/config.json`
4. Создайте issue в репозитории с приложенными файлами

## Безопасность

### 1. Рекомендации по безопасности

- Используйте сильные JWT секреты
- Включите TLS для всех соединений
- Настройте rate limiting
- Ограничьте доступ по IP-адресам
- Регулярно обновляйте сертификаты
- Мониторьте логи на подозрительную активность

### 2. Аудит безопасности

```bash
# Проверка прав доступа
sudo find /etc/relay -type f -exec ls -la {} \;

# Проверка открытых портов
sudo netstat -tlnp

# Проверка процессов
ps aux | grep relay

# Проверка логов безопасности
sudo journalctl -u relay | grep -i "security\|auth\|error"
```

## Поддержка

### Контакты для поддержки

- **GitHub Issues:** https://github.com/maxlanies/cloudbridge-relay-installer/issues
- **Документация:** https://github.com/maxlanies/cloudbridge-relay-installer/docs
- **Wiki:** https://github.com/maxlanies/cloudbridge-relay-installer/wiki

### Полезные команды

```bash
# Полная диагностика системы
sudo relay --diagnostic

# Проверка здоровья сервиса
curl http://localhost:8081/health

# Получение метрик
curl http://localhost:8081/metrics

# Просмотр логов в реальном времени
sudo journalctl -u relay -f

# Перезапуск сервиса
sudo systemctl restart relay

# Проверка конфигурации
relay --config /etc/relay/config.json --validate
``` 