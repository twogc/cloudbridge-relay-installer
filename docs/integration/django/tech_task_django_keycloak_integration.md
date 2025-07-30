# Техническое задание: Интеграция Django (palm) с Keycloak Multi-Tenancy

## 1. Цели и задачи
- Автоматизировать процесс регистрации компаний (tenants) через Django.
- Создавать главную учётную запись клиента (owner/admin) с возможностью приглашать пользователей и назначать им права.
- Обеспечить возможность включения интеграции LDAP/AD и других внешних IDP для арендатора.
- Гарантировать изоляцию данных, безопасность и аудит всех операций.

---

## 2. Архитектурные требования и best practices

- **Следовать best practices и рекомендациям из официального [README расширения Keycloak Multi-Tenancy](https://github.com/anarsultanov/keycloak-multi-tenancy#readme):**
  - Features, Configuration, Tenant Creation and Management, IDP and SSO Integration, Token claims, API.
- **Required Actions:**
  - После регистрации пользователь должен пройти flow: Review tenant invitations → Create tenant → Select active tenant.
  - Если пользователь не приглашён, он создаёт новый tenant.
- **Token claims:**
  - В access token должны быть корректные claims: `active_tenant`, `all_tenants`, tenant-roles.
  - Django backend обязан валидировать эти claims для авторизации и изоляции данных.
- **Tenant switching:**
  - Поддерживать Application-Initiated Actions (AIA) для смены активного арендатора без выхода из системы.
- **SSO/IDP:**
  - Для tenant-specific IDP (LDAP, OIDC, SAML) Django должен позволять добавить IDP через UI и вызвать нужный API Keycloak.
  - В настройках IDP обязательно указывать `multi-tenancy.tenants` через API.
  - Для автоматического добавления пользователей в tenant при логине через IDP — использовать authenticator Create tenant membership в flow Keycloak.
- **Tenant Management Role:**
  - Для админов Django/суперпользователей можно назначать роль `manage-tenants` для управления всеми арендаторами.
- **Безопасность:**
  - Django backend не должен реализовывать собственную логику аутентификации — только вызывать API Keycloak и работать с JWT.
  - Все операции выполнять через публичные REST API расширения.
  - Валидировать tenant claims из токена для изоляции данных.
- **UX:**
  - Следовать flow, описанному в README, для seamless onboarding и управления арендаторами.

---

## 3. Основные сценарии

### 3.1. Регистрация компании и главного пользователя
- Пользователь заполняет форму регистрации компании (название, email, имя, пароль).
- Django backend:
  - Создаёт запись о компании в своей БД.
  - Вызывает API Keycloak для создания арендатора (tenant).
  - Создаёт главного пользователя (owner/admin) с ролью tenant-admin (или отправляет приглашение).
  - Сохраняет tenant_id, user_id в своей БД.

### 3.2. Приглашение пользователей и назначение ролей
- Главный пользователь может приглашать новых пользователей (email, роль).
- Django вызывает API Keycloak для создания приглашения и назначения ролей.
- Пользователь получает email-приглашение, регистрируется и автоматически добавляется в нужный tenant с нужной ролью.

### 3.3. Включение интеграции LDAP/AD (по желанию)
- В личном кабинете есть форма для ввода параметров LDAP/AD.
- Django вызывает bash-скрипт или REST API для настройки интеграции с LDAP/AD для конкретного tenant.
- После успешной интеграции пользователи и группы синхронизируются автоматически.

---

## 4. API Keycloak Multi-Tenancy (OAS 3.1)

**Базовый URL:**
```
{keycloakUrl}/realms/{realmName}
```

### Основные эндпоинты:

#### **Тенанты**
- `GET /tenants` — список тенантов (фильтры: first, max, q, search)
- `POST /tenants` — создать тенанта
- `GET /tenants/{tenantId}` — получить тенанта
- `PUT /tenants/{tenantId}` — обновить тенанта
- `DELETE /tenants/{tenantId}` — удалить тенанта

#### **Приглашения**
- `GET /tenants/{tenantId}/invitations` — список приглашений (фильтры: first, max, search)
- `POST /tenants/{tenantId}/invitations` — создать приглашение
- `DELETE /tenants/{tenantId}/invitations/{invitationId}` — отозвать приглашение

#### **Членства**
- `GET /tenants/{tenantId}/memberships` — список членств (фильтры: first, max, search)
- `PATCH /tenants/{tenantId}/memberships/{membershipId}` — обновить роли членства
- `DELETE /tenants/{tenantId}/memberships/{membershipId}` — удалить членство

### Примеры схем:

#### **TenantRepresentation**
```json
{
  "id": "string",
  "name": "string",
  "realm": "string",
  "attributes": {
    "additionalProp1": ["string"]
  }
}
```

#### **TenantInvitationRepresentation**
```json
{
  "id": "string",
  "tenantId": "string",
  "email": "string",
  "locale": "string",
  "roles": ["string"],
  "invitedBy": "string"
}
```

#### **TenantMembershipRepresentation**
```json
{
  "id": "string",
  "user": {
    "id": "string",
    "createdTimestamp": 0,
    "username": "string",
    "enabled": true,
    "firstName": "string",
    "lastName": "string",
    "email": "string",
    "emailVerified": true,
    "federationLink": "string"
  },
  "roles": ["string"]
}
```

### Примеры запросов:

#### **Создать тенанта**
```http
POST /tenants
Content-Type: application/json
{
  "name": "company-a",
  "attributes": {"company": ["Company A"]}
}
```

#### **Создать приглашение**
```http
POST /tenants/{tenantId}/invitations
Content-Type: application/json
{
  "email": "user@company.com",
  "roles": ["tenant-user"]
}
```

#### **Обновить роли членства**
```http
PATCH /tenants/{tenantId}/memberships/{membershipId}
Content-Type: application/json
{
  "roles": ["tenant-admin"]
}
```

### Ожидаемые коды ответов:
- 200/201 — успех
- 204 — успешно, без тела
- 401 — неавторизован
- 403 — запрещено
- 404 — не найдено
- 409 — конфликт (например, имя уже занято)

---

## 5. Требования к безопасности
- Все действия выполнять от имени сервисного admin-токена Keycloak.
- Не хранить пароли и секреты в коде или git, использовать защищённые переменные окружения.
- Валидация всех входных данных (email, DN, параметры LDAP).
- Логирование и аудит всех операций регистрации, приглашения, назначения ролей, интеграции IDP.
- Гарантировать изоляцию данных между компаниями (tenant_id).

---

## 6. Best practices
- Использовать асинхронные вызовы для API и интеграций.
- Обрабатывать ошибки и возвращать информативные сообщения пользователю.
- Для интеграции LDAP/AD — валидировать соединение до сохранения настроек.
- Для приглашений — реализовать повторную отправку и трекинг статуса.
- Для безопасности — реализовать MFA, password policy, session management (см. best practices).

---

## 7. Acceptance Criteria (критерии приёмки)
- [ ] Регистрация компании и главного пользователя полностью автоматизирована.
- [ ] Главный пользователь может приглашать пользователей и назначать им роли через UI.
- [ ] Включение LDAP/AD интеграции работает через UI и backend.
- [ ] Все операции логируются и доступны для аудита.
- [ ] Изоляция данных между компаниями гарантирована.
- [ ] Все API-вызовы защищены и валидируются.
- [ ] Документация по интеграции и usage для backend-методов предоставлена.

---

**Вопросы и уточнения приветствуются до начала разработки!** 