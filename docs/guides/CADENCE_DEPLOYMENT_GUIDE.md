# 🚀 Полное руководство по развертыванию Cadence в Minikube

## 📋 Содержание

1. [Предварительные требования](#предварительные-требования)
2. [Быстрый старт](#быстрый-старт)
3. [Развертывание базы данных](#развертывание-базы-данных)
4. [Развертывание Cadence](#развертывание-cadence)
5. [Проверка работы](#проверка-работы)
6. [Частые ошибки и решения](#частые-ошибки-и-решения)
7. [Мониторинг и отладка](#мониторинг-и-отладка)
8. [Альтернативные решения](#альтернативные-решения)

##  Предварительные требования

### Системные требования
- **Minikube**: версия 1.28+
- **kubectl**: версия 1.28+
- **Docker**: версия 20.10+
- **RAM**: минимум 4GB (рекомендуется 8GB)
- **CPU**: минимум 2 ядра

### Проверка установки
```bash
# Проверить Minikube
minikube version

# Проверить kubectl
kubectl version --client

# Проверить Docker
docker --version
```

## ⚡ Быстрый старт

### 1. Запуск Minikube
```bash
# Запустить Minikube с достаточными ресурсами
minikube start --driver=docker --memory=4096 --cpus=2

# Проверить статус
minikube status
```

### 2. Создание namespace
```bash
# Создать namespace
kubectl create namespace cloudbridge

# Или использовать существующий
kubectl get namespaces | grep cloudbridge
```

### 3. Развертывание (выберите один вариант)

#### Вариант A: Автоматический (рекомендуется)
```bash
cd /home/ubuntu/cloudbridge-relay-installer
./scripts/start-cadence-minikube.sh
```

#### Вариант B: Ручной
```bash
# PostgreSQL + Cadence
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
kubectl apply -f k8s/services/cadence-nodeport.yaml -n cloudbridge
```

## 🗄️ Развертывание базы данных

### PostgreSQL (рекомендуется для разработки)

#### Конфигурация
```yaml
# k8s/deployments/postgres-simple.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: postgres
  template:
    metadata:
      labels:
        app.kubernetes.io/name: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:13
        env:
        - name: POSTGRES_DB
          value: "cadence"
        - name: POSTGRES_USER
          value: "cadence"
        - name: POSTGRES_PASSWORD
          value: "cadence"
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-data
        persistentVolumeClaim:
          claimName: postgres-pvc
```

#### Развертывание PostgreSQL
```bash
# Применить конфигурацию
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge -l app.kubernetes.io/name=postgres

# Проверить логи
kubectl logs -f deployment/postgres -n cloudbridge
```

### Cassandra (для продакшена)

#### Конфигурация
```yaml
# k8s/deployments/cassandra-simple.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cassandra
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: cassandra
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cassandra
    spec:
      containers:
      - name: cassandra
        image: cassandra:3.11
        env:
        - name: CASSANDRA_CLUSTER_NAME
          value: "cloudbridge-cluster"
        - name: CASSANDRA_DC
          value: "datacenter1"
        - name: CASSANDRA_ENDPOINT_SNITCH
          value: "SimpleSnitch"
        - name: CASSANDRA_SEEDS
          value: "127.0.0.1"
        ports:
        - containerPort: 9042
        volumeMounts:
        - name: cassandra-data
          mountPath: /var/lib/cassandra
      volumes:
      - name: cassandra-data
        persistentVolumeClaim:
          claimName: cassandra-pvc
```

## 🎯 Развертывание Cadence

### Конфигурация с PostgreSQL
```yaml
# k8s/deployments/cadence-working.yaml
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
        - containerPort: 7933  # gRPC
        - containerPort: 7934  # Web UI
        env:
        # Основные настройки
        - name: SERVICES
          value: "history,matching,frontend,worker"
        - name: LOG_LEVEL
          value: "info"
        - name: SKIP_SCHEMA_SETUP
          value: "false"
        - name: AUTO_SETUP
          value: "true"
        
        # База данных - SQLite (проще для разработки)
        - name: DB
          value: "sqlite"
        - name: DB_PATH
          value: "/var/lib/cadence/cadence.db"
        
        # Visibility Store
        - name: VISIBILITY_STORE
          value: "sqlite"
        - name: VISIBILITY_DB_PATH
          value: "/var/lib/cadence/visibility.db"
        
        # Ringpop
        - name: RINGPOP_NAME
          value: "cadence"
        - name: RINGPOP_BOOTSTRAP_MODE
          value: "hosts"
        - name: RINGPOP_BOOTSTRAP_HOSTS
          value: "cadence:7933"
        
        # Cluster metadata
        - name: CLUSTER_METADATA_ENABLE_GLOBAL_DOMAIN
          value: "true"
        - name: CLUSTER_METADATA_FAILOVER_VERSION_INCREMENT
          value: "10"
        - name: CLUSTER_METADATA_MASTER_CLUSTER_NAME
          value: "active"
        - name: CLUSTER_METADATA_CURRENT_CLUSTER_NAME
          value: "active"
        
        # Отключить конфигурационный файл
        - name: CONFIG_TEMPLATE_PATH
          value: ""
        - name: CONFIG_FILE_PATH
          value: ""
        - name: DOCKERIZE_TEMPLATE
          value: ""
        - name: DOCKERIZE_DELAY
          value: "0"
        
        volumeMounts:
        - name: cadence-data
          mountPath: /var/lib/cadence
      volumes:
      - name: cadence-data
        persistentVolumeClaim:
          claimName: cadence-pvc
```

### Развертывание Cadence
```bash
# Применить конфигурацию
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cadence

# Проверить логи
kubectl logs -f deployment/cadence -n cloudbridge
```

### Создание сервисов
```bash
# Применить сервисы
kubectl apply -f k8s/services/cadence-nodeport.yaml -n cloudbridge

# Проверить сервисы
kubectl get services -n cloudbridge
```

## ✅ Проверка работы

### 1. Проверка статуса подов
```bash
# Проверить все поды
kubectl get pods -n cloudbridge

# Проверить конкретно Cadence
kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cadence

# Проверить события
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

### 2. Проверка логов
```bash
# Логи Cadence
kubectl logs -f deployment/cadence -n cloudbridge

# Логи базы данных
kubectl logs -f deployment/postgres -n cloudbridge
```

### 3. Проверка доступности
```bash
# Получить IP Minikube
MINIKUBE_IP=$(minikube ip)

# Получить порты сервиса
CADENCE_GRPC_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[0].nodePort}')
CADENCE_WEB_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[1].nodePort}')

echo "Cadence gRPC: $MINIKUBE_IP:$CADENCE_GRPC_PORT"
echo "Cadence Web UI: http://$MINIKUBE_IP:$CADENCE_WEB_PORT"
```

### 4. Тестирование API
```bash
# Port forwarding для тестирования
kubectl port-forward service/cadence-nodeport 7933:7933 -n cloudbridge &
kubectl port-forward service/cadence-nodeport 7934:7934 -n cloudbridge &

# Тестировать Web UI
curl http://localhost:7934/health
```

## 🐛 Частые ошибки и решения

### 1. Проблемы с базой данных

#### Ошибка: `CrashLoopBackOff` для PostgreSQL
```bash
# Симптомы
kubectl get pods -n cloudbridge
# NAME                       READY   STATUS             RESTARTS
# postgres-xxx               0/1     CrashLoopBackOff   3

# Диагностика
kubectl logs postgres-xxx -n cloudbridge

# Решения:
# 1. Проверить PVC
kubectl get pvc -n cloudbridge
kubectl describe pvc postgres-pvc -n cloudbridge

# 2. Удалить и пересоздать PVC
kubectl delete pvc postgres-pvc -n cloudbridge
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge

# 3. Проверить storage class
kubectl get storageclass
```

#### Ошибка: `CrashLoopBackOff` для Cassandra
```bash
# Симптомы
kubectl logs cassandra-xxx -n cloudbridge
# ERROR: The seed provider lists no seeds.

# Решения:
# 1. Использовать SQLite вместо Cassandra
kubectl delete deployment cassandra -n cloudbridge
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge

# 2. Или исправить конфигурацию Cassandra
# Изменить CASSANDRA_SEEDS на правильный IP
```

### 2. Проблемы с Cadence

#### Ошибка: `config validation failed: Must provide one of VisibilityStore and AdvancedVisibilityStore`
```bash
# Симптомы
kubectl logs cadence-xxx -n cloudbridge
# config validation failed: Must provide one of VisibilityStore and AdvancedVisibilityStore

# Решение: Добавить переменную окружения
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","env":[{"name":"VISIBILITY_STORE","value":"sqlite"},{"name":"VISIBILITY_DB_PATH","value":"/var/lib/cadence/visibility.db"}]}]}}}}'
```

#### Ошибка: `unable to create SQL connection: dial tcp [::1]:5432: connection refused`
```bash
# Симптомы
kubectl logs cadence-xxx -n cloudbridge
# unable to create SQL connection: dial tcp [::1]:5432: connection refused

# Решения:
# 1. Использовать SQLite вместо PostgreSQL
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge

# 2. Или исправить DB_HOST в PostgreSQL конфигурации
# Изменить на: postgres-service.cloudbridge.svc.cluster.local
```

#### Ошибка: `Loading configFiles=[/etc/cadence/config/docker.yaml]`
```bash
# Симптомы
kubectl logs cadence-xxx -n cloudbridge
# Loading configFiles=[/etc/cadence/config/docker.yaml]

# Решение: Отключить конфигурационный файл
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","env":[{"name":"CONFIG_TEMPLATE_PATH","value":""},{"name":"CONFIG_FILE_PATH","value":""}]}]}}}}'
```

### 3. Проблемы с ресурсами

#### Ошибка: `Insufficient memory`
```bash
# Симптомы
kubectl describe pod cadence-xxx -n cloudbridge
# Warning  FailedScheduling  pod/cadence-xxx 0/1 nodes are available: 1 Insufficient memory.

# Решение: Увеличить память Minikube
minikube stop
minikube start --memory=8192 --cpus=4
```

#### Ошибка: `ImagePullBackOff`
```bash
# Симптомы
kubectl get pods -n cloudbridge
# NAME                       READY   STATUS             RESTARTS
# cadence-xxx                0/1     ImagePullBackOff   3

# Решения:
# 1. Проверить доступность образа
docker pull ubercadence/server:v0.22.4

# 2. Использовать локальный образ
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","imagePullPolicy":"IfNotPresent"}]}}}}'
```

### 4. Проблемы с сетью

#### Ошибка: `Service not accessible`
```bash
# Симптомы
curl http://localhost:7934/health
# curl: (7) Failed to connect to localhost port 7934: Connection refused

# Решения:
# 1. Проверить port forwarding
kubectl port-forward service/cadence-nodeport 7934:7934 -n cloudbridge

# 2. Проверить сервисы
kubectl get services -n cloudbridge

# 3. Использовать NodePort напрямую
MINIKUBE_IP=$(minikube ip)
CADENCE_WEB_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[1].nodePort}')
curl http://$MINIKUBE_IP:$CADENCE_WEB_PORT/health
```

## 📊 Мониторинг и отладка

### 1. Основные команды мониторинга
```bash
# Статус всех ресурсов
kubectl get all -n cloudbridge

# Детальная информация о поде
kubectl describe pod -l app.kubernetes.io/name=cadence -n cloudbridge

# Логи в реальном времени
kubectl logs -f deployment/cadence -n cloudbridge

# Использование ресурсов
kubectl top pods -n cloudbridge
kubectl top nodes
```

### 2. Проверка здоровья
```bash
# Health check через kubectl
kubectl exec -it deployment/cadence -n cloudbridge -- curl http://localhost:7934/health

# Health check через port forwarding
kubectl port-forward service/cadence-nodeport 7934:7934 -n cloudbridge &
curl http://localhost:7934/health
```

### 3. Отладка подключений
```bash
# Проверить endpoints
kubectl get endpoints -n cloudbridge

# Проверить DNS
kubectl exec -it deployment/cadence -n cloudbridge -- nslookup postgres-service

# Проверить сетевые политики
kubectl get networkpolicies -n cloudbridge
```

## 🔄 Альтернативные решения

### 1. Docker Compose (для локальной разработки)
```yaml
# docker-compose.yml
version: '3.8'
services:
  cadence:
    image: ubercadence/server:v0.22.4
    ports:
      - "7933:7933"
      - "7934:7934"
    environment:
      - SERVICES=history,matching,frontend,worker
      - LOG_LEVEL=info
      - SKIP_SCHEMA_SETUP=false
      - AUTO_SETUP=true
      - DB=sqlite
      - DB_PATH=/var/lib/cadence/cadence.db
      - VISIBILITY_STORE=sqlite
      - VISIBILITY_DB_PATH=/var/lib/cadence/visibility.db
      - RINGPOP_NAME=cadence
      - RINGPOP_BOOTSTRAP_MODE=hosts
      - RINGPOP_BOOTSTRAP_HOSTS=cadence:7933
      - CLUSTER_METADATA_ENABLE_GLOBAL_DOMAIN=true
      - CLUSTER_METADATA_FAILOVER_VERSION_INCREMENT=10
      - CLUSTER_METADATA_MASTER_CLUSTER_NAME=active
      - CLUSTER_METADATA_CURRENT_CLUSTER_NAME=active
      - CONFIG_TEMPLATE_PATH=
      - CONFIG_FILE_PATH=
      - DOCKERIZE_TEMPLATE=
      - DOCKERIZE_DELAY=0
    volumes:
      - cadence-data:/var/lib/cadence

volumes:
  cadence-data:
```

Запуск:
```bash
docker-compose up -d
```

### 2. Helm Chart (для продакшена)
```bash
# Добавить репозиторий Cadence
helm repo add cadence https://cadenceworkflow.github.io/helm-charts
helm repo update

# Установить Cadence
helm install cadence cadence/cadence \
  --namespace cloudbridge \
  --create-namespace \
  --set server.image.tag=v0.22.4 \
  --set server.config.database.type=sqlite \
  --set server.config.database.sqlite.path=/var/lib/cadence/cadence.db
```

### 3. Готовые образы с предустановленной конфигурацией
```bash
# Создать собственный образ
docker build -f docker/Dockerfile.cadence -t cloudbridge-cadence:latest .

# Развернуть с собственным образом
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","image":"cloudbridge-cadence:latest"}]}}}}'
```

## 🧹 Очистка

### Полная очистка
```bash
# Удалить все ресурсы
kubectl delete namespace cloudbridge

# Остановить Minikube
minikube stop

# Очистить образы
docker rmi ubercadence/server:v0.22.4
docker rmi postgres:13
```

### Частичная очистка
```bash
# Удалить только Cadence
kubectl delete deployment cadence -n cloudbridge
kubectl delete service cadence-nodeport -n cloudbridge

# Удалить только базу данных
kubectl delete deployment postgres -n cloudbridge
kubectl delete pvc postgres-pvc -n cloudbridge
```

## 📚 Полезные ссылки

- [Cadence Documentation](https://cadenceworkflow.io/docs/)
- [Cadence Docker Hub](https://hub.docker.com/r/ubercadence/server)
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

## 🤝 Поддержка

Если у вас возникли проблемы:

1. Проверьте логи: `kubectl logs -f deployment/cadence -n cloudbridge`
2. Проверьте статус: `kubectl get all -n cloudbridge`
3. Проверьте события: `kubectl get events -n cloudbridge --sort-by='.lastTimestamp'`
4. Создайте issue в репозитории проекта

---

**🎯 Результат: Полнофункциональный Cadence развернут в Minikube с подробной документацией по устранению неполадок!**
