#!/bin/bash

# Полная установка Zabbix Server на Ubuntu 22.04
# Выполнить на сервере 10.0.0.241

set -e

echo "🚀 Начинаем установку Zabbix Server..."

# Обновление системы
echo "📦 Обновление системы..."
sudo apt update && sudo apt upgrade -y

# Установка необходимых пакетов
echo "📦 Установка необходимых пакетов..."
sudo apt install -y wget curl gnupg2 software-properties-common apt-transport-https ca-certificates lsb-release

# Настройка часового пояса
echo "🕐 Настройка часового пояса..."
sudo timedatectl set-timezone Europe/Moscow

# Установка PostgreSQL
echo "🐘 Установка PostgreSQL..."
sudo apt install -y postgresql postgresql-contrib

# Запуск и включение PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Создание пользователя и базы данных для Zabbix
echo "🗄️ Настройка базы данных..."
sudo -u postgres psql -c "CREATE USER zabbix WITH PASSWORD 'zabbix_password_2024';"
sudo -u postgres psql -c "CREATE DATABASE zabbix OWNER zabbix;"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE zabbix TO zabbix;"
sudo -u postgres psql -c "ALTER USER zabbix CREATEDB;"

# Добавление репозитория Zabbix
echo "📥 Добавление репозитория Zabbix..."
wget https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
sudo apt update

# Установка Zabbix Server и компонентов
echo "📦 Установка Zabbix Server..."
sudo apt install -y zabbix-server-pgsql zabbix-frontend-php php8.1-pgsql zabbix-sql-scripts

# Установка Apache и PHP
echo "🌐 Установка Apache и PHP..."
sudo apt install -y apache2 php8.1 libapache2-mod-php8.1

# Импорт схемы базы данных
echo "📊 Импорт схемы базы данных..."
sudo zcat /usr/share/zabbix-sql-scripts/postgresql/server.sql.gz | sudo -u zabbix psql zabbix

# Создание директорий для логов и PID
echo "📁 Создание необходимых директорий..."
sudo mkdir -p /var/log/zabbix /run/zabbix
sudo chown zabbix:zabbix /var/log/zabbix /run/zabbix

# Настройка конфигурации Zabbix Server
echo "⚙️ Настройка конфигурации Zabbix Server..."
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

# Установка правильных прав на конфигурацию
sudo chown zabbix:zabbix /etc/zabbix/zabbix_server.conf
sudo chmod 640 /etc/zabbix/zabbix_server.conf

# Настройка PHP
echo "🐘 Настройка PHP..."
sudo sed -i 's/;date.timezone =/date.timezone = Europe\/Moscow/' /etc/php/8.1/apache2/php.ini
sudo sed -i 's/max_execution_time = 30/max_execution_time = 300/' /etc/php/8.1/apache2/php.ini
sudo sed -i 's/memory_limit = 128M/memory_limit = 256M/' /etc/php/8.1/apache2/php.ini

# Настройка Apache для Zabbix
echo "🌐 Настройка Apache..."
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

# Включение сайта Zabbix
sudo a2ensite zabbix.conf
sudo a2enmod rewrite
sudo systemctl restart apache2
sudo systemctl enable apache2

# Запуск Zabbix Server
echo "🚀 Запуск Zabbix Server..."
sudo systemctl start zabbix-server
sudo systemctl enable zabbix-server

# Установка Zabbix Agent
echo "📡 Установка Zabbix Agent..."
sudo apt install -y zabbix-agent

# Настройка Zabbix Agent
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

# Настройка файрвола
echo "🔥 Настройка файрвола..."
sudo apt install -y ufw
sudo ufw allow ssh
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow 10050/tcp
sudo ufw allow 10051/tcp
sudo ufw --force enable

# Проверка статуса сервисов
echo "✅ Проверка статуса сервисов..."
echo "PostgreSQL:"
sudo systemctl status postgresql --no-pager -l
echo ""
echo "Zabbix Server:"
sudo systemctl status zabbix-server --no-pager -l
echo ""
echo "Apache:"
sudo systemctl status apache2 --no-pager -l
echo ""
echo "Zabbix Agent:"
sudo systemctl status zabbix-agent --no-pager -l

# Проверка портов
echo "🔍 Проверка портов..."
sudo netstat -tlnp | grep -E "(80|443|10050|10051|5432)"

echo ""
echo "🎉 Установка Zabbix завершена!"
echo ""
echo "📋 Информация для доступа:"
echo "   Веб-интерфейс: http://10.0.0.241/zabbix"
echo "   Логин по умолчанию: Admin"
echo "   Пароль по умолчанию: zabbix"
echo ""
echo "🔧 Полезные команды:"
echo "   sudo systemctl status zabbix-server"
echo "   sudo systemctl restart zabbix-server"
echo "   sudo tail -f /var/log/zabbix/zabbix_server.log"
echo ""
echo "⚠️  Не забудьте изменить пароль по умолчанию после первого входа!"




