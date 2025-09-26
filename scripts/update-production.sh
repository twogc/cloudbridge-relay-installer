#!/bin/bash
set -e

# CloudBridge Relay Production Update Script
# Безопасное обновление production системы

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
NAMESPACE=${NAMESPACE:-"cloudbridge"}
DEPLOYMENT_NAME=${DEPLOYMENT_NAME:-"cloudbridge-relay"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
IMAGE_REPO=${IMAGE_REPO:-"ghcr.io/twogc/cloudbridge-relay"}
BACKUP_DIR=${BACKUP_DIR:-"./backups"}
TIMEOUT=${TIMEOUT:-"300s"}

echo -e "${BLUE}🚀 CloudBridge Relay Production Update${NC}"
echo "======================================="
echo "  Namespace: $NAMESPACE"
echo "  Deployment: $DEPLOYMENT_NAME" 
echo "  New Image: $IMAGE_REPO:$IMAGE_TAG"
echo "  Backup Dir: $BACKUP_DIR"
echo ""

# Check prerequisites
echo -e "${YELLOW}Проверка требований...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl не найден${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Не подключен к Kubernetes кластеру${NC}"
    exit 1
fi

if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    echo -e "${RED}❌ Namespace '$NAMESPACE' не найден${NC}"
    exit 1
fi

if ! kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE &> /dev/null; then
    echo -e "${RED}❌ Deployment '$DEPLOYMENT_NAME' не найден в namespace '$NAMESPACE'${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Все требования выполнены${NC}"

# Create backup directory
mkdir -p $BACKUP_DIR
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

echo -e "${YELLOW}Создание резервных копий...${NC}"

# Backup current deployment
kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o yaml > \
    "$BACKUP_DIR/deployment-$DEPLOYMENT_NAME-$TIMESTAMP.yaml"
echo -e "${GREEN}✅ Deployment сохранен: $BACKUP_DIR/deployment-$DEPLOYMENT_NAME-$TIMESTAMP.yaml${NC}"

# Backup current config
if kubectl get configmap ${DEPLOYMENT_NAME}-config -n $NAMESPACE &> /dev/null; then
    kubectl get configmap ${DEPLOYMENT_NAME}-config -n $NAMESPACE -o yaml > \
        "$BACKUP_DIR/configmap-$DEPLOYMENT_NAME-$TIMESTAMP.yaml"
    echo -e "${GREEN}✅ ConfigMap сохранен: $BACKUP_DIR/configmap-$DEPLOYMENT_NAME-$TIMESTAMP.yaml${NC}"
fi

# Backup current service
if kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE &> /dev/null; then
    kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE -o yaml > \
        "$BACKUP_DIR/service-$DEPLOYMENT_NAME-$TIMESTAMP.yaml"
    echo -e "${GREEN}✅ Service сохранен: $BACKUP_DIR/service-$DEPLOYMENT_NAME-$TIMESTAMP.yaml${NC}"
fi

# Show current status
echo ""
echo -e "${BLUE}Текущее состояние системы:${NC}"
kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME
echo ""

# Confirm update
if [[ "${AUTO_CONFIRM}" != "true" ]]; then
    echo -e "${YELLOW}Продолжить обновление до $IMAGE_REPO:$IMAGE_TAG? [y/N]${NC}"
    read -r CONFIRM
    if [[ ! $CONFIRM =~ ^[Yy]$ ]]; then
        echo "Обновление отменено"
        exit 0
    fi
fi

# Pre-update health check
echo -e "${YELLOW}Проверка здоровья системы перед обновлением...${NC}"

# Check if all pods are ready
READY_PODS=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers | grep "Running" | grep "1/1" | wc -l)
TOTAL_PODS=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers | wc -l)

if [ $READY_PODS -ne $TOTAL_PODS ]; then
    echo -e "${RED}❌ Не все поды в рабочем состоянии ($READY_PODS/$TOTAL_PODS)${NC}"
    kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME
    exit 1
fi

# Check service endpoint
SERVICE_IP=$(kubectl get service $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "")
if [ -n "$SERVICE_IP" ]; then
    if ! curl -sf http://$SERVICE_IP/health > /dev/null; then
        echo -e "${RED}❌ Health endpoint недоступен${NC}"
        exit 1
    fi
    echo -e "${GREEN}✅ Health endpoint доступен${NC}"
fi

# Perform rolling update
echo ""
echo -e "${YELLOW}Запуск rolling update...${NC}"

kubectl set image deployment/$DEPLOYMENT_NAME \
    $DEPLOYMENT_NAME=$IMAGE_REPO:$IMAGE_TAG \
    -n $NAMESPACE

echo -e "${GREEN}✅ Команда обновления отправлена${NC}"

# Monitor rollout
echo -e "${YELLOW}Мониторинг процесса обновления...${NC}"

if kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=$TIMEOUT; then
    echo -e "${GREEN}✅ Rolling update завершен успешно${NC}"
else
    echo -e "${RED}❌ Rolling update не завершился в течение $TIMEOUT${NC}"
    echo -e "${YELLOW}Откат к предыдущей версии...${NC}"
    kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE
    kubectl rollout status deployment/$DEPLOYMENT_NAME -n $NAMESPACE --timeout=$TIMEOUT
    echo -e "${RED}❌ Обновление провалено, выполнен откат${NC}"
    exit 1
fi

# Post-update verification
echo ""
echo -e "${YELLOW}Проверка после обновления...${NC}"

# Wait a bit for pods to stabilize
sleep 10

# Check pod status
echo -e "${BLUE}Состояние подов после обновления:${NC}"
kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME

# Verify all pods are ready
READY_PODS_AFTER=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers | grep "Running" | grep "1/1" | wc -l)
TOTAL_PODS_AFTER=$(kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME --no-headers | wc -l)

if [ $READY_PODS_AFTER -ne $TOTAL_PODS_AFTER ]; then
    echo -e "${RED}❌ Не все поды в рабочем состоянии после обновления ($READY_PODS_AFTER/$TOTAL_PODS_AFTER)${NC}"
    echo -e "${YELLOW}Логи подов для отладки:${NC}"
    kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE --tail=20
    exit 1
fi

# Health check after update
if [ -n "$SERVICE_IP" ]; then
    echo -e "${YELLOW}Проверка health endpoint...${NC}"
    sleep 5  # Give service time to update
    
    if curl -sf http://$SERVICE_IP/health > /dev/null; then
        echo -e "${GREEN}✅ Health endpoint работает${NC}"
    else
        echo -e "${RED}❌ Health endpoint недоступен после обновления${NC}"
        exit 1
    fi
fi

# Test basic kubectl functionality if kubeconfig exists
if [ -f ~/.kube/config-relay-production ]; then
    echo -e "${YELLOW}Тестирование kubectl proxy функциональности...${NC}"
    
    export KUBECTL_REMOTE_COMMAND_WEBSOCKETS=1
    export KUBECTL_PORT_FORWARD_WEBSOCKETS=1
    
    if kubectl --kubeconfig ~/.kube/config-relay-production get namespaces > /dev/null 2>&1; then
        echo -e "${GREEN}✅ kubectl proxy работает${NC}"
    else
        echo -e "${RED}❌ kubectl proxy не работает${NC}"
        echo "Проверьте логи и конфигурацию"
        kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE --tail=20
        exit 1
    fi
fi

# Check metrics if available
if [ -n "$SERVICE_IP" ]; then
    echo -e "${YELLOW}Проверка метрик...${NC}"
    if curl -sf http://$SERVICE_IP:5551/metrics | grep -q "kube_proxy_requests_total"; then
        echo -e "${GREEN}✅ Метрики доступны${NC}"
    else
        echo -e "${YELLOW}⚠ Метрики недоступны (может быть нормально для новых подов)${NC}"
    fi
fi

# Show final status
echo ""
echo -e "${BLUE}Финальное состояние системы:${NC}"
kubectl get deployment,pods,service -n $NAMESPACE -l app=$DEPLOYMENT_NAME

# Show image info
echo ""
echo -e "${BLUE}Версии образов:${NC}"
kubectl get deployment $DEPLOYMENT_NAME -n $NAMESPACE -o jsonpath='{.spec.template.spec.containers[0].image}'
echo ""

# Success message
echo ""
echo -e "${GREEN}🎉 Обновление завершено успешно!${NC}"
echo ""
echo -e "${BLUE}Информация об обновлении:${NC}"
echo "  ✅ Старая версия сохранена в $BACKUP_DIR/"
echo "  ✅ Новая версия: $IMAGE_REPO:$IMAGE_TAG"
echo "  ✅ Все поды в рабочем состоянии: $READY_PODS_AFTER/$TOTAL_PODS_AFTER"

if [ -n "$SERVICE_IP" ]; then
    echo "  ✅ Service доступен: http://$SERVICE_IP"
    echo "  ✅ Health endpoint: http://$SERVICE_IP/health" 
    echo "  ✅ Metrics endpoint: http://$SERVICE_IP:5551/metrics"
fi

echo ""
echo -e "${YELLOW}Для отката к предыдущей версии:${NC}"
echo "  kubectl rollout undo deployment/$DEPLOYMENT_NAME -n $NAMESPACE"
echo ""
echo -e "${YELLOW}Для мониторинга:${NC}"
echo "  kubectl logs deployment/$DEPLOYMENT_NAME -n $NAMESPACE -f"
echo "  kubectl get pods -n $NAMESPACE -l app=$DEPLOYMENT_NAME -w"

if [ -f ~/.kube/config-relay-production ]; then
    echo ""
    echo -e "${YELLOW}Для тестирования kubectl proxy:${NC}"
    echo "  kubectl --kubeconfig ~/.kube/config-relay-production get pods"
    echo "  kubectl --kubeconfig ~/.kube/config-relay-production exec -it deployment/test -- bash"
fi

echo ""
echo -e "${GREEN}Обновление production системы завершено! ✨${NC}"
