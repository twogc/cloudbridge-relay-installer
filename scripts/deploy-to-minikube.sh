#!/bin/bash

# 🚀 Скрипт автоматического развертывания CloudBridge Relay в Minikube
# Версия: 1.0
# Дата: 13 сентября 2025

set -e  # Остановка при ошибке

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
    echo -e "${GREEN}$1${NC}"
}

warning() {
    echo -e "${YELLOW}  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Проверка предварительных требований
check_requirements() {
    log "Проверяем предварительные требования..."
    
    # Проверяем Minikube
    if ! command -v minikube &> /dev/null; then
        error "Minikube не установлен"
    fi
    
    # Проверяем kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не установлен"
    fi
    
    # Проверяем Docker
    if ! command -v docker &> /dev/null; then
        error "Docker не установлен"
    fi
    
    # Проверяем статус Minikube
    if ! minikube status &> /dev/null; then
        error "Minikube не запущен. Запустите: minikube start"
    fi
    
    success "Все предварительные требования выполнены"
}

# Сборка Docker образов
build_images() {
    log "Собираем Docker образы..."
    
    # Переходим в корневую директорию
    cd "$(dirname "$0")/.."
    
    # Собираем Relay Server
    log "Собираем образ cloudbridge/relay-server:latest..."
    if docker build -t cloudbridge/relay-server:latest -f Dockerfile .; then
        success "Образ cloudbridge/relay-server:latest собран"
    else
        error "Ошибка сборки образа relay-server"
    fi
    
    # Собираем AI Service
    log "Собираем образ cloudbridge/ai-service:latest..."
    if docker build -t cloudbridge/ai-service:latest -f Dockerfile.ai-service .; then
        success "Образ cloudbridge/ai-service:latest собран"
    else
        error "Ошибка сборки образа ai-service"
    fi
}

# Загрузка образов в Minikube
load_images() {
    log "Загружаем образы в Minikube..."
    
    # Загружаем Relay Server
    log "Загружаем cloudbridge/relay-server:latest..."
    if minikube image load cloudbridge/relay-server:latest; then
        success "Образ relay-server загружен в Minikube"
    else
        error "Ошибка загрузки образа relay-server"
    fi
    
    # Загружаем AI Service
    log "Загружаем cloudbridge/ai-service:latest..."
    if minikube image load cloudbridge/ai-service:latest; then
        success "Образ ai-service загружен в Minikube"
    else
        error "Ошибка загрузки образа ai-service"
    fi
    
    # Проверяем загруженные образы
    log "Проверяем загруженные образы..."
    minikube image ls | grep cloudbridge || warning "Образы не найдены в Minikube"
}

# Включение AI в конфигурации
enable_ai() {
    log "Включаем AI в конфигурации..."
    
    # Обновляем ConfigMap
    kubectl patch configmap cloudbridge-config -n cloudbridge --type merge -p '{
        "data": {
            "config.json": "{\n  \"ai\": {\n    \"enabled\": true,\n    \"behavior_analysis\": {\n      \"enabled\": true,\n      \"model_path\": \"/models/behavior\",\n      \"threshold\": 0.8,\n      \"learning_rate\": 0.01\n    },\n    \"predictive_security\": {\n      \"enabled\": true,\n      \"confidence\": 0.9\n    },\n    \"automated_response\": {\n      \"enabled\": true\n    },\n    \"monitoring\": {\n      \"enabled\": true,\n      \"dashboard_url\": \"http://grafana:3000\",\n      \"refresh_rate\": \"30s\"\n    }\n  }\n}"
        }
    }' || error "Ошибка обновления конфигурации AI"
    
    success "AI включен в конфигурации"
}

# Обновление deployment
update_deployment() {
    log "Обновляем deployment..."
    
    # Перезапускаем deployment
    log "Перезапускаем deployment cloudbridge-relay..."
    if kubectl rollout restart deployment/cloudbridge-relay -n cloudbridge; then
        success "Deployment перезапущен"
    else
        error "Ошибка перезапуска deployment"
    fi
    
    # Ждем завершения обновления
    log "Ждем завершения обновления..."
    if kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=300s; then
        success "Deployment обновлен успешно"
    else
        error "Timeout при обновлении deployment"
    fi
}

# Проверка статуса
check_status() {
    log "Проверяем статус системы..."
    
    # Проверяем поды
    log "Статус подов:"
    kubectl get pods -n cloudbridge
    
    # Проверяем инициализацию AI
    log "Проверяем инициализацию AI..."
    if kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 | grep -q "AI Analytics initialized"; then
        success "AI Analytics инициализирован"
    else
        warning "AI Analytics не найден в логах"
    fi
    
    # Проверяем сервисы
    log "Статус сервисов:"
    kubectl get services -n cloudbridge | grep -E "(NodePort|ClusterIP)"
}

# Настройка port-forwarding
setup_port_forwarding() {
    log "Настраиваем port-forwarding..."
    
    # Убиваем старые процессы
    pkill -f "kubectl port-forward" || true
    sleep 2
    
    # Запускаем port-forward для P2P API
    log "Запускаем port-forward для P2P API (порт 8082)..."
    kubectl port-forward -n cloudbridge service/p2p-api-nodeport 8082:8082 &
    sleep 2
    
    # Запускаем port-forward для AI Service
    log "Запускаем port-forward для AI Service (порт 8083)..."
    kubectl port-forward -n cloudbridge service/ai-service-nodeport 8083:8083 &
    sleep 2
    
    # Проверяем, что port-forward работает
    if ps aux | grep -q "kubectl port-forward.*8082"; then
        success "Port-forward для P2P API запущен"
    else
        warning "Port-forward для P2P API не запущен"
    fi
    
    if ps aux | grep -q "kubectl port-forward.*8083"; then
        success "Port-forward для AI Service запущен"
    else
        warning "Port-forward для AI Service не запущен"
    fi
}

# Тестирование API
test_apis() {
    log "Тестируем API endpoints..."
    
    # Ждем, пока сервисы будут готовы
    sleep 5
    
    # Тестируем P2P API
    log "Тестируем P2P API..."
    if curl -s -f http://localhost:8082/health > /dev/null; then
        success "P2P API работает"
    else
        warning "P2P API не отвечает"
    fi
    
    # Тестируем AI Service
    log "Тестируем AI Service..."
    if curl -s -f http://localhost:8083/health > /dev/null; then
        success "AI Service работает"
    else
        warning "AI Service не отвечает"
    fi
    
    # Тестируем AI Models
    log "Тестируем AI Models API..."
    if curl -s -f http://localhost:8083/api/v1/ai/models > /dev/null; then
        success "AI Models API работает"
    else
        warning "AI Models API не отвечает"
    fi
    
    # Тестируем Anomaly Detection
    log "Тестируем Anomaly Detection..."
    response=$(curl -s -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
        -H "Content-Type: application/json" \
        -d '{"metrics": {"connections": 100, "bytes_in": 1000000, "bytes_out": 500000, "latency": 50, "cpu_usage": 45, "memory_usage": 60, "packet_loss": 0.01, "protocol": "TCP"}}')
    
    if echo "$response" | grep -q "result"; then
        success "Anomaly Detection работает"
    else
        warning "Anomaly Detection не отвечает"
    fi
}

# Показать информацию о доступе
show_access_info() {
    log "Информация о доступе к сервисам:"
    echo ""
    echo "🌐 P2P API: http://localhost:8082"
    echo "   Health: http://localhost:8082/health"
    echo ""
    echo "🧠 AI Service: http://localhost:8083"
    echo "   Health: http://localhost:8083/health"
    echo "   Models: http://localhost:8083/api/v1/ai/models"
    echo "   Anomaly: http://localhost:8083/api/v1/ai/anomaly/detect"
    echo ""
    echo "📊 Grafana:"
    echo "   kubectl port-forward -n cloudbridge-monitoring service/grafana 3000:3000"
    echo "   http://localhost:3000"
    echo ""
    echo "🔍 Полезные команды:"
    echo "   kubectl get pods -n cloudbridge"
    echo "   kubectl logs -n cloudbridge deployment/cloudbridge-relay -f"
    echo "   kubectl logs -n cloudbridge deployment/ai-service -f"
    echo ""
}

# Основная функция
main() {
    echo "🚀 Начинаем развертывание CloudBridge Relay в Minikube..."
    echo ""
    
    check_requirements
    build_images
    load_images
    enable_ai
    update_deployment
    check_status
    setup_port_forwarding
    test_apis
    show_access_info
    
    echo ""
    success "🎉 Развертывание завершено успешно!"
    echo ""
    echo "Система готова к работе. AI интеграция активна."
    echo "Для остановки port-forwarding: pkill -f 'kubectl port-forward'"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --test-only    Только тестирование (без развертывания)"
        echo "  --build-only   Только сборка образов"
        echo ""
        exit 0
        ;;
    --test-only)
        log "Запускаем только тестирование..."
        setup_port_forwarding
        test_apis
        show_access_info
        exit 0
        ;;
    --build-only)
        log "Запускаем только сборку образов..."
        check_requirements
        build_images
        load_images
        exit 0
        ;;
    "")
        main
        ;;
    *)
        error "Неизвестная опция: $1. Используйте --help для справки."
        ;;
esac
