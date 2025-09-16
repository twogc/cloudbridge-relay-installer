# Задание разработчику: Установка и настройка Zabbix на удаленном сервере

## 📋 Описание задачи

**Цель**: Установить и настроить Zabbix Server на удаленном сервере с IP адресом `10.0.0.241` для мониторинга CloudBridge Relay системы.

**Приоритет**: Высокий
**Срок выполнения**: 2-3 дня
**Ответственный**: DevOps/System Administrator

## 🎯 Требования

### Целевой сервер:
- **IP адрес**: 10.0.0.241
- **ОС**: Ubuntu 22.04 LTS (рекомендуется)
- **Доступ**: SSH с ключами
- **Права**: sudo/root доступ

### Функциональные требования:
1. **Zabbix Server** - основной сервер мониторинга
2. **Zabbix Agent** - агент для сбора метрик
3. **Zabbix Web Interface** - веб-интерфейс
4. **Database** - PostgreSQL (рекомендуется) или MySQL
5. **Monitoring Templates** - шаблоны для CloudBridge
6. **Alerting** - настройка уведомлений

## 🔧 Технические требования

### Минимальные системные требования:
- **CPU**: 4 cores
- **RAM**: 8 GB
- **Диск**: 100 GB SSD
- **Сеть**: 1 Gbps

### Рекомендуемые системные требования:
- **CPU**: 8 cores
- **RAM**: 16 GB
- **Диск**: 500 GB SSD
- **Сеть**: 10 Gbps

### Порты для открытия:
- **80/443** - Zabbix Web Interface
- **10050** - Zabbix Agent
- **10051** - Zabbix Server
- **5432** - PostgreSQL (если используется)
- **3306** - MySQL (если используется)

## 📝 Пошаговые инструкции

### Этап 1: Подготовка сервера

```bash
# Подключение к серверу
ssh ubuntu@10.0.0.241

# Обновление системы
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
sudo apt install -y wget curl gnupg2 software-properties-common apt-transport-https ca-certificates lsb-release

# Настройка часового пояса
sudo timedatectl set-timezone Europe/Moscow
```

### Этап 2: Установка PostgreSQL

```bash
# Установка PostgreSQL
sudo apt install -y postgresql postgresql-contrib

# Запуск и включение PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Создание пользователя и базы данных для Zabbix
sudo -u postgres psql -c "CREATE USER zabbix WITH PASSWORD 'zabbix_password_2024';"
sudo -u postgres psql -c "CREATE DATABASE zabbix OWNER zabbix;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE zabbix TO zabbix;"

# Настройка PostgreSQL для Zabbix
sudo -u postgres psql -c "ALTER USER zabbix CREATEDB;"
```

### Этап 3: Установка Zabbix Server

```bash
# Добавление репозитория Zabbix
wget https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo apt update

# Установка Zabbix Server
sudo apt install -y zabbix-server-pgsql zabbix-frontend-php php8.1-pgsql

# Импорт схемы базы данных
sudo zcat /usr/share/zabbix-sql-scripts/postgresql/server.sql.gz | sudo -u zabbix psql zabbix

# Настройка конфигурации Zabbix Server
sudo nano /etc/zabbix/zabbix_server.conf
```

### Конфигурация Zabbix Server (/etc/zabbix/zabbix_server.conf):

```ini
# Основные настройки
DBHost=localhost
DBName=zabbix
DBUser=zabbix
DBPassword=zabbix_password_2024
DBPort=5432

# Настройки производительности
StartPollers=10
StartPollersUnreachable=5
StartTrappers=5
StartPingers=1
StartDiscoverers=1
StartHTTPPollers=1
StartTimers=1
StartEscalators=1

# Настройки кэша
CacheSize=128M
CacheUpdateFrequency=60
StartDBSyncers=4
HistoryCacheSize=64M
HistoryIndexCacheSize=32M
TrendCacheSize=32M
ValueCacheSize=256M

# Настройки логирования
LogFile=/var/log/zabbix/zabbix_server.log
LogFileSize=10
DebugLevel=3

# Настройки безопасности
AllowRoot=0
User=zabbix
```

### Этап 4: Настройка веб-интерфейса

```bash
# Настройка PHP для Zabbix
sudo nano /etc/php/8.1/apache2/php.ini
```

### Конфигурация PHP:

```ini
# Основные настройки
max_execution_time = 300
memory_limit = 256M
post_max_size = 32M
upload_max_filesize = 16M
max_input_time = 300

# Настройки сессий
session.auto_start = 0
session.cookie_httponly = 1

# Настройки даты
date.timezone = Europe/Moscow
```

```bash
# Настройка Apache
sudo nano /etc/apache2/sites-available/zabbix.conf
```

### Конфигурация Apache:

```apache
<VirtualHost *:80>
    ServerName 10.0.0.241
    DocumentRoot /usr/share/zabbix
    
    <Directory "/usr/share/zabbix">
        Options FollowSymLinks
        AllowOverride None
        Require all granted
    </Directory>
    
    <Directory "/usr/share/zabbix/conf">
        Require all denied
    </Directory>
    
    <Directory "/usr/share/zabbix/app">
        Require all denied
    </Directory>
    
    <Directory "/usr/share/zabbix/include">
        Require all denied
    </Directory>
    
    <Directory "/usr/share/zabbix/local">
        Require all denied
    </Directory>
    
    ErrorLog ${APACHE_LOG_DIR}/zabbix_error.log
    CustomLog ${APACHE_LOG_DIR}/zabbix_access.log combined
</VirtualHost>
```

```bash
# Включение сайта и перезапуск Apache
sudo a2ensite zabbix.conf
sudo a2enmod rewrite
sudo systemctl restart apache2
sudo systemctl enable apache2
```

### Этап 5: Запуск Zabbix Server

```bash
# Запуск и включение Zabbix Server
sudo systemctl start zabbix-server
sudo systemctl enable zabbix-server

# Проверка статуса
sudo systemctl status zabbix-server
```

### Этап 6: Установка Zabbix Agent

```bash
# Установка Zabbix Agent
sudo apt install -y zabbix-agent

# Настройка конфигурации
sudo nano /etc/zabbix/zabbix_agentd.conf
```

### Конфигурация Zabbix Agent:

```ini
# Основные настройки
Server=127.0.0.1,10.0.0.241
ServerActive=127.0.0.1,10.0.0.241
Hostname=Zabbix server

# Настройки безопасности
EnableRemoteCommands=0
LogRemoteCommands=1

# Настройки производительности
StartAgents=3
Timeout=4
```

```bash
# Запуск Zabbix Agent
sudo systemctl start zabbix-agent
sudo systemctl enable zabbix-agent
```

### Этап 7: Настройка файрвола

```bash
# Установка UFW
sudo apt install -y ufw

# Настройка правил файрвола
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 10050/tcp
sudo ufw allow 10051/tcp

# Включение файрвола
sudo ufw --force enable
```

### Этап 8: Настройка мониторинга CloudBridge

```bash
# Создание директории для кастомных шаблонов
sudo mkdir -p /etc/zabbix/templates/cloudbridge

# Создание шаблона для CloudBridge Relay
sudo nano /etc/zabbix/templates/cloudbridge/cloudbridge-relay.xml
```

### Этап 9: Настройка уведомлений

```bash
# Создание скрипта для отправки уведомлений
sudo mkdir -p /usr/lib/zabbix/alertscripts
sudo nano /usr/lib/zabbix/alertscripts/telegram.sh
```

### Скрипт для Telegram уведомлений:

```bash
#!/bin/bash

# Конфигурация
BOT_TOKEN="YOUR_BOT_TOKEN"
CHAT_ID="YOUR_CHAT_ID"

# Получение параметров
TO="$1"
SUBJECT="$2"
MESSAGE="$3"

# Формирование сообщения
TEXT="🚨 *Zabbix Alert*
📋 *Subject:* $SUBJECT
📝 *Message:* $MESSAGE
🕐 *Time:* $(date)
🖥️ *Server:* 10.0.0.241"

# Отправка в Telegram
curl -s -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
     -H "Content-Type: application/json" \
     -d "{\"chat_id\":\"$CHAT_ID\",\"text\":\"$TEXT\",\"parse_mode\":\"Markdown\"}"

echo "Telegram notification sent"
```

```bash
# Установка прав на выполнение
sudo chmod +x /usr/lib/zabbix/alertscripts/telegram.sh
```

## 🔍 Проверка установки

### Команды для проверки:

```bash
# Проверка статуса сервисов
sudo systemctl status zabbix-server
sudo systemctl status zabbix-agent
sudo systemctl status postgresql
sudo systemctl status apache2

# Проверка портов
sudo netstat -tlnp | grep -E "(80|443|10050|10051|5432)"

# Проверка логов
sudo tail -f /var/log/zabbix/zabbix_server.log
sudo tail -f /var/log/zabbix/zabbix_agentd.log

# Проверка веб-интерфейса
curl -I http://10.0.0.241/zabbix
```

### Доступ к веб-интерфейсу:
- **URL**: http://10.0.0.241/zabbix
- **Логин по умолчанию**: Admin
- **Пароль по умолчанию**: zabbix

## 📊 Настройка мониторинга CloudBridge

### 1. Создание хоста для CloudBridge Relay Server

```bash
# В веб-интерфейсе Zabbix:
# Configuration → Hosts → Create host
# Host name: CloudBridge-Relay-Server
# Groups: CloudBridge
# IP address: [IP адрес CloudBridge сервера]
# Port: 10050
```

### 2. Импорт шаблонов

```bash
# Создание шаблона для CloudBridge
sudo nano /etc/zabbix/templates/cloudbridge/cloudbridge-template.xml
```

### 3. Настройка элементов данных

```bash
# Основные метрики для мониторинга:
# - CPU usage
# - Memory usage
# - Disk usage
# - Network traffic
# - Active connections
# - AI metrics
# - Response time
```

## 🚨 Настройка алертов

### Критические алерты:
1. **Сервис недоступен** - CloudBridge Relay Server down
2. **Высокая загрузка CPU** - CPU > 80%
3. **Нехватка памяти** - Memory > 90%
4. **Заполнение диска** - Disk > 85%
5. **Высокая задержка** - Response time > 5s
6. **AI сервис недоступен** - AI Service down

### Предупреждающие алерты:
1. **Средняя загрузка CPU** - CPU > 60%
2. **Среднее использование памяти** - Memory > 70%
3. **Среднее заполнение диска** - Disk > 70%
4. **Средняя задержка** - Response time > 2s

## 📈 Настройка дашбордов

### Основные дашборды:
1. **CloudBridge Overview** - общий обзор системы
2. **AI Performance** - производительность AI
3. **Network Metrics** - сетевые метрики
4. **System Resources** - системные ресурсы
5. **Alerts Summary** - сводка алертов

## 🔧 Дополнительные настройки

### 1. Настройка резервного копирования

```bash
# Создание скрипта резервного копирования
sudo nano /usr/local/bin/zabbix-backup.sh
```

### 2. Настройка автоматического обновления

```bash
# Создание скрипта обновления
sudo nano /usr/local/bin/zabbix-update.sh
```

### 3. Настройка мониторинга самого Zabbix

```bash
# Создание шаблона для мониторинга Zabbix
sudo nano /etc/zabbix/templates/cloudbridge/zabbix-self-monitoring.xml
```

## 📋 Чек-лист выполнения

### Обязательные задачи:
- [ ] Установка PostgreSQL
- [ ] Установка Zabbix Server
- [ ] Настройка веб-интерфейса
- [ ] Установка Zabbix Agent
- [ ] Настройка файрвола
- [ ] Создание пользователей и групп
- [ ] Настройка шаблонов CloudBridge
- [ ] Настройка алертов
- [ ] Настройка уведомлений
- [ ] Создание дашбордов
- [ ] Тестирование мониторинга
- [ ] Настройка резервного копирования

### Дополнительные задачи:
- [ ] Настройка SSL сертификатов
- [ ] Настройка LDAP аутентификации
- [ ] Настройка API для интеграции
- [ ] Настройка автоматического масштабирования
- [ ] Настройка мониторинга самого Zabbix

## 🎯 Критерии успеха

### Функциональные критерии:
1. **Доступность веб-интерфейса** - 99.9%
2. **Время отклика** - < 2 секунды
3. **Покрытие мониторинга** - 100% критических компонентов
4. **Время обнаружения проблем** - < 1 минуты
5. **Время доставки алертов** - < 30 секунд

### Технические критерии:
1. **Стабильность работы** - без перезагрузок
2. **Производительность** - CPU < 50%, Memory < 70%
3. **Безопасность** - все порты защищены
4. **Резервное копирование** - ежедневные бэкапы
5. **Документация** - полная документация настроек

## 📞 Контакты и поддержка

### Ответственные:
- **DevOps Lead**: devops-lead@cloudbridge.com
- **System Administrator**: sysadmin@cloudbridge.com
- **AI Team**: ai-team@cloudbridge.com

### Документация:
- [Zabbix Official Documentation](https://www.zabbix.com/documentation)
- [CloudBridge Monitoring Guide](AI_MONITORING_SETUP.md)
- [Troubleshooting Guide](AI_TROUBLESHOOTING.md)

### Полезные команды:

```bash
# Перезапуск сервисов
sudo systemctl restart zabbix-server
sudo systemctl restart zabbix-agent
sudo systemctl restart postgresql
sudo systemctl restart apache2

# Проверка логов
sudo journalctl -u zabbix-server -f
sudo journalctl -u zabbix-agent -f

# Проверка конфигурации
sudo zabbix_server -t
sudo zabbix_agentd -t

# Очистка кэша
sudo systemctl restart zabbix-server
```

## 🚀 Заключение

После выполнения всех задач у вас будет полностью настроенная система мониторинга Zabbix, которая обеспечит:

1. **Полный мониторинг** CloudBridge Relay системы
2. **Автоматические алерты** при возникновении проблем
3. **Веб-интерфейс** для просмотра метрик и управления
4. **Масштабируемость** для будущего роста системы
5. **Надежность** с резервным копированием и восстановлением

Система готова к использованию в продакшене и может быть интегрирована с существующими системами мониторинга.
