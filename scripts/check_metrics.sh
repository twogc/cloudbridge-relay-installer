#!/bin/bash

# Скрипт для проверки реальных метрик AI Service
echo "🔍 ПРОВЕРКА РЕАЛЬНЫХ МЕТРИК AI SERVICE"
echo "======================================"

MINIKUBE_IP=$(minikube ip)
echo "Minikube IP: $MINIKUBE_IP"
echo ""

# Функция для проверки метрики
check_metric() {
    local name="$1"
    local prometheus_name="$2"
    local api_endpoint="$3"
    
    echo "📊 Проверка: $name"
    echo "Prometheus: $prometheus_name"
    
    # Получаем значение из Prometheus
    prometheus_value=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "^$prometheus_name" | awk '{print $2}' | head -1)
    
    # Получаем значение из API
    api_value=$(curl -s "$api_endpoint" | jq -r '.metrics.'"$name" 2>/dev/null || echo "N/A")
    
    echo "  Prometheus: $prometheus_value"
    echo "  API: $api_value"
    
    # Проверяем, что значения не равны 0 или пустые
    if [ "$prometheus_value" != "0" ] && [ "$prometheus_value" != "" ]; then
        echo "  ✅ Prometheus: РЕАЛЬНЫЕ ДАННЫЕ"
    else
        echo "  ⚠️  Prometheus: НЕТ ДАННЫХ"
    fi
    
    if [ "$api_value" != "0" ] && [ "$api_value" != "null" ] && [ "$api_value" != "N/A" ]; then
        echo "  ✅ API: РЕАЛЬНЫЕ ДАННЫЕ"
    else
        echo "  ⚠️  API: НЕТ ДАННЫХ"
    fi
    echo ""
}

# Проверяем все метрики
echo "🤖 AI SERVICE METRICS"
echo "====================="

check_metric "anomalies_detected" "ai_anomalies_detected_total" "http://$MINIKUBE_IP:30083/api/v1/ai/metrics"
check_metric "predictions_made" "ai_predictions_made_total" "http://$MINIKUBE_IP:30083/api/v1/ai/metrics"
check_metric "routes_optimized" "ai_routes_optimized_total" "http://$MINIKUBE_IP:30083/api/v1/ai/metrics"
check_metric "requests_processed" "ai_requests_processed_total" "http://$MINIKUBE_IP:30083/api/v1/ai/metrics"

echo "📈 ДОПОЛНИТЕЛЬНЫЕ МЕТРИКИ"
echo "========================="

# Проверяем точность предсказаний
echo "🎯 Route Optimization Accuracy:"
accuracy=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "ai_prediction_accuracy" | awk '{print $2}' | head -1)
echo "  Значение: $accuracy"
if [ "$accuracy" != "0" ] && [ "$accuracy" != "" ]; then
    echo "  ✅ РЕАЛЬНЫЕ ДАННЫЕ"
else
    echo "  ⚠️  НЕТ ДАННЫХ"
fi
echo ""

# Проверяем время обработки
echo "⏱️  AI Processing Time:"
processing_time=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "ai_processing_time" | awk '{print $2}' | head -1)
echo "  Значение: $processing_time ms"
if [ "$processing_time" != "0" ] && [ "$processing_time" != "" ]; then
    echo "  ✅ РЕАЛЬНЫЕ ДАННЫЕ"
else
    echo "  ⚠️  НЕТ ДАННЫХ"
fi
echo ""

# Проверяем CPU и Memory
echo "💻 System Metrics:"
cpu_usage=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "ai_cpu_usage_percent" | awk '{print $2}' | head -1)
memory_usage=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "ai_memory_usage_bytes" | awk '{print $2}' | head -1)
echo "  CPU Usage: $cpu_usage%"
echo "  Memory Usage: $memory_usage bytes"
echo ""

# Проверяем статус AI Service
echo "🔍 AI Service Status:"
status=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.status' 2>/dev/null)
echo "  Status: $status"
if [ "$status" = "running" ]; then
    echo "  ✅ СЕРВИС РАБОТАЕТ"
else
    echo "  ❌ СЕРВИС НЕ РАБОТАЕТ"
fi
echo ""

# Проверяем компоненты
echo "🧩 AI Components:"
components=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.components' 2>/dev/null)
echo "  Components: $components"
echo ""

# Проверяем логи на наличие реальных данных
echo "📋 Log Analysis:"
echo "Проверяем логи AI Service на наличие реальных данных..."
ai_logs=$(kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service -c ai-service --tail=50 2>/dev/null | grep -E "(real|actual|metrics|data)" | wc -l)
echo "  Найдено упоминаний реальных данных: $ai_logs"
if [ "$ai_logs" -gt 0 ]; then
    echo "  ✅ ЛОГИ СОДЕРЖАТ РЕАЛЬНЫЕ ДАННЫЕ"
else
    echo "  ⚠️  ЛОГИ НЕ СОДЕРЖАТ РЕАЛЬНЫХ ДАННЫХ"
fi
echo ""

# Итоговый отчет
echo "📊 ИТОГОВЫЙ ОТЧЕТ"
echo "=================="

total_metrics=0
real_metrics=0

# Подсчитываем реальные метрики
for metric in "ai_anomalies_detected_total" "ai_predictions_made_total" "ai_routes_optimized_total" "ai_requests_processed_total"; do
    value=$(curl -s "http://$MINIKUBE_IP:30083/metrics" | grep "^$metric" | awk '{print $2}' | head -1)
    total_metrics=$((total_metrics + 1))
    if [ "$value" != "0" ] && [ "$value" != "" ]; then
        real_metrics=$((real_metrics + 1))
    fi
done

echo "Всего метрик: $total_metrics"
echo "Реальных метрик: $real_metrics"
echo "Процент реальных данных: $((real_metrics * 100 / total_metrics))%"

if [ "$real_metrics" -eq "$total_metrics" ]; then
    echo "🎉 ВСЕ МЕТРИКИ СОДЕРЖАТ РЕАЛЬНЫЕ ДАННЫЕ!"
elif [ "$real_metrics" -gt 0 ]; then
    echo "✅ ЧАСТИЧНО РЕАЛЬНЫЕ ДАННЫЕ"
else
    echo "❌ НЕТ РЕАЛЬНЫХ ДАННЫХ"
fi

echo ""
echo "🔍 ДЕТАЛЬНАЯ ПРОВЕРКА ИСТОЧНИКОВ ДАННЫХ"
echo "======================================="

# Проверяем, откуда берутся данные
echo "Проверяем источники данных в AI Service..."
echo "1. Data Pipeline:"
pipeline_status=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.components.data_pipeline' 2>/dev/null)
echo "   Status: $pipeline_status"

echo "2. Anomaly Detector:"
anomaly_status=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.components.anomaly_detector' 2>/dev/null)
echo "   Status: $anomaly_status"

echo "3. Load Predictor:"
predictor_status=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.components.load_predictor' 2>/dev/null)
echo "   Status: $predictor_status"

echo "4. Route Optimizer:"
optimizer_status=$(curl -s "http://$MINIKUBE_IP:30083/api/v1/ai/status" | jq -r '.components.route_optimizer' 2>/dev/null)
echo "   Status: $optimizer_status"

echo ""
echo "✅ ПРОВЕРКА ЗАВЕРШЕНА"
