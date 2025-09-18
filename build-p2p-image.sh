#!/bin/bash

# 🚀 Скрипт для сборки образа с P2P изменениями
# Использование: ./build-p2p-image.sh

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
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Проверка предварительных требований
check_requirements() {
    log "Проверяем предварительные требования..."
    
    if ! command -v go &> /dev/null; then
        error "Go не установлен"
    fi
    
    if ! command -v docker &> /dev/null; then
        error "Docker не установлен"
    fi
    
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не установлен"
    fi
    
    success "Все предварительные требования выполнены"
}

# Сборка нового бинарника
build_binary() {
    log "Собираем новый бинарник с P2P изменениями..."
    
    if go build -o cloudbridge-relay-p2p-new ./cmd/relay; then
        success "Бинарник собран: cloudbridge-relay-p2p-new"
    else
        error "Ошибка сборки бинарника"
    fi
    
    # Проверяем размер бинарника
    size=$(ls -lh cloudbridge-relay-p2p-new | awk '{print $5}')
    log "Размер бинарника: $size"
}

# Сборка Docker образа
build_image() {
    log "Собираем Docker образ..."
    
    # Создаем новый тег
    export NEW_TAG="p2p-routing-$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD)"
    log "Новый тег: $NEW_TAG"
    
    # Собираем образ
    if docker build -f Dockerfile.p2p -t cloudbridge-relay:$NEW_TAG .; then
        success "Образ собран: cloudbridge-relay:$NEW_TAG"
    else
        error "Ошибка сборки образа"
    fi
    
    # Сохраняем тег для дальнейшего использования
    echo "$NEW_TAG" > .last_build_tag
}

# Загрузка образа в Minikube
load_to_minikube() {
    log "Загружаем образ в Minikube..."
    
    if [ -f .last_build_tag ]; then
        TAG=$(cat .last_build_tag)
        log "Используем тег: $TAG"
        
        if eval $(minikube docker-env) && docker tag cloudbridge-relay:$TAG cloudbridge-relay:$TAG; then
            success "Образ загружен в Minikube: cloudbridge-relay:$TAG"
        else
            error "Ошибка загрузки образа в Minikube"
        fi
    else
        error "Файл с тегом не найден"
    fi
}

# Обновление deployment
update_deployment() {
    log "Обновляем deployment..."
    
    if [ -f .last_build_tag ]; then
        TAG=$(cat .last_build_tag)
        log "Обновляем deployment с тегом: $TAG"
        
        if kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge-relay; then
            success "Deployment обновлен"
        else
            error "Ошибка обновления deployment"
        fi
        
        # Принудительно перезапускаем deployment
        log "Перезапускаем deployment..."
        if kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge-relay; then
            success "Deployment перезапущен"
        else
            error "Ошибка перезапуска deployment"
        fi
    else
        error "Файл с тегом не найден"
    fi
}

# Проверка развертывания
check_deployment() {
    log "Проверяем статус развертывания..."
    
    if kubectl rollout status deployment/cloudbridge-relay -n cloudbridge-relay --timeout=300s; then
        success "Deployment успешно обновлен"
    else
        error "Timeout при обновлении deployment"
    fi
    
    # Проверяем поды
    log "Статус подов:"
    kubectl get pods -n cloudbridge-relay
}

# Проверка логов
check_logs() {
    log "Проверяем логи на наличие P2P компонентов..."
    
    # Ждем немного, чтобы логи появились
    sleep 10
    
    if kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay --tail=50 | grep -q "P2P Relay components initialized successfully"; then
        success "P2P компоненты запущены успешно!"
    else
        warning "P2P компоненты не найдены в логах. Возможно, образ не содержит новые изменения."
        log "Проверьте логи вручную:"
        echo "kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay --tail=50"
    fi
}

# Показать информацию о доступе
show_access_info() {
    log "Информация о доступе к сервисам:"
    echo ""
    echo "🌐 HTTP API: http://localhost:8080"
    echo "   Health: http://localhost:8080/health"
    echo ""
    echo "🔗 P2P API: http://localhost:8082"
    echo "   Health: http://localhost:8082/health"
    echo ""
    echo "📊 Полезные команды:"
    echo "   kubectl get pods -n cloudbridge-relay"
    echo "   kubectl logs -n cloudbridge-relay deployment/cloudbridge-relay -f"
    echo "   kubectl describe pod -n cloudbridge-relay -l app=cloudbridge-relay"
    echo ""
}

# Основная функция
main() {
    echo "🚀 Начинаем сборку образа с P2P изменениями..."
    echo ""
    
    check_requirements
    build_binary
    build_image
    load_to_minikube
    update_deployment
    check_deployment
    check_logs
    show_access_info
    
    echo ""
    success "🎉 Сборка и развертывание завершены!"
    echo ""
    echo "Теперь вы можете тестировать P2P функциональность."
    echo "Проверьте логи на наличие сообщений о P2P компонентах."
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --build-only   Только сборка образа (без развертывания)"
        echo "  --check-only   Только проверка логов"
        echo ""
        exit 0
        ;;
    --build-only)
        log "Запускаем только сборку образа..."
        check_requirements
        build_binary
        build_image
        load_to_minikube
        exit 0
        ;;
    --check-only)
        log "Запускаем только проверку логов..."
        check_logs
        exit 0
        ;;
    "")
        main
        ;;
    *)
        error "Неизвестная опция: $1. Используйте --help для справки."
        ;;
esac
