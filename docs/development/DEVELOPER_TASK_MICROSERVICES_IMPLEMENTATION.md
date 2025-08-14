# 🚀 Техническое задание: Реализация микросервисов CloudBridge Relay

## 📋 Общая информация

**Проект**: CloudBridge Relay  
**Тип**: Миграция с монолитной архитектуры на микросервисы  
**Статус**: Инфраструктура готова, требуется реализация сервисов  
**Приоритет**: Высокий  

## 🎯 Цель

Реализовать полнофункциональные микросервисы для замены монолитного приложения CloudBridge Relay с возможностью внешнего доступа и production-ready функциональностью.

## 📊 Текущий статус

### ✅ Готово:
- **Kubernetes кластер** (VK Cloud, 6 vCPU / 12 GB RAM)
- **Инфраструктура**: Namespaces, ConfigMaps, Secrets, Ingress
- **Placeholder сервисы**: Auth Service, Metrics Service (базовая структура)
- **NodePort сервисы**: Настроены для внешнего доступа
- **Автоматическое масштабирование**: HPA настроены

### 🔧 Проблемы для решения:
- **🚨 КРИТИЧНО**: Docker Hub rate limit - поды в состоянии ImagePullBackOff
- **🚨 КРИТИЧНО**: VK Cloud превышена квота на floating IP (лимит: 4, использовано: 4)
- **🚨 КРИТИЧНО**: LoadBalancer не может получить внешний IP
- **Placeholder сервисы** не предоставляют реальную функциональность
- **Отсутствуют health checks** и метрики
- **VK Cloud quota exceeded** для floating IP (лимит: 4, использовано: 4)
- **Внешний доступ** настроен через NodePort, но поды не работают

### 🚨 КРИТИЧЕСКАЯ ПРОБЛЕМА:
```
Pods Status:
- auth-service: ImagePullBackOff (50 restarts)
- metrics-service: ErrImagePull (50 restarts)

External Access:
- NodePort services configured: 30080, 30084, 30090, 30091
- Connection refused due to pods not running
```

## 🚨 КРИТИЧЕСКИЕ ПРОБЛЕМЫ И РЕШЕНИЯ

### Проблема 1: Docker Hub Rate Limit и блокировка России
**Симптомы**: Поды в состоянии ImagePullBackOff, ошибки подключения к Docker Hub
**Решение**: 
1. ✅ Создать локальные tar архивы образов
2. ✅ Использовать `imagePullPolicy: Never` в Kubernetes
3. ✅ Загружать образы напрямую в кластер
4. ✅ Альтернативно: использовать VK Cloud Container Registry

### Проблема 2: VK Cloud Floating IP Quota
**Симптомы**: LoadBalancer не может получить внешний IP
**Решение**:
1. Назначить внешний IP master узлу: `185.22.153.XX`
2. Использовать NodePort для внешнего доступа
3. Настроить Ingress с внешним IP

### Проблема 3: Внешний доступ
**Симптомы**: Сервисы недоступны извне
**Решение**:
1. Настроить DNS записи на внешний IP
2. Настроить SSL сертификаты
3. Протестировать доступность

## ✅ ГОТОВЫЕ РЕШЕНИЯ

### 1. Локальные образы (решение проблемы Docker Hub)
```bash
# Сохранение образов в tar файлы
./scripts/save-and-load-images.sh save

# Создание Kubernetes манифестов
./scripts/save-and-load-images.sh manifests

# Развертывание в Kubernetes
./scripts/deploy-to-k8s.sh deploy
```

### 2. Container Registry с VK Cloud Storage
```bash
# Настройка Container Registry
./scripts/setup-container-registry.sh

# Конфигурация в configs/registry-values.yaml
# - VK Cloud Storage настроен
# - htpasswd сгенерирован
# - Плавающий IP: 89.208.86.150
```

### 3. Готовые образы
- ✅ `images/auth-service-v1.0.0.tar` - Auth Service
- ✅ `images/metrics-service-v1.0.0.tar` - Metrics Service  
- ✅ `images/registry-2.tar` - Container Registry

### 4. Kubernetes манифесты
- ✅ `k8s/cloudbridge-namespace.yaml` - Namespace
- ✅ `k8s/auth-service-deployment.yaml` - Auth Service
- ✅ `k8s/metrics-service-deployment.yaml` - Metrics Service
- ✅ `k8s/cloudbridge-ingress.yaml` - Ingress
- ✅ `k8s/cloudbridge-loadbalancer.yaml` - LoadBalancer

## 🛠️ Задачи разработчика

### 1. Реализация Auth Service

#### 1.1 Создание Go приложения
```go
// cmd/auth-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
    "os"
)

func main() {
    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    r.GET("/ready", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ready"})
    })
    
    // Auth endpoints
    r.POST("/login", handleLogin)
    r.POST("/logout", handleLogout)
    r.GET("/profile", handleProfile)
    r.POST("/refresh", handleRefresh)
    r.GET("/status", handleStatus)
    
    // Metrics
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Auth Service starting on port %s", port)
    r.Run(":" + port)
}
```

#### 1.2 Dockerfile
```dockerfile
# Dockerfile.auth
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service ./cmd/auth-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/auth-service .
COPY --from=builder /app/configs/auth-service.yaml ./configs/

EXPOSE 8080 9090

CMD ["./auth-service"]
```

### 2. Реализация Metrics Service

#### 2.1 Создание Go приложения
```go
// cmd/metrics-service/main.go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
    "os"
)

var (
    connectionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "cloudbridge_connections_total",
        Help: "Total number of connections",
    })
    
    tunnelsActive = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "cloudbridge_tunnels_active",
        Help: "Number of active tunnels",
    })
)

func main() {
    prometheus.MustRegister(connectionsTotal, tunnelsActive)
    
    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy"})
    })
    
    r.GET("/ready", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ready"})
    })
    
    // Metrics endpoints
    r.POST("/metrics/connections/increment", handleIncrementConnections)
    r.POST("/metrics/connections/decrement", handleDecrementConnections)
    r.POST("/metrics/tunnels/increment", handleIncrementTunnels)
    r.POST("/metrics/tunnels/decrement", handleDecrementTunnels)
    
    // Prometheus metrics
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    
    port := os.Getenv("METRICS_PORT")
    if port == "" {
        port = "8084"
    }
    
    log.Printf("Metrics Service starting on port %s", port)
    r.Run(":" + port)
}
```

#### 2.2 Dockerfile
```dockerfile
# Dockerfile.metrics
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o metrics-service ./cmd/metrics-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/metrics-service .
COPY --from=builder /app/configs/metrics-service.yaml ./configs/

EXPOSE 8084 9090

CMD ["./metrics-service"]
```

### 3. Настройка Container Registry

#### 3.1 VK Cloud Container Registry (ОБЯЗАТЕЛЬНО)
```bash
# Создание registry
vk cloud container registry create --name cloudbridge-registry

# Аутентификация
vk cloud container registry get-credentials --name cloudbridge-registry

# Тегирование и отправка образов
docker tag cloudbridge/auth-service:v1.0.0 registry.vkcs.cloud/cloudbridge/auth-service:v1.0.0
docker tag cloudbridge/metrics-service:v1.0.0 registry.vkcs.cloud/cloudbridge/metrics-service:v1.0.0

docker push registry.vkcs.cloud/cloudbridge/auth-service:v1.0.0
docker push registry.vkcs.cloud/cloudbridge/metrics-service:v1.0.0
```

#### 3.2 Обновление Kubernetes манифестов
```yaml
# k8s/auth-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: registry.vkcs.cloud/cloudbridge/auth-service:v1.0.0  # ИСПОЛЬЗОВАТЬ VK REGISTRY
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: cloudbridge-secrets
              key: jwt-secret
        - name: LOG_LEVEL
          value: "info"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

### 4. Настройка внешнего доступа

#### 4.1 Текущая конфигурация NodePort
```bash
# Auth Service
auth-service-external: NodePort 10.254.241.6
- HTTP: 8080:30080/TCP
- Metrics: 9090:30090/TCP

# Metrics Service  
metrics-service-external: NodePort 10.254.204.64
- HTTP: 8084:30084/TCP
- Metrics: 9090:30091/TCP

# Kubernetes узлы
kubernetes-cluster-8926-master-0: 10.0.0.137
kubernetes-cluster-8926-default-group-0: 10.0.0.130
```

#### 4.2 Тестирование доступа
```bash
# После развертывания реальных сервисов
curl http://10.0.0.137:30080/health
curl http://10.0.0.137:30084/health
curl http://10.0.0.137:30090/metrics
curl http://10.0.0.137:30091/metrics
```

#### 4.3 DNS настройка (опционально)
```bash
# Добавить в /etc/hosts для тестирования
10.0.0.137 auth.cloudbridge.local
10.0.0.137 metrics.cloudbridge.local
10.0.0.137 api.cloudbridge.local
```

### 5. Мониторинг и логирование

#### 5.1 Prometheus + Grafana
```yaml
# k8s/monitoring.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
  namespace: cloudbridge-monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: prometheus-config
          mountPath: /etc/prometheus
      volumes:
      - name: prometheus-config
        configMap:
          name: prometheus-config
```

#### 5.2 ELK Stack
```yaml
# k8s/logging.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: elasticsearch
  namespace: cloudbridge-monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: elasticsearch
  template:
    metadata:
      labels:
        app: elasticsearch
    spec:
      containers:
      - name: elasticsearch
        image: docker.elastic.co/elasticsearch/elasticsearch:8.11.0
        ports:
        - containerPort: 9200
        env:
        - name: discovery.type
          value: single-node
```

## 📅 План выполнения

### Этап 1: Критическое исправление (1 день)
- [ ] **ОБЯЗАТЕЛЬНО**: Настроить VK Cloud Container Registry
- [ ] Создать реальные Go приложения с health checks
- [ ] Собрать и загрузить образы в VK Registry
- [ ] Обновить Kubernetes манифесты
- [ ] Развернуть работающие сервисы

### Этап 2: Внешний доступ (1 день)
- [ ] Протестировать доступ через NodePort
- [ ] Настроить DNS записи для удобства
- [ ] Протестировать Ingress маршрутизацию
- [ ] Настроить базовую безопасность

### Этап 3: Мониторинг (1-2 дня)
- [ ] Развернуть Prometheus + Grafana
- [ ] Настроить алерты
- [ ] Развернуть ELK Stack
- [ ] Настроить логирование микросервисов

### Этап 4: Production готовность (1 день)
- [ ] Настроить backup стратегии
- [ ] Настроить security policies
- [ ] Произвести нагрузочное тестирование
- [ ] Документировать deployment процесс

## 🔧 Технические требования

### Язык программирования
- **Go 1.23+** - основной язык разработки
- **Gin framework** - для HTTP API
- **Prometheus client** - для метрик

### Контейнеризация
- **Docker** - для создания образов
- **VK Cloud Container Registry** - ОБЯЗАТЕЛЬНО для хранения образов
- **Alpine Linux** - базовый образ

### Kubernetes
- **Namespaces**: cloudbridge, cloudbridge-monitoring
- **Deployments** с HPA
- **Services** типа ClusterIP и NodePort
- **Ingress** с SSL
- **ConfigMaps** и **Secrets**

### Мониторинг
- **Prometheus** - сбор метрик
- **Grafana** - визуализация
- **AlertManager** - алерты
- **ELK Stack** - логирование

## 📋 Критерии приемки

### Функциональные требования
- [ ] Auth Service отвечает на все endpoints (/health, /ready, /login, /logout, /profile, /refresh, /status)
- [ ] Metrics Service собирает и экспортирует метрики
- [ ] Внешний доступ работает через NodePort (порты 30080, 30084, 30090, 30091)
- [ ] Поды запускаются без ошибок ImagePullBackOff
- [ ] Health checks работают корректно
- [ ] Мониторинг показывает метрики сервисов

### Нефункциональные требования
- [ ] Время отклика API < 100ms
- [ ] Доступность сервисов > 99.9%
- [ ] Автоматическое масштабирование работает
- [ ] Backup и восстановление настроены
- [ ] Security policies применены

## 🚨 КРИТИЧЕСКИЕ ЗАДАЧИ

### 1. Исправить ImagePullBackOff
```bash
# Проблема: Docker Hub rate limit
# Решение: Использовать VK Cloud Container Registry
vk cloud container registry create --name cloudbridge-registry
docker tag cloudbridge/auth-service:v1.0.0 registry.vkcs.cloud/cloudbridge/auth-service:v1.0.0
docker push registry.vkcs.cloud/cloudbridge/auth-service:v1.0.0
```

### 2. Обеспечить внешний доступ
```bash
# Текущие порты NodePort:
# Auth Service: 30080 (HTTP), 30090 (Metrics)
# Metrics Service: 30084 (HTTP), 30091 (Metrics)
# Узлы: 10.0.0.137 (master), 10.0.0.130 (worker)
```

### 3. Реализовать health checks
```go
// Обязательные endpoints:
r.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
})
r.GET("/ready", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ready"})
})
```

## 🚀 Следующие этапы миграции

### Phase 3: WireGuard Service
- Извлечение WireGuard функциональности
- Создание отдельного сервиса
- Интеграция с Auth Service

### Phase 4: QUIC Service
- Реализация QUIC протокола
- Интеграция с Metrics Service
- Настройка P2P соединений

### Phase 5: Cadence Workflows
- Настройка Cadence для оркестрации
- Интеграция всех сервисов
- Реализация сложных workflow'ов

## 📞 Контакты и поддержка

**Ответственный**: Разработчик CloudBridge Relay  
**Статус**: КРИТИЧЕСКИЙ - требуется немедленная реализация  
**Приоритет**: Критический  

---

**🎯 Цель**: Полностью функциональные микросервисы с внешним доступом и мониторингом

**🚨 КРИТИЧНО**: Исправить ImagePullBackOff через VK Cloud Container Registry
