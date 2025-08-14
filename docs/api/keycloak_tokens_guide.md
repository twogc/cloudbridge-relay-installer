# 🔐 Руководство по получению токенов авторизации в Keycloak

## 📋 Обзор

Keycloak настроен и доступен на порту 9443. Для получения токенов авторизации нужно настроить клиентов и пользователей через Admin Console.

## 🌐 Доступ к Keycloak

- **Admin Console**: https://89.208.86.150:9443/admin/
- **Realm**: cloudbridge
- **Внешний URL**: https://89.208.86.150:9443

## 🔧 Настройка через Admin Console

### 1. Вход в Admin Console

1. Откройте https://89.208.86.150:9443/admin/
2. Войдите с учетными данными администратора
3. Перейдите в realm "cloudbridge"

### 2. Создание клиента

1. В левом меню выберите "Clients"
2. Нажмите "Create"
3. Заполните поля:
   - **Client ID**: test-client
   - **Client Protocol**: openid-connect
   - **Root URL**: http://localhost:8080

4. На вкладке "Settings":
   - **Access Type**: confidential
   - **Standard Flow Enabled**: ON
   - **Direct Access Grants Enabled**: ON
   - **Service Accounts Enabled**: ON

5. На вкладке "Credentials" скопируйте **Client Secret**

### 3. Создание пользователя

1. В левом меню выберите "Users"
2. Нажмите "Add user"
3. Заполните поля:
   - **Username**: testuser
   - **Email**: test@example.com
   - **Email Verified**: ON

4. На вкладке "Credentials":
   - **Password**: testpass123
   - **Temporary**: OFF

## 🔑 Получение токенов

### OAuth Token (Password Grant)

```bash
curl -k -X POST "https://89.208.86.150:9443/realms/cloudbridge/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=test-client" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "username=testuser" \
  -d "password=testpass123"
```

### API Key (Client Credentials)

```bash
curl -k -X POST "https://89.208.86.150:9443/realms/cloudbridge/protocol/openid-connect/token" \
  -d "grant_type=client_credentials" \
  -d "client_id=test-client" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

## 📝 Пример ответа

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "not-before-policy": 0,
  "session_state": "12345678-1234-1234-1234-123456789012",
  "scope": "openid email profile"
}
```

## 🔍 Проверка токена

```bash
curl -k -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  "https://89.208.86.150:9443/realms/cloudbridge/protocol/openid-connect/userinfo"
```

## 🛠️ Автоматизация

### Скрипт для получения токена

```bash
#!/bin/bash
KEYCLOAK_URL="https://89.208.86.150:9443"
REALM="cloudbridge"
CLIENT_ID="test-client"
CLIENT_SECRET="YOUR_CLIENT_SECRET"
USERNAME="testuser"
PASSWORD="testpass123"

# Получение токена
TOKEN_RESPONSE=$(curl -s -k -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -d "grant_type=password" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "username=$USERNAME" \
  -d "password=$PASSWORD")

# Извлечение access_token
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "Access Token: $ACCESS_TOKEN"
```

## 🔒 Безопасность

- Храните client_secret в безопасном месте
- Используйте HTTPS для всех запросов
- Регулярно обновляйте токены
- Не передавайте токены в логах

## 📞 Поддержка

Если возникли проблемы:
1. Проверьте настройки клиента в Admin Console
2. Убедитесь, что realm "cloudbridge" существует
3. Проверьте права доступа пользователя
4. Проверьте логи Keycloak: `docker logs cloudbridge-keycloak` 