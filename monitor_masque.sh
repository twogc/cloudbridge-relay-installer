#!/bin/bash

# Скрипт мониторинга MASQUE HTTP/3 сервера

echo "🔍 Мониторинг MASQUE HTTP/3 сервера"
echo "=================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Функция для проверки статуса
check_status() {
    local name="$1"
    local command="$2"
    local expected="$3"
    
    echo -n "Проверка $name... "
    
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ OK${NC}"
        return 0
    else
        echo -e "${RED}❌ FAIL${NC}"
        return 1
    fi
}

# Функция для проверки с предупреждением
check_warning() {
    local name="$1"
    local command="$2"
    local warning="$3"
    
    echo -n "Проверка $name... "
    
    if eval "$command" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ OK${NC}"
    else
        echo -e "${YELLOW}⚠️  $warning${NC}"
    fi
}

echo "📊 Проверка компонентов:"
echo "------------------------"

# 1. Проверка pod'ов
check_status "Pod'ы CloudBridge Relay" "kubectl get pods -n cloudbridge | grep cloudbridge-relay | grep Running"

# 2. Проверка UDP порта 8443
check_status "UDP порт 8443" "kubectl exec -n cloudbridge \$(kubectl get pods -n cloudbridge | grep cloudbridge-relay | head -1 | awk '{print \$1}') -- netstat -ulnp | grep 8443"

# 3. Проверка UDP подключения
check_status "UDP подключение к NodePort" "nc -u -w 2 192.168.49.2 30443 < /dev/null"

# 4. Проверка сертификата
check_warning "Сертификат Let's Encrypt" "sudo test -f /etc/letsencrypt/live/edge.2gc.ru/fullchain.pem" "Сертификат не найден"

# 5. Проверка логов MASQUE
echo -n "Проверка логов MASQUE... "
if kubectl logs -n cloudbridge $(kubectl get pods -n cloudbridge | grep cloudbridge-relay | head -1 | awk '{print $1}') | grep -q "MASQUE server started successfully"; then
    echo -e "${GREEN}✅ OK${NC}"
else
    echo -e "${RED}❌ FAIL${NC}"
fi

echo ""
echo "📈 Статистика:"
echo "--------------"

# Подсчет pod'ов
POD_COUNT=$(kubectl get pods -n cloudbridge | grep cloudbridge-relay | grep Running | wc -l)
echo "Активных pod'ов: $POD_COUNT"

# Проверка сертификата
if sudo test -f /etc/letsencrypt/live/edge.2gc.ru/fullchain.pem; then
    CERT_EXPIRY=$(sudo openssl x509 -in /etc/letsencrypt/live/edge.2gc.ru/fullchain.pem -noout -dates | grep notAfter | cut -d= -f2)
    echo "Сертификат действителен до: $CERT_EXPIRY"
fi

# Проверка портов
echo "Порты MASQUE:"
kubectl get service cloudbridge-relay-external -n cloudbridge -o jsonpath='{.spec.ports[?(@.name=="masque-proxy")]}' | jq -r '"  - " + .name + ": " + .port + "/" + .protocol + " (NodePort: " + (.nodePort|tostring) + ")"'

echo ""
echo "🎯 Рекомендации:"
echo "----------------"

# Проверка мониторинга
if ! kubectl get pods -n cloudbridge | grep -q prometheus; then
    echo "⚠️  Рекомендуется настроить Prometheus для мониторинга"
fi

# Проверка HTTP/3 клиента
if ! command -v curl --http3 &> /dev/null; then
    echo "⚠️  Рекомендуется установить curl с поддержкой HTTP/3 для полного тестирования"
fi

echo ""
echo "✅ Мониторинг завершен"

