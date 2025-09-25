#!/bin/bash
set -euo pipefail

# CloudBridge Relay WAF Stack Compatibility Check
# Проверяет совместимость WAF стека с существующими компонентами

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

# Счетчики тестов
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNING_TESTS=0

# Функция для выполнения теста
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_result="${3:-0}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_info "Running test: $test_name"
    
    if eval "$test_command" >/dev/null 2>&1; then
        if [ $? -eq $expected_result ]; then
            log_success "$test_name: PASSED"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            return 0
        else
            log_warning "$test_name: WARNING (unexpected result)"
            WARNING_TESTS=$((WARNING_TESTS + 1))
            return 1
        fi
    else
        log_error "$test_name: FAILED"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# Функция для проверки доступности сервиса
check_service() {
    local service_name="$1"
    local namespace="$2"
    local port="${3:-80}"
    
    log_info "Checking service: $service_name in namespace: $namespace"
    
    if kubectl get svc "$service_name" -n "$namespace" >/dev/null 2>&1; then
        log_success "Service $service_name exists in namespace $namespace"
        return 0
    else
        log_error "Service $service_name not found in namespace $namespace"
        return 1
    fi
}

# Функция для проверки endpoint'а
check_endpoint() {
    local url="$1"
    local expected_status="${2:-200}"
    local timeout="${3:-10}"
    
    log_info "Checking endpoint: $url"
    
    local status_code
    status_code=$(curl -s -o /dev/null -w "%{http_code}" --max-time "$timeout" "$url" 2>/dev/null || echo "000")
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "Endpoint $url returned status $status_code"
        return 0
    else
        log_error "Endpoint $url returned status $status_code (expected $expected_status)"
        return 1
    fi
}

# Функция для проверки ресурсов
check_resources() {
    local resource_type="$1"
    local namespace="$2"
    local min_replicas="${3:-1}"
    
    log_info "Checking $resource_type resources in namespace: $namespace"
    
    local current_replicas
    current_replicas=$(kubectl get "$resource_type" -n "$namespace" --no-headers 2>/dev/null | wc -l || echo "0")
    
    if [ "$current_replicas" -ge "$min_replicas" ]; then
        log_success "Found $current_replicas $resource_type resources (minimum: $min_replicas)"
        return 0
    else
        log_error "Found only $current_replicas $resource_type resources (minimum: $min_replicas)"
        return 1
    fi
}

# Основная функция проверки
main() {
    echo "🔍 CloudBridge Relay WAF Stack Compatibility Check"
    echo "=================================================="
    echo
    
    # Проверка 1: Kubernetes API совместимость
    echo "1. Kubernetes API Compatibility"
    echo "----------------------------------"
    run_test "Kubernetes API version" "kubectl version --short"
    run_test "Ingress API support" "kubectl api-resources | grep -q 'ingresses'"
    run_test "IngressClass API support" "kubectl api-resources | grep -q 'ingressclasses'"
    run_test "CRD API support" "kubectl api-resources | grep -q 'customresourcedefinitions'"
    echo
    
    # Проверка 2: Существующие сервисы
    echo "2. Existing Services Check"
    echo "----------------------------"
    check_service "cloudbridge-relay" "cloudbridge"
    check_service "zitadel" "zitadel"
    check_service "zitadel-login" "zitadel"
    echo
    
    # Проверка 3: Zitadel OIDC совместимость
    echo "3. Zitadel OIDC Compatibility"
    echo "-------------------------------"
    
    # Проверка доступности Zitadel
    if kubectl get pods -n zitadel | grep -q "Running"; then
        log_success "Zitadel pods are running"
        
        # Проброс порта для проверки
        log_info "Setting up port-forward for Zitadel..."
        kubectl port-forward -n zitadel service/zitadel 8080:8080 >/dev/null 2>&1 &
        PF_PID=$!
        sleep 5
        
        # Проверка OIDC endpoints
        check_endpoint "https://localhost:8080/.well-known/openid-configuration" "200" 10
        check_endpoint "https://localhost:8080/.well-known/jwks.json" "200" 10
        check_endpoint "https://localhost:8080/healthz" "200" 10
        
        # Остановка port-forward
        kill $PF_PID 2>/dev/null || true
    else
        log_error "Zitadel pods are not running"
    fi
    echo
    
    # Проверка 4: CloudBridge Relay API
    echo "4. CloudBridge Relay API Compatibility"
    echo "----------------------------------------"
    
    if kubectl get pods -n cloudbridge | grep -q "Running"; then
        log_success "CloudBridge Relay pods are running"
        
        # Проброс порта для проверки
        log_info "Setting up port-forward for CloudBridge Relay..."
        kubectl port-forward -n cloudbridge deployment/cloudbridge-relay 9090:8080 >/dev/null 2>&1 &
        PF_PID=$!
        sleep 5
        
        # Проверка API endpoints
        check_endpoint "http://localhost:9090/health" "200" 10
        check_endpoint "http://localhost:9090/metrics" "200" 10
        check_endpoint "http://localhost:9090/api/v1/status" "200" 10
        
        # Остановка port-forward
        kill $PF_PID 2>/dev/null || true
    else
        log_error "CloudBridge Relay pods are not running"
    fi
    echo
    
    # Проверка 5: Ресурсы кластера
    echo "5. Cluster Resources"
    echo "---------------------"
    run_test "Node resources" "kubectl top nodes | head -1"
    run_test "Pod resources" "kubectl top pods -A | head -1"
    run_test "Namespace resources" "kubectl get namespaces | grep -q 'cloudbridge'"
    echo
    
    # Проверка 6: Мониторинг
    echo "6. Monitoring Compatibility"
    echo "-----------------------------"
    check_service "prometheus" "monitoring" || log_warning "Prometheus not found in monitoring namespace"
    check_service "grafana" "monitoring" || log_warning "Grafana not found in monitoring namespace"
    echo
    
    # Проверка 7: Секреты
    echo "7. Secrets Management"
    echo "----------------------"
    run_test "Kubernetes secrets" "kubectl get secrets -n cloudbridge | head -1"
    run_test "Kubernetes secrets" "kubectl get secrets -n zitadel | head -1"
    echo
    
    # Проверка 8: Сетевые политики
    echo "8. Network Policies"
    echo "--------------------"
    run_test "Network policies" "kubectl get networkpolicies -A | head -1" || log_warning "No network policies found"
    echo
    
    # Проверка 9: Ingress
    echo "9. Ingress Configuration"
    echo "-------------------------"
    run_test "Ingress resources" "kubectl get ingress -A | head -1" || log_warning "No ingress resources found"
    echo
    
    # Проверка 10: Helm
    echo "10. Helm Compatibility"
    echo "-----------------------"
    run_test "Helm installation" "helm version --short"
    run_test "Helm repositories" "helm repo list | head -1"
    echo
    
    # Итоговый отчет
    echo "📊 Compatibility Check Summary"
    echo "============================="
    echo "Total tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo -e "Warnings: ${YELLOW}$WARNING_TESTS${NC}"
    echo
    
    # Расчет процента совместимости
    local compatibility_percent
    compatibility_percent=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    
    echo "Overall compatibility: $compatibility_percent%"
    echo
    
    if [ $compatibility_percent -ge 90 ]; then
        log_success "High compatibility - Ready for WAF stack deployment"
        exit 0
    elif [ $compatibility_percent -ge 70 ]; then
        log_warning " Medium compatibility - Some issues need to be resolved"
        exit 1
    else
        log_error "❌ Low compatibility - Major issues need to be resolved"
        exit 2
    fi
}

# Функция для отображения справки
show_help() {
    echo "CloudBridge Relay WAF Stack Compatibility Check"
    echo
    echo "Usage: $0 [OPTIONS]"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --verbose  Enable verbose output"
    echo "  -q, --quiet    Suppress output except errors"
    echo
    echo "Examples:"
    echo "  $0                # Run full compatibility check"
    echo "  $0 --verbose      # Run with verbose output"
    echo "  $0 --quiet        # Run quietly"
}

# Обработка аргументов командной строки
VERBOSE=false
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -q|--quiet)
            QUIET=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Запуск основной функции
main "$@"



