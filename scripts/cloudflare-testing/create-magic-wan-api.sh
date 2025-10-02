#!/bin/bash
set -e

echo "🚀 Создание Magic WAN через Cloudflare API..."

# API параметры
CF_TOKEN="gBGCQ-8qBu49xHIrcxJIki76AjeTEa5a0okZ_LP1"
ACC_ID="073ae0130b7cee5e55a1ac1a335431a8"
API_BASE="https://api.cloudflare.com/client/v4/accounts/$ACC_ID"

# Получение публичного IP
echo "🔍 Получение публичного IP адреса..."
PUBLIC_IP=$(curl -s https://api.ipify.org)
echo "   Публичный IP: $PUBLIC_IP"

# Создание IPsec туннеля
echo "🔧 Создание IPsec туннеля Magic WAN..."

TUNNEL_RESPONSE=$(curl -s -X POST "$API_BASE/magic/ipsec_tunnels" \
  -H "Authorization: Bearer $CF_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"cloudbridge-russia-eu-01\",
    \"cloudflare_endpoint\": \"162.159.75.92\",
    \"customer_endpoint\": \"$PUBLIC_IP\",
    \"interface_address\": \"10.200.0.0/31\",
    \"ike_version\": 2
  }")

echo "📋 Ответ создания туннеля:"
echo "$TUNNEL_RESPONSE" | jq .

# Извлечение tunnel_id
TUNNEL_ID=$(echo "$TUNNEL_RESPONSE" | jq -r '.result.id // empty')
if [[ -z "$TUNNEL_ID" || "$TUNNEL_ID" == "null" ]]; then
    echo "❌ Ошибка: Не удалось получить tunnel_id"
    echo "   Проверьте API токен и права доступа"
    exit 1
fi

echo "✅ IPsec туннель создан с ID: $TUNNEL_ID"

# Генерация PSK
echo "🔐 Генерация PSK для туннеля $TUNNEL_ID..."

PSK_RESPONSE=$(curl -s -X POST "$API_BASE/magic/ipsec_tunnels/$TUNNEL_ID/psk_generate" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "📋 Ответ генерации PSK:"
echo "$PSK_RESPONSE" | jq .

# Извлечение PSK
PSK=$(echo "$PSK_RESPONSE" | jq -r '.result.psk // empty')
if [[ -z "$PSK" || "$PSK" == "null" ]]; then
    echo "❌ Ошибка: Не удалось получить PSK"
    exit 1
fi

echo "✅ PSK сгенерирован: $PSK"

# Получение информации о туннеле
echo "📊 Получение информации о туннеле..."

TUNNEL_INFO=$(curl -s -X GET "$API_BASE/magic/ipsec_tunnels/$TUNNEL_ID" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "📋 Информация о туннеле:"
echo "$TUNNEL_INFO" | jq .

# Извлечение Cloudflare endpoint
CF_ENDPOINT=$(echo "$TUNNEL_INFO" | jq -r '.result.cloudflare_endpoint // empty')
if [[ -z "$CF_ENDPOINT" || "$CF_ENDPOINT" == "null" ]]; then
    echo "❌ Ошибка: Не удалось получить Cloudflare endpoint"
    exit 1
fi

echo "✅ Cloudflare endpoint: $CF_ENDPOINT"

# Создание GRE туннеля (опционально)
echo "🔧 Создание GRE туннеля (опционально)..."

GRE_RESPONSE=$(curl -s -X POST "$API_BASE/magic/gre_tunnels" \
  -H "Authorization: Bearer $CF_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"gre-russia-eu-01\",
    \"cloudflare_endpoint\": \"172.64.251.46\",
    \"customer_endpoint\": \"$PUBLIC_IP\",
    \"interface_address\": \"10.200.2.0/31\",
    \"ttl\": 64
  }")

echo "📋 Ответ создания GRE туннеля:"
echo "$GRE_RESPONSE" | jq .

GRE_TUNNEL_ID=$(echo "$GRE_RESPONSE" | jq -r '.result.id // empty')
if [[ -n "$GRE_TUNNEL_ID" && "$GRE_TUNNEL_ID" != "null" ]]; then
    echo "✅ GRE туннель создан с ID: $GRE_TUNNEL_ID"
else
    echo "⚠️ GRE туннель не создан (возможно, не поддерживается в вашем аккаунте)"
fi

# Настройка маршрутов
echo "🛣️ Настройка маршрутов..."

# Получение списка доступных префиксов
echo "📋 Получение списка префиксов..."
PREFIXES_RESPONSE=$(curl -s -X GET "$API_BASE/magic/prefixes" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "📋 Доступные префиксы:"
echo "$PREFIXES_RESPONSE" | jq .

# Создание статического маршрута
echo "🔧 Создание статического маршрута..."

ROUTE_RESPONSE=$(curl -s -X POST "$API_BASE/magic/static_routes" \
  -H "Authorization: Bearer $CF_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"cloudbridge-internal\",
    \"prefix\": \"10.0.0.0/24\",
    \"nexthop\": \"$TUNNEL_ID\",
    \"priority\": 100
  }")

echo "📋 Ответ создания маршрута:"
echo "$ROUTE_RESPONSE" | jq .

# Сохранение параметров в файл
echo "💾 Сохранение параметров Magic WAN..."

cat > /tmp/magic-wan-params.env << EOF
# Magic WAN параметры
export CF_PEER_IP="$CF_ENDPOINT"
export CF_PSK="$PSK"
export LOCAL_PUBLIC_IP="$PUBLIC_IP"
export TUNNEL_ID="$TUNNEL_ID"
export GRE_TUNNEL_ID="$GRE_TUNNEL_ID"

# API параметры
export CF_TOKEN="$CF_TOKEN"
export ACC_ID="$ACC_ID"
EOF

echo "✅ Параметры сохранены в /tmp/magic-wan-params.env"

# Вывод итоговой информации
echo ""
echo "🎉 Magic WAN успешно создан!"
echo ""
echo "📋 Параметры для настройки IPsec:"
echo "   CF_PEER_IP: $CF_ENDPOINT"
echo "   CF_PSK: $PSK"
echo "   LOCAL_PUBLIC_IP: $PUBLIC_IP"
echo "   TUNNEL_ID: $TUNNEL_ID"
echo ""
echo "🚀 Следующие шаги:"
echo "1. Загрузите параметры: source /tmp/magic-wan-params.env"
echo "2. Запустите IPsec туннель: ./scripts/cloudflare-testing/deploy-ipsec-magic-wan.sh"
echo "3. Протестируйте: ./scripts/cloudflare-testing/test-ipsec-magic-wan.sh"
echo ""
echo "💡 Для проверки статуса туннеля в Cloudflare Dashboard:"
echo "   https://dash.cloudflare.com/$ACC_ID/magic-wan"
