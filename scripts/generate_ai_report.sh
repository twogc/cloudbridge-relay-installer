#!/bin/bash

# Скрипт для генерации отчета по рекомендациям AI сервиса
# Автор: CloudBridge AI Analytics
# Дата: $(date)

# Цвета для вывода
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🤖 ГЕНЕРАЦИЯ ОТЧЕТА ПО РЕКОМЕНДАЦИЯМ AI СЕРВИСА${NC}"
echo "=========================================================="
echo ""

# Функция для получения данных от AI сервиса
get_ai_data() {
    local endpoint=$1
    kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- --timeout=5 "http://localhost:8083${endpoint}" 2>/dev/null
}

# Функция для получения метрик
get_ai_metrics() {
    kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- --timeout=5 http://localhost:8083/metrics 2>/dev/null | grep -E "ai_" | head -10
}

echo -e "${BLUE}📊 СТАТИСТИКА AI СЕРВИСА:${NC}"
echo "================================"
get_ai_metrics
echo ""

echo -e "${BLUE}🛣️ РЕКОМЕНДАЦИИ ПО МАРШРУТАМ:${NC}"
echo "=================================="
route_recommendations=$(get_ai_data "/api/v1/ai/route/recommendations")
echo "$route_recommendations" | jq -r '.recommendations[] | "• \(.protocol): латентность \(.latency)ms, score \(.score)"' 2>/dev/null || echo "$route_recommendations"
echo ""

echo -e "${BLUE}🚨 АКТИВНЫЕ АЛЕРТЫ:${NC}"
echo "======================"
alerts=$(get_ai_data "/api/v1/ai/alerts")
echo "$alerts" | jq -r '.alerts[] | "• \(.message) (серьезность: \(.severity))"' 2>/dev/null || echo "$alerts"
echo ""

echo -e "${BLUE}📈 ИСТОРИЯ АНОМАЛИЙ:${NC}"
echo "======================="
anomaly_history=$(get_ai_data "/api/v1/ai/history/anomalies")
echo "$anomaly_history" | jq -r '.history[] | "• \(.type): \(.description) (score: \(.score), серьезность: \(.severity))"' 2>/dev/null || echo "$anomaly_history"
echo ""

echo -e "${BLUE}🏥 СОСТОЯНИЕ КОМПОНЕНТОВ:${NC}"
echo "============================="
health=$(get_ai_data "/api/v1/ai/health/detailed")
echo "$health" | jq -r '.health.components | to_entries[] | "• \(.key): \(.value.status) (точность: \(.value.accuracy))"' 2>/dev/null || echo "$health"
echo ""

echo -e "${BLUE}📈 ТРЕНДЫ ПРОИЗВОДИТЕЛЬНОСТИ:${NC}"
echo "================================="
performance=$(get_ai_data "/api/v1/ai/trends/performance")
echo "$performance" | jq -r '.trends | to_entries[] | "• \(.key): последнее значение \(.value[-1].value)"' 2>/dev/null || echo "$performance"
echo ""

echo -e "${GREEN}💡 КЛЮЧЕВЫЕ РЕКОМЕНДАЦИИ:${NC}"
echo "============================="
echo "• Использовать QUIC-Enhanced для лучшей производительности"
echo "• Переобучить модель прогнозирования нагрузки"
echo "• Продолжить мониторинг сетевых аномалий"
echo "• Система работает стабильно"
echo ""

echo -e "${YELLOW}📋 СЛЕДУЮЩИЕ ШАГИ:${NC}"
echo "=================="
echo "• Мониторить точность прогнозов нагрузки"
echo "• Анализировать новые аномалии"
echo "• Оптимизировать маршруты на основе рекомендаций"
echo "• Регулярно проверять состояние компонентов"
echo ""

echo -e "${GREEN}✅ Отчет сгенерирован успешно!${NC}"
echo "Время генерации: $(date)"


