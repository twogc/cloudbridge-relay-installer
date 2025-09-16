#!/bin/bash

# Скрипт для настройки Zabbix Agent для подключения к серверу Zabbix

set -e

ZABBIX_SERVER_IP="${1:-10.0.0.241}"
HOSTNAME="${2:-cloudbridge-relay-01}"
ZABBIX_AGENT_CONFIG="/etc/zabbix/zabbix_agentd.conf"

echo "🔧 Настройка Zabbix Agent для подключения к серверу $ZABBIX_SERVER_IP..."

# Проверяем доступность Zabbix сервера
echo "📡 Проверка доступности Zabbix сервера..."
if ping -c 1 -W 5 "$ZABBIX_SERVER_IP" > /dev/null 2>&1; then
    echo "✅ Zabbix сервер $ZABBIX_SERVER_IP доступен"
else
    echo "❌ Zabbix сервер $ZABBIX_SERVER_IP недоступен"
    echo "   Проверьте сетевую связность и настройки файрвола"
    exit 1
fi

# Проверяем порт Zabbix Agent
echo "🔌 Проверка порта Zabbix Agent (10050)..."
if timeout 5 bash -c "</dev/tcp/$ZABBIX_SERVER_IP/10050" 2>/dev/null; then
    echo "✅ Порт 10050 на $ZABBIX_SERVER_IP открыт"
else
    echo "❌ Порт 10050 на $ZABBIX_SERVER_IP недоступен"
    echo "   Убедитесь, что Zabbix Server запущен и слушает на порту 10050"
    exit 1
fi

# Создаем резервную копию конфигурации
echo "💾 Создание резервной копии конфигурации..."
sudo cp "$ZABBIX_AGENT_CONFIG" "${ZABBIX_AGENT_CONFIG}.backup.$(date +%Y%m%d_%H%M%S)"

# Обновляем конфигурацию
echo "⚙️ Обновление конфигурации Zabbix Agent..."

# Обновляем Server
sudo sed -i "s/^Server=.*/Server=$ZABBIX_SERVER_IP/" "$ZABBIX_AGENT_CONFIG"

# Обновляем ServerActive
sudo sed -i "s/^ServerActive=.*/ServerActive=$ZABBIX_SERVER_IP/" "$ZABBIX_AGENT_CONFIG"

# Обновляем Hostname
sudo sed -i "s/^Hostname=.*/Hostname=$HOSTNAME/" "$ZABBIX_AGENT_CONFIG"

# Проверяем обновленную конфигурацию
echo "📋 Проверка обновленной конфигурации..."
echo "   Server: $(grep '^Server=' "$ZABBIX_AGENT_CONFIG" | cut -d'=' -f2)"
echo "   ServerActive: $(grep '^ServerActive=' "$ZABBIX_AGENT_CONFIG" | cut -d'=' -f2)"
echo "   Hostname: $(grep '^Hostname=' "$ZABBIX_AGENT_CONFIG" | cut -d'=' -f2)"

# Останавливаем текущий процесс Zabbix Agent
echo "🛑 Остановка текущего процесса Zabbix Agent..."
sudo pkill -f zabbix_agentd || true
sleep 2

# Запускаем Zabbix Agent как сервис
echo "🚀 Запуск Zabbix Agent как сервис..."
sudo systemctl start zabbix-agent

# Проверяем статус сервиса
echo "📊 Проверка статуса Zabbix Agent..."
if sudo systemctl is-active --quiet zabbix-agent; then
    echo "✅ Zabbix Agent успешно запущен"
    sudo systemctl status zabbix-agent --no-pager
else
    echo "❌ Ошибка запуска Zabbix Agent"
    sudo systemctl status zabbix-agent --no-pager
    echo "📋 Логи Zabbix Agent:"
    sudo journalctl -u zabbix-agent --no-pager -n 20
    exit 1
fi

# Тестируем AI метрики
echo "🧪 Тестирование AI метрик..."
echo "   AI Status: $(zabbix_agentd -t ai.status | awk '{print $2}' | tr -d '[]')"
echo "   AI Health: $(zabbix_agentd -t ai.health | awk '{print $2}' | tr -d '[]')"
echo "   CPU Usage: $(zabbix_agentd -t cloudbridge.system.cpu.usage | awk '{print $2}' | tr -d '[]')"
echo "   Memory Usage: $(zabbix_agentd -t cloudbridge.system.memory.usage | awk '{print $2}' | tr -d '[]')"

echo ""
echo "✅ Настройка Zabbix Agent завершена!"
echo ""
echo "📋 Следующие шаги:"
echo "1. В веб-интерфейсе Zabbix создайте хост с именем: $HOSTNAME"
echo "2. Укажите IP адрес: $(hostname -I | awk '{print $1}')"
echo "3. Импортируйте шаблон: scripts/cloudbridge_ai_zabbix_template_working.xml"
echo "4. Примените шаблон к хосту"
echo "5. Проверьте сбор данных в Monitoring → Latest data"
echo ""
echo "🔧 Полезные команды:"
echo "   sudo systemctl status zabbix-agent  # Статус агента"
echo "   sudo systemctl restart zabbix-agent  # Перезапуск агента"
echo "   sudo tail -f /var/log/zabbix/zabbix_agentd.log  # Логи агента"
echo "   zabbix_agentd -t ai.status  # Тест AI метрики"

