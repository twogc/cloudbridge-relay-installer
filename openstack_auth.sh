#!/usr/bin/env bash

# Скрипт для настройки переменных окружения OpenStack
# Основан на документации VK Cloud и предоставленных данных

export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_PROJECT_ID="8530029fa7b64199b2ad87c18d4620df"
export OS_REGION_NAME="RegionOne"
unset OS_PROJECT_NAME
unset OS_PROJECT_DOMAIN_ID

# unset v2.0 items in case set
unset OS_TENANT_ID
unset OS_TENANT_NAME

if [[ -z $OS_USERNAME ]] || [[ -z $OS_PASSWORD ]] || [[ "$OS_USERNAME" != "info@2gc.ru" ]]; then
    export OS_USERNAME="info@2gc.ru"
    export OS_USER_DOMAIN_NAME="users"
    
    # Используем предоставленный пароль
    export OS_PASSWORD="7Thbngy6"
fi  

export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

echo "Переменные окружения OpenStack настроены:"
echo "OS_AUTH_URL: $OS_AUTH_URL"
echo "OS_PROJECT_ID: $OS_PROJECT_ID"
echo "OS_REGION_NAME: $OS_REGION_NAME"
echo "OS_USERNAME: $OS_USERNAME"
echo "OS_USER_DOMAIN_NAME: $OS_USER_DOMAIN_NAME"
echo "OS_INTERFACE: $OS_INTERFACE"
echo "OS_IDENTITY_API_VERSION: $OS_IDENTITY_API_VERSION"
echo ""
echo "Для применения переменных выполните: source openstack_auth.sh"

