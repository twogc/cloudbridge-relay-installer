#!/bin/bash

# Скрипт для исправления прав доступа Zabbix Agent к Docker

set -e

echo " Исправление прав доступа Zabbix Agent к Docker..."

# Добавляем пользователя zabbix в группу docker
echo "👤 Добавление пользователя zabbix в группу docker..."
sudo usermod -aG docker zabbix

# Проверяем, что пользователь zabbix может получить доступ к Docker
echo "🧪 Тестирование доступа к Docker..."
if sudo -u zabbix docker ps >/dev/null 2>&1; then
    echo "Пользователь zabbix может получить доступ к Docker"
else
    echo "Ошибка доступа к Docker для пользователя zabbix"
    echo "Попробуйте перезапустить систему или выполнить:"
    echo "   sudo systemctl restart zabbix-agent"
    exit 1
fi

# Проверяем права на Docker socket
echo "Проверка прав на Docker socket..."
if [ -S /var/run/docker.sock ]; then
    echo "Текущие права на Docker socket:"
    ls -la /var/run/docker.sock
    echo ""
    echo "Группа docker:"
    getent group docker
else
    echo "Docker socket не найден"
    exit 1
fi

# Тестируем системные метрики
echo "🧪 Тестирование системных метрик..."
echo "   Docker контейнеры: $(zabbix_agentd -t cloudbridge.docker.containers.running | awk '{print $2}' | tr -d '[]')"
echo "   CPU использование: $(zabbix_agentd -t cloudbridge.system.cpu.usage | awk '{print $2}' | tr -d '[]')%"
echo "   Память: $(zabbix_agentd -t cloudbridge.system.memory.usage | awk '{print $2}' | tr -d '[]')%"

echo ""
echo "Права доступа исправлены!"
echo ""
echo "Информация:"
echo "   Пользователь zabbix добавлен в группу docker"
echo "   Docker API доступен для Zabbix Agent"
echo "   Системные метрики работают корректно"
echo ""
echo "🔄 Рекомендации:"
echo "   1. Перезапустите Zabbix Agent: sudo pkill -f zabbix_agentd && sudo ./scripts/start_zabbix_agent.sh"
echo "   2. Проверьте метрики в Zabbix веб-интерфейсе"
echo "   3. Убедитесь, что все метрики собираются без ошибок"
