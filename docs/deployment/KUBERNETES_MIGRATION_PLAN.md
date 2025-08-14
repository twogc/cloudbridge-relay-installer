# 🚀 План миграции CloudBridge Relay в Kubernetes

## 📊 Анализ текущего состояния

### ✅ Существующая инфраструктура (terraform-relay)
- **Кластер**: `cloudbridge-cluster` в VK Cloud
- **API сервер**: https://91.219.226.200:6443
- **Версия Kubernetes**: v1.31
- **Ресурсы**: 1 master (STD3-4-8) + 2 workers (STD2-2-4)
- **Сеть**: 192.168.1.0/24
- **Аддоны**: Prometheus Stack, Docker Registry, Cert Manager
- **Keycloak**: Работает отдельно (не участвует в миграции)

### ✅ Готовые компоненты
- **Terraform конфигурация**: Полностью развернута
- **Kubernetes манифесты**: Базовые манифесты готовы
- **Мониторинг**: Prometheus + Grafana настроены
- **Скрипты развертывания**: Автоматизированы

### 🔄 Текущий статус CloudBridge
- **Архитектура**: Монолитное приложение
- **Компоненты**: WireGuard, QUIC, Cadence, AI/ML, Quantum
- **Готовность к микросервисам**: 95% (документация готова)

## 🎯 Стратегия миграции

### Phase 1: Подготовка инфраструктуры (1-2 дня)

#### 1.1 Анализ и планирование
```bash
# Анализ текущего кода
cd cloudbridge-relay-installer
find . -name "*.go" -exec grep -l "main()" {} \;
grep -r "package main" ./cmd/
```

#### 1.2 Создание микросервисной архитектуры
```
cloudbridge-relay-installer/
├── cmd/
│   ├── wireguard-service/     # WireGuard P2P Mesh
│   ├── quic-service/          # Enhanced QUIC
│   ├── cadence-service/       # Workflow Orchestration
│   ├── ai-service/            # AI/ML Monitoring
│   ├── quantum-service/       # Post-Quantum Crypto
│   └── metrics-service/       # Metrics & Monitoring
├── internal/
│   ├── wireguard/            # WireGuard модули
│   ├── quic/                 # QUIC модули
│   ├── cadence/              # Cadence модули
│   ├── ai/                   # AI модули
│   ├── quantum/              # Quantum модули
│   └── shared/               # Общие компоненты
└── k8s/
    ├── wireguard-service/
    ├── quic-service/
    ├── cadence-service/
    ├── ai-service/
    ├── quantum-service/
    └── metrics-service/
```

### Phase 2: Извлечение микросервисов (3-5 дней)

#### 2.1 WireGuard Service
```go
// cmd/wireguard-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/wireguard"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Инициализация WireGuard
    wgManager := wireguard.NewWireGuardManager(logger)
    
    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "wireguard"})
    })
    
    // WireGuard API
    r.POST("/tunnels", wgManager.CreateTunnel)
    r.GET("/tunnels", wgManager.ListTunnels)
    r.GET("/tunnels/:id", wgManager.GetTunnel)
    r.DELETE("/tunnels/:id", wgManager.DeleteTunnel)
    r.POST("/tunnels/:id/start", wgManager.StartTunnel)
    r.POST("/tunnels/:id/stop", wgManager.StopTunnel)
    
    // P2P Mesh API
    r.GET("/peers", wgManager.ListPeers)
    r.POST("/peers/discover", wgManager.DiscoverPeers)
    r.GET("/topology", wgManager.GetTopology)
    
    // Metrics
    r.GET("/metrics", wgManager.MetricsHandler)
    
    port := os.Getenv("WIREGUARD_PORT")
    if port == "" {
        port = "8080"
    }
    
    logger.Info("WireGuard Service starting", zap.String("port", port))
    r.Run(":" + port)
}
```

#### 2.2 QUIC Service
```go
// cmd/quic-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/quic"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Инициализация QUIC
    quicManager := quic.NewQUICManager(logger)
    
    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "quic"})
    })
    
    // QUIC API
    r.POST("/connections", quicManager.CreateConnection)
    r.GET("/connections", quicManager.ListConnections)
    r.GET("/connections/:id", quicManager.GetConnection)
    r.DELETE("/connections/:id", quicManager.CloseConnection)
    
    // Performance API
    r.GET("/performance", quicManager.GetPerformance)
    r.POST("/performance/optimize", quicManager.OptimizePerformance)
    
    // Metrics
    r.GET("/metrics", quicManager.MetricsHandler)
    
    port := os.Getenv("QUIC_PORT")
    if port == "" {
        port = "8081"
    }
    
    logger.Info("QUIC Service starting", zap.String("port", port))
    r.Run(":" + port)
}
```

#### 2.3 Cadence Service
```go
// cmd/cadence-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/twogc/cloudbridge-relay-installer/internal/cadence"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Инициализация Cadence
    cadenceManager := cadence.NewCadenceManager(logger)
    
    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "cadence"})
    })
    
    // Workflow API
    r.POST("/workflows", cadenceManager.StartWorkflow)
    r.GET("/workflows", cadenceManager.ListWorkflows)
    r.GET("/workflows/:id", cadenceManager.GetWorkflow)
    r.POST("/workflows/:id/cancel", cadenceManager.CancelWorkflow)
    
    // Activity API
    r.GET("/activities", cadenceManager.ListActivities)
    r.GET("/activities/:id", cadenceManager.GetActivity)
    
    // Metrics
    r.GET("/metrics", cadenceManager.MetricsHandler)
    
    port := os.Getenv("CADENCE_PORT")
    if port == "" {
        port = "8082"
    }
    
    logger.Info("Cadence Service starting", zap.String("port", port))
    r.Run(":" + port)
}
```

### Phase 3: Kubernetes развертывание (2-3 дня)

#### 3.1 WireGuard Service Deployment
```yaml
# k8s/wireguard-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wireguard-service
  namespace: cloudbridge-relay
spec:
  replicas: 2
  selector:
    matchLabels:
      app: wireguard-service
  template:
    metadata:
      labels:
        app: wireguard-service
    spec:
      containers:
      - name: wireguard-service
        image: registry.vkcs.cloud/cloudbridge/wireguard-service:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 51820
          name: wireguard
          protocol: UDP
        env:
        - name: WIREGUARD_PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
            - SYS_MODULE
        volumeMounts:
        - name: wireguard-configs
          mountPath: /etc/wireguard
        - name: wireguard-keys
          mountPath: /etc/wireguard/keys
        resources:
          limits:
            cpu: "1000m"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: wireguard-configs
        persistentVolumeClaim:
          claimName: wireguard-configs-pvc
      - name: wireguard-keys
        secret:
          secretName: wireguard-keys
---
apiVersion: v1
kind: Service
metadata:
  name: wireguard-service
  namespace: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 51820
    targetPort: 51820
    name: wireguard
    protocol: UDP
  selector:
    app: wireguard-service
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: wireguard-service-hpa
  namespace: cloudbridge-relay
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wireguard-service
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 80
```

#### 3.2 QUIC Service Deployment
```yaml
# k8s/quic-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quic-service
  namespace: cloudbridge-relay
spec:
  replicas: 3
  selector:
    matchLabels:
      app: quic-service
  template:
    metadata:
      labels:
        app: quic-service
    spec:
      containers:
      - name: quic-service
        image: registry.vkcs.cloud/cloudbridge/quic-service:v1.0.0
        ports:
        - containerPort: 8081
          name: http
        - containerPort: 9090
          name: quic
          protocol: UDP
        env:
        - name: QUIC_PORT
          value: "8081"
        - name: LOG_LEVEL
          value: "info"
        resources:
          limits:
            cpu: "2000m"
            memory: "2Gi"
          requests:
            cpu: "1000m"
            memory: "1Gi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: quic-service
  namespace: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8081
    targetPort: 8081
    name: http
  - port: 9090
    targetPort: 9090
    name: quic
    protocol: UDP
  selector:
    app: quic-service
```

#### 3.3 Cadence Service Deployment
```yaml
# k8s/cadence-service/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cadence-service
  namespace: cloudbridge-relay
spec:
  replicas: 2
  selector:
    matchLabels:
      app: cadence-service
  template:
    metadata:
      labels:
        app: cadence-service
    spec:
      containers:
      - name: cadence-service
        image: registry.vkcs.cloud/cloudbridge/cadence-service:v1.0.0
        ports:
        - containerPort: 8082
          name: http
        env:
        - name: CADENCE_PORT
          value: "8082"
        - name: LOG_LEVEL
          value: "info"
        - name: CADENCE_DOMAIN
          value: "cloudbridge"
        - name: CADENCE_HOSTPORT
          value: "cadence-server:7933"
        resources:
          limits:
            cpu: "1000m"
            memory: "1Gi"
          requests:
            cpu: "500m"
            memory: "512Mi"
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
---
apiVersion: v1
kind: Service
metadata:
  name: cadence-service
  namespace: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8082
    targetPort: 8082
    name: http
  selector:
    app: cadence-service
```

### Phase 4: Интеграция и тестирование (2-3 дня)

#### 4.1 API Gateway
```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudbridge-ingress
  namespace: cloudbridge-relay
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - api.cloudbridge.local
    secretName: cloudbridge-tls
  rules:
  - host: api.cloudbridge.local
    http:
      paths:
      - path: /wireguard
        pathType: Prefix
        backend:
          service:
            name: wireguard-service
            port:
              number: 8080
      - path: /quic
        pathType: Prefix
        backend:
          service:
            name: quic-service
            port:
              number: 8081
      - path: /cadence
        pathType: Prefix
        backend:
          service:
            name: cadence-service
            port:
              number: 8082
      - path: /ai
        pathType: Prefix
        backend:
          service:
            name: ai-service
            port:
              number: 8083
      - path: /quantum
        pathType: Prefix
        backend:
          service:
            name: quantum-service
            port:
              number: 8084
      - path: /metrics
        pathType: Prefix
        backend:
          service:
            name: metrics-service
            port:
              number: 8085
```

#### 4.2 Service Mesh (Istio)
```yaml
# k8s/istio/virtual-service.yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: cloudbridge-vs
  namespace: cloudbridge-relay
spec:
  hosts:
  - "api.cloudbridge.local"
  gateways:
  - cloudbridge-gateway
  http:
  - match:
    - uri:
        prefix: /wireguard
    route:
    - destination:
        host: wireguard-service
        port:
          number: 8080
  - match:
    - uri:
        prefix: /quic
    route:
    - destination:
        host: quic-service
        port:
          number: 8081
  - match:
    - uri:
        prefix: /cadence
    route:
    - destination:
        host: cadence-service
        port:
          number: 8082
```

### Phase 5: Мониторинг и алерты (1-2 дня)

#### 5.1 ServiceMonitor для каждого сервиса
```yaml
# k8s/monitoring/servicemonitors.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: wireguard-service
  namespace: cloudbridge-relay
spec:
  selector:
    matchLabels:
      app: wireguard-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: quic-service
  namespace: cloudbridge-relay
spec:
  selector:
    matchLabels:
      app: quic-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
---
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cadence-service
  namespace: cloudbridge-relay
spec:
  selector:
    matchLabels:
      app: cadence-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

#### 5.2 PrometheusRules для алертов
```yaml
# k8s/monitoring/prometheusrules.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cloudbridge-alerts
  namespace: cloudbridge-relay
spec:
  groups:
  - name: cloudbridge.rules
    rules:
    - alert: WireGuardServiceDown
      expr: up{job="wireguard-service"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "WireGuard service is down"
        description: "WireGuard service has been down for more than 1 minute"
    
    - alert: QUICServiceDown
      expr: up{job="quic-service"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "QUIC service is down"
        description: "QUIC service has been down for more than 1 minute"
    
    - alert: CadenceServiceDown
      expr: up{job="cadence-service"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "Cadence service is down"
        description: "Cadence service has been down for more than 1 minute"
    
    - alert: HighCPUUsage
      expr: rate(container_cpu_usage_seconds_total{container=~"wireguard-service|quic-service|cadence-service"}[5m]) > 0.8
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "High CPU usage detected"
        description: "CPU usage is above 80% for 5 minutes"
```

## 🚀 Пошаговый план выполнения

### День 1-2: Подготовка
1. **Анализ кода**: Извлечение компонентов из монолита
2. **Создание Dockerfile**: Для каждого микросервиса
3. **Настройка CI/CD**: Автоматическая сборка образов

### День 3-5: Разработка микросервисов
1. **WireGuard Service**: Извлечение и тестирование
2. **QUIC Service**: Извлечение и тестирование
3. **Cadence Service**: Извлечение и тестирование
4. **AI Service**: Извлечение и тестирование
5. **Quantum Service**: Извлечение и тестирование
6. **Metrics Service**: Извлечение и тестирование

### День 6-8: Kubernetes развертывание
1. **Создание манифестов**: Для каждого сервиса
2. **Настройка Ingress**: API Gateway
3. **Service Mesh**: Istio конфигурация
4. **Тестирование**: End-to-end тесты

### День 9-10: Мониторинг и оптимизация
1. **Prometheus метрики**: Настройка мониторинга
2. **Grafana дашборды**: Визуализация метрик
3. **Алерты**: Настройка уведомлений
4. **Performance tuning**: Оптимизация ресурсов

## 📋 Чек-лист миграции

### ✅ Подготовка
- [ ] Анализ существующего кода завершен
- [ ] Микросервисная архитектура спроектирована
- [ ] Docker образы созданы и протестированы
- [ ] Kubernetes манифесты подготовлены

### ✅ Развертывание
- [ ] Namespaces созданы
- [ ] ConfigMaps и Secrets настроены
- [ ] Deployments развернуты
- [ ] Services настроены
- [ ] Ingress настроен

### ✅ Мониторинг
- [ ] ServiceMonitors созданы
- [ ] PrometheusRules настроены
- [ ] Grafana дашборды импортированы
- [ ] Алерты настроены

### ✅ Тестирование
- [ ] Health checks работают
- [ ] End-to-end тесты пройдены
- [ ] Performance тесты выполнены
- [ ] Load testing завершен

## 🔧 Полезные команды

### Развертывание
```bash
# Применение всех манифестов
kubectl apply -f k8s/ -R

# Проверка статуса
kubectl get all -n cloudbridge-relay

# Логи сервисов
kubectl logs -f deployment/wireguard-service -n cloudbridge-relay
kubectl logs -f deployment/quic-service -n cloudbridge-relay
kubectl logs -f deployment/cadence-service -n cloudbridge-relay
```

### Мониторинг
```bash
# Доступ к Grafana
kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n prometheus-monitoring

# Доступ к Prometheus
kubectl port-forward svc/kube-prometheus-stack-prometheus 9090:9090 -n prometheus-monitoring

# Проверка метрик
curl http://localhost:9090/api/v1/query?query=up
```

### Масштабирование
```bash
# Масштабирование сервисов
kubectl scale deployment wireguard-service --replicas=3 -n cloudbridge-relay
kubectl scale deployment quic-service --replicas=5 -n cloudbridge-relay
kubectl scale deployment cadence-service --replicas=2 -n cloudbridge-relay
```

## 🎯 Ожидаемые результаты

### Производительность
- **Лучшая масштабируемость**: Независимое масштабирование сервисов
- **Высокая доступность**: Отказоустойчивость на уровне сервисов
- **Оптимизация ресурсов**: Точное выделение ресурсов

### Операционная эффективность
- **Упрощенное развертывание**: CI/CD пайплайны
- **Улучшенный мониторинг**: Детальная метрика каждого сервиса
- **Быстрое восстановление**: Автоматическое перезапуск подов

### Бизнес-ценность
- **Гибкость**: Независимое развитие сервисов
- **Надежность**: Изоляция отказов
- **Масштабируемость**: Поддержка роста нагрузки

---

**⏱️ Общее время миграции**: 10-12 дней  
**🎯 Статус**: ГОТОВ К НАЧАЛУ МИГРАЦИИ
