#!/usr/bin/env bash

# Настройка переменных окружения для VK Cloud
export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="8530029fa7b64199b2ad87c18d4620df"
export OS_REGION_NAME="RegionOne"
export OS_USERNAME="info@2gc.ru"
export OS_USER_DOMAIN_NAME="users"
export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

# Токен для API
export VK_TOKEN="gAAAAABonbSaPNEzCT9R3_qbqtO_XQuDGhO3-DzJEoBhvLJNOxcYnd8tSTV7ao99HP6t9r756YJx4qfJc66ECzXwKPnoM4aklMwLgEe6AYqC3HLUp4AxQfuShZIpq3mruYyYEIZCVwtT2EEDlThP4Gl81lINZpRxWnqOEIzf0rztwq5XrV0NkBgxmL2wE9xQKnf7tY-JHLX8"

# Устанавливаем токен как OS_TOKEN для использования в скриптах
export OS_TOKEN=$VK_TOKEN

# Пароль для аутентификации
export OS_PASSWORD="7Thbngy6"

# Эндпоинты API
export NEUTRON_URL="https://infra.mail.ru:9696"
export OCTAVIA_URL="https://public.infra.mail.ru:9876"

echo "Переменные окружения настроены для проекта $OS_PROJECT_ID"
echo "Neutron API: $NEUTRON_URL"
echo "Octavia API: $OCTAVIA_URL"
