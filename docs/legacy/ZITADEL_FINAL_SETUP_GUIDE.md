# CloudBridge Relay + ZITADEL - Final Setup Guide

## **Текущий статус: 95% готово**

Все компоненты настроены. Остается только создать OIDC приложения через веб-интерфейс.

### **Имеющиеся учетные данные**

#### **Admin User (веб-интерфейс)**
- **URL:** https://edge.2gc.ru/ui/console/
- **Username:** `admin` или `admin@cloudbridge.local`
- **Password:** `CloudBridge123!`

####  **Service User JSON Key**
```json
{
  "type": "application",
  "keyId": "339535653324718382",
  "appId": "339340397199491183", 
  "clientId": "339340399631990895",
  "key": "-----BEGIN RSA PRIVATE KEY----- ... -----END RSA PRIVATE KEY-----"
}
```

#### 🎫 **Personal Access Token**
- **Token:** `1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g`
- **ID:** `339535554926346543`

---

## **Шаги для завершения настройки**

### Шаг 1: Создайте OIDC приложения

Войдите в ZITADEL консоль: https://edge.2gc.ru/ui/console/

#### 📱 **Приложение 1: CloudBridge Relay API**
1. **Projects** → **ZITADEL** → **Applications** → **New**
2. **Настройки:**
   ```
   Name: cloudbridge-relay-api
   Type: Web Application
   Auth Method: Client Secret (Basic)
   Grant Types: Authorization Code + Client Credentials
   Response Types: Code
   Access Token Type: JWT
   Roles in Token: Enabled
   ```
3. **Redirect URIs:**
   ```
   https://b1.2gc.space/auth/callback
   https://relay.2gc.space/auth/callback  
   http://localhost:8080/auth/callback
   ```

#### **Приложение 2: CloudBridge Relay SPA**
1. **Applications** → **New** 
2. **Настройки:**
   ```
   Name: cloudbridge-relay-spa
   Type: Single Page Application
   Auth Method: PKCE
   Grant Types: Authorization Code + Device Code
   Response Types: Code
   Access Token Type: JWT
   Dev Mode: Enabled
   ```
3. **Redirect URIs:**
   ```
   https://relay.2gc.space/ui/callback
   http://localhost:3000/callback
   http://localhost:8080/callback
   urn:ietf:wg:oauth:2.0:oob
   ```

### Шаг 2: Скопируйте Client ID и Secret

После создания приложений запишите:
- **API App Client ID:** `_____________`
- **API App Client Secret:** `_____________`
- **SPA App Client ID:** `_____________`

### Шаг 3: Финализация конфигурации

```bash
cd /home/ubuntu/cloudbridge-relay-installer

# Запустите финальный скрипт с полученными данными
./scripts/finalize-cloudbridge-setup.sh \
  <API_CLIENT_ID> \
  <API_CLIENT_SECRET> \
  <SPA_CLIENT_ID>
```

---

## 🛠️ **Готовые компоненты**

### **Уже настроены:**
- **ZITADEL Core:** 3/3 pods Running
- **PostgreSQL:** база данных с пользователями  
- **SSL/TLS:** сертификаты и Ingress
- **mTLS:** CA, server, client сертификаты
- **RBAC:** Kubernetes роли и привязки
- **Service User:** Machine User с ключами
- **Monitoring:** Prometheus/Grafana готовы
- **Documentation:** полная документация

### **Конфигурационные файлы:**
```
configs/
├── zitadel-final-config.yaml           # Финальная конфигурация
├── zitadel-scopes-claims.yaml          # Custom scopes/claims
├── oidc-app-1-api.json                 # Шаблон API приложения
└── oidc-app-2-spa.json                 # Шаблон SPA приложения

certs/relay-mtls/
├── cloudbridge-relay-ca-cert.pem       # CA сертификат
├── relay-server-cert.pem               # Серверный сертификат
├── relay-client-cert.pem               # Клиентский сертификат
└── *.p12                               # PKCS#12 файлы

scripts/
├── finalize-cloudbridge-setup.sh       # Финальная настройка
├── setup-zitadel-scopes.sh            # Настройка scopes
└── reset-admin-password.sh             # Сброс паролей
```

### **Kubernetes секреты:**
```bash
# Проверьте созданные секреты
kubectl get secrets -n cloudbridge
kubectl get secrets -n zitadel
```

---

## 🧪 **Тестирование после настройки**

### 1. Проверка OIDC endpoints
```bash
curl -s https://edge.2gc.ru/.well-known/openid_configuration | jq .
```

### 2. Тест Client Credentials flow
```bash
curl -X POST https://edge.2gc.ru/oauth/v2/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=YOUR_API_CLIENT_ID" \
  -d "client_secret=YOUR_API_CLIENT_SECRET" \
  -d "scope=openid profile relay.api"
```

### 3. Тест gRPC с JWT
```bash
# Получите JWT токен и тестируйте gRPC API
grpcurl -authority relay.2gc.space:8444 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --insecure list
```

### 4. Тест kubectl exec-plugin
```bash
kubectl get namespaces --context=relay
```

---

## **Мониторинг и логи**

### Prometheus метрики:
- `zitadel_oauth_requests_total`
- `zitadel_token_issued_total` 
- `zitadel_login_latency_seconds`

### Grafana Dashboard:
- **Login RPS:** количество входов в секунду
- **Error Rate:** процент ошибок аутентификации
- **P95 Latency:** 95-й процентиль задержки login

### Audit логи:
Экспорт в Loki через WebHook с `audit_session_id` в JWT claims.

---

## 🆘 **Troubleshooting**

### Проблема: "Not Found" в Management API
**Решение:** Используйте веб-интерфейс для создания приложений

### Проблема: JWT Bearer не работает  
**Решение:** Проверьте Service User ключ и endpoint'ы

### Проблема: gRPC недоступен
**Решение:** Проверьте mTLS сертификаты и Ingress

### Проблема: kubectl plugin не работает
**Решение:** Проверьте PKCE настройки SPA приложения

---

## **Результат**

После выполнения всех шагов у вас будет:

 **Полностью рабочая интеграция CloudBridge Relay + ZITADEL**
 **OAuth 2.0/OIDC аутентификация** для gRPC и Web UI  
 **mTLS защищенные соединения**
 **Kubernetes RBAC интеграция**
 **kubectl exec-plugin** для прозрачного входа
 **Audit логи и метрики** для мониторинга
 **Disaster recovery** план и документация

**Время выполнения:** 15-30 минут (после создания OIDC приложений)

**Документация обновлена:** $(date)
