# Отчет о проблеме с Cadence для DevOps

## Краткое описание проблемы
Cadence не может запуститься из-за проблем с правами доступа к файловой системе и конфигурацией.

## Детальное описание проблемы

### Основная ошибка
```
unable to create open /etc/cadence/config/docker.yaml: permission denied
```

### Что происходит
1. Cadence использует утилиту `dockerize` для создания конфигурационного файла
2. `dockerize` пытается создать файл `/etc/cadence/config/docker.yaml` из шаблона
3. Папка `/etc/cadence/config` монтируется как read-only ConfigMap
4. Cadence не может записать файл из-за отсутствия прав на запись

### Попытки решения

#### 1. Исправление прав доступа в minikube
```bash
minikube ssh "sudo mkdir -p /etc/cadence/config && sudo chown -R 1000:1000 /etc/cadence/config && sudo chmod -R 755 /etc/cadence/config"
```
**Результат**: Не помогло, так как ConfigMap все равно монтируется как read-only

#### 2. Использование HostPath volume
```yaml
volumes:
- name: cadence-config-host
  hostPath:
    path: /etc/cadence/config
    type: DirectoryOrCreate
```
**Результат**: Частично помогло - Cadence смог создать файл, но конфигурация была неполной

#### 3. Использование init container с emptyDir
```yaml
initContainers:
- name: setup-config-dir
  image: busybox:1.35
  command: ['sh', '-c', 'mkdir -p /etc/cadence/config && chown -R 1000:1000 /etc/cadence/config && chmod -R 755 /etc/cadence/config']
```
**Результат**: Cadence может создать папку, но не может найти файл конфигурации

#### 4. Попытки отключить dockerize
Добавлены переменные окружения:
```yaml
- name: DOCKERIZE_TEMPLATE
  value: ""
- name: DOCKERIZE_DELAY
  value: "0"
- name: CONFIG_FILE_PATH
  value: ""
- name: CONFIG_TEMPLATE_PATH
  value: ""
```
**Результат**: Не помогло - Cadence все равно пытается использовать dockerize

### Текущее состояние
- **Prometheus**: ✅ Работает (исправлена проблема с query log)
- **Grafana**: ✅ Работает (исправлены права доступа)
- **Cassandra**: ✅ Работает (keyspaces созданы)
- **CloudBridge Relay**: ✅ Работает (3 пода)
- **Cadence**: ❌ Не работает (2 пода в CrashLoopBackOff)

### Логи Cadence
```
+ CONFIG_TEMPLATE_PATH=/etc/cadence/config/config_template.yaml
+ dockerize -template /etc/cadence/config/config_template.yaml:/etc/cadence/config/docker.yaml
2025/09/07 18:12:53 unable to stat /etc/cadence/config/config_template.yaml, error: stat /etc/cadence/config/config_template.yaml: no such file or directory
```

## Рекомендации для DevOps

### Вариант 1: Использовать готовый Helm chart
```bash
helm repo add cadence https://cadenceworkflow.github.io/helm-charts
helm install cadence cadence/cadence
```

### Вариант 2: Исправить текущую конфигурацию
1. **Создать правильный init container**:
```yaml
initContainers:
- name: setup-config
  image: busybox:1.35
  command: ['sh', '-c', 'mkdir -p /etc/cadence/config && cat > /etc/cadence/config/config_template.yaml << EOF
log:
  level: info
  output: stdout

services:
  frontend:
    rpc:
      grpcPort: 7933
      membershipPort: 7933
      bindOnLocalHost: false
      bindOnIP: "0.0.0.0"
    pprof:
      port: 7934
    metrics:
      prometheus:
        timerType: histogram
        listenAddress: "0.0.0.0:9090"
    clusterMetadata:
      enableGlobalDomain: true
      failoverVersionIncrement: 10
      masterClusterName: "active"
      currentClusterName: "active"
      clusterInformation:
        active:
          enabled: true
          initialFailoverVersion: 1
          rpcName: "cadence-frontend"
          rpcAddress: "cadence.cloudbridge.svc.cluster.local:7933"
    ringpop:
      name: "cadence"
      bootstrapMode: "hosts"
      bootstrapHosts:
        - "cadence:7933"
    persistence:
      defaultStore: "cassandra-default"
      numHistoryShards: 4
      datastores:
        cassandra-default:
          cassandra:
            hosts: "cassandra-service.cloudbridge.svc.cluster.local"
            port: 9042
            keyspace: "cadence"
            datacenter: "dc1"
            maxConns: 2
            connectConsistency: "local-one"
            timeout: "600ms"
            connectTimeout: "600ms"
            maxRetryAttempts: 3
            retryPolicy: "exponential"
            convConsistency: "local-quorum"
            reconnectionPolicy: "constant"
            disableInitialHostLookup: false
            consistency:
              default:
                consistency: "local-quorum"
              serial:
                consistency: "local-serial"
              strong:
                consistency: "local-quorum"
              eventual:
                consistency: "local-one"
            auth:
              enabled: false
            protoVersion: 4
EOF
chown -R 1000:1000 /etc/cadence/config && chmod -R 755 /etc/cadence/config']
  volumeMounts:
  - name: cadence-config-dir
    mountPath: /etc/cadence/config
  securityContext:
    runAsUser: 0
```

2. **Использовать emptyDir volume**:
```yaml
volumes:
- name: cadence-config-dir
  emptyDir: {}
```

### Вариант 3: Использовать переменные окружения
Полностью отказаться от файлов конфигурации и использовать только переменные окружения. Cadence поддерживает конфигурацию через ENV переменные.

### Вариант 4: Использовать другую версию Cadence
Попробовать более старую версию, которая может лучше работать с переменными окружения:
```yaml
image: ubercadence/server:v0.22.4
```

## Файлы для исправления
- `k8s/deployments/cadence-deployment.yaml` - основной deployment
- `k8s/cadence/configmap.yaml` - конфигурация (если используется)

## Команды для тестирования
```bash
# Проверить статус подов
kubectl get pods -n cloudbridge

# Посмотреть логи Cadence
kubectl logs -l app.kubernetes.io/name=cadence -n cloudbridge --tail=50

# Проверить подключение к Cassandra
kubectl exec -it cassandra-6dc4499998-5qgl9 -n cloudbridge -- cqlsh -e "DESCRIBE KEYSPACES;"

# Перезапустить deployment
kubectl rollout restart deployment/cadence -n cloudbridge
```

## Приоритет
**Высокий** - Cadence критически важен для работы системы workflow.

## Статус
**В работе** - требуется вмешательство DevOps для окончательного решения проблемы с конфигурацией.
