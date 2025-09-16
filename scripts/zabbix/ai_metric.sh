#!/usr/bin/env bash

# ai_metric.sh - Fetch AI Service health and metrics for Zabbix UserParameter
# Usage:
#   ai_metric.sh <metric> <base_url> [timeout_seconds]
# Metrics:
#   status              -> prints textual status if JSON has .status, else "up"/"down" by HTTP 200
#   http200             -> prints 1 if HTTP 200, else 0
#   latency_ms          -> prints curl total time in milliseconds (float)
#   requests_total      -> from /metrics (Prometheus) or /api/v1/ai/metrics JSON
#   errors_total        -> same as above
#   anomalies_total     -> same as above
#   predictions_total   -> same as above
#   error_rate          -> percentage if derivable from /metrics or JSON, else ZBX_NOTSUPPORTED

set -euo pipefail

metric="${1:-}"
base_url="${2:-http://127.0.0.1:8083}"
timeout_s="${3:-5}"

curl_bin=$(command -v curl || true)
jq_bin=$(command -v jq || true)

if [[ -z "${curl_bin}" ]]; then
  echo "ZBX_NOTSUPPORTED: curl not found"
  exit 1
fi

fetch_http() {
  local path="$1"
  local out http_code time_total
  out=$("${curl_bin}" -sS -m "${timeout_s}" -w "%{http_code} %{time_total}" -o /tmp/ai_metric_body.$$ "${base_url}${path}" || true)
  http_code=$(awk '{print $1}' <<<"${out}")
  time_total=$(awk '{print $2}' <<<"${out}")
  [[ -z "${http_code}" ]] && http_code=000
  [[ -z "${time_total}" ]] && time_total=0
  echo "${http_code}" "${time_total}"
}

get_body() {
  local path="$1"
  "${curl_bin}" -sS -m "${timeout_s}" "${base_url}${path}" || true
}

parse_prometheus_metric() {
  local body="$1" name="$2"
  # pick the last non-comment line starting with the metric name and no labels if possible
  # if multiple series exist, sum them
  local lines values sum=0
  lines=$(grep -E "^${name}(|\{.*\})\s" <<<"${body}" || true)
  if [[ -z "${lines}" ]]; then
    echo ""
    return 0
  fi
  # Sum numeric values across all matching lines
  while IFS= read -r line; do
    val=$(awk '{print $NF}' <<<"${line}")
    if [[ "${val}" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
      sum=$(awk -v a="${sum}" -v b="${val}" 'BEGIN{printf (a+b)}')
    fi
  done <<<"${lines}"
  echo "${sum}"
}

case "${metric}" in
  status)
    read -r code t <<<"$(fetch_http "/api/v1/ai/status")"
    if [[ "${code}" == "200" ]]; then
      if [[ -n "${jq_bin}" ]]; then
        body=$(get_body "/api/v1/ai/status")
        # Try to extract .status, fallback to "running"
        status=$(echo "${body}" | "${jq_bin}" -r 'try .status // empty' 2>/dev/null || true)
        if [[ -n "${status}" && "${status}" != "null" ]]; then
          echo "${status}"
        else
          echo "running"
        fi
      else
        echo "running"
      fi
    else
      echo "down"
    fi
    ;;
  http200)
    read -r code t <<<"$(fetch_http "/api/v1/ai/status")"
    if [[ "${code}" == "200" ]]; then echo 1; else echo 0; fi
    ;;
  latency_ms)
    read -r code t <<<"$(fetch_http "/api/v1/ai/status")"
    ms=$(awk -v s="${t}" 'BEGIN{printf (s*1000)}')
    echo "${ms}"
    ;;
  requests_total|errors_total|anomalies_total|predictions_total|error_rate)
    # Try Prometheus /metrics first
    metrics_body=$(get_body "/metrics")
    val=""
    case "${metric}" in
      requests_total)
        val=$(parse_prometheus_metric "${metrics_body}" "http_requests_total")
        ;;
      errors_total)
        # count 5xx as errors if available
        # Sum all series where status matches 5xx if present
        lines=$(grep -E '^http_requests_total\{.*status="5..".*\}' <<<"${metrics_body}" || true)
        if [[ -n "${lines}" ]]; then
          sum=0
          while IFS= read -r line; do
            v=$(awk '{print $NF}' <<<"${line}")
            if [[ "${v}" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
              sum=$(awk -v a="${sum}" -v b="${v}" 'BEGIN{printf (a+b)}')
            fi
          done <<<"${lines}"
          val="${sum}"
        else
          val=""
        fi
        ;;
      anomalies_total)
        val=$(parse_prometheus_metric "${metrics_body}" "ai_anomalies_detected_total")
        ;;
      predictions_total)
        val=$(parse_prometheus_metric "${metrics_body}" "ai_predictions_total")
        ;;
      error_rate)
        # Compute error rate from requests_total and errors_total if possible
        req=$(parse_prometheus_metric "${metrics_body}" "http_requests_total")
        err=""
        lines=$(grep -E '^http_requests_total\{.*status="5..".*\}' <<<"${metrics_body}" || true)
        if [[ -n "${lines}" ]]; then
          sum=0
          while IFS= read -r line; do
            v=$(awk '{print $NF}' <<<"${line}")
            if [[ "${v}" =~ ^-?[0-9]+(\.[0-9]+)?$ ]]; then
              sum=$(awk -v a="${sum}" -v b="${v}" 'BEGIN{printf (a+b)}')
            fi
          done <<<"${lines}"
          err="${sum}"
        fi
        if [[ -n "${req}" && -n "${err}" && "${req}" != "0" ]]; then
          rate=$(awk -v e="${err}" -v r="${req}" 'BEGIN{printf (100*e/r)}')
          val="${rate}"
        else
          val=""
        fi
        ;;
    esac
    if [[ -z "${val}" ]]; then
      # Try JSON endpoint fallback
      body=$(get_body "/api/v1/ai/metrics")
      if [[ -n "${jq_bin}" && -n "${body}" ]]; then
        case "${metric}" in
          requests_total) jq_key='.requests_total' ;;
          errors_total) jq_key='.errors_total' ;;
          anomalies_total) jq_key='.anomalies_total' ;;
          predictions_total) jq_key='.predictions_total' ;;
          error_rate) jq_key='.error_rate' ;;
        esac
        if [[ -n "${jq_key:-}" ]]; then
          val=$(echo "${body}" | "${jq_bin}" -r "try ${jq_key} // empty" 2>/dev/null | tr -d '"')
        fi
      fi
    fi
    if [[ -n "${val}" ]]; then
      echo "${val}"
    else
      echo "ZBX_NOTSUPPORTED: metric ${metric} not available"
      exit 1
    fi
    ;;
  *)
    echo "ZBX_NOTSUPPORTED: unknown metric"
    exit 1
    ;;
esac


