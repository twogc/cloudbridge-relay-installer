#!/bin/bash
# CloudBridge Relay Zabbix Agent Installer
# Автоматическая установка всех компонентов Zabbix мониторинга

set -euo pipefail

echo "🚀 CloudBridge Relay - Zabbix Agent Installer"
echo "============================================="

# Проверка прав root
if [[ $EUID -ne 0 ]]; then
   echo "❌ Этот скрипт должен запускаться с правами root"
   echo "Используйте: sudo $0"
   exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📁 Устанавливаем конфигурацию Zabbix Agent..."
cp "$SCRIPT_DIR/conf.d/cloudbridge-relay-template.conf" /etc/zabbix/zabbix_agentd.d/
echo "✅ Конфигурация установлена"

echo "🔧 Устанавливаем коллекторы..."
cp "$SCRIPT_DIR/collectors"/*.sh /usr/local/bin/
chmod +x /usr/local/bin/{zitadel_collector.sh,p2p_metrics_collector.sh,ai_extended_collector.sh}
echo "✅ Коллекторы установлены"

echo "🔄 Перезапускаем Zabbix Agent..."
systemctl restart zabbix-agent
systemctl enable zabbix-agent
echo "✅ Zabbix Agent перезапущен"

echo "🧪 Тестируем установку..."
echo "Status Zitadel: $(zabbix_get -s localhost -k zitadel.status 2>/dev/null || echo 'ОШИБКА')"
echo "Users Zitadel: $(zabbix_get -s localhost -k zitadel.users.total 2>/dev/null || echo 'ОШИБКА')"
echo "AI Load: $(zabbix_get -s localhost -k ai.load_predictions 2>/dev/null || echo 'ОШИБКА')"

echo ""
echo "🎊 УСТАНОВКА ЗАВЕРШЕНА!"
echo "======================="
echo "📄 Template для импорта: $SCRIPT_DIR/templates/cloudbridge-relay-with-zitadel.xml"
echo "📚 Документация: $SCRIPT_DIR/README.md"
echo "🔧 Логи агента: tail -f /var/log/zabbix/zabbix_agentd.log"
