#!/bin/bash
# Надёжный сбор AI-метрик для Zabbix: безопасные fallback при недоступном API
# По умолчанию используем NodePort релэя
AI_API_URL="${AI_API_URL:-https://auth.2gc.ru:30082}"
AI_METRICS_URL="${AI_METRICS_URL:-http://192.168.49.2:30083/metrics}"
METRIC_TYPE="$1"

fetch() {
  local path="$1"
  curl -sk --max-time 3 "$AI_API_URL$path" 2>/dev/null || true
}

fetch_prometheus() {
  curl -s --max-time 3 "$AI_METRICS_URL" 2>/dev/null || true
}

get_prometheus_metric() {
  local metric_name="$1"
  local value=$(fetch_prometheus | grep "^${metric_name} " | awk '{print $2}' | head -1)
  echo "${value:-0}"
}

jq_num_or0() {
  local json="$1"; local expr="$2"
  if [ -z "$json" ]; then echo 0; return; fi
  printf '%s' "$json" | jq -r "$expr // 0" 2>/dev/null || echo 0
}

jq_text_or_unknown() {
  local json="$1"; local expr="$2"
  if [ -z "$json" ]; then echo "unknown"; return; fi
  printf '%s' "$json" | jq -r "$expr // \"unknown\"" 2>/dev/null || echo "unknown"
}

case "$METRIC_TYPE" in
  status)
    body=$(fetch "/api/v1/ai/status"); jq_text_or_unknown "$body" '.status' ;;
  health)
    body=$(fetch "/api/v1/ai/status"); st=$(jq_text_or_unknown "$body" '.status'); if [ "$st" = "running" ] || [ "$st" = "healthy" ]; then echo 1; else echo 0; fi ;;
  anomaly_detections)
    get_prometheus_metric "ai_anomalies_detected_total" ;;
  load_predictions)
    get_prometheus_metric "ai_predictions_made_total" ;;
  routes_optimized)
    get_prometheus_metric "ai_routes_optimized_total" ;;
  processing_time)
    get_prometheus_metric "ai_service_uptime_seconds" ;;
  anomaly_detection_accuracy)
    get_prometheus_metric "ai_prediction_accuracy" ;;
  load_prediction_accuracy)
    get_prometheus_metric "ai_prediction_accuracy" ;;
  route_optimization_accuracy)
    get_prometheus_metric "ai_prediction_accuracy" ;;
  active_connections)
    get_prometheus_metric "ai_requests_processed_total" ;;
  requests_per_second)
    get_prometheus_metric "ai_requests_processed_total" ;;
  error_rate)
    get_prometheus_metric "ai_active_alerts" ;;
  cpu_usage)
    get_prometheus_metric "ai_cpu_usage_percent" ;;
  memory_usage)
    get_prometheus_metric "ai_memory_usage_bytes" ;;
  disk_usage)
    get_prometheus_metric "ai_disk_usage_bytes" ;;
  active_alerts)
    get_prometheus_metric "ai_active_alerts" ;;
  requests_processed)
    get_prometheus_metric "ai_requests_processed_total" ;;
  service_uptime)
    get_prometheus_metric "ai_service_uptime_seconds" ;;
  docker_containers_running)
    docker ps --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo 0 ;;
  docker_containers_total)
    docker ps -a --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo 0 ;;
  k8s_pods_running)
    kubectl get pods --all-namespaces --no-headers 2>/dev/null | grep -c "Running" 2>/dev/null | grep -v "error\|denied\|permission" || echo 0 ;;
  k8s_pods_total)
    kubectl get pods --all-namespaces --no-headers 2>/dev/null | wc -l 2>/dev/null | grep -v "error\|denied\|permission" || echo 0 ;;
  k8s_nodes_ready)
    kubectl get nodes --no-headers 2>/dev/null | grep -c "Ready" 2>/dev/null | grep -v "error\|denied\|permission" || echo 0 ;;
  relay_connections)
    netstat -an 2>/dev/null | grep :8080 | grep ESTABLISHED | wc -l || echo 0 ;;
  relay_connections_total)
    netstat -an 2>/dev/null | grep :8080 | wc -l || echo 0 ;;
  relay_uptime)
    pgrep -f cloudbridge-relay | head -1 | xargs -I {} ps -o etime= -p {} 2>/dev/null | tr -d ' ' | awk -F: '{if(NF==3) print $1*3600+$2*60+$3; else if(NF==2) print $1*60+$2; else print $1}' || echo 0 ;;
  *)
    echo 0 ;;
esac
