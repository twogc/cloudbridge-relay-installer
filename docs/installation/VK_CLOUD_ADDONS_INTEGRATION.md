# 🚀 Интеграция аддонов VK Cloud для CloudBridge Relay

## 📊 Доступные аддоны VK Cloud

### ✅ Рекомендуемые для развертывания:

#### 1. **kube-prometheus-stack** (КРИТИЧНО)
- **Назначение**: Система мониторинга на базе Prometheus, Alertmanager и Grafana
- **Приоритет**: Высокий
- **Зачем**: Мониторинг микросервисов, метрики, алерты
- **Преимущества**: Готовые дашборды, алерты, интеграция с Kubernetes

#### 2. **ingress-nginx** (КРИТИЧНО)
- **Назначение**: Управление и маршрутизация трафика
- **Приоритет**: Высокий
- **Зачем**: Внешний доступ к сервисам, SSL termination
- **Преимущества**: Автоматическая настройка, SSL сертификаты

#### 3. **cert-manager** (ВАЖНО)
- **Назначение**: Контроллер сертификатов для Kubernetes
- **Приоритет**: Средний
- **Зачем**: Автоматические SSL сертификаты Let's Encrypt
- **Преимущества**: Автоматическое обновление сертификатов

#### 4. **docker-registry** (ВАЖНО)
- **Назначение**: Приватный репозиторий для Docker образов
- **Приоритет**: Средний
- **Зачем**: Хранение образов микросервисов
- **Преимущества**: Интеграция с VK Cloud, безопасность

#### 5. **jaeger** (ОПЦИОНАЛЬНО)
- **Назначение**: Отслеживание и устранение неисправностей в микросервисах
- **Приоритет**: Низкий
- **Зачем**: Distributed tracing, отладка
- **Преимущества**: Визуализация запросов между сервисами

## 🛠️ План интеграции

### Этап 1: Критические аддоны (День 1)

#### 1.1 Установка kube-prometheus-stack
```bash
# В VK Cloud Console:
# 1. Перейти к Kubernetes кластеру
# 2. Выбрать "Аддоны"
# 3. Установить "kube-prometheus-stack"
# 4. Настроить параметры:
#    - Namespace: cloudbridge-monitoring
#    - Grafana Admin Password: secure-password
#    - Retention: 30d
```

#### 1.2 Установка ingress-nginx
```bash
# В VK Cloud Console:
# 1. Установить "ingress-nginx"
# 2. Настроить параметры:
#    - LoadBalancer IP: использовать внешний IP
#    - SSL: включить
#    - Metrics: включить
```

### Этап 2: Важные аддоны (День 2)

#### 2.1 Установка cert-manager
```bash
# В VK Cloud Console:
# 1. Установить "cert-manager"
# 2. Настроить ClusterIssuer для Let's Encrypt
# 3. Создать сертификаты для доменов
```

#### 2.2 Установка docker-registry
```bash
# В VK Cloud Console:
# 1. Установить "docker-registry"
# 2. Настроить параметры:
#    - Storage: VK Cloud Storage
#    - Authentication: включить
#    - SSL: включить
```

### Этап 3: Опциональные аддоны (День 3)

#### 3.1 Установка jaeger
```bash
# В VK Cloud Console:
# 1. Установить "jaeger"
# 2. Настроить параметры:
#    - Storage: Elasticsearch
#    - Sampling: 100%
```

## 🔧 Конфигурация после установки

### 1. Настройка Prometheus для мониторинга микросервисов

```yaml
# k8s/monitoring/prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
  namespace: cloudbridge-monitoring
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    scrape_configs:
      - job_name: 'kubernetes-pods'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
      
      - job_name: 'auth-service'
        static_configs:
          - targets: ['auth-service.cloudbridge.svc.cluster.local:9090']
      
      - job_name: 'metrics-service'
        static_configs:
          - targets: ['metrics-service.cloudbridge.svc.cluster.local:9090']
```

### 2. Настройка Grafana дашбордов

```yaml
# k8s/monitoring/grafana-dashboards.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: grafana-dashboards
  namespace: cloudbridge-monitoring
data:
  cloudbridge-overview.json: |
    {
      "dashboard": {
        "title": "CloudBridge Relay Overview",
        "panels": [
          {
            "title": "Auth Service Requests",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total{service=\"auth-service\"}[5m])",
                "legendFormat": "{{method}} {{endpoint}}"
              }
            ]
          },
          {
            "title": "Metrics Service Requests",
            "type": "graph",
            "targets": [
              {
                "expr": "rate(http_requests_total{service=\"metrics-service\"}[5m])",
                "legendFormat": "{{method}} {{endpoint}}"
              }
            ]
          },
          {
            "title": "Active Connections",
            "type": "stat",
            "targets": [
              {
                "expr": "cloudbridge_connections_active",
                "legendFormat": "Active Connections"
              }
            ]
          },
          {
            "title": "Active Tunnels",
            "type": "stat",
            "targets": [
              {
                "expr": "cloudbridge_tunnels_active",
                "legendFormat": "Active Tunnels"
              }
            ]
          }
        ]
      }
    }
```

### 3. Настройка алертов

```yaml
# k8s/monitoring/alertmanager-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alertmanager-config
  namespace: cloudbridge-monitoring
data:
  alertmanager.yml: |
    global:
      smtp_smarthost: 'smtp.gmail.com:587'
      smtp_from: 'alerts@cloudbridge.local'
      smtp_auth_username: 'your-email@gmail.com'
      smtp_auth_password: 'your-app-password'
    
    route:
      group_by: ['alertname']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 1h
      receiver: 'team-cloudbridge'
    
    receivers:
      - name: 'team-cloudbridge'
        email_configs:
          - to: 'team@cloudbridge.local'
    
    inhibit_rules:
      - source_match:
          severity: 'critical'
        target_match:
          severity: 'warning'
        equal: ['alertname']
```

### 4. Настройка Ingress с SSL

```yaml
# k8s/cloudbridge-ingress-ssl.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudbridge-ingress-ssl
  namespace: cloudbridge
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - auth.cloudbridge.local
    - metrics.cloudbridge.local
    - api.cloudbridge.local
    secretName: cloudbridge-tls
  rules:
  - host: auth.cloudbridge.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 8080
  - host: metrics.cloudbridge.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: metrics-service
            port:
              number: 8084
  - host: api.cloudbridge.local
    http:
      paths:
      - path: /auth
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 8080
      - path: /metrics
        pathType: Prefix
        backend:
          service:
            name: metrics-service
            port:
              number: 8084
```

### 5. Настройка Docker Registry

```yaml
# k8s/registry-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: registry-config
  namespace: cloudbridge
data:
  config.yml: |
    version: 0.1
    log:
      level: debug
    storage:
      filesystem:
        rootdirectory: /var/lib/registry
      delete:
        enabled: true
    http:
      addr: :5000
      headers:
        X-Content-Type-Options: [nosniff]
    auth:
      htpasswd:
        realm: basic-realm
        path: /etc/registry/htpasswd
    health:
      storagedriver:
        enabled: true
        interval: 10s
        threshold: 3
```

## 📊 Дашборды и мониторинг

### 1. Grafana дашборды
- **CloudBridge Overview**: Общий обзор всех сервисов
- **Auth Service**: Детальная метрика аутентификации
- **Metrics Service**: Метрики и производительность
- **Infrastructure**: Состояние Kubernetes кластера

### 2. Prometheus метрики
- **HTTP запросы**: Количество и время ответа
- **Активные подключения**: Текущие пользователи
- **Активные туннели**: WireGuard туннели
- **Ошибки**: Количество ошибок по типам
- **Ресурсы**: CPU, Memory, Disk usage

### 3. Алерты
- **Критические**: Сервисы недоступны
- **Предупреждения**: Высокая нагрузка
- **Информационные**: Обновления и события

## 🚀 Преимущества интеграции

### 1. **Автоматизация**
- Автоматические SSL сертификаты
- Автоматическое масштабирование
- Автоматические бэкапы

### 2. **Мониторинг**
- Готовые дашборды
- Автоматические алерты
- Distributed tracing

### 3. **Безопасность**
- SSL/TLS шифрование
- Приватный registry
- RBAC политики

### 4. **Производительность**
- Оптимизированные образы
- Кэширование
- Load balancing

## 📋 Чек-лист установки

### ✅ Этап 1: Критические аддоны
- [ ] Установить kube-prometheus-stack
- [ ] Установить ingress-nginx
- [ ] Настроить базовый мониторинг
- [ ] Протестировать доступ к Grafana

### ✅ Этап 2: Важные аддоны
- [ ] Установить cert-manager
- [ ] Установить docker-registry
- [ ] Настроить SSL сертификаты
- [ ] Настроить приватный registry

### ✅ Этап 3: Опциональные аддоны
- [ ] Установить jaeger
- [ ] Настроить distributed tracing
- [ ] Создать дополнительные дашборды
- [ ] Настроить расширенные алерты

---

**🎯 Цель**: Полная интеграция с VK Cloud экосистемой для production-ready развертывания

