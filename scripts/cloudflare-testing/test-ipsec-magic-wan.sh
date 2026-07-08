#!/bin/bash
set -e

echo "🧪 Тестирование IPsec L3-туннеля Magic WAN..."

# Получение имени пода
POD_NAME=$(kubectl get pods -n cloudbridge-testing -l app=ipsec-magic-wan -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [[ -z "$POD_NAME" ]]; then
    echo "IPsec под не найден. Сначала запустите deploy-ipsec-magic-wan.sh"
    exit 1
fi

echo "Проверка статуса IPsec туннеля в поде $POD_NAME..."

# Проверка статуса IPsec
echo "Статус IPsec соединений:"
kubectl exec -n cloudbridge-testing $POD_NAME -- ipsec statusall

echo ""
echo "Активные соединения:"
kubectl exec -n cloudbridge-testing $POD_NAME -- ipsec listall

echo ""
echo "🛣️ Маршруты через туннель:"
kubectl exec -n cloudbridge-testing $POD_NAME -- ip route show | grep -E "(magic-wan|tunnel)" || echo "   Нет специальных маршрутов через туннель"

echo ""
echo "Тест связности через Magic WAN:"
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo 'Тест ping к Google DNS:'
    if ping -c 3 8.8.8.8; then
        echo 'Ping к 8.8.8.8 успешен'
    else
        echo 'Ping к 8.8.8.8 неудачен'
    fi
    
    echo 'Тест ping к Cloudflare DNS:'
    if ping -c 3 1.1.1.1; then
        echo 'Ping к 1.1.1.1 успешен'
    else
        echo 'Ping к 1.1.1.1 неудачен'
    fi
"

echo ""
echo "Проверка логов IPsec:"
kubectl exec -n cloudbridge-testing $POD_NAME -- tail -20 /var/log/strongswan.log 2>/dev/null || echo "   Логи недоступны"

echo ""
echo "🔧 Проверка конфигурации:"
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo ' Конфигурация IPsec:'
    cat /etc/ipsec/ipsec.conf | head -20
    
    echo ''
    echo 'Проверка PSK (без отображения значения):'
    if [[ -f /etc/ipsec/ipsec.secrets ]]; then
        echo '   Файл PSK существует'
        wc -l /etc/ipsec/ipsec.secrets
    else
        echo '   Файл PSK не найден'
    fi
"

echo ""
echo " Тест внешней связности:"
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo 'Тест HTTP через туннель:'
    if curl -s --connect-timeout 5 https://httpbin.org/ip; then
        echo 'HTTP запрос через туннель успешен'
    else
        echo 'HTTP запрос через туннель неудачен'
    fi
"

echo ""
echo "Статистика туннеля:"
kubectl exec -n cloudbridge-testing $POD_NAME -- /bin/bash -c "
    echo 'Статистика IPsec:'
    ipsec statusall | grep -E '(ESTABLISHED|bytes|packets)' || echo '   Нет активных соединений'
"

echo ""
echo "Тестирование IPsec L3-туннеля завершено!"
echo ""
echo " Если туннель не работает:"
echo "1. Проверьте параметры CF_PEER_IP, CF_PSK, LOCAL_PUBLIC_IP"
echo "2. Убедитесь, что порты UDP 500/4500 открыты"
echo "3. Проверьте логи: kubectl logs -n cloudbridge-testing $POD_NAME"
echo "4. Обратитесь к Cloudflare для получения правильных параметров"
