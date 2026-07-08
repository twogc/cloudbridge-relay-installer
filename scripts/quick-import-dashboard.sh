#!/bin/bash

# Quick Dashboard Import Script
# Shows dashboard JSON for manual import

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Quick Dashboard Import${NC}"
echo "================================================"

# Function to show dashboard JSON
show_dashboard() {
    local dashboard_name="$1"
    local configmap_key="$2"
    
    echo -e "${YELLOW}$dashboard_name Dashboard JSON:${NC}"
    echo "================================================"
    
    kubectl get configmap grafana-dashboards -n cloudbridge-monitoring -o jsonpath="{.data.$configmap_key}" | jq .
    
    echo
    echo -e "${BLUE}Instructions:${NC}"
    echo "1. Copy the JSON above"
    echo "2. Go to http://localhost:3000"
    echo "3. Login with admin/admin123"
    echo "4. Click '+' → 'Import'"
    echo "5. Paste the JSON and click 'Load'"
    echo
}

# Main function
main() {
    echo -e "${YELLOW}Available Dashboards:${NC}"
    echo
    
    # Show AI Service Dashboard
    show_dashboard "AI Service" "ai-dashboard.json"
    
    echo -e "${GREEN}Dashboard JSON ready for import!${NC}"
    echo
    echo -e "${BLUE}Grafana Access:${NC}"
    echo "   URL: http://localhost:3000"
    echo "   Login: admin"
    echo "   Password: admin123"
}

# Run main function
main "$@"


