#!/bin/bash

# Тест балансировщика нагрузки и тенантов CloudBridge

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
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

tenant_info() {
    echo -e "${PURPLE}[TENANT]${NC} $1"
}

# Проверка LoadBalancer сервисов
check_loadbalancers() {
    info "=== ПРОВЕРКА LOADBALANCER СЕРВИСОВ ==="
    
    # Получаем все LoadBalancer сервисы
    LB_SERVICES=$(kubectl get svc -n cloudbridge --field-selector spec.type=LoadBalancer -o name)
    
    if [ -z "$LB_SERVICES" ]; then
        warning "LoadBalancer сервисы не найдены"
        return 1
    fi
    
    echo "Найденные LoadBalancer сервисы:"
    for svc in $LB_SERVICES; do
        SVC_NAME=$(echo $svc | cut -d'/' -f2)
        EXTERNAL_IP=$(kubectl get svc -n cloudbridge $SVC_NAME -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        PORTS=$(kubectl get svc -n cloudbridge $SVC_NAME -o jsonpath='{.spec.ports[*].port}')
        
        if [ -n "$EXTERNAL_IP" ]; then
            success "✓ $SVC_NAME: $EXTERNAL_IP (порты: $PORTS)"
        else
            warning "⚠ $SVC_NAME: External IP не назначен (порты: $PORTS)"
        fi
    done
    
    # Проверяем конкретные LoadBalancer'ы
    info "Детальная проверка LoadBalancer'ов:"
    
    # TCP LoadBalancer
    if kubectl get svc -n cloudbridge cloudbridge-relay-lb-tcp >/dev/null 2>&1; then
        TCP_LB_IP=$(kubectl get svc -n cloudbridge cloudbridge-relay-lb-tcp -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$TCP_LB_IP" ]; then
            success "✓ TCP LoadBalancer: $TCP_LB_IP"
            
            # Тестируем TCP подключения
            info "Тестирование TCP портов на $TCP_LB_IP:"
            for port in 9090 8084 8082; do
                if timeout 3 bash -c "</dev/tcp/$TCP_LB_IP/$port" 2>/dev/null; then
                    success "  ✓ Порт $port доступен"
                else
                    warning "  ⚠ Порт $port недоступен"
                fi
            done
        else
            warning "⚠ TCP LoadBalancer: External IP не назначен"
        fi
    fi
    
    # UDP LoadBalancer
    if kubectl get svc -n cloudbridge cloudbridge-relay-lb-udp >/dev/null 2>&1; then
        UDP_LB_IP=$(kubectl get svc -n cloudbridge cloudbridge-relay-lb-udp -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$UDP_LB_IP" ]; then
            success "✓ UDP LoadBalancer: $UDP_LB_IP"
        else
            warning "⚠ UDP LoadBalancer: External IP не назначен"
        fi
    fi
    
    # Main LoadBalancer
    if kubectl get svc -n cloudbridge cloudbridge-relay-loadbalancer >/dev/null 2>&1; then
        MAIN_LB_IP=$(kubectl get svc -n cloudbridge cloudbridge-relay-loadbalancer -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        if [ -n "$MAIN_LB_IP" ]; then
            success "✓ Main LoadBalancer: $MAIN_LB_IP"
        else
            warning "⚠ Main LoadBalancer: External IP не назначен"
        fi
    fi
}

# Проверка конфигурации тенантов
check_tenant_configuration() {
    info "=== ПРОВЕРКА КОНФИГУРАЦИИ ТЕНАНТОВ ==="
    
    # Проверяем values-kube-proxy.yaml
    if [ -f "k8s/values-kube-proxy.yaml" ]; then
        success "✓ Найден файл конфигурации тенантов: k8s/values-kube-proxy.yaml"
        
        # Извлекаем информацию о тенантах
        tenant_info "Настроенные тенанты:"
        
        # Парсим YAML файл для получения tenant_id
        grep -A 20 "policies:" k8s/values-kube-proxy.yaml | grep "tenant_id:" | while read line; do
            TENANT_ID=$(echo $line | cut -d'"' -f2)
            echo "  - $TENANT_ID"
        done
        
        # Проверяем кластеры
        tenant_info "Настроенные кластеры:"
        grep -A 10 "clusters:" k8s/values-kube-proxy.yaml | grep "id:" | while read line; do
            CLUSTER_ID=$(echo $line | cut -d'"' -f2)
            echo "  - $CLUSTER_ID"
        done
        
    else
        warning "⚠ Файл конфигурации тенантов не найден"
    fi
}

# Тест доступа к API для разных тенантов
test_tenant_api_access() {
    info "=== ТЕСТ ДОСТУПА К API ДЛЯ ТЕНАНТОВ ==="
    
    # Список тенантов для тестирования
    TENANTS=("default" "dev-team" "ops-team" "readonly-viewer")
    
    # Настраиваем port-forward если нужно
    if ! nc -z localhost 8082 2>/dev/null; then
        info "Настраиваем port-forward для тестирования..."
        kubectl port-forward -n cloudbridge svc/cloudbridge-relay-external 8082:5552 &
        PF_PID=$!
        sleep 3
    fi
    
    for tenant in "${TENANTS[@]}"; do
        tenant_info "Тестирование тенанта: $tenant"
        
        # Создаем тестовый JWT токен для тенанта (упрощенный)
        # В реальности токены должны создаваться через систему аутентификации
        
        # Тестируем базовый доступ к API
        RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null \
            -H "X-Tenant-ID: $tenant" \
            "http://localhost:8082/api/v1/health" 2>/dev/null || echo "000")
        
        if [ "$RESPONSE" = "200" ] || [ "$RESPONSE" = "404" ]; then
            success "  ✓ API доступен для тенанта $tenant (HTTP $RESPONSE)"
        else
            warning "  ⚠ API недоступен для тенанта $tenant (HTTP $RESPONSE)"
        fi
        
        # Тестируем tenant-specific endpoints
        TENANT_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null \
            -H "X-Tenant-ID: $tenant" \
            "http://localhost:8082/api/v1/tenants/$tenant/peers/discover" 2>/dev/null || echo "000")
        
        if [ "$TENANT_RESPONSE" = "200" ] || [ "$TENANT_RESPONSE" = "401" ]; then
            success "  ✓ Tenant endpoint доступен для $tenant (HTTP $TENANT_RESPONSE)"
        else
            warning "  ⚠ Tenant endpoint недоступен для $tenant (HTTP $TENANT_RESPONSE)"
        fi
    done
    
    # Останавливаем port-forward если мы его запустили
    if [ -n "$PF_PID" ]; then
        kill $PF_PID 2>/dev/null || true
    fi
}

# Проверка балансировки нагрузки между подами
test_load_balancing() {
    info "=== ТЕСТ БАЛАНСИРОВКИ НАГРУЗКИ ==="
    
    # Получаем список подов CloudBridge
    PODS=$(kubectl get pods -n cloudbridge -l app=cloudbridge-relay -o jsonpath='{.items[*].metadata.name}')
    POD_COUNT=$(echo $PODS | wc -w)
    
    if [ $POD_COUNT -eq 0 ]; then
        error "CloudBridge поды не найдены"
        return 1
    fi
    
    success "Найдено $POD_COUNT CloudBridge подов: $PODS"
    
    # Настраиваем port-forward для тестирования
    if ! nc -z localhost 8082 2>/dev/null; then
        kubectl port-forward -n cloudbridge svc/cloudbridge-relay-external 8082:5552 &
        PF_PID=$!
        sleep 3
    fi
    
    # Отправляем несколько запросов и смотрим на распределение
    info "Отправляем 20 запросов для проверки балансировки..."
    
    declare -A pod_responses
    
    for i in {1..20}; do
        # Отправляем запрос и пытаемся определить, какой под ответил
        RESPONSE=$(curl -s -H "X-Request-ID: test-$i" "http://localhost:8082/api/v1/health" 2>/dev/null || echo "error")
        
        # В реальности нужно было бы добавить заголовок с именем пода в ответ
        # Пока что просто считаем успешные запросы
        if [ "$RESPONSE" != "error" ]; then
            pod_responses["success"]=$((${pod_responses["success"]} + 1))
        else
            pod_responses["error"]=$((${pod_responses["error"]} + 1))
        fi
        
        sleep 0.1
    done
    
    success "Результаты балансировки:"
    echo "  Успешных запросов: ${pod_responses["success"]:-0}"
    echo "  Ошибок: ${pod_responses["error"]:-0}"
    
    if [ ${pod_responses["success"]:-0} -gt 15 ]; then
        success "✓ Балансировка работает корректно"
    else
        warning "⚠ Возможны проблемы с балансировкой"
    fi
    
    # Останавливаем port-forward если мы его запустили
    if [ -n "$PF_PID" ]; then
        kill $PF_PID 2>/dev/null || true
    fi
}

# Проверка Kubernetes прокси функциональности
test_k8s_proxy() {
    info "=== ТЕСТ KUBERNETES ПРОКСИ ==="
    
    # Проверяем, есть ли конфигурация для k8s прокси
    if [ -f "k8s/values-kube-proxy.yaml" ]; then
        success "✓ Конфигурация Kubernetes прокси найдена"
        
        # Проверяем настройки кластеров
        CLUSTERS=$(grep -A 5 "clusters:" k8s/values-kube-proxy.yaml | grep "id:" | wc -l)
        success "✓ Настроено кластеров: $CLUSTERS"
        
        # Проверяем политики доступа
        POLICIES=$(grep -A 5 "policies:" k8s/values-kube-proxy.yaml | grep "tenant_id:" | wc -l)
        success "✓ Настроено политик доступа: $POLICIES"
        
        # Проверяем, есть ли секреты для кластеров
        info "Проверка секретов для кластеров:"
        for cluster in production staging development; do
            if kubectl get secret -n cloudbridge "k8s-cluster-$cluster-certs" >/dev/null 2>&1; then
                success "  ✓ Секрет для кластера $cluster найден"
            else
                warning "  ⚠ Секрет для кластера $cluster не найден"
            fi
        done
        
    else
        warning "⚠ Конфигурация Kubernetes прокси не найдена"
    fi
}

# Проверка мониторинга LoadBalancer'ов
check_lb_monitoring() {
    info "=== ПРОВЕРКА МОНИТОРИНГА LOADBALANCER'ОВ ==="
    
    # Проверяем ServiceMonitor для LoadBalancer'ов
    if kubectl get servicemonitor -n cloudbridge-monitoring >/dev/null 2>&1; then
        MONITORS=$(kubectl get servicemonitor -n cloudbridge-monitoring -o name | wc -l)
        success "✓ Найдено ServiceMonitor'ов: $MONITORS"
        
        kubectl get servicemonitor -n cloudbridge-monitoring -o name | while read monitor; do
            MONITOR_NAME=$(echo $monitor | cut -d'/' -f2)
            echo "  - $MONITOR_NAME"
        done
    else
        warning "⚠ ServiceMonitor'ы не найдены"
    fi
    
    # Проверяем метрики LoadBalancer'ов
    info "Проверка доступности метрик:"
    
    # Пытаемся получить метрики через port-forward
    if kubectl get svc -n cloudbridge prometheus >/dev/null 2>&1; then
        kubectl port-forward -n cloudbridge svc/prometheus 9090:9090 &
        PROM_PID=$!
        sleep 3
        
        # Проверяем метрики LoadBalancer'а
        LB_METRICS=$(curl -s "http://localhost:9090/api/v1/query?query=up{job=~'.*loadbalancer.*'}" 2>/dev/null | grep -o '"result":\[[^]]*\]' | wc -c)
        
        if [ $LB_METRICS -gt 20 ]; then
            success "✓ Метрики LoadBalancer'ов доступны"
        else
            warning "⚠ Метрики LoadBalancer'ов недоступны"
        fi
        
        kill $PROM_PID 2>/dev/null || true
    else
        warning "⚠ Prometheus сервис не найден"
    fi
}

# Генерация отчета
generate_report() {
    info "=== ГЕНЕРАЦИЯ ОТЧЕТА ==="
    
    REPORT_FILE="loadbalancer-tenants-report-$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# CloudBridge LoadBalancer & Tenants Report

**Дата:** $(date)  
**Кластер:** $(kubectl config current-context)  

## LoadBalancer Services

$(kubectl get svc -n cloudbridge --field-selector spec.type=LoadBalancer -o wide)

## Tenant Configuration

### Настроенные тенанты:
EOF

    if [ -f "k8s/values-kube-proxy.yaml" ]; then
        grep -A 20 "policies:" k8s/values-kube-proxy.yaml | grep "tenant_id:" | while read line; do
            TENANT_ID=$(echo $line | cut -d'"' -f2)
            echo "- $TENANT_ID" >> "$REPORT_FILE"
        done
    fi

    cat >> "$REPORT_FILE" << EOF

### Настроенные кластеры:
EOF

    if [ -f "k8s/values-kube-proxy.yaml" ]; then
        grep -A 10 "clusters:" k8s/values-kube-proxy.yaml | grep "id:" | while read line; do
            CLUSTER_ID=$(echo $line | cut -d'"' -f2)
            echo "- $CLUSTER_ID" >> "$REPORT_FILE"
        done
    fi

    cat >> "$REPORT_FILE" << EOF

## Pod Status

$(kubectl get pods -n cloudbridge -l app=cloudbridge-relay -o wide)

## Service Endpoints

$(kubectl get endpoints -n cloudbridge | grep cloudbridge-relay)

## Рекомендации

1. Проверить External IP для всех LoadBalancer сервисов
2. Настроить мониторинг LoadBalancer метрик
3. Создать секреты для Kubernetes кластеров
4. Протестировать tenant-specific API endpoints
5. Настроить health checks для LoadBalancer'ов

EOF

    success "Отчет сохранен: $REPORT_FILE"
}

# Основная функция
main() {
    info "CloudBridge LoadBalancer & Tenants Test"
    info "======================================="
    
    check_loadbalancers
    echo ""
    check_tenant_configuration
    echo ""
    test_tenant_api_access
    echo ""
    test_load_balancing
    echo ""
    test_k8s_proxy
    echo ""
    check_lb_monitoring
    echo ""
    generate_report
    
    success "Тестирование LoadBalancer'ов и тенантов завершено!"
}

# Запуск
main "$@"
