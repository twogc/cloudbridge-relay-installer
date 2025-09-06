#!/bin/bash

# Скрипт для тестирования интеграции Cadence

set -e

# Конфигурация
RELAY_HOST="localhost:9092"
CADENCE_HOST="localhost:7933"
DOMAIN_NAME="cloudbridge"

echo "Testing Cadence integration..."

# Проверяем доступность relay сервера
echo "1. Checking relay server availability..."
if ! curl -s -f http://$RELAY_HOST/health > /dev/null; then
    echo "Error: Relay server is not running on $RELAY_HOST"
    echo "Please start the relay server first"
    exit 1
fi
echo "✓ Relay server is running"

# Проверяем доступность Cadence сервера
echo "2. Checking Cadence server availability..."
if ! nc -z localhost 7933; then
    echo "Error: Cadence server is not running on localhost:7933"
    echo "Please start Cadence server first using: docker-compose up -d cadence cassandra"
    exit 1
fi
echo "✓ Cadence server is running"

# Проверяем Cadence health endpoint
echo "3. Testing Cadence health endpoint..."
if curl -s -f http://$RELAY_HOST/api/v1/cadence/health > /dev/null; then
    echo "✓ Cadence health endpoint is working"
else
    echo "✗ Cadence health endpoint is not working"
    exit 1
fi

# Тестируем запуск workflow развертывания
echo "4. Testing relay deployment workflow..."
DEPLOYMENT_RESPONSE=$(curl -s -X POST http://$RELAY_HOST/api/v1/cadence/workflows/relay-deployment \
    -H "Content-Type: application/json" \
    -d '{
        "config": {"test": true},
        "wireguard_config": {"enabled": true},
        "quic_config": {"enabled": true},
        "services": ["relay", "wireguard", "quic"],
        "timeout": "5m"
    }')

if echo "$DEPLOYMENT_RESPONSE" | grep -q "workflow_id"; then
    WORKFLOW_ID=$(echo "$DEPLOYMENT_RESPONSE" | jq -r '.workflow_id')
    echo "✓ Relay deployment workflow started with ID: $WORKFLOW_ID"
else
    echo "✗ Failed to start relay deployment workflow"
    echo "Response: $DEPLOYMENT_RESPONSE"
    exit 1
fi

# Проверяем статус workflow
echo "5. Checking workflow status..."
sleep 2
STATUS_RESPONSE=$(curl -s "http://$RELAY_HOST/api/v1/cadence/workflows/status?workflow_id=$WORKFLOW_ID")

if echo "$STATUS_RESPONSE" | grep -q "status"; then
    STATUS=$(echo "$STATUS_RESPONSE" | jq -r '.status')
    echo "✓ Workflow status retrieved: $STATUS"
else
    echo "✗ Failed to get workflow status"
    echo "Response: $STATUS_RESPONSE"
    exit 1
fi

# Тестируем запуск workflow резервного копирования
echo "6. Testing backup workflow..."
BACKUP_RESPONSE=$(curl -s -X POST http://$RELAY_HOST/api/v1/cadence/workflows/backup \
    -H "Content-Type: application/json" \
    -d '{
        "backup_config": {"type": "full"},
        "config_paths": ["/etc/relay", "/var/lib/relay"],
        "data_paths": ["/var/log/relay"],
        "cleanup_config": {"retention_days": 7}
    }')

if echo "$BACKUP_RESPONSE" | grep -q "workflow_id"; then
    BACKUP_WORKFLOW_ID=$(echo "$BACKUP_RESPONSE" | jq -r '.workflow_id')
    echo "✓ Backup workflow started with ID: $BACKUP_WORKFLOW_ID"
else
    echo "✗ Failed to start backup workflow"
    echo "Response: $BACKUP_RESPONSE"
    exit 1
fi

echo ""
echo "🎉 All Cadence integration tests passed!"
echo ""
echo "Test Summary:"
echo "- Relay server: ✓ Running"
echo "- Cadence server: ✓ Running"
echo "- Health endpoint: ✓ Working"
echo "- Deployment workflow: ✓ Started ($WORKFLOW_ID)"
echo "- Workflow status: ✓ Retrieved"
echo "- Backup workflow: ✓ Started ($BACKUP_WORKFLOW_ID)"
echo ""
echo "Cadence integration is working correctly!"
