# 🔄 Альтернативные способы установки Prometheus Stack

## 🚨 Если аддон VK Cloud не работает

Если установка через аддон VK Cloud не удается, используйте эти альтернативные способы.

## 🚀 Способ 1: Простая установка Grafana (рекомендуется)

### Быстрая установка только Grafana:

```bash
# 1. Примените простой манифест
kubectl apply -f k8s/grafana-simple.yaml

# 2. Проверьте установку
kubectl get pods -n cloudbridge-monitoring

# 3. Получите доступ
kubectl -n cloudbridge-monitoring port-forward service/grafana 8001:80
```

**URL**: http://127.0.0.1:8001  
**Логин**: `admin`  
**Пароль**: `CloudBridge2025!`

### Что включено:
- ✅ Grafana с минимальными ресурсами (25m CPU / 32Mi Memory)
- ✅ SQLite база данных (не требует persistent storage)
- ✅ Готовая конфигурация
- ✅ Health checks

## 📦 Способ 2: Helm установка

### Установка через Helm:

```bash
# 1. Установите Helm (если не установлен)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# 2. Добавьте репозиторий
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# 3. Создайте namespace
kubectl create namespace cloudbridge-monitoring

# 4. Установите с минимальной конфигурацией
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace cloudbridge-monitoring \
  --set grafana.adminPassword=CloudBridge2025! \
  --set prometheus.prometheusSpec.resources.requests.cpu=50m \
  --set prometheus.prometheusSpec.resources.requests.memory=128Mi \
  --set grafana.resources.requests.cpu=25m \
  --set grafana.resources.requests.memory=32Mi \
  --set kube-state-metrics.enabled=false \
  --set prometheus-node-exporter.enabled=false
```

## 🔧 Способ 3: Интерактивный скрипт

### Использование готового скрипта:

```bash
# 1. Сделайте скрипт исполняемым
chmod +x scripts/install-prometheus-alternatives.sh

# 2. Запустите скрипт
./scripts/install-prometheus-alternatives.sh
```

Скрипт предоставляет интерактивное меню с 4 способами установки:
1. Helm установка
2. Только Grafana
3. kubectl apply
4. Готовый манифест

## 📄 Способ 4: Ручная установка компонентов

### Установка Prometheus Operator:

```bash
# 1. Создайте namespace
kubectl create namespace cloudbridge-monitoring

# 2. Установите CRD
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml

# 3. Установите Prometheus Operator
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml
```

## 🎯 Рекомендуемый порядок установки

### Для быстрого старта:
1. **Способ 1** - Простая установка Grafana (5 минут)
2. **Способ 2** - Helm установка (10 минут)
3. **Способ 3** - Интерактивный скрипт (выбор)
4. **Способ 4** - Ручная установка (15 минут)

## 🔍 Проверка установки

### Общие команды для проверки:

```bash
# Проверка namespace
kubectl get namespace cloudbridge-monitoring

# Проверка подов
kubectl get pods -n cloudbridge-monitoring

# Проверка сервисов
kubectl get services -n cloudbridge-monitoring

# Проверка логов Grafana
kubectl logs deployment/grafana -n cloudbridge-monitoring

# Проверка статуса подов
kubectl describe pods -n cloudbridge-monitoring
```

### Проверка доступа:

```bash
# Port-forward для Grafana
kubectl -n cloudbridge-monitoring port-forward service/grafana 8001:80

# Port-forward для Prometheus (если установлен)
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-prometheus 9090:9090

# Port-forward для Alertmanager (если установлен)
kubectl -n cloudbridge-monitoring port-forward service/prometheus-kube-prometheus-alertmanager 9093:9093
```

## 🚨 Устранение проблем

### Проблема: Pods остаются в Pending
```bash
# Проверьте ресурсы
kubectl describe nodes | grep -A 5 "Allocated resources"

# Проверьте события
kubectl get events -n cloudbridge-monitoring --sort-by='.lastTimestamp'
```

### Проблема: Не удается подключиться к кластеру
```bash
# Проверьте kubeconfig
kubectl config view

# Проверьте текущий контекст
kubectl config current-context

# Проверьте доступность кластера
kubectl cluster-info
```

### Проблема: Образы не загружаются
```bash
# Проверьте статус подов
kubectl describe pod -n cloudbridge-monitoring

# Проверьте сетевые политики
kubectl get networkpolicies -n cloudbridge-monitoring
```

## 📊 После успешной установки

### Добавление дашбордов:

1. **Войдите в Grafana**
2. **Импортируйте дашборды:**
   - Kubernetes Cluster Monitoring (ID: 315)
   - Kubernetes Pods (ID: 6417)
   - Kubernetes Services (ID: 6418)

### Добавление CloudBridge метрик:

```bash
# Создайте ServiceMonitor для CloudBridge сервисов
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

## 🎯 Преимущества каждого способа

### Способ 1 (Простая установка Grafana):
- ✅ Быстрая установка (5 минут)
- ✅ Минимальные ресурсы
- ✅ Не требует persistent storage
- ✅ Работает в любом кластере

### Способ 2 (Helm):
- ✅ Полный Prometheus Stack
- ✅ Автоматическая настройка
- ✅ Легкое обновление
- ✅ Гибкая конфигурация

### Способ 3 (Интерактивный скрипт):
- ✅ Выбор способа установки
- ✅ Автоматическая диагностика
- ✅ Проверка установки
- ✅ Удобный интерфейс

### Способ 4 (Ручная установка):
- ✅ Полный контроль
- ✅ Понимание компонентов
- ✅ Возможность кастомизации
- ✅ Обучение

## 📞 Поддержка

Если ни один способ не работает:

1. **Проверьте ресурсы кластера** - минимум 2 CPU, 4GB RAM
2. **Проверьте сетевые политики** - могут блокировать доступ
3. **Проверьте storage классы** - должны быть доступны
4. **Обратитесь в поддержку VK Cloud** - предоставьте логи ошибок

---

**Рекомендация**: Начните со **Способа 1** - простой установки Grafana. Это самый надежный способ для быстрого старта! 🚀
