#!/bin/bash

# Тест P2P маршрутизации в Minikube
echo "🚀 Тестирование P2P маршрутизации в Minikube"

# Получаем IP Minikube
MINIKUBE_IP=$(minikube ip)
P2P_PORT=30085

echo "Minikube IP: $MINIKUBE_IP"
echo "P2P API Port: $P2P_PORT"

# Проверяем доступность P2P API
echo "Проверка доступности P2P API..."
HEALTH_RESPONSE=$(curl -s "http://$MINIKUBE_IP:$P2P_PORT/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo "✅ P2P API доступен"
    echo "Health response: $HEALTH_RESPONSE"
else
    echo "❌ P2P API недоступен"
    echo "Response: $HEALTH_RESPONSE"
    exit 1
fi

# Создаем простой JWT токен для тестирования (без подписи)
echo "Тестирование без JWT токена..."
REGISTER_RESPONSE=$(curl -s -X POST "http://$MINIKUBE_IP:$P2P_PORT/api/v1/tenants/test-tenant/peers/register" \
    -H "Content-Type: application/json" \
    -d '{
        "public_key": "test-peer-public-key",
        "allowed_ips": ["10.100.77.10/32"],
        "peer_info": {
            "name": "Test Peer",
            "location": "Minikube Test"
        }
    }')

echo "Register response: $REGISTER_RESPONSE"

# Тестируем обнаружение пиров
echo "Тестирование обнаружения пиров..."
DISCOVER_RESPONSE=$(curl -s -X GET "http://$MINIKUBE_IP:$P2P_PORT/api/v1/tenants/test-tenant/peers/discover")
echo "Discover response: $DISCOVER_RESPONSE"

# Тестируем heartbeat
echo "Тестирование heartbeat..."
HEARTBEAT_RESPONSE=$(curl -s -X POST "http://$MINIKUBE_IP:$P2P_PORT/api/v1/tenants/test-tenant/peers/test-peer/heartbeat" \
    -H "Content-Type: application/json" \
    -d '{
        "status": "active",
        "relay_session_id": "test-session-123"
    }')
echo "Heartbeat response: $HEARTBEAT_RESPONSE"

echo "🎯 Тестирование P2P маршрутизации в Minikube завершено!"
