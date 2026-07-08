# CloudBridge Relay Kubectl Proxy - Quick Start

Быстрый тест kubectl proxy функциональности с minikube.

## 1. Подготовка minikube

```bash
# Запуск minikube
minikube start

# Получение IP API сервера
minikube ip
# Обновите api_server_url в configs/minikube-example.json

# Создание тестового пода
kubectl create deployment nginx --image=nginx
kubectl expose deployment nginx --port=80 --target-port=80
```

## 2. Запуск CloudBridge Relay

```bash
# Сборка
go build -o cloudbridge-relay ./cmd/relay

# Запуск с minikube конфигом
./cloudbridge-relay --config=configs/minikube-example.json
```

## 3. Настройка kubectl

Создайте kubeconfig для доступа через relay:

```yaml
# ~/.kube/config-relay
apiVersion: v1
kind: Config
clusters:
- name: minikube-via-relay
  cluster:
    server: http://localhost:8444/kube/minikube
    insecure-skip-tls-verify: true
contexts:
- name: minikube-via-relay
  context:
    cluster: minikube-via-relay
    user: test-user
current-context: minikube-via-relay
users:
- name: test-user
  user: {}
```

## 4. Тестирование

```bash
# Экспорт переменных для WebSocket
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1

# Базовые команды
kubectl --kubeconfig ~/.kube/config-relay get pods
kubectl --kubeconfig ~/.kube/config-relay get services

# WebSocket streams
kubectl --kubeconfig ~/.kube/config-relay logs deployment/nginx
kubectl --kubeconfig ~/.kube/config-relay exec -it deployment/nginx -- /bin/bash

# Port forwarding
kubectl --kubeconfig ~/.kube/config-relay port-forward svc/nginx 8080:80
# В другом терминале: curl http://localhost:8080
```

## 5. Проверка метрик

```bash
# Prometheus метрики
curl http://localhost:5551/metrics | grep kube_proxy

# Примеры:
# kube_proxy_requests_total{cluster_id="minikube",verb="get",resource="pods",code="200"} 5
# kube_ws_handshakes_total{cluster_id="minikube",code="101",subprotocol="v4.channel.k8s.io"} 2
# kube_ws_active_streams 0
```

## 6. Тест авторизации

Измените конфиг для добавления deny-правила:

```json
{
  "name": "deny-kube-system", 
  "verbs": ["*"],
  "api_groups": ["*"],
  "resources": ["*"],
  "namespaces": ["kube-system"],
  "deny": true
}
```

Перезапустите relay и проверьте:

```bash
# Должно работать
kubectl --kubeconfig ~/.kube/config-relay get pods -n default

# Должно быть запрещено (403 Forbidden)  
kubectl --kubeconfig ~/.kube/config-relay get pods -n kube-system
```

## 7. Аудит логи

В логах relay'я должны появиться записи:

```json
{
  "level": "info",
  "msg": "kube-proxy-audit", 
  "tenant_id": "default",
  "user_id": "anonymous",
  "verb": "get",
  "resource": "pods",
  "namespace": "default",
  "allowed": true,
  "matched_rule": "minikube-full-access"
}
```

---

**Готово!** Kubectl работает через CloudBridge Relay с полной поддержкой WebSocket streams, метриками и аудитом.
