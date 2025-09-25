#!/usr/bin/env bash

# 🐳 Автоматизация сборки и отправки Docker образов в VK Cloud Registry
# Автор: VK Cloud API Manager
# Версия: 1.0.0

set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции логирования
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

# Загружаем переменные окружения
source ./scripts/setup-env.sh

# Конфигурация
REGISTRY_IP="${REGISTRY_IP:-localhost}"
REGISTRY_PORT="${REGISTRY_PORT:-5000}"
REGISTRY_URL="${REGISTRY_IP}:${REGISTRY_PORT}"
REGISTRY_USERNAME="${REGISTRY_USERNAME:-cloudbridge}"
REGISTRY_PASSWORD="${REGISTRY_PASSWORD:-password}"
REGISTRY_NAMESPACE="${REGISTRY_NAMESPACE:-cloudbridge}"

# Версии образов
VERSION="${VERSION:-$(git describe --tags --always --dirty)}"
BUILD_TIME="${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse HEAD)}"

# Список сервисов для сборки
SERVICES=(
    "relay"
    "ai"
    "cadence"
    "quic"
    "wireguard"
)

# Функция для проверки зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("docker" "kubectl" "curl" "jq")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Отсутствуют зависимости: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Функция для проверки подключения к registry
check_registry_connection() {
    log_info "Проверка подключения к registry..."
    
    if ! curl -s --connect-timeout 10 "http://${REGISTRY_URL}/v2/" > /dev/null; then
        log_error "Не удается подключиться к registry по адресу ${REGISTRY_URL}"
        log_info "Убедитесь, что registry запущен и доступен"
        return 1
    fi
    
    log_success "Registry доступен"
}

# Функция для входа в registry
login_to_registry() {
    log_info "Вход в registry..."
    
    if echo "${REGISTRY_PASSWORD}" | docker login "${REGISTRY_URL}" --username "${REGISTRY_USERNAME}" --password-stdin; then
        log_success "Успешный вход в registry"
    else
        log_error "Ошибка входа в registry"
        return 1
    fi
}

# Функция для сборки образа
build_image() {
    local service=$1
    local dockerfile="Dockerfile.${service}"
    local image_name="${REGISTRY_URL}/${REGISTRY_NAMESPACE}/${service}"
    local full_tag="${image_name}:${VERSION}"
    local latest_tag="${image_name}:latest"
    
    log_info "Сборка образа для сервиса: ${service}"
    
    # Проверяем существование Dockerfile
    if [ ! -f "$dockerfile" ]; then
        log_warning "Dockerfile не найден: ${dockerfile}, пропускаем"
        return 0
    fi
    
    # Собираем образ
    if docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg GIT_COMMIT="${GIT_COMMIT}" \
        -f "$dockerfile" \
        -t "$full_tag" \
        -t "$latest_tag" \
        .; then
        log_success "Образ собран: ${full_tag}"
    else
        log_error "Ошибка сборки образа: ${service}"
        return 1
    fi
}

# Функция для отправки образа в registry
push_image() {
    local service=$1
    local image_name="${REGISTRY_URL}/${REGISTRY_NAMESPACE}/${service}"
    local full_tag="${image_name}:${VERSION}"
    local latest_tag="${image_name}:latest"
    
    log_info "Отправка образа в registry: ${service}"
    
    # Отправляем версионированный тег
    if docker push "$full_tag"; then
        log_success "Образ отправлен: ${full_tag}"
    else
        log_error "Ошибка отправки образа: ${full_tag}"
        return 1
    fi
    
    # Отправляем latest тег
    if docker push "$latest_tag"; then
        log_success "Образ отправлен: ${latest_tag}"
    else
        log_error "Ошибка отправки образа: ${latest_tag}"
        return 1
    fi
}

# Функция для создания секрета Kubernetes
create_k8s_secret() {
    log_info "Создание секрета для Kubernetes..."
    
    # Проверяем, существует ли секрет
    if kubectl --insecure-skip-tls-verify get secret "${REGISTRY_NAMESPACE}-registry-secret" -n "${REGISTRY_NAMESPACE}" &> /dev/null; then
        log_info "Секрет уже существует, обновляем..."
        kubectl --insecure-skip-tls-verify delete secret "${REGISTRY_NAMESPACE}-registry-secret" -n "${REGISTRY_NAMESPACE}"
    fi
    
    # Создаем секрет
    if kubectl --insecure-skip-tls-verify create secret docker-registry "${REGISTRY_NAMESPACE}-registry-secret" \
        --docker-server="${REGISTRY_URL}" \
        --docker-username="${REGISTRY_USERNAME}" \
        --docker-password="${REGISTRY_PASSWORD}" \
        --docker-email="info@2gc.ru" \
        -n "${REGISTRY_NAMESPACE}"; then
        log_success "Секрет создан: ${REGISTRY_NAMESPACE}-registry-secret"
    else
        log_error "Ошибка создания секрета"
        return 1
    fi
}

# Функция для обновления deployment файлов
update_deployments() {
    log_info "Обновление deployment файлов..."
    
    local deployment_dir="k8s"
    local image_prefix="${REGISTRY_URL}/${REGISTRY_NAMESPACE}"
    
    # Обновляем каждый deployment файл
    for service in "${SERVICES[@]}"; do
        local deployment_file="${deployment_dir}/${service}-deployment.yaml"
        
        if [ -f "$deployment_file" ]; then
            log_info "Обновление deployment: ${deployment_file}"
            
            # Создаем временный файл
            local temp_file=$(mktemp)
            
            # Обновляем образ и добавляем imagePullSecrets
            sed -e "s|image:.*${service}.*|image: ${image_prefix}/${service}:${VERSION}|g" \
                -e "/imagePullSecrets:/,/name: ${REGISTRY_NAMESPACE}-registry-secret/d" \
                -e "/containers:/a\      imagePullSecrets:\n      - name: ${REGISTRY_NAMESPACE}-registry-secret" \
                "$deployment_file" > "$temp_file"
            
            # Заменяем оригинальный файл
            mv "$temp_file" "$deployment_file"
            
            log_success "Deployment обновлен: ${deployment_file}"
        else
            log_warning "Deployment файл не найден: ${deployment_file}"
        fi
    done
}

# Функция для проверки образов в registry
verify_images() {
    log_info "Проверка образов в registry..."
    
    for service in "${SERVICES[@]}"; do
        local catalog_url="http://${REGISTRY_URL}/v2/${REGISTRY_NAMESPACE}/${service}/tags/list"
        
        if curl -s -u "${REGISTRY_USERNAME}:${REGISTRY_PASSWORD}" "$catalog_url" | jq -e '.tags' > /dev/null; then
            log_success "Образ ${service} найден в registry"
        else
            log_warning "Образ ${service} не найден в registry"
        fi
    done
}

# Функция для развертывания в Kubernetes
deploy_to_k8s() {
    log_info "Развертывание в Kubernetes..."
    
    local deployment_dir="k8s"
    
    # Применяем все deployment файлы
    for service in "${SERVICES[@]}"; do
        local deployment_file="${deployment_dir}/${service}-deployment.yaml"
        
        if [ -f "$deployment_file" ]; then
            log_info "Применение deployment: ${service}"
            
            if kubectl --insecure-skip-tls-verify apply -f "$deployment_file"; then
                log_success "Deployment применен: ${service}"
            else
                log_error "Ошибка применения deployment: ${service}"
                return 1
            fi
        fi
    done
}

# Функция для проверки статуса развертывания
check_deployment_status() {
    log_info "Проверка статуса развертывания..."
    
    # Ждем немного для запуска подов
    sleep 10
    
    for service in "${SERVICES[@]}"; do
        log_info "Статус сервиса: ${service}"
        
        if kubectl --insecure-skip-tls-verify get pods -n "${REGISTRY_NAMESPACE}" -l app="${service}" 2>/dev/null; then
            log_success "Поды найдены для сервиса: ${service}"
        else
            log_warning "Поды не найдены для сервиса: ${service}"
        fi
    done
}

# Функция для очистки локальных образов
cleanup_local_images() {
    log_info "Очистка локальных образов..."
    
    for service in "${SERVICES[@]}"; do
        local image_name="${REGISTRY_URL}/${REGISTRY_NAMESPACE}/${service}"
        local full_tag="${image_name}:${VERSION}"
        local latest_tag="${image_name}:latest"
        
        # Удаляем образы
        docker rmi "$full_tag" 2>/dev/null || true
        docker rmi "$latest_tag" 2>/dev/null || true
    done
    
    log_success "Локальные образы очищены"
}

# Главная функция
main() {
    log_info "🚀 Запуск автоматизации сборки и отправки Docker образов"
    log_info "Registry: ${REGISTRY_URL}"
    log_info "Версия: ${VERSION}"
    log_info "Сервисы: ${SERVICES[*]}"
    
    # Проверяем зависимости
    check_dependencies
    
    # Проверяем подключение к registry
    if ! check_registry_connection; then
        exit 1
    fi
    
    # Входим в registry
    if ! login_to_registry; then
        exit 1
    fi
    
    # Собираем и отправляем образы
    for service in "${SERVICES[@]}"; do
        if build_image "$service"; then
            if push_image "$service"; then
                log_success "Сервис ${service} успешно обработан"
            else
                log_error "❌ Ошибка отправки сервиса ${service}"
                exit 1
            fi
        else
            log_warning " Сервис ${service} пропущен"
        fi
    done
    
    # Создаем секрет Kubernetes
    if create_k8s_secret; then
        # Обновляем deployment файлы
        update_deployments
        
        # Проверяем образы в registry
        verify_images
        
        # Развертываем в Kubernetes
        if deploy_to_k8s; then
            # Проверяем статус развертывания
            check_deployment_status
        fi
    fi
    
    # Очищаем локальные образы
    cleanup_local_images
    
    log_success "🎉 Автоматизация завершена успешно!"
}

# Обработка аргументов командной строки
case "${1:-}" in
    "build")
        log_info "Режим: только сборка"
        check_dependencies
        for service in "${SERVICES[@]}"; do
            build_image "$service"
        done
        ;;
    "push")
        log_info "Режим: только отправка"
        check_dependencies
        check_registry_connection
        login_to_registry
        for service in "${SERVICES[@]}"; do
            push_image "$service"
        done
        ;;
    "deploy")
        log_info "Режим: только развертывание"
        create_k8s_secret
        update_deployments
        deploy_to_k8s
        check_deployment_status
        ;;
    "cleanup")
        log_info "Режим: очистка"
        cleanup_local_images
        ;;
    "verify")
        log_info "Режим: проверка"
        check_registry_connection
        verify_images
        ;;
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  build    - Только сборка образов"
        echo "  push     - Только отправка образов"
        echo "  deploy   - Только развертывание в Kubernetes"
        echo "  cleanup  - Очистка локальных образов"
        echo "  verify   - Проверка образов в registry"
        echo "  help     - Показать эту справку"
        echo ""
        echo "Переменные окружения:"
        echo "  REGISTRY_IP       - IP адрес registry (по умолчанию: 95.163.250.128)"
        echo "  REGISTRY_PORT     - Порт registry (по умолчанию: 5000)"
        echo "  REGISTRY_USERNAME - Имя пользователя (по умолчанию: cloudbridge)"
        echo "  REGISTRY_PASSWORD - Пароль (по умолчанию: password)"
        echo "  VERSION           - Версия образов (по умолчанию: git describe)"
        echo ""
        echo "Примеры:"
        echo "  $0                    # Полный цикл"
        echo "  $0 build              # Только сборка"
        echo "  $0 push               # Только отправка"
        echo "  REGISTRY_IP=192.168.1.100 $0  # С кастомным IP"
        ;;
    "")
        main
        ;;
    *)
        log_error "Неизвестная команда: $1"
        echo "Используйте '$0 help' для справки"
        exit 1
        ;;
esac
