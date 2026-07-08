# Tenant Policy Controller & IPAM System

## Обзор

Этот проект содержит улучшенную систему изоляции тенантов и управления IP-адресами для CloudBridge.

## Компоненты

### 1. Tenant Policy Controller (`cmd/tenant-policy-controller/`)
Kubernetes контроллер для автоматического создания сетевых политик Calico для изоляции тенантов.

**Ключевые улучшения:**
- Исправлен селектор подов (использует "label exists")
- езопасный cleanup по индексу
- демпотентный CreateOrUpdate
- NS правила с правильными портами/протоколами
- eader election для HA
- ealth/ready checks

### 2. IPAM Repository (`internal/ipam/`)
PostgreSQL-based система управления IP-адресами с детерминированной аллокацией.

**Ключевые улучшения:**
- ипы данных CIDR/INET вместо TEXT
- ривязка IP к подсетям
- ostgreSQL функции с FOR UPDATE SKIP LOCKED
- сключенные CIDR для избежания конфликтов
- равильная обработка транзакций

## Быстрый старт

### Развертывание в Minikube

1. **Сборка контроллера:**
```bash
docker build -f Dockerfile.tenant-policy-controller -t tenant-policy-controller:latest .
minikube image load tenant-policy-controller:latest
```

2. **Развертывание PostgreSQL:**
```bash
kubectl apply -f k8s/postgres-init-scripts.yaml
kubectl apply -f k8s/postgres-deployment.yaml
```

3. **Развертывание контроллера:**
```bash
kubectl apply -f k8s/tenant-policy-controller.yaml
```

4. **Тестирование изоляции:**
```bash
kubectl apply -f k8s/test-pods.yaml
```

## Тестирование

### IPAM функции
```bash
# Создание тенанта
kubectl exec -n cloudbridge-system ipam-test-client -- psql -c "
INSERT INTO tenants (tenant_id, status) VALUES ('test-tenant-001', 'active');"

# Создание подсети
kubectl exec -n cloudbridge-system ipam-test-client -- psql -c "
SELECT create_tenant_subnet('test-tenant-001', '10.100.1.0/24', '10.100.1.1');"

# Аллокация IP
kubectl exec -n cloudbridge-system ipam-test-client -- psql -c "
SELECT allocate_ip_for_peer('test-tenant-001', 'peer-001', '10.100.1.0/24');"
```

### Изоляция тенантов
```bash
# Проверка связи между подами одного тенанта
kubectl exec -n tenant-a-test test-pod-tenant-a -- ping -c 3 10.244.6.2

# Проверка изоляции между тенантами
kubectl exec -n tenant-a-test test-pod-tenant-a -- ping -c 3 10.244.6.4
```

## 📁 Структура файлов

```
cmd/tenant-policy-controller/     # Основное приложение контроллера
internal/controllers/              # Логика контроллера
internal/ipam/                     # IPAM репозиторий
k8s/                              # Kubernetes манифесты
migrations/                       # SQL миграции
```

##  Конфигурация

### Переменные окружения
- `WATCH_NAMESPACE` - namespace для наблюдения (пустой = весь кластер)
- `METRICS_BIND_ADDRESS` - адрес для метрик (по умолчанию :8080)
- `HEALTH_PROBE_BIND_ADDRESS` - адрес для health checks (по умолчанию :8081)

### RBAC
Контроллер требует следующие права:
- `pods`: get, list, watch
- `globalnetworkpolicies`: get, list, watch, create, update, patch, delete
- `leases`: get, list, watch, create, update, patch, delete

## 🐛 Известные проблемы

1. **API версия Calico**: Контроллер использует `projectcalico.org/v3`, но кластер может использовать `crd.projectcalico.org/v1`
2. **Изоляция между тенантами**: Требует доработки политик для полной изоляции

## Мониторинг

### Метрики
- `/metrics` - Prometheus метрики
- `/healthz` - Health check
- `/readyz` - Readiness check

### Логи
```bash
kubectl logs -n cloudbridge-system deployment/tenant-policy-controller
```

## Миграции

SQL миграции находятся в `migrations/`:
- `001_tenant_ipam_schema.sql` - Базовая схема
- `002_tenant_ipam_improved_schema.sql` - Улучшенная схема
- `003_test_sql_functions.sql` - Тестовые функции

## Документация

- `CONTROLLER_IMPROVEMENTS_REPORT.md` - Отчет об улучшениях контроллера
- `PRODUCTION_READY_SUMMARY.md` - Готовность к продакшену
- `DEPLOYMENT_GUIDE.md` - Руководство по развертыванию
