#!/bin/bash
set -e

echo "Проверка статуса Magic WAN..."

# API параметры
CF_TOKEN="gBGCQ-8qBu49xHIrcxJIki76AjeTEa5a0okZ_LP1"
ACC_ID="073ae0130b7cee5e55a1ac1a335431a8"
API_BASE="https://api.cloudflare.com/client/v4/accounts/$ACC_ID"

# Проверка IPsec туннелей
echo "Список IPsec туннелей:"
IPSEC_TUNNELS=$(curl -s -X GET "$API_BASE/magic/ipsec_tunnels" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "$IPSEC_TUNNELS" | jq '.result[] | {id: .id, name: .name, status: .status, cloudflare_endpoint: .cloudflare_endpoint, customer_endpoint: .customer_endpoint}'

# Проверка GRE туннелей
echo "Список GRE туннелей:"
GRE_TUNNELS=$(curl -s -X GET "$API_BASE/magic/gre_tunnels" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "$GRE_TUNNELS" | jq '.result[] | {id: .id, name: .name, status: .status, cloudflare_endpoint: .cloudflare_endpoint, customer_endpoint: .customer_endpoint}'

# Проверка статических маршрутов
echo "🛣️ Статические маршруты:"
STATIC_ROUTES=$(curl -s -X GET "$API_BASE/magic/static_routes" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "$STATIC_ROUTES" | jq '.result[] | {id: .id, name: .name, prefix: .prefix, nexthop: .nexthop, priority: .priority}'

# Проверка префиксов
echo "Доступные префиксы:"
PREFIXES=$(curl -s -X GET "$API_BASE/magic/prefixes" \
  -H "Authorization: Bearer $CF_TOKEN")

echo "$PREFIXES" | jq '.result[] | {id: .id, prefix: .prefix, type: .type, status: .status}'

# Проверка BGP сессий (если есть)
echo "BGP сессии:"
BGP_SESSIONS=$(curl -s -X GET "$API_BASE/magic/bgp_sessions" \
  -H "Authorization: Bearer $CF_TOKEN" 2>/dev/null || echo '{"result": []}')

echo "$BGP_SESSIONS" | jq '.result[] | {id: .id, name: .name, status: .status, asn: .asn}'

# Проверка сайтов Magic WAN
echo " Сайты Magic WAN:"
SITES=$(curl -s -X GET "$API_BASE/magic/sites" \
  -H "Authorization: Bearer $CF_TOKEN" 2>/dev/null || echo '{"result": []}')

echo "$SITES" | jq '.result[] | {id: .id, name: .name, status: .status, location: .location}'

# Проверка аналитики (если доступно)
echo "Аналитика Magic WAN:"
ANALYTICS=$(curl -s -X GET "$API_BASE/magic/analytics" \
  -H "Authorization: Bearer $CF_TOKEN" 2>/dev/null || echo '{"result": {}}')

echo "$ANALYTICS" | jq '.result'

echo ""
echo "Проверка статуса Magic WAN завершена!"
echo ""
echo " Для детальной информации посетите:"
echo "   https://dash.cloudflare.com/$ACC_ID/magic-wan"
