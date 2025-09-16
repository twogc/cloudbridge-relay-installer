#!/bin/bash

# AI Service Dashboard Setup Script
# Creates Grafana dashboard for AI service monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="cloudbridge-monitoring"
GRAFANA_NAMESPACE="cloudbridge-monitoring"

echo -e "${BLUE}🚀 AI Service Dashboard Setup Script${NC}"
echo "================================================"

# Function to check if namespace exists
check_namespace() {
    if kubectl get namespace $NAMESPACE >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Namespace $NAMESPACE exists${NC}"
        return 0
    else
        echo -e "${RED}❌ Namespace $NAMESPACE not found${NC}"
        return 1
    fi
}

# Function to wait for deployment
wait_for_deployment() {
    local deployment="$1"
    local namespace="$2"
    local timeout="${3:-300}"
    
    echo -n "Waiting for $deployment to be ready... "
    
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace >/dev/null 2>&1; then
        echo -e "${GREEN}✅ Ready${NC}"
        return 0
    else
        echo -e "${RED}❌ Timeout${NC}"
        return 1
    fi
}

# Function to create Grafana dashboard
create_dashboard() {
    local dashboard_name="$1"
    local dashboard_json="$2"
    
    echo -n "Creating dashboard $dashboard_name... "
    
    # Create ConfigMap with dashboard
    kubectl create configmap $dashboard_name --from-literal=dashboard.json="$dashboard_json" -n $GRAFANA_NAMESPACE --dry-run=client -o yaml | kubectl apply -f - >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Created${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed${NC}"
        return 1
    fi
}

echo -e "${YELLOW}📋 Setting up AI Service Dashboard...${NC}"
echo

# Check namespaces
echo -e "${BLUE}🔍 Checking namespaces...${NC}"
check_namespace $NAMESPACE
check_namespace $GRAFANA_NAMESPACE

echo

# Check if Grafana is running
echo -e "${BLUE}🔍 Checking Grafana deployment...${NC}"
if kubectl get deployment grafana -n $GRAFANA_NAMESPACE >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Grafana deployment found${NC}"
    wait_for_deployment "grafana" $GRAFANA_NAMESPACE
else
    echo -e "${RED}❌ Grafana deployment not found${NC}"
    echo -e "${YELLOW}   Please deploy Grafana first${NC}"
    exit 1
fi

echo

# Create AI Service Dashboard
echo -e "${BLUE}📊 Creating AI Service Dashboard...${NC}"

# AI Service Dashboard JSON
AI_DASHBOARD_JSON='{
  "dashboard": {
    "id": null,
    "title": "AI Service Monitoring",
    "tags": ["ai", "cloudbridge"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "AI Service Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"ai-service\"}",
            "legendFormat": "AI Service"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "thresholds"
            },
            "thresholds": {
              "steps": [
                {"color": "red", "value": 0},
                {"color": "green", "value": 1}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"ai-service\"}[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"ai-service\"}[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job=\"ai-service\"}[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 4,
        "title": "Anomaly Detection Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(ai_anomalies_detected_total[5m])",
            "legendFormat": "Anomalies/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 8}
      },
      {
        "id": 5,
        "title": "Prediction Accuracy",
        "type": "graph",
        "targets": [
          {
            "expr": "ai_prediction_accuracy",
            "legendFormat": "Accuracy"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 8}
      },
      {
        "id": 6,
        "title": "Active Alerts",
        "type": "table",
        "targets": [
          {
            "expr": "ai_active_alerts",
            "legendFormat": "{{alert_type}}"
          }
        ],
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 16}
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}'

create_dashboard "ai-service-dashboard" "$AI_DASHBOARD_JSON"

echo

# Create Prometheus ServiceMonitor for AI Service
echo -e "${BLUE}📊 Creating Prometheus ServiceMonitor...${NC}"

SERVICEMONITOR_YAML='apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: ai-service
  namespace: cloudbridge
  labels:
    app: ai-service
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: ai-service
  endpoints:
  - port: http
    path: /metrics
    interval: 30s'

echo -n "Creating ServiceMonitor... "
echo "$SERVICEMONITOR_YAML" | kubectl apply -f - >/dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Created${NC}"
else
    echo -e "${RED}❌ Failed${NC}"
fi

echo

# Create PrometheusRule for AI Service alerts
echo -e "${BLUE}📊 Creating PrometheusRule for alerts...${NC}"

PROMETHEUSRULE_YAML='apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: ai-service-alerts
  namespace: cloudbridge
  labels:
    app: ai-service
spec:
  groups:
  - name: ai-service
    rules:
    - alert: AIServiceDown
      expr: up{job="ai-service"} == 0
      for: 1m
      labels:
        severity: critical
      annotations:
        summary: "AI Service is down"
        description: "AI Service has been down for more than 1 minute"
    
    - alert: HighErrorRate
      expr: rate(http_requests_total{job="ai-service",status=~"5.."}[5m]) > 0.1
      for: 2m
      labels:
        severity: warning
      annotations:
        summary: "High error rate in AI Service"
        description: "Error rate is above 10% for more than 2 minutes"
    
    - alert: HighResponseTime
      expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="ai-service"}[5m])) > 1
      for: 3m
      labels:
        severity: warning
      annotations:
        summary: "High response time in AI Service"
        description: "95th percentile response time is above 1 second"
    
    - alert: AnomalyDetectionFailure
      expr: rate(ai_anomalies_detected_total[5m]) == 0
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Anomaly detection not working"
        description: "No anomalies detected in the last 5 minutes"'

echo -n "Creating PrometheusRule... "
echo "$PROMETHEUSRULE_YAML" | kubectl apply -f - >/dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Created${NC}"
else
    echo -e "${RED}❌ Failed${NC}"
fi

echo

# Summary
echo "================================================"
echo -e "${GREEN}🎉 AI Service Dashboard Setup Complete!${NC}"
echo

echo -e "${BLUE}📊 Access Information:${NC}"
echo "  - Grafana: kubectl port-forward -n $GRAFANA_NAMESPACE svc/grafana 3000:80"
echo "  - Prometheus: kubectl port-forward -n $GRAFANA_NAMESPACE svc/prometheus 9090:9090"
echo "  - AI Service: kubectl port-forward -n cloudbridge svc/ai-service 8083:8083"
echo

echo -e "${BLUE}🔗 URLs:${NC}"
echo "  - Grafana: http://localhost:3000 (admin/admin)"
echo "  - Prometheus: http://localhost:9090"
echo "  - AI Service: http://localhost:8083"
echo

echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "  1. Access Grafana and import the AI Service dashboard"
echo "  2. Configure alerting channels (email, Slack, etc.)"
echo "  3. Set up additional custom panels as needed"
echo "  4. Monitor the AI service performance and alerts"