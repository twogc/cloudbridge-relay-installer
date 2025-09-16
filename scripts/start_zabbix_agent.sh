#!/bin/bash

# Скрипт для запуска Zabbix Agent с AI метриками

set -e

echo "🚀 Запуск Zabbix Agent с AI метриками..."

# Останавливаем Zabbix Agent 2 (если запущен)
echo "🛑 Остановка Zabbix Agent 2..."
sudo systemctl stop zabbix-agent2 2>/dev/null || true
sudo systemctl disable zabbix-agent2 2>/dev/null || true

# Останавливаем все процессы zabbix_agentd
echo "🛑 Остановка всех процессов Zabbix Agent..."
sudo pkill -f zabbix_agentd || true
sleep 2

# Проверяем, что порт 10050 свободен
echo "🔌 Проверка порта 10050..."
if sudo netstat -tlnp | grep -q ":10050"; then
    echo "❌ Порт 10050 занят. Останавливаем процесс..."
    sudo fuser -k 10050/tcp 2>/dev/null || true
    sleep 2
fi

# Запускаем Zabbix Agent
echo "🚀 Запуск Zabbix Agent..."
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &
sleep 3

# Проверяем, что Agent запустился
echo "📊 Проверка статуса Zabbix Agent..."
if sudo netstat -tlnp | grep -q "zabbix_agen.*:10050"; then
    echo "✅ Zabbix Agent успешно запущен на порту 10050"
else
    echo "❌ Ошибка запуска Zabbix Agent"
    exit 1
fi

# Тестируем AI метрики
echo "🧪 Тестирование AI метрик..."
echo "   AI Status: $(zabbix_agentd -t ai.status | awk '{print $2}' | tr -d '[]')"
echo "   AI Health: $(zabbix_agentd -t ai.health | awk '{print $2}' | tr -d '[]')"
echo "   CPU Usage: $(zabbix_agentd -t cloudbridge.system.cpu.usage | awk '{print $2}' | tr -d '[]')"

echo ""
echo "✅ Zabbix Agent с AI метриками запущен!"
echo ""
echo "📋 Информация:"
echo "   Порт: 10050"
echo "   Конфигурация: /etc/zabbix/zabbix_agentd.conf"
echo "   Логи: /var/log/zabbix/zabbix_agentd.log"
echo "   Скрипт метрик: /usr/local/bin/ai_metrics_collector.sh"
echo ""
echo "🔧 Полезные команды:"
echo "   zabbix_agentd -t ai.status  # Тест AI метрики"
echo "   sudo pkill -f zabbix_agentd  # Остановка агента"
echo "   sudo tail -f /var/log/zabbix/zabbix_agentd.log  # Просмотр логов"

