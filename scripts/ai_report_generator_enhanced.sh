#!/bin/bash

# Улучшенный генератор отчетов AI без Cloudflare AI
# Автор: CloudBridge AI Analytics
# Дата: $(date)

# Цвета для вывода
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE} УЛУЧШЕННЫЙ ГЕНЕРАТОР ОТЧЕТОВ AI${NC}"
echo "============================================="
echo ""

# Функция для получения данных от AI сервиса
get_ai_data() {
    local endpoint=$1
    kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- --timeout=5 "http://localhost:8083${endpoint}" 2>/dev/null
}

# Функция для анализа данных и генерации рекомендаций
analyze_data() {
    local data=$1
    local type=$2
    
    case $type in
        "route")
            # Анализ рекомендаций по маршрутам
            best_protocol=$(echo "$data" | jq -r '.recommendations[] | select(.score == (.recommendations | max_by(.score) | .score)) | .protocol' 2>/dev/null)
            best_latency=$(echo "$data" | jq -r '.recommendations[] | select(.score == (.recommendations | max_by(.score) | .score)) | .latency' 2>/dev/null)
            best_score=$(echo "$data" | jq -r '.recommendations[] | select(.score == (.recommendations | max_by(.score) | .score)) | .score' 2>/dev/null)
            
            if [ "$best_score" != "null" ] && [ "$best_score" != "" ]; then
                echo "• РЕКОМЕНДУЕТСЯ: $best_protocol (латентность: ${best_latency}ms, score: $best_score)"
            fi
            ;;
        "alerts")
            # Анализ алертов
            alert_count=$(echo "$data" | jq -r '.count' 2>/dev/null)
            if [ "$alert_count" != "null" ] && [ "$alert_count" != "0" ]; then
                echo "• ВНИМАНИЕ: Обнаружено $alert_count активных алертов"
                echo "$data" | jq -r '.alerts[] | "  - \(.message) (серьезность: \(.severity))"' 2>/dev/null
            else
                echo "• Активных алертов нет"
            fi
            ;;
        "anomalies")
            # Анализ аномалий
            anomaly_count=$(echo "$data" | jq -r '.count' 2>/dev/null)
            if [ "$anomaly_count" != "null" ] && [ "$anomaly_count" != "0" ]; then
                echo "• Обнаружено $anomaly_count аномалий за последние 24 часа"
                echo "$data" | jq -r '.history[] | "  - \(.type): \(.description) (score: \(.score), серьезность: \(.severity))"' 2>/dev/null
            else
                echo "• Аномалий не обнаружено"
            fi
            ;;
        "health")
            # Анализ здоровья системы
            overall_health=$(echo "$data" | jq -r '.health.overall' 2>/dev/null)
            if [ "$overall_health" = "healthy" ]; then
                echo "• Общее состояние: $overall_health"
            else
                echo "•  Общее состояние: $overall_health"
            fi
            
            echo "$data" | jq -r '.health.components | to_entries[] | "• \(.key): \(.value.status) (точность: \(.value.accuracy))"' 2>/dev/null
            ;;
    esac
}

# Генерация отчета
echo -e "${BLUE}🕐 Время генерации: $(date)${NC}"
echo ""

echo -e "${BLUE}📊 СТАТИСТИКА AI СЕРВИСА:${NC}"
echo "================================"
ai_metrics=$(kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- --timeout=5 http://localhost:8083/metrics 2>/dev/null | grep -E "ai_" | head -5)
echo "$ai_metrics"
echo ""

echo -e "${BLUE}🛣️ РЕКОМЕНДАЦИИ ПО МАРШРУТАМ:${NC}"
echo "=================================="
route_data=$(get_ai_data "/api/v1/ai/route/recommendations")
echo "$route_data" | jq -r '.recommendations[] | "• \(.protocol): латентность \(.latency)ms, score \(.score)"' 2>/dev/null || echo "$route_data"
echo ""
echo -e "${GREEN}АНАЛИЗ РЕКОМЕНДАЦИЙ:${NC}"
analyze_data "$route_data" "route"
echo ""

echo -e "${BLUE}АКТИВНЫЕ АЛЕРТЫ:${NC}"
echo "======================"
alerts_data=$(get_ai_data "/api/v1/ai/alerts")
echo "$alerts_data" | jq -r '.alerts[] | "• \(.message) (серьезность: \(.severity))"' 2>/dev/null || echo "$alerts_data"
echo ""
echo -e "${GREEN}АНАЛИЗ АЛЕРТОВ:${NC}"
analyze_data "$alerts_data" "alerts"
echo ""

echo -e "${BLUE}📈 ИСТОРИЯ АНОМАЛИЙ:${NC}"
echo "======================="
anomaly_data=$(get_ai_data "/api/v1/ai/history/anomalies")
echo "$anomaly_data" | jq -r '.history[] | "• \(.type): \(.description) (score: \(.score), серьезность: \(.severity))"' 2>/dev/null || echo "$anomaly_data"
echo ""
echo -e "${GREEN}АНАЛИЗ АНОМАЛИЙ:${NC}"
analyze_data "$anomaly_data" "anomalies"
echo ""

echo -e "${BLUE}🏥 СОСТОЯНИЕ КОМПОНЕНТОВ:${NC}"
echo "============================="
health_data=$(get_ai_data "/api/v1/ai/health/detailed")
echo "$health_data" | jq -r '.health.components | to_entries[] | "• \(.key): \(.value.status) (точность: \(.value.accuracy))"' 2>/dev/null || echo "$health_data"
echo ""
echo -e "${GREEN}АНАЛИЗ ЗДОРОВЬЯ СИСТЕМЫ:${NC}"
analyze_data "$health_data" "health"
echo ""

echo -e "${BLUE}📈 ТРЕНДЫ ПРОИЗВОДИТЕЛЬНОСТИ:${NC}"
echo "================================="
performance_data=$(get_ai_data "/api/v1/ai/trends/performance")
echo "$performance_data" | jq -r '.trends | to_entries[] | "• \(.key): последнее значение \(.value[-1].value)"' 2>/dev/null || echo "$performance_data"
echo ""

echo -e "${GREEN}КЛЮЧЕВЫЕ РЕКОМЕНДАЦИИ:${NC}"
echo "============================="

# Генерация рекомендаций на основе данных
if echo "$route_data" | jq -r '.recommendations[] | select(.protocol == "QUIC-Enhanced") | .score' 2>/dev/null | grep -q "0.95"; then
    echo "• Использовать QUIC-Enhanced для лучшей производительности"
fi

if echo "$alerts_data" | jq -r '.alerts[] | select(.type == "low_accuracy")' 2>/dev/null | grep -q "low_accuracy"; then
    echo "• Переобучить модель прогнозирования нагрузки"
fi

if echo "$anomaly_data" | jq -r '.history[] | select(.severity == "high")' 2>/dev/null | grep -q "high"; then
    echo "• Продолжить мониторинг сетевых аномалий"
fi

if echo "$health_data" | jq -r '.health.overall' 2>/dev/null | grep -q "healthy"; then
    echo "• Система работает стабильно"
fi

echo ""

echo -e "${YELLOW}СЛЕДУЮЩИЕ ШАГИ:${NC}"
echo "=================="
echo "• Мониторить точность прогнозов нагрузки"
echo "• Анализировать новые аномалии"
echo "• Оптимизировать маршруты на основе рекомендаций"
echo "• Регулярно проверять состояние компонентов"
echo ""

echo -e "${GREEN}Отчет сгенерирован успешно!${NC}"
echo "Время генерации: $(date)"
echo ""
echo -e "${BLUE}ℹ️ Примечание: Отчет сгенерирован без Cloudflare AI (сервис недоступен)${NC}"
echo "Использованы локальные алгоритмы анализа данных"


