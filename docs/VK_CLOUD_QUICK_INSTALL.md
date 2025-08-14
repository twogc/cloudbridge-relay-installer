# 🚀 Быстрая установка Prometheus Stack в VK Cloud

## 📋 Информация о кластере
- **Кластер**: kubernetes-cluster-6379
- **API**: https://90.156.150.25:6443
- **Ресурсы**: 3 vCPU / 8 ГБ RAM
- **Доверенные Registry**: docker.io, registry.infra.mail.ru:5010

## 🎯 Конфигурация для копирования

Скопируйте **ВСЕ** содержимое ниже в поле "Код настройки аддона":

```yaml
# Конфигурация kube-prometheus-stack для VK Cloud
# Основана на официальной инструкции VK Cloud

nameOverride: prometheus
fullnameOverride: prometheus

crds:
  enabled: true

kubernetesServiceMonitors:
  enabled: true

global:
  imageRegistry: "docker.io"

cleanPrometheusOperatorObjectNames: false

extraManifests: []

# Grafana Dashboard
grafana:
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
  adminPassword: "CloudBridge2025!"  # Пароль к порталу Grafana
  ingress:
    enabled: false
  persistence:
    enabled: true
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    size: 1Gi
  sidecar:
    enabled: false
  initChownData:
    enabled: false
  image:
    repository: "grafana/grafana"
    tag: "10.1.5"
  deploymentStrategy:
    type: Recreate

# Alertmanager
alertmanager:
  enabled: true
  ingress:
    enabled: false
  alertmanagerSpec:
    image:
      repository: prometheus/alertmanager
      tag: v0.28.0
    resources:
      requests:
        cpu: 50m
        memory: 64Mi
      limits:
        cpu: 200m
        memory: 256Mi
    storage:
      volumeClaimTemplate:
        spec:
          storageClassName: "standard"
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi

# Prometheus Operator
prometheusOperator:
  repository: prometheus-operator/prometheus-operator
  tag: "v0.69.1"
  revisionHistoryLimit: 10
  prometheusConfigReloader:
    image:
      repository: prometheus-operator/prometheus-config-reloader
      tag: "v0.79.2"
    resources:
      requests:
        cpu: 25m
        memory: 25Mi
      limits:
        cpu: 50m
        memory: 50Mi
  resources:
    requests:
      cpu: 50m
      memory: 50Mi
    limits:
      cpu: 100m
      memory: 100Mi
  cleanupCustomResource: true
  admissionWebhooks:
    patch:
      image:
        repository: ingress-nginx/kube-webhook-certgen
        tag: v1.5.1
      resources:
        requests:
          cpu: 10m
          memory: 50Mi
        limits:
          cpu: 50m
          memory: 100Mi

# Prometheus
prometheus:
  ingress:
    enabled: false
  prometheusSpec:
    scrapeInterval: "30s"
    image:
      repository: prometheus/prometheus
      tag: v3.1.0
    ruleSelectorNilUsesHelmValues: false
    serviceMonitorSelectorNilUsesHelmValues: false
    retention: "7d"
    resources:
      requests:
        cpu: 200m
        memory: 256Mi
      limits:
        cpu: 400m
        memory: 512Mi
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: "standard"
          accessModes:
          - ReadWriteOnce
          resources:
            requests:
              storage: 5Gi

# Kube State Metrics
kube-state-metrics:
  image:
    repository: "kube-state-metrics/kube-state-metrics"
    tag: v2.10.1
  resources:
    limits:
      cpu: 50m
      memory: 32Mi
    requests:
      cpu: 10m
      memory: 16Mi

# Node Exporter
prometheus-node-exporter:
  image:
    repository: "prometheus/node-exporter"
    tag: v1.7.0
  namespaceOverride: "kube-system"
  resources:
    limits:
      cpu: 100m
      memory: 30Mi
    requests:
      cpu: 50m
      memory: 20Mi

# Kubernetes Components Monitoring
kubeApiServer:
  enabled: true
  tlsConfig:
    insecureSkipVerify: False

kubelet:
  enabled: true
  serviceMonitor:
    https: https

coreDns:
  enabled: true
  service:
    port: 9153
    targetPort: 9153
    selector:
      k8s-app: coredns

kubeEtcd:
  enabled: false

kubeScheduler:
  enabled: true
  endpoints: []
  service:
    selector:
      k8s-app: kube-scheduler

kubeControllerManager:
  enabled: true
  endpoints: []
  service:
    selector:
      k8s-app: kube-controller-manager
```

## 📝 Шаги установки

### 1. Откройте VK Cloud
- https://mcs.mail.ru/
- Войдите в аккаунт `info@2gc.ru`

### 2. Перейдите к кластеру
- **Kubernetes** → `kubernetes-cluster-6379`

### 3. Добавьте аддон
- **Аддоны** → **Добавить аддон** → **kube-prometheus-stack**

### 4. Заполните поля
- **Название**: `cloudbridge-prometheus`
- **Namespace**: `prometheus-monitoring`
- **Пароль**: `CloudBridge2025!`

### 5. Вставьте конфигурацию
Скопируйте конфигурацию выше в поле **Код настройки аддона**

### 6. Установите
- **Установить аддон**
- Ждите 10-15 минут

## 🔍 Проверка

### В консоли VK Cloud:
1. **Ресурсы кластера** → **Поды**
2. Namespace: `prometheus-monitoring`
3. Должны быть поды:
   - `prometheus-grafana-xxx`
   - `prometheus-kube-prometheus-prometheus-xxx`
   - `prometheus-kube-prometheus-alertmanager-xxx`

## 🌐 Доступ

### Через port-forward:
```bash
# Grafana
kubectl -n prometheus-monitoring port-forward service/prometheus-grafana 8001:80

# Prometheus
kubectl -n prometheus-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090
```

### Логин в Grafana:
- **URL**: http://127.0.0.1:8001
- **Логин**: `admin`
- **Пароль**: `CloudBridge2025!`

## ✅ Что будет установлено

- **Grafana** - визуализация метрик
- **Prometheus** - сбор метрик
- **Alertmanager** - алерты
- **Kube State Metrics** - метрики Kubernetes
- **Node Exporter** - метрики узлов
- **Мониторинг Kubernetes** - API Server, Scheduler, Controller Manager

---

**Статус**: ✅ Готов к установке  
**Время**: 10-15 минут  
**Ресурсы**: ~1.5 CPU, ~1GB RAM
