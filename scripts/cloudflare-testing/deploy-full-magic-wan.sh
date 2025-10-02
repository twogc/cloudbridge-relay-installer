#!/bin/bash
set -e

echo "🚀 Полное развертывание Magic WAN с автоматическим созданием через API..."

# Проверка наличия jq
if ! command -v jq &> /dev/null; then
    echo "📦 Установка jq..."
    apt-get update && apt-get install -y jq
fi

# Шаг 1: Создание Magic WAN через API
echo "🔧 Шаг 1: Создание Magic WAN через Cloudflare API..."
./scripts/cloudflare-testing/create-magic-wan-api.sh

# Загрузка параметров
echo "📋 Загрузка параметров Magic WAN..."
source /tmp/magic-wan-params.env

# Проверка параметров
if [[ -z "$CF_PEER_IP" || -z "$CF_PSK" || -z "$LOCAL_PUBLIC_IP" ]]; then
    echo "❌ Ошибка: Параметры Magic WAN не загружены"
    exit 1
fi

echo "✅ Параметры загружены:"
echo "   CF_PEER_IP: $CF_PEER_IP"
echo "   LOCAL_PUBLIC_IP: $LOCAL_PUBLIC_IP"
echo "   TUNNEL_ID: $TUNNEL_ID"

# Шаг 2: Настройка файрвола
echo "🔧 Шаг 2: Настройка файрвола для IPsec..."

# Проверка ufw
if command -v ufw &> /dev/null; then
    echo "🔥 Настройка UFW для IPsec..."
    ufw allow 500/udp comment "IPsec IKE"
    ufw allow 4500/udp comment "IPsec NAT-T"
    ufw allow esp comment "IPsec ESP"
    echo "✅ UFW настроен для IPsec"
else
    echo "⚠️ UFW не найден, настройте файрвол вручную:"
    echo "   - UDP 500/4500 для IPsec"
    echo "   - ESP протокол"
fi

# Шаг 3: Развертывание IPsec туннеля
echo "🔧 Шаг 3: Развертывание IPsec туннеля..."
./scripts/cloudflare-testing/deploy-ipsec-magic-wan.sh

# Шаг 4: Тестирование туннеля
echo "🧪 Шаг 4: Тестирование IPsec туннеля..."
./scripts/cloudflare-testing/test-ipsec-magic-wan.sh

# Шаг 5: Проверка статуса в Cloudflare
echo "📊 Шаг 5: Проверка статуса в Cloudflare..."
./scripts/cloudflare-testing/check-magic-wan-status.sh

# Шаг 6: Интеграция с CloudBridge Relay
echo "🔧 Шаг 6: Настройка CloudBridge Relay для Magic WAN..."

# Обновление конфигурации CloudBridge Relay
kubectl get configmap cloudbridge-magic-wan-config -n cloudbridge-testing -o yaml > /tmp/cloudbridge-config.yaml

# Добавление параметров Magic WAN в конфигурацию
cat >> /tmp/cloudbridge-config.yaml << EOF
  magic_wan_peer_ip: "$CF_PEER_IP"
  magic_wan_psk: "$CF_PSK"
  magic_wan_tunnel_id: "$TUNNEL_ID"
EOF

kubectl apply -f /tmp/cloudbridge-config.yaml

# Перезапуск CloudBridge Relay
kubectl rollout restart deployment/cloudbridge-relay-magic-wan -n cloudbridge-testing

echo "⏳ Ожидание перезапуска CloudBridge Relay..."
kubectl wait --for=condition=Ready pod -l app=cloudbridge-relay-magic-wan -n cloudbridge-testing --timeout=300s

# Финальная проверка
echo "🎉 Развертывание Magic WAN завершено!"
echo ""
echo "📋 Итоговая информация:"
echo "   Magic WAN Tunnel ID: $TUNNEL_ID"
echo "   Cloudflare Endpoint: $CF_PEER_IP"
echo "   Local Public IP: $LOCAL_PUBLIC_IP"
echo ""
echo "🔍 Проверка компонентов:"
echo "   IPsec туннель: kubectl get pods -n cloudbridge-testing -l app=ipsec-magic-wan"
echo "   CloudBridge Relay: kubectl get pods -n cloudbridge-testing -l app=cloudbridge-relay-magic-wan"
echo "   Статус в Cloudflare: ./scripts/cloudflare-testing/check-magic-wan-status.sh"
echo ""
echo "🌐 Тестирование:"
echo "   Локальный тест: ./scripts/cloudflare-testing/test-ipsec-magic-wan.sh"
echo "   CloudBridge тест: kubectl port-forward -n cloudbridge-testing svc/cloudbridge-relay-magic-wan 6003:6003"
echo ""
echo "💡 Следующие шаги:"
echo "1. Настройте DNS записи в Cloudflare"
echo "2. Настройте политики трафика в Magic WAN"
echo "3. Протестируйте end-to-end подключение"
echo "4. Настройте мониторинг и алерты"
