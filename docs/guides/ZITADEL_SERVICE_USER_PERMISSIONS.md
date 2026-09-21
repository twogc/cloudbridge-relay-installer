# Service User Permissions Fix

## **Проблема**
Service User `cloudbridge-admin` не имеет прав доступа к Management API ZITADEL.

**Ошибка:**
```json
{
  "code": 5,
  "message": "membership not found (AUTHZ-cdgFk)"
}
```

## **Причина**
Согласно официальной документации ZITADEL (`reference-repos/auth/zitadel/docs/docs/guides/integrate/zitadel-apis/access-zitadel-apis.md`), для доступа к Management API Service User **обязательно должен иметь Manager роль**.

## **Решение: Назначение Manager роли**

### Шаг 1: Войдите в ZITADEL Console
- **URL:** https://edge.2gc.ru/ui/console/
- **Логин:** `admin@cloudbridge.local`
- **Пароль:** `CloudBridge123!`

### Шаг 2: Найдите Service User
1. Перейдите в **Users** → **Service Users**
2. Найдите пользователя: `cloudbridge-admin`
3. Откройте его профиль

### Шаг 3: Назначьте роль ORG_OWNER

**Способ 1: Через организацию**
1. Перейдите в **Organizations** → выберите свою организацию
2. В правом верхнем углу нажмите **"+"** (Add Manager)
3. Найдите `cloudbridge-admin`
4. Назначьте роль **"Org Owner"** (`ORG_OWNER`)
5. Нажмите **Save**

**Способ 2: Через Service User**
1. В профиле Service User найдите раздел **"Authorizations"** или **"Grants"**
2. Нажмите **"New Authorization"**
3. Выберите вашу организацию
4. Назначьте роль **"Org Owner"**
5. Сохраните изменения

### Шаг 4: Проверьте права
После назначения роли Service User получит права:
- **ORG_OWNER**: Управление всем в рамках организации
- **Доступ к Management API**: создание проектов, приложений, пользователей

## **Использование токенов после назначения ролей**

### Personal Access Token (готов к использованию)
```bash
ZITADEL_TOKEN="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"

# Тест доступа
curl -s -H "Authorization: Bearer $ZITADEL_TOKEN" \
  "https://edge.2gc.ru/management/v1/orgs/_search" \
  -H "Content-Type: application/json" \
  -d '{"queries": []}' | jq .
```

### JSON Key (требует JWT Bearer flow)
```bash
# Обязательный scope для ZITADEL API
SCOPE="openid profile urn:zitadel:iam:org:project:id:zitadel:aud"

# JWT Bearer запрос
curl --request POST \
  --url https://edge.2gc.ru/oauth/v2/token \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data "grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer" \
  --data "scope=$SCOPE" \
  --data "assertion=YOUR_JWT_ASSERTION"
```

## **Доступные Manager роли**

| Роль | Описание | Права |
|------|----------|-------|
| `ORG_OWNER` | **Владелец организации** | Полное управление организацией |
| `ORG_USER_MANAGER` | Менеджер пользователей | Управление пользователями |
| `PROJECT_OWNER` | Владелец проекта | Управление конкретным проектом |
| `IAM_OWNER` | Владелец экземпляра | Управление всем экземпляром ZITADEL |

**Рекомендация:** Используйте `ORG_OWNER` для CloudBridge Service User.

## **После назначения роли**

1. **Создайте OIDC приложения:**
   ```bash
   cd /home/ubuntu/cloudbridge-relay-installer
   export ZITADEL_TOKEN="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"
   ./scripts/setup-cloudbridge-zitadel.sh
   ```

2. **Настройте custom scopes:**
   ```bash
   ./scripts/setup-zitadel-scopes.sh
   ```

3. **Финализируйте интеграцию:**
   ```bash
   ./scripts/finalize-cloudbridge-setup.sh <CLIENT_ID> <CLIENT_SECRET> <SPA_CLIENT_ID>
   ```

## **Ссылки на документацию**

- [Access ZITADEL APIs](reference-repos/auth/zitadel/docs/docs/guides/integrate/zitadel-apis/access-zitadel-apis.md)
- [Service User Authentication](reference-repos/auth/zitadel/docs/docs/guides/integrate/service-users/authenticate-service-users.md)
- [Manager Roles](reference-repos/auth/zitadel/docs/docs/guides/manage/console/managers.mdx)
- [Private Key JWT](reference-repos/auth/zitadel/docs/docs/guides/integrate/service-users/private-key-jwt.md)

---

**После назначения ORG_OWNER роли CloudBridge + ZITADEL интеграция будет полностью работоспособна!**
