#!/bin/bash

# Тест P2P API с валидным токеном
echo "🚀 Тестирование P2P API с валидным токеном"

# Читаем токен из файла
JWT_TOKEN=$(cat /home/ubuntu/cloudbridge-relay-installer/token.txt)
echo "JWT токен загружен: ${JWT_TOKEN:0:50}..."

# Тестируем регистрацию пира
echo "1. Тестирование регистрации пира..."
REGISTER_RESPONSE=$(curl -s -X POST "http://localhost:8083/api/v1/tenants/tenant-216420165/peers/register" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "public_key": "test-peer-public-key-$(date +%s)",
        "allowed_ips": ["10.100.77.11/32"],
        "peer_info": {
            "name": "Test Peer $(date +%s)",
            "location": "Minikube Test"
        }
    }')

echo "Register response: $REGISTER_RESPONSE"

# Извлекаем peer_id из ответа
PEER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"peer_id":"[^"]*"' | cut -d'"' -f4)
echo "Peer ID: $PEER_ID"

# Тестируем обнаружение пиров
echo "2. Тестирование обнаружения пиров..."
DISCOVER_RESPONSE=$(curl -s -X GET "http://localhost:8083/api/v1/tenants/tenant-216420165/peers/discover" \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "Discover response: $DISCOVER_RESPONSE"

# Тестируем heartbeat
echo "3. Тестирование heartbeat..."
HEARTBEAT_RESPONSE=$(curl -s -X POST "http://localhost:8083/api/v1/tenants/tenant-216420165/peers/$PEER_ID/heartbeat" \
    -H "Authorization: Bearer $JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "status": "active",
        "relay_session_id": "test-session-123"
    }')

echo "Heartbeat response: $HEARTBEAT_RESPONSE"

# Тестируем получение информации о пире
echo "4. Тестирование получения информации о пире..."
INFO_RESPONSE=$(curl -s -X GET "http://localhost:8083/api/v1/tenants/tenant-216420165/peers/$PEER_ID/info" \
    -H "Authorization: Bearer $JWT_TOKEN")

echo "Info response: $INFO_RESPONSE"

echo "🎯 Тестирование P2P API с валидным токеном завершено!"
