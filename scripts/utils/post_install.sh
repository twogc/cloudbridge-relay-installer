#!/bin/bash
set -e

echo -e "\033[1;32m"
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                                                                              ║"
echo "║                    CloudBridge Relay Successfully Installed!                 ║"
echo "║                                                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo -e "\033[0m"

echo -e "\033[1;36mURLs сервисов:\033[0m"
echo -e "  \033[1;36mZitadel:\033[0m     \${ZITADEL_DOMAIN:-https://zitadel.local}"
echo -e "  \033[1;36mRelay:\033[0m        http://localhost:8081"
echo -e "  \033[1;36mGrafana:\033[0m      http://localhost:3000"
echo -e "  \033[1;36mPrometheus:\033[0m   http://localhost:9090"
echo -e "  \033[1;36mAlertManager:\033[0m http://localhost:9093"
echo

echo -e "\033[1;36mПолезные команды:\033[0m"
echo -e "  \033[1;33m./scripts/utils/start_production.sh\033[0m - Запуск production окружения"
echo -e "  \033[1;33mdocker-compose logs -f [service]\033[0m - Просмотр логов"
echo -e "  \033[1;33m./scripts/setup/setup_dashboards.sh\033[0m - Обновление дашбордов"
echo

echo -e "\033[1;36mДокументация:\033[0m"
echo -e "  \033[1;33mopenwiki/quickstart.md\033[0m - Быстрый старт"
echo -e "  \033[1;33mdocs/guides/\033[0m - Руководства по развёртыванию"
echo

echo -e "\033[0;32mУстановка CloudBridge Relay завершена успешно!\033[0m"
