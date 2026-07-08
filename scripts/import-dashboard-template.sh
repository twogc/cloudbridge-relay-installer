#!/bin/bash

# Dashboard Template Import Script
# Imports the AI Service dashboard template into Grafana

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
DASHBOARD_FILE="grafana-dashboard-template.json"

echo -e "${BLUE}AI Service Dashboard Template Import${NC}"
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

# Function to check if dashboard file exists
check_dashboard_file() {
    echo -n "Checking dashboard template file... "
    
    if [ -f "$DASHBOARD_FILE" ]; then
        echo -e "${GREEN}Found${NC}"
        return 0
    else
        echo -e "${RED}Not found${NC}"
        return 1
    fi
}

# Function to import dashboard
import_dashboard() {
    echo -n "Importing AI Service dashboard... "
    
    # Import dashboard using curl
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d @"$DASHBOARD_FILE" \
        "$GRAFANA_URL/api/dashboards/db" \
        -u "$GRAFANA_USER:$GRAFANA_PASSWORD" 2>/dev/null)
    
    if echo "$response" | grep -q "success\|id"; then
        echo -e "${GREEN}Imported successfully${NC}"
        return 0
    else
        echo -e "${RED}Failed to import${NC}"
        echo "Response: $response"
        return 1
    fi
}

# Function to show manual import instructions
show_manual_instructions() {
    echo
    echo -e "${YELLOW}Manual Import Instructions:${NC}"
    echo "================================================"
    echo
    echo "1. Open Grafana in your browser:"
    echo "   http://localhost:3000"
    echo
    echo "2. Login with credentials:"
    echo "   👤 Username: admin"
    echo "   Password: admin123"
    echo
    echo "3. Click the '+' button in the left sidebar"
    echo
    echo "4. Select 'Import'"
    echo
    echo "5. Copy the contents of the dashboard file:"
    echo "   $DASHBOARD_FILE"
    echo
    echo "6. Paste the JSON into the 'Import via panel json' text area"
    echo
    echo "7. Click 'Load'"
    echo
    echo "8. Configure the dashboard settings and click 'Import'"
    echo
    echo -e "${BLUE}Dashboard Features:${NC}"
    echo "   AI Service Status"
    echo "   Request Rate"
    echo "   Response Time"
    echo "   CPU Usage"
    echo "   Memory Usage"
    echo "   Anomaly Detection Rate"
    echo "   Prediction Accuracy"
    echo "   Active Alerts"
    echo "   Error Rate"
    echo "   Throughput"
    echo
}

# Function to show dashboard content
show_dashboard_content() {
    echo -e "${YELLOW} Dashboard Template Content:${NC}"
    echo "================================================"
    echo
    cat "$DASHBOARD_FILE"
    echo
}

# Main function
main() {
    echo -e "${YELLOW}Importing AI Service Dashboard Template...${NC}"
    echo
    
    # Check prerequisites
    if ! check_grafana; then
        echo -e "${RED}Cannot access Grafana. Make sure it's running and port-forward is active.${NC}"
        echo -e "${YELLOW}   Run: kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000${NC}"
        echo
        show_manual_instructions
        exit 1
    fi
    
    if ! check_dashboard_file; then
        echo -e "${RED}Dashboard template file not found.${NC}"
        echo -e "${YELLOW}   Make sure $DASHBOARD_FILE exists in the current directory.${NC}"
        exit 1
    fi
    
    echo
    
    # Try to import dashboard
    if import_dashboard; then
        echo
        echo -e "${GREEN}Dashboard imported successfully!${NC}"
        echo
        echo -e "${BLUE}Access your dashboard:${NC}"
        echo "   Grafana: http://localhost:3000"
        echo "   👤 Login: admin"
        echo "   Password: admin123"
        echo
        echo -e "${YELLOW}📝 The dashboard should now be available in the 'Dashboards' section.${NC}"
    else
        echo
        echo -e "${YELLOW}  Automatic import failed. Using manual import method.${NC}"
        show_manual_instructions
    fi
    
    echo
    echo -e "${BLUE}Additional Information:${NC}"
    echo "   - Dashboard file: $DASHBOARD_FILE"
    echo "   - Grafana URL: $GRAFANA_URL"
    echo "   - Namespace: cloudbridge-monitoring"
    echo
    echo -e "${GREEN}Import process complete!${NC}"
}

# Run main function
main "$@"


