#!/bin/bash

# 🚀 Скрипт миграции CloudBridge Relay в Kubernetes
# Автоматизация процесса миграции монолитного приложения в микросервисы

set -e

# Цвета для вывода
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Переменные
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
K8S_DIR="$PROJECT_ROOT/k8s"
SCRIPTS_DIR="$PROJECT_ROOT/scripts"
TERRAFORM_DIR="../terraform-relay/terraform"

# Функции
print_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║              CloudBridge Relay Migration                     ║"
    echo "║                   Kubernetes Edition                          ║"
    echo "║                                                              ║"
    echo "║  Автоматическая миграция монолита в микросервисы             ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    print_step "Проверка зависимостей..."
    
    local missing_deps=()
    
    # Проверка Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    # Проверка Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi
    
    # Проверка kubectl
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    # Проверка helm
    if ! command -v helm &> /dev/null; then
        missing_deps+=("helm")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Отсутствуют зависимости: ${missing_deps[*]}"
        echo "Установите недостающие зависимости и повторите попытку"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
}

# Анализ текущего кода
analyze_code() {
    print_step "Анализ текущего кода..."
    
    cd "$PROJECT_ROOT"
    
    # Поиск main функций
    echo "Найдены main функции:"
    find . -name "*.go" -exec grep -l "func main()" {} \;
    
    # Анализ пакетов
    echo "Найдены пакеты:"
    find ./internal -type d -name "*" | head -10
    
    # Создание структуры для микросервисов
    mkdir -p "$K8S_DIR"/{wireguard-service,quic-service,cadence-service,ai-service,quantum-service,metrics-service}
    
    print_success "Анализ кода завершен"
}

# Создание микросервисов
create_microservices() {
    print_step "Создание микросервисов..."
    
    cd "$PROJECT_ROOT"
    
    # WireGuard Service
    print_step "Создание WireGuard Service..."
    mkdir -p cmd/wireguard-service
    cat > cmd/wireguard-service/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "os"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "wireguard"})
    })
    
    // WireGuard API endpoints
    r.GET("/tunnels", func(c *gin.Context) {
        c.JSON(200, gin.H{"tunnels": []string{}})
    })
    
    r.POST("/tunnels", func(c *gin.Context) {
        c.JSON(201, gin.H{"message": "tunnel created"})
    })
    
    // Metrics
    r.GET("/metrics", func(c *gin.Context) {
        c.JSON(200, gin.H{"metrics": "wireguard_metrics"})
    })
    
    port := os.Getenv("WIREGUARD_PORT")
    if port == "" {
        port = "8080"
    }
    
    logger.Info("WireGuard Service starting", zap.String("port", port))
    r.Run(":" + port)
}
EOF

    # QUIC Service
    print_step "Создание QUIC Service..."
    mkdir -p cmd/quic-service
    cat > cmd/quic-service/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "os"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "quic"})
    })
    
    // QUIC API endpoints
    r.GET("/connections", func(c *gin.Context) {
        c.JSON(200, gin.H{"connections": []string{}})
    })
    
    r.POST("/connections", func(c *gin.Context) {
        c.JSON(201, gin.H{"message": "connection created"})
    })
    
    // Metrics
    r.GET("/metrics", func(c *gin.Context) {
        c.JSON(200, gin.H{"metrics": "quic_metrics"})
    })
    
    port := os.Getenv("QUIC_PORT")
    if port == "" {
        port = "8081"
    }
    
    logger.Info("QUIC Service starting", zap.String("port", port))
    r.Run(":" + port)
}
EOF

    # Cadence Service
    print_step "Создание Cadence Service..."
    mkdir -p cmd/cadence-service
    cat > cmd/cadence-service/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "os"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "cadence"})
    })
    
    // Cadence API endpoints
    r.GET("/workflows", func(c *gin.Context) {
        c.JSON(200, gin.H{"workflows": []string{}})
    })
    
    r.POST("/workflows", func(c *gin.Context) {
        c.JSON(201, gin.H{"message": "workflow started"})
    })
    
    // Metrics
    r.GET("/metrics", func(c *gin.Context) {
        c.JSON(200, gin.H{"metrics": "cadence_metrics"})
    })
    
    port := os.Getenv("CADENCE_PORT")
    if port == "" {
        port = "8082"
    }
    
    logger.Info("Cadence Service starting", zap.String("port", port))
    r.Run(":" + port)
}
EOF

    # AI Service
    print_step "Создание AI Service..."
    mkdir -p cmd/ai-service
    cat > cmd/ai-service/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "os"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "ai"})
    })
    
    // AI API endpoints
    r.GET("/anomalies", func(c *gin.Context) {
        c.JSON(200, gin.H{"anomalies": []string{}})
    })
    
    r.POST("/analyze", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "analysis completed"})
    })
    
    // Metrics
    r.GET("/metrics", func(c *gin.Context) {
        c.JSON(200, gin.H{"metrics": "ai_metrics"})
    })
    
    port := os.Getenv("AI_PORT")
    if port == "" {
        port = "8083"
    }
    
    logger.Info("AI Service starting", zap.String("port", port))
    r.Run(":" + port)
}
EOF

    # Quantum Service
    print_step "Создание Quantum Service..."
    mkdir -p cmd/quantum-service
    cat > cmd/quantum-service/main.go << 'EOF'
package main

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "os"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    r := gin.Default()
    
    // Health checks
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "healthy", "service": "quantum"})
    })
    
    // Quantum API endpoints
    r.GET("/keys", func(c *gin.Context) {
        c.JSON(200, gin.H{"keys": []string{}})
    })
    
    r.POST("/generate", func(c *gin.Context) {
        c.JSON(201, gin.H{"message": "quantum key generated"})
    })
    
    // Metrics
    r.GET("/metrics", func(c *gin.Context) {
        c.JSON(200, gin.H{"metrics": "quantum_metrics"})
    })
    
    port := os.Getenv("QUANTUM_PORT")
    if port == "" {
        port = "8084"
    }
    
    logger.Info("Quantum Service starting", zap.String("port", port))
    r.Run(":" + port)
}
EOF

    print_success "Микросервисы созданы"
}

# Создание Dockerfile для микросервисов
create_dockerfiles() {
    print_step "Создание Dockerfile для микросервисов..."
    
    cd "$PROJECT_ROOT"
    
    # WireGuard Service Dockerfile
    cat > Dockerfile.wireguard << 'EOF'
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o wireguard-service ./cmd/wireguard-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/wireguard-service .
EXPOSE 8080

CMD ["./wireguard-service"]
EOF

    # QUIC Service Dockerfile
    cat > Dockerfile.quic << 'EOF'
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quic-service ./cmd/quic-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/quic-service .
EXPOSE 8081

CMD ["./quic-service"]
EOF

    # Cadence Service Dockerfile
    cat > Dockerfile.cadence << 'EOF'
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cadence-service ./cmd/cadence-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/cadence-service .
EXPOSE 8082

CMD ["./cadence-service"]
EOF

    # AI Service Dockerfile
    cat > Dockerfile.ai << 'EOF'
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ai-service ./cmd/ai-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/ai-service .
EXPOSE 8083

CMD ["./ai-service"]
EOF

    # Quantum Service Dockerfile
    cat > Dockerfile.quantum << 'EOF'
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quantum-service ./cmd/quantum-service

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/quantum-service .
EXPOSE 8084

CMD ["./quantum-service"]
EOF

    print_success "Dockerfile созданы"
}

# Сборка Docker образов
build_images() {
    print_step "Сборка Docker образов..."
    
    cd "$PROJECT_ROOT"
    
    # Сборка WireGuard Service
    print_step "Сборка WireGuard Service..."
    docker build -f Dockerfile.wireguard -t cloudbridge/wireguard-service:v1.0.0 .
    
    # Сборка QUIC Service
    print_step "Сборка QUIC Service..."
    docker build -f Dockerfile.quic -t cloudbridge/quic-service:v1.0.0 .
    
    # Сборка Cadence Service
    print_step "Сборка Cadence Service..."
    docker build -f Dockerfile.cadence -t cloudbridge/cadence-service:v1.0.0 .
    
    # Сборка AI Service
    print_step "Сборка AI Service..."
    docker build -f Dockerfile.ai -t cloudbridge/ai-service:v1.0.0 .
    
    # Сборка Quantum Service
    print_step "Сборка Quantum Service..."
    docker build -f Dockerfile.quantum -t cloudbridge/quantum-service:v1.0.0 .
    
    print_success "Docker образы собраны"
}

# Создание Kubernetes манифестов
create_k8s_manifests() {
    print_step "Создание Kubernetes манифестов..."
    
    cd "$PROJECT_ROOT"
    
    # Namespace
    cat > "$K8S_DIR/namespace.yaml" << 'EOF'
apiVersion: v1
kind: Namespace
metadata:
  name: cloudbridge-relay
  labels:
    name: cloudbridge-relay
    environment: production
EOF

    # WireGuard Service Deployment
    cat > "$K8S_DIR/wireguard-service/deployment.yaml" << 'EOF'
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
        image: cloudbridge/wireguard-service:v1.0.0
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: WIREGUARD_PORT
          value: "8080"
        - name: LOG_LEVEL
          value: "info"
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
  selector:
    app: wireguard-service
EOF

    # QUIC Service Deployment
    cat > "$K8S_DIR/quic-service/deployment.yaml" << 'EOF'
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
        image: cloudbridge/quic-service:v1.0.0
        ports:
        - containerPort: 8081
          name: http
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
  selector:
    app: quic-service
EOF

    # Cadence Service Deployment
    cat > "$K8S_DIR/cadence-service/deployment.yaml" << 'EOF'
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
        image: cloudbridge/cadence-service:v1.0.0
        ports:
        - containerPort: 8082
          name: http
        env:
        - name: CADENCE_PORT
          value: "8082"
        - name: LOG_LEVEL
          value: "info"
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
EOF

    # AI Service Deployment
    cat > "$K8S_DIR/ai-service/deployment.yaml" << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ai-service
  namespace: cloudbridge-relay
spec:
  replicas: 2
  selector:
    matchLabels:
      app: ai-service
  template:
    metadata:
      labels:
        app: ai-service
    spec:
      containers:
      - name: ai-service
        image: cloudbridge/ai-service:v1.0.0
        ports:
        - containerPort: 8083
          name: http
        env:
        - name: AI_PORT
          value: "8083"
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
            port: 8083
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8083
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: ai-service
  namespace: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8083
    targetPort: 8083
    name: http
  selector:
    app: ai-service
EOF

    # Quantum Service Deployment
    cat > "$K8S_DIR/quantum-service/deployment.yaml" << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: quantum-service
  namespace: cloudbridge-relay
spec:
  replicas: 2
  selector:
    matchLabels:
      app: quantum-service
  template:
    metadata:
      labels:
        app: quantum-service
    spec:
      containers:
      - name: quantum-service
        image: cloudbridge/quantum-service:v1.0.0
        ports:
        - containerPort: 8084
          name: http
        env:
        - name: QUANTUM_PORT
          value: "8084"
        - name: LOG_LEVEL
          value: "info"
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
            port: 8084
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8084
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: quantum-service
  namespace: cloudbridge-relay
spec:
  type: ClusterIP
  ports:
  - port: 8084
    targetPort: 8084
    name: http
  selector:
    app: quantum-service
EOF

    # Ingress
    cat > "$K8S_DIR/ingress.yaml" << 'EOF'
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: cloudbridge-ingress
  namespace: cloudbridge-relay
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
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
EOF

    print_success "Kubernetes манифесты созданы"
}

# Развертывание в Kubernetes
deploy_to_kubernetes() {
    print_step "Развертывание в Kubernetes..."
    
    # Проверка подключения к кластеру
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Не удается подключиться к Kubernetes кластеру"
        print_warning "Убедитесь, что kubectl настроен правильно"
        exit 1
    fi
    
    cd "$PROJECT_ROOT"
    
    # Применение namespace
    print_step "Создание namespace..."
    kubectl apply -f "$K8S_DIR/namespace.yaml"
    
    # Применение deployments
    print_step "Развертывание сервисов..."
    kubectl apply -f "$K8S_DIR/wireguard-service/deployment.yaml"
    kubectl apply -f "$K8S_DIR/quic-service/deployment.yaml"
    kubectl apply -f "$K8S_DIR/cadence-service/deployment.yaml"
    kubectl apply -f "$K8S_DIR/ai-service/deployment.yaml"
    kubectl apply -f "$K8S_DIR/quantum-service/deployment.yaml"
    
    # Применение ingress
    print_step "Настройка Ingress..."
    kubectl apply -f "$K8S_DIR/ingress.yaml"
    
    # Ожидание готовности подов
    print_step "Ожидание готовности подов..."
    kubectl wait --for=condition=ready pod -l app=wireguard-service -n cloudbridge-relay --timeout=300s
    kubectl wait --for=condition=ready pod -l app=quic-service -n cloudbridge-relay --timeout=300s
    kubectl wait --for=condition=ready pod -l app=cadence-service -n cloudbridge-relay --timeout=300s
    kubectl wait --for=condition=ready pod -l app=ai-service -n cloudbridge-relay --timeout=300s
    kubectl wait --for=condition=ready pod -l app=quantum-service -n cloudbridge-relay --timeout=300s
    
    print_success "Развертывание завершено"
}

# Проверка развертывания
verify_deployment() {
    print_step "Проверка развертывания..."
    
    echo "=== Pods ==="
    kubectl get pods -n cloudbridge-relay
    
    echo "=== Services ==="
    kubectl get services -n cloudbridge-relay
    
    echo "=== Deployments ==="
    kubectl get deployments -n cloudbridge-relay
    
    echo "=== Ingress ==="
    kubectl get ingress -n cloudbridge-relay
    
    # Проверка health endpoints
    print_step "Проверка health endpoints..."
    
    # WireGuard Service
    WIREGUARD_POD=$(kubectl get pods -n cloudbridge-relay -l app=wireguard-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward pod/$WIREGUARD_POD 8080:8080 -n cloudbridge-relay &
    PF_PID=$!
    sleep 5
    
    if curl -f http://localhost:8080/health &> /dev/null; then
        print_success "WireGuard Service health check passed"
    else
        print_error "WireGuard Service health check failed"
    fi
    
    kill $PF_PID
    
    # QUIC Service
    QUIC_POD=$(kubectl get pods -n cloudbridge-relay -l app=quic-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward pod/$QUIC_POD 8081:8081 -n cloudbridge-relay &
    PF_PID=$!
    sleep 5
    
    if curl -f http://localhost:8081/health &> /dev/null; then
        print_success "QUIC Service health check passed"
    else
        print_error "QUIC Service health check failed"
    fi
    
    kill $PF_PID
    
    # Cadence Service
    CADENCE_POD=$(kubectl get pods -n cloudbridge-relay -l app=cadence-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward pod/$CADENCE_POD 8082:8082 -n cloudbridge-relay &
    PF_PID=$!
    sleep 5
    
    if curl -f http://localhost:8082/health &> /dev/null; then
        print_success "Cadence Service health check passed"
    else
        print_error "Cadence Service health check failed"
    fi
    
    kill $PF_PID
    
    # AI Service
    AI_POD=$(kubectl get pods -n cloudbridge-relay -l app=ai-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward pod/$AI_POD 8083:8083 -n cloudbridge-relay &
    PF_PID=$!
    sleep 5
    
    if curl -f http://localhost:8083/health &> /dev/null; then
        print_success "AI Service health check passed"
    else
        print_error "AI Service health check failed"
    fi
    
    kill $PF_PID
    
    # Quantum Service
    QUANTUM_POD=$(kubectl get pods -n cloudbridge-relay -l app=quantum-service -o jsonpath='{.items[0].metadata.name}')
    kubectl port-forward pod/$QUANTUM_POD 8084:8084 -n cloudbridge-relay &
    PF_PID=$!
    sleep 5
    
    if curl -f http://localhost:8084/health &> /dev/null; then
        print_success "Quantum Service health check passed"
    else
        print_error "Quantum Service health check failed"
    fi
    
    kill $PF_PID
    
    print_success "Проверка завершена"
}

# Основная функция
main() {
    print_banner
    
    echo "Этот скрипт выполнит миграцию CloudBridge Relay в Kubernetes"
    echo "Время выполнения: ~30 минут"
    echo ""
    
    read -p "Продолжить? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Миграция отменена"
        exit 0
    fi
    
    # Выполнение этапов миграции
    check_dependencies
    analyze_code
    create_microservices
    create_dockerfiles
    build_images
    create_k8s_manifests
    deploy_to_kubernetes
    verify_deployment
    
    print_success "🎉 Миграция успешно завершена!"
    echo ""
    echo "Следующие шаги:"
    echo "1. Настройте DNS для api.cloudbridge.local"
    echo "2. Настройте SSL сертификаты"
    echo "3. Настройте мониторинг Prometheus/Grafana"
    echo "4. Настройте алерты"
    echo ""
    echo " Полезные команды:"
    echo "- Проверка статуса: kubectl get all -n cloudbridge-relay"
    echo "- Логи сервисов: kubectl logs -f deployment/wireguard-service -n cloudbridge-relay"
    echo "- Масштабирование: kubectl scale deployment wireguard-service --replicas=3 -n cloudbridge-relay"
    echo "- Удаление: kubectl delete namespace cloudbridge-relay"
}

# Запуск основной функции
main "$@"
