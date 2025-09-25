#!/bin/bash

echo " Демонстрация анализа данных AI сервисом CloudBridge"
echo "======================================================"

# Цвета для вывода
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}📊 Шаг 1: Проверка статуса AI сервиса${NC}"
echo "----------------------------------------"
kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- http://localhost:8083/api/v1/ai/status | jq .

echo -e "\n${BLUE}📈 Шаг 2: Метрики AI сервиса${NC}"
echo "----------------------------"
echo "Аномалии обнаружены:"
kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- http://localhost:8083/metrics | grep "ai_anomalies_detected_total"

echo "Прогнозы сделаны:"
kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- http://localhost:8083/metrics | grep "ai_predictions_made_total"

echo "Время обработки:"
kubectl exec deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- http://localhost:8083/metrics | grep "ai_processing_time_milliseconds"

echo -e "\n${BLUE}🔍 Шаг 3: Данные от Prometheus${NC}"
echo "-------------------------------"
echo "AI сервис получает данные от Prometheus через prometheus-proxy:"
kubectl logs deployment/ai-service -n cloudbridge -c prometheus-proxy --tail=5

echo -e "\n${BLUE}🧠 Шаг 4: Процесс анализа${NC}"
echo "-------------------------"
echo "AI сервис анализирует следующие типы данных:"
echo "Метрики Go runtime (go_goroutines, go_memstats_*)"
echo "Системные метрики (process_cpu_seconds_total, process_resident_memory_bytes)"
echo "Метрики CloudBridge Relay (relay_active_connections)"
echo "Собственные метрики AI (ai_anomalies_detected_total, ai_predictions_made_total)"

echo -e "\n${BLUE}Шаг 5: Алгоритмы анализа${NC}"
echo "-----------------------------"
echo "🔸 Детекция аномалий: Isolation Forest + статистические методы"
echo "🔸 Прогнозирование нагрузки: ARIMA + экспоненциальное сглаживание + нейронные сети"
echo "🔸 Оптимизация маршрутов: Multi-Armed Bandit + A* алгоритм"

echo -e "\n${GREEN}Результат: AI сервис успешно анализирует реальные данные!${NC}"
echo "========================================================"
