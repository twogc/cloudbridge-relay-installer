#!/usr/bin/env bash
set -euo pipefail

URL=${1:-https://zitadel.local}

echo "Fetching OIDC discovery from $URL/.well-known/openid-configuration" >&2
curl -sk "$URL/.well-known/openid-configuration" | jq '{issuer, jwks_uri, authorization_endpoint, token_endpoint, userinfo_endpoint}'






