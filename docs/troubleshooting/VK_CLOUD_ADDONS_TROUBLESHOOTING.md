# 🔧 Устранение проблем с установкой аддонов VK Cloud

## 🚨 Проблема: kube-prometheus-stack не устанавливается

### Диагностика проблемы

#### 1. Проверка статуса кластера
```bash
# Установить переменную окружения
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml

# Проверить подключение к кластеру
kubectl --insecure-skip-tls-verify cluster-info
```

#### 2. Проверка событий
```bash
# Общие события кластера
kubectl --insecure-skip-tls-verify get events --all-namespaces --sort-by='.lastTimestamp' | tail -20

# События в namespace prometheus-monitoring
kubectl --insecure-skip-tls-verify get events -n prometheus-monitoring --sort-by='.lastTimestamp'
```

#### 3. Проверка ресурсов
```bash
# Проверить namespace
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus

# Проверить все ресурсы в namespace
kubectl --insecure-skip-tls-verify get all -n prometheus-monitoring
```

## 🔍 Возможные причины и решения

### 1. Проблема с ресурсами кластера

#### Проверка доступных ресурсов
```bash
# Проверить узлы кластера
kubectl --insecure-skip-tls-verify get nodes

# Проверить использование ресурсов
kubectl --insecure-skip-tls-verify top nodes
kubectl --insecure-skip-tls-verify top pods --all-namespaces
```

#### Решение: Увеличить ресурсы кластера
1. Перейти в VK Cloud Console
2. Выбрать кластер `kubernetes-cluster-8926`
3. Перейти на вкладку "Узлы"
4. Увеличить количество узлов или ресурсы существующих узлов

### 2. Проблема с квотой floating IP

#### Диагностика
```bash
# Проверить события LoadBalancer
kubectl --insecure-skip-tls-verify get events --all-namespaces | grep -i "floatingip\|quota"
```

#### Решение: Запросить увеличение квоты
1. Обратиться в поддержку VK Cloud
2. Запросить увеличение квоты floating IP
3. Или использовать NodePort вместо LoadBalancer

### 3. Проблема с образами контейнеров

#### Диагностика
```bash
# Проверить поды с проблемами ImagePullBackOff
kubectl --insecure-skip-tls-verify get pods --all-namespaces | grep -i "ImagePullBackOff\|ErrImagePull"
```

#### Решение: Использовать локальные образы
```bash
# Создать простые образы для тестирования
cat > Dockerfile.test << EOF
FROM alpine:latest
RUN echo "Test service running" > /tmp/status
CMD ["sh", "-c", "while true; do echo 'Service is running'; sleep 30; done"]
EOF

# Собрать образ
podman build -t localhost/cloudbridge/test-service -f Dockerfile.test .

# Сохранить образ в tar
podman save localhost/cloudbridge/test-service -o test-service.tar
```

## 🛠️ Пошаговое решение

### Шаг 1: Очистка неудачной установки
```bash
# Удалить namespace prometheus-monitoring
kubectl --insecure-skip-tls-verify delete namespace prometheus-monitoring

# Подождать полного удаления
kubectl --insecure-skip-tls-verify get namespaces | grep prometheus
```

### Шаг 2: Проверка ресурсов кластера
```bash
# Проверить доступные ресурсы
kubectl --insecure-skip-tls-verify describe nodes

# Убедиться, что есть достаточно ресурсов для kube-prometheus-stack
# Минимум: 2 CPU, 4GB RAM, 20GB дискового пространства
```

### Шаг 3: Установка с минимальной конфигурацией
```yaml
# configs/kube-prometheus-stack-minimal.yaml
nameOverride: cloudbridge-prometheus
fullnameOverride: cloudbridge-prometheus

# Минимальные ресурсы
grafana:
  resources:
    limits:
      cpu: 200m
      memory: 256Mi
    requests:
      cpu: 100m
      memory: 128Mi
  persistence:
    enabled: false  # Отключить persistent storage

prometheus:
  prometheusSpec:
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
      requests:
        cpu: 200m
        memory: 512Mi
    storageSpec:
      volumeClaimTemplate:
        spec:
          resources:
            requests:
              storage: 5Gi  # Уменьшить размер

alertmanager:
  alertmanagerSpec:
    resources:
      limits:
        cpu: 200m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi
    storage:
      volumeClaimTemplate:
        spec:
          resources:
            requests:
              storage: 1Gi  # Уменьшить размер
```

### Шаг 4: Альтернативная установка через Helm
```bash
# Добавить репозиторий Prometheus
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Установить с минимальной конфигурацией
helm install cloudbridge-prometheus prometheus-community/kube-prometheus-stack \
  --namespace prometheus-monitoring \
  --create-namespace \
  --values configs/kube-prometheus-stack-minimal.yaml
```

## 🔄 Альтернативные решения

### 1. Установка только Prometheus (без Grafana)
```yaml
# configs/prometheus-only.yaml
prometheus:
  prometheusSpec:
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
      requests:
        cpu: 200m
        memory: 512Mi
    storageSpec:
      volumeClaimTemplate:
        spec:
          resources:
            requests:
              storage: 5Gi

grafana:
  enabled: false

alertmanager:
  enabled: false
```

### 2. Использование NodePort вместо LoadBalancer
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

alertmanager:
  service:
    type: NodePort
```

### 3. Установка компонентов по отдельности
```bash
# 1. Установить только Prometheus
helm install prometheus prometheus-community/prometheus \
  --namespace prometheus-monitoring \
  --create-namespace \
  --set server.resources.limits.cpu=500m \
  --set server.resources.limits.memory=1Gi

# 2. Установить Grafana отдельно
helm install grafana grafana/grafana \
  --namespace prometheus-monitoring \
  --set resources.limits.cpu=200m \
  --set resources.limits.memory=256Mi
```

## 📊 Мониторинг установки

### Проверка статуса установки
```bash
# Проверить поды
kubectl --insecure-skip-tls-verify get pods -n prometheus-monitoring

# Проверить сервисы
kubectl --insecure-skip-tls-verify get svc -n prometheus-monitoring

# Проверить события
kubectl --insecure-skip-tls-verify get events -n prometheus-monitoring --sort-by='.lastTimestamp'
```

### Проверка работоспособности
```bash
# Проверить Prometheus
kubectl --insecure-skip-tls-verify port-forward svc/prometheus-operated 9090:9090 -n prometheus-monitoring

# Проверить Grafana
kubectl --insecure-skip-tls-verify port-forward svc/grafana 3000:80 -n prometheus-monitoring
```

## 🚨 Критические проблемы

### 1. Недостаточно ресурсов
**Симптомы**: Pods в состоянии Pending, ошибки "Insufficient memory/cpu"
**Решение**: Увеличить ресурсы кластера или уменьшить запросы ресурсов в конфигурации

### 2. Проблемы с сетью
**Симптомы**: Pods не могут связаться друг с другом, ошибки DNS
**Решение**: Проверить настройки сети кластера, CNI плагины

### 3. Проблемы с хранилищем
**Симптомы**: Pods не могут создать PersistentVolumeClaims
**Решение**: Проверить StorageClass, доступность storage

## 📞 Поддержка VK Cloud

Если проблемы не решаются, обратиться в поддержку VK Cloud:

1. **Через консоль**: Создать тикет в разделе "Поддержка"
2. **Email**: support@vk.com
3. **Telegram**: @vkcloud_support

### Информация для поддержки:
- ID кластера: `kubernetes-cluster-8926`
- Версия Kubernetes: (указать версию)
- Описание проблемы: "Не удается установить kube-prometheus-stack"
- Логи ошибок: (приложить логи)

---

**💡 Совет**: Начните с минимальной конфигурации и постепенно увеличивайте ресурсы по мере необходимости.

