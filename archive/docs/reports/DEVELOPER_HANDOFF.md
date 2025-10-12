# CloudBridge + ZITADEL: Developer Handoff Package
## Данные для доработки через API

**Дата передачи:** $(date)  
**Статус:** Готово к разработке  
**Приоритет:** Критичный (до 100% соответствия)

---

## 🔐 **ZITADEL API Доступ**

### Management API Endpoints
```bash
Base URL: https://edge.2gc.ru
Management API: https://edge.2gc.ru/management/v1
Auth API: https://edge.2gc.ru/auth/v1
Token Endpoint: https://edge.2gc.ru/oauth/v2/token
JWKS URI: https://edge.2gc.ru/oauth/v2/keys
Introspection: https://edge.2gc.ru/oauth/v2/introspect
```

### Authentication Tokens
```bash
# Personal Access Token (Admin права)
ZITADEL_ADMIN_TOKEN="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"

# Service User Credentials
SERVICE_USER_CLIENT_ID="339537393138794797"
SERVICE_USER_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"

# Service User JSON Key (located in ~/.zitadel/service-key.json)
SERVICE_USER_JSON_KEY_PATH="/home/ubuntu/.zitadel/service-key.json"
```

### Organization & Project IDs
```bash
ORGANIZATION_ID="339340397199294575"  # ZITADEL org
PROJECT_ID="339537327304999215"       # cloudbridge-relay project
SERVICE_USER_ID="339533564527771946"  # cloudbridge-admin
```

---

## 🌐 **Созданные OIDC Приложения**

### API Application (Back-channel)
```json
{
  "app_id": "339537393122017581",
  "client_id": "339537393138794797", 
  "client_secret": "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh",
  "name": "cloudbridge-relay-api",
  "type": "Web Application",
  "grant_types": ["client_credentials", "authorization_code"],
  "redirect_uris": [
    "https://b1.2gc.space/auth/callback",
    "https://relay.2gc.space/auth/callback",
    "http://localhost:8080/auth/callback"
  ]
}
```

### SPA Application (Front-channel)  
```json
{
  "app_id": "339537407432982831",
  "client_id": "339537407433048367",
  "name": "cloudbridge-relay-spa", 
  "type": "Single Page Application",
  "auth_method": "PKCE",
  "grant_types": ["authorization_code", "device_code"],
  "redirect_uris": [
    "https://relay.2gc.space/ui/callback",
    "http://localhost:3000/callback",
    "http://localhost:8080/callback", 
    "urn:ietf:wg:oauth:2.0:oob"
  ]
}
```

---

## 🚨 **ЗАДАЧИ ДЛЯ РАЗРАБОТЧИКА**

### 🔴 Критичные (обязательные):

#### 1. Настроить Custom Scopes в проекте
```bash
# API вызов для добавления scopes
curl -X POST "https://edge.2gc.ru/management/v1/projects/{PROJECT_ID}/roles" \
-H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "key": "relay.api",
  "display_name": "Relay API Access",
  "group": "api"
}'

curl -X POST "https://edge.2gc.ru/management/v1/projects/{PROJECT_ID}/roles" \
-H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "key": "relay.k8s", 
  "display_name": "Kubernetes Access",
  "group": "kubernetes"
}'
```

#### 2. Включить Token Exchange для mTLS
```bash
# Обновить приложение для поддержки Token Exchange
curl -X PUT "https://edge.2gc.ru/management/v1/projects/{PROJECT_ID}/apps/{APP_ID}/oidc_config" \
-H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "additional_origins": ["https://b1.2gc.space", "https://*.relay.2gc.space"],
  "allowed_origins": ["https://b1.2gc.space", "https://*.relay.2gc.space"],
  "skip_native_app_success_page": true
}'
```

#### 3. Реализовать kubectl exec-plugin (Go)
**Архитектура готова в:** `cmd/kubectl-zitadel/main.go`

**Требуется реализовать:**
- Device Code Flow для CLI
- PKCE Flow для браузерного логина
- Automatic refresh token handling
- ExecCredential v1beta1 output

**Пример структуры:**
```go
type ExecCredential struct {
    TypeMeta metav1.TypeMeta `json:",inline"`
    Status   *ExecCredentialStatus `json:"status,omitempty"`
}

type ExecCredentialStatus struct {
    Token               string      `json:"token,omitempty"`
    ExpirationTimestamp *metav1.Time `json:"expirationTimestamp,omitempty"`
}
```

#### 4. Настроить Custom Claims в JWT
```bash
# Добавить custom claims в токены проекта
curl -X POST "https://edge.2gc.ru/management/v1/projects/{PROJECT_ID}/apps/{APP_ID}/keys" \
-H "Authorization: Bearer $ZITADEL_ADMIN_TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "type": "KEY_TYPE_JSON",
  "expiration_date": "2029-12-31T23:59:59Z"
}'
```

---

## 📁 **Готовые конфигурационные файлы**

### Конфигурации для использования:
```bash
# Финальная конфигурация
/home/ubuntu/cloudbridge-relay-installer/configs/zitadel-final-config.yaml

# Scopes и Claims модель  
/home/ubuntu/cloudbridge-relay-installer/configs/zitadel-scopes-claims.yaml

# Kubernetes секреты
kubectl get secret cloudbridge-zitadel -n cloudbridge -o yaml

# mTLS сертификаты
/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/
```

### Service User креды (готовы к использованию):
```bash
# Загрузить переменные окружения
source /home/ubuntu/cloudbridge-relay-installer/configs/service-user-credentials.env

# Или использовать напрямую
CLIENT_ID="339537393138794797"
CLIENT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
PAT="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"
```

---

## 🧪 **Примеры API вызовов**

### Получение токена для тестирования
```bash
# Client Credentials Grant
curl -X POST "https://edge.2gc.ru/oauth/v2/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  -d "scope=openid profile email"

# Management API вызов
curl -H "Authorization: Bearer $PAT" \
  "https://edge.2gc.ru/management/v1/projects/$PROJECT_ID"
```

### JWKS проверка
```bash
curl "https://edge.2gc.ru/oauth/v2/keys" | jq .
```

### Token Introspection
```bash
curl -X POST "https://edge.2gc.ru/oauth/v2/introspect" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "token=YOUR_ACCESS_TOKEN"
```

---

## 📖 **Документация для разработчика**

### Созданные руководства:
- `docs/zitadel/SETUP.md` - Полная документация (27k строк)
- `docs/zitadel/SERVICE_USER_PERMISSIONS.md` - Права доступа
- `docs/NEXT_STEPS.md` - План развития
- `docs/COMPLIANCE_CHECK.md` - Анализ соответствия

### ZITADEL API Reference:
- [Management API](https://zitadel.com/docs/apis/resources/mgmt/management-service)
- [OIDC Endpoints](https://zitadel.com/docs/apis/openidoauth/endpoints)  
- [Service Users](https://zitadel.com/docs/guides/integrate/service-users)

---

## ⚙️ **Технические детали**

### Kubernetes Integration
```bash
# Namespace для развертывания
NAMESPACE="cloudbridge"

# RBAC уже настроен
kubectl get clusterrole cloudbridge-cluster-admin
kubectl get clusterrolebinding cloudbridge-cluster-admin-binding

# Секреты доступны
kubectl get secret cloudbridge-zitadel -n cloudbridge
```

### mTLS Certificates (готовы)
```bash
CA_CERT="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/cloudbridge-relay-ca-cert.pem"
CLIENT_CERT="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/relay-client-cert.pem" 
CLIENT_KEY="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/relay-client-key.pem"
SERVER_CERT="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/relay-server-cert.pem"
```

### PostgreSQL доступ (если нужен)
```bash
PG_PASSWORD=$(kubectl get secret -n zitadel pg-postgresql -o jsonpath='{.data.postgres-password}' | base64 -d)
kubectl exec -n zitadel pg-postgresql-0 -- env PGPASSWORD="$PG_PASSWORD" psql -U postgres -d zitadel
```

---

## ✅ **Acceptance Criteria для проверки**

После реализации должны работать:

### 1. gRPC с JWT токенами
```bash
grpcurl -authority relay.2gc.space:8444 --insecure \
  -H "Authorization: Bearer JWT_TOKEN" list
```

### 2. kubectl через exec-plugin  
```bash
kubectl get ns --context=relay
# Должно работать через ~/.kube/config с exec section
```

### 3. Логи с tenant_id/client_id
Логи access-denied должны содержать custom claims

### 4. Метрики в Grafana
- Login RPS
- Error rate <1%
- P95 latency при нагрузке RPS ≥ 200

---

## 🏃‍♂️ **Быстрый старт для разработчика**

### 1. Подготовить среду
```bash
cd /home/ubuntu/cloudbridge-relay-installer
source configs/service-user-credentials.env
```

### 2. Проверить доступ
```bash
curl -H "Authorization: Bearer $PAT" \
  "https://edge.2gc.ru/management/v1/orgs/me"
```

### 3. Начать разработку
```bash
# Склонировать kubectl plugin template
mkdir -p cmd/kubectl-zitadel
# Реализовать Go код по примеру в финальной конфигурации
```

---

## 💪 **Что уже готово (не нужно разрабатывать):**

✅ **ZITADEL развернут и работает**  
✅ **PostgreSQL с persistent данными**  
✅ **OIDC приложения созданы через Management API**  
✅ **Service User с admin правами**  
✅ **mTLS сертификаты сгенерированы**  
✅ **Kubernetes RBAC настроен**  
✅ **Мониторинг (Prometheus/Grafana) готов**  
✅ **Исчерпывающая документация**

## 🎯 **Что нужно доработать (3-5 дней):**

🔴 **kubectl exec-plugin Go код**  
🔴 **Custom scopes в ZITADEL проекте**  
🔴 **Token Exchange для mTLS**  
🔴 **Custom claims в JWT**  
🔴 **Acceptance тесты**

---

## 📞 **Поддержка**

Все файлы, конфигурации и документация находятся в:
```
/home/ubuntu/cloudbridge-relay-installer/
```

**Передано разработчику:** $(date)  
**Приоритет:** Критичный  
**Время на доработку:** 3-5 дней  
**Цель:** 100% соответствие техническому заданию
