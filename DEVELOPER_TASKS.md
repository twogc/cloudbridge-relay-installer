# Задания для разработчика - CloudBridge Relay

## 🎯 Общий статус проекта

### ✅ Что работает:
- **CloudBridge Relay** - 3 пода работают стабильно
- **Cassandra** - 1 под работает (использует PV на 20Gi)
- **Cadence** - настроен с PV на 100Gi (как требовалось)

### ❌ Что нужно исправить:
1. **Prometheus** - падает с ошибкой query log
2. **Grafana** - падает с ошибкой прав доступа
3. **Cadence** - 2 пода падают (нужно диагностировать)

---

## 🔧 Задание 1: Исправить Prometheus

### Проблема:
Prometheus падает с ошибкой:
```
ERROR source=query_logger.go:113 msg="Error opening query log file" 
component=activeQueryTracker file=/prometheus/queries.active 
err="open /prometheus/queries.active: permission denied"
panic: Unable to create mmap-ed active query log
```

### Файлы для исправления:
- `k8s/monitoring/prometheus-deployment.yaml` (строки 29-45)

### Текущая конфигурация:
```yaml
containers:
- name: prometheus
  image: prom/prometheus:latest
  args:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus/'
    - '--web.console.libraries=/etc/prometheus/console_libraries'
    - '--web.console.templates=/etc/prometheus/consoles'
    - '--storage.tsdb.retention.time=200h'
    - '--web.enable-lifecycle'
    - '--query.max-concurrency=0'
    - '--query.max-samples=0'
    - '--query.lookback-delta=5m'
    - '--query.timeout=2m'
    - '--query.active-query-tracker-path=/tmp'
```

### Возможные решения:
1. **Отключить активный query tracker полностью:**
   ```yaml
   - '--query.active-query-tracker-path=""'
   ```

2. **Использовать более старую версию Prometheus:**
   ```yaml
   image: prom/prometheus:v2.40.0
   ```

3. **Исправить права доступа к директории:**
   ```yaml
   initContainers:
   - name: fix-permissions
     image: busybox:1.35
     command: ['sh', '-c', 'mkdir -p /prometheus && chown -R 65534:65534 /prometheus && chmod -R 755 /prometheus']
     volumeMounts:
     - name: prometheus-storage-volume
       mountPath: /prometheus
     securityContext:
       runAsUser: 0
   ```

---

## 🔧 Задание 2: Исправить Grafana

### Проблема:
Grafana падает с ошибкой:
```
GF_PATHS_DATA='/var/lib/grafana' is not writable.
mkdir: can't create directory '/var/lib/grafana/plugins': Permission denied
```

### Файлы для исправления:
- `k8s/monitoring/grafana-deployment.yaml` (строки 18-50)

### Текущая конфигурация:
```yaml
initContainers:
- name: fix-permissions
  image: busybox:1.35
  command: ['sh', '-c', 'chown -R 0:0 /var/lib/grafana && chmod -R 755 /var/lib/grafana']
  volumeMounts:
  - name: grafana-storage
    mountPath: /var/lib/grafana
  securityContext:
    runAsUser: 0
containers:
- name: grafana
  image: grafana/grafana:latest
  securityContext:
    runAsUser: 0
```

### Возможные решения:
1. **Использовать правильного пользователя Grafana:**
   ```yaml
   securityContext:
     runAsUser: 472
     runAsGroup: 472
     fsGroup: 472
   ```

2. **Исправить init container:**
   ```yaml
   command: ['sh', '-c', 'chown -R 472:472 /var/lib/grafana && chmod -R 755 /var/lib/grafana']
   ```

---

## 🔧 Задание 3: Исправить Cadence

### Проблема:
2 пода Cadence падают, нужно диагностировать причину.

### Файлы для диагностики:
- `k8s/deployments/cadence-deployment.yaml`
- `k8s/cadence/configmap.yaml`

### Команды для диагностики:
```bash
# Проверить статус подов
kubectl get pods -n cloudbridge

# Посмотреть логи Cadence
kubectl logs cadence-99c44d9b8-zbrdw -n cloudbridge --tail=20
kubectl logs cadence-d4c7676c6-2gzk7 -n cloudbridge --tail=20

# Проверить конфигурацию
kubectl describe pod cadence-99c44d9b8-zbrdw -n cloudbridge
```

### Возможные проблемы:
1. **Проблемы с подключением к Cassandra**
2. **Неправильная конфигурация Cadence**
3. **Проблемы с правами доступа к PV**

---

## 📁 Структура файлов

### Мониторинг:
```
k8s/monitoring/
├── prometheus-deployment.yaml    # Основной манифест Prometheus
├── prometheus-config.yaml        # Конфигурация Prometheus
├── grafana-deployment.yaml       # Основной манифест Grafana
└── grafana-datasources.yaml      # Источники данных Grafana
```

### Основные сервисы:
```
k8s/deployments/
├── cloudbridge-relay-deployment.yaml  # ✅ Работает
├── cassandra-deployment.yaml          # ✅ Работает (использует cassandra-minimal.yaml)
├── cassandra-minimal.yaml             # ✅ Работает
└── cadence-deployment.yaml            # ❌ Нужно исправить
```

### Конфигурация:
```
k8s/configmaps/
├── cloudbridge-config.yaml       # ✅ Работает
└── cassandra-config.yaml         # ✅ Работает

k8s/cadence/
└── configmap.yaml                # ❌ Нужно проверить
```

### Хранилище:
```
k8s/persistent-volumes/
└── cloudbridge-pv.yaml           # ✅ Работает (Cassandra 20Gi, Cadence 100Gi)

k8s/storage-classes.yaml          # ✅ Работает
```

---

## 🚀 Команды для тестирования

### Проверить статус всех подов:
```bash
kubectl get pods -n cloudbridge
kubectl get pods -n cloudbridge-monitoring
```

### Перезапустить сервисы:
```bash
kubectl rollout restart deployment/prometheus -n cloudbridge-monitoring
kubectl rollout restart deployment/grafana -n cloudbridge-monitoring
kubectl rollout restart deployment/cadence -n cloudbridge
```

### Применить изменения:
```bash
kubectl apply -f k8s/monitoring/prometheus-deployment.yaml
kubectl apply -f k8s/monitoring/grafana-deployment.yaml
kubectl apply -f k8s/deployments/cadence-deployment.yaml
```

### Проверить логи:
```bash
kubectl logs -f deployment/prometheus -n cloudbridge-monitoring
kubectl logs -f deployment/grafana -n cloudbridge-monitoring
kubectl logs -f deployment/cadence -n cloudbridge
```

---

## 🎯 Ожидаемый результат

После исправления всех проблем:

1. **Prometheus** должен запуститься и собирать метрики
2. **Grafana** должен запуститься и подключиться к Prometheus
3. **Cadence** должен запуститься и подключиться к Cassandra
4. **Все сервисы** должны быть доступны через port-forward:
   - Prometheus: `kubectl port-forward -n cloudbridge-monitoring service/prometheus 9090:9090`
   - Grafana: `kubectl port-forward -n cloudbridge-monitoring service/grafana 3000:3000`
   - Cadence: `kubectl port-forward -n cloudbridge service/cadence-service 7933:7933`

---

## 📝 Приоритеты

1. **Высокий**: Исправить Prometheus (критично для мониторинга)
2. **Высокий**: Исправить Grafana (нужен для визуализации)
3. **Средний**: Исправить Cadence (нужен для workflow)

---

## 🔍 Дополнительная информация

- **Namespace**: `cloudbridge` (основные сервисы), `cloudbridge-monitoring` (мониторинг)
- **Minikube**: Используется для локальной разработки
- **Storage**: PV/PVC настроены для всех сервисов
- **Сеть**: Все сервисы доступны через ClusterIP

---

*Создано: 2025-09-07*
*Статус: В разработке*
