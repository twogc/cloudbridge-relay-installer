# Руководство по развертыванию мультитенантной изоляции

## Быстрый старт

### **1. Предварительные требования**

```bash
# Проверяем версии
kubectl version --client
docker --version
psql --version

# Проверяем доступ к кластеру
kubectl cluster-info
kubectl get nodes
```

### **2. Развертывание за 5 шагов**

```bash
# Шаг 1: Исправляем Calico конфигурацию
kubectl apply -f k8s/calico-config-fixed.yaml

# Шаг 2: Применяем миграции БД
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_tenant_ipam_schema.sql

# Шаг 3: Собираем и разворачиваем контроллер
docker build -f Dockerfile.tenant-policy-controller -t cloudbridge/tenant-policy-controller:latest .
kubectl apply -f k8s/tenant-policy-controller.yaml

# Шаг 4: Применяем политики изоляции
kubectl apply -f k8s/tenant-isolation-policies-fixed.yaml

# Шаг 5: Настраиваем мониторинг
kubectl apply -f k8s/tenant-isolation-alerts-fixed.yaml
```

### **3. Проверка развертывания**

```bash
# Проверяем статус компонентов
kubectl get pods -n cloudbridge
kubectl get globalnetworkpolicies
kubectl get prometheusrule -n cloudbridge

# Запускаем тесты изоляции
go test ./internal/testing -v -run TestTenantIsolation

# Проверяем метрики
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_total
```

##  Конфигурация

### **Переменные окружения**

```bash
# База данных
export DB_HOST="postgres.cloudbridge.local"
export DB_USER="cloudbridge"
export DB_NAME="cloudbridge"
export DB_PASSWORD="secure-password"

# JWT секреты
export JWT_SECRET="your-jwt-secret-here"

# Мониторинг
export PROMETHEUS_URL="http://prometheus.monitoring.svc.cluster.local:9090"
export GRAFANA_URL="http://grafana.monitoring.svc.cluster.local:3000"
```

### **Настройка IP пространства**

```yaml
# config/ipam-config.yaml
ipam:
  base_subnet: "10.0.0.0/8"
  tenant_subnet_size: "/24"
  max_tenants: 65536
  salt: "cloudbridge-production-2025"
```

## Мониторинг

### **Ключевые метрики**

- `cloudbridge_tenant_total` - Общее количество тенантов
- `cloudbridge_tenant_isolation_violations_total` - Нарушения изоляции
- `cloudbridge_tenant_health` - Здоровье тенантов
- `cloudbridge_tenant_policy_reconcile_duration_seconds` - Время reconciliation

### **Алерты**

- `TenantIsolationViolation` - Нарушение изоляции между тенантами
- `TenantPolicyViolation` - Нарушение политик безопасности
- `TenantHealthDegraded` - Деградация здоровья тенанта
- `TenantIPExhaustion` - Исчерпание IP адресов

## 🧪 Тестирование

### **Автоматические тесты**

```bash
# Запуск всех тестов
go test ./internal/testing -v

# Тест изоляции
go test ./internal/testing -v -run TestTenantIsolation

# Тест IPAM
go test ./internal/ipam -v

# Тест контроллера
go test ./internal/controllers -v
```

### **Ручное тестирование**

```bash
# Создаем тестовые поды
kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: tenant-a-pod
  namespace: cloudbridge
  labels:
    tenant.cloudbridge.io/id: "tenant-test-a"
spec:
  containers:
  - name: test
    image: nginx:alpine
    ports:
    - containerPort: 80
---
apiVersion: v1
kind: Pod
metadata:
  name: tenant-b-pod
  namespace: cloudbridge
  labels:
    tenant.cloudbridge.io/id: "tenant-test-b"
spec:
  containers:
  - name: test
    image: nginx:alpine
    ports:
    - containerPort: 80
EOF

# Тестируем изоляцию
kubectl exec tenant-a-pod -- ping -c 1 tenant-b-pod
kubectl exec tenant-a-pod -- curl -s http://tenant-b-pod/health
```

## Безопасность

### **Проверка безопасности**

```bash
# Проверяем политики
kubectl get globalnetworkpolicies -o yaml | grep -E "(selector|action)"

# Проверяем RBAC
kubectl auth can-i create globalnetworkpolicies --as=system:serviceaccount:cloudbridge:tenant-policy-controller

# Проверяем логи безопасности
kubectl logs -n cloudbridge -l app=tenant-policy-controller | grep -i security
```

### **Аудит**

```bash
# Экспорт конфигурации для аудита
kubectl get globalnetworkpolicies -o yaml > audit-policies-$(date +%Y%m%d).yaml
kubectl get pods -n cloudbridge -o yaml > audit-pods-$(date +%Y%m%d).yaml
```

## Масштабирование

### **Горизонтальное масштабирование**

```bash
# Увеличиваем количество реплик контроллера
kubectl scale deployment/tenant-policy-controller -n cloudbridge --replicas=3

# Настраиваем HPA
kubectl apply -f - <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: tenant-policy-controller-hpa
  namespace: cloudbridge
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: tenant-policy-controller
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
EOF
```

### **Вертикальное масштабирование**

```bash
# Увеличиваем ресурсы
kubectl patch deployment/tenant-policy-controller -n cloudbridge -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "controller",
          "resources": {
            "requests": {"cpu": "100m", "memory": "128Mi"},
            "limits": {"cpu": "500m", "memory": "512Mi"}
          }
        }]
      }
    }
  }
}'
```

## 🆘 Устранение неполадок

### **Частые проблемы**

1. **Изоляция не работает**
   ```bash
   kubectl rollout restart daemonset/calico-node -n kube-system
   kubectl delete globalnetworkpolicies --all
   kubectl apply -f k8s/tenant-isolation-policies-fixed.yaml
   ```

2. **Контроллер не создает политики**
   ```bash
   kubectl logs -n cloudbridge -l app=tenant-policy-controller
   kubectl auth can-i create globalnetworkpolicies --as=system:serviceaccount:cloudbridge:tenant-policy-controller
   ```

3. **IPAM не выделяет IP**
   ```bash
   kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM tenant_subnets"
   ```

### **Логи и диагностика**

```bash
# Логи контроллера
kubectl logs -n cloudbridge -l app=tenant-policy-controller -f

# Логи Calico
kubectl logs -n kube-system -l k8s-app=calico-node -f

# Статус политик
kubectl get globalnetworkpolicies
calicoctl get globalnetworkpolicies
```

## 📞 Поддержка

- **Документация**: [docs/TENANT_ISOLATION_OPERATIONS_GUIDE.md](docs/TENANT_ISOLATION_OPERATIONS_GUIDE.md)
- **Техподдержка**: support@cloudbridge.io
- **Безопасность**: security@cloudbridge.io

---

**Статус**: Production Ready  
**Версия**: 1.0  
**Последнее обновление**: $(date)


