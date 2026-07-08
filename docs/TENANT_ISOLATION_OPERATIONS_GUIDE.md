# Руководство по эксплуатации мультитенантной изоляции CloudBridge

## Содержание

1. [Матрица допусков](#матрица-допусков)
2. [Процедуры развертывания](#процедуры-развертывания)
3. [Процедуры мониторинга](#процедуры-мониторинга)
4. [Процедуры устранения неполадок](#процедуры-устранения-неполадок)
5. [Процедуры масштабирования](#процедуры-масштабирования)
6. [Процедуры безопасности](#процедуры-безопасности)

---

## Матрица допусков

### **Разрешенные соединения**

| Источник | Назначение | Протокол | Порт | Статус | Обоснование |
|----------|------------|----------|------|--------|-------------|
| **Внутритенантное взаимодействие** |
| Tenant A Pod | Tenant A Pod | HTTP/HTTPS | 80, 443 | Разрешено | Внутренняя коммуникация |
| Tenant A Pod | Tenant A Pod | TCP | Любой | Разрешено | Внутренняя коммуникация |
| Tenant A Pod | Tenant A Pod | UDP | Любой | Разрешено | Внутренняя коммуникация |
| Tenant A Pod | Tenant A Pod | ICMP | - | Разрешено | Диагностика |
| **Системные сервисы** |
| Tenant Pod | kube-dns | UDP | 53 | Разрешено | DNS разрешение |
| Tenant Pod | kube-dns | TCP | 53 | Разрешено | DNS разрешение |
| Tenant Pod | kube-system | HTTP | 443 | Разрешено | API сервер |
| Tenant Pod | Prometheus | HTTP | 9090 | Разрешено | Мониторинг |
| Tenant Pod | Grafana | HTTP | 3000 | Разрешено | Дашборды |
| **Межтенантная изоляция** |
| Tenant A Pod | Tenant B Pod | HTTP/HTTPS | Любой |  Заблокировано | Изоляция тенантов |
| Tenant A Pod | Tenant B Pod | TCP | Любой |  Заблокировано | Изоляция тенантов |
| Tenant A Pod | Tenant B Pod | UDP | Любой |  Заблокировано | Изоляция тенантов |
| Tenant A Pod | Tenant B Pod | ICMP | - |  Заблокировано | Изоляция тенантов |

### **Запрещенные соединения**

| Источник | Назначение | Протокол | Порт | Статус | Обоснование |
|----------|------------|----------|------|--------|-------------|
| Tenant Pod | Metadata Service | HTTP | 169.254.169.254 |  Заблокировано | Защита от SSRF |
| Tenant Pod | External Registry | HTTP/HTTPS | 443 | ⚠️ Условно | Требует настройки |
| Tenant Pod | Internet | Любой | Любой | ⚠️ Условно | Требует настройки |

---

## Процедуры развертывания

### **1. Развертывание Calico с правильной конфигурацией**

```bash
# 1. Применяем исправленную конфигурацию Calico
kubectl apply -f k8s/calico-config-fixed.yaml

# 2. Проверяем статус Calico
kubectl get pods -n kube-system -l k8s-app=calico-node
kubectl get pods -n kube-system -l k8s-app=calico-typha

# 3. Проверяем, что Global Network Policies работают
kubectl get globalnetworkpolicies
```

### **2. Развертывание Tenant Policy Controller**

```bash
# 1. Собираем образ контроллера
docker build -f Dockerfile.tenant-policy-controller -t cloudbridge/tenant-policy-controller:latest .

# 2. Применяем манифесты
kubectl apply -f k8s/tenant-policy-controller.yaml

# 3. Проверяем статус контроллера
kubectl get pods -n cloudbridge -l app=tenant-policy-controller
kubectl logs -n cloudbridge -l app=tenant-policy-controller
```

### **3. Развертывание IPAM с миграциями БД**

```bash
# 1. Применяем миграции БД
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_tenant_ipam_schema.sql

# 2. Проверяем созданные таблицы
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt tenant*"

# 3. Тестируем IPAM функции
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT allocate_ip_for_peer('test-tenant', 'test-peer', '10.1.1.0/24');"
```

### **4. Развертывание мониторинга и алертов**

```bash
# 1. Применяем исправленные алерты
kubectl apply -f k8s/tenant-isolation-alerts-fixed.yaml

# 2. Проверяем ServiceMonitor
kubectl get servicemonitor -n cloudbridge

# 3. Проверяем PrometheusRule
kubectl get prometheusrule -n cloudbridge
```

---

## Процедуры мониторинга

### **1. Проверка статуса изоляции**

```bash
# Проверяем активные политики
kubectl get globalnetworkpolicies | grep tenant

# Проверяем метрики изоляции
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_isolation_violations_total

# Проверяем здоровье тенантов
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_health
```

### **2. Мониторинг производительности**

```bash
# Проверяем время reconciliation
curl -s http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,cloudbridge_tenant_policy_reconcile_duration_seconds_bucket)

# Проверяем время выделения IP
curl -s http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,cloudbridge_tenant_ip_allocation_duration_seconds_bucket)

# Проверяем использование IP адресов
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_allocated_ips_total
```

### **3. Проверка алертов**

```bash
# Проверяем активные алерты
kubectl get prometheusrule -n cloudbridge -o yaml

# Проверяем статус алертов в Prometheus
curl -s http://localhost:9090/api/v1/alerts | jq '.data.alerts[] | select(.labels.alertname | contains("Tenant"))'
```

---

##  Процедуры устранения неполадок

### **1. Проблема: Изоляция не работает**

**Симптомы:**
- Ping между разными тенантами проходит
- HTTP соединения между тенантами работают

**Диагностика:**
```bash
# 1. Проверяем статус Calico
kubectl get pods -n kube-system -l k8s-app=calico-node
kubectl logs -n kube-system -l k8s-app=calico-node | grep -i policy

# 2. Проверяем политики
kubectl get globalnetworkpolicies
calicoctl get globalnetworkpolicies

# 3. Проверяем селекторы
kubectl get pods --show-labels | grep tenant.cloudbridge.io/id
```

**Решение:**
```bash
# 1. Перезапускаем Calico
kubectl rollout restart daemonset/calico-node -n kube-system

# 2. Пересоздаем политики
kubectl delete globalnetworkpolicies --all
kubectl apply -f k8s/tenant-isolation-policies-fixed.yaml

# 3. Перезапускаем контроллер
kubectl rollout restart deployment/tenant-policy-controller -n cloudbridge
```

### **2. Проблема: Контроллер не создает политики**

**Симптомы:**
- Нет политик для новых тенантов
- Ошибки в логах контроллера

**Диагностика:**
```bash
# 1. Проверяем логи контроллера
kubectl logs -n cloudbridge -l app=tenant-policy-controller

# 2. Проверяем RBAC
kubectl auth can-i create globalnetworkpolicies --as=system:serviceaccount:cloudbridge:tenant-policy-controller

# 3. Проверяем CRD
kubectl get crd globalnetworkpolicies.crd.projectcalico.org
```

**Решение:**
```bash
# 1. Проверяем RBAC
kubectl apply -f k8s/tenant-policy-controller.yaml

# 2. Перезапускаем контроллер
kubectl rollout restart deployment/tenant-policy-controller -n cloudbridge

# 3. Проверяем схему
kubectl get globalnetworkpolicies
```

### **3. Проблема: IPAM не выделяет IP**

**Симптомы:**
- Ошибки при выделении IP
- Дублирование IP адресов

**Диагностика:**
```bash
# 1. Проверяем подключение к БД
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1"

# 2. Проверяем таблицы IPAM
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM tenant_subnets"

# 3. Проверяем коллизии
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT ip_address, COUNT(*) FROM tenant_ip_allocations GROUP BY ip_address HAVING COUNT(*) > 1"
```

**Решение:**
```bash
# 1. Очищаем дублированные записи
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM tenant_ip_allocations WHERE id NOT IN (SELECT MIN(id) FROM tenant_ip_allocations GROUP BY ip_address)"

# 2. Пересоздаем IP tracking
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "UPDATE ip_usage_tracking SET status = 'available' WHERE status = 'allocated'"

# 3. Перезапускаем сервис
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
```

---

## Процедуры масштабирования

### **1. Масштабирование IP пространства**

```bash
# 1. Проверяем текущее использование
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_subnets_total

# 2. Если приближаемся к лимиту, расширяем пространство
# Изменяем generateSubnetCIDR для использования 10.0.0.0/8 вместо 10.100.0.0/16

# 3. Применяем изменения
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
```

### **2. Масштабирование контроллера**

```bash
# 1. Увеличиваем количество реплик контроллера
kubectl scale deployment/tenant-policy-controller -n cloudbridge --replicas=3

# 2. Настраиваем leader election
kubectl patch deployment/tenant-policy-controller -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"controller","env":[{"name":"LEADER_ELECT","value":"true"}]}]}}}}'
```

### **3. Масштабирование мониторинга**

```bash
# 1. Увеличиваем retention для метрик
kubectl patch prometheus/prometheus -n monitoring -p '{"spec":{"retention":"30d"}}'

# 2. Настраиваем агрегацию метрик
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-rules
  namespace: monitoring
data:
  tenant-aggregation.yml: |
    groups:
    - name: tenant.aggregation
      rules:
      - record: cloudbridge:tenant_isolation_violations:rate5m
        expr: rate(cloudbridge_tenant_isolation_violations_total[5m])
EOF
```

---

## Процедуры безопасности

### **1. Аудит политик безопасности**

```bash
# 1. Проверяем все активные политики
kubectl get globalnetworkpolicies -o yaml > /tmp/current-policies.yaml

# 2. Проверяем соответствие стандартам
calicoctl get globalnetworkpolicies -o yaml | grep -E "(selector|action)" > /tmp/policy-audit.txt

# 3. Проверяем логи нарушений
kubectl logs -n cloudbridge -l app=tenant-policy-controller | grep -i violation
```

### **2. Ротация секретов**

```bash
# 1. Генерируем новые JWT секреты
openssl rand -base64 32 > /tmp/new-jwt-secret

# 2. Обновляем секреты
kubectl create secret generic jwt-secret --from-file=secret=/tmp/new-jwt-secret -n cloudbridge --dry-run=client -o yaml | kubectl apply -f -

# 3. Перезапускаем сервисы
kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge
```

### **3. Мониторинг безопасности**

```bash
# 1. Настраиваем алерты на нарушения
kubectl apply -f k8s/tenant-isolation-alerts-fixed.yaml

# 2. Проверяем логи безопасности
kubectl logs -n cloudbridge -l app=cloudbridge-relay | grep -E "(security|violation|breach)"

# 3. Экспортируем метрики безопасности
curl -s http://localhost:9090/api/v1/query?query=cloudbridge_tenant_policy_violations_total > /tmp/security-metrics.json
```

---

## Дополнительные ресурсы

### **Полезные команды**

```bash
# Проверка статуса всей системы
kubectl get pods -n cloudbridge
kubectl get globalnetworkpolicies
kubectl get prometheusrule -n cloudbridge

# Мониторинг в реальном времени
watch -n 5 'kubectl get globalnetworkpolicies | wc -l'

# Экспорт конфигурации
kubectl get globalnetworkpolicies -o yaml > backup-policies-$(date +%Y%m%d).yaml
```

### **Контакты для поддержки**

- **Техническая поддержка**: support@cloudbridge.io
- **Безопасность**: security@cloudbridge.io
- **Документация**: docs@cloudbridge.io

### **Ссылки на документацию**

- [Calico Network Policies](https://docs.projectcalico.org/security/policy)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Prometheus Alerting](https://prometheus.io/docs/alerting/latest/overview/)

---

**Последнее обновление**: $(date)  
**Версия документа**: 1.0  
**Статус**: Production Ready


