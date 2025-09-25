#!/bin/bash

# CloudBridge Relay - Полная демонстрация
# Включает P2P mesh, QUIC интеграцию и мониторинг

echo "🌟 CloudBridge Relay - Полная демонстрация"
echo "=========================================="

echo "Эта демонстрация включает:"
echo "  • P2P Mesh Network тестирование"
echo "  • QUIC-to-P2P интеграцию"
echo "  • Мониторинг в реальном времени"
echo "  • Prometheus метрики"
echo ""

read -p "❓ Продолжить демонстрацию? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
fi

# Проверка статуса системы
echo "🔍 Проверка системы..."
../../tools/utils/start_production.sh status

# Запуск P2P демо
echo ""
echo " Запуск P2P Mesh демонстрации..."
cd ../../tests/demo
./run_demo.sh

echo ""
echo "✨ Демонстрация завершена!"
echo "📊 Проверьте Grafana дашборды: http://localhost:3000"
