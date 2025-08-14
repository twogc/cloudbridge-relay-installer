# ⚡ Быстрая установка Prometheus Stack в VK Cloud

## 🚨 Если возникла ошибка "Не удалось установить аддон"

### 🔧 Быстрое решение:

1. **Используйте минимальную конфигурацию** из файла `configs/kube-prometheus-stack-basic.yaml`
2. **Проверьте ресурсы кластера** - должно быть минимум 2 CPU и 4GB RAM свободно
3. **Проверьте storage классы** - убедитесь что `csi-ceph-hdd-me1` доступен

### 📋 Пошаговое исправление:

```bash
# 1. Очистите предыдущую установку
kubectl delete namespace cloudbridge-monitoring --ignore-not-found=true

# 2. Создайте namespace заново
kubectl create namespace cloudbridge-monitoring

# 3. Проверьте ресурсы
kubectl describe nodes | grep -A 5 "Allocated resources"

# 4. Проверьте storage классы
kubectl get storageclass
```

## 🚀 Быстрый старт (5 минут)

### 1. Подготовка
```bash
# Создайте namespace
kubectl create namespace cloudbridge-monitoring
```

### 2. Установка в консоли VK Cloud

1. **Откройте консоль VK Cloud** → Kubernetes → Ваш кластер → Аддоны
2. **Добавить аддон** → kube-prometheus-stack
3. **Заполните поля:**
   - Название: `cloudbridge-prometheus`
   - Namespace: `cloudbridge-monitoring`
   - Пароль: `CloudBridge2025!`
4. **Вставьте код** из файла `configs/kube-prometheus-stack-basic.yaml` ⚠️ **ИСПОЛЬЗУЙТЕ ЭТОТ ФАЙЛ**
5. **Установите аддон**

### 3. Доступ к Grafana
```bash
# Способ 1: Port-forward
kubectl -n cloudbridge-monitoring port-forward service/prometheus-grafana 8001:80

# Способ 2: kauthproxy
kauthproxy -n cloudbridge-monitoring https://prometheus-grafana.svc
```

**URL**: http://127.0.0.1:8001  
**Логин**: `admin`  
**Пароль**: `CloudBridge2025!`

### 4. Проверка установки
```bash
# Проверка подов
kubectl get pods -n cloudbridge-monitoring

# Проверка сервисов
kubectl get services -n cloudbridge-monitoring

# Проверка PVC
kubectl get pvc -n cloudbridge-monitoring
```

## 📊 Что включено (минимальная версия)

✅ **Grafana** - визуализация метрик  
✅ **Prometheus** - сбор метрик  
✅ **Alertmanager** - алерты  
⚠️ **ServiceMonitors** - добавятся после установки  
⚠️ **PrometheusRules** - добавятся после установки  

## 🔧 Основные команды

```bash
# Доступ к Prometheus
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090

# Доступ к Alertmanager
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-alertmanager 9093:9093

# Логи Grafana
kubectl logs deployment/prometheus-grafana -n cloudbridge-monitoring

# Логи Prometheus
kubectl logs prometheus/prometheus -n cloudbridge-monitoring

# Проверка статуса подов
kubectl describe pods -n cloudbridge-monitoring
```

## 🚨 Устранение проблем

### Grafana не запускается
```bash
kubectl describe pod -l app.kubernetes.io/name=grafana -n cloudbridge-monitoring
kubectl logs deployment/prometheus-grafana -n cloudbridge-monitoring
```

### PVC остаются в Pending
```bash
kubectl get pvc -n cloudbridge-monitoring
kubectl describe pvc -n cloudbridge-monitoring
# Проверьте доступные storage классы
kubectl get storageclass
```

### Недостаточно ресурсов
```bash
kubectl top pods -n cloudbridge-monitoring
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### Pods остаются в Pending
```bash
kubectl describe pods -n cloudbridge-monitoring
# Ищите ошибки "Insufficient cpu" или "Insufficient memory"
```

## 📈 После успешной установки

### Добавление ServiceMonitors для CloudBridge
```bash
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloudbridge-auth-service
  namespace: cloudbridge-monitoring
spec:
  selector:
    matchLabels:
      app: auth-service
  namespaceSelector:
    matchNames:
    - cloudbridge
  endpoints:
  - port: metrics
    interval: 15s
    path: /metrics
EOF
```

### Добавление правил алертов
```bash
kubectl apply -f - <<EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cloudbridge-alerts
  namespace: cloudbridge-monitoring
spec:
  groups:
  - name: cloudbridge.rules
    rules:
    - alert: CloudBridgeServiceDown
      expr: up{app=~"auth-service|metrics-service"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "CloudBridge service is down"
EOF
```

## 📚 Подробная документация

- Полная инструкция: `docs/VK_CLOUD_PROMETHEUS_INSTALL_GUIDE.md`
- Устранение ошибок: `docs/TROUBLESHOOTING_PROMETHEUS_INSTALL.md`

## 🎯 Альтернативные решения

Если установка через аддон не работает:

### 1. Helm установка
```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace cloudbridge-monitoring \
  --create-namespace \
  --set grafana.adminPassword=CloudBridge2025!
```

### 2. Ручная установка Grafana
```bash
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: grafana
  namespace: cloudbridge-monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: grafana
  template:
    metadata:
      labels:
        app: grafana
    spec:
      containers:
      - name: grafana
        image: grafana/grafana:10.1.5
        ports:
        - containerPort: 3000
        env:
        - name: GF_SECURITY_ADMIN_PASSWORD
          value: "CloudBridge2025!"
EOF
```

---

**Время установки**: 5 минут  
**Статус**: ✅ Готов к использованию  
**При проблемах**: Используйте `configs/kube-prometheus-stack-basic.yaml`
