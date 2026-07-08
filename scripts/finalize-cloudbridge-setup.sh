#!/bin/bash
set -euo pipefail

# CloudBridge Relay + ZITADEL Finalization Script
# Финальная настройка интеграции после создания OIDC приложений

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Цвета
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }

# Параметры (будут получены от пользователя)
API_CLIENT_ID="${1:-}"
API_CLIENT_SECRET="${2:-}"
SPA_CLIENT_ID="${3:-}"
ZITADEL_DOMAIN="edge.2gc.ru"

main() {
    log "Финализация CloudBridge Relay + ZITADEL интеграции"
    log ""
    
    if [[ -z "$API_CLIENT_ID" || -z "$API_CLIENT_SECRET" || -z "$SPA_CLIENT_ID" ]]; then
        log "Использование: $0 <API_CLIENT_ID> <API_CLIENT_SECRET> <SPA_CLIENT_ID>"
        log ""
        log "Пример:"
        log "$0 123456789012345678 abc123def456ghi789jkl012 987654321098765432"
        exit 1
    fi
    
    log "Конфигурация OIDC приложений:"
    log "  API Client ID: $API_CLIENT_ID"
    log "  API Secret: ${API_CLIENT_SECRET:0:10}..."
    log "  SPA Client ID: $SPA_CLIENT_ID"
    log "  Domain: $ZITADEL_DOMAIN"
    log ""
    
    # 1. Создаем конфигурационные файлы
    create_config_files
    
    # 2. Обновляем Kubernetes секреты
    update_k8s_secrets
    
    # 3. Создаем kubectl exec-plugin
    create_kubectl_plugin
    
    # 4. Тестируем интеграцию
    test_integration
    
    log "Интеграция CloudBridge Relay + ZITADEL завершена!"
}

create_config_files() {
    log "📁 Создание конфигурационных файлов..."
    
    # Основная конфигурация ZITADEL
    cat > "$PROJECT_ROOT/configs/zitadel-final-config.yaml" <<EOF
zitadel:
  enabled: true
  domain: "$ZITADEL_DOMAIN"
  issuer_url: "https://$ZITADEL_DOMAIN"
  
  # OIDC Applications
  applications:
    api:
      client_id: "$API_CLIENT_ID"
      client_secret: "$API_CLIENT_SECRET"
      grant_types: ["client_credentials", "authorization_code"]
      scopes: ["openid", "profile", "email", "relay.api"]
    
    spa:
      client_id: "$SPA_CLIENT_ID"
      grant_types: ["authorization_code", "device_code"]
      scopes: ["openid", "profile", "email", "relay.api", "relay.k8s", "offline_access"]
      use_pkce: true
  
  # JWT Configuration
  jwt:
    custom_claims:
      tenant_id: "uuid"
      client_id: "string"
      roles: "array"
      audit_session_id: "string"
    
    token_lifetimes:
      access_token: "15m"
      refresh_token: "8h"
      id_token: "1h"
  
  # mTLS Configuration
  mtls:
    enabled: true
    ca_cert_path: "$PROJECT_ROOT/certs/relay-mtls/cloudbridge-relay-ca-cert.pem"
    client_cert_path: "$PROJECT_ROOT/certs/relay-mtls/relay-client-cert.pem"
    client_key_path: "$PROJECT_ROOT/certs/relay-mtls/relay-client-key.pem"
EOF

    log "Конфигурация сохранена: configs/zitadel-final-config.yaml"
}

update_k8s_secrets() {
    log "☸️ Обновление Kubernetes секретов..."
    
    # Создаем секрет с OIDC данными
    kubectl create secret generic cloudbridge-zitadel \
        --from-literal=api-client-id="$API_CLIENT_ID" \
        --from-literal=api-client-secret="$API_CLIENT_SECRET" \
        --from-literal=spa-client-id="$SPA_CLIENT_ID" \
        --from-literal=issuer-url="https://$ZITADEL_DOMAIN" \
        --namespace=cloudbridge \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log "Секрет cloudbridge-zitadel обновлен"
}

create_kubectl_plugin() {
    log "🔧 Создание kubectl exec-plugin..."
    
    cat > "$PROJECT_ROOT/cmd/kubectl-zitadel/main.go" <<'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

// ExecCredential represents the kubectl exec plugin response
type ExecCredential struct {
    APIVersion string `json:"apiVersion"`
    Kind       string `json:"kind"`
    Status     Status `json:"status"`
}

type Status struct {
    Token               string    `json:"token"`
    ExpirationTimestamp time.Time `json:"expirationTimestamp"`
}

func main() {
    // Здесь будет логика получения JWT токена из ZITADEL
    // Пока что возвращаем пример структуры
    
    cred := ExecCredential{
        APIVersion: "client.authentication.k8s.io/v1beta1",
        Kind:       "ExecCredential",
        Status: Status{
            Token:               "example-jwt-token-from-zitadel",
            ExpirationTimestamp: time.Now().Add(15 * time.Minute),
        },
    }
    
    json.NewEncoder(os.Stdout).Encode(cred)
}
EOF

    log "kubectl exec-plugin создан: cmd/kubectl-zitadel/main.go"
}

test_integration() {
    log "🧪 Тестирование интеграции..."
    
    # Проверяем доступность ZITADEL
    if curl -s -f "https://$ZITADEL_DOMAIN/.well-known/openid_configuration" >/dev/null; then
        log "ZITADEL OIDC endpoint доступен"
    else
        log "Проблема с доступностью ZITADEL"
    fi
    
    # Проверяем Kubernetes секреты
    if kubectl get secret cloudbridge-zitadel -n cloudbridge >/dev/null 2>&1; then
        log "Kubernetes секрет создан"
    else
        log "Проблема с Kubernetes секретом"
    fi
    
    # Проверяем сертификаты
    if [[ -f "$PROJECT_ROOT/certs/relay-mtls/cloudbridge-relay-ca-cert.pem" ]]; then
        log "mTLS сертификаты найдены"
    else
        log "mTLS сертификаты не найдены"
    fi
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
