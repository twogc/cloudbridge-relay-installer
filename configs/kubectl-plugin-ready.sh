#!/bin/bash
# kubectl-zitadel plugin - финальная рабочая конфигурация

# ZITADEL Configuration
export ZITADEL_ISSUER="https://edge.2gc.ru"
export ZITADEL_CLIENT_ID="339537393138794797"
export ZITADEL_CLIENT_SECRET="Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"

# Kubernetes scopes  
export KUBECTL_SCOPES="openid profile email urn:zitadel:iam:org:project:id:zitadel:aud"
export CUSTOM_SCOPES="relay_api relay_k8s"

# mTLS Certificates (готовы)
export MTLS_CA_CERT="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/cloudbridge-relay-ca-cert.pem"
export MTLS_CLIENT_CERT="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/relay-client-cert.pem"
export MTLS_CLIENT_KEY="/home/ubuntu/cloudbridge-relay-installer/certs/relay-mtls/relay-client-key.pem"

# Token endpoints
export TOKEN_ENDPOINT="https://edge.2gc.ru/oauth/v2/token"
export JWKS_URI="https://edge.2gc.ru/oauth/v2/keys"

echo "✅ kubectl-zitadel plugin конфигурация загружена!"
echo "🔑 Client ID: $ZITADEL_CLIENT_ID"
echo "🌐 Issuer: $ZITADEL_ISSUER"
echo "🛡️ mTLS certs: готовы"
echo "📊 Scopes: $KUBECTL_SCOPES + $CUSTOM_SCOPES"
