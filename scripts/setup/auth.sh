#!/bin/bash
export OS_AUTH_URL="https://infra.mail.ru:35357/v3/"
export OS_USERNAME="info@2gc.ru"
export OS_PROJECT_ID="8530029fa7b64199b2ad87c18d4620df"
export OS_REGION_NAME="RegionOne"
export OS_INTERFACE="public"
export OS_IDENTITY_API_VERSION=3
export OS_AUTH_TYPE="password"
export OS_USER_DOMAIN_NAME="users"
export OS_PROJECT_DOMAIN_NAME="users"

echo "Введите ваш пароль для VK Cloud:"
read -sr OS_PASSWORD_INPUT
export OS_PASSWORD=$OS_PASSWORD_INPUT

echo "Проверка аутентификации..."
openstack token issue
