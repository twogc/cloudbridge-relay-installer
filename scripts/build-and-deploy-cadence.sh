#!/bin/bash

# Скрипт для сборки и развертывания Cadence с новым тегом
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

# Параметры по умолчанию
IMAGE_NAME="cloudbridge-relay"
TAG="latest"
REGISTRY=""
PUSH_TO_REGISTRY=false
MINIKUBE_BUILD=false

# Функция для показа помощи
show_help() {
    echo "Использование: $0 [ОПЦИИ]"
    echo ""
    echo "Опции:"
    echo "  -t, --tag TAG           Тег для образа (по умолчанию: latest)"
    echo "  -r, --registry REGISTRY Реестр Docker (например: localhost:5000)"
    echo "  -p, --push              Отправить образ в реестр"
    echo "  -m, --minikube          Собрать образ в Minikube"
    echo "  -h, --help              Показать эту справку"
    echo ""
    echo "Примеры:"
    echo "  $0 -t v1.0.0 -m                    # Собрать тег v1.0.0 в Minikube"
    echo "  $0 -t v1.0.0 -r localhost:5000 -p  # Собрать и отправить в локальный реестр"
    echo "  $0 -t dev -m                       # Собрать dev тег в Minikube"
}

# Парсинг аргументов командной строки
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -r|--registry)
            REGISTRY="$2"
            shift 2
            ;;
        -p|--push)
            PUSH_TO_REGISTRY=true
            shift
            ;;
        -m|--minikube)
            MINIKUBE_BUILD=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Неизвестная опция: $1"
            show_help
            exit 1
            ;;
    esac
done

# Проверяем, что Minikube запущен
check_minikube() {
    if [ "$MINIKUBE_BUILD" = true ]; then
        log "Проверяем статус Minikube..."
        if ! minikube status > /dev/null 2>&1; then
            error "Minikube не запущен. Запустите: minikube start"
            exit 1
        fi
        success "Minikube запущен"
    fi
}

# Собираем образ
build_image() {
    log "Собираем Docker образ..."
    
    # Формируем полное имя образа
    if [ -n "$REGISTRY" ]; then
        FULL_IMAGE_NAME="${REGISTRY}/${IMAGE_NAME}:${TAG}"
    else
        FULL_IMAGE_NAME="${IMAGE_NAME}:${TAG}"
    fi
    
    log "Имя образа: $FULL_IMAGE_NAME"
    
    # Собираем образ
    if [ "$MINIKUBE_BUILD" = true ]; then
        log "Собираем образ в Minikube..."
        eval $(minikube docker-env)
        docker build -f docker/Dockerfile.cadence -t "$FULL_IMAGE_NAME" .
    else
        log "Собираем образ локально..."
        docker build -f docker/Dockerfile.cadence -t "$FULL_IMAGE_NAME" .
    fi
    
    success "Образ собран: $FULL_IMAGE_NAME"
}

# Отправляем образ в реестр
push_image() {
    if [ "$PUSH_TO_REGISTRY" = true ]; then
        log "Отправляем образ в реестр..."
        docker push "$FULL_IMAGE_NAME"
        success "Образ отправлен в реестр"
    else
        log "Пропускаем отправку в реестр (используйте -p для отправки)"
    fi
}

# Обновляем deployment
update_deployment() {
    log "Обновляем deployment с новым образом..."
    
    # Создаем временный файл с обновленным deployment
    TEMP_DEPLOYMENT="/tmp/cadence-deployment-${TAG}.yaml"
    
    # Копируем оригинальный deployment
    cp k8s/deployments/cadence-postgres.yaml "$TEMP_DEPLOYMENT"
    
    # Обновляем образ в deployment
    if [ "$MINIKUBE_BUILD" = true ]; then
        # Для Minikube используем локальный образ
        sed -i "s|image: ubercadence/server:v0.22.4|image: ${IMAGE_NAME}:${TAG}|g" "$TEMP_DEPLOYMENT"
    else
        # Для внешнего реестра используем полное имя
        sed -i "s|image: ubercadence/server:v0.22.4|image: ${FULL_IMAGE_NAME}|g" "$TEMP_DEPLOYMENT"
    fi
    
    # Применяем обновленный deployment
    kubectl apply -f "$TEMP_DEPLOYMENT" -n cloudbridge
    
    # Удаляем временный файл
    rm -f "$TEMP_DEPLOYMENT"
    
    success "Deployment обновлен с образом $FULL_IMAGE_NAME"
}

# Ждем готовности deployment
wait_for_deployment() {
    log "Ожидаем готовности deployment..."
    kubectl rollout status deployment/cadence -n cloudbridge --timeout=300s
    success "Deployment готов"
}

# Показываем информацию о развертывании
show_deployment_info() {
    log "Получаем информацию о развертывании..."
    
    MINIKUBE_IP=$(minikube ip)
    CADENCE_GRPC_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "30033")
    CADENCE_WEB_PORT=$(kubectl get service cadence-nodeport -n cloudbridge -o jsonpath='{.spec.ports[1].nodePort}' 2>/dev/null || echo "30034")
    
    echo ""
    echo "=========================================="
    echo "🎉 Cadence успешно обновлен!"
    echo "=========================================="
    echo ""
    echo "📋 Информация о развертывании:"
    echo "   Образ: $FULL_IMAGE_NAME"
    echo "   Тег: $TAG"
    echo "   Minikube IP: $MINIKUBE_IP"
    echo "   Cadence gRPC: $MINIKUBE_IP:$CADENCE_GRPC_PORT"
    echo "   Cadence Web UI: http://$MINIKUBE_IP:$CADENCE_WEB_PORT"
    echo ""
    echo " Полезные команды:"
    echo "   kubectl get pods -n cloudbridge"
    echo "   kubectl logs -f deployment/cadence -n cloudbridge"
    echo "   kubectl describe deployment cadence -n cloudbridge"
    echo ""
    echo "🌐 Открыть Cadence Web UI:"
    echo "   minikube service cadence-nodeport -n cloudbridge --url"
    echo ""
}

# Основная функция
main() {
    echo "🚀 Сборка и развертывание Cadence с тегом $TAG..."
    echo ""
    
    check_minikube
    build_image
    push_image
    update_deployment
    wait_for_deployment
    show_deployment_info
    
    success "Cadence успешно обновлен с тегом $TAG!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        main
        ;;
esac
