#!/bin/bash

# 🧪 Тест AI Workflows через Cadence
# Этот скрипт тестирует AI workflows через Cadence API

set -e

echo "🧪 Тестирование AI Workflows через Cadence..."

# Проверяем, что Cadence Web UI доступен
echo "📡 Проверка доступности Cadence Web UI..."
if curl -s http://localhost:8088 > /dev/null; then
    echo "Cadence Web UI доступен на http://localhost:8088"
else
    echo "❌ Cadence Web UI недоступен. Запускаем port-forward..."
    kubectl port-forward -n cloudbridge service/cadence-web-service 8088:8088 &
    sleep 3
fi

# Проверяем, что AI Service доступен
echo " Проверка доступности AI Service..."
if curl -s http://localhost:8083/api/v1/ai/status > /dev/null; then
    echo "AI Service доступен на http://localhost:8083"
else
    echo "❌ AI Service недоступен. Запускаем port-forward..."
    kubectl port-forward -n cloudbridge service/ai-service-nodeport 8083:8083 &
    sleep 3
fi

echo ""
echo "🎯 Тестирование AI компонентов..."

# 1. Тест статуса AI Service
echo "1️⃣ Проверка статуса AI Service..."
curl -s http://localhost:8083/api/v1/ai/status | jq '.status, .components'

# 2. Тест anomaly detection
echo ""
echo "2️⃣ Тест anomaly detection..."
curl -s -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
  -H "Content-Type: application/json" \
  -d '{
    "metrics": {
      "packets_per_second": 1500,
      "bytes_per_second": 75000,
      "connection_count": 200
    }
  }' | jq '.result'

# 3. Тест load prediction
echo ""
echo "3️⃣ Тест load prediction..."
curl -s -X POST http://localhost:8083/api/v1/ai/predict/load \
  -H "Content-Type: application/json" \
  -d '{
    "horizon": 24
  }' | jq '.prediction'

# 4. Тест route optimization
echo ""
echo "4️⃣ Тест route optimization..."
curl -s -X POST http://localhost:8083/api/v1/ai/route/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "source": "192.168.1.1",
    "destination": "192.168.1.100",
    "constraints": {
      "max_latency": 100,
      "min_bandwidth": 1000
    }
  }' | jq '.route'

# 5. Проверка метрик
echo ""
echo "5️⃣ Проверка AI метрик..."
curl -s http://localhost:8083/metrics | grep -E "(ai_|anomaly_|prediction_)" | head -10

echo ""
echo "🎉 Тестирование завершено!"
echo ""
echo "📊 Доступные интерфейсы:"
echo "   • AI Service API: http://localhost:8083"
echo "   • Cadence Web UI: http://localhost:8088"
echo "   • AI Metrics: http://localhost:8083/metrics"
echo ""
echo " Полезные ссылки:"
echo "   • AI API Docs: http://localhost:8083/api/v1/ai/status"
echo "   • Cadence Workflows: http://localhost:8088/domains/cloudbridge/workflows"
echo ""
