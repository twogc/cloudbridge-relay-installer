#!/usr/bin/env bash

# VK Cloud OpenStack credentials
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
    
    # With Keystone you pass the keystone password.
    echo "Please enter your OpenStack Password for project $OS_PROJECT_ID as user $OS_USERNAME: "
    read -sr OS_PASSWORD_INPUT
    export OS_PASSWORD=$OS_PASSWORD_INPUT
fi  

export OS_INTERFACE=public
export OS_IDENTITY_API_VERSION=3

echo "OpenStack environment configured for VK Cloud"
echo "Project ID: $OS_PROJECT_ID"
echo "Username: $OS_USERNAME"
echo "Auth URL: $OS_AUTH_URL"
