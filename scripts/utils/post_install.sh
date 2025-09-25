#!/bin/bash
set -e

echo -e "\033[1;32m"
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                              ║"
echo "║                    🎉 CloudBridge Relay Successfully Installed! 🎉          ║"
echo "║                                                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo -e "\033[0m"

echo -e "\033[1;36m📊 URLs сервисов:\033[0m"
echo -e "  \033[1;36m🔐 Keycloak:\033[0m     http://localhost:8080"
echo -e "  \033[1;36m🚀 Relay:\033[0m        http://localhost:8081"
echo -e "  \033[1;36m📊 Grafana:\033[0m      http://localhost:3000"
echo -e "  \033[1;36m📈 Prometheus:\033[0m   http://localhost:9090"
echo -e "  \033[1;36mAlertManager:\033[0m http://localhost:9093"
echo

echo -e "\033[1;36m🔧 Полезные команды:\033[0m"
echo -e "  \033[1;33m./start_production.sh\033[0m - Запуск production окружения"
echo -e "  \033[1;33mdocker-compose logs -f [service]\033[0m - Просмотр логов"
echo -e "  \033[1;33m./scripts/setup/setup_dashboards.sh\033[0m - Обновление дашбордов"
echo

echo -e "\033[1;36m📚 Документация:\033[0m"
echo -e "  \033[1;33mdocs/deployment_guide.md\033[0m - Руководство по развертыванию"
echo -e "  \033[1;33mdocs/dashboard_guide.md\033[0m - Руководство по дашбордам"
echo

echo -e "\033[0;32mУстановка CloudBridge Relay завершена успешно!\033[0m" 