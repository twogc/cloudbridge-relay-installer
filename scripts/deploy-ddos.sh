#!/bin/bash

# CloudBridge Relay - DDoS Protected Deployment Script
# Скрипт для развертывания с защитой от DDoS атак

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функции для вывода
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

print_section() {
    echo -e "${CYAN}$1${NC}"
}

# Переменные
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TERRAFORM_DIR="$PROJECT_ROOT/terraform"
DDOS_CONFIG="$TERRAFORM_DIR/cloudbridge-relay-ddos.tf"

# Параметры развертывания
DDOS_LEVEL="${DDOS_LEVEL:-advanced}"  # basic, advanced, enterprise
REGION="${REGION:-MS1}"
FLAVOR="${FLAVOR:-STD2-8-16}"
DISK_SIZE="${DISK_SIZE:-120}"

print_header "CloudBridge Relay - DDoS Protected Deployment"
print_status "Уровень защиты: $DDOS_LEVEL"
print_status "Регион: $REGION"
print_status "Конфигурация: $FLAVOR"
print_status "Диск: ${DISK_SIZE}GB"

# Проверка зависимостей
check_dependencies() {
    print_section "1. Проверка зависимостей"
    
    local deps=("terraform" "curl" "jq")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Отсутствуют зависимости: ${missing_deps[*]}"
        print_status "Установите их и повторите попытку"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
}

# Проверка конфигурации
check_configuration() {
    print_section "2. Проверка конфигурации"
    
    if [ ! -f "$DDOS_CONFIG" ]; then
        print_error "Файл конфигурации DDoS не найден: $DDOS_CONFIG"
        exit 1
    fi
    
    if [ ! -f "$TERRAFORM_DIR/cloud-init-ddos.yaml" ]; then
        print_error "Cloud-init конфигурация не найдена"
        exit 1
    fi
    
    print_success "Конфигурация проверена"
}

# Инициализация Terraform
init_terraform() {
    print_section "3. Инициализация Terraform"
    
    cd "$TERRAFORM_DIR"
    
    if [ ! -d ".terraform" ]; then
        print_status "Инициализация Terraform..."
        terraform init
    else
        print_status "Terraform уже инициализирован"
    fi
    
    print_success "Terraform готов к работе"
}

# Планирование развертывания
plan_deployment() {
    print_section "4. Планирование развертывания"
    
    cd "$TERRAFORM_DIR"
    
    print_status "Создание плана развертывания..."
    
    terraform plan \
        -var="ddos_protection_level=$DDOS_LEVEL" \
        -var="availability_zone=$REGION" \
        -var="compute_flavor=$FLAVOR" \
        -var="disk_size=$DISK_SIZE" \
        -out=ddos-deployment.tfplan
    
    print_success "План создан: ddos-deployment.tfplan"
}

# Развертывание
deploy_infrastructure() {
    print_section "5. Развертывание инфраструктуры"
    
    cd "$TERRAFORM_DIR"
    
    print_status "Применение конфигурации..."
    
    terraform apply ddos-deployment.tfplan
    
    print_success "Инфраструктура развернута"
}

# Получение информации о развертывании
get_deployment_info() {
    print_section "6. Информация о развертывании"
    
    cd "$TERRAFORM_DIR"
    
    local instance_ip=$(terraform output -raw relay_ddos_instance_ip 2>/dev/null || echo "N/A")
    local instance_id=$(terraform output -raw relay_ddos_instance_id 2>/dev/null || echo "N/A")
    local loadbalancer_ip=$(terraform output -raw loadbalancer_ip 2>/dev/null || echo "N/A")
    
    print_success "Развертывание завершено!"
    echo ""
    echo "Информация о развертывании:"
    echo "   Instance IP: $instance_ip"
    echo "   Instance ID: $instance_id"
    echo "   Load Balancer: $loadbalancer_ip"
    echo ""
    echo " Следующие шаги:"
    echo "   1. SSH подключение: ssh ubuntu@$instance_ip"
    echo "   2. Проверка DDoS защиты: sudo ddos-status"
    echo "   3. Запуск relay: sudo relay-start"
    echo "   4. Мониторинг: http://$instance_ip:8081"
    echo ""
    echo "DDoS защита:"
    echo "   - L4: nftables + SYNPROXY + XDP/eBPF"
    echo "   - L7: Nginx + CrowdSec + Rate Limiting"
    echo "   - Мониторинг: FastNetMon + Grafana"
    echo ""
    echo "Доступные порты:"
    echo "   - 80: HTTP (Nginx с DDoS защитой)"
    echo "   - 443: HTTPS (Nginx с DDoS защитой)"
    echo "   - 8080: Relay Server API"
    echo "   - 8081: Relay Server Metrics"
    echo "   - 9090: QUIC Protocol (UDP)"
    echo "   - 51820: WireGuard P2P Mesh (UDP)"
}

# Тестирование развертывания
test_deployment() {
    print_section "7. Тестирование развертывания"
    
    cd "$TERRAFORM_DIR"
    
    local instance_ip=$(terraform output -raw relay_ddos_instance_ip 2>/dev/null || echo "")
    
    if [ -z "$instance_ip" ] || [ "$instance_ip" = "N/A" ]; then
        print_warning "Не удалось получить IP адрес для тестирования"
        return
    fi
    
    print_status "Тестирование доступности сервисов..."
    
    # Тест SSH
    if timeout 10 ssh -o ConnectTimeout=5 -o StrictHostKeyChecking=no ubuntu@"$instance_ip" "echo 'SSH OK'" 2>/dev/null; then
        print_success "SSH доступен"
    else
        print_warning "SSH недоступен"
    fi
    
    # Тест HTTP
    if curl -s --max-time 10 "http://$instance_ip/health" > /dev/null 2>&1; then
        print_success "HTTP доступен"
    else
        print_warning "HTTP недоступен"
    fi
    
    # Тест метрик
    if curl -s --max-time 10 "http://$instance_ip:8081/metrics" > /dev/null 2>&1; then
        print_success "Метрики доступны"
    else
        print_warning "Метрики недоступны"
    fi
}

# Основная функция
main() {
    print_header "Начинаем развертывание CloudBridge Relay с DDoS защитой"
    
    # Выполнение этапов
    check_dependencies
    check_configuration
    init_terraform
    plan_deployment
    
    # Запрос подтверждения
    echo ""
    print_warning "Проверьте план развертывания выше"
    read -p "Продолжить развертывание? (y/N): " -n 1 -r
    echo ""
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        deploy_infrastructure
        get_deployment_info
        test_deployment
        
        print_header "Развертывание завершено успешно!"
    else
        print_warning "Развертывание отменено"
        exit 0
    fi
}

# Обработка аргументов командной строки
while [[ $# -gt 0 ]]; do
    case $1 in
        --level)
            DDOS_LEVEL="$2"
            shift 2
            ;;
        --region)
            REGION="$2"
            shift 2
            ;;
        --flavor)
            FLAVOR="$2"
            shift 2
            ;;
        --disk-size)
            DISK_SIZE="$2"
            shift 2
            ;;
        --help)
            echo "Использование: $0 [опции]"
            echo ""
            echo "Опции:"
            echo "  --level LEVEL     Уровень DDoS защиты (basic/advanced/enterprise)"
            echo "  --region REGION   Регион развертывания"
            echo "  --flavor FLAVOR   Конфигурация VM"
            echo "  --disk-size SIZE  Размер диска в GB"
            echo "  --help           Показать эту справку"
            echo ""
            echo "Примеры:"
            echo "  $0 --level advanced --region MS1"
            echo "  $0 --level enterprise --flavor STD2-16-32"
            exit 0
            ;;
        *)
            print_error "Неизвестная опция: $1"
            echo "Используйте --help для справки"
            exit 1
            ;;
    esac
done

# Запуск основной функции
main "$@"

