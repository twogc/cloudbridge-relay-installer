# 🚀 CloudBridge Relay - Kubernetes Deployment

## Обзор

Этот раздел содержит Kubernetes манифесты для развертывания CloudBridge Relay Server с полным стеком сервисов, включая Cadence Workflow Engine, Cassandra, Prometheus и Grafana.

## 🏗️ Архитектура

### Компоненты системы:

- **CloudBridge Relay** - основной сервис (порт 9090)
- **Cassandra** - база данных (порт 9042)
- **Cadence** - workflow engine (порт 7933)
- **Prometheus** - метрики (порт 9090)
- **Grafana** - дашборды (порт 3000)

### Namespaces:

- `cloudbridge` - основные сервисы
- `cloudbridge-monitoring` - мониторинг

## 📁 Структура файлов

```
k8s/
├── namespaces/
│   └── cloudbridge-namespace.yaml
├── configmaps/
│   └── cloudbridge-config.yaml
├── secrets/
│   └── cloudbridge-secrets.yaml
├── persistent-volumes/
│   └── cloudbridge-pv.yaml
├── deployments/
│   ├── cloudbridge-relay-deployment.yaml
│   ├── cassandra-deployment.yaml
│   └── cadence-deployment.yaml
├── services/
│   └── cloudbridge-services.yaml
├── ingress/
│   └── cloudbridge-ingress.yaml
└── monitoring/
    ├── prometheus-deployment.yaml
    └── grafana-deployment.yaml
```

## 🚀 Быстрый старт

### 1. **Предварительные требования**

```bash
# Установка kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Проверка подключения к кластеру
kubectl cluster-info
```

### 2. **Развертывание**

```bash
# Автоматическое развертывание
./deploy-k8s.sh

# Или пошаговое развертывание
kubectl apply -f k8s/namespaces/
kubectl apply -f k8s/configmaps/
kubectl apply -f k8s/secrets/
kubectl apply -f k8s/persistent-volumes/
kubectl apply -f k8s/deployments/
kubectl apply -f k8s/services/
kubectl apply -f k8s/monitoring/
kubectl apply -f k8s/ingress/
```

### 3. **Проверка статуса**

```bash
# Статус подов
kubectl get pods -n cloudbridge
kubectl get pods -n cloudbridge-monitoring

# Статус сервисов
kubectl get services -n cloudbridge
kubectl get services -n cloudbridge-monitoring

# Логи
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

## 🔧 Конфигурация

### **ConfigMaps**

- `cloudbridge-config` - основная конфигурация
- `cassandra-config` - конфигурация Cassandra

### **Secrets**

- `cloudbridge-secrets` - пароли и ключи
- `wireguard-keys` - WireGuard ключи
- `prometheus-secrets` - конфигурация Prometheus
- `grafana-secrets` - конфигурация Grafana

### **Persistent Volumes**

- `cassandra-pv` - данные Cassandra (20Gi)
- `prometheus-pv` - метрики Prometheus (10Gi)
- `grafana-pv` - дашборды Grafana (5Gi)
- `cloudbridge-logs-pv` - логи (5Gi)

## 🌐 Доступ к сервисам

### **Внутренний доступ**

```bash
# CloudBridge Relay
http://cloudbridge-relay.cloudbridge.svc.cluster.local:9090

# Grafana
http://grafana.cloudbridge-monitoring.svc.cluster.local:3000

# Prometheus
http://prometheus.cloudbridge-monitoring.svc.cluster.local:9090

# Cassandra
cassandra-service.cloudbridge.svc.cluster.local:9042

# Cadence
cadence-service.cloudbridge.svc.cluster.local:7933
```

### **Внешний доступ (через Ingress)**

```bash
# CloudBridge Relay
https://cloudbridge.example.com

# Grafana
https://grafana.cloudbridge.example.com

# Prometheus
https://prometheus.cloudbridge.example.com
```

## 📊 Мониторинг

### **Health Checks**

Все сервисы имеют настроенные health checks:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9090
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 9090
  initialDelaySeconds: 5
  periodSeconds: 5
```

### **Метрики**

- **CloudBridge Relay**: `:8084/metrics`
- **Cassandra**: `:9500/metrics`
- **Prometheus**: `:9090/metrics`
- **Grafana**: `:3000/metrics`

## 🔒 Безопасность

### **RBAC**

Все сервисы имеют настроенные ServiceAccounts и RBAC:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
```

### **Security Context**

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
```

### **Network Policies**

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: cloudbridge-network-policy
  namespace: cloudbridge
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: cloudbridge-relay
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: cloudbridge
    ports:
    - protocol: TCP
      port: 9090
```

## 🚀 Масштабирование

### **Horizontal Pod Autoscaler**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: cloudbridge-relay-hpa
  namespace: cloudbridge
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cloudbridge-relay
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### **Vertical Pod Autoscaler**

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: cloudbridge-relay-vpa
  namespace: cloudbridge
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: cloudbridge-relay
  updatePolicy:
    updateMode: "Auto"
```

## 🔧 Troubleshooting

### **Общие проблемы**

#### Проблема: Pod не запускается

```bash
# Проверка событий
kubectl describe pod <pod-name> -n cloudbridge

# Проверка логов
kubectl logs <pod-name> -n cloudbridge

# Проверка конфигурации
kubectl get configmap cloudbridge-config -n cloudbridge -o yaml
```

#### Проблема: Сервис недоступен

```bash
# Проверка endpoints
kubectl get endpoints -n cloudbridge

# Проверка сервиса
kubectl describe service cloudbridge-relay -n cloudbridge

# Тест подключения
kubectl run test-pod --image=busybox -it --rm -- nslookup cloudbridge-relay.cloudbridge.svc.cluster.local
```

#### Проблема: Persistent Volume не монтируется

```bash
# Проверка PV
kubectl get pv

# Проверка PVC
kubectl get pvc -n cloudbridge

# Проверка событий
kubectl get events -n cloudbridge
```

### **Полезные команды**

```bash
# Просмотр всех ресурсов
kubectl get all -n cloudbridge
kubectl get all -n cloudbridge-monitoring

# Просмотр логов всех подов
kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge

# Вход в под
kubectl exec -it <pod-name> -n cloudbridge -- /bin/bash

# Порт-форвардинг для тестирования
kubectl port-forward service/cloudbridge-relay 9090:9090 -n cloudbridge
kubectl port-forward service/grafana 3000:3000 -n cloudbridge-monitoring
```

## 📚 Дополнительные ресурсы

### **Полезные ссылки**

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [Grafana Operator](https://github.com/grafana-operator/grafana-operator)
- [Cassandra Operator](https://github.com/datastax/cass-operator)

### **Best Practices**

1. **Используйте StatefulSet для Cassandra** - для персистентных данных
2. **Настройте Resource Limits** - для стабильности кластера
3. **Используйте Health Checks** - для мониторинга
4. **Настройте RBAC** - для безопасности
5. **Используйте Network Policies** - для изоляции сети
6. **Настройте Monitoring** - для observability

## 🎯 Заключение

Kubernetes deployment обеспечивает:

- **Высокую доступность** - реплики и health checks
- **Масштабируемость** - HPA и VPA
- **Безопасность** - RBAC и Network Policies
- **Мониторинг** - Prometheus и Grafana
- **Персистентность** - Persistent Volumes
- **Управляемость** - автоматическое развертывание

---

*Документ создан: 2025-09-06*  
*Статус: Готов к использованию* 🚀

