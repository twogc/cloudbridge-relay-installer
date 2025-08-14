# 🔄 Переустановка kube-prometheus-stack в VK Cloud

## 🚨 Проблема
Аддон kube-prometheus-stack не установился из-за недостатка ресурсов в кластере.

## 📊 Анализ ресурсов кластера

### Текущие ресурсы:
- **Worker node**: 2 CPU, ~3.7GB RAM
- **Master node**: 4 CPU, ~7.8GB RAM (с taints, поды не размещаются)
- **Доступно на worker**: ~1.9 CPU, ~2.5GB RAM

### Требования kube-prometheus-stack:
- **Стандартная конфигурация**: 2+ CPU, 4+ GB RAM
- **Наша ультра-минимальная**: 0.3 CPU, 0.5GB RAM

## 🛠️ Пошаговое решение

### Шаг 1: Удаление неудачной установки

#### 1.1 Удалить namespace prometheus-monitoring
```bash
# Установить переменную окружения
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Удалить namespace
kubectl --insecure-skip-tls-verify delete namespace prometheus-monitoring

# Подождать полного удаления
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus
```

#### 1.2 Проверить, что все ресурсы удалены
```bash
# Проверить события удаления
kubectl --insecure-skip-tls-verify get events --all-namespaces | grep -i "prometheus\|monitoring"
```

### Шаг 2: Подготовка ультра-минимальной конфигурации

#### 2.1 Использовать файл `configs/kube-prometheus-stack-ultra-minimal.yaml`
Этот файл содержит:
- **Grafana**: 50m-100m CPU, 64Mi-128Mi RAM
- **Prometheus**: 100m-200m CPU, 256Mi-512Mi RAM
- **Alertmanager**: отключен
- **Дополнительные компоненты**: отключены
- **Хранилище**: 2Gi для Prometheus

#### 2.2 Проверить доступные ресурсы
```bash
# Проверить свободные ресурсы
kubectl --insecure-skip-tls-verify describe nodes kubernetes-cluster-8926-default-group-0 | grep -A 10 "Allocated resources"

# Убедиться, что есть достаточно места
# Нужно: ~300m CPU, ~600Mi RAM
```

### Шаг 3: Установка через VK Cloud Console

#### 3.1 Перейти к кластеру
1. Открыть VK Cloud Console
2. Перейти к разделу "Kubernetes"
3. Выбрать кластер `kubernetes-cluster-8926`
4. Перейти на вкладку "Аддоны"

#### 3.2 Удалить старый аддон (если есть)
1. Найти аддон "kube-prometheus-stack"
2. Нажать "Удалить" (если статус "Ошибка")
3. Подождать полного удаления

#### 3.3 Установить новый аддон
1. Нажать "Установить" для kube-prometheus-stack
2. В поле "Values" вставить содержимое файла `configs/kube-prometheus-stack-ultra-minimal.yaml`
3. Нажать "Установить"

### Шаг 4: Мониторинг установки

#### 4.1 Проверка статуса
```bash
# Проверить namespace
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus

# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring

# Проверить сервисы
kubectl --insecure-skip-tls-verify get svc -n prometheus-monitoring
```

#### 4.2 Проверка событий
```bash
# События в namespace
kubectl --insecure-skip-tls-verify get events -n prometheus-monitoring --sort-by='.lastTimestamp'

# Общие события кластера
kubectl --insecure-skip-tls-verify get events --all-namespaces | grep -i "prometheus\|monitoring"
```

#### 4.3 Проверка ресурсов
```bash
# Использование ресурсов
kubectl --insecure-skip-tls-verify top pods -n prometheus-monitoring

# Проверить, что поды запустились
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring -o wide
```

### Шаг 5: Тестирование работоспособности

#### 5.1 Проверка Prometheus
```bash
# Port-forward для Prometheus
kubectl --insecure-skip-tls-verify port-forward svc/prometheus-operated 9090:9090 -n prometheus-monitoring

# В другом терминале проверить доступность
curl http://localhost:9090/api/v1/status/config
```

#### 5.2 Проверка Grafana
```bash
# Port-forward для Grafana
kubectl --insecure-skip-tls-verify port-forward svc/grafana 3000:80 -n prometheus-monitoring

# В браузере открыть http://localhost:3000
# Логин: admin
# Пароль: CloudBridge2025!
```

#### 5.3 Проверка метрик CloudBridge
```bash
# Проверить ServiceMonitors
kubectl --insecure-skip-tls-verify get servicemonitors -n prometheus-monitoring

# Проверить, что CloudBridge сервисы мониторятся
kubectl --insecure-skip-tls-verify get servicemonitors cloudbridge-auth-service -n prometheus-monitoring -o yaml
```

## 🔄 Альтернативные решения

### Вариант 1: Увеличение ресурсов кластера
```bash
# В VK Cloud Console:
# 1. Перейти к кластеру
# 2. Вкладка "Узлы"
# 3. Изменить размер узла на STD3-4-8 (4 CPU, 8GB RAM)
# 4. Или добавить дополнительный узел
```

### Вариант 2: Установка только Prometheus
```yaml
# configs/prometheus-only.yaml
prometheus:
  prometheusSpec:
    resources:
      limits:
        cpu: 200m
        memory: 512Mi
      requests:
        cpu: 100m
        memory: 256Mi
    storageSpec:
      volumeClaimTemplate:
        spec:
          resources:
            requests:
              storage: 2Gi

grafana:
  enabled: false

alertmanager:
  enabled: false
```

### Вариант 3: Использование NodePort
```yaml
# configs/kube-prometheus-stack-nodeport.yaml
grafana:
  service:
    type: NodePort
  ingress:
    enabled: false

prometheus:
  prometheusSpec:
    serviceMonitorSelectorNilUsesHelmValues: false
    serviceMonitorSelector: {}
    serviceMonitorNamespaceSelector: {}
```

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
grafana-xxx                             1/1     Running   0          2m
prometheus-operated-xxx                 1/1     Running   0          2m

# Services
NAME                    TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)
grafana                 ClusterIP   10.254.xxx.xxx   <none>        80/TCP
prometheus-operated     ClusterIP   10.254.xxx.xxx   <none>        9090/TCP
```

### Доступ к сервисам:
- **Prometheus**: http://localhost:9090 (через port-forward)
- **Grafana**: http://localhost:3000 (через port-forward)
- **Логин Grafana**: admin / CloudBridge2025!

## 🎯 Следующие шаги

### 1. Настройка дашбордов Grafana
- Импорт дашбордов для CloudBridge
- Настройка алертов
- Конфигурация уведомлений

### 2. Интеграция с CloudBridge микросервисами
- Проверка метрик auth-service
- Проверка метрик metrics-service
- Настройка ServiceMonitors

### 3. Настройка внешнего доступа
- Установка ingress-nginx
- Настройка SSL сертификатов
- Конфигурация доменов

---

**💡 Совет**: Если установка все еще не удается, рассмотрите возможность увеличения ресурсов кластера или использования альтернативных решений мониторинга.

