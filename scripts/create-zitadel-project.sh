#!/bin/bash
set -euo pipefail

# Конфигурация
ZITADEL_DOMAIN="https://zitadel.local"
ADMIN_EMAIL="admin@local.test"
ADMIN_PASSWORD="jrcCaigzmqzHpRcxo61RDiS+xejxX4/K"
PROJECT_NAME="cloudbridge-project"
APP_NAME="cloudbridge-relay"

echo "🔐 Creating Zitadel project and application..."

# Используем Personal Access Token
echo "📝 Using Personal Access Token..."
ACCESS_TOKEN="b4b59e03c344795e8863e9c8f9ff4dd427af46aaa3c1aa65"

echo "Using PAT for authentication"

# Получаем информацию об организации
echo "🏢 Getting organization info..."
ORG_RESPONSE=$(curl -k -s -X GET "$ZITADEL_DOMAIN/admin/v1/orgs/me" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json")

ORG_ID=$(echo "$ORG_RESPONSE" | jq -r '.org.id')
ORG_NAME=$(echo "$ORG_RESPONSE" | jq -r '.org.name')

echo "Organization: $ORG_NAME (ID: $ORG_ID)"

# Создаем проект
echo "📁 Creating project: $PROJECT_NAME..."
PROJECT_RESPONSE=$(curl -k -s -X POST "$ZITADEL_DOMAIN/admin/v1/projects" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$PROJECT_NAME\",
    \"projectRoleAssertion\": true,
    \"projectRoleCheck\": true,
    \"hasProjectCheck\": true
  }")

PROJECT_ID=$(echo "$PROJECT_RESPONSE" | jq -r '.id')

if [ "$PROJECT_ID" = "null" ] || [ -z "$PROJECT_ID" ]; then
  echo "❌ Failed to create project"
  echo "Response: $PROJECT_RESPONSE"
  exit 1
fi

echo "Project created successfully! ID: $PROJECT_ID"

# Создаем OIDC приложение
echo " Creating OIDC application: $APP_NAME..."
APP_RESPONSE=$(curl -k -s -X POST "$ZITADEL_DOMAIN/admin/v1/projects/$PROJECT_ID/apps/oidc" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"$APP_NAME\",
    \"redirectUris\": [
      \"http://localhost:8080/callback\",
      \"https://localhost:8443/callback\"
    ],
    \"postLogoutRedirectUris\": [
      \"http://localhost:8080/\",
      \"https://localhost:8443/\"
    ],
    \"responseTypes\": [\"CODE\"],
    \"grantTypes\": [\"AUTHORIZATION_CODE\"],
    \"appType\": \"WEB\",
    \"authMethodType\": \"BASIC\",
    \"accessTokenType\": \"BEARER\",
    \"devMode\": true
  }")

APP_ID=$(echo "$APP_RESPONSE" | jq -r '.appId')
CLIENT_ID=$(echo "$APP_RESPONSE" | jq -r '.clientId')
CLIENT_SECRET=$(echo "$APP_RESPONSE" | jq -r '.clientSecret')

if [ "$APP_ID" = "null" ] || [ -z "$APP_ID" ]; then
  echo "❌ Failed to create application"
  echo "Response: $APP_RESPONSE"
  exit 1
fi

echo "Application created successfully!"
echo "   App ID: $APP_ID"
echo "   Client ID: $CLIENT_ID"
echo "   Client Secret: $CLIENT_SECRET"

# Выводим конфигурацию для обновления
echo ""
echo "=== CONFIGURATION TO UPDATE ==="
echo "ZITADEL_PROJECT_ID=$PROJECT_ID"
echo "ZITADEL_CLIENT_ID=$CLIENT_ID"
echo "ZITADEL_CLIENT_SECRET=$CLIENT_SECRET"
echo "ZITADEL_DOMAIN=$ZITADEL_DOMAIN"

echo ""
echo "=== JSON CONFIG ==="
cat << EOF
{
  "zitadel": {
    "enabled": true,
    "domain": "$ZITADEL_DOMAIN",
    "project_id": "$PROJECT_ID",
    "client_id": "$CLIENT_ID",
    "client_secret": "$CLIENT_SECRET",
    "fallback_secret": "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=",
    "token_validation": {
      "audience": "cloudbridge-relay",
      "clock_skew": "30s"
    },
    "cache_ttl": "5m",
    "http_timeout": "30s",
    "tenant_management": {
      "enabled": true,
      "auto_create": false,
      "default_quotas": {
        "max_peers": 10,
        "max_connections": 100,
        "max_tunnels": 20,
        "bandwidth_limit": "1GB",
        "storage_limit": "10GB"
      }
    }
  }
}
EOF

echo ""
echo "Zitadel project and application created successfully!"
echo ""
echo " Next steps:"
echo "1. Update configs/config.json with the above configuration"
echo "2. Update k8s/configmaps/cloudbridge-config.yaml"
echo "3. Restart CloudBridge Relay deployment"
