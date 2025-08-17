#!/usr/bin/env bash

# Настройка переменных окружения для VK Cloud
export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="8530029fa7b64199b2ad87c18d4620df"
export OS_REGION_NAME="RegionOne"
export OS_USERNAME="info@2gc.ru"
export OS_USER_DOMAIN_NAME="users"
export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

# Пароль для аутентификации
export OS_PASSWORD="7Thbngy6"

# Получаем новый токен
echo "Получаем новый токен аутентификации..."
NEW_TOKEN=$(curl -s -D - -X POST -H "Content-Type: application/json" -d "{\"auth\": {\"identity\": {\"methods\": [\"password\"], \"password\": {\"user\": {\"name\": \"$OS_USERNAME\", \"password\": \"$OS_PASSWORD\", \"domain\": {\"name\": \"$OS_USER_DOMAIN_NAME\"}}}}, \"scope\": {\"project\": {\"id\": \"$OS_PROJECT_ID\"}}}}" "$OS_AUTH_URL/auth/tokens" 2>/dev/null | grep -i "x-subject-token" | cut -d' ' -f2 | tr -d '\r')

if [ -n "$NEW_TOKEN" ]; then
    export OS_TOKEN="$NEW_TOKEN"
    echo "Токен получен успешно: ${OS_TOKEN:0:20}..."
else
    echo "Ошибка получения токена"
    exit 1
fi

# Эндпоинты API
export NEUTRON_URL="https://infra.mail.ru:9696"
export OCTAVIA_URL="https://public.infra.mail.ru:9876"

echo "Переменные окружения настроены для проекта $OS_PROJECT_ID"
echo "Neutron API: $NEUTRON_URL"
echo "Octavia API: $OCTAVIA_URL"
