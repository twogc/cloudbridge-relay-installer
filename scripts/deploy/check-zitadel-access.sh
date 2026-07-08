#!/bin/bash

# Скрипт для проверки доступности Zitadel через Ingress

echo "Проверка доступности Zitadel через Ingress..."
echo "================================================"

# Получаем IP адрес Ingress
INGRESS_IP=$(kubectl get ingress zitadel -n zitadel -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
if [ -z "$INGRESS_IP" ]; then
    INGRESS_IP=$(kubectl get ingress zitadel -n zitadel -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
fi

echo "Ingress IP: $INGRESS_IP"
echo "Host: zitadel.cloudbridge.local"
echo ""

# Проверяем основные эндпоинты
echo "Проверка эндпоинтов:"
echo "------------------------"

# Главная страница
echo -n "• Главная страница: "
if curl -s -o /dev/null -w "%{http_code}" http://$INGRESS_IP -H "Host: zitadel.cloudbridge.local" | grep -q "302\|200"; then
    echo "Доступна"
else
    echo "Недоступна"
fi

# UI Login
echo -n "• UI Login: "
if curl -s -o /dev/null -w "%{http_code}" http://$INGRESS_IP/ui/login/ -H "Host: zitadel.cloudbridge.local" | grep -q "302\|200"; then
    echo "Доступна"
else
    echo "Недоступна"
fi

# UI Console
echo -n "• UI Console: "
if curl -s -o /dev/null -w "%{http_code}" http://$INGRESS_IP/ui/console/ -H "Host: zitadel.cloudbridge.local" | grep -q "200"; then
    echo "Доступна"
else
    echo "Недоступна"
fi

# API Health
echo -n "• API Health: "
if curl -s -o /dev/null -w "%{http_code}" http://$INGRESS_IP/debug/healthz -H "Host: zitadel.cloudbridge.local" | grep -q "200"; then
    echo "Доступна"
else
    echo "Недоступна"
fi

echo ""
echo "Доступ к Zitadel:"
echo "• URL: http://zitadel.cloudbridge.local"
echo "• IP: http://$INGRESS_IP (с заголовком Host: zitadel.cloudbridge.local)"
echo "• Console: http://zitadel.cloudbridge.local/ui/console/"
echo "• Login: http://zitadel.cloudbridge.local/ui/login/"

echo ""
echo "Для добавления в /etc/hosts:"
echo "echo '$INGRESS_IP zitadel.cloudbridge.local' | sudo tee -a /etc/hosts"
