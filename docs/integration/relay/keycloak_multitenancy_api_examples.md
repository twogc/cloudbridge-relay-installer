# Примеры использования REST API Keycloak Multi-Tenancy

## Получение списка арендаторов
```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/realms/<realm>/tenants"
```

## Создание арендатора
```bash
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "tenant1", "attributes": {"company": ["Acme"]}}' \
  "$KEYCLOAK_URL/realms/<realm>/tenants"
```

## Обновление арендатора
```bash
curl -X PUT -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "tenant1-new", "attributes": {"company": ["Acme"]}}' \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>"
```

## Удаление арендатора
```bash
curl -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>"
```

## Список приглашений
```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>/invitations"
```

## Создание приглашения
```bash
curl -X POST -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "roles": ["tenant-user"]}' \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>/invitations"
```

## Список членств арендатора
```bash
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>/memberships"
```

## Обновление членства (роли)
```bash
curl -X PATCH -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"roles": ["tenant-admin"]}' \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>/memberships/<membershipId>"
```

## Удаление членства
```bash
curl -X DELETE -H "Authorization: Bearer $ADMIN_TOKEN" \
  "$KEYCLOAK_URL/realms/<realm>/tenants/<tenantId>/memberships/<membershipId>"
```

## Привязка IDP к арендатору (пример для OIDC)
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

**Для всех запросов используйте admin-токен, полученный через стандартный endpoint Keycloak.** 