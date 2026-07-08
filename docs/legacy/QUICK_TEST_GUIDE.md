# Быстрое тестирование CloudBridge Relay kubectl proxy

Пошаговое руководство по тестированию и обновлению системы.

## 1. Предварительная проверка

```bash
# Проверка версий
kubectl version --client    # Должен быть >= 1.29 для WebSocket
go version                  # Должен быть >= 1.21
docker version              # Для сборки образов

# Проверка minikube
minikube status
# Если не запущен:
# minikube start --cpus=4 --memory=8192mb
```

## 2. Быстрый тест (5 минут)

### Шаг 1: Подготовка конфигурации

```bash
# Получение IP minikube
MINIKUBE_IP=$(minikube ip)
MINIKUBE_PORT=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | sed 's/.*://')

echo "Minikube API Server: https://$MINIKUBE_IP:$MINIKUBE_PORT"

# Обновление конфига (автоматически)
sed -i "s/192.168.49.2:8443/$MINIKUBE_IP:$MINIKUBE_PORT/g" configs/minikube-example.json
```

### Шаг 2: Сборка и запуск Relay

```bash
# Настройка Docker для minikube
eval $(minikube docker-env)

# Сборка образа
docker build -t cloudbridge-relay:test .

# Создание простого Deployment
cat > /tmp/quick-test-deploy.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: relay-config
data:
  config.json: |
$(cat configs/minikube-example.json | sed 's/^/    /')
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudbridge-relay
  template:
    metadata:
      labels:
        app: cloudbridge-relay
    spec:
      containers:
      - name: relay
        image: cloudbridge-relay:test
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        - containerPort: 8444
        - containerPort: 5551
        args: ["--config=/etc/config/config.json"]
        volumeMounts:
        - name: config
          mountPath: /etc/config
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 100m
            memory: 128Mi
      volumes:
      - name: config
        configMap:
          name: relay-config
---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay
spec:
  type: NodePort
  selector:
    app: cloudbridge-relay
  ports:
  - name: api
    port: 8080
    targetPort: 8080
    nodePort: 30080
  - name: grpc
    port: 8444
    targetPort: 8444
    nodePort: 30444  
  - name: metrics
    port: 5551
    targetPort: 5551
    nodePort: 30551
EOF

# Развертывание
kubectl apply -f /tmp/quick-test-deploy.yaml

# Ожидание готовности
kubectl wait --for=condition=ready pod -l app=cloudbridge-relay --timeout=120s
```

### Шаг 3: Создание kubeconfig для тестов

```bash
# Создание kubeconfig для relay
RELAY_URL="http://$(minikube ip):30444"

cat > ~/.kube/config-relay-test << EOF
apiVersion: v1
kind: Config
clusters:
- name: minikube-via-relay
  cluster:
    server: $RELAY_URL/kube/minikube
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
EOF

echo " Kubeconfig создан: ~/.kube/config-relay-test"
echo " Relay URL: $RELAY_URL"
```

### Шаг 4: Базовое тестирование

```bash
# Экспорт переменных для WebSocket
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1

echo "🧪 Тестирование базовых команд..."

# Тест 1: Получение pods
kubectl --kubeconfig ~/.kube/config-relay-test get pods --all-namespaces
echo " Тест 1 пройден: получение pods"

# Тест 2: Получение services  
kubectl --kubeconfig ~/.kube/config-relay-test get services --all-namespaces
echo " Тест 2 пройден: получение services"

# Тест 3: Создание тестового pod
kubectl --kubeconfig ~/.kube/config-relay-test run test-pod --image=busybox:1.35 --restart=Never -- sleep 3600
sleep 5

# Тест 4: WebSocket exec
kubectl --kubeconfig ~/.kube/config-relay-test exec test-pod -- echo "Hello from relay!"
echo " Тест 4 пройден: kubectl exec через WebSocket"

# Тест 5: WebSocket logs
kubectl --kubeconfig ~/.kube/config-relay-test logs test-pod
echo " Тест 5 пройден: kubectl logs через WebSocket"

# Cleanup
kubectl --kubeconfig ~/.kube/config-relay-test delete pod test-pod

echo "Базовые тесты пройдены успешно!"
```

### Шаг 5: Проверка метрик

```bash
echo "Проверка метрик..."

METRICS_URL="http://$(minikube ip):30551"
curl -s $METRICS_URL/metrics | grep -E "(kube_proxy_requests_total|kube_ws_)" | head -10

echo " Метрики доступны по адресу: $METRICS_URL/metrics"
```

## 3. Полные E2E тесты (10-15 минут)

```bash
# Настройка переменных для тестов
export RELAY_URL="http://$(minikube ip):30444"
export METRICS_URL="http://$(minikube ip):30551"
export CLUSTER_ID="minikube"

# Запуск всех E2E тестов
./test/e2e/run_tests.sh

# Или отдельные тест-суиты:
cd test/e2e

# Базовые команды + WebSocket streams
go test -v -timeout 300s -run TestKubeProxyE2E

# Авторизация
go test -v -timeout 300s -run TestAuthorizationE2E  

# Метрики
go test -v -timeout 300s -run TestMetricsE2E

# Полная интеграция
go test -v -timeout 600s -run TestIntegration
```

## 4. Обновление системы

### Локальное обновление (development)

```bash
# 1. Пересборка образа с новыми изменениями
eval $(minikube docker-env)
docker build -t cloudbridge-relay:v2 .

# 2. Обновление Deployment
kubectl patch deployment cloudbridge-relay -p '{"spec":{"template":{"spec":{"containers":[{"name":"relay","image":"cloudbridge-relay:v2"}]}}}}'

# 3. Проверка обновления
kubectl rollout status deployment/cloudbridge-relay
kubectl get pods -l app=cloudbridge-relay

# 4. Тестирование после обновления
kubectl --kubeconfig ~/.kube/config-relay-test get pods
```

### Обновление конфигурации

```bash
# Обновление ConfigMap
kubectl create configmap relay-config \
  --from-file=config.json=configs/minikube-example.json \
  --dry-run=client -o yaml | kubectl replace -f -

# Перезапуск для применения новой конфигурации
kubectl rollout restart deployment/cloudbridge-relay
kubectl rollout status deployment/cloudbridge-relay

# Проверка новой конфигурации
kubectl logs deployment/cloudbridge-relay | tail -20
```

## 5. Обновление Production системы

### Подготовка к обновлению

```bash
# 1. Резервное копирование текущей конфигурации
kubectl get configmap cloudbridge-relay-config -o yaml > backup-config-$(date +%Y%m%d-%H%M).yaml
kubectl get deployment cloudbridge-relay -o yaml > backup-deployment-$(date +%Y%m%d-%H%M).yaml

# 2. Проверка готовности кластера
kubectl get nodes
kubectl get pods -n cloudbridge --no-headers | wc -l

echo " Backup создан"
```

### Blue-Green развертывание

```bash
# 1. Создание новой версии с другим именем
sed 's/name: cloudbridge-relay/name: cloudbridge-relay-v2/' k8s/deployment-kube-proxy.yaml > /tmp/deployment-v2.yaml

# 2. Развертывание новой версии
kubectl apply -f /tmp/deployment-v2.yaml

# 3. Проверка готовности новой версии
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cloudbridge-relay-v2 --timeout=300s

# 4. Тестирование новой версии
RELAY_V2_URL=$(kubectl get service cloudbridge-relay-v2 -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl -s http://$RELAY_V2_URL/health

# 5. Переключение трафика (обновление LoadBalancer)
kubectl patch service cloudbridge-relay -p '{"spec":{"selector":{"app.kubernetes.io/instance":"cloudbridge-relay-v2"}}}'

# 6. Проверка работоспособности
kubectl --kubeconfig ~/.kube/config-relay-production get pods

# 7. Удаление старой версии (через час после проверки)
# kubectl delete deployment cloudbridge-relay-v1
```

### Rolling Update (рекомендуется)

```bash
# 1. Обновление образа с новым тегом
kubectl set image deployment/cloudbridge-relay \
  cloudbridge-relay=ghcr.io/twogc/cloudbridge-relay:v1.2.0

# 2. Мониторинг процесса обновления
kubectl rollout status deployment/cloudbridge-relay --watch

# 3. Проверка здоровья всех подов
kubectl get pods -l app=cloudbridge-relay
kubectl logs deployment/cloudbridge-relay --tail=50

# 4. Тестирование функциональности
kubectl --kubeconfig ~/.kube/config-relay-production get namespaces
kubectl --kubeconfig ~/.kube/config-relay-production exec deployment/test-app -- echo "test"

# 5. Откат при необходимости
# kubectl rollout undo deployment/cloudbridge-relay
```

## 6. Мониторинг после обновления

### Проверка метрик

```bash
# Основные метрики работоспособности
curl -s http://$RELAY_URL/metrics | grep -E "(kube_proxy_requests_total|up|process_start_time)"

# WebSocket метрики
curl -s http://$RELAY_URL/metrics | grep -E "kube_ws_"

# Ошибки upstream
curl -s http://$RELAY_URL/metrics | grep kube_upstream_errors_total
```

### Проверка логов

```bash
# Последние логи
kubectl logs deployment/cloudbridge-relay --tail=100

# Логи аудита авторизации
kubectl logs deployment/cloudbridge-relay | grep "kube-proxy-audit"

# Ошибки
kubectl logs deployment/cloudbridge-relay | grep -i error
```

### Health Check

```bash
# Health endpoint
curl -s http://$RELAY_URL/health | jq '.'

# Ready endpoint  
curl -s http://$RELAY_URL/ready | jq '.'

# Состояние Kubernetes ресурсов
kubectl get deployment,service,configmap -l app=cloudbridge-relay
```

## 7.  Troubleshooting

### Общие проблемы

```bash
# Проблема: Pod не запускается
kubectl describe pod -l app=cloudbridge-relay
kubectl logs deployment/cloudbridge-relay

# Проблема: Service недоступен
kubectl describe service cloudbridge-relay
kubectl get endpoints cloudbridge-relay

# Проблема: Конфигурация не применилась
kubectl get configmap relay-config -o yaml
kubectl exec deployment/cloudbridge-relay -- cat /etc/config/config.json

# Проблема: kubectl команды не работают
kubectl --kubeconfig ~/.kube/config-relay-test get pods -v=8
```

### Восстановление после сбоя

```bash
# Восстановление из backup
kubectl apply -f backup-deployment-$(date +%Y%m%d)*.yaml
kubectl apply -f backup-config-$(date +%Y%m%d)*.yaml

# Принудительная перезагрузка
kubectl delete pods -l app=cloudbridge-relay
kubectl rollout status deployment/cloudbridge-relay

# Проверка восстановления
kubectl --kubeconfig ~/.kube/config-relay-test get pods
```

## 8. Checklist обновления

### Перед обновлением:
- [ ] Создан backup конфигурации и deployment
- [ ] Протестированы изменения в dev/staging  
- [ ] Проверена совместимость с kubectl >= 1.29
- [ ] Проверены метрики и дашборды
- [ ] Уведомлена команда об обновлении

### Во время обновления:
- [ ] Мониторинг здоровья подов
- [ ] Проверка метрик в реальном времени
- [ ] Тестирование kubectl команд
- [ ] Проверка WebSocket функциональности
- [ ] Мониторинг логов на ошибки

### После обновления:
- [ ] Полное тестирование E2E сценариев
- [ ] Проверка всех типов kubectl операций
- [ ] Мониторинг производительности 30 минут
- [ ] Проверка audit логов
- [ ] Документирование изменений
- [ ] Уведомление команды об успешном обновлении

---

## Готово!

Теперь у вас есть полный набор инструментов для:
- **Быстрого тестирования** (5 минут)  
- **Полного E2E тестирования** (15 минут)
- **Безопасного обновления** production системы
- **Мониторинга** здоровья системы

**Система готова к production использованию!**
