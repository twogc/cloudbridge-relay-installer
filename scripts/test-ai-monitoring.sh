#!/bin/bash

# Тестирование AI мониторинга и дашборда
echo "📊 Тестирование AI мониторинга и дашборда..."

# Проверяем доступность Prometheus
echo "1️⃣ Проверка Prometheus..."
if curl -s http://localhost:9090/api/v1/query?query=up > /dev/null; then
    echo "✅ Prometheus доступен"
else
    echo "❌ Prometheus недоступен. Запускаем port-forward..."
    kubectl port-forward -n cloudbridge service/prometheus-service 9090:9090 &
    sleep 5
fi

# Проверяем доступность Grafana
echo "2️⃣ Проверка Grafana..."
if curl -s http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana доступен"
else
    echo "❌ Grafana недоступен. Запускаем port-forward..."
    kubectl port-forward -n cloudbridge service/grafana-service 3000:3000 &
    sleep 5
fi

# Проверяем AI метрики в Prometheus
echo "3️⃣ Проверка AI метрик в Prometheus..."

# Проверяем метрики AI connections
echo "   🔍 AI Connections Processed:"
curl -s "http://localhost:9090/api/v1/query?query=ai_connections_processed_total" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

# Проверяем метрики аномалий
echo "   🚨 Anomalies Detected:"
curl -s "http://localhost:9090/api/v1/query?query=ai_anomalies_detected_total" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

# Проверяем метрики кэша
echo "   💾 Cache Hit Rate:"
cache_hits=$(curl -s "http://localhost:9090/api/v1/query?query=ai_cache_hits_total" | jq -r '.data.result[0].value[1] // "0"')
cache_misses=$(curl -s "http://localhost:9090/api/v1/query?query=ai_cache_misses_total" | jq -r '.data.result[0].value[1] // "0"')
if [ "$cache_hits" != "0" ] || [ "$cache_misses" != "0" ]; then
    total=$((cache_hits + cache_misses))
    hit_rate=$((cache_hits * 100 / total))
    echo "   Cache Hit Rate: ${hit_rate}% (${cache_hits}/${total})"
else
    echo "   Cache Hit Rate: N/A (no data)"
fi

echo ""

# Проверяем метрики circuit breaker
echo "   🔄 Circuit Breaker State:"
curl -s "http://localhost:9090/api/v1/query?query=ai_circuit_breaker_state" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

# Проверяем метрики здоровья сервиса
echo "   ❤️ AI Service Health:"
curl -s "http://localhost:9090/api/v1/query?query=ai_service_health" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

# Проверяем метрики производительности
echo "   ⚡ Processing Duration:"
curl -s "http://localhost:9090/api/v1/query?query=ai_processing_duration_seconds" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

# Проверяем алерты
echo "4️⃣ Проверка AI алертов..."
echo "   🚨 Active Alerts:"
curl -s "http://localhost:9090/api/v1/alerts" | jq '.data.alerts[] | select(.labels.service == "ai") | {alertname: .labels.alertname, state: .state, severity: .labels.severity}'

echo ""

# Генерируем тестовую нагрузку для метрик
echo "5️⃣ Генерация тестовой нагрузки..."
echo "   📈 Отправка тестовых запросов к AI Service..."

for i in {1..10}; do
    # Anomaly detection
    curl -s -X POST http://localhost:8083/api/v1/ai/anomaly/detect \
      -H "Content-Type: application/json" \
      -d '{
        "metrics": {
          "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
          "connections_per_second": '$((100 + i * 10))',
          "bytes_per_second": '$((1000000 + i * 100000))',
          "packet_loss_rate": 0.01,
          "average_latency": '$((50 + i))',
          "cpu_usage": '$((45 + i))',
          "memory_usage": '$((60 + i))',
          "active_connections": '$((150 + i * 5))',
          "failed_connections": '$((5 + i))',
          "protocol_distribution": {
            "QUIC": 80,
            "TCP": 20
          }
        }
      }' > /dev/null

    # Load prediction
    curl -s -X POST http://localhost:8083/api/v1/ai/predict/load \
      -H "Content-Type: application/json" \
      -d '{
        "horizon": 24
      }' > /dev/null

    # Route optimization
    curl -s -X POST http://localhost:8083/api/v1/ai/route/optimize \
      -H "Content-Type: application/json" \
      -d '{
        "source": "192.168.1.'$i'",
        "destination": "192.168.1.100",
        "constraints": {
          "max_latency": 100,
          "min_bandwidth": 1000
        }
      }' > /dev/null

    echo "   Запрос $i/10 отправлен"
    sleep 1
done

echo ""

# Проверяем обновленные метрики
echo "6️⃣ Проверка обновленных метрик..."
echo "   📊 AI Connections Processed (после нагрузки):"
curl -s "http://localhost:9090/api/v1/query?query=ai_connections_processed_total" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

echo "   🚨 Anomalies Detected (после нагрузки):"
curl -s "http://localhost:9090/api/v1/query?query=ai_anomalies_detected_total" | jq '.data.result[] | {metric: .metric, value: .value[1]}'

echo ""

echo "🎉 Тестирование AI мониторинга завершено!"
echo ""
echo "📊 Доступные интерфейсы:"
echo "   • Prometheus: http://localhost:9090"
echo "   • Grafana: http://localhost:3000"
echo "   • AI Dashboard: http://localhost:3000/d/ai-dashboard"
echo "   • AI Alerts: http://localhost:9090/alerts"
echo ""
echo "🔗 Полезные ссылки:"
echo "   • AI Metrics: http://localhost:9090/api/v1/query?query=ai_"
echo "   • AI Alerts: http://localhost:9090/api/v1/alerts"
echo "   • Grafana AI Dashboard: http://localhost:3000/d/ai-dashboard"
echo ""
echo "📈 Ключевые метрики для мониторинга:"
echo "   • ai_connections_processed_total - обработанные соединения"
echo "   • ai_anomalies_detected_total - обнаруженные аномалии"
echo "   • ai_cache_hits_total / ai_cache_misses_total - эффективность кэша"
echo "   • ai_circuit_breaker_state - состояние circuit breaker"
echo "   • ai_service_health - здоровье AI сервисов"
