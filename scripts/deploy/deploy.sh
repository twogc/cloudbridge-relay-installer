#!/bin/bash

# Скрипт автоматического развертывания CloudBridge Relay в VK Cloud Kubernetes
# Автор: CloudBridge Team
# Версия: 1.0.0

set -e

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$REPO_ROOT"

# Deploy mode: manifests (default, file-by-file k8s-manifests/) | overlay (kustomize)
DEPLOY_MODE="${DEPLOY_MODE:-manifests}"
OVERLAY_DIR="${OVERLAY_DIR:-k8s/overlays/production}"
DRY_RUN="${DRY_RUN:-0}"
SKIP_MIRRORS="${SKIP_MIRRORS:-0}"

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

usage() {
    cat <<'EOF'
CloudBridge Relay — deploy to Kubernetes (VK Cloud)

Usage:
  ./scripts/deploy/deploy.sh [options]

Options:
  --use-overlay          Apply core stack via kustomize overlay
                         (k8s/overlays/production → k8s-manifests/00–08)
  --overlay DIR          Overlay path (default: k8s/overlays/production)
  --mode MODE            manifests (default) | overlay
  --dry-run              kubectl --dry-run=client where supported; no cluster wait
  --skip-mirrors         Skip Docker Hub mirror setup
  --render-overlay       Only print kustomize output (no apply); requires kubectl
  --validate-overlay     Run scripts/deploy/validate-overlay.sh --all and exit
  -h, --help             Show this help

Environment:
  DEPLOY_MODE=manifests|overlay
  OVERLAY_DIR=k8s/overlays/production
  DRY_RUN=1
  SKIP_MIRRORS=1
  KUBECONFIG=...

Canonical default remains file-by-file k8s-manifests/ (no overlay).
Zitadel (09–15) is not applied by this script — use zitadel-setup after core.
EOF
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --use-overlay)
                DEPLOY_MODE=overlay
                shift
                ;;
            --overlay)
                OVERLAY_DIR="$2"
                DEPLOY_MODE=overlay
                shift 2
                ;;
            --overlay=*)
                OVERLAY_DIR="${1#*=}"
                DEPLOY_MODE=overlay
                shift
                ;;
            --mode)
                DEPLOY_MODE="$2"
                shift 2
                ;;
            --mode=*)
                DEPLOY_MODE="${1#*=}"
                shift
                ;;
            --dry-run)
                DRY_RUN=1
                shift
                ;;
            --skip-mirrors)
                SKIP_MIRRORS=1
                shift
                ;;
            --render-overlay)
                DEPLOY_MODE=overlay
                RENDER_ONLY=1
                shift
                ;;
            --validate-overlay)
                VALIDATE_ONLY=1
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    case "$DEPLOY_MODE" in
        manifests|overlay) ;;
        *)
            log_error "Invalid DEPLOY_MODE/mode: $DEPLOY_MODE (use manifests|overlay)"
            exit 1
            ;;
    esac
}

# kubectl apply helper (supports dry-run)
kapply() {
    if [[ "$DRY_RUN" == "1" ]]; then
        kubectl apply --dry-run=client "$@"
    else
        kubectl apply "$@"
    fi
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("kubectl" "docker")
    # helm/terraform optional for core app deploy; warn only
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "$dep не установлен. Пожалуйста, установите его перед продолжением."
            exit 1
        fi
    done
    for dep in helm terraform; do
        if ! command -v "$dep" &> /dev/null; then
            log_warning "$dep не найден (не обязателен для core deploy.sh)"
        fi
    done

    if [[ "$DEPLOY_MODE" == "overlay" ]]; then
        if ! kubectl kustomize "$OVERLAY_DIR" &> /dev/null; then
            log_error "kubectl kustomize failed for $OVERLAY_DIR (need kubectl with kustomize support)"
            exit 1
        fi
        log_info "Overlay mode: $OVERLAY_DIR"
    fi
    
    log_success "Все зависимости установлены (mode=$DEPLOY_MODE)"
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
        kapply -f k8s-manifests/00-namespace.yaml
        log_success "Namespace cloudbridge создан"
    fi
}

# Создание PVC
create_pvc() {
    log_info "Создание PersistentVolumeClaim..."
    
    kapply -f k8s-manifests/01-pvc.yaml
    
    if [[ "$DRY_RUN" != "1" ]]; then
        # Ожидание готовности PVC
        log_info "Ожидание готовности PVC..."
        kubectl wait --for=condition=Bound pvc/cloudbridge-logs-pvc -n cloudbridge --timeout=300s
        kubectl wait --for=condition=Bound pvc/cadence-pvc -n cloudbridge --timeout=300s
        kubectl wait --for=condition=Bound pvc/registry-pvc -n cloudbridge --timeout=300s
    fi
    
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
    
    kapply -f k8s-manifests/03-configmap.yaml
    log_success "ConfigMap создан"
}

# Развертывание registry
deploy_registry() {
    log_info "Развертывание Docker Registry..."
    
    kapply -f k8s-manifests/04-registry.yaml
    
    if [[ "$DRY_RUN" != "1" ]]; then
        # Ожидание готовности registry
        log_info "Ожидание готовности registry..."
        kubectl wait --for=condition=Available deployment/docker-registry -n cloudbridge --timeout=300s
    fi
    
    log_success "Docker Registry развернут"
}

# Загрузка образов в registry
load_images() {
    log_info "Загрузка образов в registry..."
    
    local registry_host="localhost:5000"
    local ai_dockerfile="docker/dockerfiles/Dockerfile.ai-service"
    
    # Проверка существования образов
    if ! docker image inspect cloudbridge-relay:latest &> /dev/null; then
        log_warning "Образ cloudbridge-relay:latest не найден. Создание образа..."
        docker build -f docker/Dockerfile -t cloudbridge-relay:latest .
    fi
    
    if ! docker image inspect ai-service:latest &> /dev/null; then
        log_warning "Образ ai-service:latest не найден. Сборка из ${ai_dockerfile}..."
        docker build -f "${ai_dockerfile}" -t ai-service:latest .
    fi
    
    log_info "Cadence: образ ubercadence/server:v0.22.4 берётся с Docker Hub (см. k8s-manifests/05-deployments.yaml), push в локальный registry не требуется"
    
    # Настройка port-forward для registry
    log_info "Настройка port-forward для registry..."
    kubectl port-forward service/docker-registry 5000:5000 -n cloudbridge &
    PORT_FORWARD_PID=$!
    
    # Ожидание готовности registry
    sleep 10
    
    # Тегирование и загрузка образов (имена совпадают с image: в 05-deployments.yaml)
    log_info "Тегирование образов для ${registry_host}..."
    docker tag cloudbridge-relay:latest "${registry_host}/cloudbridge-relay:latest"
    docker tag ai-service:latest "${registry_host}/ai-service:latest"
    
    log_info "Загрузка образов в registry..."
    docker push "${registry_host}/cloudbridge-relay:latest"
    docker push "${registry_host}/ai-service:latest"
    
    # Остановка port-forward
    kill $PORT_FORWARD_PID 2>/dev/null || true
    
    log_success "Образы cloudbridge-relay и ai-service загружены в registry"
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
    
    kapply -f k8s-manifests/05-deployments.yaml
    kapply -f k8s-manifests/06-services.yaml
    
    wait_core_deployments
    log_success "Сервисы развернуты"
}

# Создание Ingress
create_ingress() {
    log_info "Создание Ingress..."
    
    kapply -f k8s-manifests/07-ingress.yaml
    log_success "Ingress создан"
}

# Создание ServiceMonitor
create_servicemonitor() {
    log_info "Создание ServiceMonitor..."
    
    kapply -f k8s-manifests/08-servicemonitor.yaml
    log_success "ServiceMonitor создан"
}

wait_core_deployments() {
    if [[ "$DRY_RUN" == "1" ]]; then
        log_info "dry-run: skip waiting for deployments"
        return 0
    fi
    log_info "Ожидание готовности сервисов..."
    kubectl wait --for=condition=Available deployment/cloudbridge-relay -n cloudbridge --timeout=300s || true
    kubectl wait --for=condition=Available deployment/ai-service -n cloudbridge --timeout=300s || true
    kubectl wait --for=condition=Available deployment/cadence -n cloudbridge --timeout=300s || true
}

# Apply production kustomize overlay (core stack 00–08)
apply_overlay() {
    log_info "Applying kustomize overlay: $OVERLAY_DIR"
    if [[ ! -f "$OVERLAY_DIR/kustomization.yaml" ]]; then
        log_error "Missing $OVERLAY_DIR/kustomization.yaml"
        exit 1
    fi
    if [[ "$DRY_RUN" == "1" ]]; then
        kubectl apply -k "$OVERLAY_DIR" --dry-run=client
    else
        kubectl apply -k "$OVERLAY_DIR"
    fi
    log_success "Overlay applied: $OVERLAY_DIR"
}

render_overlay() {
    log_info "Rendering overlay (no apply): $OVERLAY_DIR"
    kubectl kustomize "$OVERLAY_DIR"
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
    parse_args "$@"
    RENDER_ONLY="${RENDER_ONLY:-0}"
    VALIDATE_ONLY="${VALIDATE_ONLY:-0}"

    echo "Развертывание CloudBridge Relay в VK Cloud Kubernetes"
    echo "=================================================="
    echo "Mode: $DEPLOY_MODE"
    if [[ "$DEPLOY_MODE" == "overlay" ]]; then
        echo "Overlay: $OVERLAY_DIR"
    fi
    if [[ "$DRY_RUN" == "1" ]]; then
        echo "Dry-run: client-side only"
    fi
    echo ""

    if [[ "$VALIDATE_ONLY" == "1" ]]; then
        exec "$REPO_ROOT/scripts/deploy/validate-overlay.sh" --all
    fi

    if [[ "$RENDER_ONLY" == "1" ]]; then
        check_dependencies
        render_overlay
        exit 0
    fi
    
    check_dependencies
    if [[ "$SKIP_MIRRORS" != "1" ]]; then
        setup_docker_mirrors
    else
        log_info "Skipping Docker mirrors (--skip-mirrors)"
    fi
    if [[ "$DRY_RUN" != "1" ]]; then
        check_cluster
    else
        log_warning "dry-run: skip cluster-info check"
    fi

    if [[ "$DEPLOY_MODE" == "overlay" ]]; then
        # Bootstrap secrets the same way as manifests mode (overlay 02 may be placeholders).
        create_namespace
        create_secrets
        # Full core stack via kustomize (00–08)
        apply_overlay
        if [[ "$DRY_RUN" != "1" ]]; then
            log_info "Ожидание готовности registry (overlay)..."
            kubectl wait --for=condition=Available deployment/docker-registry -n cloudbridge --timeout=300s || true
        fi
        if [[ "$DRY_RUN" != "1" ]]; then
            load_images
            create_registry_secret
            # Re-apply so pods restart against pushed images if needed
            apply_overlay
            wait_core_deployments
            check_deployment
        fi
    else
        create_namespace
        create_pvc
        create_secrets
        create_configmap
        deploy_registry
        if [[ "$DRY_RUN" != "1" ]]; then
            load_images
        else
            log_info "dry-run: skip load_images"
        fi
        create_registry_secret
        deploy_services
        create_ingress
        create_servicemonitor
        if [[ "$DRY_RUN" != "1" ]]; then
            check_deployment
        fi
    fi
    
    echo ""
    log_success "Развертывание завершено! (mode=$DEPLOY_MODE)"
    echo ""
    log_info "Полезные команды:"
    echo "  kubectl get pods -n cloudbridge                    # Статус подов"
    echo "  kubectl logs -f -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge  # Логи CloudBridge Relay"
    echo "  kubectl port-forward service/cloudbridge-relay-service 9090:9090 -n cloudbridge  # Доступ к сервису"
    echo ""
    log_info "Overlay (optional):"
    echo "  ./scripts/deploy/deploy.sh --use-overlay"
    echo "  ./scripts/deploy/deploy.sh --render-overlay"
    echo ""
    log_info "Мониторинг:"
    echo "  kubectl port-forward service/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring  # Prometheus"
    echo "  kubectl port-forward service/prometheus-grafana 3000:80 -n monitoring  # Grafana"
    echo ""
}

# Запуск скрипта
main "$@"
