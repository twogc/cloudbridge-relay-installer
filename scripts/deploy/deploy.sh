#!/bin/bash

# Скрипт автоматического развертывания CloudBridge Relay в VK Cloud Kubernetes
# Автор: CloudBridge Team
# Версия: 1.0.0

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("kubectl" "helm" "docker" "terraform")
    for dep in "${deps[@]}"; do
        if ! command -v $dep &> /dev/null; then
            log_error "$dep не установлен. Пожалуйста, установите его перед продолжением."
            exit 1
        fi
    done
    
    log_success "Все зависимости установлены"
}

# Настройка зеркал Docker Hub
setup_docker_mirrors() {
    log_info "Настройка зеркал Docker Hub..."
    
    # Проверка существования конфигурации
    if [ -f /etc/docker/daemon.json ]; then
        log_warning "Конфигурация Docker уже существует"
        if grep -q "registry-mirrors" /etc/docker/daemon.json; then
            log_warning "Зеркала уже настроены"
            return
        fi
    fi
    
    # Создание резервной копии
    if [ -f /etc/docker/daemon.json ]; then
        sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.backup
        log_info "Создана резервная копия /etc/docker/daemon.json.backup"
    fi
    
    # Настройка зеркал
    cat << EOF | sudo tee /etc/docker/daemon.json
{
  "registry-mirrors": [
    "https://dockerhub.timeweb.cloud",
    "https://huecker.io",
    "https://mirror.gcr.io",
    "https://c.163.com",
    "https://registry.docker-cn.com",
    "https://daocloud.io",
    "https://cr.yandex/mirror",
    "https://noohub.ru"
  ]
}
EOF
    
    # Перезапуск Docker
    log_info "Перезапуск Docker..."
    sudo systemctl restart docker
    
    # Проверка настройки
    log_info "Проверка настройки зеркал..."
    if docker info | grep -q "Registry Mirrors"; then
        log_success "Зеркала Docker Hub настроены"
    else
        log_warning "Не удалось проверить настройку зеркал"
    fi
}

# Проверка подключения к кластеру
check_cluster() {
    log_info "Проверка подключения к кластеру..."
    
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Не удается подключиться к кластеру Kubernetes"
        log_info "Убедитесь, что KUBECONFIG настроен правильно"
        exit 1
    fi
    
    local node_count=$(kubectl get nodes --no-headers | wc -l)
    log_success "Подключение к кластеру установлено ($node_count узлов)"
}

# Создание namespace
create_namespace() {
    log_info "Создание namespace cloudbridge..."
    
    if kubectl get namespace cloudbridge &> /dev/null; then
        log_warning "Namespace cloudbridge уже существует"
    else
        kubectl apply -f k8s-manifests/00-namespace.yaml
        log_success "Namespace cloudbridge создан"
    fi
}

# Создание PVC
create_pvc() {
    log_info "Создание PersistentVolumeClaim..."
    
    kubectl apply -f k8s-manifests/01-pvc.yaml
    
    # Ожидание готовности PVC
    log_info "Ожидание готовности PVC..."
    kubectl wait --for=condition=Bound pvc/cloudbridge-logs-pvc -n cloudbridge --timeout=300s
    kubectl wait --for=condition=Bound pvc/cadence-pvc -n cloudbridge --timeout=300s
    kubectl wait --for=condition=Bound pvc/registry-pvc -n cloudbridge --timeout=300s
    
    log_success "PVC созданы и готовы"
}

# Создание секретов
create_secrets() {
    log_info "Создание секретов..."
    
    # Проверка существования секретов
    if kubectl get secret cloudbridge-secrets -n cloudbridge &> /dev/null; then
        log_warning "Секреты уже существуют"
    else
        log_warning "ВНИМАНИЕ: Необходимо обновить секреты в файле k8s-manifests/02-secrets.yaml"
        log_info "Создание базовых секретов..."
        
        # Создание базовых секретов
        kubectl create secret generic cloudbridge-secrets \
          --from-literal=cassandra-username=cassandra \
          --from-literal=cassandra-password=cassandra \
          --from-literal=api-key=your-api-key \
          --from-literal=jwt-secret=your-jwt-secret \
          --namespace=cloudbridge --dry-run=client -o yaml | kubectl apply -f -
        
        kubectl create secret generic wireguard-keys \
          --from-literal=private-key=your-private-key \
          --from-literal=public-key=your-public-key \
          --namespace=cloudbridge --dry-run=client -o yaml | kubectl apply -f -
        
        log_success "Секреты созданы"
    fi
}

# Создание ConfigMap
create_configmap() {
    log_info "Создание ConfigMap..."
    
    kubectl apply -f k8s-manifests/03-configmap.yaml
    log_success "ConfigMap создан"
}

# Развертывание registry
deploy_registry() {
    log_info "Развертывание Docker Registry..."
    
    kubectl apply -f k8s-manifests/04-registry.yaml
    
    # Ожидание готовности registry
    log_info "Ожидание готовности registry..."
    kubectl wait --for=condition=Available deployment/docker-registry -n cloudbridge --timeout=300s
    
    log_success "Docker Registry развернут"
}

# Загрузка образов в registry
load_images() {
    log_info "Загрузка образов в registry..."
    
    # Проверка существования образов
    if ! docker image inspect cloudbridge-relay:latest &> /dev/null; then
        log_warning "Образ cloudbridge-relay:latest не найден. Создание образа..."
        docker build -f docker/Dockerfile -t cloudbridge-relay:latest .
    fi
    
    if ! docker image inspect ai-service:latest &> /dev/null; then
        log_warning "Образ ai-service:latest не найден. Создание образа..."
        # Создание простого образа для тестирования
        docker build -f - -t ai-service:latest << 'EOF'
FROM alpine:3.18
RUN apk add --no-cache wget
EXPOSE 8080
CMD ["sh", "-c", "while true; do echo 'AI Service running on port 8080'; sleep 30; done"]
EOF
    fi
    
    if ! docker image inspect cadence:latest &> /dev/null; then
        log_warning "Образ cadence:latest не найден. Создание образа..."
        # Создание простого образа для тестирования
        docker build -f - -t cadence:latest << 'EOF'
FROM alpine:3.18
RUN apk add --no-cache wget
EXPOSE 7933
CMD ["sh", "-c", "while true; do echo 'Cadence Service running on port 7933'; sleep 30; done"]
EOF
    fi
    
    # Настройка port-forward для registry
    log_info "Настройка port-forward для registry..."
    kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &
    PORT_FORWARD_PID=$!
    
    # Ожидание готовности registry
    sleep 10
    
    # Тегирование и загрузка образов
    log_info "Тегирование образов..."
    docker tag cloudbridge-relay:latest localhost:5000/cloudbridge-relay:latest
    docker tag ai-service:latest localhost:5000/ai-service:latest
    docker tag cadence:latest localhost:5000/cadence:latest
    
    log_info "Загрузка образов в registry..."
    docker push localhost:5000/cloudbridge-relay:latest
    docker push localhost:5000/ai-service:latest
    docker push localhost:5000/cadence:latest
    
    # Остановка port-forward
    kill $PORT_FORWARD_PID 2>/dev/null || true
    
    log_success "Образы загружены в registry"
}

# Создание секрета для registry
create_registry_secret() {
    log_info "Создание секрета для registry..."
    
    kubectl create secret docker-registry k8s-registry-creds \
      --docker-server=http://docker-registry:5000 \
      --docker-username=admin \
      --docker-password=admin \
      --namespace=cloudbridge --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Секрет для registry создан"
}

# Развертывание сервисов
deploy_services() {
    log_info "Развертывание сервисов..."
    
    kubectl apply -f k8s-manifests/05-deployments.yaml
    kubectl apply -f k8s-manifests/06-services.yaml
    
    # Ожидание готовности deployment
    log_info "Ожидание готовности сервисов..."
    kubectl wait --for=condition=Available deployment/cloudbridge-relay -n cloudbridge --timeout=300s || true
    kubectl wait --for=condition=Available deployment/ai-service -n cloudbridge --timeout=300s || true
    kubectl wait --for=condition=Available deployment/cadence -n cloudbridge --timeout=300s || true
    
    log_success "Сервисы развернуты"
}

# Создание Ingress
create_ingress() {
    log_info "Создание Ingress..."
    
    kubectl apply -f k8s-manifests/07-ingress.yaml
    log_success "Ingress создан"
}

# Создание ServiceMonitor
create_servicemonitor() {
    log_info "Создание ServiceMonitor..."
    
    kubectl apply -f k8s-manifests/08-servicemonitor.yaml
    log_success "ServiceMonitor создан"
}

# Проверка статуса развертывания
check_deployment() {
    log_info "Проверка статуса развертывания..."
    
    echo ""
    log_info "Статус подов:"
    kubectl get pods -n cloudbridge
    
    echo ""
    log_info "Статус сервисов:"
    kubectl get services -n cloudbridge
    
    echo ""
    log_info "Статус PVC:"
    kubectl get pvc -n cloudbridge
    
    echo ""
    log_info "События:"
    kubectl get events -n cloudbridge --sort-by='.lastTimestamp' | tail -10
}

# Основная функция
main() {
    echo "Развертывание CloudBridge Relay в VK Cloud Kubernetes"
    echo "=================================================="
    echo ""
    
    check_dependencies
    setup_docker_mirrors
    check_cluster
    create_namespace
    create_pvc
    create_secrets
    create_configmap
    deploy_registry
    load_images
    create_registry_secret
    deploy_services
    create_ingress
    create_servicemonitor
    check_deployment
    
    echo ""
    log_success "Развертывание завершено!"
    echo ""
    log_info "Полезные команды:"
    echo "  kubectl get pods -n cloudbridge                    # Статус подов"
    echo "  kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge  # Логи CloudBridge Relay"
    echo "  kubectl port-forward service/cloudbridge-relay-service 9090:9090 -n cloudbridge  # Доступ к сервису"
    echo ""
    log_info "Мониторинг:"
    echo "  kubectl port-forward service/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring  # Prometheus"
    echo "  kubectl port-forward service/prometheus-grafana 3000:80 -n monitoring  # Grafana"
    echo ""
}

# Запуск скрипта
main "$@"
