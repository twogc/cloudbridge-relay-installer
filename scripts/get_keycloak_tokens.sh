#!/bin/bash

# Скрипт для получения токенов авторизации из Keycloak
# Использование: ./scripts/get_keycloak_tokens.sh [username] [password]

KEYCLOAK_URL="https://localhost:9443"
REALM="cloudbridge"
CLIENT_ID="relay-client"
CLIENT_SECRET="JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM"

# Параметры по умолчанию - используем существующего пользователя
USERNAME=${1:-"info@2gc.ru"}
PASSWORD=${2:-"admin"}

echo "=== Получение токенов авторизации Keycloak ==="
echo "URL: $KEYCLOAK_URL"
echo "Realm: $REALM"
echo "Client ID: $CLIENT_ID"
echo "Username: $USERNAME"
echo ""

# 1. Попытка получить токен через password grant
echo "1. Попытка получить токен через password grant..."
RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
    -d "grant_type=password" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET" \
    -d "username=$USERNAME" \
    -d "password=$PASSWORD")

if echo "$RESPONSE" | grep -q "access_token"; then
    echo "✅ Успешно получен токен через password grant"
    echo "$RESPONSE" | jq .
else
    echo "❌ Ошибка получения токена через password grant:"
    echo "$RESPONSE"
fi

echo ""
echo "2. Попытка получить токен через client credentials..."
RESPONSE2=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
    -d "grant_type=client_credentials" \
    -d "client_id=$CLIENT_ID" \
    -d "client_secret=$CLIENT_SECRET")

if echo "$RESPONSE2" | grep -q "access_token"; then
    echo "✅ Успешно получен токен через client credentials"
    echo "$RESPONSE2" | jq .
else
    echo "❌ Ошибка получения токена через client credentials:"
    echo "$RESPONSE2"
fi

echo ""
echo "3. Проверка доступных клиентов в realm..."
# Попробуем получить информацию о realm
REALM_INFO=$(curl -s -k "$KEYCLOAK_URL/realms/$REALM")
if echo "$REALM_INFO" | grep -q "realm"; then
    echo "✅ Realm $REALM доступен"
    echo "Public Key: $(echo "$REALM_INFO" | jq -r '.public_key' | head -c 50)..."
else
    echo "❌ Realm $REALM недоступен"
fi

echo ""
echo "=== Инструкции по использованию ==="
echo "1. Откройте Admin Console: $KEYCLOAK_URL/admin/"
echo "2. Войдите с учетными данными администратора"
echo "3. Перейдите в Realm: $REALM"
echo "4. Проверьте настройки клиента: $CLIENT_ID"
echo "5. Убедитесь, что client_secret правильный"
echo "6. Проверьте, что пользователь $USERNAME существует и активен"
echo ""
echo "=== Примеры использования токенов ==="
echo "Для API запросов используйте заголовок:"
echo "Authorization: Bearer <access_token>"
echo ""
echo "Для проверки токена:"
echo "curl -k -H 'Authorization: Bearer <access_token>' $KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/userinfo"
echo ""
echo "=== Альтернативные способы ==="
echo "Если клиент relay-client не работает, создайте новый клиент:"
echo "1. В Admin Console создайте клиент 'test-client'"
echo "2. Установите Access Type: confidential"
echo "3. Включите Direct Access Grants"
echo "4. Скопируйте новый client_secret"
echo "5. Используйте новый client_id и client_secret" 