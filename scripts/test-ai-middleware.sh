#!/bin/bash

# Тестирование AI Middleware интеграции
echo "🧪 Тестирование AI Middleware интеграции..."

# Проверяем, что AI Service запущен
echo "1️⃣ Проверка AI Service..."
if curl -s http://localhost:8083/health > /dev/null; then
    echo "✅ AI Service доступен"
else
    echo "❌ AI Service недоступен. Запускаем port-forward..."
    kubectl port-forward -n cloudbridge service/ai-service-nodeport 8083:8083 &
    sleep 5
fi

# Тестируем AI middleware endpoints
echo "2️⃣ Тестирование AI middleware endpoints..."

# Тест anomaly detection
echo "   🔍 Тест anomaly detection..."
curl -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
  -H "Content-Type: application/json" \
  -d '{
    "metrics": {
      "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
      "connections_per_second": 100,
      "bytes_per_second": 1000000,
      "packet_loss_rate": 0.01,
      "average_latency": 50.0,
      "cpu_usage": 45.0,
      "memory_usage": 60.0,
      "active_connections": 150,
      "failed_connections": 5,
      "protocol_distribution": {
        "QUIC": 80,
        "TCP": 20
      }
    }
  }' | jq '.'

echo ""

# Тест load prediction
echo "   📊 Тест load prediction..."
curl -X POST http://localhost:8083/api/v1/ai/predict/load \
  -H "Content-Type: application/json" \
  -d '{
    "horizon": 24
  }' | jq '.'

echo ""

# Тест route optimization
echo "   🛣️ Тест route optimization..."
curl -X POST http://localhost:8083/api/v1/ai/route/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "source": "192.168.1.1",
    "destination": "192.168.1.100",
    "constraints": {
      "max_latency": 100,
      "min_bandwidth": 1000
    }
  }' | jq '.'

echo ""

# Тест AI status
echo "   📈 Тест AI status..."
curl -X GET http://localhost:8083/api/v1/ai/status | jq '.'

echo ""

# Тест AI metrics
echo "   📊 Тест AI metrics..."
curl -X GET http://localhost:8083/api/v1/ai/metrics | jq '.'

echo ""

echo "🎉 Тестирование AI Middleware завершено!"
echo ""
echo "📊 Доступные интерфейсы:"
echo "   • AI Service API: http://localhost:8083"
echo "   • AI Middleware: интегрирован в Relay Server"
echo "   • AI Metrics: http://localhost:8083/api/v1/ai/metrics"
echo ""
echo "🔗 Полезные ссылки:"
echo "   • AI API Docs: http://localhost:8083/api/v1/ai/status"
echo "   • Anomaly Detection: POST /api/v1/ai/anomaly/detect"
echo "   • Load Prediction: POST /api/v1/ai/predict/load"
echo "   • Route Optimization: POST /api/v1/ai/route/optimize"
