#!/usr/bin/env bash

# 🐳 Установка Docker Registry в VK Cloud
# Автор: VK Cloud API Manager
# Версия: 1.0.0

set -euo pipefail

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Функции логирования
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Загружаем переменные окружения
source ./scripts/setup-env.sh

# Конфигурация
REGISTRY_IP="${REGISTRY_IP:-212.233.89.183}"
REGISTRY_NAME="${REGISTRY_NAME:-cloudbridge-registry}"
REGISTRY_NAMESPACE="${REGISTRY_NAMESPACE:-kube-system}"
CONFIG_FILE="configs/docker-registry-vkcloud.yaml"

# Функция для проверки зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    local deps=("kubectl" "helm" "curl")
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            log_error "Отсутствует зависимость: $dep"
            exit 1
        fi
    done
    log_success "Все зависимости установлены"
}

# Функция для проверки доступа к Kubernetes
check_k8s_access() {
    log_info "Проверка доступа к Kubernetes..."
    if kubectl --insecure-skip-tls-verify cluster-info &> /dev/null; then
        log_success "Доступ к Kubernetes подтвержден"
    else
        log_error "Нет доступа к Kubernetes"
        exit 1
    fi
}

# Функция для создания плавающего IP
create_floating_ip() {
    log_info "Создание плавающего IP..."
    
    # Используем VK Cloud API для создания плавающего IP
    local ip_response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{"floatingip": {"floating_network_id": "external"}}' \
        "$NEUTRON_URL/v2.0/floatingips")
    
    local ip_id=$(echo "$ip_response" | jq -r '.floatingip.id')
    local ip_address=$(echo "$ip_response" | jq -r '.floatingip.floating_ip_address')
    
    if [ "$ip_id" != "null" ] && [ "$ip_address" != "null" ]; then
        log_success "Плавающий IP создан: $ip_address (ID: $ip_id)"
        echo "$ip_address" > .registry_ip
        REGISTRY_IP="$ip_address"
    else
        log_error "Ошибка создания плавающего IP"
        exit 1
    fi
}

# Функция для обновления конфигурации
update_config() {
    log_info "Обновление конфигурации registry..."
    
    if [ ! -f "$CONFIG_FILE" ]; then
        log_error "Конфигурационный файл не найден: $CONFIG_FILE"
        exit 1
    fi
    
    # Обновляем IP адрес в конфигурации
    sed -i "s/loadBalancerIP:.*/loadBalancerIP: \"$REGISTRY_IP\"/g" "$CONFIG_FILE"
    sed -i "s/service.beta.kubernetes.io\/vk-cs-load-balancer-ip:.*/service.beta.kubernetes.io\/vk-cs-load-balancer-ip: \"$REGISTRY_IP\"/g" "$CONFIG_FILE"
    
    log_success "Конфигурация обновлена с IP: $REGISTRY_IP"
}

# Функция для установки registry через Helm
install_registry() {
    log_info "Установка Docker Registry через Helm..."
    
    # Добавляем репозиторий Helm
    helm repo add twuni https://helm.twun.io
    helm repo update
    
    # Устанавливаем registry
    if helm install "$REGISTRY_NAME" twuni/docker-registry \
        --namespace "$REGISTRY_NAMESPACE" \
        --values "$CONFIG_FILE" \
        --wait --timeout 10m; then
        log_success "Docker Registry установлен"
    else
        log_error "Ошибка установки Docker Registry"
        exit 1
    fi
}

# Функция для проверки установки
verify_installation() {
    log_info "Проверка установки registry..."
    
    # Ждем запуска пода
    sleep 30
    
    # Проверяем поды
    if kubectl --insecure-skip-tls-verify get pods -n "$REGISTRY_NAMESPACE" | grep -q "$REGISTRY_NAME"; then
        log_success "Поды registry запущены"
    else
        log_error "Поды registry не найдены"
        exit 1
    fi
    
    # Проверяем сервис
    if kubectl --insecure-skip-tls-verify get svc -n "$REGISTRY_NAMESPACE" | grep -q "$REGISTRY_NAME"; then
        log_success "Сервис registry создан"
    else
        log_error "Сервис registry не найден"
        exit 1
    fi
    
    # Проверяем доступность
    if curl -s --connect-timeout 10 "http://$REGISTRY_IP:5000/v2/" > /dev/null; then
        log_success "Registry доступен по адресу: $REGISTRY_IP:5000"
    else
        log_warning "Registry пока недоступен, подождите несколько минут"
    fi
}

# Функция для создания namespace
create_namespace() {
    log_info "Создание namespace: $REGISTRY_NAMESPACE"
    
    if kubectl --insecure-skip-tls-verify get namespace "$REGISTRY_NAMESPACE" &> /dev/null; then
        log_info "Namespace уже существует"
    else
        kubectl --insecure-skip-tls-verify create namespace "$REGISTRY_NAMESPACE"
        log_success "Namespace создан"
    fi
}

# Функция для настройки доступа
setup_access() {
    log_info "Настройка доступа к registry..."
    
    # Создаем секрет для доступа
    kubectl --insecure-skip-tls-verify create secret docker-registry cloudbridge-registry-secret \
        --docker-server="$REGISTRY_IP:5000" \
        --docker-username=cloudbridge \
        --docker-password=password \
        --docker-email=info@2gc.ru \
        -n cloudbridge --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Секрет доступа создан"
}

# Функция для вывода информации
print_info() {
    echo ""
    log_success "Docker Registry успешно установлен!"
    echo ""
    echo "Информация о registry:"
    echo "   IP адрес: $REGISTRY_IP"
    echo "   Порт: 5000"
    echo "   URL: http://$REGISTRY_IP:5000"
    echo "   Пользователь: cloudbridge"
    echo "   Пароль: password"
    echo ""
    echo " Команды для работы:"
    echo "   # Вход в registry"
    echo "   docker login $REGISTRY_IP:5000 --username cloudbridge --password password"
    echo ""
    echo "   # Тегирование образа"
    echo "   docker tag localhost/myapp:latest $REGISTRY_IP:5000/cloudbridge/myapp:latest"
    echo ""
    echo "   # Отправка образа"
    echo "   docker push $REGISTRY_IP:5000/cloudbridge/myapp:latest"
    echo ""
    echo "   # Проверка образов"
    echo "   curl -u cloudbridge:password http://$REGISTRY_IP:5000/v2/_catalog"
    echo ""
    echo "Для автоматизации используйте:"
    echo "   ./scripts/docker-registry-automation.sh"
    echo ""
}

# Главная функция
main() {
    log_info "Установка Docker Registry в VK Cloud"
    
    check_dependencies
    check_k8s_access
    create_namespace
    
    # Создаем плавающий IP если не указан
    if [ "$REGISTRY_IP" = "212.233.89.183" ]; then
        log_info "Используем существующий IP: $REGISTRY_IP"
    fi
    
    update_config
    install_registry
    verify_installation
    setup_access
    print_info
}

# Обработка аргументов
case "${1:-}" in
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  install  - Установить registry (по умолчанию)"
        echo "  verify   - Проверить установку"
        echo "  uninstall - Удалить registry"
        echo "  help     - Показать эту справку"
        echo ""
        echo "Переменные окружения:"
        echo "  REGISTRY_IP       - IP адрес registry"
        echo "  REGISTRY_NAME     - Имя установки (по умолчанию: cloudbridge-registry)"
        echo "  REGISTRY_NAMESPACE - Namespace (по умолчанию: kube-system)"
        ;;
    "verify")
        verify_installation
        ;;
    "uninstall")
        log_info "Удаление Docker Registry..."
        helm uninstall "$REGISTRY_NAME" -n "$REGISTRY_NAMESPACE"
        log_success "Docker Registry удален"
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
