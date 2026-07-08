#!/bin/bash

# Script to update DNS records to new LoadBalancer IP
# New LoadBalancer: 109.120.181.108
# Old LoadBalancer: 109.120.180.160

echo "=== CloudBridge DNS Update Script ==="
echo "Updating DNS records to new LoadBalancer IP: 109.120.181.108"
echo

# New LoadBalancer details
NEW_LB_IP="109.120.181.108"
OLD_LB_IP="109.120.180.160"

# Domains to update
DOMAINS=(
    "relay.2gc.ru"
    "masque.2gc.ru"
    "p2p.2gc.ru"
    "admin.2gc.ru"
    "monitor.2gc.ru"
    "api.2gc.ru"
)

echo "=== Current DNS Status ==="
echo "Checking current DNS records..."
for domain in "${DOMAINS[@]}"; do
    current_ip=$(nslookup "$domain" | grep "Address:" | tail -1 | awk '{print $2}' 2>/dev/null || echo "Not resolved")
    echo "$domain -> $current_ip"
done
echo

echo "=== DNS Update Instructions ==="
echo "Please update the following DNS A records:"
echo
for domain in "${DOMAINS[@]}"; do
    echo "$domain    A    $NEW_LB_IP"
done
echo

echo "=== LoadBalancer Comparison ==="
echo "OLD LoadBalancer (BROKEN):"
echo "  External IP: $OLD_LB_IP"
echo "  Internal IP: 10.0.0.242"
echo "  Subnet: 10.0.0.0/24"
echo "  Status: Cross-subnet issues"
echo
echo "NEW LoadBalancer (CONFIGURED):"
echo "  External IP: $NEW_LB_IP"
echo "  Internal IP: 192.168.49.8"
echo "  Subnet: 192.168.49.0/24 (same as Kubernetes)"
echo "  Status: Proper architecture, but still 503 errors"
echo

echo "=== Testing Commands ==="
echo "After DNS update, test with:"
echo "curl -k https://relay.2gc.ru/health"
echo "curl -k https://masque.2gc.ru/health"
echo "curl -k https://api.2gc.ru/health"
echo

echo "=== Firewall Rules Added ==="
echo "Security Group: 811039da-7d5a-4986-84d0-cc8eefc67e69"
echo "✓ Port 22 (SSH)"
echo "✓ Port 80 (HTTP)"
echo "✓ Port 443 (HTTPS)"
echo "✓ Port 8080 (Nginx Proxy) - NEW"
echo "✓ Port 10050 (Zabbix)"
echo "✓ Port 32500 (Kubernetes NodePort) - NEW"
echo

echo "=== Current Issues ==="
echo "LoadBalancer still returns 503 Service Unavailable"
echo "Member health check may be failing"
echo "Possible VK Cloud LoadBalancer internal issues"
echo

echo "=== Recommended Actions ==="
echo "1. Update DNS records to new IP ($NEW_LB_IP)"
echo "2. Test HTTPS endpoints after DNS propagation"
echo "3. If still 503, contact VK Cloud support"
echo "4. Consider using direct Kubernetes access for pilot"
echo

echo "=== Pilot Project Fallback ==="
echo "If LoadBalancer issues persist, use direct access:"
echo "Base URL: http://192.168.49.2:32500"
echo "Health: http://192.168.49.2:32500/health"
echo "Metrics: http://192.168.49.2:32500/metrics"
echo

echo "DNS update script completed."
echo "Manual DNS update required in your DNS provider control panel."
