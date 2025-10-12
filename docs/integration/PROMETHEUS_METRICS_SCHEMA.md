# CloudBridge QUIC Experimental Features - Prometheus Metrics Schema

**Version**: 1.0.0
**Date**: 2025-01-12
**Status**: Production Ready

## Table of Contents

1. [Overview](#overview)
2. [Metrics Endpoint](#metrics-endpoint)
3. [BBRv2 Congestion Control Metrics](#bbrv2-congestion-control-metrics)
4. [ACK Frequency Optimization Metrics](#ack-frequency-optimization-metrics)
5. [Forward Error Correction (FEC) Metrics](#forward-error-correction-fec-metrics)
6. [QUIC Bit Greasing (RFC 9287) Metrics](#quic-bit-greasing-rfc-9287-metrics)
7. [Cross-Cutting Metrics](#cross-cutting-metrics)
8. [Grafana Dashboard](#grafana-dashboard)
9. [Alerting Rules](#alerting-rules)
10. [Performance Benchmarks](#performance-benchmarks)

---

## Overview

CloudBridge QUIC implements comprehensive Prometheus metrics for all experimental features:
- **BBRv2 Congestion Control** (Google's latest algorithm)
- **ACK Frequency Optimization** (draft-ietf-quic-ack-frequency)
- **Forward Error Correction (FEC)** (XOR and Reed-Solomon)
- **QUIC Bit Greasing** (RFC 9287)

All metrics are automatically exported via the `/metrics` endpoint in Prometheus exposition format.

### Naming Convention

All experimental metrics follow the pattern:
```
cloudbridge_quic_{feature}_{metric}[_{unit}]
```

Where:
- `feature`: `bbrv2`, `ack_freq`, `fec`, `greasing`, or `experimental`
- `metric`: descriptive name of the metric
- `unit`: optional unit suffix (e.g., `bytes`, `seconds`, `bps`)

---

## Metrics Endpoint

### HTTP Endpoints

| Endpoint | Description | Format |
|----------|-------------|--------|
| `/metrics` | Prometheus metrics (all features) | Prometheus exposition format |
| `/metrics/summary` | Human-readable metrics summary | JSON |

### Example Usage

```bash
# Prometheus metrics
curl http://localhost:8081/metrics

# Human-readable summary
curl http://localhost:8081/metrics/summary
```

### Configuration

```yaml
# config/experimental-features.yaml
metrics:
  enable_prometheus: true
  prometheus_port: 9091
  export_interval: 10s
  enable_detailed_metrics: true
```

---

## BBRv2 Congestion Control Metrics

### Overview

BBRv2 metrics track congestion control state, bandwidth estimation, and pacing behavior.

### Metrics

#### `cloudbridge_quic_bbrv2_state`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 congestion control state
**Values**:
- `0.0` - Startup (aggressive bandwidth probing)
- `1.0` - Drain (drain inflight packets)
- `2.0` - ProbeBW (bandwidth probing cycles)
- `3.0` - ProbeRTT (minimum RTT probing)

**Example**:
```prometheus
cloudbridge_quic_bbrv2_state{connection_id="conn-123",tenant_id="tenant-abc"} 2.0
```

#### `cloudbridge_quic_bbrv2_congestion_window_bytes`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: Current congestion window size in bytes
**Expected Range**: 14600 - 2MB (10 MTU - 1400 packets)

**Example**:
```prometheus
cloudbridge_quic_bbrv2_congestion_window_bytes{connection_id="conn-123",tenant_id="tenant-abc"} 1460000
```

#### `cloudbridge_quic_bbrv2_pacing_rate_bps`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 pacing rate in bits per second
**Expected Range**: 1 Mbps - 10 Gbps

**Example**:
```prometheus
cloudbridge_quic_bbrv2_pacing_rate_bps{connection_id="conn-123",tenant_id="tenant-abc"} 100000000
```

#### `cloudbridge_quic_bbrv2_bandwidth_bps`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 estimated bandwidth in bits per second
**Expected Range**: 1 Mbps - 10 Gbps

**Example**:
```prometheus
cloudbridge_quic_bbrv2_bandwidth_bps{connection_id="conn-123",tenant_id="tenant-abc"} 95000000
```

#### `cloudbridge_quic_bbrv2_min_rtt_seconds`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 minimum RTT in seconds
**Expected Range**: 0.001s - 1.0s

**Example**:
```prometheus
cloudbridge_quic_bbrv2_min_rtt_seconds{connection_id="conn-123",tenant_id="tenant-abc"} 0.05
```

#### `cloudbridge_quic_bbrv2_throughput_mbps`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 throughput in Mbps
**Expected Range**: 1 - 1000 Mbps

**Example**:
```prometheus
cloudbridge_quic_bbrv2_throughput_mbps{connection_id="conn-123",tenant_id="tenant-abc"} 95.5
```

#### `cloudbridge_quic_bbrv2_loss_rate`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: BBRv2 packet loss rate (0.0 to 1.0)
**Expected Range**: 0.0 - 0.1 (0% - 10%)

**Example**:
```prometheus
cloudbridge_quic_bbrv2_loss_rate{connection_id="conn-123",tenant_id="tenant-abc"} 0.01
```

#### `cloudbridge_quic_bbrv2_state_transitions_total`
**Type**: Counter
**Labels**: `from_state`, `to_state`, `tenant_id`
**Description**: Total number of BBRv2 state transitions

**Example**:
```prometheus
cloudbridge_quic_bbrv2_state_transitions_total{from_state="Startup",to_state="Drain",tenant_id="tenant-abc"} 42
```

#### `cloudbridge_quic_bbrv2_packets_in_flight`
**Type**: Gauge
**Labels**: `connection_id`, `tenant_id`
**Description**: Number of packets in flight (sent but not acknowledged)
**Expected Range**: 0 - 1000 packets

**Example**:
```prometheus
cloudbridge_quic_bbrv2_packets_in_flight{connection_id="conn-123",tenant_id="tenant-abc"} 150
```

### Queries

```promql
# Average congestion window per tenant
avg by (tenant_id) (cloudbridge_quic_bbrv2_congestion_window_bytes)

# BBRv2 throughput improvement vs baseline CUBIC
(avg(cloudbridge_quic_bbrv2_throughput_mbps) - avg(cloudbridge_quic_cubic_throughput_mbps)) / avg(cloudbridge_quic_cubic_throughput_mbps) * 100

# Packet loss rate 95th percentile
histogram_quantile(0.95, cloudbridge_quic_bbrv2_loss_rate)

# State transition rate
rate(cloudbridge_quic_bbrv2_state_transitions_total[5m])
```

---

## ACK Frequency Optimization Metrics

### Overview

ACK Frequency metrics track ACK sending behavior, bandwidth savings, and adaptive optimization.

### Metrics

#### `cloudbridge_quic_ack_freq_total_connections`
**Type**: Gauge
**Description**: Total number of connections with ACK frequency optimization

**Example**:
```prometheus
cloudbridge_quic_ack_freq_total_connections 250
```

#### `cloudbridge_quic_ack_freq_active_connections`
**Type**: Gauge
**Description**: Number of active connections with ACK frequency optimization

**Example**:
```prometheus
cloudbridge_quic_ack_freq_active_connections 180
```

#### `cloudbridge_quic_ack_freq_total_acks`
**Type**: Counter
**Description**: Total number of ACKs sent

**Example**:
```prometheus
cloudbridge_quic_ack_freq_total_acks 1500000
```

#### `cloudbridge_quic_ack_freq_delayed_acks_total`
**Type**: Counter
**Description**: Total number of delayed ACKs

**Example**:
```prometheus
cloudbridge_quic_ack_freq_delayed_acks_total 450000
```

#### `cloudbridge_quic_ack_freq_adaptive_acks_total`
**Type**: Counter
**Description**: Total number of adaptive ACKs

**Example**:
```prometheus
cloudbridge_quic_ack_freq_adaptive_acks_total 300000
```

#### `cloudbridge_quic_ack_freq_average_delay_seconds`
**Type**: Histogram
**Description**: Average ACK delay in seconds
**Buckets**: [0.001, 0.002, 0.005, 0.01, 0.025, 0.05, 0.1]

**Example**:
```prometheus
cloudbridge_quic_ack_freq_average_delay_seconds_bucket{le="0.005"} 1200
cloudbridge_quic_ack_freq_average_delay_seconds_bucket{le="0.01"} 1450
cloudbridge_quic_ack_freq_average_delay_seconds_sum 7.5
cloudbridge_quic_ack_freq_average_delay_seconds_count 1500
```

#### `cloudbridge_quic_ack_freq_bandwidth_saved_bytes`
**Type**: Counter
**Description**: Total bandwidth saved by ACK frequency optimization (bytes)

**Example**:
```prometheus
cloudbridge_quic_ack_freq_bandwidth_saved_bytes 50000000
```

#### `cloudbridge_quic_ack_freq_packets_received_total`
**Type**: Counter
**Labels**: `connection_id`, `tenant_id`
**Description**: Total packets received per connection

**Example**:
```prometheus
cloudbridge_quic_ack_freq_packets_received_total{connection_id="conn-123",tenant_id="tenant-abc"} 15000
```

#### `cloudbridge_quic_ack_freq_bytes_received_total`
**Type**: Counter
**Labels**: `connection_id`, `tenant_id`
**Description**: Total bytes received per connection

**Example**:
```prometheus
cloudbridge_quic_ack_freq_bytes_received_total{connection_id="conn-123",tenant_id="tenant-abc"} 21900000
```

### Queries

```promql
# ACK overhead reduction percentage
(1 - cloudbridge_quic_ack_freq_total_acks / cloudbridge_quic_ack_freq_packets_received_total) * 100

# Average ACK delay (P50, P95, P99)
histogram_quantile(0.5, rate(cloudbridge_quic_ack_freq_average_delay_seconds_bucket[5m]))
histogram_quantile(0.95, rate(cloudbridge_quic_ack_freq_average_delay_seconds_bucket[5m]))
histogram_quantile(0.99, rate(cloudbridge_quic_ack_freq_average_delay_seconds_bucket[5m]))

# Bandwidth savings rate (bytes/sec)
rate(cloudbridge_quic_ack_freq_bandwidth_saved_bytes[5m])

# Adaptive ACK ratio
cloudbridge_quic_ack_freq_adaptive_acks_total / cloudbridge_quic_ack_freq_total_acks
```

---

## Forward Error Correction (FEC) Metrics

### Overview

FEC metrics track error correction performance, recovery rate, and efficiency.

### Metrics

#### `cloudbridge_quic_fec_total_groups`
**Type**: Counter
**Description**: Total number of FEC groups created

**Example**:
```prometheus
cloudbridge_quic_fec_total_groups 50000
```

#### `cloudbridge_quic_fec_active_groups`
**Type**: Gauge
**Description**: Number of active FEC groups

**Example**:
```prometheus
cloudbridge_quic_fec_active_groups 120
```

#### `cloudbridge_quic_fec_redundancy_bytes_total`
**Type**: Counter
**Description**: Total redundancy bytes added by FEC

**Example**:
```prometheus
cloudbridge_quic_fec_redundancy_bytes_total 5000000
```

#### `cloudbridge_quic_fec_recovery_events_total`
**Type**: Counter
**Description**: Total number of successful packet recoveries

**Example**:
```prometheus
cloudbridge_quic_fec_recovery_events_total 4500
```

#### `cloudbridge_quic_fec_failed_recoveries_total`
**Type**: Counter
**Description**: Total number of failed recovery attempts

**Example**:
```prometheus
cloudbridge_quic_fec_failed_recoveries_total 150
```

#### `cloudbridge_quic_fec_recovery_rate`
**Type**: Gauge
**Description**: FEC recovery rate (0.0 to 1.0)
**Target**: ≥ 0.95 (95%)

**Example**:
```prometheus
cloudbridge_quic_fec_recovery_rate 0.967
```

#### `cloudbridge_quic_fec_efficiency`
**Type**: Gauge
**Description**: FEC efficiency (recovered_bytes / redundancy_bytes)
**Expected Range**: 0.5 - 10.0

**Example**:
```prometheus
cloudbridge_quic_fec_efficiency 2.5
```

#### `cloudbridge_quic_fec_latency_seconds`
**Type**: Histogram
**Description**: FEC encoding/decoding latency in seconds
**Buckets**: [0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05]

**Example**:
```prometheus
cloudbridge_quic_fec_latency_seconds_bucket{le="0.001"} 45000
cloudbridge_quic_fec_latency_seconds_sum 25.5
cloudbridge_quic_fec_latency_seconds_count 50000
```

#### `cloudbridge_quic_fec_bytes_saved_total`
**Type**: Counter
**Description**: Total bytes saved by avoiding retransmission

**Example**:
```prometheus
cloudbridge_quic_fec_bytes_saved_total 12000000
```

#### `cloudbridge_quic_fec_data_packets_total`
**Type**: Counter
**Labels**: `algorithm`, `tenant_id`
**Description**: Total data packets in FEC groups

**Example**:
```prometheus
cloudbridge_quic_fec_data_packets_total{algorithm="xor",tenant_id="tenant-abc"} 480000
```

#### `cloudbridge_quic_fec_redundant_packets_total`
**Type**: Counter
**Labels**: `algorithm`, `tenant_id`
**Description**: Total redundant packets created by FEC

**Example**:
```prometheus
cloudbridge_quic_fec_redundant_packets_total{algorithm="xor",tenant_id="tenant-abc"} 48000
```

### Queries

```promql
# FEC recovery success rate
cloudbridge_quic_fec_recovery_events_total / (cloudbridge_quic_fec_recovery_events_total + cloudbridge_quic_fec_failed_recoveries_total)

# FEC overhead percentage
(cloudbridge_quic_fec_redundancy_bytes_total / cloudbridge_quic_fec_data_packets_total) * 100

# FEC latency P95
histogram_quantile(0.95, rate(cloudbridge_quic_fec_latency_seconds_bucket[5m]))

# Bytes saved per second
rate(cloudbridge_quic_fec_bytes_saved_total[5m])

# FEC efficiency trend (7 days)
avg_over_time(cloudbridge_quic_fec_efficiency[7d])
```

---

## QUIC Bit Greasing (RFC 9287) Metrics

### Overview

Greasing metrics track middlebox compatibility and ossification prevention.

### Metrics

#### `cloudbridge_quic_greasing_version_total`
**Type**: Counter
**Description**: Total number of version greasing applications

**Example**:
```prometheus
cloudbridge_quic_greasing_version_total 25000
```

#### `cloudbridge_quic_greasing_transport_param_total`
**Type**: Counter
**Description**: Total number of transport parameter greasing applications

**Example**:
```prometheus
cloudbridge_quic_greasing_transport_param_total 25000
```

#### `cloudbridge_quic_greasing_frame_type_total`
**Type**: Counter
**Description**: Total number of frame type greasing applications

**Example**:
```prometheus
cloudbridge_quic_greasing_frame_type_total 0
```

#### `cloudbridge_quic_greasing_applications_total`
**Type**: Counter
**Description**: Total number of greasing applications (all types)

**Example**:
```prometheus
cloudbridge_quic_greasing_applications_total 50000
```

#### `cloudbridge_quic_greasing_compatibility_issues_total`
**Type**: Counter
**Description**: Total number of compatibility issues detected

**Example**:
```prometheus
cloudbridge_quic_greasing_compatibility_issues_total 5
```

#### `cloudbridge_quic_greasing_active_connections`
**Type**: Gauge
**Description**: Number of active connections with greasing enabled

**Example**:
```prometheus
cloudbridge_quic_greasing_active_connections 200
```

### Queries

```promql
# Greasing success rate
1 - (cloudbridge_quic_greasing_compatibility_issues_total / cloudbridge_quic_greasing_applications_total)

# Greasing application rate
rate(cloudbridge_quic_greasing_applications_total[5m])

# Version greasing percentage
cloudbridge_quic_greasing_version_total / cloudbridge_quic_greasing_applications_total * 100

# Transport param greasing percentage
cloudbridge_quic_greasing_transport_param_total / cloudbridge_quic_greasing_applications_total * 100
```

---

## Cross-Cutting Metrics

### Overview

Cross-cutting metrics track overall experimental feature adoption and performance improvements.

### Metrics

#### `cloudbridge_quic_experimental_connections_total`
**Type**: Counter
**Labels**: `feature`, `tenant_id`
**Description**: Total connections using experimental features

**Example**:
```prometheus
cloudbridge_quic_experimental_connections_total{feature="bbrv2",tenant_id="tenant-abc"} 1500
cloudbridge_quic_experimental_connections_total{feature="ack_frequency",tenant_id="tenant-abc"} 1500
cloudbridge_quic_experimental_connections_total{feature="fec",tenant_id="tenant-abc"} 500
cloudbridge_quic_experimental_connections_total{feature="greasing",tenant_id="tenant-abc"} 1400
```

#### `cloudbridge_quic_experimental_connections_active`
**Type**: Gauge
**Labels**: `feature`, `tenant_id`
**Description**: Active connections using experimental features

**Example**:
```prometheus
cloudbridge_quic_experimental_connections_active{feature="bbrv2",tenant_id="tenant-abc"} 250
```

#### `cloudbridge_quic_experimental_latency_improvement_percent`
**Type**: Gauge
**Labels**: `feature`, `tenant_id`
**Description**: Latency improvement compared to baseline (%)
**Expected Range**: -10% to +50%

**Example**:
```prometheus
cloudbridge_quic_experimental_latency_improvement_percent{feature="bbrv2",tenant_id="tenant-abc"} 15.5
```

#### `cloudbridge_quic_experimental_throughput_improvement_percent`
**Type**: Gauge
**Labels**: `feature`, `tenant_id`
**Description**: Throughput improvement compared to baseline (%)
**Expected Range**: -10% to +60%

**Example**:
```prometheus
cloudbridge_quic_experimental_throughput_improvement_percent{feature="bbrv2",tenant_id="tenant-abc"} 45.2
```

#### `cloudbridge_quic_experimental_packet_loss_reduction_percent`
**Type**: Gauge
**Labels**: `feature`, `tenant_id`
**Description**: Packet loss reduction compared to baseline (%)
**Expected Range**: -10% to +95%

**Example**:
```prometheus
cloudbridge_quic_experimental_packet_loss_reduction_percent{feature="fec",tenant_id="tenant-abc"} 85.0
```

### Queries

```promql
# Overall experimental feature adoption
sum by (feature) (cloudbridge_quic_experimental_connections_active)

# BBRv2 throughput improvement (average across all tenants)
avg(cloudbridge_quic_experimental_throughput_improvement_percent{feature="bbrv2"})

# ACK Frequency latency improvement
avg(cloudbridge_quic_experimental_latency_improvement_percent{feature="ack_frequency"})

# FEC packet loss reduction
avg(cloudbridge_quic_experimental_packet_loss_reduction_percent{feature="fec"})

# Multi-feature performance comparison
avg by (feature) (cloudbridge_quic_experimental_throughput_improvement_percent)
```

---

## Grafana Dashboard

### Dashboard JSON

See [grafana-dashboard.json](./grafana-dashboard.json) for the complete Grafana dashboard configuration.

### Dashboard Panels

1. **BBRv2 Overview**
   - State distribution (pie chart)
   - Congestion window timeline
   - Throughput comparison (BBRv2 vs CUBIC)
   - State transition heatmap

2. **ACK Frequency**
   - ACK delay histogram
   - Bandwidth savings trend
   - Delayed ACK ratio
   - Adaptive ACK percentage

3. **FEC Performance**
   - Recovery rate gauge
   - FEC efficiency trend
   - Latency histogram
   - Bytes saved counter

4. **Greasing Status**
   - Greasing application types (bar chart)
   - Compatibility issues timeline
   - Active connections gauge

5. **Cross-Cutting Analysis**
   - Feature adoption by tenant
   - Performance improvements comparison
   - Multi-feature correlation matrix

### Import Instructions

```bash
# Import dashboard to Grafana
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-api-key>" \
  -d @grafana-dashboard.json
```

---

## Alerting Rules

### Prometheus Alerting Rules

```yaml
# alerts.yaml
groups:
  - name: cloudbridge_quic_experimental
    interval: 30s
    rules:
      # BBRv2 Alerts
      - alert: BBRv2HighLossRate
        expr: cloudbridge_quic_bbrv2_loss_rate > 0.05
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High packet loss rate detected ({{ $value }}%)"
          description: "Connection {{ $labels.connection_id }} for tenant {{ $labels.tenant_id }} has packet loss > 5%"

      - alert: BBRv2LowThroughput
        expr: cloudbridge_quic_bbrv2_throughput_mbps < 10
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Low throughput detected ({{ $value }} Mbps)"
          description: "Connection {{ $labels.connection_id }} has throughput < 10 Mbps"

      # ACK Frequency Alerts
      - alert: ACKFrequencyHighDelay
        expr: histogram_quantile(0.95, rate(cloudbridge_quic_ack_freq_average_delay_seconds_bucket[5m])) > 0.025
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High ACK delay detected (P95: {{ $value }}s)"
          description: "ACK delay P95 > 25ms"

      # FEC Alerts
      - alert: FECLowRecoveryRate
        expr: cloudbridge_quic_fec_recovery_rate < 0.90
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Low FEC recovery rate ({{ $value }})"
          description: "FEC recovery rate < 90%"

      - alert: FECHighLatency
        expr: histogram_quantile(0.95, rate(cloudbridge_quic_fec_latency_seconds_bucket[5m])) > 0.005
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High FEC latency detected (P95: {{ $value }}s)"
          description: "FEC latency P95 > 5ms"

      # Greasing Alerts
      - alert: GreasingCompatibilityIssues
        expr: rate(cloudbridge_quic_greasing_compatibility_issues_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Greasing compatibility issues detected ({{ $value }}/s)"
          description: "Greasing is causing compatibility issues"

      # Cross-Cutting Alerts
      - alert: ExperimentalFeaturePerformanceDegradation
        expr: cloudbridge_quic_experimental_throughput_improvement_percent < -5
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Experimental feature causing performance degradation"
          description: "Feature {{ $labels.feature }} for tenant {{ $labels.tenant_id }} shows {{ $value }}% degradation"
```

---

## Performance Benchmarks

### Expected Performance Improvements

Based on integration testing and RFC expectations:

| Feature | Metric | Expected Improvement | Measurement Method |
|---------|--------|---------------------|-------------------|
| **BBRv2** | Throughput | +40-60% (high RTT) | `cloudbridge_quic_experimental_throughput_improvement_percent` |
| **BBRv2** | Latency | +15-25% | `cloudbridge_quic_experimental_latency_improvement_percent` |
| **ACK Frequency** | ACK Overhead | -20-40% | `(1 - ack_freq_total_acks / packets_received) * 100` |
| **ACK Frequency** | Bandwidth | -5-15% overhead | `ack_freq_bandwidth_saved_bytes / bytes_received * 100` |
| **FEC** | Packet Loss | -80-95% recovery | `cloudbridge_quic_fec_recovery_rate` |
| **FEC** | Retransmission | -50-80% | `fec_bytes_saved_total / retransmission_bytes * 100` |
| **Greasing** | Compatibility | 99.5%+ success | `1 - (compatibility_issues / applications) * 100` |

### Benchmark Scenarios

```bash
# Scenario 1: High RTT (200ms)
# Expected: BBRv2 +50% throughput vs CUBIC
cloudbridge_quic_experimental_throughput_improvement_percent{feature="bbrv2",rtt="200ms"} ~ 50

# Scenario 2: High packet loss (5%)
# Expected: FEC 90%+ recovery rate
cloudbridge_quic_fec_recovery_rate{loss_rate="5%"} >= 0.90

# Scenario 3: Bulk transfer
# Expected: ACK Frequency -30% overhead
(1 - cloudbridge_quic_ack_freq_total_acks{traffic="bulk"} / packets) * 100 ~ 30

# Scenario 4: Interactive traffic
# Expected: ACK Frequency <5ms delay
histogram_quantile(0.95, ack_freq_delay{traffic="interactive"}) < 0.005
```

---

## Appendix

### Related Documentation

- [Experimental Features Integration Report](./EXPERIMENTAL_FEATURES_INTEGRATION_REPORT.md)
- [Configuration Guide](../../config/experimental-features.yaml)
- [Usage Examples](../../internal/quic/experimental_example.go)
- [Unit Tests Report](./UNIT_TESTS_COMPLETION_REPORT.md)

### References

- [BBRv2: Congestion Control Based on Explicit Congestion Notification](https://datatracker.ietf.org/doc/html/draft-cardwell-iccrg-bbr-congestion-control)
- [QUIC ACK Frequency](https://datatracker.ietf.org/doc/html/draft-ietf-quic-ack-frequency)
- [QUIC Bit Greasing (RFC 9287)](https://www.rfc-editor.org/rfc/rfc9287.html)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/naming/)

### Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-01-12 | Initial metrics schema documentation |

---

**Status**: ✅ Complete
**Next Steps**: Grafana Dashboard Creation, Performance Testing
