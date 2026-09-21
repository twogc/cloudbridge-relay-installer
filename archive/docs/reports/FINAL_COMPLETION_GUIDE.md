# Финальное завершение CloudBridge + ZITADEL интеграции

**Время на завершение:** 2-3 часа
**Статус:** 75% готово → 100% готово
**Приоритет:** Критичный

---

## **ДОСТУПЫ К ZITADEL CONSOLE**

```
URL: https://edge.2gc.ru/
Email: zitadel-admin@zitadel.edge.2gc.ru
Password: suppIw-tegvu9-jocxiq
```

---

## **КРИТИЧНЫЕ ЗАДАЧИ ДЛЯ ЗАВЕРШЕНИЯ**

### 1⃣ **Настроить Custom Scopes (30 мин)**

**Проблема:** Нет scopes `relay.api` и `relay.k8s`
**Решение через Console:**

1. Зайти в https://edge.2gc.ru/
2. Projects → `cloudbridge-relay` (ID: 339537327304999215)
3. Roles → Add Role:
   - Key: `relay-api`, Display: `CloudBridge Relay API Access`, Group: `api`
   - Key: `relay-k8s`, Display: `Kubernetes Proxy Access`, Group: `k8s`
4. Applications → `cloudbridge-relay-api` → Settings
5. Scopes → Add: `relay.api`, `relay.k8s`

**Альтернатива (API):**
```bash
source configs/developer-env.sh
# Создание ролей через Management API
curl -k -X POST "$ZITADEL_MANAGEMENT_API/projects/$PROJECT_ID/roles" \
  -H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key": "relay-api", "displayName": "Relay API Access"}'
```

---

### 2⃣ **Включить Client Credentials Flow (15 мин)**

**Проблема:** API приложение не поддерживает `client_credentials`
**Решение через Console:**

1. Projects → `cloudbridge-relay`
2. Applications → `cloudbridge-relay-api` (ID: 339537393122017581)
3. Settings → Grant Types
4. Включить: `Client Credentials`
5. Включить: `Token Exchange` (для mTLS)
6. Auth Method → `Client Secret (Basic)` или `Private Key JWT` (для mTLS)

**Проверка:**
```bash
# После настройки - тест получения токена
curl -k -X POST "https://edge.2gc.ru/oauth/v2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&client_id=339537393138794797&client_secret=Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh&scope=openid profile relay.api"
```

---

### 3⃣ **Настроить Custom Claims (20 мин)**

**Проблема:** JWT не содержит `tenant_id`, `client_id`, `roles`
**Решение через Console:**

1. Projects → `cloudbridge-relay` → Settings
2. Token Settings / Claims Mapping
3. Добавить custom claims:
   - `tenant_id` → User Metadata: `tenant_id`
   - `client_id` → Static Value: `${client_id}`
   - `roles` → User Roles: `project_roles`
   - `audit_session_id` → Generate: UUID

**Альтернативный путь:**
1. Applications → `cloudbridge-relay-api`
2. Token Settings
3. Include Role Assertion in Access Token
4. Include Role Assertion in ID Token

---

## **ОПЦИОНАЛЬНЫЕ ДОРАБОТКИ**

### 4⃣ **Создать организацию "cloudbridge" (1 час)**

**Текущее:** Работаем в org "ZITADEL"
**Требуется:** Отдельная org "cloudbridge"

1. Console → Organizations → Create
2. Name: `cloudbridge`
3. Migrate projects:
   - Export project `cloudbridge-relay`
   - Import в новую org

### 5⃣ **Настроить Audit WebHook → Loki (30 мин)**

1. Console → Notifications → WebHooks
2. Add WebHook:
   - URL: `http://promtail:3100/loki/api/v1/push`
   - Events: `user.token.added`, `org.project.app.*`

---

## **ПРОВЕРКА ГОТОВНОСТИ**

### Acceptance Criteria Tests:

#### 1. **kubectl exec-plugin работает:**
```bash
cd /home/ubuntu/cloudbridge-relay-installer/cmd/kubectl-zitadel
./kubectl-zitadel --config ../../configs/kubectl-zitadel-config.json --debug
# Ожидается: JSON с ExecCredential и access_token
```

#### 2. **gRPC доступен с JWT:**
```bash
# Получить токен через plugin
TOKEN=$(./kubectl-zitadel --config ../../configs/kubectl-zitadel-config.json | jq -r '.status.token')

# Тест gRPC
grpcurl -authority relay.2gc.space:8444 -H "Authorization: Bearer $TOKEN" -insecure list
```

#### 3. **Kubernetes RBAC работает:**
```bash
# Создать kubeconfig с exec plugin
cat > ~/.kube/relay-config << EOF
apiVersion: v1
kind: Config
users:
- name: zitadel-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: /home/ubuntu/cloudbridge-relay-installer/cmd/kubectl-zitadel/kubectl-zitadel
      args:
      - --config=/home/ubuntu/cloudbridge-relay-installer/configs/kubectl-zitadel-config.json
EOF

# Тест kubectl
kubectl get ns --kubeconfig ~/.kube/relay-config
```

---

## **СТАТУС ПО COMPLIANCE_CHECK.md**

| Пункт | До | После | Статус |
|-------|----|----|--------|
| 1. Базовая настройка | 75% | 95% |→|
| 2. Токены/scopes | 40% | 95% |→|
| 3. Service Account | 95% | 100% | |
| 4. mTLS/JWT Bearer | 70% | 95% |→|
| 5. RBAC Kubernetes | 100% | 100% | |
| 6. kubectl plugin | 25% | 95% |→|
| 7. Audit/Compliance | 50% | 80% |→ |
| 8. Observability | 100% | 100% | |
| 9. Документация | 100% | 100% | |
| 10. Acceptance | 30% | 90% |→|

**Итого: 68% → 95% готовности**

---

## **ПОЛЕЗНЫЕ КОМАНДЫ ДЛЯ ДИАГНОСТИКИ**

### Environment setup:
```bash
cd /home/ubuntu/cloudbridge-relay-installer
source configs/developer-env.sh
```

### Проверка ZITADEL доступа:
```bash
curl -k -H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" "$ZITADEL_MANAGEMENT_API/orgs/me"
```

### Список приложений в проекте:
```bash
curl -k -H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
  "$ZITADEL_MANAGEMENT_API/projects/$PROJECT_ID/apps/_search" \
  -H "Content-Type: application/json" -d '{"queries": []}'
```

### Проверка grant types:
```bash
curl -k -H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
  "$ZITADEL_MANAGEMENT_API/projects/$PROJECT_ID/apps/$API_APP_ID"
```

---

## **РЕЗУЛЬТАТ ПОСЛЕ ЗАВЕРШЕНИЯ**

**100% соответствие техзаданию**
**kubectl exec-plugin работает с ZITADEL**
**mTLS аутентификация готова**
**Custom scopes и claims настроены**
**Token Exchange активирован**
**Kubernetes RBAC интегрирован**
**Мониторинг и observability готовы**
**Полная документация**

**СИСТЕМА ГОТОВА К PRODUCTION ИСПОЛЬЗОВАНИЮ**
