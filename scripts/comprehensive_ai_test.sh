#!/bin/bash

# Комплексное тестирование AI сервиса CloudBridge
echo "🧪 Комплексное тестирование AI сервиса CloudBridge"
echo "=================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Счетчики
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Функция для логирования
log_test() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}PASS${NC} - $test_name: $message"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}❌ FAIL${NC} - $test_name: $message"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}  WARN${NC} - $test_name: $message"
    elif [ "$status" = "INFO" ]; then
        echo -e "${BLUE}ℹ️  INFO${NC} - $test_name: $message"
    fi
}

# Функция для проверки HTTP ответа
check_http_response() {
    local url="$1"
    local expected_status="$2"
    local test_name="$3"
    
    response=$(curl -s -w "%{http_code}" -o /tmp/response.json "$url" 2>/dev/null)
    status_code="${response: -3}"
    
    if [ "$status_code" = "$expected_status" ]; then
        log_test "$test_name" "PASS" "HTTP $status_code"
        return 0
    else
        log_test "$test_name" "FAIL" "Expected HTTP $expected_status, got $status_code"
        return 1
    fi
}

# Функция для проверки JSON ответа
check_json_response() {
    local url="$1"
    local expected_field="$2"
    local test_name="$3"
    
    response=$(curl -s "$url" 2>/dev/null)
    if echo "$response" | jq -e ".$expected_field" >/dev/null 2>&1; then
        log_test "$test_name" "PASS" "Field '$expected_field' found"
        return 0
    else
        log_test "$test_name" "FAIL" "Field '$expected_field' not found"
        return 1
    fi
}

# Функция для проверки POST запроса
check_post_response() {
    local url="$1"
    local data="$2"
    local expected_field="$3"
    local test_name="$4"
    
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$data" "$url" 2>/dev/null)
    if echo "$response" | jq -e ".$expected_field" >/dev/null 2>&1; then
        log_test "$test_name" "PASS" "POST request successful"
        return 0
    else
        log_test "$test_name" "FAIL" "POST request failed"
        return 1
    fi
}

echo -e "${CYAN}🔍 Этап 1: Проверка доступности сервисов${NC}"
echo "=============================================="

# Проверка подов
if kubectl get pods -n cloudbridge | grep -q "cloudbridge-relay.*Running"; then
    log_test "CloudBridge Relay Pod" "PASS" "Pod is running"
else
    log_test "CloudBridge Relay Pod" "FAIL" "Pod is not running"
fi

if kubectl get pods -n cloudbridge | grep -q "ai-service.*Running"; then
    log_test "AI Service Pod" "PASS" "Pod is running"
else
    log_test "AI Service Pod" "FAIL" "Pod is not running"
fi

if kubectl get pods -n cloudbridge-monitoring | grep -q "prometheus.*Running"; then
    log_test "Prometheus Pod" "PASS" "Pod is running"
else
    log_test "Prometheus Pod" "FAIL" "Pod is not running"
fi

if kubectl get pods -n cloudbridge | grep -q "cadence.*Running"; then
    log_test "Cadence Pod" "PASS" "Pod is running"
else
    log_test "Cadence Pod" "FAIL" "Pod is not running"
fi

echo -e "\n${CYAN}🔍 Этап 2: Проверка портов и сервисов${NC}"
echo "=============================================="

# Проверка сервисов
if kubectl get svc -n cloudbridge | grep -q "cloudbridge-relay"; then
    log_test "CloudBridge Relay Service" "PASS" "Service exists"
else
    log_test "CloudBridge Relay Service" "FAIL" "Service not found"
fi

if kubectl get svc -n cloudbridge | grep -q "ai-service"; then
    log_test "AI Service Service" "PASS" "Service exists"
else
    log_test "AI Service Service" "FAIL" "Service not found"
fi

if kubectl get svc -n cloudbridge-monitoring | grep -q "prometheus"; then
    log_test "Prometheus Service" "PASS" "Service exists"
else
    log_test "Prometheus Service" "FAIL" "Service not found"
fi

echo -e "\n${CYAN}🔍 Этап 3: Проверка метрик CloudBridge Relay${NC}"
echo "=================================================="

# Пробрасываем порт CloudBridge Relay
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 5551:5551 >/dev/null 2>&1 &
RELAY_PF_PID=$!
sleep 3

if check_http_response "http://localhost:5551/metrics" "200" "CloudBridge Relay Metrics"; then
    # Проверяем наличие ключевых метрик
    if curl -s "http://localhost:5551/metrics" | grep -q "relay_active_connections"; then
        log_test "Relay Active Connections Metric" "PASS" "Metric found"
    else
        log_test "Relay Active Connections Metric" "FAIL" "Metric not found"
    fi
    
    if curl -s "http://localhost:5551/metrics" | grep -q "go_goroutines"; then
        log_test "Go Goroutines Metric" "PASS" "Metric found"
    else
        log_test "Go Goroutines Metric" "FAIL" "Metric not found"
    fi
fi

echo -e "\n${CYAN}🔍 Этап 4: Проверка AI Service API${NC}"
echo "====================================="

# Пробрасываем порт AI Service
kubectl port-forward -n cloudbridge svc/ai-service 8080:8080 >/dev/null 2>&1 &
AI_PF_PID=$!
sleep 3

# Проверка health check
if check_http_response "http://localhost:8080/health" "200" "AI Service Health Check"; then
    log_test "AI Service Health" "PASS" "Service is healthy"
else
    log_test "AI Service Health" "FAIL" "Service is not healthy"
fi

# Проверка статуса AI
if check_json_response "http://localhost:8080/api/v1/ai/status" "anomaly_detector" "AI Service Status"; then
    # Проверяем отдельные компоненты
    if curl -s "http://localhost:8080/api/v1/ai/status" | jq -e '.anomaly_detector' | grep -q "true"; then
        log_test "Anomaly Detector Status" "PASS" "Component is active"
    else
        log_test "Anomaly Detector Status" "FAIL" "Component is not active"
    fi
    
    if curl -s "http://localhost:8080/api/v1/ai/status" | jq -e '.load_predictor' | grep -q "true"; then
        log_test "Load Predictor Status" "PASS" "Component is active"
    else
        log_test "Load Predictor Status" "FAIL" "Component is not active"
    fi
    
    if curl -s "http://localhost:8080/api/v1/ai/status" | jq -e '.route_optimizer' | grep -q "true"; then
        log_test "Route Optimizer Status" "PASS" "Component is active"
    else
        log_test "Route Optimizer Status" "FAIL" "Component is not active"
    fi
fi

echo -e "\n${CYAN}🔍 Этап 5: Тестирование AI функций${NC}"
echo "====================================="

# Тест обнаружения аномалий
anomaly_data='{
    "connections": 100,
    "bandwidth": 1000,
    "latency": 50,
    "cpu_usage": 75,
    "memory_usage": 60
}'

if check_post_response "http://localhost:8080/api/v1/ai/anomaly" "$anomaly_data" "is_anomaly" "Anomaly Detection"; then
    # Проверяем качество ответа
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$anomaly_data" "http://localhost:8080/api/v1/ai/anomaly")
    if echo "$response" | jq -e '.confidence' | grep -q "[0-9]"; then
        log_test "Anomaly Detection Quality" "PASS" "Confidence score provided"
    else
        log_test "Anomaly Detection Quality" "WARN" "No confidence score"
    fi
fi

# Тест прогнозирования нагрузки
load_data='{"horizon": 24}'

if check_post_response "http://localhost:8080/api/v1/ai/predict" "$load_data" "confidence" "Load Prediction"; then
    # Проверяем качество ответа
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$load_data" "http://localhost:8080/api/v1/ai/predict")
    if echo "$response" | jq -e '.model_id' | grep -q "real_data"; then
        log_test "Load Prediction Model" "PASS" "Using real data model"
    else
        log_test "Load Prediction Model" "WARN" "Model type unclear"
    fi
fi

# Тест оптимизации маршрутов
route_data='{
    "source": "node1",
    "destination": "node2",
    "constraints": {
        "max_latency": 100,
        "min_bandwidth": 1000
    }
}'

if check_post_response "http://localhost:8080/api/v1/ai/optimize" "$route_data" "protocol" "Route Optimization"; then
    # Проверяем качество ответа
    response=$(curl -s -X POST -H "Content-Type: application/json" -d "$route_data" "http://localhost:8080/api/v1/ai/optimize")
    if echo "$response" | jq -e '.path' | grep -q "node1"; then
        log_test "Route Optimization Path" "PASS" "Valid path provided"
    else
        log_test "Route Optimization Path" "WARN" "Path validation unclear"
    fi
fi

echo -e "\n${CYAN}🔍 Этап 6: Проверка Prometheus интеграции${NC}"
echo "============================================="

# Пробрасываем порт Prometheus
kubectl port-forward -n cloudbridge-monitoring svc/prometheus 9090:9090 >/dev/null 2>&1 &
PROM_PF_PID=$!
sleep 3

if check_http_response "http://localhost:9090/api/v1/query?query=up" "200" "Prometheus API"; then
    # Проверяем наличие метрик CloudBridge
    if curl -s "http://localhost:9090/api/v1/query?query=relay_active_connections" | jq -e '.data.result' | grep -q "cloudbridge"; then
        log_test "Prometheus CloudBridge Metrics" "PASS" "Metrics found"
    else
        log_test "Prometheus CloudBridge Metrics" "WARN" "Metrics not found or no data"
    fi
fi

echo -e "\n${CYAN}🔍 Этап 7: Проверка Cadence интеграции${NC}"
echo "=========================================="

# Пробрасываем порт Cadence
kubectl port-forward -n cloudbridge svc/cadence 8080:8080 >/dev/null 2>&1 &
CADENCE_PF_PID=$!
sleep 3

if check_http_response "http://localhost:8080/health" "200" "Cadence Health"; then
    log_test "Cadence Integration" "PASS" "Cadence is accessible"
else
    log_test "Cadence Integration" "WARN" "Cadence health check failed"
fi

echo -e "\n${CYAN}🔍 Этап 8: Проверка логов${NC}"
echo "============================="

# Проверка логов AI Service
ai_logs=$(kubectl logs -n cloudbridge deployment/ai-service -c ai-service --tail=10 2>/dev/null)
if echo "$ai_logs" | grep -q "error\|Error\|ERROR"; then
    log_test "AI Service Logs" "WARN" "Errors found in logs"
    echo "$ai_logs" | grep -i error | head -3
else
    log_test "AI Service Logs" "PASS" "No critical errors in logs"
fi

# Проверка логов CloudBridge Relay
relay_logs=$(kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=10 2>/dev/null)
if echo "$relay_logs" | grep -q "error\|Error\|ERROR"; then
    log_test "CloudBridge Relay Logs" "WARN" "Errors found in logs"
    echo "$relay_logs" | grep -i error | head -3
else
    log_test "CloudBridge Relay Logs" "PASS" "No critical errors in logs"
fi

echo -e "\n${CYAN}🔍 Этап 9: Проверка конфигурации${NC}"
echo "====================================="

# Проверка AI конфигурации
if kubectl get configmap ai-config -n cloudbridge >/dev/null 2>&1; then
    log_test "AI ConfigMap" "PASS" "ConfigMap exists"
    
    # Проверяем наличие ключевых настроек
    if kubectl get configmap ai-config -n cloudbridge -o yaml | grep -q "prometheus_url"; then
        log_test "Prometheus URL Config" "PASS" "Prometheus URL configured"
    else
        log_test "Prometheus URL Config" "FAIL" "Prometheus URL not configured"
    fi
    
    if kubectl get configmap ai-config -n cloudbridge -o yaml | grep -q "cloudbridge_url"; then
        log_test "CloudBridge URL Config" "PASS" "CloudBridge URL configured"
    else
        log_test "CloudBridge URL Config" "FAIL" "CloudBridge URL not configured"
    fi
else
    log_test "AI ConfigMap" "FAIL" "ConfigMap not found"
fi

# Проверка CloudBridge конфигурации
if kubectl get configmap cloudbridge-config -n cloudbridge >/dev/null 2>&1; then
    log_test "CloudBridge ConfigMap" "PASS" "ConfigMap exists"
else
    log_test "CloudBridge ConfigMap" "FAIL" "ConfigMap not found"
fi

echo -e "\n${PURPLE}📊 Итоговый отчет${NC}"
echo "=================="

echo -e "${BLUE}Всего тестов:${NC} $TOTAL_TESTS"
echo -e "${GREEN}Пройдено:${NC} $PASSED_TESTS"
echo -e "${RED}Провалено:${NC} $FAILED_TESTS"

# Расчет процента успеха
if [ $TOTAL_TESTS -gt 0 ]; then
    success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo -e "${CYAN}Процент успеха:${NC} $success_rate%"
    
    if [ $success_rate -ge 90 ]; then
        echo -e "${GREEN}🎉 Отличный результат! Система готова к продакшн использованию.${NC}"
    elif [ $success_rate -ge 75 ]; then
        echo -e "${YELLOW}  Хороший результат, но есть области для улучшения.${NC}"
    else
        echo -e "${RED}❌ Требуется доработка системы.${NC}"
    fi
fi

# Очистка фоновых процессов
kill $RELAY_PF_PID $AI_PF_PID $PROM_PF_PID $CADENCE_PF_PID 2>/dev/null

echo -e "\n${CYAN}🔧 Рекомендации по улучшению:${NC}"
echo "================================="

if [ $FAILED_TESTS -gt 0 ]; then
    echo "1. Исправить проваленные тесты"
    echo "2. Проверить конфигурацию сервисов"
    echo "3. Убедиться в правильности портов"
    echo "4. Проверить сетевую связность между сервисами"
fi

echo "5. Настроить мониторинг и алерты"
echo "6. Добавить автоматические тесты в CI/CD"
echo "7. Документировать процедуры развертывания"

echo -e "\n${GREEN}Тестирование завершено!${NC}"
