# 📋 Пошаговая установка kube-prometheus-stack в VK Cloud

## 🎯 Цель
Установить kube-prometheus-stack через VK Cloud Console с внутренним IP адресом для мониторинга CloudBridge Relay.

## 📋 Подготовка

### 1. Проверка подключения к кластеру
```bash
# Установить переменную окружения
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Проверить подключение
kubectl --insecure-skip-tls-verify cluster-info

# Проверить доступные ресурсы
kubectl --insecure-skip-tls-verify describe nodes kubernetes-cluster-8926-default-group-0 | grep -A 10 "Allocated resources"
```

### 2. Очистка предыдущей установки (если есть)
```bash
# Удалить namespace prometheus-monitoring
kubectl --insecure-skip-tls-verify delete namespace prometheus-monitoring

# Подождать полного удаления
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus
```

## 🛠️ Установка через VK Cloud Console

### Шаг 1: Переход к кластеру
1. Открыть VK Cloud Console
2. Перейти к разделу **"Kubernetes"**
3. Выбрать кластер **`kubernetes-cluster-8926`**
4. Перейти на вкладку **"Аддоны"**

### Шаг 2: Поиск аддона
1. Найти аддон **"kube-prometheus-stack"**
2. Нажать кнопку **"Установить"**

### Шаг 3: Настройка параметров

#### 3.1 Основные параметры
- **Название приложения**: `cloudbridge-prometheus`
- **Пространство имен**: `prometheus-monitoring` (будет создано автоматически)

#### 3.2 Конфигурация Values
В поле **"Values"** вставить содержимое файла `configs/kube-prometheus-stack-vkcloud.yaml`:

```yaml
# Конфигурация kube-prometheus-stack для VK Cloud Console
# Установка с внутренним IP адресом

# Название приложения
nameOverride: cloudbridge-prometheus
fullnameOverride: cloudbridge-prometheus

# Пространство имен
# Будет создано автоматически: prometheus-monitoring

# Глобальные настройки
global:
  imageRegistry: "registry.infra.mail.ru:5010"
  imagePullSecrets: []

# Grafana - основная конфигурация
grafana:
  enabled: true
  
  # Пароль к порталу Grafana (ОБЯЗАТЕЛЬНОЕ ПОЛЕ)
  adminPassword: "CloudBridge2025!"
  
  # Ресурсы
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
  
  # Отключить persistent storage для экономии ресурсов
  persistence:
    enabled: false
  
  # Отключить Ingress (будет настроен позже)
  ingress:
    enabled: false
  
  # Конфигурация Grafana
  grafana.ini:
    server:
      root_url: http://localhost:8001/
    security:
      admin_user: admin
      admin_password: CloudBridge2025!
    users:
      allow_sign_up: false
    auth.anonymous:
      enabled: false
  
  # Отключить дополнительные компоненты
  sidecar:
    enabled: false
  initChownData:
    enabled: false
  
  # Использовать простой образ
  image:
    repository: "grafana/grafana"
    tag: "10.1.5"
    pullPolicy: IfNotPresent

# Prometheus - основная конфигурация
prometheus:
  enabled: true
  prometheusSpec:
    # Ресурсы
    resources:
      limits:
        cpu: 300m
        memory: 1Gi
      requests:
        cpu: 150m
        memory: 512Mi
    
    # Хранилище
    storageSpec:
      volumeClaimTemplate:
        spec:
          storageClassName: "csi-ceph-hdd-me1"
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 5Gi
    
    # Отключить Ingress
    ingress:
      enabled: false
    
    # Конфигурация
    scrapeInterval: "30s"
    retention: "7d"
    
    # Использовать простой образ
    image:
      repository: "prometheus/prometheus"
      tag: "v3.1.0"
      pullPolicy: IfNotPresent
    
    # Отключить дополнительные компоненты
    additionalScrapeConfigs: []
    additionalAlertManagerConfigs: []

# Alertmanager - отключить для экономии ресурсов
alertmanager:
  enabled: false

# Отключить все дополнительные компоненты для экономии ресурсов
kube-state-metrics:
  enabled: false

node-exporter:
  enabled: false

pushgateway:
  enabled: false

kubelet:
  enabled: false

kube-proxy:
  enabled: false

coreDns:
  enabled: false

kubeControllerManager:
  enabled: false

kubeScheduler:
  enabled: false

kubeEtcd:
  enabled: false

kubeApiserver:
  enabled: false

# Конфигурация для CloudBridge микросервисов
additionalServiceMonitors:
- name: cloudbridge-auth-service
  selector:
    matchLabels:
      app: auth-service
  namespaceSelector:
    matchNames:
    - cloudbridge
  endpoints:
  - port: metrics
    interval: 15s

- name: cloudbridge-metrics-service
  selector:
    matchLabels:
      app: metrics-service
  namespaceSelector:
    matchNames:
    - cloudbridge
  endpoints:
  - port: metrics
    interval: 15s

# Простые правила алертов
additionalPrometheusRules:
- name: cloudbridge-alerts
  groups:
  - name: cloudbridge
    rules:
    - alert: CloudBridgeServiceDown
      expr: up{job=~"cloudbridge.*"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "CloudBridge service is down"
        description: "Service {{ $labels.job }} has been down for more than 1 minute"
    
    - alert: CloudBridgeHighErrorRate
      expr: rate(http_requests_total{job=~"cloudbridge.*",status=~"5.."}[5m]) > 0.1
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate in CloudBridge service"
        description: "Error rate is {{ $value }} errors per second"

# Отключить дополнительные компоненты
cleanPrometheusOperatorObjectNames: true
kubernetesServiceMonitors:
  enabled: false
```

#### 3.3 Обязательные поля
- ✅ **Название приложения**: `cloudbridge-prometheus`
- ✅ **Пространство имен**: `prometheus-monitoring`
- ✅ **Пароль Grafana**: `CloudBridge2025!`

### Шаг 4: Установка
1. Нажать кнопку **"Установить аддон"**
2. Дождаться завершения установки (5-10 минут)

## 📊 Проверка установки

### Шаг 1: Проверка статуса
```bash
# Проверить namespace
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus

# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring

# Проверить сервисы
kubectl --insecure-skip-tls-verify get svc -n prometheus-monitoring
```

### Шаг 2: Проверка событий
```bash
# События в namespace
kubectl --insecure-skip-tls-verify get events -n prometheus-monitoring --sort-by='.lastTimestamp'

# Общие события кластера
kubectl --insecure-skip-tls-verify get events --all-namespaces | grep -i "prometheus\|monitoring"
```

### Шаг 3: Проверка ресурсов
```bash
# Использование ресурсов
kubectl --insecure-skip-tls-verify top pods -n prometheus-monitoring

# Проверить, что поды запустились
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring -o wide
```

## 🌐 Доступ к Grafana

### Шаг 1: Port-forward
```bash
# Запустить port-forward для Grafana
kubectl --insecure-skip-tls-verify -n prometheus-monitoring port-forward service/kube-prometheus-stack-grafana 8001:80
```

### Шаг 2: Открыть в браузере
1. Открыть браузер
2. Перейти по адресу: **http://127.0.0.1:8001**
3. Войти с учетными данными:
   - **Логин**: `admin`
   - **Пароль**: `CloudBridge2025!`

## 📈 Проверка Prometheus

### Шаг 1: Port-forward для Prometheus
```bash
# В другом терминале запустить port-forward для Prometheus
kubectl --insecure-skip-tls-verify -n prometheus-monitoring port-forward service/kube-prometheus-stack-prometheus 9090:9090
```

### Шаг 2: Проверка метрик
```bash
# Проверить доступность Prometheus
curl http://localhost:9090/api/v1/status/config

# Проверить метрики CloudBridge
curl http://localhost:9090/api/v1/query?query=up{job=~"cloudbridge.*"}
```

## 🔧 Настройка дашбордов

### Шаг 1: Импорт дашбордов в Grafana
1. Войти в Grafana (http://127.0.0.1:8001)
2. Перейти в **"Dashboards"** → **"Import"**
3. Импортировать дашборды:
   - **Node Exporter**: ID 1860
   - **Kubernetes Cluster**: ID 315
   - **CloudBridge Services**: создать вручную

### Шаг 2: Настройка источников данных
1. Перейти в **"Configuration"** → **"Data Sources"**
2. Добавить Prometheus:
   - **URL**: `http://kube-prometheus-stack-prometheus:9090`
   - **Access**: Server (default)

## 🚨 Устранение неполадок

### Проблема: Pods в состоянии Pending
```bash
# Проверить события
kubectl --insecure-skip-tls-verify describe pod <pod-name> -n prometheus-monitoring

# Возможные причины:
# - Insufficient memory/cpu
# - No persistent volumes available
# - Image pull errors
```

### Проблема: ImagePullBackOff
```bash
# Проверить логи
kubectl --insecure-skip-tls-verify logs <pod-name> -n prometheus-monitoring

# Решение: Использовать образы из registry.infra.mail.ru:5010
```

### Проблема: PersistentVolumeClaim в состоянии Pending
```bash
# Проверить StorageClass
kubectl --insecure-skip-tls-verify get storageclass

# Проверить PVC
kubectl --insecure-skip-tls-verify get pvc -n prometheus-monitoring
```

## 📊 Ожидаемый результат

### Успешная установка:
```bash
# Namespace
prometheus-monitoring    Active   2m

# Pods
NAME                                    READY   STATUS    RESTARTS   AGE
kube-prometheus-stack-grafana-xxx       1/1     Running   0          2m
kube-prometheus-stack-prometheus-xxx    1/1     Running   0          2m

# Services
NAME                                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)
kube-prometheus-stack-grafana           ClusterIP   10.254.xxx.xxx   <none>        80/TCP
kube-prometheus-stack-prometheus        ClusterIP   10.254.xxx.xxx   <none>        9090/TCP
```

### Доступ к сервисам:
- **Grafana**: http://127.0.0.1:8001 (через port-forward)
- **Prometheus**: http://localhost:9090 (через port-forward)
- **Логин Grafana**: admin / CloudBridge2025!

## 🎯 Следующие шаги

### 1. Настройка алертов
- Настройка уведомлений в Grafana
- Интеграция с Slack/Email
- Настройка правил алертов

### 2. Интеграция с CloudBridge микросервисами
- Проверка метрик auth-service
- Проверка метрик metrics-service
- Настройка ServiceMonitors

### 3. Настройка внешнего доступа
- Установка ingress-nginx
- Настройка SSL сертификатов
- Конфигурация доменов

---

**🎉 Готово!** kube-prometheus-stack установлен и настроен для CloudBridge Relay!

