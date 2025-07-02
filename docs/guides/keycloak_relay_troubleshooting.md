# Интеграция Keycloak и CloudBridge Relay: устранение ошибок

## Ошибка в дашборде: `Invalid parameter: redirect_uri`

**Причина:**
В настройках клиента Keycloak не были указаны корректные redirect URIs для вашего UI (например, http://localhost:5173/callback).

**Решение:**
1. Получите admin token Keycloak:
   ```bash
   export ADMIN_TOKEN=$(curl -s -X POST "http://localhost:8080/realms/master/protocol/openid-connect/token" \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password&client_id=admin-cli&username=admin&password=<ВАШ_ПАРОЛЬ>" | jq -r '.access_token')
   ```
2. Получите ID клиента:
   ```bash
   CLIENT_ID=$(curl -s -X GET "http://localhost:8080/admin/realms/cloudbridge/clients?clientId=relay-client" \
     -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')
   ```
3. Обновите redirectUris и webOrigins:
   ```bash
   curl -s -X PUT "http://localhost:8080/admin/realms/cloudbridge/clients/$CLIENT_ID" \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "redirectUris": [
         "http://localhost:5173/*",
         "http://localhost:5173/callback",
         "https://localhost:5173/*",
         "https://localhost:5173/callback"
       ],
       "webOrigins": [
         "http://localhost:5173",
         "https://localhost:5173"
       ]
     }'
   ```
4. Перезапустите UI и попробуйте снова войти.

---

## Ошибка API: `invalid token` или `token is unverifiable: no keyfunc was provided`

**Причина:**
- В конфиге relay/adminapi был указан неправильный client_secret, public_key, issuer, audience или server_url.
- Контейнер relay не мог получить JWKS (ключи подписи) от Keycloak.

**Решение:**
1. Проверьте, что в `config/config.json` и `config/keycloak_config.json`:
   ```json
   "keycloak": {
     "enabled": true,
     "server_url": "http://keycloak:8080",
     "realm": "cloudbridge",
     "client_id": "relay-client",
     "client_secret": "<АКТУАЛЬНЫЙ_СЕКРЕТ>",
     "public_key": "-----BEGIN PUBLIC KEY-----\n...новый ключ...\n-----END PUBLIC KEY-----",
     "token_validation": {
       "issuer": "http://keycloak:8080/realms/cloudbridge",
       "audience": "relay-client",
       "clock_skew": "30s"
     }
   }
   ```
   - client_secret можно получить через API или UI Keycloak.
   - public_key можно получить так:
     ```bash
     curl -s http://localhost:8080/realms/cloudbridge | jq -r '.public_key'
     ```
   - issuer и audience должны совпадать с настройками клиента и realm.

2. Перезапустите relay:
   ```bash
   docker-compose down
   docker-compose up --build -d
   ```

3. Проверьте логи:
   ```bash
   docker logs relay --tail 50
   docker exec relay cat /var/log/relay/relay.log | tail -30
   ```

4. Проверьте сетевой доступ из контейнера relay:
   ```bash
   docker exec relay wget -qO- http://keycloak:8080/realms/cloudbridge/protocol/openid-connect/certs
   ```
   Если не работает — проверьте, что оба сервиса в одной сети (relay-network).

---

## Проверка работы

1. Получите access_token:
   ```bash
   TOKEN=$(curl -s -X POST http://localhost:8080/realms/cloudbridge/protocol/openid-connect/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=password&client_id=relay-client&client_secret=<АКТУАЛЬНЫЙ_СЕКРЕТ>&username=admin&password=<ПАРОЛЬ>" | jq -r '.access_token')
   ```
2. Проверьте admin API:
   ```bash
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8083/api/status
   ```
   Ожидаемый ответ:
   ```json
   {"status":"ok","message":"Admin API is running"}
   ```

---

## Примечания

- После каждого изменения конфигов или секретов — обязательно пересобирайте и перезапускайте контейнер relay.
- Для production обязательно используйте HTTPS и ограничьте origins в CORS.
- Все логи backend пишутся в logs/relay.log (или /var/log/relay/relay.log внутри контейнера).

---

**Если после всех шагов ошибка сохраняется — приложите свежие логи relay.log и docker logs relay для диагностики.** 