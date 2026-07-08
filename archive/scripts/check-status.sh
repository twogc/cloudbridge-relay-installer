#!/bin/bash
# CloudBridge Relay - Проверка статуса

echo "Проверка статуса CloudBridge Relay..."
echo ""

# Проверка подов
echo "Статус подов:"
kubectl get pods -n cloudbridge -o wide
echo ""

# Проверка сервисов
echo "Доступные сервисы:"
kubectl get svc -n cloudbridge
echo ""

# Проверка NodePort сервисов
echo "NodePort сервисы:"
kubectl get svc -n cloudbridge | grep NodePort
echo ""

# Проверка LoadBalancer сервисов
echo "⚖️ LoadBalancer сервисы:"
kubectl get svc -n cloudbridge | grep LoadBalancer
echo ""

# Проверка ресурсов
echo "💾 Использование ресурсов:"
kubectl top pods -n cloudbridge 2>/dev/null || echo "Метрики недоступны"
echo ""

# Проверка логов (последние 5 строк)
echo "📝 Последние логи CloudBridge Relay:"
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=5 2>/dev/null || echo "Логи недоступны"
echo ""

# Проверка логов AI Service
echo " Последние логи AI Service:"
kubectl logs -n cloudbridge deployment/ai-service --tail=5 2>/dev/null || echo "Логи недоступны"
echo ""

# Проверка Cassandra
echo "🗄️ Статус Cassandra:"
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "describe keyspaces" 2>/dev/null || echo "Cassandra недоступна"
echo ""

# Проверка TURN Pool
echo "🔄 Проверка TURN Pool:"
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i "turn\|pool" | tail -3 || echo "TURN Pool логи не найдены"
echo ""

# Команды для доступа
echo "Команды для доступа:"
echo "  CloudBridge Relay: kubectl port-forward svc/cloudbridge-relay 9090:9090 -n cloudbridge"
echo "  AI Service:        kubectl port-forward svc/ai-service 8083:8083 -n cloudbridge"
echo "  Cassandra:         kubectl port-forward svc/cassandra-service 9042:9042 -n cloudbridge"
echo ""

# IP Minikube
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "недоступен")
echo "Minikube IP: $MINIKUBE_IP"
if [ "$MINIKUBE_IP" != "недоступен" ]; then
    echo "  CloudBridge Relay: http://$MINIKUBE_IP:30990/health"
    echo "  AI Service:        http://$MINIKUBE_IP:30083/health"
fi
echo ""

echo "Проверка завершена!"
