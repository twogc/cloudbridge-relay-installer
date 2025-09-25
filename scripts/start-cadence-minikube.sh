#!/bin/bash

# Скрипт для запуска Cadence в Minikube
# Автор: CloudBridge Relay Team
# Версия: 1.0

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверяем, что Minikube запущен
check_minikube() {
    log "Проверяем статус Minikube..."
    if ! minikube status > /dev/null 2>&1; then
        error "Minikube не запущен. Запустите: minikube start"
        exit 1
    fi
    success "Minikube запущен"
}

# Проверяем kubectl
check_kubectl() {
    log "Проверяем kubectl..."
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не найден. Установите kubectl"
        exit 1
    fi
    success "kubectl доступен"
}

# Создаем namespace
create_namespace() {
    log "Создаем namespace cloudbridge..."
    kubectl apply -f k8s/namespace.yaml
    success "Namespace cloudbridge создан"
}

# Создаем конфигурацию базы данных
create_database_config() {
    log "Используем SQLite для простоты..."
    success "Конфигурация базы данных готова"
}

# Запускаем базу данных (пропускаем для SQLite)
deploy_database() {
    log "SQLite не требует отдельного deployment..."
    success "База данных готова"
}

# Запускаем Cadence
deploy_cadence() {
    log "Развертываем Cadence..."
    kubectl apply -f k8s/deployments/cadence-working.yaml
    
    log "Ожидаем готовности Cadence..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cadence -n cloudbridge --timeout=300s
    success "Cadence развернут и готов"
}

# Создаем сервисы
create_services() {
    log "Создаем сервисы..."
    kubectl apply -f k8s/services/cadence-nodeport.yaml
    success "Сервисы созданы"
}

# Показываем информацию о доступе
show_access_info() {
    log "Получаем информацию о доступе..."
    
    MINIKUBE_IP=$(minikube ip)
    CADENCE_GRPC_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[0].nodePort}')
    CADENCE_WEB_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[1].nodePort}')
    
    echo ""
    echo "=========================================="
    echo "🎉 Cadence успешно развернут в Minikube!"
    echo "=========================================="
    echo ""
    echo "📋 Информация о доступе:"
    echo "   Minikube IP: $MINIKUBE_IP"
    echo "   Cadence gRPC: $MINIKUBE_IP:$CADENCE_GRPC_PORT"
    echo "   Cadence Web UI: http://$MINIKUBE_IP:$CADENCE_WEB_PORT"
    echo ""
    echo " Полезные команды:"
    echo "   kubectl get pods -n cloudbridge"
    echo "   kubectl logs -f deployment/cadence -n cloudbridge"
    echo "   kubectl logs -f statefulset/cassandra -n cloudbridge"
    echo ""
    echo "🌐 Открыть Cadence Web UI:"
    echo "   minikube service cadence-nodeport -n cloudbridge --url"
    echo ""
    echo "📊 Проверить статус:"
    echo "   kubectl get all -n cloudbridge"
    echo ""
}

# Основная функция
main() {
    echo "🚀 Запуск Cadence в Minikube..."
    echo ""
    
    check_minikube
    check_kubectl
    create_namespace
    create_database_config
    deploy_database
    deploy_cadence
    create_services
    show_access_info
    
    success "Cadence успешно запущен в Minikube!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    "stop")
        log "Останавливаем Cadence..."
        kubectl delete -f k8s/deployments/cadence-working.yaml --ignore-not-found=true
        kubectl delete -f k8s/services/cadence-nodeport.yaml --ignore-not-found=true
        success "Cadence остановлен"
        ;;
    "status")
        log "Проверяем статус Cadence..."
        kubectl get all -n cloudbridge
        ;;
    "logs")
        log "Показываем логи Cadence..."
        kubectl logs -f deployment/cadence -n cloudbridge
        ;;
    "cassandra-logs")
        log "Показываем логи Cassandra..."
        kubectl logs -f statefulset/cassandra -n cloudbridge
        ;;
    "web")
        log "Открываем Cadence Web UI..."
        minikube service cadence-nodeport -n cloudbridge --url
        ;;
    *)
        main
        ;;
esac
