# Техническое задание: Исправление метрик Cadence

## Контекст
В системе CloudBridge Relay установлен Cadence для оркестрации workflow'ов, но сервис не запускается и не предоставляет метрики для Prometheus. Это блокирует мониторинг workflow'ов и интеграцию с AI сервисом.

## Проблема
- **Статус**: Cadence в состоянии `CrashLoopBackOff`
- **Ошибка**: `config validation failed: Must provide one of VisibilityStore and AdvancedVisibilityStore`
- **Дополнительная ошибка**: `unable to create open /etc/cadence/config/docker.yaml: read-only file system`
- **Результат**: Prometheus не может получить метрики от Cadence (статус `up = 0`)

## Текущая конфигурация

### Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cadence
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: cadence
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cadence
    spec:
      containers:
      - name: cadence
        image: ubercadence/server:v0.22.4
        ports:
        - containerPort: 7933
          name: grpc
        - containerPort: 7934
          name: web
        env:
        - name: SERVICES
          value: "history,matching,frontend,worker"
        - name: LOG_LEVEL
          value: "info"
        - name: SKIP_SCHEMA_SETUP
          value: "false"
        - name: AUTO_SETUP
          value: "true"
        - name: RINGPOP_NAME
          value: "cadence"
        - name: RINGPOP_BOOTSTRAP_MODE
          value: "hosts"
        - name: RINGPOP_BOOTSTRAP_HOSTS
          value: "cadence:7933"
        - name: DB
          value: "sqlite"
        - name: DB_PATH
          value: "/var/lib/cadence/cadence.db"
        - name: VISIBILITY_STORE
          value: "sqlite"
        - name: VISIBILITY_DB_PATH
          value: "/var/lib/cadence/visibility.db"
        - name: CLUSTER_METADATA_ENABLE_GLOBAL_DOMAIN
          value: "true"
        - name: CLUSTER_METADATA_FAILOVER_VERSION_INCREMENT
          value: "10"
        - name: CLUSTER_METADATA_MASTER_CLUSTER_NAME
          value: "active"
        - name: CLUSTER_METADATA_CURRENT_CLUSTER_NAME
          value: "active"
        - name: CONFIG_TEMPLATE_PATH
          value: ""
        - name: CONFIG_FILE_PATH
          value: ""
        - name: DOCKERIZE_TEMPLATE
          value: ""
        - name: DOCKERIZE_DELAY
          value: "0"
        - name: CADENCE_USE_ENV_VARS
          value: "true"
        - name: CADENCE_VISIBILITY_STORE
          value: "sqlite"
        - name: CADENCE_PERSISTENCE_VISIBILITY_STORE
          value: "sqlite"
        - name: CADENCE_SKIP_CONFIG_VALIDATION
          value: "true"
        volumeMounts:
        - mountPath: /var/lib/cadence
          name: cadence-data
      volumes:
      - name: cadence-data
        persistentVolumeClaim:
          claimName: cadence-pvc
```

### Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cadence-service
  namespace: cloudbridge
spec:
  selector:
    app.kubernetes.io/name: cadence
  ports:
  - port: 7933
    targetPort: 7933
    name: grpc
  - port: 7934
    targetPort: 7934
    name: web
  type: ClusterIP
```

### Prometheus Configuration
```yaml
- job_name: 'cadence'
  static_configs:
    - targets: ['cadence-service.cloudbridge.svc.cluster.local:7934']
  metrics_path: '/metrics'
  scrape_interval: 30s
  metric_relabel_configs:
    - source_labels: [__name__]
      regex: 'cadence_(workflow_executions_total|workflow_failures_total|activity_executions_total)'
      action: keep
```

## Технические требования

### 1. Исправить конфигурацию Cadence
- **Проблема**: Cadence не может найти `VisibilityStore` или `AdvancedVisibilityStore`
- **Решение**: Настроить правильную конфигурацию для SQLite visibility store
- **Ожидаемый результат**: Cadence запускается без ошибок конфигурации

### 2. Решить проблему с файловой системой
- **Проблема**: `unable to create open /etc/cadence/config/docker.yaml: read-only file system`
- **Решение**: Настроить правильные volume mounts или использовать переменные окружения
- **Ожидаемый результат**: Cadence может создавать необходимые файлы конфигурации

### 3. Настроить метрики
- **Проблема**: Cadence не предоставляет метрики на порту 7934
- **Решение**: Настроить Cadence для предоставления Prometheus метрик
- **Ожидаемый результат**: Prometheus может получить метрики от Cadence

### 4. Проверить интеграцию с AI сервисом
- **Проблема**: AI сервис не может получить статистику workflow'ов от Cadence
- **Решение**: Убедиться, что AI сервис может подключиться к Cadence
- **Ожидаемый результат**: AI сервис получает данные о workflow'ах

## Критерии приемки

### Функциональные требования
1. ✅ Cadence запускается без ошибок (`kubectl get pods -n cloudbridge | grep cadence` показывает `Running`)
2. ✅ Cadence предоставляет метрики на порту 7934 (`/metrics` endpoint доступен)
3. ✅ Prometheus может получить метрики от Cadence (статус `up = 1`)
4. ✅ AI сервис может подключиться к Cadence для получения статистики workflow'ов

### Технические требования
1. ✅ Логи Cadence не содержат ошибок конфигурации
2. ✅ Все необходимые переменные окружения настроены правильно
3. ✅ Volume mounts работают корректно
4. ✅ Service и Endpoints настроены правильно

### Метрики
1. ✅ Доступны метрики `cadence_workflow_executions_total`
2. ✅ Доступны метрики `cadence_workflow_failures_total`
3. ✅ Доступны метрики `cadence_activity_executions_total`
4. ✅ Метрики обновляются в реальном времени

## Возможные решения

### Вариант 1: Использовать переменные окружения
```yaml
env:
- name: CADENCE_PERSISTENCE_VISIBILITY_STORE
  value: "sqlite"
- name: CADENCE_PERSISTENCE_SQL_VISIBILITY_DRIVERNAME
  value: "sqlite3"
- name: CADENCE_PERSISTENCE_SQL_VISIBILITY_DATASOURCENAME
  value: "/var/lib/cadence/visibility.db"
```

### Вариант 2: Использовать ConfigMap с правильными правами
```yaml
volumes:
- name: cadence-config
  configMap:
    name: cadence-config
    defaultMode: 0644
volumeMounts:
- name: cadence-config
  mountPath: /etc/cadence/config
  readOnly: false
```

### Вариант 3: Использовать другой образ Cadence
```yaml
image: ubercadence/server:v0.22.4-alpine
# или
image: ubercadence/server:latest
```

## Тестирование

### 1. Проверка запуска
```bash
kubectl -n cloudbridge get pods | grep cadence
kubectl -n cloudbridge logs deploy/cadence --tail=50
```

### 2. Проверка метрик
```bash
kubectl -n cloudbridge exec deploy/prometheus -- wget -qO- http://cadence-service.cloudbridge.svc.cluster.local:7934/metrics
```

### 3. Проверка в Prometheus
```bash
kubectl -n cloudbridge exec deploy/ai-service -- wget -qO- http://prometheus.cloudbridge.svc.cluster.local:9090/api/v1/query?query=up | jq '.data.result[] | select(.metric.job == "cadence")'
```

### 4. Проверка интеграции с AI
```bash
kubectl -n cloudbridge exec deploy/ai-service -- wget -qO- http://cadence-service.cloudbridge.svc.cluster.local:7934/api/v1/domains
```

## Приоритет
**Высокий** - блокирует мониторинг workflow'ов и интеграцию с AI сервисом

## Временные рамки
**2-4 часа** - включает диагностику, исправление конфигурации, тестирование и проверку интеграции

## Дополнительные ресурсы
- [Cadence Configuration Reference](https://cadenceworkflow.io/docs/operation-guide/setup/)
- [Cadence Docker Configuration](https://github.com/uber/cadence/tree/master/docker)
- [Cadence Metrics](https://cadenceworkflow.io/docs/operation-guide/monitoring/)
