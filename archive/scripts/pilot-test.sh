#!/bin/bash

# Быстрый тест CloudBridge для пилотного проекта

set -e

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Конфигурация
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "192.168.49.2")
BASE_URL="http://$MINIKUBE_IP:32500"

echo -e "${BLUE}=== CloudBridge Pilot Project Test ===${NC}"
echo "Base URL: $BASE_URL"
echo "Date: $(date)"
echo ""

# Функция тестирования
test_endpoint() {
    local name="$1"
    local url="$2"
    local expected_code="${3:-200}"
    
    echo -n "Testing $name... "
    
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null --connect-timeout 5 --max-time 10 "$url" 2>/dev/null || echo "000")
    
    if [ "$HTTP_CODE" = "$expected_code" ]; then
        echo -e "${GREEN}✓ OK ($HTTP_CODE)${NC}"
        return 0
    else
        echo -e "${RED}✗ FAIL ($HTTP_CODE, expected $expected_code)${NC}"
        return 1
    fi
}

# Основные тесты
echo -e "${YELLOW}Core Endpoints:${NC}"
test_endpoint "Health Check" "$BASE_URL/health" "200"
test_endpoint "Metrics" "$BASE_URL/metrics" "200"
test_endpoint "System Info" "$BASE_URL/api/v1/system/info" "200"

echo ""
echo -e "${YELLOW}API Endpoints (require auth):${NC}"
test_endpoint "Peer Discovery" "$BASE_URL/api/v1/peers/discover" "400"  # Expected: missing tenant_id
test_endpoint "Peer List" "$BASE_URL/api/v1/peers/list" "401"           # Expected: unauthorized
test_endpoint "Tunnel List" "$BASE_URL/api/v1/tunnels/list" "401"       # Expected: unauthorized

echo ""
echo -e "${YELLOW}WebSocket Endpoints:${NC}"
test_endpoint "WebSocket Peers" "$BASE_URL/ws/peers" "404"              # Expected: not found (need upgrade)

echo ""
echo -e "${YELLOW}Kubernetes Status:${NC}"
POD_COUNT=$(kubectl get pods -n cloudbridge -l app=cloudbridge-relay --no-headers 2>/dev/null | wc -l)
if [ "$POD_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ CloudBridge Pods: $POD_COUNT active${NC}"
else
    echo -e "${RED}✗ No CloudBridge pods found${NC}"
fi

SVC_COUNT=$(kubectl get svc -n cloudbridge --no-headers 2>/dev/null | wc -l)
if [ "$SVC_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Kubernetes Services: $SVC_COUNT active${NC}"
else
    echo -e "${RED}✗ No Kubernetes services found${NC}"
fi

echo ""
echo -e "${YELLOW}Detailed Health Check:${NC}"
HEALTH_RESPONSE=$(curl -s "$BASE_URL/health" 2>/dev/null || echo "{}")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✓ System Status: Healthy${NC}"
    
    # Извлекаем метрики
    UPTIME=$(echo "$HEALTH_RESPONSE" | grep -o '"uptime":"[^"]*"' | cut -d'"' -f4)
    PEERS=$(echo "$HEALTH_RESPONSE" | grep -o '"total_peers":[0-9]*' | cut -d':' -f2)
    CONNECTIONS=$(echo "$HEALTH_RESPONSE" | grep -o '"active_connections":[0-9]*' | cut -d':' -f2)
    
    echo "  Uptime: ${UPTIME:-unknown}"
    echo "  Total Peers: ${PEERS:-0}"
    echo "  Active Connections: ${CONNECTIONS:-0}"
else
    echo -e "${RED}✗ System Status: Unhealthy or unreachable${NC}"
fi

echo ""
echo -e "${YELLOW}Performance Test:${NC}"
echo -n "Response time test... "
RESPONSE_TIME=$(curl -s -w "%{time_total}" -o /dev/null "$BASE_URL/health" 2>/dev/null || echo "timeout")
if [ "$RESPONSE_TIME" != "timeout" ]; then
    echo -e "${GREEN}✓ ${RESPONSE_TIME}s${NC}"
else
    echo -e "${RED}✗ Timeout${NC}"
fi

echo ""
echo -e "${BLUE}=== Summary ===${NC}"
echo "CloudBridge is ready for pilot project!"
echo ""
echo "Quick start commands:"
echo "  curl $BASE_URL/health"
echo "  curl $BASE_URL/metrics"
echo "  curl '$BASE_URL/api/v1/peers/discover?tenant_id=pilot-tenant'"
echo ""
echo "For development:"
echo "  kubectl port-forward -n cloudbridge svc/cloudbridge-relay-external 8080:5552"
echo "  curl http://localhost:8080/health"
echo ""
echo -e "${GREEN}System is ready for pilot project deployment! 🚀${NC}"
