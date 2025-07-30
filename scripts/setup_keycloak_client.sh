#!/bin/bash

# Скрипт для настройки клиента и пользователя в Keycloak
# Этот скрипт создает тестового пользователя и клиента для получения токенов

KEYCLOAK_URL="https://localhost:9443"
REALM="cloudbridge"
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin"

echo "=== Настройка Keycloak для получения токенов ==="
echo "URL: $KEYCLOAK_URL"
echo "Realm: $REALM"
echo ""

# 1. Попытка получить токен администратора
echo "1. Получение токена администратора..."
ADMIN_TOKEN_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" \
    -d "username=$ADMIN_USERNAME" \
    -d "password=$ADMIN_PASSWORD")

if echo "$ADMIN_TOKEN_RESPONSE" | grep -q "access_token"; then
    ADMIN_TOKEN=$(echo "$ADMIN_TOKEN_RESPONSE" | jq -r '.access_token')
    echo "✅ Токен администратора получен"
else
    echo "❌ Не удалось получить токен администратора:"
    echo "$ADMIN_TOKEN_RESPONSE"
    echo ""
    echo "Попробуйте войти в Admin Console: $KEYCLOAK_URL/admin/"
    echo "И проверить учетные данные администратора"
    exit 1
fi

echo ""
echo "2. Создание тестового пользователя..."
# Создаем тестового пользователя
USER_DATA='{
  "username": "testuser",
  "email": "test@example.com",
  "enabled": true,
  "credentials": [
    {
      "type": "password",
      "value": "testpass123",
      "temporary": false
    }
  ]
}'

USER_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/admin/realms/$REALM/users" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$USER_DATA")

if [ $? -eq 0 ]; then
    echo "✅ Тестовый пользователь создан"
    echo "Username: testuser"
    echo "Password: testpass123"
else
    echo "❌ Ошибка создания пользователя:"
    echo "$USER_RESPONSE"
fi

echo ""
echo "3. Создание тестового клиента..."
# Создаем тестовый клиент
CLIENT_DATA='{
  "clientId": "test-client",
  "enabled": true,
  "publicClient": false,
  "standardFlowEnabled": true,
  "directAccessGrantsEnabled": true,
  "serviceAccountsEnabled": true,
  "redirectUris": ["http://localhost:8080/*"],
  "webOrigins": ["http://localhost:8080"]
}'

CLIENT_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/admin/realms/$REALM/clients" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$CLIENT_DATA")

if [ $? -eq 0 ]; then
    echo "✅ Тестовый клиент создан"
    echo "Client ID: test-client"
    
    # Получаем client secret
    CLIENT_ID=$(curl -s -k -X GET "$KEYCLOAK_URL/admin/realms/$REALM/clients?clientId=test-client" \
        -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')
    
    CLIENT_SECRET_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/admin/realms/$REALM/clients/$CLIENT_ID/client-secret" \
        -H "Authorization: Bearer $ADMIN_TOKEN")
    
    CLIENT_SECRET=$(echo "$CLIENT_SECRET_RESPONSE" | jq -r '.value')
    echo "Client Secret: $CLIENT_SECRET"
else
    echo "❌ Ошибка создания клиента:"
    echo "$CLIENT_RESPONSE"
fi

echo ""
echo "=== Тестирование получения токенов ==="

# Тест получения токена через password grant
echo "4. Тест получения токена через password grant..."
TOKEN_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=test-client" \
    -d "client_secret=$CLIENT_SECRET" \
    -d "username=testuser" \
    -d "password=testpass123")

if echo "$TOKEN_RESPONSE" | grep -q "access_token"; then
    ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
    echo "✅ Токен получен успешно!"
    echo "Access Token: ${ACCESS_TOKEN:0:50}..."
    echo ""
    echo "Полный ответ:"
    echo "$TOKEN_RESPONSE" | jq .
else
    echo "❌ Ошибка получения токена:"
    echo "$TOKEN_RESPONSE"
fi

echo ""
echo "=== Инструкции по использованию ==="
echo "1. Admin Console: $KEYCLOAK_URL/admin/"
echo "2. Тестовый пользователь: testuser / testpass123"
echo "3. Тестовый клиент: test-client"
echo "4. Client Secret: $CLIENT_SECRET"
echo ""
echo "Для получения токена используйте:"
echo "curl -k -X POST '$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token' \\"
echo "  -d 'grant_type=password' \\"
echo "  -d 'client_id=test-client' \\"
echo "  -d 'client_secret=$CLIENT_SECRET' \\"
echo "  -d 'username=testuser' \\"
echo "  -d 'password=testpass123'" 