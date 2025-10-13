#!/bin/bash

# ================================
# TURN Server Deployment Script
# ================================
# Разворачивает TURN server в Kubernetes cluster
# для CloudBridge Relay NAT traversal

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация
NAMESPACE="${NAMESPACE:-cloudbridge}"
TURN_PASSWORD="${TURN_PASSWORD:-$(openssl rand -base64 32)}"
DEPLOYMENT_FILE="k8s/deployments/turn-server-deployment.yaml"

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

# Проверка kubectl
check_kubectl() {
    log_info "Проверка kubectl..."
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl не установлен. Установите kubectl для продолжения."
        exit 1
    fi
    log_success "kubectl найден: $(kubectl version --client --short 2>/dev/null || echo 'kubectl installed')"
}

# Проверка доступа к кластеру
check_cluster_access() {
    log_info "Проверка доступа к Kubernetes кластеру..."
    if ! kubectl cluster-info &> /dev/null; then
        log_error "Нет доступа к Kubernetes кластеру. Проверьте kubeconfig."
        exit 1
    fi
    log_success "Доступ к кластеру подтвержден"
    kubectl cluster-info | head -n 1
}

# Создание namespace
create_namespace() {
    log_info "Проверка namespace '$NAMESPACE'..."
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        log_warning "Namespace '$NAMESPACE' уже существует"
    else
        log_info "Создание namespace '$NAMESPACE'..."
        kubectl create namespace "$NAMESPACE"
        log_success "Namespace '$NAMESPACE' создан"
    fi
}

# Создание secret для TURN credentials
create_turn_secret() {
    log_info "Создание Secret для TURN credentials..."

    # Проверяем существует ли уже secret
    if kubectl get secret turn-server-credentials -n "$NAMESPACE" &> /dev/null; then
        log_warning "Secret 'turn-server-credentials' уже существует"
        read -p "Пересоздать secret? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Пропускаем создание secret"
            return 0
        fi
        kubectl delete secret turn-server-credentials -n "$NAMESPACE"
    fi

    # Создаем secret
    kubectl create secret generic turn-server-credentials \
        -n "$NAMESPACE" \
        --from-literal=TURN_PASSWORD="$TURN_PASSWORD" \
        --from-literal=TURN_EXTERNAL_IP="0.0.0.0"

    log_success "Secret 'turn-server-credentials' создан"
    log_info "TURN Password (сохраните!): $TURN_PASSWORD"
}

# Проверка TLS сертификатов
check_tls_certs() {
    log_info "Проверка TLS сертификатов..."
    if ! kubectl get secret cloudbridge-tls -n "$NAMESPACE" &> /dev/null; then
        log_warning "Secret 'cloudbridge-tls' не найден"
        log_info "TURN server будет ожидать TLS сертификаты"
        log_info "Создайте secret 'cloudbridge-tls' с ключами tls.crt и tls.key"

        read -p "Продолжить без TLS? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_error "Развертывание отменено. Создайте TLS secret и повторите."
            exit 1
        fi

        # Создаем self-signed сертификат для dev
        log_info "Создание self-signed сертификата для development..."
        openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout /tmp/tls.key -out /tmp/tls.crt \
            -subj "/CN=turn.2gc.ru/O=2GC" \
            2>/dev/null

        kubectl create secret tls cloudbridge-tls \
            -n "$NAMESPACE" \
            --cert=/tmp/tls.crt \
            --key=/tmp/tls.key

        rm -f /tmp/tls.key /tmp/tls.crt
        log_success "Self-signed сертификат создан"
    else
        log_success "TLS сертификаты найдены"
    fi
}

# Развертывание TURN server
deploy_turn_server() {
    log_info "Развертывание TURN server..."

    if [ ! -f "$DEPLOYMENT_FILE" ]; then
        log_error "Deployment файл не найден: $DEPLOYMENT_FILE"
        exit 1
    fi

    kubectl apply -f "$DEPLOYMENT_FILE"
    log_success "TURN server deployment применен"
}

# Ожидание готовности pods
wait_for_pods() {
    log_info "Ожидание готовности TURN server pods..."

    kubectl wait --for=condition=ready pod \
        -l app=turn-server \
        -n "$NAMESPACE" \
        --timeout=300s || {
        log_warning "Pods не готовы в течение 5 минут"
        log_info "Проверка статуса pods:"
        kubectl get pods -n "$NAMESPACE" -l app=turn-server
        return 1
    }

    log_success "TURN server pods готовы"
}

# Получение external IP
get_external_ip() {
    log_info "Получение external IP LoadBalancer..."

    local max_attempts=30
    local attempt=0
    local external_ip=""

    while [ $attempt -lt $max_attempts ]; do
        external_ip=$(kubectl get svc turn-server -n "$NAMESPACE" \
            -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

        if [ -n "$external_ip" ]; then
            log_success "External IP получен: $external_ip"

            # Обновляем secret с external IP
            kubectl patch secret turn-server-credentials \
                -n "$NAMESPACE" \
                --type merge \
                -p "{\"stringData\":{\"TURN_EXTERNAL_IP\":\"$external_ip\"}}"

            log_success "Secret обновлен с external IP"
            return 0
        fi

        attempt=$((attempt + 1))
        log_info "Ожидание external IP... (попытка $attempt/$max_attempts)"
        sleep 10
    done

    log_warning "Не удалось получить external IP за отведенное время"
    log_info "LoadBalancer может быть еще в процессе создания"
    return 1
}

# Вывод статуса
show_status() {
    log_info "=== TURN Server Status ==="
    echo ""

    log_info "Pods:"
    kubectl get pods -n "$NAMESPACE" -l app=turn-server
    echo ""

    log_info "Service:"
    kubectl get svc turn-server -n "$NAMESPACE"
    echo ""

    log_info "HPA:"
    kubectl get hpa turn-server-hpa -n "$NAMESPACE" 2>/dev/null || log_info "HPA не найден"
    echo ""
}

# Тест connectivity
test_connectivity() {
    log_info "Тестирование TURN server connectivity..."

    local external_ip=$(kubectl get svc turn-server -n "$NAMESPACE" \
        -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")

    if [ -z "$external_ip" ]; then
        log_warning "External IP не доступен, пропускаем connectivity test"
        return 0
    fi

    log_info "Testing TURN server на $external_ip:3478..."

    # Простой UDP тест (требует nc/netcat)
    if command -v nc &> /dev/null; then
        if timeout 5 nc -vzu "$external_ip" 3478 2>&1 | grep -q "succeeded"; then
            log_success "TURN UDP порт 3478 доступен"
        else
            log_warning "TURN UDP порт может быть недоступен (это ожидаемо для UDP)"
        fi
    else
        log_info "netcat не установлен, пропускаем connectivity test"
    fi
}

# Вывод инструкций
show_instructions() {
    log_success "=== TURN Server Deployed Successfully ==="
    echo ""
    log_info "📝 Следующие шаги:"
    echo ""
    echo "1. Обновите configs/config.json с полученным external IP:"
    echo "   \"turn_pool\": {"
    echo "     \"pools\": {"
    echo "       \"primary\": {"
    echo "         \"servers\": ["
    echo "           {"
    echo "             \"address\": \"<EXTERNAL_IP>:3478\","
    echo "             \"username\": \"cloudbridge\","
    echo "             \"password\": \"$TURN_PASSWORD\""
    echo "           }"
    echo "         ]"
    echo "       }"
    echo "     }"
    echo "   }"
    echo ""
    echo "2. Проверьте metrics:"
    echo "   kubectl port-forward -n $NAMESPACE svc/turn-server 9641:9641"
    echo "   curl http://localhost:9641/metrics"
    echo ""
    echo "3. Проверьте logs:"
    echo "   kubectl logs -n $NAMESPACE -l app=turn-server --tail=100 -f"
    echo ""
    echo "4. Масштабирование (опционально):"
    echo "   kubectl scale deployment turn-server -n $NAMESPACE --replicas=3"
    echo ""
}

# Основной процесс
main() {
    echo ""
    log_info "==================================="
    log_info "TURN Server Deployment для CloudBridge"
    log_info "==================================="
    echo ""

    check_kubectl
    check_cluster_access
    create_namespace
    create_turn_secret
    check_tls_certs
    deploy_turn_server

    echo ""
    log_info "Ожидание развертывания..."
    sleep 5

    wait_for_pods
    get_external_ip

    echo ""
    show_status
    test_connectivity

    echo ""
    show_instructions

    log_success "Развертывание завершено! 🚀"
}

# Запуск
main "$@"
