# Observability Dashboards CloudBridge Relay

## 1. Prometheus Configuration

### 1.1 Prometheus Server Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'cloudbridge-relay'
    environment: 'production'

rule_files:
  - "cloudbridge-alerts.yml"
  - "cloudbridge-recording-rules.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  - job_name: 'cloudbridge-relay'
    static_configs:
      - targets: ['relay-server:9091']
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s
    
  - job_name: 'cloudbridge-ai'
    static_configs:
      - targets: ['ai-service:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    
  - job_name: 'cloudbridge-cadence'
    static_configs:
      - targets: ['cadence-service:7933']
    metrics_path: '/metrics'
    scrape_interval: 30s
    
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__
```

### 1.2 Recording Rules

```yaml
# cloudbridge-recording-rules.yml
groups:
  - name: cloudbridge.rules
    rules:
      - record: cloudbridge:connections:rate
        expr: rate(cloudbridge_connections_total[5m])
        
      - record: cloudbridge:latency:p95
        expr: histogram_quantile(0.95, rate(cloudbridge_request_duration_seconds_bucket[5m]))
        
      - record: cloudbridge:latency:p99
        expr: histogram_quantile(0.99, rate(cloudbridge_request_duration_seconds_bucket[5m]))
        
      - record: cloudbridge:throughput:bytes_per_second
        expr: rate(cloudbridge_bytes_transferred_total[1m])
        
      - record: cloudbridge:error:rate
        expr: rate(cloudbridge_errors_total[5m])
        
      - record: cloudbridge:ai:anomaly:detection:rate
        expr: rate(cloudbridge_ai_anomalies_detected_total[5m])
        
      - record: cloudbridge:cadence:workflow:success:rate
        expr: rate(cloudbridge_cadence_workflows_completed_total[5m])
```

## 2. Grafana Dashboards

### 2.1 Main Overview Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CloudBridge Relay - Main Overview",
    "tags": ["cloudbridge", "relay", "overview"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Active Connections",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge_connections_active",
            "legendFormat": "Active Connections"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "green", "value": null},
                {"color": "yellow", "value": 5000},
                {"color": "red", "value": 8000}
              ]
            },
            "unit": "short"
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Connection Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge:connections:rate",
            "legendFormat": "Connections/sec"
          }
        ],
        "yAxes": [
          {
            "label": "Connections/sec",
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Latency P95/P99",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge:latency:p95",
            "legendFormat": "P95 Latency"
          },
          {
            "expr": "cloudbridge:latency:p99",
            "legendFormat": "P99 Latency"
          }
        ],
        "yAxes": [
          {
            "label": "Latency (ms)",
            "min": 0,
            "unit": "ms"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 4,
        "title": "Throughput",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge:throughput:bytes_per_second",
            "legendFormat": "Throughput"
          }
        ],
        "yAxes": [
          {
            "label": "Bytes/sec",
            "min": 0,
            "unit": "Bps"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 5,
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge:error:rate",
            "legendFormat": "Error Rate"
          }
        ],
        "yAxes": [
          {
            "label": "Errors/sec",
            "min": 0
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

### 2.2 AI Service Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CloudBridge AI Service",
    "tags": ["cloudbridge", "ai", "ml"],
    "panels": [
      {
        "id": 1,
        "title": "Anomaly Detection Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge:ai:anomaly:detection:rate",
            "legendFormat": "Anomalies/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Model Accuracy",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge_ai_model_accuracy",
            "legendFormat": "Accuracy"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Prediction Latency",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(cloudbridge_ai_prediction_duration_seconds_bucket[5m]))",
            "legendFormat": "P95 Prediction Latency"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 4,
        "title": "Load Predictions",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge_ai_load_predictions",
            "legendFormat": "Predicted Load"
          },
          {
            "expr": "cloudbridge_connections_active",
            "legendFormat": "Actual Load"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      }
    ]
  }
}
```

### 2.3 Cadence Workflow Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CloudBridge Cadence Workflows",
    "tags": ["cloudbridge", "cadence", "workflows"],
    "panels": [
      {
        "id": 1,
        "title": "Workflow Success Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge:cadence:workflow:success:rate",
            "legendFormat": "Success Rate"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Active Workflows",
        "type": "graph",
        "targets": [
          {
            "expr": "cloudbridge_cadence_workflows_active",
            "legendFormat": "Active Workflows"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Workflow Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(cloudbridge_cadence_workflow_duration_seconds_bucket[5m]))",
            "legendFormat": "P95 Duration"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 4,
        "title": "Workflow Types",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by (workflow_type) (cloudbridge_cadence_workflows_completed_total)",
            "legendFormat": "{{workflow_type}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      }
    ]
  }
}
```

## 3. SLO/SLA Definitions

### 3.1 Service Level Objectives

```yaml
# slo-definitions.yml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloudbridge-slo
  labels:
    app: cloudbridge-relay
spec:
  selector:
    matchLabels:
      app: cloudbridge-relay
  endpoints:
    - port: metrics
      path: /metrics
      interval: 30s
---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cloudbridge-slo-rules
spec:
  groups:
    - name: cloudbridge.slo
      rules:
        # Availability SLO: 99.9% uptime
        - record: cloudbridge:slo:availability:ratio
          expr: |
            (
              sum(rate(cloudbridge_requests_total{status!~"5.."}[5m])) /
              sum(rate(cloudbridge_requests_total[5m]))
            )
        
        # Latency SLO: 95% of requests < 100ms
        - record: cloudbridge:slo:latency:ratio
          expr: |
            (
              sum(rate(cloudbridge_request_duration_seconds_bucket{le="0.1"}[5m])) /
              sum(rate(cloudbridge_request_duration_seconds_count[5m]))
            )
        
        # Error Rate SLO: < 0.1% error rate
        - record: cloudbridge:slo:error_rate:ratio
          expr: |
            (
              sum(rate(cloudbridge_requests_total{status=~"5.."}[5m])) /
              sum(rate(cloudbridge_requests_total[5m]))
            )
```

### 3.2 SLA Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "CloudBridge SLA Monitoring",
    "tags": ["cloudbridge", "sla", "slo"],
    "panels": [
      {
        "id": 1,
        "title": "Availability SLO (99.9%)",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge:slo:availability:ratio * 100",
            "legendFormat": "Availability %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100,
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 99.5},
                {"color": "green", "value": 99.9}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 8, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Latency SLO (95% < 100ms)",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge:slo:latency:ratio * 100",
            "legendFormat": "Latency SLO %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 100,
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "yellow", "value": 90},
                {"color": "green", "value": 95}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 8, "x": 8, "y": 0}
      },
      {
        "id": 3,
        "title": "Error Rate SLO (< 0.1%)",
        "type": "stat",
        "targets": [
          {
            "expr": "cloudbridge:slo:error_rate:ratio * 100",
            "legendFormat": "Error Rate %"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percent",
            "min": 0,
            "max": 1,
            "thresholds": {
              "steps": [
                {"color": "green", "value": 0},
                {"color": "yellow", "value": 0.05},
                {"color": "red", "value": 0.1}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 8, "x": 16, "y": 0}
      }
    ]
  }
}
```

## 4. Alerting Rules

### 4.1 Critical Alerts

```yaml
# cloudbridge-alerts.yml
groups:
  - name: cloudbridge.critical
    rules:
      - alert: CloudBridgeHighErrorRate
        expr: cloudbridge:error:rate > 0.01
        for: 2m
        labels:
          severity: critical
          service: cloudbridge-relay
        annotations:
          summary: "High error rate detected"
          description: "CloudBridge error rate is {{ $value }} errors/sec"
          
      - alert: CloudBridgeHighLatency
        expr: cloudbridge:latency:p95 > 0.5
        for: 5m
        labels:
          severity: critical
          service: cloudbridge-relay
        annotations:
          summary: "High latency detected"
          description: "CloudBridge P95 latency is {{ $value }}s"
          
      - alert: CloudBridgeConnectionDrop
        expr: rate(cloudbridge_connections_dropped_total[5m]) > 10
        for: 1m
        labels:
          severity: critical
          service: cloudbridge-relay
        annotations:
          summary: "High connection drop rate"
          description: "CloudBridge dropping {{ $value }} connections/sec"
          
      - alert: CloudBridgeSLOViolation
        expr: cloudbridge:slo:availability:ratio < 0.999
        for: 5m
        labels:
          severity: critical
          service: cloudbridge-relay
        annotations:
          summary: "SLA violation detected"
          description: "Availability SLO violated: {{ $value }}"
```

### 4.2 Warning Alerts

```yaml
  - name: cloudbridge.warning
    rules:
      - alert: CloudBridgeHighCPUUsage
        expr: rate(container_cpu_usage_seconds_total{pod=~"cloudbridge-.*"}[5m]) > 0.8
        for: 10m
        labels:
          severity: warning
          service: cloudbridge-relay
        annotations:
          summary: "High CPU usage"
          description: "Pod {{ $labels.pod }} CPU usage is {{ $value }}"
          
      - alert: CloudBridgeHighMemoryUsage
        expr: container_memory_usage_bytes{pod=~"cloudbridge-.*"} / container_spec_memory_limit_bytes > 0.8
        for: 10m
        labels:
          severity: warning
          service: cloudbridge-relay
        annotations:
          summary: "High memory usage"
          description: "Pod {{ $labels.pod }} memory usage is {{ $value }}"
          
      - alert: CloudBridgeAIServiceDown
        expr: up{job="cloudbridge-ai"} == 0
        for: 2m
        labels:
          severity: warning
          service: cloudbridge-ai
        annotations:
          summary: "AI Service is down"
          description: "CloudBridge AI Service is not responding"
          
      - alert: CloudBridgeCadenceWorkflowFailure
        expr: rate(cloudbridge_cadence_workflows_failed_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          service: cloudbridge-cadence
        annotations:
          summary: "High workflow failure rate"
          description: "Cadence workflow failure rate is {{ $value }}"
```

## 5. AlertManager Configuration

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@cloudbridge.io'
  smtp_auth_username: 'alerts@cloudbridge.io'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
    - match:
        severity: warning
      receiver: 'warning-alerts'

receivers:
  - name: 'default'
    email_configs:
      - to: 'team@cloudbridge.io'
        subject: 'CloudBridge Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
          
  - name: 'critical-alerts'
    email_configs:
      - to: 'oncall@cloudbridge.io'
        subject: 'CRITICAL: CloudBridge Alert'
        body: |
          {{ range .Alerts }}
          CRITICAL ALERT: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/...'
        channel: '#alerts-critical'
        title: 'Critical Alert'
        text: '{{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        
  - name: 'warning-alerts'
    email_configs:
      - to: 'team@cloudbridge.io'
        subject: 'WARNING: CloudBridge Alert'
        body: |
          {{ range .Alerts }}
          WARNING: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          {{ end }}
```

## 6. Custom Metrics

### 6.1 CloudBridge Metrics

```go
// metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // Connection metrics
    ConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "cloudbridge_connections_active",
        Help: "Number of active connections",
    })
    
    ConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
        Name: "cloudbridge_connections_total",
        Help: "Total number of connections",
    })
    
    ConnectionsDropped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "cloudbridge_connections_dropped_total",
        Help: "Total number of dropped connections",
    })
    
    // Request metrics
    RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "cloudbridge_request_duration_seconds",
        Help: "Request duration in seconds",
        Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
    })
    
    RequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "cloudbridge_requests_total",
        Help: "Total number of requests",
    }, []string{"method", "status"})
    
    // Throughput metrics
    BytesTransferred = promauto.NewCounter(prometheus.CounterOpts{
        Name: "cloudbridge_bytes_transferred_total",
        Help: "Total bytes transferred",
    })
    
    // AI metrics
    AIAnomaliesDetected = promauto.NewCounter(prometheus.CounterOpts{
        Name: "cloudbridge_ai_anomalies_detected_total",
        Help: "Total number of anomalies detected by AI",
    })
    
    AIModelAccuracy = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "cloudbridge_ai_model_accuracy",
        Help: "AI model accuracy percentage",
    })
    
    // Cadence metrics
    CadenceWorkflowsActive = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "cloudbridge_cadence_workflows_active",
        Help: "Number of active Cadence workflows",
    })
    
    CadenceWorkflowsCompleted = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "cloudbridge_cadence_workflows_completed_total",
        Help: "Total number of completed workflows",
    }, []string{"workflow_type", "status"})
)
```

---

*Эта конфигурация observability предоставляет полный мониторинг CloudBridge Relay с Prometheus метриками, Grafana дашбордами, SLO/SLA определениями и системой алертинга для обеспечения высокой доступности и производительности системы.*
