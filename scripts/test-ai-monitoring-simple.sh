#!/bin/bash

# Простое тестирование AI мониторинга без Prometheus
echo "📊 Простое тестирование AI мониторинга..."

# Проверяем доступность AI Service
echo "1️⃣ Проверка AI Service..."
if curl -s http://localhost:8083/health > /dev/null; then
    echo "✅ AI Service доступен"
else
    echo "❌ AI Service недоступен"
    exit 1
fi

# Проверяем AI метрики через API
echo "2️⃣ Проверка AI метрик через API..."

echo "   📊 AI Status:"
curl -s http://localhost:8083/api/v1/ai/status | jq .

echo ""

echo "   📈 AI Metrics:"
curl -s http://localhost:8083/api/v1/ai/metrics | jq .

echo ""

# Генерируем тестовую нагрузку
echo "3️⃣ Генерация тестовой нагрузки для метрик..."

echo "   🚨 Тестирование Anomaly Detection..."
for i in {1..5}; do
    curl -s -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
      -H "Content-Type: application/json" \
      -d '{
        "metrics": {
          "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
          "connections_per_second": '$((100 + i * 20))',
          "bytes_per_second": '$((1000000 + i * 200000))',
          "packet_loss_rate": 0.01,
          "average_latency": '$((50 + i * 5))',
          "cpu_usage": '$((45 + i * 2))',
          "memory_usage": '$((60 + i * 3))',
          "active_connections": '$((150 + i * 10))',
          "failed_connections": '$((5 + i))',
          "protocol_distribution": {
            "QUIC": 80,
            "TCP": 20
          }
        }
      }' | jq '.result | {score: .score, confidence: .confidence, is_anomaly: .is_anomaly}'
    sleep 1
done

echo ""

echo "   📈 Тестирование Load Prediction..."
for i in {1..3}; do
    curl -s -X POST http://localhost:8083/api/v1/ai/predict/load \
      -H "Content-Type: application/json" \
      -d '{
        "horizon": 24
      }' | jq '.result | {forecast: .forecast, confidence: .confidence}'
    sleep 1
done

echo ""

echo "   🛣️ Тестирование Route Optimization..."
for i in {1..3}; do
    curl -s -X POST http://localhost:8083/api/v1/ai/route/optimize \
      -H "Content-Type: application/json" \
      -d '{
        "source": "192.168.1.'$i'",
        "destination": "192.168.1.100",
        "constraints": {
          "max_latency": 100,
          "min_bandwidth": 1000
        }
      }' | jq '.result | {protocol: .protocol, latency: .latency, confidence: .confidence}'
    sleep 1
done

echo ""

# Проверяем обновленные метрики
echo "4️⃣ Проверка обновленных метрик..."

echo "   📊 AI Status (после нагрузки):"
curl -s http://localhost:8083/api/v1/ai/status | jq .

echo ""

echo "   📈 AI Metrics (после нагрузки):"
curl -s http://localhost:8083/api/v1/ai/metrics | jq .

echo ""

# Создаем простой отчет о мониторинге
echo "5️⃣ Создание отчета о мониторинге..."

# Получаем текущие метрики
status=$(curl -s http://localhost:8083/api/v1/ai/status)
metrics=$(curl -s http://localhost:8083/api/v1/ai/metrics)

echo "📋 Отчет о состоянии AI системы:"
echo "   • Время: $(date)"
echo "   • AI Service: $(echo $status | jq -r '.status')"
echo "   • Anomaly Detector: $(echo $status | jq -r '.anomaly_detector')"
echo "   • Load Predictor: $(echo $status | jq -r '.load_predictor')"
echo "   • Route Optimizer: $(echo $status | jq -r '.route_optimizer')"
echo "   • Data Pipeline: $(echo $status | jq -r '.data_pipeline')"
echo "   • Monitoring: $(echo $status | jq -r '.monitoring')"
echo "   • Requests Processed: $(echo $metrics | jq -r '.requests_processed')"
echo "   • Anomalies Detected: $(echo $metrics | jq -r '.anomalies_detected')"
echo "   • Predictions Made: $(echo $metrics | jq -r '.predictions_made')"
echo "   • Routes Optimized: $(echo $metrics | jq -r '.routes_optimized')"

echo ""

echo "🎉 Тестирование AI мониторинга завершено!"
echo ""
echo "📊 Доступные интерфейсы:"
echo "   • AI Service: http://localhost:8083"
echo "   • AI Health: http://localhost:8083/health"
echo "   • AI Status: http://localhost:8083/api/v1/ai/status"
echo "   • AI Metrics: http://localhost:8083/api/v1/ai/metrics"
echo ""
echo "🔗 API Endpoints для мониторинга:"
echo "   • Anomaly Detection: POST http://localhost:8083/api/v1/ai/anomaly/detect"
echo "   • Load Prediction: POST http://localhost:8083/api/v1/ai/predict/load"
echo "   • Route Optimization: POST http://localhost:8083/api/v1/ai/route/optimize"
echo ""
echo "📈 Ключевые метрики для мониторинга:"
echo "   • requests_processed - обработанные запросы"
echo "   • anomalies_detected - обнаруженные аномалии"
echo "   • predictions_made - сделанные предсказания"
echo "   • routes_optimized - оптимизированные маршруты"
echo "   • processing_time_ns - время обработки"
echo "   • cache_hits/cache_misses - эффективность кэша"
