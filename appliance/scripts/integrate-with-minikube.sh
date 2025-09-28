#!/bin/bash
# Скрипт интеграции appliance с существующим minikube

set -e

# Конфигурация
SOURCE_DIR="/home/ubuntu/cloudbridge-relay-installer"
APPLIANCE_DIR="/home/ubuntu/cloudbridge-relay-installer/appliance"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверить наличие minikube
check_minikube() {
    if ! command -v minikube &> /dev/null; then
        log_error "Minikube не найден. Установите minikube сначала."
        exit 1
    fi
    
    if ! minikube status &> /dev/null; then
        log_warning "Minikube не запущен. Запускаю..."
        minikube start
    fi
    
    log_success "Minikube готов к работе"
}

# Копировать основной код
copy_main_code() {
    log_info "Копирование основного кода в appliance..."
    ./scripts/copy-main-code.sh
}

# Обновить Dockerfile для minikube
update_dockerfile() {
    log_info "Обновление Dockerfile для работы с minikube..."
    
    # Создать Dockerfile для minikube
    cat > "$APPLIANCE_DIR/docker/Dockerfile.minikube" << 'EOF'
# CloudBridge Appliance for Minikube
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy source code
COPY ../src/ ./

# Build CloudBridge Relay
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=appliance-$(date +%Y%m%d)" \
    -o cloudbridge-appliance \
    ./cmd/relay

# Final stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    iptables \
    iproute2 \
    wireguard-tools \
    curl \
    jq

# Create non-root user
RUN adduser -D -s /bin/sh cloudbridge

# Copy binary
COPY --from=builder /app/cloudbridge-appliance /usr/local/bin/
COPY --from=builder /app/configs/appliance-config.yaml /etc/cloudbridge/

# Create directories
RUN mkdir -p /var/log/cloudbridge \
    /var/lib/cloudbridge \
    /etc/cloudbridge/ssl \
    /etc/cloudbridge/wireguard

# Set permissions
RUN chown -R cloudbridge:cloudbridge /var/log/cloudbridge \
    /var/lib/cloudbridge \
    /etc/cloudbridge

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Expose ports
EXPOSE 8444 8080 19302 3478 5553

# Switch to non-root user
USER cloudbridge

# Start CloudBridge Appliance
CMD ["cloudbridge-appliance", "--config", "/etc/cloudbridge/appliance-config.yaml"]
EOF

    log_success "Dockerfile для minikube создан"
}

# Создать Kubernetes манифесты для appliance
create_k8s_manifests() {
    log_info "Создание Kubernetes манифестов для appliance..."
    
    mkdir -p "$APPLIANCE_DIR/k8s"
    
    # Deployment
    cat > "$APPLIANCE_DIR/k8s/appliance-deployment.yaml" << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-appliance
  labels:
    app: cloudbridge-appliance
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudbridge-appliance
  template:
    metadata:
      labels:
        app: cloudbridge-appliance
    spec:
      containers:
      - name: cloudbridge-appliance
        image: cloudbridge-appliance:latest
        ports:
        - containerPort: 8444
          name: grpc
        - containerPort: 8080
          name: http
        - containerPort: 19302
          name: wireguard
        - containerPort: 3478
          name: stun-turn
        - containerPort: 5553
          name: quic
        env:
        - name: CLOUDBRIDGE_MODE
          value: "appliance"
        - name: CLOUDBRIDGE_LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "2Gi"
            cpu: "1000m"
          limits:
            memory: "4Gi"
            cpu: "2000m"
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
EOF

    # Service
    cat > "$APPLIANCE_DIR/k8s/appliance-service.yaml" << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-appliance-service
spec:
  selector:
    app: cloudbridge-appliance
  ports:
  - name: grpc
    port: 8444
    targetPort: 8444
  - name: http
    port: 8080
    targetPort: 8080
  - name: wireguard
    port: 19302
    targetPort: 19302
  - name: stun-turn
    port: 3478
    targetPort: 3478
  - name: quic
    port: 5553
    targetPort: 5553
  type: LoadBalancer
EOF

    # ConfigMap
    cat > "$APPLIANCE_DIR/k8s/appliance-configmap.yaml" << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-appliance-config
data:
  appliance-config.yaml: |
    app:
      name: "CloudBridge Appliance"
      version: "1.0.0"
      mode: "appliance"
      log_level: "info"
    
    server:
      host: "0.0.0.0"
      port: 8444
      http_port: 8080
    
    relay:
      enabled: true
      mode: "appliance"
      throughput_limit: "2Gbps"
      max_connections: 1000
EOF

    log_success "Kubernetes манифесты созданы"
}

# Создать скрипт развертывания в minikube
create_deploy_script() {
    log_info "Создание скрипта развертывания в minikube..."
    
    cat > "$APPLIANCE_DIR/scripts/deploy-to-minikube.sh" << 'EOF'
#!/bin/bash
# Развертывание CloudBridge Appliance в minikube

set -e

# Цвета
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Проверить minikube
if ! minikube status &> /dev/null; then
    log_info "Запуск minikube..."
    minikube start
fi

# Собрать Docker образ в minikube
log_info "Сборка Docker образа в minikube..."
eval $(minikube docker-env)
docker build -f docker/Dockerfile.minikube -t cloudbridge-appliance:latest .

# Применить Kubernetes манифесты
log_info "Применение Kubernetes манифестов..."
kubectl apply -f k8s/appliance-configmap.yaml
kubectl apply -f k8s/appliance-deployment.yaml
kubectl apply -f k8s/appliance-service.yaml

# Ждать готовности
log_info "Ожидание готовности pod..."
kubectl wait --for=condition=ready pod -l app=cloudbridge-appliance --timeout=300s

# Получить URL сервиса
log_info "Получение URL сервиса..."
SERVICE_URL=$(minikube service cloudbridge-appliance-service --url)
echo "CloudBridge Appliance доступен по адресу: $SERVICE_URL"

log_success "CloudBridge Appliance успешно развернут в minikube!"
EOF

    chmod +x "$APPLIANCE_DIR/scripts/deploy-to-minikube.sh"
    log_success "Скрипт развертывания создан"
}

# Создать скрипт управления
create_management_script() {
    log_info "Создание скрипта управления appliance в minikube..."
    
    cat > "$APPLIANCE_DIR/scripts/manage-minikube-appliance.sh" << 'EOF'
#!/bin/bash
# Управление CloudBridge Appliance в minikube

set -e

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Функции управления
start_appliance() {
    log_info "Запуск CloudBridge Appliance в minikube..."
    kubectl scale deployment cloudbridge-appliance --replicas=1
    kubectl wait --for=condition=ready pod -l app=cloudbridge-appliance --timeout=300s
    log_success "Appliance запущен"
}

stop_appliance() {
    log_info "Остановка CloudBridge Appliance..."
    kubectl scale deployment cloudbridge-appliance --replicas=0
    log_success "Appliance остановлен"
}

restart_appliance() {
    log_info "Перезапуск CloudBridge Appliance..."
    kubectl rollout restart deployment cloudbridge-appliance
    kubectl wait --for=condition=ready pod -l app=cloudbridge-appliance --timeout=300s
    log_success "Appliance перезапущен"
}

show_status() {
    log_info "Статус CloudBridge Appliance:"
    echo
    kubectl get pods -l app=cloudbridge-appliance
    echo
    kubectl get services cloudbridge-appliance-service
    echo
    kubectl get configmap cloudbridge-appliance-config
}

show_logs() {
    log_info "Логи CloudBridge Appliance:"
    kubectl logs -l app=cloudbridge-appliance --tail=100
}

show_metrics() {
    log_info "Метрики CloudBridge Appliance:"
    SERVICE_URL=$(minikube service cloudbridge-appliance-service --url 2>/dev/null || echo "http://localhost:8080")
    curl -s "$SERVICE_URL/metrics" | grep -E "(appliance_|cloudbridge_)" | head -20
}

# Основная функция
main() {
    local command="$1"
    
    case "$command" in
        start)
            start_appliance
            ;;
        stop)
            stop_appliance
            ;;
        restart)
            restart_appliance
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        metrics)
            show_metrics
            ;;
        *)
            echo "Использование: $0 {start|stop|restart|status|logs|metrics}"
            exit 1
            ;;
    esac
}

main "$@"
EOF

    chmod +x "$APPLIANCE_DIR/scripts/manage-minikube-appliance.sh"
    log_success "Скрипт управления создан"
}

# Основная функция интеграции
main() {
    log_info "Интеграция CloudBridge Appliance с minikube..."
    
    check_minikube
    copy_main_code
    update_dockerfile
    create_k8s_manifests
    create_deploy_script
    create_management_script
    
    log_success "Интеграция завершена!"
    echo
    echo "📋 Что дальше:"
    echo "1. Развернуть appliance: ./scripts/deploy-to-minikube.sh"
    echo "2. Управлять appliance: ./scripts/manage-minikube-appliance.sh"
    echo "3. Проверить статус: ./scripts/manage-minikube-appliance.sh status"
    echo "4. Посмотреть логи: ./scripts/manage-minikube-appliance.sh logs"
    echo
    echo "🌐 После развертывания appliance будет доступен через minikube service"
}

main "$@"
