# Руководство по внедрению Keycloak Multi-Tenancy для CloudBridge Relay

## 1. Сборка и установка расширения Multi-Tenancy

1. Соберите JAR (если не скачивали готовый):
   ```bash
   cd keycloak-multi-tenancy
   mvn package
   ```
2. Поместите JAR в папку `keycloak_providers/` и убедитесь, что он монтируется в контейнер Keycloak через `docker-compose.yml`.

---

## 2. Запуск и базовая настройка Keycloak

1. Запустите Keycloak через docker-compose (убедитесь, что volume с JAR подключен):
   ```bash
   docker-compose up -d keycloak
   ```
2. Дождитесь, пока Keycloak будет доступен на нужном порту (обычно 8080 или 443).

---

## 3. Автоматическая настройка Keycloak (realm, client, роли, темы)

1. Используйте скрипт:
   ```bash
   bash scripts/setup/setup_keycloak.sh
   ```
   - Скрипт создаст realm, клиента, роли, применит базовые настройки.
   - Все параметры можно задать через переменные окружения или `.env`.

---

## 4. Включение Required Actions для Multi-Tenancy

1. Вручную или через API/скрипт включите:
   - Review tenant invitations
   - Create tenant
   - Select active tenant
2. Можно автоматизировать через Keycloak REST API (добавить в `setup_keycloak.sh` или отдельный скрипт).

---

## 5. Интеграция с внешними IDP (LDAP/AD/SAML/OIDC) для арендаторов

### LDAP/AD

1. Для LDAP/AD используйте:
   ```bash
   bash scripts/setup/setup_ldap_integration.sh --tenant <tenant_name> --ldap-url <ldap_url> --bind-dn "<bind_dn>" --bind-credential "<password>"
   ```
   - Скрипт создаст туннель, настроит User Federation, синхронизацию, роли и маппинг.
   - Все параметры для компаний берутся из `config/company_mapping.yaml`.

### SAML/OIDC

1. Для SAML/OIDC — настройте через UI или REST API (можно добавить автоматизацию по аналогии с LDAP).

---

## 6. Настройка claims (active_tenant, all_tenants) в токенах

1. Через UI:
   - Clients → <ваш клиент> → Client Scopes → Mappers → Add → User Attribute Mapper
   - Name: active_tenant, User Attribute: active_tenant, Token Claim Name: active_tenant
2. Можно автоматизировать через REST API (curl или отдельный bash-скрипт).

---

## 7. Проверка работы Multi-Tenancy

- Зарегистрируйте нового пользователя — должен появиться flow создания арендатора.
- Пригласите пользователя в арендатора через API или UI.
- Проверьте переключение арендатора (AIA flow).
- Проверьте, что в токене есть нужные claims.

---

## 8. Мониторинг и аудит

- Проверьте, что метрики Keycloak и Relay доступны в Prometheus/Grafana.
- Логи Keycloak и Relay должны содержать события по арендаторам, пользователям, ошибкам.

---

## 9. (Опционально) Массовое тестирование и нагрузка

- Используйте тестовые скрипты или утилиты для создания 1000+ арендаторов и пользователей.
- Проверьте изоляцию данных, скорость переключения арендаторов, работу с внешними IDP.

---

## Примеры вызова скриптов

```bash
# Настройка Keycloak
KEYCLOAK_URL=https://keycloak.example.com \
KEYCLOAK_ADMIN=admin \
KEYCLOAK_ADMIN_PASSWORD=supersecret \
REALM_NAME=cloudbridge \
CLIENT_ID=relay-client \
CLIENT_SECRET=relay-client-secret \
bash scripts/setup/setup_keycloak.sh

# Интеграция LDAP для арендатора company-a
bash scripts/setup/setup_ldap_integration.sh --tenant company-a --ldap-url ldaps://ldap.company-a.com --bind-dn "CN=SyncUser,OU=ServiceAccounts,DC=company-a,DC=com" --bind-credential "password"
```

---

## Что можно автоматизировать дополнительно
- Включение Required Actions через REST API.
- Настройка claims через REST API.
- Интеграция SAML/OIDC через отдельные скрипты.
- Генерация тестовых пользователей/арендаторов для нагрузочного тестирования.

---

## Acceptance Checklist (чеклист для приемки)
- [ ] Регистрация нового пользователя с созданием арендатора
- [ ] Приглашение пользователей в арендаторы (UI и API)
- [ ] Переключение арендатора (AIA flow)
- [ ] Массовое добавление пользователей через LDAP/AD
- [ ] Настройка tenant-specific IDP
- [ ] Корректные claims в токене (active_tenant, all_tenants)
- [ ] Изоляция данных между арендаторами
- [ ] Метрики и аудит событий
- [ ] Масштабирование и нагрузочное тестирование

---

**Вопросы и предложения по автоматизации и документации — приветствуются!** 