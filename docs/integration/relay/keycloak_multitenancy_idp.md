# Интеграция внешних IDP (LDAP/AD, SAML, OIDC) для арендаторов Keycloak Multi-Tenancy

## 1. Введение

Keycloak Multi-Tenancy позволяет каждому арендатору (tenant) использовать собственный корпоративный Identity Provider (IDP): LDAP/AD, SAML, OIDC. Это обеспечивает изоляцию, гибкость и соответствие требованиям безопасности.

- **Tenant-specific IDP** — IDP, который доступен только определённым арендаторам.
- Поддерживаются: LDAP/AD, SAML, OIDC.

---

## 2. Интеграция LDAP/AD

### 2.1. Автоматизация через скрипт

Используйте скрипт:
```bash
bash scripts/setup/setup_ldap_integration.sh \
  --tenant company-a \
  --ldap-url ldaps://ldap.company-a.com \
  --bind-dn "CN=SyncUser,OU=ServiceAccounts,DC=company-a,DC=com" \
  --bind-credential "password"
```
- Скрипт создаёт туннель, настраивает User Federation, синхронизацию, роли и маппинг.
- Параметры для компаний берутся из `config/company_mapping.yaml`.

### 2.2. Проверка успешности
- В Keycloak → User Federation должен появиться LDAP-провайдер с именем `ldap-<tenant>`.
- Пользователи и группы должны синхронизироваться.
- В логах не должно быть ошибок соединения.

### 2.3. Troubleshooting
- Проверьте корректность DN, пароля, сертификатов.
- Используйте `ldapsearch` для ручной проверки соединения.
- Логи Keycloak и Relay помогут выявить ошибки синхронизации.

---

## 3. Интеграция SAML

### 3.1. Через UI
1. В Keycloak Admin UI: Identity Providers → Add provider → SAML
2. Заполните:
   - Alias: tenant1-saml
   - Single Sign-On Service URL, Entity ID, сертификаты
3. В настройках IDP добавьте атрибут:
   - `multi-tenancy.tenants = <tenant_id>` (через REST API или UI, если поддерживается)

### 3.2. Пример curl для создания SAML IDP
```bash
curl -X POST "$KEYCLOAK_URL/admin/realms/<realm>/identity-provider/instances" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "alias": "tenant1-saml",
    "providerId": "saml",
    "enabled": true,
    "config": {
      "singleSignOnServiceUrl": "https://sso.example.com/saml",
      "entityId": "tenant1-saml",
      "multi-tenancy.tenants": "<tenant_id>"
    }
  }'
```

---

## 4. Интеграция OIDC

### 4.1. Через UI
1. В Keycloak Admin UI: Identity Providers → Add provider → OpenID Connect
2. Заполните:
   - Alias: tenant1-oidc
   - Client ID, Client Secret, Discovery URL
3. В настройках IDP добавьте атрибут:
   - `multi-tenancy.tenants = <tenant_id>`

### 4.2. Пример curl для создания OIDC IDP
```bash
curl -X POST "$KEYCLOAK_URL/admin/realms/<realm>/identity-provider/instances" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "alias": "tenant1-oidc",
    "providerId": "oidc",
    "enabled": true,
    "config": {
      "clientId": "<client_id>",
      "clientSecret": "<client_secret>",
      "discoveryUrl": "https://idp.example.com/.well-known/openid-configuration",
      "multi-tenancy.tenants": "<tenant_id>"
    }
  }'
```

---

## 5. Автоматизация через REST API

- Для массового создания/обновления IDP используйте curl или скрипты.
- Для tenant-specific IDP обязательно указывайте `multi-tenancy.tenants` в config.
- Для обновления существующего IDP:
```bash
curl -X PUT "$KEYCLOAK_URL/admin/realms/<realm>/identity-provider/instances/<alias>" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    ...
    "config": {
      ...
      "multi-tenancy.tenants": "<tenant_id>"
    }
  }'
```

---

## 6. Best practices

- **Безопасность:**
  - Включайте MFA для всех внешних IDP.
  - Используйте только доверенные сертификаты.
  - Настройте password policy на стороне IDP и Keycloak.
- **Изоляция:**
  - Проверяйте, что каждый IDP привязан только к нужным арендаторам.
  - Не используйте публичные IDP для приватных данных.
- **Аудит и мониторинг:**
  - Включите аудит входов через IDP.
  - Мониторьте ошибки и задержки аутентификации.

---

## 7. Чеклист для проверки интеграции

- [ ] IDP создан и активен для нужного арендатора
- [ ] Пользователь может войти через внешний IDP и получить нужные tenant claims
- [ ] В токене присутствует корректный active_tenant
- [ ] Изоляция между арендаторами не нарушается
- [ ] Логи и метрики отражают события входа через IDP
- [ ] MFA и password policy работают корректно

---

**Вопросы и предложения по автоматизации и интеграции приветствуются!** 