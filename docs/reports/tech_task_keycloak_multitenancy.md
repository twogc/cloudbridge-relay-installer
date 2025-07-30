# Выполнено на 2025 год

- Собрано и интегрировано расширение [Keycloak Multi-Tenancy](https://github.com/anarsultanov/keycloak-multi-tenancy) (v26.1.0):
  - Клонирован репозиторий, выполнена сборка JAR через Maven (JDK 17)
  - JAR-файл смонтирован в Keycloak через docker-compose (volume ./keycloak_providers)
- Настроен Keycloak для работы по HTTPS (монтирование /etc/relay/certs с реальными сертификатами)
- Keycloak работает на 443 порту (внешний) с внутренним 8443
- docker-compose обновлён: добавлены volumes, переменные окружения, команда запуска с поддержкой TLS
- После перезапуска Keycloak расширение Multi-Tenancy доступно в админке
- Включены Required Actions: Review tenant invitations, Create tenant, Select active tenant
- Проверена работа REST API Multi-Tenancy (создание, просмотр, управление арендаторами, приглашениями, членствами)

---

# Техническое задание: Внедрение Keycloak Multi-Tenancy для CloudBridge Relay

## 1. Цели и задачи

- Внедрить расширение [Keycloak Multi-Tenancy](https://github.com/anarsultanov/keycloak-multi-tenancy) для поддержки мультиарендности на уровне IAM.
- Обеспечить изоляцию данных, пользователей, ролей и интеграций для каждого арендатора (компании/клиента).
- Поддержать интеграцию с внешними корпоративными IDP (Active Directory, LDAP, SAML, OIDC) для каждого арендатора.
- Гарантировать масштабируемость, безопасность и соответствие Zero Trust архитектуре.

---

## 2. Общие требования

- Использовать последнюю совместимую версию Keycloak и расширения Multi-Tenancy.
- Все изменения должны быть документированы и покрыты тестами.
- Система должна поддерживать тысячи арендаторов и сотни тысяч пользователей.
- Все операции с пользователями и арендаторами должны быть доступны через UI и API.

---

## 3. Архитектура и интеграция

- Внедрить Keycloak Multi-Tenancy как расширение (JAR) в существующий Keycloak.
- Настроить Realm/Client/Role/Group структуру согласно best practices мультиарендности.
- Для каждого арендатора:
  - Создавать отдельный realm или использовать tenant-атрибуты в рамках одного realm (по выбранной модели).
  - Поддерживать tenant-specific IDP (AD/LDAP/SAML/OIDC).
  - Обеспечить автоматическую синхронизацию пользователей и групп из внешних IDP.
- Интеграция с CloudBridge Relay:
  - Все обращения к API и сервисам должны проходить аутентификацию через JWT, выпущенный Keycloak с мультиарендными claims.
  - В JWT должны быть отражены: active_tenant, список всех tenants пользователя, роли, кастомные атрибуты.

---

## 4. Функциональные требования

- Регистрация нового пользователя с возможностью создания нового арендатора.
- Приглашение пользователей в арендаторы через UI/API.
- Переключение активного арендатора без выхода из системы (AIA flow).
- Управление ролями и правами на уровне арендатора (tenant-admin, tenant-user, кастомные роли).
- Маппинг tenant-атрибутов в токены для использования в backend-сервисах.
- Поддержка массового добавления пользователей и групп через синхронизацию с AD/LDAP.
- Возможность настройки tenant-specific IDP через UI/API.
- API для управления арендаторами, пользователями, ролями, интеграциями.

---

## 5. Требования к безопасности

- Соответствие Zero Trust архитектуре (минимальные привилегии, изоляция, аудит).
- Поддержка MFA, password policy, session management на уровне арендатора.
- Валидация JWT: подпись, срок действия, issuer, audience, tenant claims.
- Rate limiting и защита от DoS на уровне пользователя и арендатора.
- Аудит всех критических операций (создание/удаление арендаторов, добавление пользователей, смена ролей, интеграция IDP).
- Защита от несанкционированного доступа между арендаторами.

---

## 6. Требования к масштабируемости и производительности

- Поддержка горизонтального масштабирования Keycloak и CloudBridge Relay.
- Кэширование JWKS, tenant-метаданных, ролей и атрибутов.
- Оптимизация синхронизации с внешними IDP (batch-режимы, инкрементальная синхронизация).
- Мониторинг производительности (метрики по аутентификации, переключению арендаторов, интеграциям).

---

## 7. Требования к API и автоматизации

- REST API для управления арендаторами, пользователями, ролями, интеграциями.
- API для настройки и тестирования интеграции с AD/LDAP/SAML/OIDC для каждого арендатора.
- API для массового импорта/экспорта пользователей и групп.
- Документация по API (OpenAPI/Swagger).

---

## 8. Требования к мониторингу и аудиту

- Интеграция с существующей системой мониторинга (Prometheus, Grafana).
- Метрики по аутентификации, ошибкам, времени отклика, событиям безопасности.
- Аудит-логирование всех операций с арендаторами и пользователями.
- Настройка алертов по критическим событиям (ошибки синхронизации, неудачные попытки входа, DoS).

---

## 9. Требования к документации

- Подробная инструкция по установке и настройке Keycloak Multi-Tenancy.
- Гайд по интеграции с внешними IDP (AD/LDAP/SAML/OIDC) для арендаторов.
- Описание типовых сценариев (регистрация, приглашение, переключение арендатора, интеграция IDP).
- Документация по API и best practices по безопасности.

---

## 10. Критерии приемки

- Все требования реализованы и покрыты тестами (unit, integration, нагрузочные).
- Проведено нагрузочное тестирование на 10 000+ пользователей и 1000+ арендаторов.
- Все сценарии регистрации, приглашения, переключения арендатора, интеграции с AD/LDAP/SAML/OIDC работают корректно.
- JWT-токены содержат корректные claims и валидируются всеми сервисами.
- Нет утечек данных между арендаторами.
- Документация и инструкции предоставлены.

---

**Ответственный:** [ФИО, контакт]

**Дата:** [указать дату] 

# Подробности по настройке Multi-Tenancy, API и claims

## 1. Настройка realm и Required Actions

- Войдите в Keycloak Admin UI (https://<ваш-домен>/admin)
- Перейдите в нужный realm или создайте новый
- В разделе Authentication > Required Actions включите и расставьте в следующем порядке:
  1. Review tenant invitations
  2. Create tenant
  3. Select active tenant
- В Realm Settings > Login включите Verify email (только пользователи с подтверждённой почтой видят приглашения)

## 2. Настройка claims для токенов

- В разделе Client Scopes создайте или отредактируйте мапперы:
  - **active_tenant** — текущий активный tenant пользователя
  - **all_tenants** — список всех tenants, к которым принадлежит пользователь
- Пример маппера (User Attribute Mapper):
  - Name: active_tenant
  - User Attribute: active_tenant
  - Token Claim Name: active_tenant
  - Claim JSON Type: String
  - Add to ID token: true
  - Add to access token: true

## 3. Интеграция внешних IDP для арендаторов

- Для каждого tenant можно добавить отдельный Identity Provider (OIDC, SAML, LDAP, AD)
- В разделе Identity Providers создайте нужный IDP и укажите alias (например, tenant1-oidc)
- В настройках Multi-Tenancy расширения можно указать tenant-specific IDP через API или UI

## 4. Примеры работы с Multi-Tenancy API

- **Получить список арендаторов:**
  ```http
  GET /realms/<realm>/tenants
  Authorization: Bearer <admin-token>
  ```
- **Создать арендатора:**
  ```http
  POST /realms/<realm>/tenants
  Content-Type: application/json
  Authorization: Bearer <admin-token>
  {
    "name": "tenant1",
    "attributes": {"company": ["Acme"]}
  }
  ```
- **Обновить арендатора:**
  ```http
  PUT /realms/<realm>/tenants/<tenantId>
  Content-Type: application/json
  Authorization: Bearer <admin-token>
  {
    "name": "tenant1-new",
    "attributes": {"company": ["Acme"]}
  }
  ```
- **Удалить арендатора:**
  ```http
  DELETE /realms/<realm>/tenants/<tenantId>
  Authorization: Bearer <admin-token>
  ```
- **Список приглашений:**
  ```http
  GET /realms/<realm>/tenants/<tenantId>/invitations
  Authorization: Bearer <admin-token>
  ```

## 5. Рекомендации

- Для работы с API используйте admin-токен (получить через стандартный endpoint /protocol/openid-connect/token)
- Для интеграции с внешними IDP используйте tenant-specific настройки в расширении
- Для автоматизации создания/управления tenant-ами используйте предоставленный OpenAPI (openapi.json) 

## Получение токенов авторизации

### 1. OAuth Token (для пользователей)

#### Через curl (Password Grant):
```bash
curl -X POST https://edge.2gc.ru/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=relay-client" \
  -d "client_secret=JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM" \
  -d "username=your_username" \
  -d "password=your_password"
```

#### Через curl (Client Credentials Grant):
```bash
curl -X POST https://edge.2gc.ru/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=relay-client" \
  -d "client_secret=JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM"
```

#### Через JavaScript (для веб-приложений):
```javascript
const tokenResponse = await fetch('https://edge.2gc.ru/realms/cloudbridge/protocol/openid-connect/token', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
  },
  body: new URLSearchParams({
    grant_type: 'password',
    client_id: 'relay-client',
    client_secret: 'JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM',
    username: 'your_username',
    password: 'your_password'
  })
});

const tokenData = await tokenResponse.json();
const accessToken = tokenData.access_token;
```

### 2. Admin Token (для управления Multi-Tenancy)

#### Получение admin токена:
```bash
curl -X POST https://edge.2gc.ru/realms/cloudbridge/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" \
  -d "username=admin" \
  -d "password=admin"
```

### 3. API Key (для сервисов)

#### Создание API Key через Keycloak Admin UI:
1. Войдите в Keycloak Admin Console (https://edge.2gc.ru/admin)
2. Перейдите в Clients → relay-client → Credentials
3. Скопируйте Client Secret: `JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM`

#### Использование API Key в заголовках:
```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
     -H "X-API-Key: JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM" \
     https://edge.2gc.ru/realms/cloudbridge/tenants
```

### 4. Тестирование токенов

#### Проверка токена:
```bash
curl -X POST https://edge.2gc.ru/realms/cloudbridge/protocol/openid-connect/userinfo \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### Декодирование JWT токена:
```bash
echo "YOUR_ACCESS_TOKEN" | cut -d. -f2 | base64 -d | jq .
```

### 5. Примеры для Multi-Tenancy API

#### Создание tenant с admin токеном:
```bash
curl -X POST https://edge.2gc.ru/realms/cloudbridge/tenants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN" \
  -d '{
    "name": "tenant1",
    "attributes": {"company": ["Acme Corp"]}
  }'
```

#### Получение списка tenants:
```bash
curl -X GET https://edge.2gc.ru/realms/cloudbridge/tenants \
  -H "Authorization: Bearer ADMIN_ACCESS_TOKEN"
```

### 6. Настройки для тестирования

- **Realm**: cloudbridge
- **Client ID**: relay-client  
- **Client Secret**: JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM
- **Admin Client**: admin-cli
- **Admin User**: max0r
- **Admin Password**: 7Thbngy^
- **URL**: https://edge.2gc.ru (порт 443)

### 7. Актуальные токены (обновлено 2025)

#### OAuth Token (Access Token):
```
eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJsc2E0ekI0cEhqN2YyejAxWXVfMXp5dnBEUFNtS1ZaelNvTkdZb0NZenA0In0.eyJleHAiOjE3NTM2MDAyMzEsImlhdCI6MTc1MzU5OTkzMSwianRpIjoib25sdHJvOmIyMmUyODNkLWMxYzYtZGI5My04YWEzLTZkYzhmYWY4MzI1YyIsImlzcyI6Imh0dHBzOi8vZWRnZS4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwidHlwIjoiQmVhcmVyIiwiYXpwIjoiYWRtaW4tY2xpIiwic2lkIjoiNzZlNmY3ODQtMzJmMS00OWY1LTk3MDctNTIyZDFkMzI2ZGNhIiwic2NvcGUiOiJwcm9maWxlIGVtYWlsIn0.c_rmwhtsZ2G59CbB6vFwAh77cXCkxOniA6tHWCWWCDp-4-UqAYz5IqDGEgriCB4dLQHfDy390eyDccrew3l1n1mHou66eMc4ltq9-vbIyNl12Puj8HnPbCfzPsBY5pd9EQP3oXcuu6Qh6bva6n1ueJEU8hfM5Aqp2zRElIZeyoqqCjr7DT_uKs35O44xxrXXcS9bF2bn7Shz7_3y9WzMaJNvQS7a9yxI0lSQkKI7MAAVeOgyMFiDcKSWGIFZTVKubGnX5-3oLJ0eW0BS6HdVylUPwRP2M3Onfi4Bsx7IB8KHrQ-r_m-6BhcAxu2UWonL_5z_EigMxzFMjkRa5ww4nQ
```

#### Refresh Token:
```
eyJhbGciOiJIUzUxMiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI5MTdiMDFiOC1jZTI0LTQwYjktOWNiZS03YzA3ZThjNGY0YTQifQ.eyJleHAiOjE3NTM2MDE3MzEsImlhdCI6MTc1MzU5OTkzMSwianRpIjoiZDM4MGM0MjAtYWM1Mi0xYjU4LWUyNjUtMDU1NWU4NjFiZjczIiwiaXNzIjoiaHR0cHM6Ly9lZGdlLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJodHRwczovL2VkZ2UuMmdjLnJ1L3JlYWxtcy9jbG91ZGJyaWRnZSIsInR5cCI6IlJlZnJlc2giLCJhenAiOiJhZG1pbi1jbGkiLCJzaWQiOiI3NmU2Zjc4NC0zMmYxLTQ5ZjUtOTcwNy01MjJkMWQzMjZkY2EiLCJzY29wZSI6ImFjciB3ZWItb3JpZ2lucyBwcm9maWxlIGVtYWlsIGJhc2ljIHJvbGVzIn0.STh8U882j8W4ZZbPkDhGnpys0LSbS1nUOYx0R8bzqJkWD4dRS80L-2Zf2UVj5KecUH6U5YqdS5JRnGToFmDOhw
```

#### API Key (Client Secret):
```
JLowpQ24MDLYkHM7auMYzJSTbp9eCjfM
```

# Структура проекта и примеры интеграции IDP, claims, автоматизации

## Где что находится

- **keycloak-multi-tenancy/** — исходники расширения Multi-Tenancy (Java, Maven)
- **keycloak_providers/** — собранный JAR-файл расширения для Keycloak
- **config/** — конфигурационные файлы (в т.ч. keycloak_config.json)
- **scripts/setup/** — скрипты автоматизации и настройки (setup_keycloak.sh, setup_ldap_integration.sh, setup_certs.sh и др.)
- **docs/integration/** — документация по интеграции Keycloak, Multi-Tenancy, LDAP, OIDC, SAML и др.

## Примеры интеграции внешних IDP для арендаторов

### OIDC (OpenID Connect)
1. В Keycloak Admin UI: Identity Providers → Add provider → OpenID Connect
2. Заполнить поля:
   - Alias: tenant1-oidc
   - Client ID, Client Secret, Discovery URL — от внешнего провайдера
3. В настройках расширения Multi-Tenancy указать alias для tenant

### LDAP
1. В Keycloak Admin UI: User Federation → Add provider → LDAP
2. Заполнить:
   - Connection URL: ldap://ldap.example.com
   - Bind DN, Bind Credential, User Search Base и др.
3. Для tenant-specific LDAP используйте скрипт scripts/setup/setup_ldap_integration.sh

### SAML
1. Identity Providers → Add provider → SAML
2. Заполнить:
   - Alias: tenant1-saml
   - Single Sign-On Service URL, Entity ID, сертификаты
3. Привязать к нужному tenant через расширение

## Пример настройки claims (active_tenant, all_tenants)
1. В разделе Client Scopes → <scope> → Mappers → Create
2. Тип: User Attribute
3. User Attribute: active_tenant
4. Token Claim Name: active_tenant
5. Claim JSON Type: String
6. Add to ID token: true, Add to access token: true

## Пример автоматизации через bash-скрипты
- **scripts/setup/setup_keycloak.sh** — автоматизация развёртывания и настройки Keycloak
- **scripts/setup/setup_ldap_integration.sh** — автоматизация интеграции LDAP
- **scripts/setup/setup_certs.sh** — автоматизация установки сертификатов

Пример вызова скрипта для настройки LDAP:
```bash
bash scripts/setup/setup_ldap_integration.sh --tenant tenant1 --ldap-url ldap://ldap.example.com --bind-dn "cn=admin,dc=example,dc=com" --bind-credential "password"
``` 