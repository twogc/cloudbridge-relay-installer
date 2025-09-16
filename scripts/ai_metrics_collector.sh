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
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.anomaly_detection.total_detections // 0'
        ;;
    "load_predictions")
        # Количество предсказаний нагрузки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.load_prediction.total_predictions // 0'
        ;;
    "routes_optimized")
        # Количество оптимизированных маршрутов
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.route_optimization.total_optimizations // 0'
        ;;
    "processing_time")
        # Среднее время обработки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.processing.average_time // 0'
        ;;
    "cache_hits")
        # Количество попаданий в кэш (пока не реализовано в API)
        echo "0"
        ;;
    "cache_misses")
        # Количество промахов кэша (пока не реализовано в API)
        echo "0"
        ;;
    "circuit_breaker_state")
        # Состояние circuit breaker (пока не реализовано в API)
        echo "closed"
        ;;
    "anomaly_detection_accuracy")
        # Точность обнаружения аномалий
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.anomaly_detection.accuracy // 0'
        ;;
    "load_prediction_accuracy")
        # Точность предсказания нагрузки
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.load_prediction.accuracy // 0'
        ;;
    "route_optimization_accuracy")
        # Точность оптимизации маршрутов (пока не реализовано в API)
        echo "0"
        ;;
    "active_connections")
        # Активные соединения, обрабатываемые AI
        curl -s "$AI_API_URL/api/v1/ai/metrics" | jq -r '.metrics.system.active_connections // 0'
        ;;
    "requests_per_second")
        # Запросов в секунду (пока не реализовано в API)
        echo "0"
        ;;
    "error_rate")
        # Процент ошибок (пока не реализовано в API)
        echo "0"
        ;;
    "cpu_usage")
        # Использование CPU (фиксированное значение для тестирования)
        echo "5"
        ;;
    "memory_usage")
        # Использование памяти (фиксированное значение для тестирования)
        echo "32"
        ;;
    "disk_usage")
        # Использование диска (фиксированное значение для тестирования)
        echo "52"
        ;;
    "docker_containers_running")
        # Количество запущенных Docker контейнеров
        docker ps --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
        ;;
    "docker_containers_total")
        # Общее количество Docker контейнеров
        docker ps -a --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
        ;;
    "k8s_pods_running")
        # Количество запущенных Pod'ов в Kubernetes
        kubectl get pods --no-headers 2>/dev/null | grep -c "Running" 2>/dev/null || echo "0"
        ;;
    "k8s_pods_total")
        # Общее количество Pod'ов в Kubernetes
        kubectl get pods --no-headers 2>/dev/null | wc -l 2>/dev/null || echo "0"
        ;;
    "relay_connections")
        # Количество активных соединений Relay Server
        netstat -an 2>/dev/null | grep :8080 | grep ESTABLISHED | wc -l || echo "0"
        ;;
    "relay_connections_total")
        # Общее количество соединений Relay Server
        netstat -an 2>/dev/null | grep :8080 | wc -l || echo "0"
        ;;
    *)
        echo "0"
        ;;
esac
