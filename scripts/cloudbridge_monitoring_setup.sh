#!/bin/bash

# Настройка мониторинга CloudBridge в Zabbix
# Выполнить после установки Zabbix

set -e

echo " Настройка мониторинга CloudBridge..."

# Создание директории для кастомных шаблонов
sudo mkdir -p /etc/zabbix/templates/cloudbridge

# Создание шаблона для CloudBridge Relay
echo "Создание шаблона CloudBridge..."
sudo tee /etc/zabbix/templates/cloudbridge/cloudbridge-relay.xml > /dev/null <<'EOF'
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
                <item>
                    <uuid>cloudbridge-network-uuid</uuid>
                    <name>Network Traffic In</name>
                    <type>0</type>
                    <key>net.if.in[eth0]</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>3</value_type>
                    <units>bps</units>
                </item>
                <item>
                    <uuid>cloudbridge-network-out-uuid</uuid>
                    <name>Network Traffic Out</name>
                    <type>0</type>
                    <key>net.if.out[eth0]</key>
                    <history>7d</history>
                    <trends>365d</trends>
                    <value_type>3</value_type>
                    <units>bps</units>
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

# Создание скрипта для Telegram уведомлений
echo "Настройка Telegram уведомлений..."
sudo mkdir -p /usr/lib/zabbix/alertscripts

sudo tee /usr/lib/zabbix/alertscripts/telegram.sh > /dev/null <<'EOF'
#!/bin/bash

# Конфигурация Telegram
BOT_TOKEN="YOUR_BOT_TOKEN"
CHAT_ID="YOUR_CHAT_ID"

# Получение параметров
TO="$1"
SUBJECT="$2"
MESSAGE="$3"

# Формирование сообщения
TEXT="*Zabbix Alert*
*Subject:* $SUBJECT
📝 *Message:* $MESSAGE
🕐 *Time:* $(date)
 *Server:* 10.0.0.241"

# Отправка в Telegram
curl -s -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
     -H "Content-Type: application/json" \
     -d "{\"chat_id\":\"$CHAT_ID\",\"text\":\"$TEXT\",\"parse_mode\":\"Markdown\"}"

echo "Telegram notification sent"
EOF

sudo chmod +x /usr/lib/zabbix/alertscripts/telegram.sh

# Создание скрипта резервного копирования
echo "💾 Создание скрипта резервного копирования..."
sudo tee /usr/local/bin/zabbix-backup.sh > /dev/null <<'EOF'
#!/bin/bash

# Скрипт резервного копирования Zabbix
BACKUP_DIR="/var/backups/zabbix"
DATE=$(date +%Y%m%d_%H%M%S)

# Создание директории для бэкапов
mkdir -p $BACKUP_DIR

# Резервное копирование базы данных
sudo -u postgres pg_dump zabbix > $BACKUP_DIR/zabbix_db_$DATE.sql

# Резервное копирование конфигурации
tar -czf $BACKUP_DIR/zabbix_config_$DATE.tar.gz /etc/zabbix/

# Удаление старых бэкапов (старше 30 дней)
find $BACKUP_DIR -name "*.sql" -mtime +30 -delete
find $BACKUP_DIR -name "*.tar.gz" -mtime +30 -delete

echo "Backup completed: $DATE"
EOF

sudo chmod +x /usr/local/bin/zabbix-backup.sh

# Добавление задачи в cron для ежедневного бэкапа
echo "⏰ Настройка автоматического резервного копирования..."
(crontab -l 2>/dev/null; echo "0 2 * * * /usr/local/bin/zabbix-backup.sh") | crontab -

echo ""
echo "Настройка мониторинга CloudBridge завершена!"
echo ""
echo "Следующие шаги:"
echo "1. Откройте веб-интерфейс Zabbix: http://10.0.0.241/zabbix"
echo "2. Войдите с учетными данными: Admin / zabbix"
echo "3. Импортируйте шаблон: Configuration → Templates → Import"
echo "4. Создайте хост для CloudBridge сервера"
echo "5. Настройте Telegram уведомления (укажите BOT_TOKEN и CHAT_ID)"
echo ""
echo " Полезные команды:"
echo "   sudo /usr/local/bin/zabbix-backup.sh  # Ручной бэкап"
echo "   sudo tail -f /var/log/zabbix/zabbix_server.log  # Логи сервера"
echo "   sudo systemctl restart zabbix-server  # Перезапуск сервера"




