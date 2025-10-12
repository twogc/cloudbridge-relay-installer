# Задание: Исправление отображения Workflows в Cadence Web UI

## **Проблема**
Cadence Web UI не отображает workflows на странице `/domains/cloudbridge/primary/workflows`. Показывается ошибка "Failed to fetch workflows" или пустая страница.

## 🔍 **Что уже проверено и исправлено**

### **Выполненные исправления:**
1. **Созданы таблицы в Cassandra**:
   - `open_executions` - для активных workflow
   - `closed_executions` - для завершенных workflow
   - Все необходимые индексы

2. **Исправлена схема базы данных**:
   - Таблицы созданы в пространстве `cadence_visibility`
   - Настроены правильные индексы

3. **Настроено gRPC подключение**:
   - Cadence Web UI подключается к `10.244.4.68:7833` (gRPC порт)
   - Кластер настроен как `primary`

## 📁 **Файлы и конфигурация**

### **Kubernetes манифесты:**
```
k8s/deployments/cadence-web-deployment.yaml
k8s/services/cadence-web-service.yaml
k8s/configmaps/cadence-web-config.yaml
```

### **Cadence конфигурация:**
```
k8s/deployments/cadence-deployment.yaml
k8s/services/cadence-service.yaml
k8s/configmaps/cadence-config.yaml
```

### **Cassandra схема:**
```
/etc/cadence/schema/cassandra/visibility/schema.cql
```

##  **Что нужно проверить разработчику**

### **1. Проверить логи Cadence Web UI:**
```bash
kubectl logs deployment/cadence-web -n cloudbridge --tail=50
```

**Искать ошибки:**
- `GRPCError: Call cancelled`
- `unconfigured table open_executions`
- `unconfigured table closed_executions`
- `ListOpenWorkflowExecutions operation failed`
- `ListClosedWorkflowExecutions operation failed`

### **2. Проверить подключение к Cassandra:**
```bash
# Проверить таблицы в cadence_visibility
kubectl exec -n cloudbridge cassandra-6b99f898f9-j596t -- cqlsh -e "USE cadence_visibility; DESCRIBE TABLES;"

# Проверить содержимое таблиц
kubectl exec -n cloudbridge cassandra-6b99f898f9-j596t -- cqlsh -e "USE cadence_visibility; SELECT COUNT(*) FROM open_executions;"
kubectl exec -n cloudbridge cassandra-6b99f898f9-j596t -- cqlsh -e "USE cadence_visibility; SELECT COUNT(*) FROM closed_executions;"
```

### **3. Проверить gRPC подключение:**
```bash
# Проверить, слушает ли Cadence на gRPC порту
kubectl exec -n cloudbridge cadence-8678db5f7c-96dhk -- netstat -tlnp | grep 7833

# Проверить подключение из Cadence Web UI
kubectl exec -n cloudbridge cadence-web-<pod-name> -- nc -zv 10.244.4.68 7833
```

### **4. Проверить переменные окружения Cadence Web UI:**
```bash
kubectl exec -n cloudbridge cadence-web-<pod-name> -- env | grep CADENCE
```

**Ожидаемые значения:**
- `CADENCE_GRPC_PEERS=10.244.4.68:7833`
- `CADENCE_CLUSTERS_NAMES=primary`
- `CADENCE_GRPC_SERVICES_NAMES=cadence-frontend`

### **5. Проверить API запросы:**
```bash
# Проверить API доменов
curl -s "http://localhost:8088/api/v1/domains" | head -c 200

# Проверить API workflows
curl -s "http://localhost:8088/api/v1/domains/cloudbridge/primary/workflows" | head -c 200
```

## 🧪 **Как тестировать**

### **1. Создать тестовый workflow:**
```bash
# Попробовать создать workflow через CLI
kubectl exec -n cloudbridge cadence-8678db5f7c-96dhk -- cadence --address 10.244.4.68:7933 --domain cloudbridge workflow start --workflow_id test-workflow-3 --workflow_type TestWorkflow --tasklist test-task-list --input '{"test": "data"}'
```

### **2. Проверить в Web UI:**
1. Открыть `http://localhost:8088/domains/cloudbridge/primary/workflows`
2. Проверить, отображаются ли workflow
3. Проверить консоль браузера на ошибки JavaScript

### **3. Проверить логи в реальном времени:**
```bash
# В одном терминале
kubectl logs -f deployment/cadence-web -n cloudbridge

# В другом терминале - открыть Web UI и проверить workflow
```

## 🔨 **Как собирать и деплоить**

### **1. Собрать Cadence Web UI:**
```bash
# Если нужно пересобрать образ
docker build -t cadence-web:latest ./cadence-web/
docker tag cadence-web:latest localhost:5000/cadence-web:latest
docker push localhost:5000/cadence-web:latest
```

### **2. Применить изменения:**
```bash
# Применить конфигурацию
kubectl apply -f k8s/deployments/cadence-web-deployment.yaml -n cloudbridge

# Перезапустить deployment
kubectl rollout restart deployment/cadence-web -n cloudbridge

# Проверить статус
kubectl rollout status deployment/cadence-web -n cloudbridge --timeout=120s
```

### **3. Настроить port-forward:**
```bash
# Остановить старый port-forward
pkill -f "kubectl port-forward.*cadence-web-service"

# Запустить новый
kubectl port-forward -n cloudbridge svc/cadence-web-service 8088:8088 &
```

## 🐛 **Возможные проблемы и решения**

### **Проблема 1: Таблицы не созданы**
**Решение:**
```bash
# Создать таблицы вручную
kubectl exec -n cloudbridge cassandra-6b99f898f9-j596t -- cqlsh -e "USE cadence_visibility; CREATE TABLE open_executions (...);"
```

### **Проблема 2: gRPC подключение не работает**
**Решение:**
```bash
# Проверить правильный порт (7833 для gRPC, 7933 для RPC)
kubectl patch deployment cadence-web -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence-web","env":[{"name":"CADENCE_GRPC_PEERS","value":"10.244.4.68:7833"}]}]}}}}'
```

### **Проблема 3: Неправильный кластер**
**Решение:**
```bash
# Установить правильное имя кластера
kubectl patch deployment cadence-web -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence-web","env":[{"name":"CADENCE_CLUSTERS_NAMES","value":"primary"}]}]}}}}'
```

## 📊 **Ожидаемый результат**

После исправления:
1. Cadence Web UI загружается без ошибок
2. Страница `/domains/cloudbridge/primary/workflows` показывает список workflow
3. Можно создавать новые workflow
4. Можно просматривать детали workflow
5. Логи не содержат ошибок gRPC

##  **Полезные ссылки**

- **Cadence Web UI**: `http://localhost:8088`
- **Домены**: `http://localhost:8088/domains`
- **Workflows**: `http://localhost:8088/domains/cloudbridge/primary/workflows`
- **API доменов**: `http://localhost:8088/api/v1/domains`
- **API workflows**: `http://localhost:8088/api/v1/domains/cloudbridge/primary/workflows`

## 📝 **Примечания**

- Cadence использует gRPC для связи между компонентами
- Web UI подключается к Cadence через gRPC порт 7833
- Данные workflow хранятся в Cassandra в пространстве `cadence_visibility`
- Таблицы `open_executions` и `closed_executions` обязательны для работы
