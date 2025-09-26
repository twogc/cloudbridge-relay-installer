#!/bin/bash
# CloudBridge + ZITADEL Developer Environment
# Все данные для доработки системы

export ZITADEL_DOMAIN="edge.2gc.ru"
export ZITADEL_ISSUER="https://edge.2gc.ru"

# API Endpoints
export ZITADEL_MANAGEMENT_API="https://edge.2gc.ru/management/v1"
export ZITADEL_AUTH_API="https://edge.2gc.ru/auth/v1" 
export ZITADEL_TOKEN_ENDPOINT="https://edge.2gc.ru/oauth/v2/token"
export ZITADEL_JWKS_URI="https://edge.2gc.ru/oauth/v2/keys"

# Admin Token (Personal Access Token)
export ZITADEL_ADMIN_TOKEN="1awoo06EgqhX0rQPLiHpQDrMlpj-iLZJ8tc7Jpx7UU4kBLLd5R-xMlEwRR3CEHyI1g1Cp5g"

# Service User Credentials  
export SERVICE_USER_CLIENT_ID="339537393138794797"
export SERVICE_USER_CLIENT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
export SERVICE_USER_JSON_KEY_PATH="$HOME/.zitadel/service-key.json"

# Organization & Project IDs
export ORGANIZATION_ID="339340397199294575"
export PROJECT_ID="339537327304999215"
export SERVICE_USER_ID="339533564527771946"

# OIDC Applications
export API_APP_ID="339537393122017581"
export API_CLIENT_ID="339537393138794797"
export API_CLIENT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"

export SPA_APP_ID="339537407432982831" 
export SPA_CLIENT_ID="339537407433048367"

# Required Scopes  
export REQUIRED_SCOPES="openid profile email urn:zitadel:iam:org:project:id:zitadel:aud"
export CUSTOM_SCOPES="relay.api relay.k8s"

# mTLS Certificates
export MTLS_CA_CERT="$PWD/certs/relay-mtls/cloudbridge-relay-ca-cert.pem"
export MTLS_CLIENT_CERT="$PWD/certs/relay-mtls/relay-client-cert.pem"
export MTLS_CLIENT_KEY="$PWD/certs/relay-mtls/relay-client-key.pem"

echo "✅ CloudBridge + ZITADEL Developer Environment загружен"
echo "🔐 Токены и endpoints настроены"
echo "📋 Используйте: curl -H \"Authorization: Bearer \$ZITADEL_ADMIN_TOKEN\" \$ZITADEL_MANAGEMENT_API/orgs/me"
