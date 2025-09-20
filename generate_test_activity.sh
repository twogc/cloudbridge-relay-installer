#!/bin/bash

# Скрипт для генерации тестовой активности
echo "🚀 Генерация тестовой активности для AI Service"
echo "=============================================="

MINIKUBE_IP=$(minikube ip)
echo "Minikube IP: $MINIKUBE_IP"
echo ""

# Функция для генерации запросов
generate_requests() {
    local count=$1
    local endpoint=$2
    local name=$3
    
    echo "📊 Генерируем $count запросов к $name..."
    
    for i in $(seq 1 $count); do
        curl -s "$endpoint" > /dev/null 2>&1 &
        if [ $((i % 10)) -eq 0 ]; then
            echo "  Отправлено $i запросов..."
        fi
    done
    
    wait
    echo "  ✅ Завершено: $count запросов к $name"
    echo ""
}

# Генерируем активность к AI Service
echo "🤖 AI Service Activity"
echo "======================"

generate_requests 50 "http://$MINIKUBE_IP:30083/api/v1/ai/status" "AI Status"
generate_requests 30 "http://$MINIKUBE_IP:30083/api/v1/ai/metrics" "AI Metrics"
generate_requests 20 "http://$MINIKUBE_IP:30083/api/v1/ai/realtime/connections" "Realtime Connections"
generate_requests 15 "http://$MINIKUBE_IP:30083/api/v1/ai/alerts" "Active Alerts"

# Генерируем активность к Relay Server
echo "🌐 Relay Server Activity"
echo "========================"

generate_requests 40 "http://$MINIKUBE_IP:30080/health" "Relay Health"
generate_requests 25 "http://$MINIKUBE_IP:32500/api/v1/health/peers" "P2P Health Peers"
generate_requests 20 "http://$MINIKUBE_IP:32500/api/v1/health/connections" "P2P Health Connections"

# Генерируем активность к Prometheus метрикам
echo "📈 Prometheus Metrics Activity"
echo "=============================="

generate_requests 30 "http://$MINIKUBE_IP:30083/metrics" "AI Prometheus Metrics"

echo "🎉 Тестовая активность сгенерирована!"
echo "Теперь AI Service должен начать обрабатывать данные."
echo ""
echo "Проверьте метрики через 30-60 секунд:"
echo "curl -s http://$MINIKUBE_IP:30083/metrics | grep -E '^ai_.*_total'"
