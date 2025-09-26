#!/bin/bash
set -e

# CloudBridge Relay Quick Test Script
# Быстрый тест kubectl proxy функциональности (5 минут)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m' 
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 CloudBridge Relay Quick Test${NC}"
echo "=================================="

# Check prerequisites
echo -e "${YELLOW}Проверка требований...${NC}"

# Check kubectl
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl не найден${NC}"
    exit 1
fi
echo -e "${GREEN}✅ kubectl найден${NC}"

# Check minikube  
if ! minikube status &> /dev/null; then
    echo -e "${RED}❌ minikube не запущен${NC}"
    echo "Запустите: minikube start --cpus=4 --memory=8192mb"
    exit 1
fi
echo -e "${GREEN}✅ minikube запущен${NC}"

# Check docker
if ! docker info &> /dev/null; then
    echo -e "${RED}❌ Docker недоступен${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Docker доступен${NC}"

# Get minikube info
MINIKUBE_IP=$(minikube ip)
API_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
MINIKUBE_PORT=$(echo $API_SERVER | sed 's/.*://')

echo -e "${BLUE}Конфигурация:${NC}"
echo "  Minikube IP: $MINIKUBE_IP"
echo "  API Server: $API_SERVER"
echo ""

# Update config file
echo -e "${YELLOW}Обновление конфигурации...${NC}"
TEMP_CONFIG="/tmp/minikube-test-config.json"
cp configs/minikube-example.json $TEMP_CONFIG
sed -i "s/192.168.49.2:8443/$MINIKUBE_IP:$MINIKUBE_PORT/g" $TEMP_CONFIG
echo -e "${GREEN}✅ Конфигурация обновлена${NC}"

# Build Docker image
echo -e "${YELLOW}Сборка Docker образа...${NC}"
eval $(minikube docker-env)
docker build -q -t cloudbridge-relay:test . 
echo -e "${GREEN}✅ Образ собран: cloudbridge-relay:test${NC}"

# Deploy to minikube
echo -e "${YELLOW}Развертывание в minikube...${NC}"
cat > /tmp/quick-deploy.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: relay-config-test
data:
  config.json: |
$(cat $TEMP_CONFIG | sed 's/^/    /')
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay-test
  labels:
    app: cloudbridge-relay-test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudbridge-relay-test
  template:
    metadata:
      labels:
        app: cloudbridge-relay-test
    spec:
      containers:
      - name: relay
        image: cloudbridge-relay:test
        imagePullPolicy: Never
        ports:
        - containerPort: 8080
        - containerPort: 8444
        - containerPort: 5551
        args: ["--config=/etc/config/config.json", "--log-level=debug"]
        volumeMounts:
        - name: config
          mountPath: /etc/config
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
      volumes:
      - name: config
        configMap:
          name: relay-config-test
---
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-test
  labels:
    app: cloudbridge-relay-test
spec:
  type: NodePort
  selector:
    app: cloudbridge-relay-test
  ports:
  - name: api
    port: 8080
    targetPort: 8080
    nodePort: 30080
  - name: grpc
    port: 8444
    targetPort: 8444  
    nodePort: 30444
  - name: metrics
    port: 5551
    targetPort: 5551
    nodePort: 30551
EOF

kubectl apply -f /tmp/quick-deploy.yaml
echo -e "${GREEN}✅ Развертывание применено${NC}"

# Wait for pod to be ready
echo -e "${YELLOW}Ожидание готовности pod...${NC}"
kubectl wait --for=condition=ready pod -l app=cloudbridge-relay-test --timeout=120s
echo -e "${GREEN}✅ Pod готов${NC}"

# Create test kubeconfig
RELAY_URL="http://$MINIKUBE_IP:30444"
METRICS_URL="http://$MINIKUBE_IP:30551"
KUBECONFIG_TEST="/tmp/kubeconfig-test"

cat > $KUBECONFIG_TEST << EOF
apiVersion: v1
kind: Config
clusters:
- name: test-via-relay
  cluster:
    server: $RELAY_URL/kube/minikube
    insecure-skip-tls-verify: true
contexts:
- name: test-via-relay
  context:
    cluster: test-via-relay
    user: test-user
current-context: test-via-relay
users:
- name: test-user
  user: {}
EOF

echo -e "${GREEN}✅ Test kubeconfig создан${NC}"
echo ""

# Wait for relay to be fully ready
echo -e "${YELLOW}Проверка доступности relay...${NC}"
for i in {1..30}; do
    if curl -s $RELAY_URL/health > /dev/null; then
        echo -e "${GREEN}✅ Relay доступен${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ Timeout ожидания готовности relay${NC}"
        kubectl logs deployment/cloudbridge-relay-test --tail=20
        exit 1
    fi
    sleep 2
done

# Set WebSocket environment variables
export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
export KUBECTL_PORT_FORWARD_WEBSOCKETS=1

echo ""
echo -e "${BLUE}🧪 Запуск тестов...${NC}"
echo "====================="

FAILED_TESTS=0
TOTAL_TESTS=0

# Test function
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -ne "${YELLOW}Тест $TOTAL_TESTS: $test_name... ${NC}"
    
    if eval "$test_command" > /tmp/test_output_$TOTAL_TESTS 2>&1; then
        echo -e "${GREEN}✅ ПРОЙДЕН${NC}"
    else
        echo -e "${RED}❌ ПРОВАЛЕН${NC}"
        echo "  Вывод: $(cat /tmp/test_output_$TOTAL_TESTS | head -1)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# Test 1: Basic connectivity
run_test "Базовая связность" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST get namespaces"

# Test 2: Pod listing
run_test "Получение списка pods" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST get pods --all-namespaces"

# Test 3: Service listing  
run_test "Получение списка services" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST get services --all-namespaces"

# Test 4: Create test pod
echo -ne "${YELLOW}Создание тестового pod... ${NC}"
kubectl --kubeconfig=$KUBECONFIG_TEST run test-pod-$$ \
    --image=busybox:1.35 --restart=Never -- sleep 300 > /dev/null 2>&1
echo -e "${GREEN}✅ Создан${NC}"

# Wait for pod to start
sleep 5

# Test 5: WebSocket exec
run_test "kubectl exec (WebSocket)" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST exec test-pod-$$ -- echo 'Hello from relay!'"

# Test 6: WebSocket logs
run_test "kubectl logs (WebSocket)" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST logs test-pod-$$"

# Test 7: Resource describe
run_test "kubectl describe" \
    "kubectl --kubeconfig=$KUBECONFIG_TEST describe pod test-pod-$$"

# Test 8: Metrics endpoint
run_test "Prometheus метрики" \
    "curl -s $METRICS_URL/metrics | grep -q kube_proxy_requests_total"

# Cleanup test pod
kubectl --kubeconfig=$KUBECONFIG_TEST delete pod test-pod-$$ --ignore-not-found=true > /dev/null 2>&1

echo ""
echo -e "${BLUE}📊 Результаты тестирования${NC}"
echo "=========================="

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 Все тесты пройдены! ($TOTAL_TESTS/$TOTAL_TESTS)${NC}"
    echo ""
    echo -e "${BLUE}Полезная информация:${NC}"
    echo "  Relay URL: $RELAY_URL"
    echo "  Metrics URL: $METRICS_URL/metrics"
    echo "  Test kubeconfig: $KUBECONFIG_TEST"
    echo ""
    echo -e "${GREEN}Система готова к использованию!${NC}"
    
    # Show some metrics
    echo ""
    echo -e "${BLUE}📈 Текущие метрики:${NC}"
    curl -s $METRICS_URL/metrics | grep -E "kube_proxy_requests_total|kube_ws_" | head -5
    
    echo ""
    echo -e "${YELLOW}Попробуйте команды:${NC}"
    echo "  kubectl --kubeconfig=$KUBECONFIG_TEST get pods"
    echo "  kubectl --kubeconfig=$KUBECONFIG_TEST run test --image=nginx --rm -it -- bash"
    echo "  curl $METRICS_URL/metrics"
    
else
    echo -e "${RED}❌ Провалено тестов: $FAILED_TESTS/$TOTAL_TESTS${NC}"
    echo ""
    echo -e "${YELLOW}Логи relay для отладки:${NC}"
    kubectl logs deployment/cloudbridge-relay-test --tail=20
    exit 1
fi

echo ""
echo -e "${BLUE}Для полных E2E тестов запустите:${NC}"
echo "  export RELAY_URL=$RELAY_URL"
echo "  export METRICS_URL=$METRICS_URL"
echo "  ./test/e2e/run_tests.sh"

echo ""
echo -e "${YELLOW}Для cleanup:${NC}"
echo "  kubectl delete deployment,service,configmap -l app=cloudbridge-relay-test"
