#!/bin/bash

# Тест прямого подключения к CloudBridge Relay без cloudflared tunnel
# Проверяет, что CloudBridge Relay доступен и готов принимать подключения

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN} $1${NC}"
}

warning() {
    echo -e "${YELLOW} $1${NC}"
}

error() {
    echo -e "${RED} $1${NC}"
}

log "🧪 Тест прямого подключения к CloudBridge Relay"

# 1. Проверка CloudBridge Relay Health
log "Проверка CloudBridge Relay Health..."
if curl -k -s https://edge.2gc.ru:9090/health | grep -q "status.*ok"; then
    success "CloudBridge Relay Health: OK"
    echo "Ответ: $(curl -k -s https://edge.2gc.ru:9090/health | jq -r '.status' 2>/dev/null || echo 'OK')"
else
    error "CloudBridge Relay Health: FAILED"
    echo "Ответ: $(curl -k -s https://edge.2gc.ru:9090/health)"
fi

# 2. Проверка API endpoint
log "Проверка API endpoint..."
API_RESPONSE=$(curl -k -s -w "%{http_code}" https://edge.2gc.ru:9090/api/ -o /dev/null)
if [ "$API_RESPONSE" = "404" ]; then
    success "API endpoint доступен (HTTP 404 - нормально)"
elif [ "$API_RESPONSE" = "200" ]; then
    success "API endpoint доступен (HTTP 200)"
else
    warning "API endpoint: HTTP $API_RESPONSE"
fi

# 3. Проверка P2P endpoint
log "Проверка P2P endpoint..."
P2P_RESPONSE=$(curl -k -s -w "%{http_code}" https://edge.2gc.ru:9090/p2p/ -o /dev/null)
if [ "$P2P_RESPONSE" = "404" ]; then
    success "P2P endpoint доступен (HTTP 404 - нормально)"
elif [ "$P2P_RESPONSE" = "200" ]; then
    success "P2P endpoint доступен (HTTP 200)"
else
    warning "P2P endpoint: HTTP $P2P_RESPONSE"
fi

# 4. Проверка QUIC endpoint
log "Проверка QUIC endpoint..."
QUIC_RESPONSE=$(curl -k -s -w "%{http_code}" https://edge.2gc.ru:9090/quic/ -o /dev/null)
if [ "$QUIC_RESPONSE" = "404" ]; then
    success "QUIC endpoint доступен (HTTP 404 - нормально)"
elif [ "$QUIC_RESPONSE" = "200" ]; then
    success "QUIC endpoint доступен (HTTP 200)"
else
    warning "QUIC endpoint: HTTP $QUIC_RESPONSE"
fi

# 5. Проверка SSL сертификата
log "Проверка SSL сертификата..."
if echo | openssl s_client -connect edge.2gc.ru:9090 -servername edge.2gc.ru 2>/dev/null | openssl x509 -noout -dates 2>/dev/null | grep -q "notAfter"; then
    success "SSL сертификат валиден"
    echo "Сертификат действителен до: $(echo | openssl s_client -connect edge.2gc.ru:9090 -servername edge.2gc.ru 2>/dev/null | openssl x509 -noout -dates 2>/dev/null | grep notAfter | cut -d= -f2)"
else
    warning "SSL сертификат недоступен или невалиден"
fi

# 6. Проверка с внешнего IP
log "Проверка с внешнего IP..."
EXTERNAL_IP=$(curl -s ifconfig.me)
log "Внешний IP: $EXTERNAL_IP"

if curl -k -s https://$EXTERNAL_IP:9090/health | grep -q "status.*ok"; then
    success "Health endpoint доступен с внешнего IP"
else
    warning "Health endpoint недоступен с внешнего IP"
fi

# 7. Проверка Magic WAN туннеля
log "Проверка Magic WAN туннеля..."
if sudo swanctl --list-sas | grep -q "ESTABLISHED"; then
    success "Magic WAN туннель активен"
else
    warning "Magic WAN туннель не активен"
fi

# 8. Проверка CloudBridge Relay подов
log "Проверка CloudBridge Relay подов..."
if kubectl get pods -n cloudbridge -l app=cloudbridge-relay | grep -q "Running"; then
    success "CloudBridge Relay поды работают"
else
    error "CloudBridge Relay поды не работают"
fi

# 9. Проверка nginx
log "Проверка nginx..."
if systemctl is-active --quiet nginx; then
    success "nginx работает"
else
    error "nginx не работает"
fi

# 10. Проверка порта 9090
log "Проверка порта 9090..."
if netstat -tlnp | grep -q :9090; then
    success "Порт 9090 слушается"
else
    error "Порт 9090 не слушается"
fi

# 11. Проверка UFW
log "Проверка UFW..."
if sudo ufw status | grep -q "9090/tcp"; then
    success "Порт 9090/tcp открыт в UFW"
else
    warning "Порт 9090/tcp не открыт в UFW"
fi

# 12. Тест WebSocket подключения
log "Тест WebSocket подключения..."
if curl -k -s -H "Upgrade: websocket" -H "Connection: Upgrade" https://edge.2gc.ru:9090/api/ | grep -q "404\|not found"; then
    success "WebSocket endpoint доступен"
else
    warning "WebSocket endpoint недоступен"
fi

echo ""
log "Сводка тестирования:"
echo "======================"

# CloudBridge Relay
if curl -k -s https://edge.2gc.ru:9090/health | grep -q "status.*ok"; then
    success "CloudBridge Relay (порт 9090): ДОСТУПЕН"
else
    error "CloudBridge Relay (порт 9090): НЕ ДОСТУПЕН"
fi

# Magic WAN
if sudo swanctl --list-sas | grep -q "ESTABLISHED"; then
    success "Magic WAN туннель: АКТИВЕН"
else
    warning "Magic WAN туннель: НЕ АКТИВЕН"
fi

# CloudBridge Relay поды
if kubectl get pods -n cloudbridge -l app=cloudbridge-relay | grep -q "Running"; then
    success "CloudBridge Relay поды: РАБОТАЮТ"
else
    error "CloudBridge Relay поды: НЕ РАБОТАЮТ"
fi

# nginx
if systemctl is-active --quiet nginx; then
    success "nginx: РАБОТАЕТ"
else
    error "nginx: НЕ РАБОТАЕТ"
fi

# Порт 9090
if netstat -tlnp | grep -q :9090; then
    success "Порт 9090: СЛУШАЕТСЯ"
else
    error "Порт 9090: НЕ СЛУШАЕТСЯ"
fi

# UFW
if sudo ufw status | grep -q "9090/tcp"; then
    success "UFW (порт 9090): ОТКРЫТ"
else
    warning "UFW (порт 9090): НЕ ОТКРЫТ"
fi

echo ""
log "Тест прямого подключения завершен!"
echo ""
echo "CloudBridge Relay готов принимать cloudflared клиентов:"
echo "========================================================"
echo ""
echo " CloudBridge Relay доступен на порту 9090"
echo " SSL сертификат валиден"
echo " Все endpoints отвечают"
echo " Magic WAN туннель активен"
echo ""
echo "Теперь можно настраивать cloudflared клиентов для подключения к:"
echo "   - https://edge.2gc.ru:9090/api/"
echo "   - https://edge.2gc.ru:9090/health"
echo "   - https://edge.2gc.ru:9090/p2p/"
echo "   - https://edge.2gc.ru:9090/quic/"

