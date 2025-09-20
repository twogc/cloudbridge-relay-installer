# 🚀 **Kubernetes Configuration - Документация**

## 📋 **Обзор**

Документация по конфигурации Kubernetes для CloudBridge Relay. Включает deployment, services, configmaps, network policies и другие ресурсы.

---

## 🏗️ **Основные компоненты**

### **Deployment**
- `cloudbridge-relay` - Основной сервис P2P API
- `ai-service` - AI Service для анализа и оптимизации
- `ai-analytics` - AI Analytics для аналитики
- `ai-worker` - AI Worker для фоновых задач
- `cadence` - Cadence Workflow Engine
- `cadence-web` - Cadence Web UI

### **Services**
- `cloudbridge-relay-external` - External service для P2P API
- `ai-service` - Service для AI Service
- `ai-analytics-service` - Service для AI Analytics
- `ai-worker-service` - Service для AI Worker
- `cadence-service` - Service для Cadence
- `cadence-web-service` - Service для Cadence Web UI

### **ConfigMaps**
- `ai-config` - Конфигурация AI компонентов
- `ai-analytics-config` - Конфигурация AI Analytics
- `ai-service-endpoints` - AI Service API endpoints
- `ai-analytics-endpoints` - AI Analytics API endpoints
- `monitoring-config` - Конфигурация мониторинга

---

## 📁 **Структура файлов**

```
k8s/
├── deployment.yaml                    # Основной deployment
├── cloudbridge-relay-external-service.yaml  # External service
├── network-policy-prometheus.yaml     # Сетевые политики
├── prometheus-service.yaml           # Prometheus service
├── ai-service-endpoints.yaml         # AI Service endpoints
├── ai-analytics-endpoints.yaml       # AI Analytics endpoints
└── monitoring-config.yaml            # Конфигурация мониторинга
```

---

## 🔧 **Основной Deployment**

### **deployment.yaml**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
  labels:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/component: p2p-api
    app.kubernetes.io/part-of: cloudbridge-relay
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: cloudbridge-relay
      app.kubernetes.io/component: p2p-api
  template:
    metadata:
      labels:
        app.kubernetes.io/name: cloudbridge-relay
        app.kubernetes.io/component: p2p-api
        app.kubernetes.io/part-of: cloudbridge-relay
    spec:
      containers:
      - name: cloudbridge-relay
        image: cloudbridge-relay:cadence-p2p-v1
        imagePullPolicy: Never
        ports:
        - containerPort: 8082
          name: http
        - containerPort: 8080
          name: metrics
        env:
        - name: CADENCE_ENABLED
          value: "true"
        - name: CADENCE_HOST_PORT
          value: "cadence:7933"
        - name: CADENCE_DOMAIN
          value: "cloudbridge-p2p"
        - name: CADENCE_TASK_LIST
          value: "p2p-quic-tasks"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## 🌐 **Services**

### **cloudbridge-relay-external-service.yaml**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-external
  namespace: cloudbridge
  labels:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/component: p2p-api
    app.kubernetes.io/part-of: cloudbridge-relay
spec:
  type: NodePort
  ports:
  - port: 8082
    targetPort: 8082
    nodePort: 32500
    name: http
  - port: 8080
    targetPort: 8080
    nodePort: 32501
    name: metrics
  selector:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/component: p2p-api
```

### **AI Services**
```yaml
# AI Service
apiVersion: v1
kind: Service
metadata:
  name: ai-service
  namespace: cloudbridge
  labels:
    app.kubernetes.io/component: ai
    app.kubernetes.io/name: ai-service
    app.kubernetes.io/part-of: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8083
    targetPort: 8083
    name: http
  selector:
    app.kubernetes.io/component: ai
    app.kubernetes.io/name: ai-service

# AI Analytics Service
apiVersion: v1
kind: Service
metadata:
  name: ai-analytics-service
  namespace: cloudbridge
  labels:
    app.kubernetes.io/component: ai
    app.kubernetes.io/name: ai-analytics
    app.kubernetes.io/part-of: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8084
    targetPort: 8084
    name: http
  selector:
    app.kubernetes.io/component: ai
    app.kubernetes.io/name: ai-analytics
```

---

## 🔧 **ConfigMaps**

### **AI Configuration**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-config
  namespace: cloudbridge
  labels:
    app.kubernetes.io/component: ai
    app.kubernetes.io/name: ai-config
    app.kubernetes.io/part-of: cloudbridge-relay
data:
  ai-config.yaml: |
    ai:
      enabled: true
      mode: "kubernetes"
      models_path: "/app/models"
      redis_address: "redis-service:6379"
      redis_password: ""
      redis_db: 0
      kubernetes:
        namespace: "cloudbridge"
        monitoring_interval: "30s"
        metrics_collection: true
        auto_scaling: true
      anomaly_detection:
        enabled: true
        threshold: 0.8
        learning_rate: 0.01
        max_samples: 10000
        contamination: 0.1
        model_file: "anomaly_detector.gob"
        kubernetes_integration: true
        retrain_interval: "1h"
      predictive_analytics:
        enabled: true
        algorithm: "arima"
        forecast_horizon: 24
        confidence_level: 0.8
        model_file: "load_predictor.gob"
        retrain_interval: "6h"
      intelligent_routing:
        enabled: true
        algorithm: "multi_armed_bandit"
        exploration_rate: 0.1
        learning_rate: 0.05
        update_interval: "5m"
        model_file: "route_optimizer.gob"
        kubernetes_aware: true
      security_intelligence:
        enabled: true
        model_file: "security_threat_model.gob"
        threat_detection: true
        auto_response: false
      data_pipeline:
        enabled: true
        batch_size: 100
        processing_interval: "10s"
        prometheus_url: "http://prometheus-service:9090"
        retention_days: 30
        kubernetes_metrics: true
        prometheus_metrics: true
      monitoring:
        enabled: true
        metrics_endpoint: "/metrics"
        health_endpoint: "/health"
        cadence_integration: true
        prometheus_integration: true
```

---

## 🔒 **Network Policies**

### **network-policy-prometheus.yaml**
```yaml
# Egress политика для AI компонентов
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-prometheus-access
  namespace: cloudbridge
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/component: ai
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: cloudbridge-monitoring
    ports:
    - protocol: TCP
      port: 9090
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53

# Ingress политика для Prometheus
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-prometheus-ingress
  namespace: cloudbridge-monitoring
spec:
  podSelector:
    matchLabels:
      app: prometheus
  policyTypes:
  - Ingress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: cloudbridge
    ports:
    - protocol: TCP
      port: 9090
```

---

## 📊 **Monitoring Configuration**

### **monitoring-config.yaml**
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: monitoring-config
  namespace: cloudbridge
  labels:
    app.kubernetes.io/component: monitoring
    app.kubernetes.io/name: monitoring-config
    app.kubernetes.io/part-of: cloudbridge-relay
data:
  prometheus-rules.yaml: |
    groups:
    - name: cloudbridge-alerts
      rules:
      - alert: HighLatency
        expr: ai_prediction_accuracy < 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "AI prediction accuracy is low"
          description: "AI prediction accuracy is {{ $value }} which is below threshold"
      
      - alert: AIServiceDown
        expr: up{job="ai-service"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "AI Service is down"
          description: "AI Service has been down for more than 1 minute"
      
      - alert: HighMemoryUsage
        expr: ai_memory_usage_bytes / 1024 / 1024 / 1024 > 1.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "AI Service high memory usage"
          description: "AI Service memory usage is {{ $value }}GB"
      
      - alert: HighCPUUsage
        expr: ai_cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "AI Service high CPU usage"
          description: "AI Service CPU usage is {{ $value }}%"
      
      - alert: AnomaliesDetected
        expr: increase(ai_anomalies_detected_total[5m]) > 10
        for: 1m
        labels:
          severity: warning
        annotations:
          summary: "High number of anomalies detected"
          description: "{{ $value }} anomalies detected in the last 5 minutes"
      
      - alert: CadenceWorkflowFailed
        expr: increase(cadence_workflow_failed_total[5m]) > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Cadence workflow failed"
          description: "{{ $value }} Cadence workflows failed in the last 5 minutes"
```

---

## 🚀 **Развертывание**

### **1. Создание namespace:**
```bash
kubectl create namespace cloudbridge
```

### **2. Применение конфигураций:**
```bash
# Основной deployment
kubectl apply -f k8s/deployment.yaml -n cloudbridge

# External service
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Сетевые политики
kubectl apply -f k8s/network-policy-prometheus.yaml

# Prometheus service
kubectl apply -f k8s/prometheus-service.yaml

# AI endpoints
kubectl apply -f k8s/ai-service-endpoints.yaml
kubectl apply -f k8s/ai-analytics-endpoints.yaml

# Мониторинг
kubectl apply -f k8s/monitoring-config.yaml
```

### **3. Проверка статуса:**
```bash
# Проверить pods
kubectl get pods -n cloudbridge

# Проверить services
kubectl get services -n cloudbridge

# Проверить configmaps
kubectl get configmaps -n cloudbridge

# Проверить network policies
kubectl get networkpolicies -A
```

---

## 🔍 **Отладка**

### **Полезные команды:**
```bash
# Логи deployment
kubectl logs deployment/cloudbridge-relay -n cloudbridge

# Описание pod
kubectl describe pod <pod-name> -n cloudbridge

# Проверка событий
kubectl get events -n cloudbridge --sort-by=.metadata.creationTimestamp

# Проверка ресурсов
kubectl top pods -n cloudbridge
kubectl top nodes

# Проверка сетевых политик
kubectl get networkpolicies -A
kubectl describe networkpolicy allow-prometheus-access -n cloudbridge
```

### **Port forwarding:**
```bash
# P2P API
kubectl port-forward service/cloudbridge-relay-external 8082:8082 -n cloudbridge

# AI Service
kubectl port-forward service/ai-service 8083:8083 -n cloudbridge

# AI Analytics
kubectl port-forward service/ai-analytics-service 8084:8084 -n cloudbridge

# Cadence Web UI
kubectl port-forward service/cadence-web-service 8088:8088 -n cloudbridge
```

---

## 📊 **Мониторинг**

### **Метрики:**
- **P2P API метрики** - количество соединений, пиров, запросов
- **AI Service метрики** - точность предсказаний, использование ресурсов
- **Cadence метрики** - статус workflows, время выполнения
- **Системные метрики** - CPU, память, диск, сеть

### **Алерты:**
- HighLatency - низкая точность AI предсказаний
- AIServiceDown - AI Service недоступен
- HighMemoryUsage - высокое использование памяти
- HighCPUUsage - высокое использование CPU
- AnomaliesDetected - обнаружено много аномалий
- CadenceWorkflowFailed - сбой Cadence workflows

---

## 🔧 **Масштабирование**

### **Горизонтальное масштабирование:**
```bash
# Увеличить количество реплик
kubectl scale deployment/cloudbridge-relay --replicas=3 -n cloudbridge

# Автоматическое масштабирование
kubectl autoscale deployment/cloudbridge-relay --cpu-percent=70 --min=2 --max=10 -n cloudbridge
```

### **Вертикальное масштабирование:**
```yaml
# Обновить ресурсы в deployment.yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

---

## 🔐 **Безопасность**

### **Network Policies:**
- Ограничение трафика между namespace
- Контроль доступа к Prometheus
- Изоляция AI компонентов

### **RBAC:**
```yaml
# Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge

# Role
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps"]
  verbs: ["get", "list", "watch"]

# RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
subjects:
- kind: ServiceAccount
  name: cloudbridge-relay
  namespace: cloudbridge
roleRef:
  kind: Role
  name: cloudbridge-relay
  apiGroup: rbac.authorization.k8s.io
```

---

**🎯 Kubernetes Configuration готова к production развертыванию CloudBridge Relay!**
