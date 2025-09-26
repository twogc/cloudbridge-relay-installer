# CloudBridge Relay Kubernetes Proxy Setup

Пошаговое руководство по настройке CloudBridge Relay для проксирования kubectl запросов.

## Обзор архитектуры

```
kubectl → https://relay.2gc.space/kube/minikube/api/v1/pods → Kubernetes API Server
```

- **Аутентификация**: JWT/mTLS в Relay + заголовки (`X-Tenant-ID`, `X-Remote-User`)
- **Авторизация**: Policy-движок на уровне verbs/resources/namespaces
- **Транспорт**: HTTP/2 + WebSocket streams (exec/attach/logs/port-forward) 
- **Безопасность**: mTLS к API серверам, аудит всех запросов

## 1. Подготовка кластеров

### Создание Service Account для Relay

На каждом кластере создайте ServiceAccount с минимальными необходимыми правами:

```bash
# На каждом кластере (production, staging, dev)
kubectl create namespace cloudbridge-relay
kubectl create serviceaccount cloudbridge-relay-sa -n cloudbridge-relay
```

### RBAC для Relay ServiceAccount

```yaml
# rbac-relay.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cloudbridge-relay-access
rules:
# Core resources
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets", "namespaces", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Exec/attach/logs/port-forward
- apiGroups: [""]
  resources: ["pods/exec", "pods/attach", "pods/log", "pods/portforward"]
  verbs: ["get", "create"]
# Apps resources
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "daemonsets", "statefulsets"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Extensions/networking
- apiGroups: ["extensions", "networking.k8s.io"]
  resources: ["ingresses", "networkpolicies"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
# Proxy access for services/nodes  
- apiGroups: [""]
  resources: ["services/proxy", "nodes/proxy"]
  verbs: ["get", "create"]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cloudbridge-relay-binding
subjects:
- kind: ServiceAccount
  name: cloudbridge-relay-sa
  namespace: cloudbridge-relay
roleRef:
  kind: ClusterRole
  name: cloudbridge-relay-access
  apiGroup: rbac.authorization.k8s.io
```

```bash
kubectl apply -f rbac-relay.yaml
```

### Извлечение сертификатов

Для каждого кластера извлеките CA и создайте client certificate:

```bash
# Получение CA кластера
kubectl get secret -n cloudbridge-relay \
  $(kubectl get sa cloudbridge-relay-sa -n cloudbridge-relay -o jsonpath='{.secrets[0].name}') \
  -o jsonpath='{.data.ca\.crt}' | base64 -d > cluster-ca.crt

# Получение Service Account token
kubectl get secret -n cloudbridge-relay \
  $(kubectl get sa cloudbridge-relay-sa -n cloudbridge-relay -o jsonpath='{.secrets[0].name}') \
  -o jsonpath='{.data.token}' | base64 -d > cluster-token

# Альтернативно: создание client certificate (рекомендуется)
# Создайте CSR и подпишите его CA кластера
openssl genrsa -out cluster-client.key 2048
openssl req -new -key cluster-client.key -out cluster-client.csr \
  -subj "/CN=cloudbridge-relay/O=system:serviceaccounts:cloudbridge-relay"

# Подпишите CSR в кластере (или используйте kubeadm alpha certs)
kubectl certificate approve cloudbridge-relay-csr
kubectl get csr cloudbridge-relay-csr -o jsonpath='{.status.certificate}' | base64 -d > cluster-client.crt
```

## 2. Конфигурация CloudBridge Relay

### Структура конфигурации

```json
{
  "kubernetes": {
    "clusters": [
      {
        "id": "production", 
        "api_server_url": "https://k8s-prod.example.com:6443",
        "ca": "/etc/relay/clusters/production/ca.crt",
        "client_cert": "/etc/relay/clusters/production/client.crt",
        "client_key": "/etc/relay/clusters/production/client.key",
        "skip_tls_verify": false
      }
    ],
    "policies": [
      {
        "tenant_id": "dev-team",
        "groups": ["developers"],
        "rules": [
          {
            "name": "dev-full-access",
            "verbs": ["get", "list", "watch", "create", "update", "patch", "delete"],
            "api_groups": ["", "apps"],
            "resources": ["pods", "services", "deployments"],
            "namespaces": ["development", "testing"]
          }
        ]
      }
    ]
  }
}
```

### Монтирование сертификатов в Kubernetes

```yaml
# k8s-cluster-certs-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: k8s-cluster-production-certs
type: Opaque
data:
  ca.crt: <base64-encoded-ca.crt>
  client.crt: <base64-encoded-client.crt>  
  client.key: <base64-encoded-client.key>
---
apiVersion: v1
kind: Secret
metadata:
  name: k8s-cluster-staging-certs
type: Opaque  
data:
  ca.crt: <base64-encoded-ca.crt>
  client.crt: <base64-encoded-client.crt>
  client.key: <base64-encoded-client.key>
```

## 3. Развертывание с Helm

```bash
# Установка из локального репозитория
helm install cloudbridge-relay ./k8s \
  --values ./k8s/values-kube-proxy.yaml \
  --namespace cloudbridge \
  --create-namespace

# Проверка развертывания
kubectl get pods -n cloudbridge
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge

# Проверка метрик
kubectl port-forward svc/cloudbridge-relay 5551:5551 -n cloudbridge
curl http://localhost:5551/metrics | grep kube_proxy
```

## 4. Настройка kubectl

### Kubeconfig для доступа через Relay

```yaml
# ~/.kube/config-via-relay
apiVersion: v1
kind: Config
clusters:
- name: production-via-relay
  cluster:
    server: https://relay.2gc.space/kube/production
    insecure-skip-tls-verify: true  # или добавьте CA relay
- name: staging-via-relay  
  cluster:
    server: https://relay.2gc.space/kube/staging
    insecure-skip-tls-verify: true
contexts:
- name: production-via-relay
  context:
    cluster: production-via-relay
    user: dev-team-user
- name: staging-via-relay
  context:
    cluster: staging-via-relay
    user: dev-team-user
current-context: staging-via-relay
users:
- name: dev-team-user
  user:
    # Временно используем headers - позже добавим ExecCredential
    # В production интегрируйте с JWT/OAuth provider
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: echo
      args: 
      - '{"apiVersion":"client.authentication.k8s.io/v1beta1","kind":"ExecCredential","status":{"token":"dummy-token"}}'
```

### Тестирование подключения

```bash
# Установка переменных для принудительного WebSocket
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1

# Тестирование базовых команд
kubectl --kubeconfig ~/.kube/config-via-relay get pods -n development
kubectl --kubeconfig ~/.kube/config-via-relay logs deployment/app -n development

# Тестирование WebSocket streams
kubectl --kubeconfig ~/.kube/config-via-relay exec -it pod-name -n development -- /bin/bash
kubectl --kubeconfig ~/.kube/config-via-relay port-forward svc/service-name 8080:80 -n development
```

## 5. Мониторинг и аудит

### Prometheus метрики

```promql
# Общее количество запросов через kube proxy
sum(rate(kube_proxy_requests_total[5m])) by (cluster_id, verb, resource)

# Latency запросов
histogram_quantile(0.95, sum(rate(kube_proxy_request_duration_seconds_bucket[5m])) by (le, cluster_id))

# WebSocket connections
kube_ws_active_streams

# Authorization denials
sum(rate(kube_proxy_requests_total{code="403"}[5m])) by (cluster_id, resource)
```

### Аудит-логи

```json
{
  "level": "info",
  "msg": "kube-proxy-audit",
  "cluster_id": "production", 
  "tenant_id": "dev-team",
  "user_id": "alice",
  "verb": "exec",
  "resource": "pods",
  "namespace": "development",
  "allowed": false,
  "reason": "denied by explicit deny rule",
  "matched_rule": "deny-production",
  "duration": "1.2ms",
  "timestamp": "2023-12-07T10:30:00Z"
}
```

## 6. Безопасность

### Рекомендации

1. **mTLS везде**: Client certificates для всех кластеров
2. **Минимальные права**: ServiceAccount в кластере только с необходимыми RBAC
3. **Network policies**: Ограничение трафика к Relay
4. **Регулярная ротация**: Сертификаты и токены
5. **Аудит включен**: Все запросы логируются с полным контекстом

### Заголовки аутентификации

В production интегрируйте с IdP (Keycloak/Zitadel):

```bash
# Пример с OAuth2 Proxy
kubectl --kubeconfig config \
  --header "X-Tenant-ID: dev-team" \
  --header "X-Remote-User: alice@company.com" \
  --header "X-Remote-Group: developers,qa" \
  get pods
```

## 7. Troubleshooting

### Общие проблемы

```bash
# Проверка статуса Relay
curl -k https://relay.2gc.space/health

# Проверка подключения к кластеру  
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "kube-proxy"

# Проверка сертификатов
openssl x509 -in /etc/relay/clusters/production/client.crt -text -noout

# Проверка политик авторизации
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "forbidden"
```

### WebSocket debugging

```bash
# Force WebSocket for kubectl >= 1.29
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1

# Debug WebSocket handshake
kubectl --kubeconfig config --v=8 exec -it pod-name -- /bin/bash
```

## 8. ExecCredential v2 (будущее)

Подготовка к интеграции с OIDC:

```yaml
users:
- name: oidc-user
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: cloudbridge-client
      args:
      - kube
      - get-token  
      - --cluster=production
      - --oidc-issuer=https://auth.company.com
      - --client-id=kubectl-proxy
      env:
      - name: CLOUDBRIDGE_RELAY_URL
        value: https://relay.2gc.space
```

---

**Готово!** Kubectl proxy интегрирован и настроен для безопасного доступа через CloudBridge Relay с полным аудитом и гранулярными правами доступа.
