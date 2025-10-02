#!/bin/bash
set -e

echo "🚀 Развертывание IPsec L3-туннеля для Magic WAN..."

# Проверка наличия необходимых параметров
echo "📋 Проверка параметров Magic WAN:"
echo "   CF_PEER_IP: ${CF_PEER_IP:-'НЕ УСТАНОВЛЕН'}"
echo "   CF_PSK: ${CF_PSK:-'НЕ УСТАНОВЛЕН'}"
echo "   LOCAL_PUBLIC_IP: ${LOCAL_PUBLIC_IP:-'НЕ УСТАНОВЛЕН'}"

if [[ -z "$CF_PEER_IP" || -z "$CF_PSK" || -z "$LOCAL_PUBLIC_IP" ]]; then
    echo "❌ Ошибка: Необходимо установить переменные окружения:"
    echo "   export CF_PEER_IP='<PEER_IP_FROM_CLOUDFLARE>'"
    echo "   export CF_PSK='<PSK_FROM_CLOUDFLARE>'"
    echo "   export LOCAL_PUBLIC_IP='<YOUR_PUBLIC_IP>'"
    echo ""
    echo "💡 Получите эти параметры от вашего TAM/SE Cloudflare"
    exit 1
fi

# Применение манифестов
echo "📦 Применение Kubernetes манифестов..."

kubectl apply -f k8s/cloudflare-testing/ipsec-magic-wan-configmap.yaml
kubectl apply -f k8s/cloudflare-testing/ipsec-magic-wan-deployment.yaml
kubectl apply -f k8s/cloudflare-testing/ipsec-magic-wan-service.yaml

# Ожидание запуска пода
echo "⏳ Ожидание запуска IPsec пода..."
kubectl wait --for=condition=Ready pod -l app=ipsec-magic-wan -n cloudbridge-testing --timeout=300s

# Получение имени пода
POD_NAME=$(kubectl get pods -n cloudbridge-testing -l app=ipsec-magic-wan -o jsonpath='{.items[0].metadata.name}')

echo "🔧 Настройка IPsec в поде $POD_NAME..."

# Установка strongSwan и настройка IPsec
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo '📦 Установка strongSwan...'
    apt-get update
    apt-get install -y strongswan strongswan-swanctl
    
    echo '🔧 Настройка IPsec...'
    # Замена переменных в конфигурации
    envsubst < /etc/ipsec/ipsec.conf.template > /etc/ipsec/ipsec.conf
    envsubst < /etc/ipsec/ipsec.secrets.template > /etc/ipsec/ipsec.secrets
    
    # Установка прав доступа
    chmod 600 /etc/ipsec/ipsec.secrets
    chown root:root /etc/ipsec/ipsec.secrets
    
    echo '🚀 Запуск strongSwan...'
    systemctl enable strongswan
    systemctl start strongswan
    
    echo '⏳ Ожидание установки туннеля...'
    sleep 10
    
    echo '📊 Статус IPsec:'
    ipsec statusall
"

# Проверка статуса туннеля
echo "🧪 Проверка статуса IPsec туннеля..."
kubectl exec -n cloudbridge-testing $POD_NAME -- ipsec statusall

# Тест связности
echo "🌐 Тест связности через Magic WAN..."
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo '🔍 Проверка маршрутов:'
    ip route show | grep -E '(magic-wan|tunnel)'
    
    echo '🌐 Тест ping через туннель:'
    if ping -c 3 8.8.8.8; then
        echo '✅ Связность через Magic WAN работает'
    else
        echo '❌ Проблемы со связностью через Magic WAN'
    fi
"

echo "✅ IPsec L3-туннель для Magic WAN развернут!"
echo ""
echo "📋 Следующие шаги:"
echo "1. Проверьте статус туннеля: kubectl exec -n cloudbridge-testing $POD_NAME -- ipsec statusall"
echo "2. Проверьте логи: kubectl logs -n cloudbridge-testing $POD_NAME"
echo "3. Настройте маршрутизацию в Cloudflare Dashboard"
echo "4. Протестируйте трафик через туннель"
