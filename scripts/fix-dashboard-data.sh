#!/bin/bash

# Скрипт для исправления проблемы с данными в дашборде

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE} ИСПРАВЛЕНИЕ ПРОБЛЕМЫ С ДАННЫМИ В ДАШБОРДЕ${NC}"
echo "================================================"

# Проверка метрик AI Service
echo -e "${YELLOW}📊 Проверка метрик AI Service...${NC}"
AI_METRICS=$(kubectl exec -n cloudbridge deployment/ai-service -- wget -qO- "http://localhost:8083/metrics" 2>/dev/null | grep -E "ai_" | wc -l)
echo "Найдено метрик AI Service: $AI_METRICS"

if [ $AI_METRICS -gt 0 ]; then
    echo -e "${GREEN}AI Service генерирует метрики${NC}"
else
    echo -e "${RED}❌ AI Service не генерирует метрики${NC}"
    exit 1
fi

# Проверка targets в Prometheus
echo -e "${YELLOW} Проверка targets в Prometheus...${NC}"
TARGETS=$(kubectl exec -n cloudbridge-monitoring deployment/prometheus -- wget -qO- "http://localhost:9090/api/v1/targets" 2>/dev/null | jq '.data.activeTargets[] | select(.labels.job=="ai-service") | .health' 2>/dev/null)
echo "Статус AI Service target: $TARGETS"

if [ "$TARGETS" = '"up"' ]; then
    echo -e "${GREEN}Prometheus собирает метрики AI Service${NC}"
else
    echo -e "${RED}❌ Prometheus не собирает метрики AI Service${NC}"
    exit 1
fi

# Проверка datasource в Grafana
echo -e "${YELLOW}📊 Проверка datasource в Grafana...${NC}"
DATASOURCE=$(kubectl exec -n cloudbridge-monitoring deployment/grafana -- wget -qO- "http://localhost:3000/api/datasources" --header="Authorization: Basic YWRtaW46YWRtaW4xMjM=" 2>/dev/null | jq '.[] | select(.name=="Prometheus") | .url' 2>/dev/null)
echo "URL datasource: $DATASOURCE"

if [ "$DATASOURCE" = '"http://prometheus.cloudbridge-monitoring.svc.cluster.local:9090"' ]; then
    echo -e "${GREEN}Datasource Prometheus настроен правильно${NC}"
else
    echo -e "${RED}❌ Datasource Prometheus настроен неправильно${NC}"
    exit 1
fi

# Проверка дашборда
echo -e "${YELLOW}Проверка дашборда...${NC}"
DASHBOARD=$(kubectl exec -n cloudbridge-monitoring deployment/grafana -- wget -qO- "http://localhost:3000/api/search?query=AI" --header="Authorization: Basic YWRtaW46YWRtaW4xMjM=" 2>/dev/null | jq '.[] | select(.title | contains("AI")) | .title' 2>/dev/null)
echo "Найден дашборд: $DASHBOARD"

if [ -n "$DASHBOARD" ]; then
    echo -e "${GREEN}Дашборд AI Service найден${NC}"
else
    echo -e "${RED}❌ Дашборд AI Service не найден${NC}"
    exit 1
fi

# Проверка доступности метрик через Prometheus API
echo -e "${YELLOW}🔍 Проверка доступности метрик через Prometheus API...${NC}"
UP_METRIC=$(kubectl exec -n cloudbridge-monitoring deployment/prometheus -- wget -qO- "http://localhost:9090/api/v1/query?query=up{job=\"ai-service\"}" 2>/dev/null | jq '.data.result | length' 2>/dev/null)
echo "Количество результатов для up{job=\"ai-service\"}: $UP_METRIC"

if [ "$UP_METRIC" -gt 0 ]; then
    echo -e "${GREEN}Метрика up{job=\"ai-service\"} доступна${NC}"
else
    echo -e "${RED}❌ Метрика up{job=\"ai-service\"} недоступна${NC}"
    echo -e "${YELLOW}Возможные причины:${NC}"
    echo "   1. Prometheus не может подключиться к AI Service"
    echo "   2. Неправильная конфигурация scrape"
    echo "   3. Проблемы с сетью в кластере"
fi

# Проверка HTTP метрик
echo -e "${YELLOW}🔍 Проверка HTTP метрик...${NC}"
HTTP_METRICS=$(kubectl exec -n cloudbridge-monitoring deployment/prometheus -- wget -qO- "http://localhost:9090/api/v1/query?query=http_requests_total{job=\"ai-service\"}" 2>/dev/null | jq '.data.result | length' 2>/dev/null)
echo "Количество HTTP метрик: $HTTP_METRICS"

if [ "$HTTP_METRICS" -gt 0 ]; then
    echo -e "${GREEN}HTTP метрики доступны${NC}"
else
    echo -e "${YELLOW}  HTTP метрики недоступны (это нормально, если нет запросов)${NC}"
fi

# Проверка AI метрик
echo -e "${YELLOW}🔍 Проверка AI метрик...${NC}"
AI_METRICS_COUNT=$(kubectl exec -n cloudbridge-monitoring deployment/prometheus -- wget -qO- "http://localhost:9090/api/v1/query?query=ai_requests_processed_total" 2>/dev/null | jq '.data.result | length' 2>/dev/null)
echo "Количество AI метрик: $AI_METRICS_COUNT"

if [ "$AI_METRICS_COUNT" -gt 0 ]; then
    echo -e "${GREEN}AI метрики доступны${NC}"
else
    echo -e "${RED}❌ AI метрики недоступны${NC}"
fi

echo
echo "================================================"
echo -e "${GREEN}🎉 ПРОВЕРКА ЗАВЕРШЕНА!${NC}"
echo
echo -e "${BLUE}📊 Информация о доступе:${NC}"
echo "   Grafana: http://localhost:3000 (admin/admin123)"
echo "   Prometheus: http://localhost:9090"
echo
echo -e "${YELLOW}Если данные не отображаются в дашборде:${NC}"
echo "   1. Проверьте временной диапазон (должен быть 'Last 1 hour' или больше)"
echo "   2. Убедитесь, что datasource 'Prometheus' выбран"
echo "   3. Проверьте, что панели используют правильные запросы"
echo "   4. Попробуйте обновить дашборд (F5)"

