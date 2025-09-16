#!/bin/bash

# Обновленная интеграция AI метрик с Zabbix
# Этот скрипт добавляет AI метрики в Zabbix Agent

set -e

echo "🤖 Настройка интеграции AI метрик с Zabbix..."

# Конфигурация
ZABBIX_AGENT_CONFIG="/etc/zabbix/zabbix_agentd.conf"
AI_METRICS_SCRIPT="/usr/local/bin/ai_metrics_collector.sh"
AI_API_URL="http://localhost:8083"

# Создание скрипта для сбора AI метрик
echo "📊 Создание скрипта сбора AI метрик..."
sudo tee $AI_METRICS_SCRIPT > /dev/null <<'EOF'
#!/bin/bash

# Скрипт для сбора AI метрик для Zabbix
AI_API_URL="http://localhost:8083"
METRIC_TYPE="$1"

case "$METRIC_TYPE" in
    "status")
        # Статус AI сервиса
        curl -s "$AI_API_URL/api/v1/ai/status" | jq -r '.status // "unknown"'
        ;;
    "health")
        # Общее здоровье AI сервиса (1 = здоров, 0 = нездоров)
        status=$(curl -s "$AI_API_URL/api/v1/ai/status" | jq -r '.status // "unknown"')
        if [ "$status" = "running" ]; then
            echo "1"
        else
            echo "0"
        fi
        ;;
    "anomaly_detections")
        # Количество обнаруженных аномалий
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.anomalies_detected // 0'
        ;;
    "load_predictions")
        # Количество предсказаний нагрузки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.load_predictions // 0'
        ;;
    "routes_optimized")
        # Количество оптимизированных маршрутов
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.routes_optimized // 0'
        ;;
    "processing_time")
        # Среднее время обработки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.avg_processing_time // 0'
        ;;
    "cache_hits")
        # Количество попаданий в кэш
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.cache_hits // 0'
        ;;
    "cache_misses")
        # Количество промахов кэша
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.cache_misses // 0'
        ;;
    "circuit_breaker_state")
        # Состояние circuit breaker
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.circuit_breaker_state // "closed"'
        ;;
    "anomaly_detection_accuracy")
        # Точность обнаружения аномалий
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.anomaly_detection_accuracy // 0'
        ;;
    "load_prediction_accuracy")
        # Точность предсказания нагрузки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.load_prediction_accuracy // 0'
        ;;
    "route_optimization_accuracy")
        # Точность оптимизации маршрутов
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.route_optimization_accuracy // 0'
        ;;
    "active_connections")
        # Активные соединения, обрабатываемые AI
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.active_connections // 0'
        ;;
    "requests_per_second")
        # Запросов в секунду
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.requests_per_second // 0'
        ;;
    "error_rate")
        # Процент ошибок
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.error_rate // 0'
        ;;
    *)
        echo "0"
        ;;
esac
EOF

# Делаем скрипт исполняемым
sudo chmod +x $AI_METRICS_SCRIPT

# Добавление AI метрик в конфигурацию Zabbix Agent
echo "🔧 Добавление AI метрик в конфигурацию Zabbix Agent..."

# Создаем резервную копию конфигурации
sudo cp $ZABBIX_AGENT_CONFIG ${ZABBIX_AGENT_CONFIG}.backup.$(date +%Y%m%d_%H%M%S)

# Добавляем AI UserParameter'ы (соответствуют ключам в шаблоне)
sudo tee -a $ZABBIX_AGENT_CONFIG > /dev/null <<'EOF'

# AI Metrics UserParameters (соответствуют ключам в шаблоне)
UserParameter=ai.status,/usr/local/bin/ai_metrics_collector.sh status
UserParameter=ai.health,/usr/local/bin/ai_metrics_collector.sh health
UserParameter=ai.anomaly_detections,/usr/local/bin/ai_metrics_collector.sh anomaly_detections
UserParameter=ai.load_predictions,/usr/local/bin/ai_metrics_collector.sh load_predictions
UserParameter=ai.routes_optimized,/usr/local/bin/ai_metrics_collector.sh routes_optimized
UserParameter=ai.processing_time,/usr/local/bin/ai_metrics_collector.sh processing_time
UserParameter=ai.cache_hits,/usr/local/bin/ai_metrics_collector.sh cache_hits
UserParameter=ai.cache_misses,/usr/local/bin/ai_metrics_collector.sh cache_misses
UserParameter=ai.circuit_breaker_state,/usr/local/bin/ai_metrics_collector.sh circuit_breaker_state
UserParameter=ai.anomaly_detection_accuracy,/usr/local/bin/ai_metrics_collector.sh anomaly_detection_accuracy
UserParameter=ai.load_prediction_accuracy,/usr/local/bin/ai_metrics_collector.sh load_prediction_accuracy
UserParameter=ai.route_optimization_accuracy,/usr/local/bin/ai_metrics_collector.sh route_optimization_accuracy
UserParameter=ai.active_connections,/usr/local/bin/ai_metrics_collector.sh active_connections
UserParameter=ai.requests_per_second,/usr/local/bin/ai_metrics_collector.sh requests_per_second
UserParameter=ai.error_rate,/usr/local/bin/ai_metrics_collector.sh error_rate

# CloudBridge Relay UserParameters (для системных метрик)
UserParameter=cloudbridge.relay.connections,netstat -an | grep :8080 | grep ESTABLISHED | wc -l
UserParameter=cloudbridge.relay.uptime,ps -o etime= -p $(pgrep -f cloudbridge-relay) | tr -d ' '
UserParameter=cloudbridge.relay.memory,ps -o rss= -p $(pgrep -f cloudbridge-relay) | tail -1 | awk '{print $1/1024}'
UserParameter=cloudbridge.relay.goroutines,curl -s http://localhost:8080/debug/pprof/goroutine?debug=1 | grep -c "goroutine"

# CloudBridge AI UserParameters (альтернативные ключи)
UserParameter=cloudbridge.ai.status,/usr/local/bin/ai_metrics_collector.sh status
UserParameter=cloudbridge.ai.detections,/usr/local/bin/ai_metrics_collector.sh anomaly_detections
UserParameter=cloudbridge.ai.predictions,/usr/local/bin/ai_metrics_collector.sh load_predictions
UserParameter=cloudbridge.ai.optimizations,/usr/local/bin/ai_metrics_collector.sh routes_optimized

# System UserParameters
UserParameter=cloudbridge.system.cpu.usage,top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}'
UserParameter=cloudbridge.system.memory.usage,free | awk 'NR==2{printf "%.2f", $3*100/$2}'
UserParameter=cloudbridge.system.disk.usage,df -h / | awk 'NR==2 {print $5}' | sed 's/%//'
UserParameter=cloudbridge.system.network.connections,netstat -an | grep ESTABLISHED | wc -l

# Docker UserParameters
UserParameter=cloudbridge.docker.containers.running,docker ps --format "table {{.Names}}" | grep -v NAMES | wc -l
UserParameter=cloudbridge.docker.containers.total,docker ps -a --format "table {{.Names}}" | grep -v NAMES | wc -l
UserParameter=cloudbridge.docker.images.total,docker images --format "table {{.Repository}}" | grep -v REPOSITORY | wc -l

# Kubernetes UserParameters
UserParameter=cloudbridge.k8s.pods.running,kubectl get pods --all-namespaces --field-selector=status.phase=Running | grep -v NAME | wc -l
UserParameter=cloudbridge.k8s.pods.total,kubectl get pods --all-namespaces | grep -v NAME | wc -l
UserParameter=cloudbridge.k8s.nodes.ready,kubectl get nodes --field-selector=status.conditions[0].type=Ready | grep -v NAME | wc -l
EOF

# Перезапуск Zabbix Agent
echo "🔄 Перезапуск Zabbix Agent..."
sudo systemctl restart zabbix-agent

# Проверка статуса
echo "✅ Проверка статуса Zabbix Agent..."
sudo systemctl status zabbix-agent --no-pager

# Тестирование AI метрик
echo "🧪 Тестирование AI метрик..."
echo "   AI Status: $(zabbix_agentd -t ai.status)"
echo "   AI Health: $(zabbix_agentd -t ai.health)"
echo "   Anomaly Detections: $(zabbix_agentd -t ai.anomaly_detections)"
echo "   Load Predictions: $(zabbix_agentd -t ai.load_predictions)"
echo "   Routes Optimized: $(zabbix_agentd -t ai.routes_optimized)"
echo "   Processing Time: $(zabbix_agentd -t ai.processing_time)"
echo "   Cache Hits: $(zabbix_agentd -t ai.cache_hits)"
echo "   Cache Misses: $(zabbix_agentd -t ai.cache_misses)"
echo "   Circuit Breaker: $(zabbix_agentd -t ai.circuit_breaker_state)"
echo "   Error Rate: $(zabbix_agentd -t ai.error_rate)"

echo ""
echo "✅ Интеграция AI метрик с Zabbix завершена!"
echo ""
echo "📋 Следующие шаги:"
echo "1. Импортируйте шаблон CloudBridge Relay Template в Zabbix"
echo "2. Примените шаблон к хосту CloudBridge"
echo "3. Проверьте сбор данных в Monitoring → Latest data"
echo "4. Настройте дашборд для AI метрик"
echo ""
echo "🔧 Полезные команды:"
echo "   sudo systemctl status zabbix-agent  # Статус агента"
echo "   sudo tail -f /var/log/zabbix/zabbix_agentd.log  # Логи агента"
echo "   zabbix_agentd -t ai.status  # Тест AI метрики"

