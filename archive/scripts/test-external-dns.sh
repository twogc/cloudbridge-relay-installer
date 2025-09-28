#!/bin/bash

# Тестирование CloudBridge через внешний DNS (2gc.ru домены)
# Этот скрипт проверяет доступность всех сервисов через VK Cloud LoadBalancer

set -e

echo "=== CloudBridge External DNS Test ==="
echo "Testing through VK Cloud LoadBalancer: 109.120.180.160"
echo "Domains: relay.2gc.ru, masque.2gc.ru, p2p.2gc.ru, admin.2gc.ru, monitor.2gc.ru, api.2gc.ru"
echo

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для проверки DNS резолюции
test_dns_resolution() {
    echo -e "${BLUE}=== DNS Resolution Test ===${NC}"
    
    domains=("relay.2gc.ru" "masque.2gc.ru" "p2p.2gc.ru" "admin.2gc.ru" "monitor.2gc.ru" "api.2gc.ru")
    
    for domain in "${domains[@]}"; do
        echo -n "Testing $domain: "
        if nslookup "$domain" > /dev/null 2>&1; then
            ip=$(nslookup "$domain" | grep "Address:" | tail -1 | awk '{print $2}')
            echo -e "${GREEN}✓ Resolved to $ip${NC}"
        else
            echo -e "${RED}✗ Failed to resolve${NC}"
        fi
    done
    echo
}

# Функция для проверки TCP подключения
test_tcp_connectivity() {
    echo -e "${BLUE}=== TCP Connectivity Test ===${NC}"
    
    echo -n "Testing LoadBalancer IP (109.120.180.160:443): "
    if timeout 5 bash -c "</dev/tcp/109.120.180.160/443" 2>/dev/null; then
        echo -e "${GREEN}✓ Port 443 is open${NC}"
    else
        echo -e "${RED}✗ Port 443 is not accessible${NC}"
    fi
    
    domains=("relay.2gc.ru" "masque.2gc.ru" "p2p.2gc.ru" "admin.2gc.ru" "monitor.2gc.ru" "api.2gc.ru")
    
    for domain in "${domains[@]}"; do
        echo -n "Testing $domain:443: "
        if timeout 5 bash -c "</dev/tcp/$domain/443" 2>/dev/null; then
            echo -e "${GREEN}✓ Connected${NC}"
        else
            echo -e "${RED}✗ Connection failed${NC}"
        fi
    done
    echo
}

# Функция для проверки SSL сертификатов
test_ssl_certificates() {
    echo -e "${BLUE}=== SSL Certificate Test ===${NC}"
    
    domains=("relay.2gc.ru" "masque.2gc.ru" "p2p.2gc.ru" "admin.2gc.ru" "monitor.2gc.ru" "api.2gc.ru")
    
    for domain in "${domains[@]}"; do
        echo "Testing SSL for $domain:"
        
        # Проверяем сертификат (игнорируем ошибки самоподписанного сертификата)
        cert_info=$(echo | timeout 10 openssl s_client -connect "$domain:443" -servername "$domain" 2>/dev/null | openssl x509 -noout -subject -dates 2>/dev/null || echo "Certificate check failed")
        
        if [[ "$cert_info" == *"Certificate check failed"* ]]; then
            echo -e "  ${RED}✗ SSL certificate check failed${NC}"
        else
            echo -e "  ${YELLOW}⚠ SSL certificate present (self-signed)${NC}"
            echo "  $cert_info" | head -3
        fi
        echo
    done
}

# Функция для проверки HTTP endpoints
test_http_endpoints() {
    echo -e "${BLUE}=== HTTP Endpoints Test ===${NC}"
    
    # Тестируем основные endpoints
    endpoints=(
        "https://relay.2gc.ru/health"
        "https://masque.2gc.ru/health"
        "https://p2p.2gc.ru/health"
        "https://admin.2gc.ru/health"
        "https://monitor.2gc.ru/health"
        "https://api.2gc.ru/health"
        "https://relay.2gc.ru/metrics"
        "https://monitor.2gc.ru/metrics"
        "https://relay.2gc.ru/api/v1/peers/discover?tenant_id=pilot-tenant"
    )
    
    for endpoint in "${endpoints[@]}"; do
        echo -n "Testing $endpoint: "
        
        # Используем curl с игнорированием SSL ошибок и следованием редиректам
        response=$(curl -k -s -w "%{http_code}|%{redirect_url}|%{time_total}" -o /tmp/response_body.txt "$endpoint" 2>/dev/null || echo "000||0")
        
        http_code=$(echo "$response" | cut -d'|' -f1)
        redirect_url=$(echo "$response" | cut -d'|' -f2)
        time_total=$(echo "$response" | cut -d'|' -f3)
        
        case "$http_code" in
            200)
                echo -e "${GREEN}✓ OK (${time_total}s)${NC}"
                # Показываем первые несколько строк ответа если это JSON
                if head -1 /tmp/response_body.txt 2>/dev/null | grep -q "{"; then
                    echo "  Response: $(head -1 /tmp/response_body.txt | cut -c1-80)..."
                fi
                ;;
            301|302|307|308)
                echo -e "${YELLOW}⚠ Redirect to: $redirect_url${NC}"
                ;;
            401)
                echo -e "${YELLOW}⚠ Authentication required (expected)${NC}"
                ;;
            404)
                echo -e "${YELLOW}⚠ Not found${NC}"
                ;;
            000)
                echo -e "${RED}✗ Connection failed${NC}"
                ;;
            *)
                echo -e "${RED}✗ HTTP $http_code${NC}"
                if [ -s /tmp/response_body.txt ]; then
                    echo "  Error: $(head -1 /tmp/response_body.txt | cut -c1-80)"
                fi
                ;;
        esac
    done
    echo
}

# Функция для анализа редиректов
test_redirect_analysis() {
    echo -e "${BLUE}=== Redirect Analysis ===${NC}"
    
    echo "Analyzing redirect behavior for relay.2gc.ru/health:"
    
    # Тестируем без следования редиректам
    curl -k -s -I "https://relay.2gc.ru/health" 2>/dev/null | head -10 | while read line; do
        if [[ "$line" == *"HTTP"* ]] || [[ "$line" == *"Location:"* ]] || [[ "$line" == *"Server:"* ]]; then
            echo "  $line"
        fi
    done
    
    echo
    echo "Testing direct LoadBalancer IP with Host header:"
    curl -k -s -I -H "Host: relay.2gc.ru" "https://109.120.180.160/health" 2>/dev/null | head -10 | while read line; do
        if [[ "$line" == *"HTTP"* ]] || [[ "$line" == *"Location:"* ]] || [[ "$line" == *"Server:"* ]]; then
            echo "  $line"
        fi
    done
    echo
}

# Функция для компиляции и запуска Go клиента
test_go_client() {
    echo -e "${BLUE}=== Go Client Test ===${NC}"
    
    cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client/examples
    
    echo "Compiling external DNS client..."
    if go build -o external-dns-client external-dns-client.go 2>/dev/null; then
        echo -e "${GREEN}✓ Compilation successful${NC}"
        echo
        echo "Running external DNS client test..."
        ./external-dns-client
    else
        echo -e "${RED}✗ Compilation failed${NC}"
        echo "Error details:"
        go build -o external-dns-client external-dns-client.go
    fi
    
    cd - > /dev/null
}

# Основная функция
main() {
    echo "Starting external DNS testing at $(date)"
    echo
    
    # Проверяем наличие необходимых утилит
    for cmd in nslookup curl openssl go; do
        if ! command -v "$cmd" &> /dev/null; then
            echo -e "${RED}Error: $cmd is not installed${NC}"
            exit 1
        fi
    done
    
    # Запускаем тесты
    test_dns_resolution
    test_tcp_connectivity
    test_ssl_certificates
    test_http_endpoints
    test_redirect_analysis
    test_go_client
    
    echo -e "${BLUE}=== Test Summary ===${NC}"
    echo "External DNS testing completed at $(date)"
    echo
    echo "Key findings:"
    echo "1. Check DNS resolution for all 2gc.ru domains"
    echo "2. Verify LoadBalancer (109.120.180.160:443) accessibility"
    echo "3. Analyze SSL certificate status (self-signed expected)"
    echo "4. Identify any redirect loops or misconfigurations"
    echo "5. Test API endpoints functionality"
    echo
    echo "If you see redirect loops, the issue is likely in the relay server"
    echo "configuration (10.0.0.249) which needs to be updated to handle"
    echo "SSL termination properly."
}

# Запускаем тесты
main "$@"
