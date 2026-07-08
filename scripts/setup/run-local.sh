#!/bin/bash

# Скрипт для локального запуска CloudBridge Relay в Docker

set -e

echo "Запуск CloudBridge Relay локально..."

# Проверяем, собран ли образ
if ! docker images | grep -q "cloudbridge-relay.*alpine"; then
    echo "Собираем Alpine Docker образ..."
    make docker-build-alpine
fi

# Останавливаем существующие контейнеры
echo "🛑 Останавливаем существующие контейнеры..."
docker stop cloudbridge-relay-alpine 2>/dev/null || true
docker rm cloudbridge-relay-alpine 2>/dev/null || true

# Запускаем новый контейнер
echo "▶️  Запускаем Alpine Docker контейнер..."
docker run -d \
    --name cloudbridge-relay-alpine \
    --cap-add=NET_ADMIN \
    --device=/dev/net/tun \
    -p 80:80 \
    -p 8080:8080 \
    -p 8081:8081 \
    -p 8082:8082 \
    -p 9090:9090 \
    -p 51820:51820/udp \
    -v $(pwd)/configs:/app/configs:ro \
    -v $(pwd)/logs:/app/logs \
    twogc/cloudbridge-relay:alpine

echo "Контейнер запущен!"
echo ""
echo "Статус:"
docker ps | grep cloudbridge-relay-alpine
echo ""
echo "Логи:"
echo "docker logs cloudbridge-relay-alpine"
echo ""
echo "🛑 Остановка:"
echo "docker stop cloudbridge-relay-alpine"
echo ""
echo "Доступные порты:"
echo "  HTTP: http://localhost:80"
echo "  API: http://localhost:8080"
echo "  Metrics: http://localhost:8081"
echo "  QUIC: localhost:9090"
echo "  WireGuard: localhost:51820/udp"
