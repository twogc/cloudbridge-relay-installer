#!/bin/bash

# Настройка внешнего балансировщика для домена 2gc.ru
# External IP: 109.120.180.160 (b1.2gc.space)
# Internal IP: 10.0.0.242

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

dns_info() {
    echo -e "${PURPLE}[DNS]${NC} $1"
}

# Конфигурация
EXTERNAL_IP="109.120.180.160"
INTERNAL_IP="10.0.0.242"
DOMAIN="2gc.ru"

# Проверка DNS записей
check_dns_records() {
    info "=== ПРОВЕРКА DNS ЗАПИСЕЙ ==="
    
    # Список доменов для проверки
    DOMAINS=(
        "b1.2gc.space"
        "masque.2gc.space" 
        "relay.2gc.space"
        "relay.2gc.ru"
        "masque.2gc.ru"
        "p2p.2gc.ru"
        "admin.2gc.ru"
        "monitor.2gc.ru"
        "api.2gc.ru"
    )
    
    for domain in "${DOMAINS[@]}"; do
        dns_info "Проверяем $domain..."
        
        # Проверяем A записи
        A_RECORD=$(dig +short A "$domain" 2>/dev/null || echo "")
        if [ -n "$A_RECORD" ]; then
            if [ "$A_RECORD" = "$EXTERNAL_IP" ]; then
                success "  ✓ A запись: $domain -> $A_RECORD"
            else
                warning "  ⚠ A запись: $domain -> $A_RECORD (ожидался $EXTERNAL_IP)"
            fi
        fi
        
        # Проверяем CNAME записи
        CNAME_RECORD=$(dig +short CNAME "$domain" 2>/dev/null || echo "")
        if [ -n "$CNAME_RECORD" ]; then
            success "  ✓ CNAME запись: $domain -> $CNAME_RECORD"
        fi
        
        # Если нет записей
        if [ -z "$A_RECORD" ] && [ -z "$CNAME_RECORD" ]; then
            error "  ✗ DNS запись не найдена для $domain"
        fi
    done
}

# Проверка доступности внешнего балансировщика
test_external_lb() {
    info "=== ТЕСТ ВНЕШНЕГО БАЛАНСИРОВЩИКА ==="
    
    # Тестируем только порт 443 (HTTPS) - все остальные сервисы проксируются через него
    info "Тестируем TCP порт 443 (HTTPS) на $EXTERNAL_IP..."
    
    if timeout 5 bash -c "</dev/tcp/$EXTERNAL_IP/443" 2>/dev/null; then
        success "  ✓ Порт 443 (HTTPS) доступен"
    else
        warning "  ⚠ Порт 443 (HTTPS) недоступен"
    fi
    
    info "Примечание: Все остальные сервисы (gRPC, P2P API, QUIC, STUN) проксируются через HTTPS (443)"
}

# Тестирование HTTP/HTTPS endpoints
test_http_endpoints() {
    info "=== ТЕСТ HTTP/HTTPS ENDPOINTS ==="
    
    # Список endpoints для тестирования
    ENDPOINTS=(
        "http://relay.2gc.ru/health"
        "https://relay.2gc.ru/health"
        "http://masque.2gc.ru/health"
        "https://masque.2gc.ru/health"
        "http://api.2gc.ru/api/v1/health"
        "https://api.2gc.ru/api/v1/health"
        "http://p2p.2gc.ru/api/v1/health"
        "https://p2p.2gc.ru/api/v1/health"
    )
    
    for endpoint in "${ENDPOINTS[@]}"; do
        info "Тестируем $endpoint..."
        
        # Используем curl с timeout
        HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null --connect-timeout 5 --max-time 10 "$endpoint" 2>/dev/null || echo "000")
        
        case $HTTP_CODE in
            200)
                success "  ✓ $endpoint (HTTP $HTTP_CODE)"
                ;;
            301|302|307|308)
                success "  ✓ $endpoint (HTTP $HTTP_CODE - редирект)"
                ;;
            404)
                warning "  ⚠ $endpoint (HTTP $HTTP_CODE - не найден)"
                ;;
            000)
                error "  ✗ $endpoint (недоступен)"
                ;;
            *)
                warning "  ⚠ $endpoint (HTTP $HTTP_CODE)"
                ;;
        esac
    done
}

# Применение конфигурации Kubernetes
apply_k8s_config() {
    info "=== ПРИМЕНЕНИЕ КОНФИГУРАЦИИ KUBERNETES ==="
    
    # Применяем конфигурацию внешнего балансировщика
    if [ -f "k8s/external-loadbalancer-config.yaml" ]; then
        info "Применяем конфигурацию внешнего балансировщика..."
        kubectl apply -f k8s/external-loadbalancer-config.yaml
        success "✓ Конфигурация внешнего балансировщика применена"
    else
        error "✗ Файл конфигурации k8s/external-loadbalancer-config.yaml не найден"
        return 1
    fi
    
    # Ждем готовности сервисов
    info "Ожидаем готовности сервисов..."
    kubectl wait --for=condition=ready pod -l app=cloudbridge-relay -n cloudbridge --timeout=60s
    
    # Проверяем статус LoadBalancer
    info "Проверяем статус LoadBalancer сервисов..."
    kubectl get svc -n cloudbridge -l component=external-loadbalancer -o wide
}

# Создание TLS сертификатов для 2gc.ru
setup_tls_certificates() {
    info "=== НАСТРОЙКА TLS СЕРТИФИКАТОВ ДЛЯ 2GC.RU ==="
    
    CERT_DIR="/etc/ssl/cloudbridge-2gc"
    DOMAIN_CERT_FILE="$CERT_DIR/2gc.ru.crt"
    DOMAIN_KEY_FILE="$CERT_DIR/2gc.ru.key"
    
    # Создаем директорию для сертификатов
    sudo mkdir -p "$CERT_DIR"
    
    # Проверяем существующие сертификаты
    if [ -f "$DOMAIN_CERT_FILE" ] && [ -f "$DOMAIN_KEY_FILE" ]; then
        info "Сертификаты для 2gc.ru уже существуют"
        
        # Проверяем срок действия
        if openssl x509 -checkend 86400 -noout -in "$DOMAIN_CERT_FILE" > /dev/null 2>&1; then
            success "✓ Сертификат валиден еще как минимум 24 часа"
            return 0
        else
            warning "⚠ Срок действия сертификата истекает, создаем новый"
        fi
    fi
    
    # Создаем конфигурационный файл для OpenSSL
    cat > "$CERT_DIR/2gc.ru.conf" << EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow
O = 2GC
OU = IT Department
CN = 2gc.ru
emailAddress = admin@2gc.ru

[v3_req]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = 2gc.ru
DNS.2 = *.2gc.ru
DNS.3 = relay.2gc.ru
DNS.4 = masque.2gc.ru
DNS.5 = p2p.2gc.ru
DNS.6 = admin.2gc.ru
DNS.7 = monitor.2gc.ru
DNS.8 = api.2gc.ru
DNS.9 = b1.2gc.space
DNS.10 = masque.2gc.space
DNS.11 = relay.2gc.space
EOF
    
    # Генерируем приватный ключ
    sudo openssl genrsa -out "$DOMAIN_KEY_FILE" 2048
    
    # Генерируем самоподписанный сертификат
    sudo openssl req -x509 -new -nodes -key "$DOMAIN_KEY_FILE" -sha256 -days 365 -out "$DOMAIN_CERT_FILE" -config "$CERT_DIR/2gc.ru.conf"
    
    # Устанавливаем права доступа
    sudo chmod 644 "$DOMAIN_CERT_FILE"
    sudo chmod 600 "$DOMAIN_KEY_FILE"
    
    success "✓ TLS сертификаты для 2gc.ru созданы"
    
    # Создаем Kubernetes Secret
    info "Создаем Kubernetes Secret для TLS сертификатов..."
    kubectl create secret tls cloudbridge-2gc-ru-tls \
        --cert="$DOMAIN_CERT_FILE" \
        --key="$DOMAIN_KEY_FILE" \
        -n cloudbridge \
        --dry-run=client -o yaml | kubectl apply -f -
    
    success "✓ Kubernetes Secret для TLS создан"
}

# Настройка мониторинга балансировщика
setup_lb_monitoring() {
    info "=== НАСТРОЙКА МОНИТОРИНГА БАЛАНСИРОВЩИКА ==="
    
    # Создаем ServiceMonitor для внешнего балансировщика
    cat > /tmp/external-lb-servicemonitor.yaml << EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cloudbridge-external-lb-monitor
  namespace: cloudbridge
  labels:
    app: cloudbridge-relay
    component: external-loadbalancer
spec:
  selector:
    matchLabels:
      app: cloudbridge-relay
      component: external-loadbalancer
  endpoints:
  - port: metrics
    path: /metrics
    interval: 30s
    scrapeTimeout: 10s
EOF
    
    kubectl apply -f /tmp/external-lb-servicemonitor.yaml
    success "✓ ServiceMonitor для внешнего балансировщика создан"
    
    # Создаем Grafana dashboard для балансировщика
    info "Создаем Grafana dashboard..."
    # Здесь можно добавить создание dashboard через ConfigMap
}

# Тестирование производительности
performance_test() {
    info "=== ТЕСТ ПРОИЗВОДИТЕЛЬНОСТИ ==="
    
    # Простой тест нагрузки с помощью curl
    info "Запускаем тест нагрузки на relay.2gc.ru..."
    
    CONCURRENT_REQUESTS=10
    TOTAL_REQUESTS=100
    
    # Используем xargs для параллельных запросов
    seq 1 $TOTAL_REQUESTS | xargs -n1 -P$CONCURRENT_REQUESTS -I{} \
        curl -s -w "%{http_code},%{time_total},%{time_connect}\n" \
        -o /dev/null "http://relay.2gc.ru/health" > /tmp/perf_results.txt
    
    # Анализируем результаты
    SUCCESS_COUNT=$(grep "^200," /tmp/perf_results.txt | wc -l)
    TOTAL_COUNT=$(wc -l < /tmp/perf_results.txt)
    SUCCESS_RATE=$(echo "scale=2; $SUCCESS_COUNT * 100 / $TOTAL_COUNT" | bc -l)
    
    AVG_RESPONSE_TIME=$(grep "^200," /tmp/perf_results.txt | cut -d',' -f2 | \
        awk '{sum+=$1} END {print sum/NR}')
    
    success "Результаты теста производительности:"
    echo "  Успешных запросов: $SUCCESS_COUNT из $TOTAL_COUNT ($SUCCESS_RATE%)"
    echo "  Среднее время ответа: ${AVG_RESPONSE_TIME}s"
    
    rm -f /tmp/perf_results.txt
}

# Генерация отчета
generate_report() {
    info "=== ГЕНЕРАЦИЯ ОТЧЕТА ==="
    
    REPORT_FILE="external-lb-2gc-report-$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# CloudBridge External LoadBalancer Report (2gc.ru)

**Дата:** $(date)  
**Внешний IP:** $EXTERNAL_IP  
**Внутренний IP:** $INTERNAL_IP  
**Домен:** $DOMAIN  

## DNS Конфигурация

### Существующие записи:
- b1.2gc.space A $EXTERNAL_IP
- masque.2gc.space CNAME b1.2gc.space  
- relay.2gc.space CNAME b1.2gc.space

### Рекомендуемые записи для 2gc.ru:
- relay.2gc.ru A $EXTERNAL_IP
- masque.2gc.ru A $EXTERNAL_IP
- p2p.2gc.ru A $EXTERNAL_IP
- admin.2gc.ru A $EXTERNAL_IP
- monitor.2gc.ru A $EXTERNAL_IP
- api.2gc.ru A $EXTERNAL_IP

## Kubernetes Services

$(kubectl get svc -n cloudbridge -o wide)

## LoadBalancer Status

$(kubectl get svc -n cloudbridge -l component=external-loadbalancer -o wide)

## Ingress Configuration

$(kubectl get ingress -n cloudbridge -o wide)

## TLS Certificates

$(kubectl get secrets -n cloudbridge | grep tls)

## Рекомендации

1. **DNS записи**: Создать A записи для всех поддоменов 2gc.ru
2. **TLS сертификаты**: Получить валидные SSL сертификаты от Let's Encrypt
3. **Мониторинг**: Настроить алерты для балансировщика
4. **Производительность**: Настроить кеширование и оптимизацию
5. **Безопасность**: Настроить WAF и DDoS защиту
6. **Backup**: Настроить резервное копирование конфигурации

## Команды для создания DNS записей

\`\`\`bash
# Создать A записи (выполнить в панели управления DNS)
relay.2gc.ru    A    $EXTERNAL_IP
masque.2gc.ru   A    $EXTERNAL_IP  
p2p.2gc.ru      A    $EXTERNAL_IP
admin.2gc.ru    A    $EXTERNAL_IP
monitor.2gc.ru  A    $EXTERNAL_IP
api.2gc.ru      A    $EXTERNAL_IP
\`\`\`

EOF

    success "Отчет сохранен: $REPORT_FILE"
}

# Основная функция
main() {
    info "CloudBridge External LoadBalancer Setup for 2gc.ru"
    info "================================================="
    info "External IP: $EXTERNAL_IP"
    info "Internal IP: $INTERNAL_IP"
    info "Domain: $DOMAIN"
    echo ""
    
    check_dns_records
    echo ""
    test_external_lb  
    echo ""
    test_http_endpoints
    echo ""
    setup_tls_certificates
    echo ""
    apply_k8s_config
    echo ""
    setup_lb_monitoring
    echo ""
    performance_test
    echo ""
    generate_report
    
    success "Настройка внешнего балансировщика для 2gc.ru завершена!"
    
    info "Следующие шаги:"
    echo "1. Создать DNS записи для поддоменов 2gc.ru"
    echo "2. Получить валидные SSL сертификаты"
    echo "3. Настроить мониторинг и алерты"
    echo "4. Протестировать все endpoints"
}

# Запуск
main "$@"
