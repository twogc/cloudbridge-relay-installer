#!/bin/bash
set -e

# CloudBridge Relay - Main Setup Script
# Основной скрипт установки с новой организованной структурой

# Красивый баннер
scripts/setup/show_banner.sh

echo -e "\033[1;36m[1/7]\033[0m Проверка зависимостей..."
scripts/setup/check_dependencies.sh

echo -e "\033[1;36m[2/7]\033[0m Настройка .env и секретов..."
scripts/setup/setup_env.sh

echo -e "\033[1;36m[3/7]\033[0m Интерактивная настройка Keycloak..."
scripts/setup/setup_keycloak.sh

echo -e "\033[1;36m[4/7]\033[0m Настройка сертификатов..."
scripts/setup/setup_certs.sh

echo -e "\033[1;36m[5/7]\033[0m Запуск сервисов..."
scripts/utils/start_services.sh

echo -e "\033[1;36m[6/7]\033[0m Установка Grafana дашбордов..."
scripts/setup/setup_dashboards.sh

echo -e "\033[1;36m[7/7]\033[0m Финальный статус..."
scripts/utils/post_install.sh 