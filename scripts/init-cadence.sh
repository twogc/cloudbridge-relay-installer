#!/bin/bash

# Скрипт для инициализации Cadence домена и схемы базы данных

set -e

# Конфигурация
CADENCE_HOST="localhost:7933"
DOMAIN_NAME="cloudbridge"
TASK_LIST="relay-task-list"

echo "Initializing Cadence domain: $DOMAIN_NAME"

# Проверяем доступность Cadence сервера
echo "Checking Cadence server availability..."
if ! nc -z localhost 7933; then
    echo "Error: Cadence server is not running on localhost:7933"
    echo "Please start Cadence server first using: docker-compose up -d cadence cassandra"
    exit 1
fi

# Создаем домен
echo "Creating Cadence domain..."
cadence --address $CADENCE_HOST domain register \
    --global_domain false \
    --retention 7 \
    --description "CloudBridge Relay Workflow Domain" \
    $DOMAIN_NAME

if [ $? -eq 0 ]; then
    echo "Domain '$DOMAIN_NAME' created successfully"
else
    echo "Domain '$DOMAIN_NAME' might already exist or creation failed"
fi

# Проверяем создание домена
echo "Verifying domain creation..."
cadence --address $CADENCE_HOST domain describe $DOMAIN_NAME

echo "Cadence initialization completed!"
echo ""
echo "Domain: $DOMAIN_NAME"
echo "Task List: $TASK_LIST"
echo "Cadence Server: $CADENCE_HOST"
echo ""
echo "You can now start the CloudBridge Relay server with Cadence integration."



