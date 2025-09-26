#!/bin/bash

# CloudBridge Relay Health Check Script
# Быстрая проверка состояния системы

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
NAMESPACE=${NAMESPACE:-"cloudbridge"}
DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-"cloudbridge-relay"}
RELAY_URL=${RELAY_URL:-"http://localhost:8444"}
METRICS_URL=${METRICS_URL:-"http://localhost:5551"}
KUBECONFIG_PATH=${KUBECONFIG_PATH:-"~/.kube/config-relay"}

echo -e "${BLUE}🏥 CloudBridge Relay Health Check${NC}"
echo "=================================="
echo ""

ISSUES=0
CHECKS=0

# Health check function
check() {
    local name="$1"
    local command="$2"
    local critical="${3:-false}"
    
    CHECKS=$((CHECKS + 1))
    echo -ne "${YELLOW}Проверка: $name... ${NC}"
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ OK${NC}"
        return 0
    else
        if [ "$critical" = "true" ]; then
            echo -e "${RED}❌ КРИТИЧНО${NC}"
        else
            echo -e "${YELLOW}⚠ ВНИМАНИЕ${NC}"
        fi
        ISSUES=$((ISSUES + 1))
        return 1
    fi
}

# Kubernetes connectivity
echo -e "${BLUE}🔌 Подключение к Kubernetes${NC}"
check "kubectl подключение" "kubectl cluster-info" true
check "Namespace существует" "kubectl get namespace $NAMESPACE" true

if kubectl get namespace $NAMESPACE > /dev/null 2>&1; then
    check "Deployment существует" "kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE" true
    
    if kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE > /dev/null 2>&1; then
        # Pod status
        echo ""
        echo -e "${BLUE}🚀 Состояние подов${NC}"
        
        READY_PODS=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers 2>/dev/null | grep "Running" | grep "1/1" | wc -l)
        TOTAL_PODS=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers 2>/dev/null | wc -l)
        
        if [ $TOTAL_PODS -gt 0 ]; then
            if [ $READY_PODS -eq $TOTAL_PODS ]; then
                echo -e "${GREEN}✅ Все поды готовы ($READY_PODS/$TOTAL_PODS)${NC}"
            else
                echo -e "${RED}❌ Не все поды готовы ($READY_PODS/$TOTAL_PODS)${NC}"
                ISSUES=$((ISSUES + 1))
                kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME
            fi
        else
            echo -e "${RED}❌ Поды не найдены${NC}"
            ISSUES=$((ISSUES + 1))
        fi
        
        # Service status
        echo ""
        echo -e "${BLUE}🌐 Сервисы${NC}"
        check "Service существует" "kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE"
        
        # Get service endpoint
        SERVICE_IP=$(kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
        SERVICE_URL=""
        
        if [ -n "$SERVICE_IP" ]; then
            SERVICE_URL="http://$SERVICE_IP"
            echo -e "${GREEN}✅ LoadBalancer IP: $SERVICE_IP${NC}"
        else
            # Try NodePort
            NODE_PORT=$(kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.spec.ports[?(@.name=="api")].nodePort}' 2>/dev/null || echo "")
            if [ -n "$NODE_PORT" ]; then
                NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' 2>/dev/null || kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}' 2>/dev/null || echo "")
                if [ -n "$NODE_IP" ]; then
                    SERVICE_URL="http://$NODE_IP:$NODE_PORT"
                    echo -e "${GREEN}✅ NodePort: $NODE_IP:$NODE_PORT${NC}"
                fi
            fi
        fi
    fi
fi

# Application endpoints
echo ""
echo -e "${BLUE}🎯 HTTP Endpoints${NC}"

# Try different URL sources
TEST_URLS=()
if [ -n "$SERVICE_URL" ]; then
    TEST_URLS+=("$SERVICE_URL")
fi
TEST_URLS+=("$RELAY_URL")

for url in "${TEST_URLS[@]}"; do
    check "Health endpoint ($url)" "curl -sf $url/health"
    check "Ready endpoint ($url)" "curl -sf $url/ready"
    break  # Test only first working URL
done

# Metrics
echo ""
echo -e "${BLUE}📊 Метрики${NC}"
METRICS_URLS=()
if [ -n "$SERVICE_URL" ]; then
    METRICS_URLS+=("$SERVICE_URL:5551")
fi
METRICS_URLS+=("$METRICS_URL")

for metrics_url in "${METRICS_URLS[@]}"; do
    if check "Prometheus endpoint ($metrics_url)" "curl -sf $metrics_url/metrics"; then
        check "Kube proxy метрики" "curl -sf $metrics_url/metrics | grep -q kube_proxy_requests_total"
        check "WebSocket метрики" "curl -sf $metrics_url/metrics | grep -q kube_ws_"
        break
    fi
done

# kubectl proxy functionality
echo ""
echo -e "${BLUE}⚙️ kubectl proxy функциональность${NC}"

# Expand tilde in kubeconfig path
KUBECONFIG_EXPANDED="${KUBECONFIG_PATH/#\~/$HOME}"

if [ -f "$KUBECONFIG_EXPANDED" ]; then
    echo -e "${GREEN}✅ Kubeconfig найден: $KUBECONFIG_EXPANDED${NC}"
    
    export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
    export KUBECTL_PORT_FORWARD_WEBSOCKETS=1
    
    check "kubectl get namespaces" "kubectl --kubeconfig=$KUBECONFIG_EXPANDED get namespaces"
    check "kubectl get pods" "kubectl --kubeconfig=$KUBECONFIG_EXPANDED get pods --all-namespaces | head -5"
    
    # Test WebSocket functionality if we have a test pod
    if kubectl --kubeconfig="$KUBECONFIG_EXPANDED" get pods -o name | grep -q "pod/"; then
        TEST_POD=$(kubectl --kubeconfig="$KUBECONFIG_EXPANDED" get pods -o name | head -1)
        check "kubectl logs (WebSocket)" "timeout 10s kubectl --kubeconfig=$KUBECONFIG_EXPANDED logs $TEST_POD"
    fi
else
    echo -e "${YELLOW}⚠ Kubeconfig не найден: $KUBECONFIG_EXPANDED${NC}"
    echo "  Создайте kubeconfig для тестирования kubectl proxy"
fi

# Resource usage (if in Kubernetes)
if kubectl get namespace $NAMESPACE > /dev/null 2>&1; then
    echo ""
    echo -e "${BLUE}💾 Использование ресурсов${NC}"
    
    # CPU and Memory usage
    if kubectl top pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME > /dev/null 2>&1; then
        echo -e "${BLUE}Текущее использование ресурсов:${NC}"
        kubectl top pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME
    else
        echo -e "${YELLOW}⚠ Metrics server недоступен${NC}"
    fi
fi

# Recent logs check for errors
if kubectl get namespace $NAMESPACE > /dev/null 2>&1 && kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE > /dev/null 2>&1; then
    echo ""
    echo -e "${BLUE}📜 Проверка логов на ошибки${NC}"
    
    ERROR_COUNT=$(kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE --tail=100 --since=10m 2>/dev/null | grep -i error | wc -l)
    WARN_COUNT=$(kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE --tail=100 --since=10m 2>/dev/null | grep -i warn | wc -l)
    
    if [ $ERROR_COUNT -eq 0 ]; then
        echo -e "${GREEN}✅ Ошибки не найдены${NC}"
    else
        echo -e "${RED}❌ Найдено ошибок: $ERROR_COUNT${NC}"
        ISSUES=$((ISSUES + 1))
        echo "Последние ошибки:"
        kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE --tail=100 --since=10m | grep -i error | tail -3
    fi
    
    if [ $WARN_COUNT -gt 0 ]; then
        echo -e "${YELLOW}⚠ Найдено предупреждений: $WARN_COUNT${NC}"
    fi
fi

# Summary
echo ""
echo -e "${BLUE}📋 Итоги проверки${NC}"
echo "=================="

if [ $ISSUES -eq 0 ]; then
    echo -e "${GREEN}🎉 Все проверки пройдены успешно!${NC}"
    echo -e "${GREEN}✅ Система работает нормально${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Обнаружено проблем: $ISSUES${NC}"
    echo ""
    echo -e "${YELLOW}Рекомендации:${NC}"
    
    if [ $ISSUES -ge 3 ]; then
        echo "  1. Проверьте логи: kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE"
        echo "  2. Проверьте статус подов: kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME"
        echo "  3. Проверьте конфигурацию: kubectl describe deployment $DEPLOYMENT_NAME -n $NAMESPACE"
    fi
    
    echo "  4. Перезапустите если необходимо: kubectl rollout restart deployment/$DEPLOYMENT_NAME -n $NAMESPACE"
    echo "  5. Проверьте health endpoints: curl $RELAY_URL/health"
    
    exit 1
fi
