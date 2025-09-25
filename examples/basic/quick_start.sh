#!/bin/bash

# CloudBridge Relay - Быстрый старт
# Простой пример запуска релей сервера

echo "🚀 CloudBridge Relay - Быстрый старт"
echo "=================================="

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен. Установите Docker для продолжения."
    exit 1
fi

# Проверка docker-compose
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose не установлен. Установите Docker Compose для продолжения."
    exit 1
fi

echo "Docker проверен"

# Запуск CloudBridge Relay
echo "🐳 Запуск CloudBridge Relay..."
docker-compose up -d

if [ $? -eq 0 ]; then
    echo "CloudBridge Relay успешно запущен!"
    echo ""
    echo "📊 Доступные эндпоинты:"
    echo "  • Relay API:        http://localhost:8082"
    echo "  • Prometheus:       http://localhost:9090"
    echo "  • Grafana:          http://localhost:3000 (admin/admin)"
    echo "  • P2P Mesh:         UDP 51820"
    echo ""
    echo "🔧 Для тестирования P2P mesh запустите:"
    echo "  ./../../tests/p2p/test_p2p_mesh.sh"
else
    echo "❌ Ошибка запуска CloudBridge Relay"
    exit 1
fi
