#!/usr/bin/env bash

# p2p_metrics_collector.sh - P2P QUIC metrics collector for Zabbix
# Usage: p2p_metrics_collector.sh <metric_type>
# 
# Metric Types:
#   connections_active     -> Active P2P connections count
#   connections_total      -> Total P2P connections established
#   peers_active          -> Number of active peers
#   peers_discovered      -> Total peers discovered
#   peers_offline         -> Number of offline peers
#   throughput_inbound    -> Inbound throughput (bytes/sec)
#   throughput_outbound   -> Outbound throughput (bytes/sec)
#   latency_average       -> Average connection latency (ms)
#   mesh_routes_active    -> Active mesh routes
#   ddos_blocked          -> DDoS attacks blocked
#   handshake_duration    -> Average handshake duration (ms)
#   packets_sent          -> Total packets sent
#   packets_received      -> Total packets received
#   packets_lost          -> Total packets lost
#   streams_active        -> Active QUIC streams count

set -euo pipefail

METRIC_TYPE="${1:-}"
P2P_API_URL="${P2P_API_URL:-http://192.168.49.2:30082}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://192.168.49.2:30083/metrics}"
TIMEOUT="${TIMEOUT:-5}"

curl_bin=$(command -v curl || true)
jq_bin=$(command -v jq || true)

if [[ -z "${curl_bin}" ]]; then
  echo "ZBX_NOTSUPPORTED: curl not found"
  exit 1
fi

# Fetch data from P2P API
fetch_p2p_api() {
  local path="$1"
  "${curl_bin}" -sS -m "${TIMEOUT}" "${P2P_API_URL}${path}" 2>/dev/null || true
}

# Fetch Prometheus metrics
fetch_prometheus() {
  "${curl_bin}" -sS -m "${TIMEOUT}" "${PROMETHEUS_URL}" 2>/dev/null || true
}

# Parse Prometheus metric
parse_prometheus_metric() {
  local metric_name="$1"
  local value=$(fetch_prometheus | grep "^${metric_name} " | awk '{print $2}' | head -1)
  echo "${value:-0}"
}

# Parse JSON value with jq
jq_num_or0() {
  local json="$1"
  local expr="$2"
  if [[ -z "$json" ]]; then echo 0; return; fi
  if [[ -n "${jq_bin}" ]]; then
    printf '%s' "$json" | "${jq_bin}" -r "$expr // 0" 2>/dev/null || echo 0
  else
    echo 0
  fi
}

# Get system network connections
get_network_connections() {
  local port="$1"
  netstat -an 2>/dev/null | grep ":${port}" | grep ESTABLISHED | wc -l || echo 0
}

case "${METRIC_TYPE}" in
  connections_active)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/health/connections")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.metrics.active_connections'
    else
      # Try alternative endpoint
      api_data=$(fetch_p2p_api "/api/v1/stats")
      if [[ -n "$api_data" ]]; then
        jq_num_or0 "$api_data" '.active_connections'
      else
        # Fallback to system monitoring
        get_network_connections "9090"
      fi
    fi
    ;;
    
  connections_total)
    # Try Prometheus first, fallback to API
    prom_value=$(parse_prometheus_metric "cloudbridge_p2p_connections_total")
    if [[ "$prom_value" != "0" ]]; then
      echo "$prom_value"
    else
      api_data=$(fetch_p2p_api "/api/v1/health/connections")
      jq_num_or0 "$api_data" '.metrics.total_connections'
    fi
    ;;
    
  peers_active)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/health/peers")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.metrics.online_peers'
    else
      # Try alternative endpoint
      api_data=$(fetch_p2p_api "/api/v1/stats")
      if [[ -n "$api_data" ]]; then
        jq_num_or0 "$api_data" '.active_peers'
      else
        parse_prometheus_metric "cloudbridge_p2p_active_peers"
      fi
    fi
    ;;
    
  peers_discovered)
    api_data=$(fetch_p2p_api "/api/v1/health/peers")
    jq_num_or0 "$api_data" '.metrics.total_peers'
    ;;
    
  peers_offline)
    api_data=$(fetch_p2p_api "/api/v1/health/peers")
    jq_num_or0 "$api_data" '.metrics.offline_peers'
    ;;
    
  throughput_inbound)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/stats")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.bytes_received'
    else
      # Fallback to Prometheus
      parse_prometheus_metric "cloudbridge_p2p_traffic_bytes_total{direction=\"inbound\"}"
    fi
    ;;
    
  throughput_outbound)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/stats")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.bytes_sent'
    else
      # Fallback to Prometheus
      parse_prometheus_metric "cloudbridge_p2p_traffic_bytes_total{direction=\"outbound\"}"
    fi
    ;;
    
  latency_average)
    # Get average latency from Prometheus histogram
    api_data=$(fetch_p2p_api "/api/v1/health/connections")
    jq_num_or0 "$api_data" '.metrics.avg_latency_ms'
    ;;
    
  mesh_routes_active)
    # Try Prometheus first, fallback to API
    prom_value=$(parse_prometheus_metric "cloudbridge_p2p_mesh_routes")
    if [[ "$prom_value" != "0" ]]; then
      echo "$prom_value"
    else
      api_data=$(fetch_p2p_api "/api/v1/relay/stats")
      jq_num_or0 "$api_data" '.metrics.mesh_routes'
    fi
    ;;
    
  ddos_blocked)
    # Parse DDoS protection metrics
    parse_prometheus_metric "cloudbridge_p2p_ddos_blocked_total"
    ;;
    
  handshake_duration)
    # Get handshake duration from Prometheus
    parse_prometheus_metric "cloudbridge_p2p_handshake_duration_seconds"
    ;;
    
  packets_sent)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/stats")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.packets_sent'
    else
      # Fallback to Prometheus
      parse_prometheus_metric "cloudbridge_p2p_packets_sent_total"
    fi
    ;;
    
  packets_received)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/stats")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.packets_received'
    else
      # Fallback to Prometheus
      parse_prometheus_metric "cloudbridge_p2p_packets_received_total"
    fi
    ;;
    
  packets_lost)
    # Get real data from P2P API
    api_data=$(fetch_p2p_api "/api/v1/stats")
    if [[ -n "$api_data" ]]; then
      jq_num_or0 "$api_data" '.packets_lost'
    else
      # Fallback: Calculate packet loss
      sent=$(parse_prometheus_metric "cloudbridge_p2p_packets_sent_total")
      received=$(parse_prometheus_metric "cloudbridge_p2p_packets_received_total")
      if [[ "$sent" != "0" && "$received" != "0" ]]; then
        echo $(($sent - $received))
      else
        echo 0
      fi
    fi
    ;;
    
  streams_active)
    # Get active QUIC streams
    parse_prometheus_metric "cloudbridge_p2p_streams_active"
    ;;
    
  # Enhanced QUIC metrics
  multipath_connections)
    parse_prometheus_metric "enhanced_quic_multipath_connections"
    ;;
    
  zero_rtt_connections)
    parse_prometheus_metric "enhanced_quic_zero_rtt_connections"
    ;;
    
  connection_migrations)
    parse_prometheus_metric "enhanced_quic_connection_migrations_total"
    ;;
    
  # Security metrics
  whitelist_violations)
    parse_prometheus_metric "cloudbridge_p2p_whitelist_violations_total"
    ;;
    
  rate_limit_hits)
    parse_prometheus_metric "cloudbridge_p2p_rate_limit_hits_total"
    ;;
    
  # Fallback events
  fallback_events)
    parse_prometheus_metric "cloudbridge_p2p_fallback_events_total"
    ;;
    
  *)
    echo "ZBX_NOTSUPPORTED: unknown metric type"
    exit 1
    ;;
esac
