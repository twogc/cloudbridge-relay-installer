#!/bin/bash
set -e

# Генерируем .env файл если его нет
if [ ! -f ".env" ]; then
    echo -e "\033[1;34m[INFO]\033[0m Генерация .env файла..."
    if [ -f "scripts/setup/generate_secrets.sh" ]; then
    bash scripts/setup/generate_secrets.sh --env .env
    else
        echo -e "\033[1;33m[WARNING]\033[0m Скрипт generate_secrets.sh не найден"
    fi
else
    echo -e "\033[1;34m[INFO]\033[0m Файл .env уже существует"
fi

# Проверяем критичные переменные
if [ -f ".env" ]; then
    if ! grep -q "KEYCLOAK_ADMIN_PASSWORD" .env; then
        echo -e "\033[0;31m[ERROR]\033[0m В файле .env отсутствует KEYCLOAK_ADMIN_PASSWORD"
        exit 1
    fi
    echo -e "\033[0;32m[SUCCESS]\033[0m Файл .env проверен"
fi 