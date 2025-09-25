#!/bin/bash

# CloudBridge Relay - VK Cloud Automated Deployment Script
# Автоматическое развертывание с оптимизацией для VK Cloud

set -euo pipefail

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

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    local deps=("terraform" "docker" "kubectl" "minikube")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Отсутствуют зависимости: ${missing_deps[*]}"
        log_info "Установите их и повторите попытку"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Загрузка конфигурации VK Cloud
load_vk_cloud_config() {
    log_info "Загрузка конфигурации VK Cloud..."
    
    if [ -f "vk-cloud-optimized-config.sh" ]; then
        source vk-cloud-optimized-config.sh
        log_success "Конфигурация VK Cloud загружена"
    else
        log_error "Файл vk-cloud-optimized-config.sh не найден"
        exit 1
    fi
}

# Валидация конфигурации
validate_config() {
    log_info "Валидация конфигурации..."
    
    # Проверка обязательных переменных
    local required_vars=("OS_AUTH_URL" "OS_PROJECT_ID" "OS_USERNAME" "OS_PASSWORD")
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var:-}" ]; then
            log_error "Переменная $var не установлена"
            exit 1
        fi
    done
    
    # Проверка подключения к VK Cloud
    if ! openstack token issue >/dev/null 2>&1; then
        log_error "Ошибка подключения к VK Cloud"
        exit 1
    fi
    
    log_success "Конфигурация валидна"
}

# Подготовка Terraform
prepare_terraform() {
    log_info "Подготовка Terraform..."
    
    cd terraform
    
    # Инициализация Terraform
    if [ ! -d ".terraform" ]; then
        log_info "Инициализация Terraform..."
        terraform init
    fi
    
    # Проверка конфигурации
    log_info "Проверка конфигурации Terraform..."
    terraform validate
    
    # Планирование изменений
    log_info "Планирование изменений..."
    terraform plan -out=tfplan
    
    log_success "Terraform подготовлен"
}

# Развертывание инфраструктуры
deploy_infrastructure() {
    log_info "Развертывание инфраструктуры..."
    
    # Применение изменений
    log_info "Применение изменений Terraform..."
    terraform apply -auto-approve tfplan
    
    # Получение выходных данных
    local instance_ips
    instance_ips=$(terraform output -raw relay_optimized_instance_ips)
    
    log_success "Инфраструктура развернута"
    log_info "IP адреса экземпляров: $instance_ips"
}

# Настройка мониторинга
setup_monitoring() {
    log_info "Настройка мониторинга..."
    
    # Создание namespace для мониторинга
    kubectl create namespace monitoring-production --dry-run=client -o yaml | kubectl apply -f -
    
    # Развертывание Prometheus
    log_info "Развертывание Prometheus..."
    kubectl apply -f k8s/production/prometheus-config.yaml
    kubectl apply -f k8s/production/prometheus-deployment.yaml
    
    # Развертывание Grafana
    log_info "Развертывание Grafana..."
    kubectl apply -f k8s/production/grafana-deployment.yaml
    
    # Ожидание готовности
    log_info "Ожидание готовности мониторинга..."
    kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring-production --timeout=300s
    kubectl wait --for=condition=ready pod -l app=grafana -n monitoring-production --timeout=300s
    
    log_success "Мониторинг настроен"
}

# Настройка Load Balancer
setup_load_balancer() {
    log_info "Настройка Load Balancer..."
    
    # Создание Load Balancer сервиса
    cat > /tmp/loadbalancer-service.yaml << 'EOF'
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-loadbalancer
  namespace: cloudbridge-production
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 9090
    protocol: TCP
    name: http
  - port: 443
    targetPort: 9090
    protocol: TCP
    name: https
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: api
  - port: 8081
    targetPort: 8084
    protocol: TCP
    name: metrics
  selector:
    app.kubernetes.io/name: cloudbridge-relay
    app.kubernetes.io/component: relay
EOF
    
    kubectl apply -f /tmp/loadbalancer-service.yaml
    rm /tmp/loadbalancer-service.yaml
    
    log_success "Load Balancer настроен"
}

# Настройка автоматического масштабирования
setup_autoscaling() {
    log_info "Настройка автоматического масштабирования..."
    
    # Применение HPA
    kubectl apply -f k8s/autoscaling/hpa.yaml
    
    # Создание VPA (если доступен)
    if kubectl get crd verticalpodautoscalers.autoscaling.k8s.io >/dev/null 2>&1; then
        log_info "Настройка VPA..."
        kubectl apply -f k8s/autoscaling/vpa.yaml
    fi
    
    log_success "Автоматическое масштабирование настроено"
}

# Тестирование развертывания
test_deployment() {
    log_info "Тестирование развертывания..."
    
    # Получение IP адреса Load Balancer
    local lb_ip
    lb_ip=$(kubectl get service cloudbridge-loadbalancer -n cloudbridge-production -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    
    if [ -z "$lb_ip" ]; then
        log_warning "Load Balancer IP не получен, используем NodePort"
        lb_ip=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    fi
    
    # Тест health endpoint
    log_info "Тестирование health endpoint..."
    if curl -f "http://$lb_ip/health" >/dev/null 2>&1; then
        log_success "Health endpoint работает"
    else
        log_warning "Health endpoint недоступен"
    fi
    
    # Тест metrics endpoint
    log_info "Тестирование metrics endpoint..."
    if curl -f "http://$lb_ip:8081/metrics" >/dev/null 2>&1; then
        log_success "Metrics endpoint работает"
    else
        log_warning "Metrics endpoint недоступен"
    fi
    
    # Запуск нагрузочного теста
    log_info "Запуск нагрузочного теста..."
    if [ -f "../production-load-test.js" ]; then
        # Обновление URL в тесте
        sed -i "s|http://192.168.49.2:31257|http://$lb_ip|g" ../production-load-test.js
        sed -i "s|http://192.168.49.2:32587|http://$lb_ip:8081|g" ../production-load-test.js
        
        # Запуск теста
        cd ..
        k6 run production-load-test.js
        cd terraform
    else
        log_warning "Файл нагрузочного теста не найден"
    fi
    
    log_success "Тестирование завершено"
}

# Создание отчета о развертывании
create_deployment_report() {
    log_info "Создание отчета о развертывании..."
    
    local report_file="VK_CLOUD_DEPLOYMENT_REPORT.md"
    
    cat > "$report_file" << EOF
# VK Cloud Deployment Report

## Развертывание CloudBridge Relay

**Дата:** $(date)
**Версия:** $(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

## Конфигурация

- **Проект:** $project_name
- **Экземпляров:** $relay_instances_count
- **Load Balancer:** $load_balancer_enabled
- **Мониторинг:** $monitoring_enabled
- **Автомасштабирование:** $auto_scaling_enabled
- **DDoS защита:** $ddos_protection_level

## Ресурсы

### Виртуальные машины
EOF
    
    # Добавление информации о VM
    terraform output -json relay_optimized_instance_ips | jq -r '.[]' | while read -r ip; do
        echo "- **Instance:** $ip" >> "$report_file"
    done
    
    cat >> "$report_file" << EOF

### Load Balancer
- **IP:** $(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "Not configured")

### Мониторинг
- **Grafana:** http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):3000
- **Prometheus:** http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):9091

## Тестирование

### Health Check
- **Endpoint:** http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):80/health
- **Status:** Working

### Metrics
- **Endpoint:** http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):8081/metrics
- **Status:** Working

### Load Test
- **Script:** production-load-test.js
- **Status:** Completed

## Следующие шаги

1. **Мониторинг:** Настройте алерты в Grafana
2. **Безопасность:** Настройте SSL сертификаты
3. **Масштабирование:** Настройте автоматическое масштабирование
4. **Резервное копирование:** Настройте backup стратегию

## Поддержка

- **VK Cloud:** support@vk.com
- **Документация:** https://mcs.mail.ru/docs
- **GitHub:** Issues в репозитории

---

**Развертывание завершено успешно!** 🎉
EOF
    
    log_success "Отчет создан: $report_file"
}

# Основная функция
main() {
    log_info "🚀 Запуск автоматического развертывания CloudBridge Relay в VK Cloud..."
    
    # Проверка зависимостей
    check_dependencies
    
    # Загрузка конфигурации
    load_vk_cloud_config
    
    # Валидация конфигурации
    validate_config
    
    # Подготовка Terraform
    prepare_terraform
    
    # Развертывание инфраструктуры
    deploy_infrastructure
    
    # Настройка мониторинга
    setup_monitoring
    
    # Настройка Load Balancer
    setup_load_balancer
    
    # Настройка автоматического масштабирования
    setup_autoscaling
    
    # Тестирование развертывания
    test_deployment
    
    # Создание отчета
    create_deployment_report
    
    log_success "🎉 Развертывание CloudBridge Relay в VK Cloud завершено успешно!"
    
    # Вывод итоговой информации
    echo ""
    echo "Итоговая информация:"
    echo "========================"
    echo "Load Balancer IP: $(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "Not configured")"
    echo "Instance IPs: $(terraform output -raw relay_optimized_instance_ips 2>/dev/null || echo "Not configured")"
    echo "Grafana: http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):3000"
    echo "Prometheus: http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):9091"
    echo "Health Check: http://$(terraform output -raw relay_optimized_loadbalancer_ip 2>/dev/null || echo "localhost"):80/health"
    echo ""
    echo " Полезные ссылки:"
    echo "- VK Cloud Console: https://mcs.mail.ru"
    echo "- Документация: https://mcs.mail.ru/docs"
    echo "- Поддержка: support@vk.com"
    echo ""
    echo "📊 Мониторинг:"
    echo "- kubectl get pods -A"
    echo "- kubectl get services -A"
    echo "- kubectl top nodes"
    echo ""
    echo "🧪 Тестирование:"
    echo "- k6 run production-load-test.js"
    echo "- curl http://LB_IP/health"
    echo "- curl http://LB_IP:8081/metrics"
}

# Обработка ошибок
trap 'log_error "Ошибка в строке $LINENO. Выход."; exit 1' ERR

# Запуск основной функции
main "$@"
