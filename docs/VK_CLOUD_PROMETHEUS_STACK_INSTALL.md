# 🚀 Установка Prometheus Stack в VK Cloud - ПОЛНАЯ ВЕРСИЯ

## 📋 Информация о кластере

- **ID кластера**: cd494d7f-937c-4cf1-be99-9f2d75e43112
- **API сервер**: https://90.156.150.25:6443
- **Ресурсы**: 3 vCPU / 8 ГБ RAM
- **Версия Kubernetes**: 1.32.1
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

### 1. Откройте консоль VK Cloud
- https://mcs.mail.ru/
- Войдите в аккаунт `info@2gc.ru`
- Выберите проект с кластером `kubernetes-cluster-6379`

### 2. Перейдите к кластеру
- Меню слева → **Kubernetes**
- Найдите `kubernetes-cluster-6379`
- Нажмите на него

### 3. Добавьте аддон
- Вкладка **Аддоны**
- **Добавить аддон**
- Выберите **kube-prometheus-stack**

### 4. Заполните поля
- **Название приложения**: `cloudbridge-prometheus`
- **Пространство имен**: `prometheus-monitoring`
- **Пароль Grafana**: `CloudBridge2025!`

### 5. Вставьте конфигурацию
Вставьте конфигурацию выше в поле **Код настройки аддона**

### 6. Установите
- **Установить аддон**
- Ждите 10-15 минут

## 🔍 Проверка установки

### В консоли VK Cloud:
1. **Ресурсы кластера** → **Поды**
2. Найдите namespace `prometheus-monitoring`
3. Должны быть поды:
   - `prometheus-grafana-xxx`
   - `prometheus-kube-prometheus-prometheus-xxx`
   - `prometheus-kube-prometheus-alertmanager-xxx`
   - `prometheus-kube-state-metrics-xxx`

### Статус всех подов должен быть **Running**

## 🌐 Доступ к веб-интерфейсам

### Способ 1: Через port-forward (рекомендуется)
```bash
# Grafana
kubectl -n prometheus-monitoring port-forward service/prometheus-grafana 8001:80

# Prometheus
kubectl -n prometheus-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090

# Alertmanager
kubectl -n prometheus-monitoring port-forward service/prometheus-kube-prometheus-alertmanager 9093:9093
```

### Способ 2: Через консоль VK Cloud
1. **Ресурсы кластера** → **Сервисы**
2. Найдите нужный сервис
3. Используйте **Внешний IP** или **Внутренний IP**

## 🔑 Доступ к веб-интерфейсам

### Grafana
- **URL**: http://127.0.0.1:8001 (при port-forward)
- **Логин**: `admin`
- **Пароль**: `CloudBridge2025!`

### Prometheus
- **URL**: http://127.0.0.1:9090 (при port-forward)
- **Логин**: не требуется

### Alertmanager
- **URL**: http://127.0.0.1:9093 (при port-forward)
- **Логин**: не требуется

## ⚠️ Что будет установлено

✅ **Grafana** - веб-интерфейс для визуализации  
✅ **Prometheus** - сбор и хранение метрик  
✅ **Alertmanager** - система алертов  
✅ **Prometheus Operator** - управление ресурсами  
✅ **Kube State Metrics** - метрики Kubernetes  
✅ **Node Exporter** - метрики узлов  
✅ **Мониторинг Kubernetes** - API Server, Scheduler, Controller Manager  

## 🚨 Если установка не удается

### Проблема: "Не удалось установить аддон"
1. **Проверьте ресурсы** - должно быть минимум 2 CPU и 4GB RAM свободно
2. **Попробуйте еще раз** - иногда нужно несколько попыток
3. **Используйте локальную установку** как альтернативу

### Альтернатива - локальная установка:
```bash
# Остановите локальные контейнеры
docker stop cloudbridge-grafana cloudbridge-prometheus

# Запустите через docker-compose
docker-compose -f docker-compose.monitoring.yml up -d
```

## 📊 Следующие шаги

После успешной установки:
1. **Настройте дашборды** в Grafana
2. **Добавьте CloudBridge сервисы** в мониторинг
3. **Настройте алерты** для критических событий
4. **Настройте внешний доступ** через Ingress

---

**Статус**: ✅ Полная установка Prometheus Stack  
**Кластер**: kubernetes-cluster-6379  
**Время установки**: 10-15 минут  
**Сложность**: Средняя  
**Ресурсы**: ~1.5 CPU, ~1GB RAM
