# 🚨 Устранение ошибок установки Prometheus Stack в VK Cloud

## ❌ Ошибка: "Не удалось установить аддон"

### 🔍 Диагностика проблемы

#### 1. Проверьте ресурсы кластера
```bash
# Проверьте доступные ресурсы
kubectl describe nodes
kubectl top nodes

# Проверьте существующие поды
kubectl get pods --all-namespaces
```

#### 2. Проверьте storage классы
```bash
# Проверьте доступные storage классы
kubectl get storageclass

# Убедитесь, что csi-ceph-hdd-me1 доступен
kubectl describe storageclass csi-ceph-hdd-me1
```

#### 3. Проверьте версию Kubernetes
```bash
kubectl version --short
```

## 🔧 Решения проблем

### Проблема 1: Недостаточно ресурсов

**Симптомы:**
- Pods остаются в состоянии Pending
- Ошибки "Insufficient cpu" или "Insufficient memory"

**Решение:**
Используйте минимальную конфигурацию из `configs/kube-prometheus-stack-basic.yaml`:

```yaml
# Минимальные ресурсы
grafana:
  resources:
    limits:
      cpu: 100m
      memory: 128Mi
    requests:
      cpu: 50m
      memory: 64Mi

prometheus:
  prometheusSpec:
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 200m
        memory: 512Mi
```

### Проблема 2: Неправильный storage класс

**Симптомы:**
- PVC остаются в состоянии Pending
- Ошибки "no persistent volumes available"

**Решение:**
Проверьте доступные storage классы и используйте правильный:

```bash
# Проверьте доступные storage классы
kubectl get storageclass

# Если csi-ceph-hdd-me1 недоступен, используйте другой
# Например: standard, default, или другой доступный
```

Измените конфигурацию:
```yaml
persistence:
  storageClassName: "standard"  # или другой доступный класс
```

### Проблема 3: Проблемы с CRD

**Симптомы:**
- Ошибки "custom resource definition not found"
- Prometheus Operator не может создать ресурсы

**Решение:**
Убедитесь, что CRD установлены:
```bash
# Проверьте CRD
kubectl get crd | grep prometheus

# Если CRD отсутствуют, установите их вручную
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagerconfigs.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_alertmanagers.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheuses.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml
kubectl apply -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml
```

### Проблема 4: Конфликт имен

**Симптомы:**
- Ошибки "already exists"
- Ресурсы не создаются

**Решение:**
Измените имена в конфигурации:
```yaml
nameOverride: cloudbridge-prometheus
fullnameOverride: cloudbridge-prometheus
```

### Проблема 5: Проблемы с образами

**Симптомы:**
- Pods остаются в состоянии ImagePullBackOff
- Ошибки "image not found"

**Решение:**
Используйте публичные образы или проверьте registry:
```yaml
global:
  imageRegistry: "docker.io"  # или другой доступный registry
```

## 🚀 Пошаговое решение

### Шаг 1: Очистка предыдущей установки
```bash
# Удалите предыдущие ресурсы
kubectl delete namespace cloudbridge-monitoring --ignore-not-found=true
kubectl delete crd alertmanagerconfigs.monitoring.coreos.com --ignore-not-found=true
kubectl delete crd alertmanagers.monitoring.coreos.com --ignore-not-found=true
kubectl delete crd prometheuses.monitoring.coreos.com --ignore-not-found=true
kubectl delete crd prometheusrules.monitoring.coreos.com --ignore-not-found=true
kubectl delete crd servicemonitors.monitoring.coreos.com --ignore-not-found=true
```

### Шаг 2: Создание namespace
```bash
kubectl create namespace cloudbridge-monitoring
```

### Шаг 3: Проверка ресурсов
```bash
# Проверьте доступные ресурсы
kubectl describe nodes | grep -A 5 "Allocated resources"

# Убедитесь, что есть свободные ресурсы
kubectl top nodes
```

### Шаг 4: Использование минимальной конфигурации
Используйте файл `configs/kube-prometheus-stack-basic.yaml` вместо полной конфигурации.

### Шаг 5: Постепенная установка
Если минимальная конфигурация работает, постепенно добавляйте компоненты:

1. **Сначала только Grafana:**
```yaml
grafana:
  enabled: true
  # минимальная конфигурация

prometheus:
  enabled: false

alertmanager:
  enabled: false
```

2. **Затем добавьте Prometheus:**
```yaml
prometheus:
  enabled: true
  # минимальная конфигурация
```

3. **Наконец добавьте Alertmanager:**
```yaml
alertmanager:
  enabled: true
  # минимальная конфигурация
```

## 🔍 Альтернативные решения

### Решение 1: Использование Helm напрямую
```bash
# Добавьте репозиторий
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Установите с минимальной конфигурацией
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace cloudbridge-monitoring \
  --create-namespace \
  --set grafana.adminPassword=CloudBridge2025! \
  --set prometheus.prometheusSpec.resources.requests.cpu=100m \
  --set prometheus.prometheusSpec.resources.requests.memory=256Mi
```

### Решение 2: Ручная установка компонентов
```bash
# Установите только Grafana
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
        resources:
          requests:
            cpu: 50m
            memory: 64Mi
          limits:
            cpu: 100m
            memory: 128Mi
---
apiVersion: v1
kind: Service
metadata:
  name: grafana
  namespace: cloudbridge-monitoring
spec:
  selector:
    app: grafana
  ports:
  - port: 80
    targetPort: 3000
EOF
```

## 📋 Чек-лист для успешной установки

- [ ] Кластер имеет достаточно ресурсов (минимум 2 CPU, 4GB RAM)
- [ ] Storage классы доступны
- [ ] Namespace создан
- [ ] Используется минимальная конфигурация
- [ ] Нет конфликтов имен
- [ ] Образы доступны
- [ ] CRD установлены

## 🆘 Если ничего не помогает

### Обратитесь в поддержку VK Cloud
1. Предоставьте логи ошибок
2. Укажите версию Kubernetes
3. Опишите шаги воспроизведения

### Используйте альтернативные решения
1. **Grafana Cloud** - облачное решение
2. **DataDog** - коммерческое решение
3. **New Relic** - APM решение

## 📞 Контакты для поддержки

- **VK Cloud Support**: https://mcs.mail.ru/help
- **Prometheus Community**: https://prometheus.io/community/
- **Grafana Community**: https://community.grafana.com/

---

**Статус**: 🔧 **Требует диагностики**  
**Сложность**: Средняя  
**Время решения**: 30-60 минут
