#!/bin/bash

# CloudBridge External Secrets Setup Script
# Автоматическая установка и настройка External Secrets Operator

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Функции для вывода
info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

security_info() {
    echo -e "${PURPLE}[SECURITY]${NC} $1"
}

# Конфигурация
EXTERNAL_SECRETS_VERSION="v0.9.11"
NAMESPACE="external-secrets-system"
CLOUDBRIDGE_NAMESPACE="cloudbridge"

# Проверка зависимостей
check_dependencies() {
    info "Проверка зависимостей..."
    
    # Проверяем kubectl
    if ! command -v kubectl &> /dev/null; then
        error "kubectl не найден. Установите kubectl и настройте доступ к кластеру."
        exit 1
    fi
    
    # Проверяем доступ к кластеру
    if ! kubectl cluster-info &> /dev/null; then
        error "Нет доступа к Kubernetes кластеру. Настройте kubeconfig."
        exit 1
    fi
    
    # Проверяем helm (опционально)
    if command -v helm &> /dev/null; then
        info "Helm найден, будем использовать Helm для установки"
        USE_HELM=true
    else
        info "Helm не найден, будем использовать kubectl"
        USE_HELM=false
    fi
    
    success "Все зависимости проверены"
}

# Создание namespace
create_namespace() {
    info "Создание namespace для External Secrets..."
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        info "Namespace $NAMESPACE уже существует"
    else
        kubectl create namespace "$NAMESPACE"
        success "Namespace $NAMESPACE создан"
    fi
}

# Установка External Secrets через Helm
install_external_secrets_helm() {
    info "Установка External Secrets через Helm..."
    
    # Добавляем репозиторий External Secrets
    helm repo add external-secrets https://charts.external-secrets.io
    helm repo update
    
    # Устанавливаем External Secrets
    helm upgrade --install external-secrets external-secrets/external-secrets \
        --namespace "$NAMESPACE" \
        --version "$EXTERNAL_SECRETS_VERSION" \
        --set installCRDs=true \
        --set webhook.port=9443 \
        --set webhook.securePort=9443 \
        --set certController.port=8080 \
        --set certController.securePort=8080 \
        --set webhook.certDir=/tmp/k8s-webhook-server/serving-certs \
        --set certController.certDir=/tmp/k8s-webhook-server/serving-certs \
        --wait
    
    success "External Secrets установлен через Helm"
}

# Установка External Secrets через kubectl
install_external_secrets_kubectl() {
    info "Установка External Secrets через kubectl..."
    
    # Скачиваем манифесты
    curl -L https://github.com/external-secrets/external-secrets/releases/download/$EXTERNAL_SECRETS_VERSION/external-secrets.yaml -o /tmp/external-secrets.yaml
    
    # Применяем манифесты
    kubectl apply -f /tmp/external-secrets.yaml
    
    # Ждем готовности
    info "Ожидание готовности External Secrets..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=external-secrets -n "$NAMESPACE" --timeout=300s
    
    success "External Secrets установлен через kubectl"
}

# Проверка установки
verify_installation() {
    info "Проверка установки External Secrets..."
    
    # Проверяем поды
    if kubectl get pods -n "$NAMESPACE" | grep -q "external-secrets.*Running"; then
        success "External Secrets поды запущены"
    else
        error "External Secrets поды не запущены"
        kubectl get pods -n "$NAMESPACE"
        exit 1
    fi
    
    # Проверяем CRD
    if kubectl get crd | grep -q "externalsecrets.external-secrets.io"; then
        success "External Secrets CRD установлены"
    else
        error "External Secrets CRD не найдены"
        exit 1
    fi
    
    # Проверяем API
    if kubectl api-resources | grep -q "externalsecrets.external-secrets.io"; then
        success "External Secrets API доступен"
    else
        error "External Secrets API недоступен"
        exit 1
    fi
}

# Настройка SecretStore
setup_secret_stores() {
    info "Настройка SecretStore для внешних систем..."
    
    # Применяем манифесты
    kubectl apply -f k8s/security/external-secrets-setup.yaml
    
    # Проверяем создание
    if kubectl get secretstore vault-secret-store -n "$CLOUDBRIDGE_NAMESPACE" &> /dev/null; then
        success "Vault SecretStore создан"
    else
        warning "Vault SecretStore не создан (Vault может быть недоступен)"
    fi
    
    if kubectl get secretstore aws-secret-store -n "$CLOUDBRIDGE_NAMESPACE" &> /dev/null; then
        success "AWS SecretStore создан"
    else
        warning "AWS SecretStore не создан (AWS credentials могут отсутствовать)"
    fi
    
    if kubectl get secretstore gcp-secret-store -n "$CLOUDBRIDGE_NAMESPACE" &> /dev/null; then
        success "GCP SecretStore создан"
    else
        warning "GCP SecretStore не создан (GCP credentials могут отсутствовать)"
    fi
}

# Создание External Secrets
create_external_secrets() {
    info "Создание External Secrets..."
    
    # Применяем манифесты External Secrets
    kubectl apply -f k8s/security/external-secrets-setup.yaml
    
    # Проверяем создание External Secrets
    info "Ожидание создания External Secrets..."
    
    # Ждем готовности основных секретов
    kubectl wait --for=condition=ready externalsecret cloudbridge-database-secret -n "$CLOUDBRIDGE_NAMESPACE" --timeout=300s || {
        warning "Database External Secret не готов, проверяем статус..."
        kubectl describe externalsecret cloudbridge-database-secret -n "$CLOUDBRIDGE_NAMESPACE"
    }
    
    success "External Secrets созданы"
}

# Обновление deployment
update_deployment() {
    info "Обновление deployment для использования External Secrets..."
    
    # Создаем backup текущего deployment
    kubectl get deployment cloudbridge-relay -n "$CLOUDBRIDGE_NAMESPACE" -o yaml > k8s/deployments/cloudbridge-relay-deployment-backup-$(date +%Y%m%d-%H%M%S).yaml
    
    # Обновляем deployment для использования External Secrets
    kubectl patch deployment cloudbridge-relay -n "$CLOUDBRIDGE_NAMESPACE" -p '{
        "spec": {
            "template": {
                "spec": {
                    "containers": [{
                        "name": "cloudbridge-relay",
                        "env": [
                            {"name": "DATABASE_URL", "valueFrom": {"secretKeyRef": {"name": "cloudbridge-database", "key": "url"}}},
                            {"name": "ZITADEL_CLIENT_ID", "valueFrom": {"secretKeyRef": {"name": "cloudbridge-zitadel", "key": "client-id"}}},
                            {"name": "ZITADEL_CLIENT_SECRET", "valueFrom": {"secretKeyRef": {"name": "cloudbridge-zitadel", "key": "client-secret"}}}
                        ]
                    }]
                }
            }
        }
    }'
    
    success "Deployment обновлен"
}

# Создание мониторинга
create_monitoring() {
    info "Создание мониторинга External Secrets..."
    
    # Создаем ServiceMonitor для Prometheus
    cat > /tmp/external-secrets-monitoring.yaml << EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: external-secrets
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: external-secrets
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: external-secrets
  endpoints:
  - port: http
    path: /metrics
EOF

    kubectl apply -f /tmp/external-secrets-monitoring.yaml
    rm -f /tmp/external-secrets-monitoring.yaml
    
    success "Мониторинг настроен"
}

# Создание алертов
create_alerts() {
    info "Создание алертов для External Secrets..."
    
    cat > /tmp/external-secrets-alerts.yaml << EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: external-secrets-alerts
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: external-secrets
spec:
  groups:
  - name: external-secrets
    rules:
    - alert: ExternalSecretSyncFailed
      expr: external_secrets_sync_calls_error_total > 0
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "External Secret sync failed"
        description: "External Secret {{ \$labels.name }} in namespace {{ \$labels.namespace }} failed to sync"
    
    - alert: ExternalSecretStoreConnectionFailed
      expr: external_secrets_secret_store_connection_error_total > 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Secret Store connection failed"
        description: "Connection to Secret Store {{ \$labels.name }} in namespace {{ \$labels.namespace }} failed"
    
    - alert: ExternalSecretRefreshFailed
      expr: external_secrets_refresh_calls_error_total > 0
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "External Secret refresh failed"
        description: "External Secret {{ \$labels.name }} in namespace {{ \$labels.namespace }} failed to refresh"
EOF

    kubectl apply -f /tmp/external-secrets-alerts.yaml
    rm -f /tmp/external-secrets-alerts.yaml
    
    success "Алерты созданы"
}

# Тестирование
test_external_secrets() {
    info "Тестирование External Secrets..."
    
    # Проверяем статус External Secrets
    kubectl get externalsecrets -n "$CLOUDBRIDGE_NAMESPACE"
    
    # Проверяем созданные secrets
    kubectl get secrets -n "$CLOUDBRIDGE_NAMESPACE" | grep cloudbridge
    
    # Проверяем статус SecretStore
    kubectl get secretstore -n "$CLOUDBRIDGE_NAMESPACE"
    
    # Проверяем содержимое секрета
    if kubectl get secret cloudbridge-database -n "$CLOUDBRIDGE_NAMESPACE" &> /dev/null; then
        info "Проверка содержимого database секрета..."
        kubectl get secret cloudbridge-database -n "$CLOUDBRIDGE_NAMESPACE" -o jsonpath='{.data.url}' | base64 -d
        success "Database секрет содержит правильные данные"
    else
        warning "Database секрет не найден"
    fi
}

# Создание документации
create_documentation() {
    info "Создание документации..."
    
    cat > "EXTERNAL_SECRETS_SETUP_COMPLETE.md" << EOF
# External Secrets Setup Complete

## Установленные компоненты

### External Secrets Operator
- **Namespace**: $NAMESPACE
- **Version**: $EXTERNAL_SECRETS_VERSION
- **Installation Method**: $([ "$USE_HELM" = true ] && echo "Helm" || echo "kubectl")

### SecretStores
- **vault-secret-store**: Vault integration
- **aws-secret-store**: AWS Secrets Manager
- **gcp-secret-store**: Google Secret Manager

### External Secrets
- **cloudbridge-database-secret**: Database credentials
- **cloudbridge-zitadel-secret**: Zitadel credentials
- **cloudbridge-api-keys**: API keys
- **cloudbridge-monitoring-secret**: Monitoring credentials
- **cloudbridge-wireguard-secret**: WireGuard keys

## Команды для управления

### Проверка статуса
\`\`\`bash
# Статус External Secrets
kubectl get pods -n $NAMESPACE

# Статус External Secrets
kubectl get externalsecrets -n $CLOUDBRIDGE_NAMESPACE

# Статус SecretStore
kubectl get secretstore -n $CLOUDBRIDGE_NAMESPACE

# Детали External Secret
kubectl describe externalsecret cloudbridge-database-secret -n $CLOUDBRIDGE_NAMESPACE
\`\`\`

### Обновление секретов
\`\`\`bash
# Принудительное обновление
kubectl annotate externalsecret cloudbridge-database-secret -n $CLOUDBRIDGE_NAMESPACE external-secrets.io/refresh="true"

# Проверка обновления
kubectl get externalsecret cloudbridge-database-secret -n $CLOUDBRIDGE_NAMESPACE -o yaml
\`\`\`

### Мониторинг
\`\`\`bash
# Логи External Secrets
kubectl logs -f deployment/external-secrets -n $NAMESPACE

# Метрики
kubectl port-forward -n $NAMESPACE svc/external-secrets 8080:8080
curl http://localhost:8080/metrics
\`\`\`

## Автоматическое обновление

External Secrets автоматически обновляет секреты согласно refreshInterval.

## Troubleshooting

### Проблемы с SecretStore
\`\`\`bash
kubectl describe secretstore vault-secret-store -n $CLOUDBRIDGE_NAMESPACE
kubectl get events -n $CLOUDBRIDGE_NAMESPACE --field-selector involvedObject.name=vault-secret-store
\`\`\`

### Проверка External Secret
\`\`\`bash
kubectl describe externalsecret cloudbridge-database-secret -n $CLOUDBRIDGE_NAMESPACE
kubectl get events -n $CLOUDBRIDGE_NAMESPACE --field-selector involvedObject.name=cloudbridge-database-secret
\`\`\`
EOF

    success "Документация создана: EXTERNAL_SECRETS_SETUP_COMPLETE.md"
}

# Основная функция
main() {
    info "CloudBridge External Secrets Setup"
    info "==================================="
    echo ""
    
    check_dependencies
    create_namespace
    
    if [ "$USE_HELM" = true ]; then
        install_external_secrets_helm
    else
        install_external_secrets_kubectl
    fi
    
    verify_installation
    setup_secret_stores
    create_external_secrets
    update_deployment
    create_monitoring
    create_alerts
    test_external_secrets
    create_documentation
    
    success "External Secrets setup завершен!"
    echo ""
    security_info "Следующие шаги:"
    echo "1. Проверьте статус External Secrets: kubectl get externalsecrets -n $CLOUDBRIDGE_NAMESPACE"
    echo "2. Проверьте работу секретов в deployment"
    echo "3. Настройте мониторинг в Prometheus/Grafana"
    echo "4. Прочитайте документацию: EXTERNAL_SECRETS_SETUP_COMPLETE.md"
}

# Запуск
main "$@"
