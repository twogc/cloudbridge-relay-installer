#!/bin/bash

# Grafana Dashboard Import Script
# Imports AI Service dashboard into Grafana

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GRAFANA_URL="http://localhost:3000"
GRAFANA_USER="admin"
GRAFANA_PASSWORD="admin123"
NAMESPACE="cloudbridge-monitoring"

echo -e "${BLUE}Grafana Dashboard Import Script${NC}"
echo "================================================"

# Function to check if Grafana is accessible
check_grafana() {
    echo -n "Checking Grafana accessibility... "
    
    if curl -s -o /dev/null -w "%{http_code}" "$GRAFANA_URL" | grep -q "200\|302"; then
        echo -e "${GREEN}Accessible${NC}"
        return 0
    else
        echo -e "${RED}Not accessible${NC}"
        return 1
    fi
}

# Function to get Grafana API key
get_api_key() {
    echo -n "Getting Grafana API key... "
    
    # Try to get API key
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"ai-dashboard-key\",\"role\":\"Admin\"}" \
        "$GRAFANA_URL/api/auth/keys" \
        -u "$GRAFANA_USER:$GRAFANA_PASSWORD" 2>/dev/null)
    
    if echo "$response" | grep -q "key"; then
        API_KEY=$(echo "$response" | grep -o '"key":"[^"]*"' | cut -d'"' -f4)
        echo -e "${GREEN}Got API key${NC}"
        return 0
    else
        echo -e "${YELLOW}  Using basic auth${NC}"
        return 1
    fi
}

# Function to import dashboard
import_dashboard() {
    local dashboard_name="$1"
    local dashboard_file="$2"
    
    echo -n "Importing $dashboard_name... "
    
    # Get dashboard JSON from ConfigMap
    dashboard_json=$(kubectl get configmap grafana-dashboards -n $NAMESPACE -o jsonpath="{.data.$dashboard_file}" 2>/dev/null)
    
    if [ -z "$dashboard_json" ]; then
        echo -e "${RED}Dashboard not found in ConfigMap${NC}"
        return 1
    fi
    
    # Import dashboard
    if [ -n "$API_KEY" ]; then
        response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $API_KEY" \
            -d "{\"dashboard\":$dashboard_json,\"overwrite\":true}" \
            "$GRAFANA_URL/api/dashboards/db" 2>/dev/null)
    else
        response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"dashboard\":$dashboard_json,\"overwrite\":true}" \
            "$GRAFANA_URL/api/dashboards/db" \
            -u "$GRAFANA_USER:$GRAFANA_PASSWORD" 2>/dev/null)
    fi
    
    if echo "$response" | grep -q "success\|id"; then
        echo -e "${GREEN}Imported${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        echo "Response: $response"
        return 1
    fi
}

# Function to create simple dashboard manually
create_simple_dashboard() {
    echo -e "${YELLOW}Creating simple AI Service dashboard...${NC}"
    
    # Simple dashboard JSON
    dashboard_json='{
        "dashboard": {
            "id": null,
            "title": "AI Service Overview",
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
                        }
                    ],
                    "gridPos": {"h": 8, "w": 6, "x": 18, "y": 0}
                }
            ],
            "time": {
                "from": "now-1h",
                "to": "now"
            },
            "refresh": "30s"
        }
    }'
    
    echo -n "Importing simple dashboard... "
    
    if [ -n "$API_KEY" ]; then
        response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $API_KEY" \
            -d "{\"dashboard\":$dashboard_json,\"overwrite\":true}" \
            "$GRAFANA_URL/api/dashboards/db" 2>/dev/null)
    else
        response=$(curl -s -X POST \
            -H "Content-Type: application/json" \
            -d "{\"dashboard\":$dashboard_json,\"overwrite\":true}" \
            "$GRAFANA_URL/api/dashboards/db" \
            -u "$GRAFANA_USER:$GRAFANA_PASSWORD" 2>/dev/null)
    fi
    
    if echo "$response" | grep -q "success\|id"; then
        echo -e "${GREEN}Created${NC}"
        return 0
    else
        echo -e "${RED}Failed${NC}"
        return 1
    fi
}

# Main function
main() {
    echo -e "${YELLOW}Importing Grafana dashboards...${NC}"
    echo
    
    # Check Grafana accessibility
    if ! check_grafana; then
        echo -e "${RED}Cannot access Grafana. Make sure it's running and port-forward is active.${NC}"
        echo -e "${YELLOW}   Run: kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000${NC}"
        exit 1
    fi
    
    echo
    
    # Get API key
    get_api_key
    
    echo
    
    # Try to import existing dashboards
    echo -e "${BLUE}Importing existing dashboards...${NC}"
    
    import_dashboard "AI Service Dashboard" "ai-dashboard.json"
    import_dashboard "Minikube Dashboard" "minikube-dashboard.json"
    import_dashboard "Cadence Dashboard" "cadence-dashboard.json"
    
    echo
    
    # Create simple dashboard if imports failed
    echo -e "${BLUE}Creating fallback dashboard...${NC}"
    create_simple_dashboard
    
    echo
    
    # Summary
    echo "================================================"
    echo -e "${GREEN}Dashboard import complete!${NC}"
    echo
    
    echo -e "${BLUE}Access your dashboards:${NC}"
    echo "   Grafana: http://localhost:3000"
    echo "   👤 Login: admin"
    echo "   Password: admin123"
    echo
    
    echo -e "${YELLOW}📝 Manual import instructions:${NC}"
    echo "   1. Go to http://localhost:3000"
    echo "   2. Login with admin/admin123"
    echo "   3. Click '+' → 'Import'"
    echo "   4. Copy dashboard JSON from ConfigMap:"
    echo "      kubectl get configmap grafana-dashboards -n $NAMESPACE -o yaml"
    echo "   5. Paste and import"
    echo
    
    echo -e "${BLUE}Available dashboards in ConfigMap:${NC}"
    kubectl get configmap grafana-dashboards -n $NAMESPACE -o jsonpath='{.data}' | jq -r 'keys[]' 2>/dev/null || echo "   - ai-dashboard.json"
}

# Run main function
main "$@"


