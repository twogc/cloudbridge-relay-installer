# DevOps Guide: Установка и настройка Zabbix для CloudBridge

## 📋 Обзор

Данное руководство предназначено для DevOps инженеров по установке и настройке системы мониторинга Zabbix на сервере `10.0.0.241` для мониторинга CloudBridge Relay системы.

## 🎯 Цели установки

- **Мониторинг производительности** CloudBridge Relay сервера
- **Автоматические алерты** при критических событиях
- **Веб-интерфейс** для просмотра метрик и управления
- **Интеграция с Telegram** для уведомлений
- **Резервное копирование** конфигурации и данных

## 🖥️ Системные требования

### Минимальные требования:
- **CPU**: 4 cores
- **RAM**: 8 GB
- **Диск**: 100 GB SSD
- **Сеть**: 1 Gbps

### Рекомендуемые требования:
- **CPU**: 8 cores
- **RAM**: 16 GB
- **Диск**: 500 GB SSD
- **Сеть**: 10 Gbps

## 🚀 Быстрая установка

### Вариант 1: Автоматическая установка

1. **Подключитесь к серверу:**
   ```bash
   ssh ubuntu@10.0.0.241
   ```

2. **Скачайте и выполните скрипт установки:**
   ```bash
   wget https://raw.githubusercontent.com/your-repo/cloudbridge-relay-installer/main/scripts/zabbix_complete_install.sh
   chmod +x zabbix_complete_install.sh
   sudo ./zabbix_complete_install.sh
   ```

3. **Настройте мониторинг CloudBridge:**
   ```bash
   wget https://raw.githubusercontent.com/your-repo/cloudbridge-relay-installer/main/scripts/cloudbridge_monitoring_setup.sh
   chmod +x cloudbridge_monitoring_setup.sh
   sudo ./cloudbridge_monitoring_setup.sh
   ```

### Вариант 2: Ручная установка

Следуйте пошаговым инструкциям ниже.

## 📝 Пошаговая установка

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
sudo -u postgres psql -c "ALTER USER zabbix CREATEDB;"
```

### Этап 3: Установка Zabbix Server

```bash
# Добавление репозитория Zabbix
wget https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo apt update

# Установка Zabbix Server и компонентов
sudo apt install -y zabbix-server-pgsql zabbix-frontend-php php8.1-pgsql zabbix-sql-scripts

# Импорт схемы базы данных
sudo zcat /usr/share/zabbix-sql-scripts/postgresql/server.sql.gz | sudo -u zabbix psql zabbix
```

### Этап 4: Настройка конфигурации Zabbix Server

```bash
# Создание директорий для логов и PID
sudo mkdir -p /var/log/zabbix /run/zabbix
sudo chown zabbix:zabbix /var/log/zabbix /run/zabbix

# Создание конфигурационного файла
sudo tee /etc/zabbix/zabbix_server.conf > /dev/null <<EOF
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
EOF

# Установка правильных прав
sudo chown zabbix:zabbix /etc/zabbix/zabbix_server.conf
sudo chmod 640 /etc/zabbix/zabbix_server.conf
```

### Этап 5: Настройка веб-интерфейса

```bash
# Установка Apache и PHP
sudo apt install -y apache2 php8.1 libapache2-mod-php8.1

# Настройка PHP
sudo sed -i 's/;date.timezone =/date.timezone = Europe\/Moscow/' /etc/php/8.1/apache2/php.ini
sudo sed -i 's/max_execution_time = 30/max_execution_time = 300/' /etc/php/8.1/apache2/php.ini
sudo sed -i 's/memory_limit = 128M/memory_limit = 256M/' /etc/php/8.1/apache2/php.ini

# Создание конфигурации Apache для Zabbix
sudo tee /etc/apache2/sites-available/zabbix.conf > /dev/null <<EOF
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
    
    ErrorLog \${APACHE_LOG_DIR}/zabbix_error.log
    CustomLog \${APACHE_LOG_DIR}/zabbix_access.log combined
</VirtualHost>
EOF

# Включение сайта и модулей
sudo a2ensite zabbix.conf
sudo a2enmod rewrite
sudo systemctl restart apache2
sudo systemctl enable apache2
```

### Этап 6: Запуск Zabbix Server

```bash
# Запуск и включение Zabbix Server
sudo systemctl start zabbix-server
sudo systemctl enable zabbix-server

# Проверка статуса
sudo systemctl status zabbix-server
```

### Этап 7: Установка Zabbix Agent

```bash
# Установка Zabbix Agent
sudo apt install -y zabbix-agent

# Настройка конфигурации
sudo tee /etc/zabbix/zabbix_agentd.conf > /dev/null <<EOF
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
EOF

# Запуск Zabbix Agent
sudo systemctl start zabbix-agent
sudo systemctl enable zabbix-agent
```

### Этап 8: Настройка файрвола

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

1. Откройте веб-интерфейс Zabbix
2. Перейдите в **Configuration → Hosts → Create host**
3. Заполните поля:
   - **Host name**: CloudBridge-Relay-Server
   - **Groups**: CloudBridge
   - **IP address**: [IP адрес CloudBridge сервера]
   - **Port**: 10050

### 2. Импорт шаблона CloudBridge

```bash
# Создание шаблона для CloudBridge
sudo tee /etc/zabbix/templates/cloudbridge/cloudbridge-template.xml > /dev/null <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<zabbix_export>
    <version>6.0</version>
    <date>2024-09-14T02:00:00Z</date>
    <groups>
        <group>
            <uuid>cloudbridge-group-uuid</uuid>
            <name>CloudBridge</name>
        </group>
    </groups>
    <templates>
        <template>
            <uuid>cloudbridge-template-uuid</uuid>
            <template>CloudBridge Relay Server</template>
            <name>CloudBridge Relay Server</name>
            <description>Template for monitoring CloudBridge Relay Server</description>
            <groups>
                <group>
                    <name>CloudBridge</name>
                </group>
            </groups>
            <items>
                <item>
                    <uuid>cloudbridge-cpu-uuid</uuid>
                    <name>CPU Usage</name>
                    <type>0</type>
                    <key>system.cpu.util[,avg1]</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>0</value_type>
                    <units>%</units>
                </item>
                <item>
                    <uuid>cloudbridge-memory-uuid</uuid>
                    <name>Memory Usage</name>
                    <type>0</type>
                    <key>vm.memory.util[pused]</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>0</value_type>
                    <units>%</units>
                </item>
                <item>
                    <uuid>cloudbridge-disk-uuid</uuid>
                    <name>Disk Usage</name>
                    <type>0</type>
                    <key>vfs.fs.util[/,pused]</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>0</value_type>
                    <units>%</units>
                </item>
            </items>
            <triggers>
                <trigger>
                    <uuid>cloudbridge-cpu-high-uuid</uuid>
                    <expression>last(/CloudBridge Relay Server/system.cpu.util[,avg1])>80</expression>
                    <name>High CPU Usage</name>
                    <priority>3</priority>
                    <description>CPU usage is above 80%</description>
                </trigger>
                <trigger>
                    <uuid>cloudbridge-memory-high-uuid</uuid>
                    <expression>last(/CloudBridge Relay Server/vm.memory.util[pused])>90</expression>
                    <name>High Memory Usage</name>
                    <priority>4</priority>
                    <description>Memory usage is above 90%</description>
                </trigger>
                <trigger>
                    <uuid>cloudbridge-disk-high-uuid</uuid>
                    <expression>last(/CloudBridge Relay Server/vfs.fs.util[/,pused])>85</expression>
                    <name>High Disk Usage</name>
                    <priority>3</priority>
                    <description>Disk usage is above 85%</description>
                </trigger>
            </triggers>
        </template>
    </templates>
</zabbix_export>
EOF
```

### 3. Импорт шаблона в веб-интерфейсе

1. Перейдите в **Configuration → Templates → Import**
2. Выберите файл `/etc/zabbix/templates/cloudbridge/cloudbridge-template.xml`
3. Нажмите **Import**

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

## 📱 Настройка Telegram уведомлений

### 1. Создание Telegram бота

1. Найдите @BotFather в Telegram
2. Отправьте команду `/newbot`
3. Следуйте инструкциям для создания бота
4. Сохраните полученный токен

### 2. Получение Chat ID

1. Добавьте бота в группу или отправьте сообщение боту
2. Отправьте GET запрос:
   ```bash
   curl https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates
   ```
3. Найдите `chat.id` в ответе

### 3. Настройка скрипта уведомлений

```bash
# Создание скрипта для Telegram уведомлений
sudo tee /usr/lib/zabbix/alertscripts/telegram.sh > /dev/null <<EOF
#!/bin/bash

# Конфигурация Telegram
BOT_TOKEN="YOUR_BOT_TOKEN"
CHAT_ID="YOUR_CHAT_ID"

# Получение параметров
TO="\$1"
SUBJECT="\$2"
MESSAGE="\$3"

# Формирование сообщения
TEXT="🚨 *Zabbix Alert*
📋 *Subject:* \$SUBJECT
📝 *Message:* \$MESSAGE
🕐 *Time:* \$(date)
🖥️ *Server:* 10.0.0.241"

# Отправка в Telegram
curl -s -X POST "https://api.telegram.org/bot\$BOT_TOKEN/sendMessage" \\
     -H "Content-Type: application/json" \\
     -d "{\\"chat_id\\":\\"\$CHAT_ID\\",\\"text\\":\\"\$TEXT\\",\\"parse_mode\\":\\"Markdown\\"}"

echo "Telegram notification sent"
EOF

sudo chmod +x /usr/lib/zabbix/alertscripts/telegram.sh
```

### 4. Настройка в веб-интерфейсе

1. Перейдите в **Administration → Media types**
2. Создайте новый тип медиа: **Telegram**
3. Укажите скрипт: `telegram.sh`
4. Перейдите в **Administration → Users**
5. Добавьте медиа для пользователя Admin

## 💾 Настройка резервного копирования

### 1. Создание скрипта резервного копирования

```bash
sudo tee /usr/local/bin/zabbix-backup.sh > /dev/null <<EOF
#!/bin/bash

# Скрипт резервного копирования Zabbix
BACKUP_DIR="/var/backups/zabbix"
DATE=\$(date +%Y%m%d_%H%M%S)

# Создание директории для бэкапов
mkdir -p \$BACKUP_DIR

# Резервное копирование базы данных
sudo -u postgres pg_dump zabbix > \$BACKUP_DIR/zabbix_db_\$DATE.sql

# Резервное копирование конфигурации
tar -czf \$BACKUP_DIR/zabbix_config_\$DATE.tar.gz /etc/zabbix/

# Удаление старых бэкапов (старше 30 дней)
find \$BACKUP_DIR -name "*.sql" -mtime +30 -delete
find \$BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: \$DATE"
EOF

sudo chmod +x /usr/local/bin/zabbix-backup.sh
```

### 2. Настройка автоматического резервного копирования

```bash
# Добавление задачи в cron для ежедневного бэкапа
(crontab -l 2>/dev/null; echo "0 2 * * * /usr/local/bin/zabbix-backup.sh") | crontab -
```

## 🔧 Полезные команды

### Управление сервисами:
```bash
# Перезапуск сервисов
sudo systemctl restart zabbix-server
sudo systemctl restart zabbix-agent
sudo systemctl restart postgresql
sudo systemctl restart apache2

# Проверка статуса
sudo systemctl status zabbix-server
sudo systemctl status zabbix-agent
```

### Работа с логами:
```bash
# Просмотр логов в реальном времени
sudo journalctl -u zabbix-server -f
sudo journalctl -u zabbix-agent -f

# Просмотр логов файлов
sudo tail -f /var/log/zabbix/zabbix_server.log
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

### Проверка конфигурации:
```bash
# Проверка конфигурации Zabbix Server
sudo -u zabbix /usr/sbin/zabbix_server -c /etc/zabbix/zabbix_server.conf

# Проверка конфигурации Zabbix Agent
sudo -u zabbix /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf
```

### Резервное копирование:
```bash
# Ручной бэкап
sudo /usr/local/bin/zabbix-backup.sh

# Восстановление из бэкапа
sudo -u postgres psql zabbix < /var/backups/zabbix/zabbix_db_YYYYMMDD_HHMMSS.sql
```

## 🚨 Устранение неполадок

### Проблема: Zabbix Server не запускается

**Диагностика:**
```bash
sudo systemctl status zabbix-server
sudo journalctl -u zabbix-server --no-pager -l
```

**Возможные решения:**
1. Проверьте права доступа к конфигурационному файлу:
   ```bash
   sudo chown zabbix:zabbix /etc/zabbix/zabbix_server.conf
   sudo chmod 640 /etc/zabbix/zabbix_server.conf
   ```

2. Проверьте подключение к базе данных:
   ```bash
   sudo -u zabbix psql -h localhost -U zabbix -d zabbix -c "SELECT version();"
   ```

3. Создайте необходимые директории:
   ```bash
   sudo mkdir -p /var/log/zabbix /run/zabbix
   sudo chown zabbix:zabbix /var/log/zabbix /run/zabbix
   ```

### Проблема: Веб-интерфейс недоступен

**Диагностика:**
```bash
sudo systemctl status apache2
sudo netstat -tlnp | grep :80
curl -I http://10.0.0.241/zabbix
```

**Возможные решения:**
1. Перезапустите Apache:
   ```bash
   sudo systemctl restart apache2
   ```

2. Проверьте конфигурацию сайта:
   ```bash
   sudo apache2ctl configtest
   ```

3. Проверьте права доступа к директории Zabbix:
   ```bash
   sudo chown -R www-data:www-data /usr/share/zabbix
   ```

### Проблема: Агент не отправляет данные

**Диагностика:**
```bash
sudo systemctl status zabbix-agent
sudo netstat -tlnp | grep :10050
```

**Возможные решения:**
1. Проверьте конфигурацию агента:
   ```bash
   sudo -u zabbix /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf
   ```

2. Проверьте файрвол:
   ```bash
   sudo ufw status
   sudo ufw allow 10050/tcp
   ```

## 📈 Мониторинг производительности

### Ключевые метрики для отслеживания:

1. **Системные ресурсы:**
   - CPU usage
   - Memory usage
   - Disk usage
   - Network traffic

2. **Zabbix производительность:**
   - Количество активных проверок
   - Время отклика базы данных
   - Размер кэша
   - Количество триггеров

3. **CloudBridge специфичные метрики:**
   - Количество активных соединений
   - Время отклика API
   - Статус AI сервисов
   - Пропускная способность

## 🔒 Безопасность

### Рекомендации по безопасности:

1. **Измените пароли по умолчанию:**
   - Пароль пользователя Admin в Zabbix
   - Пароль базы данных PostgreSQL

2. **Настройте SSL/TLS:**
   ```bash
   sudo apt install -y certbot python3-certbot-apache
   sudo certbot --apache -d 10.0.0.241
   ```

3. **Ограничьте доступ к файрволу:**
   ```bash
   sudo ufw deny 10051/tcp  # Закрыть порт сервера для внешнего доступа
   ```

4. **Регулярно обновляйте систему:**
   ```bash
   sudo apt update && sudo apt upgrade -y
   ```

## 📞 Поддержка и контакты

### Ответственные:
- **DevOps Lead**: devops-lead@cloudbridge.com
- **System Administrator**: sysadmin@cloudbridge.com
- **AI Team**: ai-team@cloudbridge.com

### Документация:
- [Zabbix Official Documentation](https://www.zabbix.com/documentation)
- [CloudBridge Monitoring Guide](AI_MONITORING_SETUP.md)
- [Troubleshooting Guide](AI_TROUBLESHOOTING.md)

### Полезные ссылки:
- [Zabbix Community](https://www.zabbix.com/community)
- [Zabbix Templates](https://share.zabbix.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

## 🎯 Заключение

После выполнения всех шагов у вас будет полностью настроенная система мониторинга Zabbix, которая обеспечит:

1. **Полный мониторинг** CloudBridge Relay системы
2. **Автоматические алерты** при возникновении проблем
3. **Веб-интерфейс** для просмотра метрик и управления
4. **Масштабируемость** для будущего роста системы
5. **Надежность** с резервным копированием и восстановлением

Система готова к использованию в продакшене и может быть интегрирована с существующими системами мониторинга.




