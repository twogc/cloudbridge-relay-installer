#!/bin/bash

# 🚀 CloudBridge Relay - Enhanced Components Build & Deploy Script
# ================================================================

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
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✅${NC} $1"
}

warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠️${NC} $1"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ❌${NC} $1"
}

# Проверка предварительных требований
check_requirements() {
    log "Проверка предварительных требований..."
    
    # Проверка Docker
    if ! command -v docker &> /dev/null; then
        error "Docker не установлен"
        exit 1
    fi
    
    # Проверка kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не установлен"
        exit 1
    fi
    
    # Проверка minikube
    if ! command -v minikube &> /dev/null; then
        error "minikube не установлен"
        exit 1
    fi
    
    # Проверка статуса minikube
    if ! minikube status &> /dev/null; then
        error "minikube не запущен"
        exit 1
    fi
    
    success "Все предварительные требования выполнены"
}

# Создание уникального тега сборки
create_build_tag() {
    BUILD_TAG="enhanced-$(date +%Y%m%d-%H%M%S)"
    log "Создан тег сборки: $BUILD_TAG"
    echo $BUILD_TAG
}

# Сборка Docker образа
build_image() {
    local build_tag=$1
    
    log "Сборка Docker образа с тегом: $build_tag"
    
    # Настройка Docker для minikube
    eval $(minikube docker-env)
    
    # Сборка образа
    docker build -t cloudbridge-relay:$build_tag .
    
    if [ $? -eq 0 ]; then
        success "Docker образ собран успешно"
    else
        error "Ошибка сборки Docker образа"
        exit 1
    fi
}

# Загрузка образа в minikube
load_image() {
    local build_tag=$1
    
    log "Загрузка образа в minikube..."
    
    minikube image load cloudbridge-relay:$build_tag
    
    if [ $? -eq 0 ]; then
        success "Образ загружен в minikube"
    else
        error "Ошибка загрузки образа в minikube"
        exit 1
    fi
}

# Обновление deployment
update_deployment() {
    local build_tag=$1
    
    log "Обновление Kubernetes deployment..."
    
    # Обновляем образ в deployment
    kubectl set image deployment/cloudbridge-relay-p2p \
        cloudbridge-relay-p2p=cloudbridge-relay:$build_tag \
        -n cloudbridge
    
    if [ $? -eq 0 ]; then
        success "Deployment обновлен"
    else
        error "Ошибка обновления deployment"
        exit 1
    fi
}

# Ожидание готовности подов
wait_for_pods() {
    log "Ожидание готовности подов..."
    
    kubectl rollout status deployment/cloudbridge-relay-p2p -n cloudbridge --timeout=300s
    
    if [ $? -eq 0 ]; then
        success "Поды готовы к работе"
    else
        error "Таймаут ожидания готовности подов"
        exit 1
    fi
}

# Проверка статуса развертывания
check_deployment_status() {
    log "Проверка статуса развертывания..."
    
    # Получаем статус подов
    kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p
    
    # Получаем логи
    log "Получение логов последнего пода..."
    POD_NAME=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p -o jsonpath='{.items[0].metadata.name}')
    
    if [ ! -z "$POD_NAME" ]; then
        kubectl logs $POD_NAME -n cloudbridge --tail=50
    fi
}

# Проверка работы сервисов
check_services() {
    log "Проверка работы сервисов..."
    
    # Проверяем доступность health endpoint
    local pod_ip=$(kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p -o jsonpath='{.items[0].status.podIP}')
    
    if [ ! -z "$pod_ip" ]; then
        log "Проверка health endpoint на $pod_ip:8080..."
        
        # Пробуем подключиться к health endpoint
        kubectl run test-pod --image=curlimages/curl --rm -i --restart=Never -- \
            curl -f http://$pod_ip:8080/health || warning "Health endpoint недоступен"
    fi
}

# Основная функция
main() {
    log "🚀 Запуск сборки и развертывания CloudBridge Relay Enhanced"
    log "=============================================================="
    
    # Проверяем предварительные требования
    check_requirements
    
    # Создаем тег сборки
    BUILD_TAG=$(create_build_tag)
    
    # Собираем образ
    build_image $BUILD_TAG
    
    # Загружаем образ в minikube
    load_image $BUILD_TAG
    
    # Обновляем deployment
    update_deployment $BUILD_TAG
    
    # Ждем готовности подов
    wait_for_pods
    
    # Проверяем статус
    check_deployment_status
    
    # Проверяем сервисы
    check_services
    
    success "🎉 Развертывание завершено успешно!"
    log "Тег образа: $BUILD_TAG"
    log "Для просмотра логов: kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p -f"
    log "Для проверки статуса: kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --build-only   Только сборка образа без развертывания"
        echo "  --deploy-only  Только развертывание (требует существующий образ)"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Полная сборка и развертывание"
        echo "  $0 --build-only       # Только сборка"
        echo "  $0 --deploy-only      # Только развертывание"
        exit 0
        ;;
    --build-only)
        log "Режим: только сборка"
        check_requirements
        BUILD_TAG=$(create_build_tag)
        build_image $BUILD_TAG
        load_image $BUILD_TAG
        success "Сборка завершена. Тег: $BUILD_TAG"
        ;;
    --deploy-only)
        log "Режим: только развертывание"
        check_requirements
        # Используем последний доступный тег
        BUILD_TAG=$(docker images --format "table {{.Tag}}" cloudbridge-relay | grep enhanced | head -1)
        if [ -z "$BUILD_TAG" ]; then
            error "Не найден образ cloudbridge-relay с тегом enhanced"
            exit 1
        fi
        log "Используется тег: $BUILD_TAG"
        update_deployment $BUILD_TAG
        wait_for_pods
        check_deployment_status
        check_services
        success "Развертывание завершено"
        ;;
    *)
        main
        ;;
esac


