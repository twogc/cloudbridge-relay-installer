#!/bin/bash

# Пример скрипта для получения токенов авторизации
# ЗАМЕНИТЕ YOUR_CLIENT_SECRET_HERE на реальный client_secret

KEYCLOAK_URL="https://89.208.86.150:9443"
REALM="cloudbridge"
CLIENT_ID="test-client"
CLIENT_SECRET="YOUR_CLIENT_SECRET_HERE"  # ЗАМЕНИТЕ НА РЕАЛЬНЫЙ SECRET
USERNAME="info@2gc.ru"
PASSWORD="admin"

echo "=== Получение токенов авторизации ==="
echo "URL: $KEYCLOAK_URL"
echo "Realm: $REALM"
echo "Client ID: $CLIENT_ID"
echo "Username: $USERNAME"
echo ""

# Проверка, что client_secret заменен
if [ "$CLIENT_SECRET" = "YOUR_CLIENT_SECRET_HERE" ]; then
    echo "ОШИБКА: Замените YOUR_CLIENT_SECRET_HERE на реальный client_secret"
    echo "Получите client_secret в Admin Console: $KEYCLOAK_URL/admin/"
    echo "Realm: $REALM -> Clients -> $CLIENT_ID -> Credentials"
    exit 1
fi

echo "1. Получение OAuth Token (Password Grant)..."
TOKEN_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" \
    -d "username=$USERNAME" \
    -d "password=$PASSWORD")

if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
    REFRESH_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.refresh_token')
    EXPIRES_IN=$(echo "$TOKEN_RESPONSE" | jq -r '.expires_in')
    
    echo "OAuth Token получен успешно!"
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
    echo "Expires in: $EXPIRES_IN секунд"
    echo ""
    echo "Полный ответ:"
    echo "$TOKEN_RESPONSE" | jq .
else
    echo "Ошибка получения OAuth Token:"
    echo "$TOKEN_RESPONSE"
fi

echo ""
echo "2. Получение API Key (Client Credentials)..."
API_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
    -d "grant_type=client_credentials" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET")

if echo "$API_RESPONSE" | grep -q "access_token"; then
    API_TOKEN=$(echo "$API_RESPONSE" | jq -r '.access_token')
    echo "API Key получен успешно!"
    echo "API Token: ${API_TOKEN:0:50}..."
    echo ""
    echo "Полный ответ:"
    echo "$API_RESPONSE" | jq .
else
    echo "Ошибка получения API Key:"
    echo "$API_RESPONSE"
fi

echo ""
echo "=== Проверка токенов ==="

if [ -n "$ACCESS_TOKEN" ]; then
    echo "3. Проверка OAuth Token..."
    USERINFO_RESPONSE=$(curl -s -k -H "Authorization: Bearer $ACCESS_TOKEN" \
        "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo")
    
    if echo "$USERINFO_RESPONSE" | grep -q "sub"; then
        echo "OAuth Token валиден!"
        echo "User Info:"
        echo "$USERINFO_RESPONSE" | jq .
    else
        echo "OAuth Token невалиден:"
        echo "$USERINFO_RESPONSE"
    fi
fi

echo ""
echo "=== Для Unified Authentication Test ==="
echo ""
echo "OAuth Token (вставьте в поле 'OAuth Token'):"
if [ -n "$ACCESS_TOKEN" ]; then
    echo "$ACCESS_TOKEN"
else
    echo "Токен не получен"
fi
echo ""
echo "API Key (вставьте в поле 'API Key'):"
echo "$CLIENT_SECRET"
echo ""
echo "URL для тестирования:"
echo "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token"
echo ""
echo "=== Примеры использования ==="
echo ""
echo "# Проверка токена"
echo "curl -k -H 'Authorization: Bearer $ACCESS_TOKEN' \\"
echo "  '$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo'"
echo ""
echo "# Обновление токена"
echo "curl -k -X POST '$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token' \\"
echo "  -d 'grant_type=refresh_token' \\"
echo "  -d 'client_id=$CLIENT_ID' \\"
echo "  -d 'client_secret=$CLIENT_SECRET' \\"
echo "  -d 'refresh_token=$REFRESH_TOKEN'" 