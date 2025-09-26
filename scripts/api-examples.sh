#!/bin/bash
# CloudBridge + ZITADEL API Examples
# Примеры вызовов для разработчика

source "$(dirname "$0")/../configs/developer-env.sh"

echo "🧪 Примеры API вызовов для разработчика"
echo ""

echo "1️⃣ Тест admin токена:"
echo "curl -H \"Authorization: Bearer \$ZITADEL_ADMIN_TOKEN\" \\"
echo "     \"\$ZITADEL_MANAGEMENT_API/orgs/me\""
echo ""

echo "2️⃣ Получение Client Credentials токена:"  
echo "curl -X POST \"\$ZITADEL_TOKEN_ENDPOINT\" \\"
echo "     -u \"\$API_CLIENT_ID:\$API_CLIENT_SECRET\" \\"
echo "     -d \"grant_type=client_credentials&scope=openid profile email\""
echo ""

echo "3️⃣ Добавление custom scope в проект:"
echo "curl -X POST \"\$ZITADEL_MANAGEMENT_API/projects/\$PROJECT_ID/roles\" \\"
echo "     -H \"Authorization: Bearer \$ZITADEL_ADMIN_TOKEN\" \\"
echo "     -H \"Content-Type: application/json\" \\"
echo "     -d '{\"key\": \"relay.api\", \"display_name\": \"Relay API Access\", \"group\": \"api\"}'"
echo ""

echo "4️⃣ Проверка JWKS:"
echo "curl \"\$ZITADEL_JWKS_URI\" | jq ."
echo ""

echo "5️⃣ Introspection токена:"
echo "curl -X POST \"\$ZITADEL_ISSUER/oauth/v2/introspect\" \\"
echo "     -u \"\$API_CLIENT_ID:\$API_CLIENT_SECRET\" \\" 
echo "     -d \"token=YOUR_ACCESS_TOKEN\""
echo ""

echo "🔧 Для выполнения примеров запустите:"
echo "  source configs/developer-env.sh"  
echo "  bash scripts/api-examples.sh"
