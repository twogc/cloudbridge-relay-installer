# ZITADEL CloudBridge Relay Setup Guide

Пошаговая инструкция по настройке ZITADEL для интеграции с CloudBridge Relay.

## Содержание

1. [Базовая настройка Zitadel](#1-базовая-настройка-zitadel)
2. [Токены и scope/claim-model](#2-токены-и-scopeclaim-model)
3. [Service Account для Relay](#3-service-account-для-relay)
4. [mTLS / JWT-Bearer flow](#4-mtls--jwt-bearer-flow)
5. [RBAC-карта Kubernetes](#5-rbac-карта-kubernetes)
6. [kubectl ExecCredential plugin](#6-kubectl-execcredential-plugin)
7. [Audit & Compliance](#7-audit--compliance)
8. [Observability метрики](#8-observability-метрики)
9. [Disaster Recovery](#9-disaster-recovery)
10. [Acceptance criteria](#10-acceptance-criteria)

## Архитектура решения

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   End Users     │    │  kubectl plugin │    │   Relay API     │
│                 │◄──►│  (PKCE flow)    │◄──►│  (JWT Bearer)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
               ┌─────────────────────────────────┐
               │         ZITADEL IDP            │
               │  • Organization: cloudbridge   │
               │  • Project: cloudbridge-relay  │
               │  • Scopes: relay.api,relay.k8s │
               │  • Claims: tenant_id,roles     │
               └─────────────────────────────────┘
                                 │
                                 ▼
               ┌─────────────────────────────────┐
               │      Kubernetes RBAC           │
               │  • ClusterRoleBinding          │
               │  • RoleBinding per Tenant      │
               │  • Policy Engine (OPA)         │
               └─────────────────────────────────┘
```

## 1. Базовая настройка Zitadel

### 1.1. Создание организации `cloudbridge`

#### Через Web UI:
1. Войти в ZITADEL Console: https://idp.example.com
2. Перейти в Organizations → Create Organization
3. Name: `cloudbridge`
4. Domain: `cloudbridge.example.com` (optional)

#### Через API:
```bash
# Создание организации через Management API
curl -X POST "https://idp.example.com/management/v1/orgs" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudbridge",
    "domain": {
      "domain": "cloudbridge.example.com",
      "type": "DOMAIN_TYPE_PRIMARY"
    }
  }'
```

### 1.2. Создание проекта `cloudbridge-relay`

#### Back-channel OIDC Client (Confidential)
```bash
# Создание проекта
curl -X POST "https://idp.example.com/management/v1/projects" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudbridge-relay",
    "project_role_assertion": true,
    "project_role_check": true
  }'

# Создание Back-channel приложения
curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudbridge-relay-api",
    "redirect_uris": [],
    "response_types": [],
    "grant_types": [
      "OIDC_GRANT_TYPE_CLIENT_CREDENTIALS",
      "OIDC_GRANT_TYPE_AUTHORIZATION_CODE",
      "OIDC_GRANT_TYPE_TOKEN_EXCHANGE"
    ],
    "app_type": "OIDC_APP_TYPE_NATIVE",
    "auth_method_type": "OIDC_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT",
    "version": "OIDC_VERSION_1_0"
  }'
```

#### Front-channel SPA Client (Public)
```bash
# Создание Front-channel приложения  
curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cloudbridge-relay-ui",
    "redirect_uris": [
      "http://localhost:8080/auth/callback",
      "https://relay.2gc.space/auth/callback"
    ],
    "response_types": [
      "OIDC_RESPONSE_TYPE_CODE"
    ],
    "grant_types": [
      "OIDC_GRANT_TYPE_AUTHORIZATION_CODE"
    ],
    "app_type": "OIDC_APP_TYPE_USER_AGENT",
    "auth_method_type": "OIDC_AUTH_METHOD_TYPE_NONE",
    "version": "OIDC_VERSION_1_0",
    "dev_mode": false,
    "access_token_type": "OIDC_TOKEN_TYPE_JWT",
    "skip_native_app_success_page": true
  }'
```

### 1.3. Issuer URL и JWKS endpoint

#### Конфигурация
- **Issuer URL**: `https://idp.example.com`
- **JWKS URI**: `https://idp.example.com/oauth/v2/keys`
- **Authorization**: `https://idp.example.com/oauth/v2/authorize`
- **Token**: `https://idp.example.com/oauth/v2/token`

#### Проверка JWKS rotation
```bash
# Проверка текущих ключей
curl -s https://idp.example.com/oauth/v2/keys | jq '.keys[].kid'

# ZITADEL автоматически ротирует ключи каждые 6 месяцев
# Настройка уведомлений о ротации через WebHook
```

## 2. Токены и scope/claim-model

### 2.1. Custom Scopes

#### Создание scope `relay.api`
```bash
curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/roles" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "relay.api",
    "display_name": "Relay API Access",
    "group": "api"
  }'
```

#### Создание scope `relay.k8s`
```bash
curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/roles" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "key": "relay.k8s",
    "display_name": "Kubernetes Proxy Access", 
    "group": "k8s"
  }'
```

### 2.2. Custom Claims Configuration

#### JWT Template для Access Token
```json
{
  "aud": ["cloudbridge-relay"],
  "iss": "https://idp.example.com",
  "sub": "{{.UserID}}",
  "tenant_id": "{{.User.Metadata.tenant_id}}",
  "client_id": "{{.User.Metadata.client_id}}",  
  "roles": "{{.ProjectRoles}}",
  "scopes": "{{.Scopes}}",
  "exp": "{{.ExpirationTime}}",
  "iat": "{{.IssuedAtTime}}",
  "auth_time": "{{.AuthTime}}",
  "session_id": "{{.SessionID}}",
  "audit_session_id": "{{.User.Metadata.audit_session_id}}"
}
```

### 2.3. Token Lifetimes

#### Настройка через Console
- **Access Token**: 15 минут
- **Refresh Token**: 8 часов  
- **Service Token** (mTLS): 5 минут
- **ID Token**: 1 час

```bash
# Обновление настроек приложения
curl -X PUT "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc/${APP_ID}" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "access_token_lifetime": "900s",
    "refresh_token_lifetime": "28800s",
    "id_token_lifetime": "3600s"
  }'
```

## 3. Service Account для Relay

### 3.1. Создание service user `relay-server`

```bash
# Создание машинного пользователя
curl -X POST "https://idp.example.com/management/v1/users/machine" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "user_name": "relay-server",
    "name": "CloudBridge Relay Service",
    "description": "Service account for CloudBridge Relay API access",
    "access_token_type": "ACCESS_TOKEN_TYPE_JWT"
  }'

# Назначение ролей
curl -X POST "https://idp.example.com/management/v1/users/${USER_ID}/grants" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "project_id": "${PROJECT_ID}",
    "role_keys": ["relay.api", "relay.k8s"]
  }'
```

### 3.2. Long-lived Client Credentials

```bash
# Создание ключа для машинного пользователя
curl -X POST "https://idp.example.com/management/v1/users/${USER_ID}/keys" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "KEY_TYPE_JSON",
    "expiration_date": "2029-12-31T23:59:59Z"
  }'

# Сохранение ключа в Kubernetes Secret
kubectl create secret generic relay-zitadel-key \
  -n cloudbridge \
  --from-file=key.json=/tmp/relay-server-key.json
```

### 3.3. Trusted Origins

```bash
# Добавление доверенных источников
curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc/${APP_ID}/origins" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "https://b1.2gc.space"
  }'

curl -X POST "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc/${APP_ID}/origins" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "https://*.relay.2gc.space"
  }'
```

## 4. mTLS / JWT-Bearer flow

### 4.1. Token Exchange Configuration

```bash
# Включение Token Exchange для relay-server
curl -X PUT "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc/${APP_ID}" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "grant_types": [
      "OIDC_GRANT_TYPE_CLIENT_CREDENTIALS",
      "OIDC_GRANT_TYPE_TOKEN_EXCHANGE"
    ],
    "auth_method_type": "OIDC_AUTH_METHOD_TYPE_PRIVATE_KEY_JWT"
  }'
```

### 4.2. Client Certificate Generation

```bash
# Генерация CA для клиентских сертификатов
openssl genrsa -out relay-ca-key.pem 4096
openssl req -new -x509 -key relay-ca-key.pem -out relay-ca.pem -days 3650 \
  -subj "/CN=CloudBridge Relay CA"

# Генерация клиентского сертификата для relay-server
openssl genrsa -out relay-server-key.pem 2048
openssl req -new -key relay-server-key.pem -out relay-server.csr \
  -subj "/CN=relay-server"

# Создание конфигурации с SAN
cat > relay-server.conf <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = relay-server
DNS.2 = relay.local
DNS.3 = relay.2gc.space
EOF

# Подпись сертификата
openssl x509 -req -in relay-server.csr -CA relay-ca.pem -CAkey relay-ca-key.pem \
  -CAcreateserial -out relay-server.pem -days 365 -extensions v3_req \
  -extfile relay-server.conf
```

### 4.3. RFC 8705 mTLS Implementation

Пример кода для получения JWT по mTLS:

```go
// pkg/auth/mtls_jwt.go
package auth

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"
    "strings"
)

type MTLSTokenClient struct {
    issuerURL    string
    clientCert   tls.Certificate
    caCert       *x509.Certificate
    clientID     string
}

func NewMTLSTokenClient(issuerURL, certFile, keyFile, caFile, clientID string) (*MTLSTokenClient, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("load client cert: %w", err)
    }

    caCertData, err := ioutil.ReadFile(caFile)
    if err != nil {
        return nil, fmt.Errorf("load CA cert: %w", err)
    }

    caCert, err := x509.ParseCertificate(caCertData)
    if err != nil {
        return nil, fmt.Errorf("parse CA cert: %w", err)
    }

    return &MTLSTokenClient{
        issuerURL:  issuerURL,
        clientCert: cert,
        caCert:     caCert,
        clientID:   clientID,
    }, nil
}

func (c *MTLSTokenClient) GetAccessToken(scopes []string) (string, error) {
    // Создание HTTP клиента с mTLS
    caCertPool := x509.NewCertPool()
    caCertPool.AddCert(c.caCert)

    client := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                Certificates: []tls.Certificate{c.clientCert},
                RootCAs:      caCertPool,
                ServerName:   "idp.example.com",
            },
        },
    }

    // RFC 8705: client_credentials grant с mTLS
    data := url.Values{
        "grant_type": {"client_credentials"},
        "client_id":  {c.clientID},
        "scope":      {strings.Join(scopes, " ")},
    }

    resp, err := client.PostForm(c.issuerURL+"/oauth/v2/token", data)
    if err != nil {
        return "", fmt.Errorf("token request: %w", err)
    }
    defer resp.Body.Close()

    // Парсинг ответа...
    // Реализация парсинга JSON ответа
    return "", nil
}
```

## 5. RBAC-карта Kubernetes

### 5.1. Групповое сопоставление ролей

#### ClusterRoleBinding для cluster-admin
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cloudbridge-cluster-admin
subjects:
- kind: User
  name: system:serviceaccount:cloudbridge:relay-server
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
---
# Правило для пользователей с ролью cluster-admin в JWT
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cloudbridge-admin-users
subjects:
- kind: User
  name: "oidc:cloudbridge-relay:roles=cluster-admin"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole 
  name: cluster-admin
  apiGroup: rbac.authorization.k8s.io
```

#### RoleBinding для tenant-reader
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tenant-reader-{{.TenantID}}
  namespace: "ns-{{.TenantID}}"
subjects:
- kind: User
  name: "oidc:cloudbridge-relay:tenant_id={{.TenantID}}:roles=tenant-reader"
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: view
  apiGroup: rbac.authorization.k8s.io
```

### 5.2. OPA/Rego Policy (Optional)

```rego
# policies/relay_authz.rego
package relay.authz

import future.keywords.if
import future.keywords.in

# Разрешить доступ если пользователь имеет cluster-admin роль
allow if {
    "cluster-admin" in input.jwt.roles
}

# Разрешить доступ к namespace если пользователь имеет tenant-reader роль 
# и tenant_id совпадает с namespace
allow if {
    "tenant-reader" in input.jwt.roles
    input.request.namespace == sprintf("ns-%s", [input.jwt.tenant_id])
    input.request.verb in ["get", "list", "watch"]
}

# Запретить доступ к системным namespace для обычных пользователей
deny if {
    not "cluster-admin" in input.jwt.roles
    input.request.namespace in ["kube-system", "kube-public", "kube-node-lease"]
}

# Audit logging
audit_log := {
    "timestamp": time.now_ns(),
    "user_id": input.jwt.sub,
    "tenant_id": input.jwt.tenant_id,
    "client_id": input.jwt.client_id,
    "action": input.request.verb,
    "resource": input.request.resource,
    "namespace": input.request.namespace,
    "allowed": allow,
    "session_id": input.jwt.audit_session_id
}
```

## 6. kubectl ExecCredential plugin

Пример минимальной реализации плагина:

```go
// cmd/kubectl-zitadel/main.go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "time"

    "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
    "golang.org/x/oauth2"
)

type Config struct {
    IssuerURL    string `json:"issuer_url"`
    ClientID     string `json:"client_id"`
    ClientSecret string `json:"client_secret"`
    Scopes       []string `json:"scopes"`
}

func main() {
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Println("kubectl-zitadel v1.0.0")
        return
    }

    config, err := loadConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
        os.Exit(1)
    }

    token, err := getToken(config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error getting token: %v\n", err)
        os.Exit(1)
    }

    response := &v1beta1.ExecCredential{
        TypeMeta: v1beta1.TypeMeta{
            APIVersion: "client.authentication.k8s.io/v1beta1",
            Kind:       "ExecCredential",
        },
        Status: &v1beta1.ExecCredentialStatus{
            Token:               token.AccessToken,
            ExpirationTimestamp: &v1beta1.Time{Time: token.Expiry},
        },
    }

    json.NewEncoder(os.Stdout).Encode(response)
}

func loadConfig() (*Config, error) {
    // Загрузка конфигурации из ~/.kube/zitadel-config.json
    return &Config{}, nil
}

func getToken(config *Config) (*oauth2.Token, error) {
    // Реализация PKCE flow для браузерной аутентификации
    // или Device Code flow для CLI
    return &oauth2.Token{}, nil
}
```

## 7. Audit & Compliance

### 7.1. Экспорт в Loki через WebHook

```yaml
# k8s/monitoring/loki-webhook.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zitadel-audit-webhook
  namespace: monitoring
spec:
  replicas: 2
  selector:
    matchLabels:
      app: zitadel-audit-webhook
  template:
    metadata:
      labels:
        app: zitadel-audit-webhook
    spec:
      containers:
      - name: webhook
        image: grafana/promtail:latest
        args:
        - -config.file=/etc/promtail/promtail.yml
        volumeMounts:
        - name: config
          mountPath: /etc/promtail
        env:
        - name: LOKI_URL
          value: "http://loki:3100"
      volumes:
      - name: config
        configMap:
          name: promtail-zitadel-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-zitadel-config
  namespace: monitoring
data:
  promtail.yml: |
    server:
      http_listen_port: 9080
      grpc_listen_port: 0

    clients:
      - url: ${LOKI_URL}/loki/api/v1/push

    scrape_configs:
      - job_name: zitadel-audit
        webhook:
          webhook_config:
            url: 0.0.0.0:9080/webhook
        relabel_configs:
          - source_labels: [__webhook_json_user_id]
            target_label: user_id
          - source_labels: [__webhook_json_tenant_id]  
            target_label: tenant_id
          - source_labels: [__webhook_json_audit_session_id]
            target_label: trace_id
```

### 7.2. Audit Session ID в JWT Claims

Настройка добавления `audit_session_id` в JWT:

```bash
# Обновление шаблона JWT через Management API
curl -X PUT "https://idp.example.com/management/v1/projects/${PROJECT_ID}/apps/oidc/${APP_ID}/jwt_profile" \
  -H "Authorization: Bearer ${ZITADEL_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "jwt": {
      "custom_claims": [
        {
          "key": "audit_session_id", 
          "value": "{{.SessionID}}"
        },
        {
          "key": "tenant_id",
          "value": "{{.User.Metadata.tenant_id}}"
        },
        {
          "key": "client_id", 
          "value": "{{.User.Metadata.client_id}}"
        }
      ]
    }
  }'
```

## 8. Observability метрики

### 8.1. Prometheus Metrics Export

```yaml
# k8s/monitoring/zitadel-servicemonitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: zitadel-metrics
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: zitadel
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
```

Кастомные метрики ZITADEL:
- `zitadel_oauth_requests_total{method, endpoint, status}`
- `zitadel_token_issued_total{grant_type, client_id}`  
- `zitadel_login_latency_seconds{method}`
- `zitadel_token_exchange_total{from_type, to_type}`

### 8.2. Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Relay-Auth",
    "panels": [
      {
        "title": "Login RPS",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(zitadel_oauth_requests_total{endpoint=\"/oauth/v2/authorize\"}[5m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "singlestat", 
        "targets": [
          {
            "expr": "rate(zitadel_oauth_requests_total{status=~\"4..|5..\"}[5m]) / rate(zitadel_oauth_requests_total[5m]) * 100"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, zitadel_login_latency_seconds)"
          }
        ]
      },
      {
        "title": "Token Exchange",
        "type": "graph", 
        "targets": [
          {
            "expr": "rate(zitadel_token_exchange_total[5m])"
          }
        ]
      }
    ]
  }
}
```

## 9. Disaster Recovery

### 9.1. PostgreSQL Backup в VK Cloud S3

```yaml
# k8s/zitadel/backup/s3-cronjob.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: zitadel-db-backup
  namespace: zitadel
spec:
  schedule: "0 2 * * *"  # Каждый день в 2:00
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:15
            command:
            - /bin/bash
            - -c
            - |
              BACKUP_FILE="zitadel-backup-$(date +%Y%m%d_%H%M%S).sql"
              pg_dump -h pg-postgresql.zitadel.svc.cluster.local \
                      -U postgres -d zitadel > /tmp/$BACKUP_FILE
              
              # Загрузка в VK Cloud S3
              s3cmd put /tmp/$BACKUP_FILE s3://cloudbridge-backups/zitadel/
              
              # Очистка локальных файлов старше 7 дней
              find /tmp -name "zitadel-backup-*.sql" -mtime +7 -delete
            env:
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: pg-postgresql
                  key: postgres-password
            - name: S3_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: vk-cloud-s3
                  key: access_key
            - name: S3_SECRET_KEY  
              valueFrom:
                secretKeyRef:
                  name: vk-cloud-s3
                  key: secret_key
            volumeMounts:
            - name: s3-config
              mountPath: /.s3cfg
              subPath: s3cfg
          volumes:
          - name: s3-config
            configMap:
              name: s3-config
          restartPolicy: OnFailure
```

### 9.2. KMS Key Management

```bash
# Генерация мастер-ключа с использованием VK Cloud KMS
curl -X POST "https://kms.api.cloud.vk.com/v1/keys" \
  -H "X-Auth-Token: ${VK_CLOUD_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "zitadel-masterkey",
    "key_spec": "AES_256",
    "usage": "ENCRYPT_DECRYPT"
  }'

# Шифрование masterkey
curl -X POST "https://kms.api.cloud.vk.com/v1/keys/${KEY_ID}/encrypt" \
  -H "X-Auth-Token: ${VK_CLOUD_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "plaintext": "'$(base64 -w 0 <<< $(kubectl get secret zitadel-masterkey -n zitadel -o jsonpath='{.data.masterkey}' | base64 -d))'"
  }'
```

## 10. Acceptance Criteria

### 10.1. gRPC API Test

```bash
# Получение токена для тестирования
TOKEN=$(curl -s -X POST https://idp.example.com/oauth/v2/token \
  -d "grant_type=client_credentials" \
  -d "client_id=${CLIENT_ID}" \
  -d "client_secret=${CLIENT_SECRET}" \
  -d "scope=relay.api" | jq -r .access_token)

# Проверка gRPC сервиса
grpcurl -authority relay.2gc.space:8444 \
  -H "Authorization: Bearer ${TOKEN}" \
  --insecure \
  list

# Ожидаемый результат: список gRPC сервисов
```

### 10.2. kubectl exec-plugin Test

```bash
# Настройка kubeconfig для использования exec-plugin
cat >> ~/.kube/config <<EOF
users:
- name: zitadel-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-zitadel
      env:
      - name: ZITADEL_ISSUER
        value: https://idp.example.com
      - name: ZITADEL_CLIENT_ID  
        value: ${CLIENT_ID}
EOF

# Проверка доступа
kubectl get ns --context=relay

# Ожидаемый результат: список namespaces
```

### 10.3. Access Denied Logging

```bash
# Попытка неавторизованного доступа
curl -H "Authorization: Bearer invalid_token" \
  https://relay.2gc.space/api/v1/clusters

# Проверка логов в Loki
curl -G "http://loki:3100/loki/api/v1/query" \
  --data-urlencode 'query={app="cloudbridge-relay"} |= "access denied"' | \
  jq '.data.result[0].values[-1][1]' | \
  jq '. | contains("tenant_id") and contains("client_id")'

# Ожидаемый результат: true
```

### 10.4. Load Testing

```bash
# Нагрузочное тестирование login endpoint
./load-test.sh https://idp.example.com/oauth/v2/authorize 200 60s

# Проверка error rate в Grafana
curl -s "http://grafana:3000/api/datasources/proxy/1/api/v1/query" \
  -d 'query=rate(zitadel_oauth_requests_total{status=~"4..|5.."}[5m]) / rate(zitadel_oauth_requests_total[5m]) * 100' | \
  jq '.data.result[0].value[1] | tonumber < 1'

# Ожидаемый результат: true (error rate < 1%)
```

## Быстрый старт

```bash
# 1. Клонирование и переход в директорию
cd /home/ubuntu/cloudbridge-relay-installer

# 2. Запуск автоматической настройки
./scripts/zitadel-setup.sh

# 3. Проверка развертывания
kubectl get pods -n zitadel
curl -k https://idp.example.com/.well-known/openid-configuration

# 4. Тестирование интеграции  
./scripts/integration-test.sh
```

## Дополнительные ресурсы

- [ZITADEL Documentation](https://zitadel.com/docs)
- [RFC 8705: OAuth 2.0 Mutual-TLS Client Authentication](https://tools.ietf.org/html/rfc8705)
- [Kubernetes ExecCredential](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins)
- [OPA/Rego Language](https://www.openpolicyagent.org/docs/latest/policy-language/)

---

*Последнее обновление: $(date)*
