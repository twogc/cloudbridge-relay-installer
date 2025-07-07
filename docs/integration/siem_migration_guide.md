# Руководство по миграции SIEM на отдельный сервер

## 🎯 Обзор миграции

### Зачем нужна миграция?

**Проблемы текущего развертывания:**
- ❌ Конкуренция ресурсов с основными сервисами
- ❌ Сложности с масштабированием
- ❌ Риски безопасности (общая инфраструктура)
- ❌ Ограничения производительности
- ❌ Сложности с backup и восстановлением

**Преимущества отдельного сервера:**
- ✅ Выделенные ресурсы для SIEM
- ✅ Изоляция security инфраструктуры
- ✅ Независимое масштабирование
- ✅ Повышенная надежность
- ✅ Соответствие стандартам compliance

## 📋 План миграции

### Этап 1: Подготовка (1-2 дня)

#### 1.1. Анализ текущего состояния
```bash
# Проверка текущих ресурсов
df -h
free -h
top -bn1 | grep "Cpu(s)"

# Проверка текущих сервисов SIEM
docker ps | grep siem
docker-compose -f siem/docker-compose.siem.yml ps

# Проверка конфигурации
ls -la siem/config/
cat siem/config/siem_config.yaml
```

#### 1.2. Подготовка нового сервера
```bash
# На новом сервере
# Скачивание и запуск скрипта развертывания
wget https://raw.githubusercontent.com/maxlanies/cloudbridge-relay-installer/main/scripts/deploy_siem_server.sh
chmod +x deploy_siem_server.sh
sudo ./deploy_siem_server.sh --full
```

#### 1.3. Настройка сетевого подключения
```bash
# Настройка firewall для доступа к новому SIEM серверу
sudo ufw allow from OLD_SERVER_IP to any port 8081  # SIEM Web UI
sudo ufw allow from OLD_SERVER_IP to any port 5601  # Kibana
sudo ufw allow from OLD_SERVER_IP to any port 9200  # Elasticsearch
```

### Этап 2: Перенос данных (1 день)

#### 2.1. Backup текущих данных
```bash
# На старом сервере - создание backup
BACKUP_DIR="/tmp/siem_migration_backup"
mkdir -p $BACKUP_DIR

# Backup конфигурации
tar -czf $BACKUP_DIR/siem_config_$(date +%Y%m%d_%H%M%S).tar.gz \
    siem/config/ \
    siem/.env \
    siem/docker-compose.siem.yml

# Backup данных Elasticsearch
curl -X PUT "localhost:9200/_snapshot/siem_migration/snapshot_$(date +%Y%m%d_%H%M%S)" \
    -H "Content-Type: application/json" \
    -d '{
        "indices": "siem-*",
        "ignore_unavailable": true,
        "include_global_state": false
    }'

# Backup пользовательских данных
docker exec siem-elasticsearch-1 tar -czf /tmp/elasticsearch_data.tar.gz /usr/share/elasticsearch/data
docker cp siem-elasticsearch-1:/tmp/elasticsearch_data.tar.gz $BACKUP_DIR/

echo "Backup создан в: $BACKUP_DIR"
```

#### 2.2. Перенос данных на новый сервер
```bash
# Копирование backup на новый сервер
scp -r $BACKUP_DIR ubuntu@NEW_SIEM_SERVER_IP:/opt/siem/

# На новом сервере - восстановление данных
cd /opt/siem

# Восстановление конфигурации
tar -xzf siem_migration_backup/siem_config_*.tar.gz -C siem/

# Восстановление данных Elasticsearch
docker exec siem-elasticsearch-1 mkdir -p /tmp/restore
docker cp siem_migration_backup/elasticsearch_data.tar.gz siem-elasticsearch-1:/tmp/restore/
docker exec siem-elasticsearch-1 tar -xzf /tmp/restore/elasticsearch_data.tar.gz -C /usr/share/elasticsearch/
docker exec siem-elasticsearch-1 chown -R elasticsearch:elasticsearch /usr/share/elasticsearch/data
```

#### 2.3. Валидация данных
```bash
# Проверка индексов в Elasticsearch
curl -X GET "localhost:9200/_cat/indices?v"

# Проверка количества документов
curl -X GET "localhost:9200/siem-*/_count"

# Проверка конфигурации
docker-compose -f siem/docker-compose.siem.yml config
```

### Этап 3: Обновление конфигурации (2-3 часа)

#### 3.1. Обновление переменных окружения
```bash
# На новом сервере - обновление .env файла
cat > /opt/siem/.env.siem << 'EOF'
# Elasticsearch
ELASTICSEARCH_PASSWORD=your_secure_password
ELASTICSEARCH_USERNAME=elastic

# SIEM Configuration
SIEM_ENVIRONMENT=production
SIEM_LOG_LEVEL=INFO

# Network Configuration
SIEM_HOST=NEW_SIEM_SERVER_IP
RELAY_HOST=OLD_SERVER_IP
KEYCLOAK_HOST=OLD_SERVER_IP
DJANGO_HOST=OLD_SERVER_IP

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
EOF
```

#### 3.2. Обновление конфигурации коллекторов
```bash
# Обновление конфигурации удаленного сбора
cat > /opt/siem/siem/config/remote_collectors.yaml << 'EOF'
remote_collectors:
  relay_server:
    host: "OLD_SERVER_IP"
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
    host: "OLD_SERVER_IP"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "http://localhost:8080/admin/events"
        type: "keycloak_events"
        interval: 120

  django_server:
    host: "OLD_SERVER_IP"
    port: 22
    username: "ubuntu"
    key_file: "/opt/siem/.ssh/id_rsa"
    sources:
      - path: "http://localhost:8000/api/audit/"
        type: "django_audit"
        interval: 60
EOF
```

#### 3.3. Настройка SSH ключей
```bash
# Генерация SSH ключей на новом SIEM сервере
ssh-keygen -t rsa -b 4096 -f /opt/siem/.ssh/id_rsa -N ""
chmod 600 /opt/siem/.ssh/id_rsa
chmod 644 /opt/siem/.ssh/id_rsa.pub

# Копирование публичного ключа на старый сервер
ssh-copy-id -i /opt/siem/.ssh/id_rsa.pub ubuntu@OLD_SERVER_IP

# Тест подключения
ssh ubuntu@OLD_SERVER_IP "echo 'SSH connection successful'"
```

### Этап 4: Тестирование (2-3 часа)

#### 4.1. Проверка сервисов
```bash
# Проверка всех сервисов на новом сервере
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

#### 4.2. Тест сбора данных
```bash
# Тест подключения к старому серверу
ssh ubuntu@OLD_SERVER_IP "ls -la /var/log/relay/"

# Тест сбора логов
curl -X POST "http://localhost:8081/api/v1/collectors/test" \
    -H "Content-Type: application/json" \
    -d '{
        "collector": "relay_server",
        "source": "relay_logs"
    }'
```

#### 4.3. Тест обработки событий
```bash
# Отправка тестового события
test_event='{
  "tenant_id": "test_tenant",
  "event_type": "migration_test",
  "severity": "medium",
  "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
  "source_ip": "192.168.1.100",
  "user_id": "test_user",
  "action": "migration_test"
}'

curl -X POST "http://localhost:8081/api/v1/events" \
    -H "Content-Type: application/json" \
    -d "$test_event"

# Проверка в Elasticsearch
curl -X GET "localhost:9200/siem-*/_search?q=event_type:migration_test"
```

### Этап 5: Переключение (1 час)

#### 5.1. Обновление DNS/конфигурации
```bash
# Обновление DNS записи (если используется)
# siem.yourcompany.com -> NEW_SIEM_SERVER_IP

# Обновление конфигурации на старом сервере
# Перенаправление запросов к SIEM на новый сервер
```

#### 5.2. Остановка SIEM на старом сервере
```bash
# На старом сервере
cd /path/to/siem
docker-compose -f docker-compose.siem.yml down

# Проверка остановки
docker ps | grep siem
```

#### 5.3. Финальная проверка
```bash
# Проверка работы всех сервисов на новом сервере
systemctl status siem.service

# Проверка сбора данных
tail -f /opt/siem/logs/collector.log

# Проверка алертов
tail -f /opt/siem/logs/alert_manager.log
```

### Этап 6: Очистка (1 час)

#### 6.1. Удаление SIEM с старого сервера
```bash
# На старом сервере
# Удаление контейнеров SIEM
docker-compose -f siem/docker-compose.siem.yml down -v

# Удаление образов
docker images | grep siem | awk '{print $3}' | xargs docker rmi

# Удаление директорий
sudo rm -rf /path/to/siem
sudo rm -rf /opt/siem  # если была установлена в /opt
```

#### 6.2. Освобождение ресурсов
```bash
# Проверка освобожденных ресурсов
df -h
free -h

# Очистка неиспользуемых Docker ресурсов
docker system prune -f
```

## 🔧 Автоматизированная миграция

### Скрипт автоматической миграции

```bash
#!/bin/bash
# migrate_siem.sh

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

# Проверка аргументов
if [ $# -ne 2 ]; then
    echo "Использование: $0 OLD_SERVER_IP NEW_SIEM_SERVER_IP"
    exit 1
fi

OLD_SERVER_IP=$1
NEW_SIEM_SERVER_IP=$2

log_info "Начинаем миграцию SIEM с $OLD_SERVER_IP на $NEW_SIEM_SERVER_IP"

# Этап 1: Backup на старом сервере
log_info "Этап 1: Создание backup на старом сервере"
ssh ubuntu@$OLD_SERVER_IP << 'EOF'
BACKUP_DIR="/tmp/siem_migration_backup_$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

# Backup конфигурации
tar -czf $BACKUP_DIR/siem_config.tar.gz siem/config/ siem/.env siem/docker-compose.siem.yml

# Backup данных Elasticsearch
curl -X PUT "localhost:9200/_snapshot/siem_migration/snapshot_$(date +%Y%m%d_%H%M%S)" \
    -H "Content-Type: application/json" \
    -d '{"indices": "siem-*", "ignore_unavailable": true, "include_global_state": false}'

echo "Backup создан в: $BACKUP_DIR"
EOF

# Этап 2: Копирование backup на новый сервер
log_info "Этап 2: Копирование backup на новый сервер"
BACKUP_DIR=$(ssh ubuntu@$OLD_SERVER_IP "ls -td /tmp/siem_migration_backup_* | head -1")
scp -r ubuntu@$OLD_SERVER_IP:$BACKUP_DIR /opt/siem/

# Этап 3: Восстановление на новом сервере
log_info "Этап 3: Восстановление на новом сервере"
cd /opt/siem
tar -xzf siem_migration_backup_*/siem_config.tar.gz -C siem/

# Этап 4: Обновление конфигурации
log_info "Этап 4: Обновление конфигурации"
sed -i "s/OLD_SERVER_IP/$OLD_SERVER_IP/g" siem/config/remote_collectors.yaml
sed -i "s/NEW_SIEM_SERVER_IP/$NEW_SIEM_SERVER_IP/g" .env.siem

# Этап 5: Перезапуск сервисов
log_info "Этап 5: Перезапуск сервисов"
systemctl restart siem.service

# Этап 6: Тестирование
log_info "Этап 6: Тестирование"
sleep 30

# Проверка сервисов
services=("8081" "5601" "9200")
for port in "${services[@]}"; do
    if curl -s "http://localhost:$port" > /dev/null 2>&1; then
        log_success "Сервис на порту $port доступен"
    else
        log_error "Сервис на порту $port недоступен"
    fi
done

log_success "Миграция завершена успешно!"
```

### Запуск автоматической миграции

```bash
# Установка прав на выполнение
chmod +x migrate_siem.sh

# Запуск миграции
./migrate_siem.sh OLD_SERVER_IP NEW_SIEM_SERVER_IP
```

## 🚨 Troubleshooting

### Проблемы с подключением

```bash
# Проверка сетевого подключения
ping NEW_SIEM_SERVER_IP
telnet NEW_SIEM_SERVER_IP 22

# Проверка SSH ключей
ssh -i /opt/siem/.ssh/id_rsa ubuntu@OLD_SERVER_IP "echo 'SSH test'"

# Проверка firewall
sudo ufw status
```

### Проблемы с данными

```bash
# Проверка индексов Elasticsearch
curl -X GET "localhost:9200/_cat/indices?v"

# Проверка snapshots
curl -X GET "localhost:9200/_snapshot/siem_migration/_all"

# Восстановление из snapshot
curl -X POST "localhost:9200/_snapshot/siem_migration/snapshot_NAME/_restore" \
    -H "Content-Type: application/json" \
    -d '{"indices": "siem-*"}'
```

### Проблемы с сервисами

```bash
# Проверка статуса сервисов
systemctl status siem.service
docker ps -a

# Просмотр логов
journalctl -u siem.service -f
docker-compose -f siem/docker-compose.siem.yml logs -f
```

## 📊 Валидация миграции

### Чек-лист успешной миграции

- [ ] Все сервисы SIEM доступны на новом сервере
- [ ] Данные успешно перенесены (проверить количество документов)
- [ ] Сбор данных с старого сервера работает
- [ ] Алерты и уведомления функционируют
- [ ] Дашборды отображают актуальные данные
- [ ] Производительность соответствует ожиданиям
- [ ] Backup стратегия настроена
- [ ] Мониторинг работает корректно

### Метрики успеха

```bash
# Проверка производительности
curl -X GET "localhost:9200/_cluster/stats?pretty"

# Проверка сбора данных
tail -n 100 /opt/siem/logs/collector.log | grep "events processed"

# Проверка алертов
tail -n 100 /opt/siem/logs/alert_manager.log | grep "alert sent"
```

## 🔗 Заключение

Миграция SIEM на отдельный сервер обеспечивает:

1. **Улучшенную производительность** - выделенные ресурсы
2. **Повышенную безопасность** - изоляция инфраструктуры
3. **Лучшую масштабируемость** - независимое масштабирование
4. **Повышенную надежность** - отказоустойчивость

**Рекомендации:**
- Проводить миграцию в нерабочее время
- Иметь план отката на случай проблем
- Тестировать все функции после миграции
- Документировать все изменения
- Настроить мониторинг новой инфраструктуры 