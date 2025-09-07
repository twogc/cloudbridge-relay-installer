# DevOps Onboarding Task - CloudBridge Relay

## Цель задания
Ознакомиться с проектом CloudBridge Relay, развернуть локальную среду разработки и проверить работу всех компонентов системы.

## Описание проекта
CloudBridge Relay - это высокопроизводительная платформа для P2P mesh сетей с поддержкой WireGuard, QUIC, TLS 1.3 и Cadence Workflow Engine. Проект включает 8 микросервисов и более 15,000 строк кода на Go.

## Архитектура системы
- **CloudBridge Relay**: Основной сервис для P2P mesh сетей
- **Cassandra**: База данных для хранения данных Cadence
- **Cadence**: Workflow Engine для оркестрации процессов
- **Prometheus + Grafana**: Мониторинг и визуализация метрик
- **Kubernetes**: Оркестрация контейнеров

## Задание

### 1. Подготовка окружения
```bash
# Убедитесь, что у вас установлены:
# - Docker
# - kubectl
# - minikube
# - Go (опционально, для разработки)

# Проверьте версии
docker --version
kubectl version --client
minikube version
```

### 2. Запуск minikube
```bash
# Удалите существующий кластер (если есть)
minikube delete

# Запустите minikube с оптимальными ресурсами
minikube start --memory=4096 --cpus=3 --driver=docker

# Проверьте статус
minikube status
kubectl get nodes
```

### 3. Развертывание Cassandra
```bash
# Создайте namespace для Cassandra
kubectl apply -f k8s/cassandra/namespace.yaml

# Создайте PersistentVolume и PersistentVolumeClaim
kubectl apply -f k8s/cassandra/pv.yaml

# Разверните Cassandra
kubectl apply -f k8s/cassandra/statefulset-simple.yaml
kubectl apply -f k8s/cassandra/service.yaml

# Проверьте статус
kubectl get pods -n cassandra
kubectl wait --for=condition=ready pod/cassandra-0 -n cassandra --timeout=300s

# Проверьте подключение к Cassandra
kubectl exec -it cassandra-0 -n cassandra -- cqlsh -e "DESCRIBE KEYSPACES;"
```

### 4. Развертывание Cadence
```bash
# Создайте namespace для Cadence
kubectl apply -f k8s/cadence/namespace.yaml

# Разверните Cadence frontend
kubectl apply -f k8s/cadence/configmap.yaml
kubectl apply -f k8s/cadence/frontend-deployment.yaml
kubectl apply -f k8s/cadence/frontend-service.yaml

# Проверьте статус
kubectl get pods -n cadence
kubectl get services -n cadence
```

### 5. Развертывание CloudBridge Relay
```bash
# Создайте namespace для production
kubectl apply -f k8s/production/namespace.yaml

# Создайте RBAC
kubectl apply -f k8s/production/rbac.yaml

# Создайте ConfigMap
kubectl apply -f k8s/production/configmap.yaml

# Соберите Docker образ
docker build -f Dockerfile.production -t cloudbridge-relay:production .

# Загрузите образ в minikube
minikube image load cloudbridge-relay:production

# Разверните CloudBridge Relay
kubectl apply -f k8s/production/deployment.yaml
kubectl apply -f k8s/production/service.yaml

# Проверьте статус
kubectl get pods -n cloudbridge-production
```

### 6. Развертывание мониторинга
```bash
# Создайте namespace для мониторинга
kubectl apply -f k8s/production/monitoring-namespace.yaml

# Разверните Prometheus
kubectl apply -f k8s/production/prometheus-config.yaml
kubectl apply -f k8s/production/prometheus-deployment.yaml

# Разверните Grafana
kubectl apply -f k8s/production/grafana-deployment.yaml

# Проверьте статус
kubectl get pods -n monitoring-production
```

### 7. Проверка работы системы
```bash
# Проверьте все поды
kubectl get pods -A

# Проверьте сервисы
kubectl get services -A

# Проверьте PersistentVolumes
kubectl get pv,pvc -A

# Проверьте логи CloudBridge Relay
kubectl logs -n cloudbridge-production -l app.kubernetes.io/name=cloudbridge-relay --tail=20
```

### 8. Тестирование функциональности
```bash
# Настройте port-forward для доступа к сервисам
kubectl port-forward -n cloudbridge-production service/cloudbridge-relay-production-internal 9090:9090 &
kubectl port-forward -n cloudbridge-production service/cloudbridge-relay-production-internal 8084:8084 &
kubectl port-forward -n monitoring-production service/prometheus 9091:9090 &
kubectl port-forward -n monitoring-production service/grafana 3001:3000 &

# Проверьте health endpoint
curl http://localhost:9090/health

# Проверьте metrics endpoint
curl http://localhost:8084/metrics

# Проверьте Prometheus
curl http://localhost:9091/targets

# Проверьте Grafana (admin/admin)
curl http://localhost:3001
```

### 9. Запуск нагрузочного тестирования
```bash
# Установите k6 (если не установлен)
# Ubuntu/Debian: sudo apt-get install k6
# Или скачайте с https://k6.io/docs/getting-started/installation/

# Запустите нагрузочный тест
k6 run production-load-test.js
```

### 10. Инициализация Cadence домена
```bash
# Создайте keyspace для Cadence в Cassandra
kubectl exec -it cassandra-0 -n cassandra -- cqlsh -e "
CREATE KEYSPACE IF NOT EXISTS cadence WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE KEYSPACE IF NOT EXISTS cadence_visibility WITH REPLICATION = {'class': 'SimpleStrategy', 'replication_factor': 1};
"

# Инициализируйте домен CloudBridge
kubectl exec -it cassandra-0 -n cassandra -- cqlsh -e "
USE cadence;
SOURCE '/etc/cassandra/cassandra.cql';
"
```

## Ожидаемые результаты

После выполнения задания вы должны увидеть:

1. **Kubernetes кластер**: minikube с 3 CPU и 4GB RAM
2. **Cassandra**: 1 под в namespace `cassandra`
3. **Cadence**: 1 под frontend в namespace `cadence`
4. **CloudBridge Relay**: 3 пода в namespace `cloudbridge-production`
5. **Мониторинг**: Prometheus и Grafana в namespace `monitoring-production`
6. **Health checks**: Все сервисы отвечают на health endpoints
7. **Metrics**: Prometheus собирает метрики со всех сервисов
8. **Load test**: k6 тест проходит успешно

## Полезные команды для отладки

```bash
# Просмотр логов
kubectl logs -n <namespace> <pod-name> --tail=50

# Описание ресурса
kubectl describe pod -n <namespace> <pod-name>

# Проверка событий
kubectl get events -n <namespace> --sort-by='.lastTimestamp'

# Проверка ресурсов
kubectl top nodes
kubectl top pods -A

# Очистка ресурсов
kubectl delete namespace cassandra cadence cloudbridge-production monitoring-production
```

## Дополнительные задачи (опционально)

1. **Масштабирование**: Попробуйте увеличить количество реплик CloudBridge Relay
2. **Обновление**: Обновите образ CloudBridge Relay и выполните rolling update
3. **Backup**: Настройте backup для Cassandra
4. **Security**: Примените Network Policies и Pod Security Policies
5. **CI/CD**: Настройте GitHub Actions для автоматического развертывания

## Время выполнения
Ожидаемое время выполнения: 2-3 часа для новичка, 1-2 часа для опытного DevOps инженера.

## Критерии успеха
- Все поды в состоянии Running
- Health endpoints отвечают 200 OK
- Metrics собираются в Prometheus
- Load test проходит без ошибок
- Grafana доступна и показывает дашборды

Удачи в изучении CloudBridge Relay! 🚀
