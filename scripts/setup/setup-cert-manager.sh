#!/bin/bash

# CloudBridge cert-manager Setup Script
# Автоматическая установка и настройка cert-manager для управления TLS сертификатами

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
CERT_MANAGER_VERSION="v1.13.3"
NAMESPACE="cert-manager"
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
    info "Создание namespace для cert-manager..."
    
    if kubectl get namespace "$NAMESPACE" &> /dev/null; then
        info "Namespace $NAMESPACE уже существует"
    else
        kubectl create namespace "$NAMESPACE"
        success "Namespace $NAMESPACE создан"
    fi
}

# Установка cert-manager через Helm
install_cert_manager_helm() {
    info "Установка cert-manager через Helm..."
    
    # Добавляем репозиторий cert-manager
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    
    # Устанавливаем cert-manager
    helm upgrade --install cert-manager jetstack/cert-manager \
        --namespace "$NAMESPACE" \
        --version "$CERT_MANAGER_VERSION" \
        --set installCRDs=true \
        --set global.leaderElection.namespace="$NAMESPACE" \
        --set webhook.securePort=10250 \
        --set cainjector.enabled=true \
        --set startupapicheck.enabled=true \
        --wait
    
    success "cert-manager установлен через Helm"
}

# Установка cert-manager через kubectl
install_cert_manager_kubectl() {
    info "Установка cert-manager через kubectl..."
    
    # Скачиваем манифесты
    curl -L https://github.com/cert-manager/cert-manager/releases/download/$CERT_MANAGER_VERSION/cert-manager.yaml -o /tmp/cert-manager.yaml
    
    # Применяем манифесты
    kubectl apply -f /tmp/cert-manager.yaml
    
    # Ждем готовности
    info "Ожидание готовности cert-manager..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cert-manager -n "$NAMESPACE" --timeout=300s
    
    success "cert-manager установлен через kubectl"
}

# Проверка установки
verify_installation() {
    info "Проверка установки cert-manager..."
    
    # Проверяем поды
    if kubectl get pods -n "$NAMESPACE" | grep -q "cert-manager.*Running"; then
        success "cert-manager поды запущены"
    else
        error "cert-manager поды не запущены"
        kubectl get pods -n "$NAMESPACE"
        exit 1
    fi
    
    # Проверяем CRD
    if kubectl get crd | grep -q "certificates.cert-manager.io"; then
        success "cert-manager CRD установлены"
    else
        error "cert-manager CRD не найдены"
        exit 1
    fi
    
    # Проверяем API
    if kubectl api-resources | grep -q "certificates.cert-manager.io"; then
        success "cert-manager API доступен"
    else
        error "cert-manager API недоступен"
        exit 1
    fi
}

# Создание ClusterIssuer
create_cluster_issuer() {
    info "Создание ClusterIssuer для Let's Encrypt..."
    
    # Применяем манифесты
    kubectl apply -f k8s/security/cert-manager-issuer.yaml
    
    # Проверяем создание
    if kubectl get clusterissuer letsencrypt-prod &> /dev/null; then
        success "ClusterIssuer letsencrypt-prod создан"
    else
        error "Не удалось создать ClusterIssuer"
        exit 1
    fi
    
    if kubectl get clusterissuer letsencrypt-staging &> /dev/null; then
        success "ClusterIssuer letsencrypt-staging создан"
    else
        error "Не удалось создать ClusterIssuer staging"
        exit 1
    fi
}

# Создание сертификатов
create_certificates() {
    info "Создание сертификатов..."
    
    # Применяем манифесты сертификатов
    kubectl apply -f k8s/security/cert-manager-certificates.yaml
    
    # Проверяем создание сертификатов
    info "Ожидание создания сертификатов (это может занять несколько минут)..."
    
    # Ждем готовности основного сертификата
    kubectl wait --for=condition=ready certificate cloudbridge-main-tls -n "$CLOUDBRIDGE_NAMESPACE" --timeout=600s || {
        warning "Основной сертификат не готов, проверяем статус..."
        kubectl describe certificate cloudbridge-main-tls -n "$CLOUDBRIDGE_NAMESPACE"
    }
    
    success "Сертификаты созданы"
}

# Обновление deployment
update_deployment() {
    info "Обновление deployment для использования cert-manager сертификатов..."
    
    # Создаем backup текущего deployment
    kubectl get deployment cloudbridge-relay -n "$CLOUDBRIDGE_NAMESPACE" -o yaml > k8s/deployments/cloudbridge-relay-deployment-backup-$(date +%Y%m%d-%H%M%S).yaml
    
    # Обновляем deployment для использования нового TLS secret
    kubectl patch deployment cloudbridge-relay -n "$CLOUDBRIDGE_NAMESPACE" -p '{
        "spec": {
            "template": {
                "spec": {
                    "containers": [{
                        "name": "cloudbridge-relay",
                        "env": [
                            {"name": "TLS_CERT_FILE", "value": "/etc/ssl/certs/tls.crt"},
                            {"name": "TLS_KEY_FILE", "value": "/etc/ssl/private/tls.key"}
                        ],
                        "volumeMounts": [
                            {"name": "tls-certs", "mountPath": "/etc/ssl/certs", "readOnly": true}
                        ]
                    }],
                    "volumes": [{
                        "name": "tls-certs",
                        "secret": {
                            "secretName": "cloudbridge-main-tls"
                        }
                    }]
                }
            }
        }
    }'
    
    success "Deployment обновлен"
}

# Создание мониторинга
create_monitoring() {
    info "Создание мониторинга сертификатов..."
    
    # Создаем ServiceMonitor для Prometheus
    cat > /tmp/cert-manager-monitoring.yaml << EOF
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: cert-manager
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: cert-manager
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: cert-manager
  endpoints:
  - port: tcp-prometheus-servicemonitor
    path: /metrics
EOF

    kubectl apply -f /tmp/cert-manager-monitoring.yaml
    rm -f /tmp/cert-manager-monitoring.yaml
    
    success "Мониторинг настроен"
}

# Создание алертов
create_alerts() {
    info "Создание алертов для сертификатов..."
    
    cat > /tmp/cert-manager-alerts.yaml << EOF
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cert-manager-alerts
  namespace: $NAMESPACE
  labels:
    app.kubernetes.io/name: cert-manager
spec:
  groups:
  - name: cert-manager
    rules:
    - alert: CertificateExpiringSoon
      expr: cert_manager_certificate_expiration_timestamp_seconds - time() < 7 * 24 * 3600
      for: 1h
      labels:
        severity: warning
      annotations:
        summary: "Certificate expiring soon"
        description: "Certificate {{ \$labels.name }} in namespace {{ \$labels.namespace }} expires in less than 7 days"
    
    - alert: CertificateExpired
      expr: cert_manager_certificate_expiration_timestamp_seconds - time() < 0
      for: 0m
      labels:
        severity: critical
      annotations:
        summary: "Certificate expired"
        description: "Certificate {{ \$labels.name }} in namespace {{ \$labels.namespace }} has expired"
    
    - alert: CertificateIssuanceFailed
      expr: cert_manager_certificate_ready_status{condition="False"} == 1
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "Certificate issuance failed"
        description: "Certificate {{ \$labels.name }} in namespace {{ \$labels.namespace }} failed to be issued"
EOF

    kubectl apply -f /tmp/cert-manager-alerts.yaml
    rm -f /tmp/cert-manager-alerts.yaml
    
    success "Алерты созданы"
}

# Тестирование
test_certificates() {
    info "Тестирование сертификатов..."
    
    # Проверяем статус сертификатов
    kubectl get certificates -n "$CLOUDBRIDGE_NAMESPACE"
    
    # Проверяем secrets
    kubectl get secrets -n "$CLOUDBRIDGE_NAMESPACE" | grep tls
    
    # Проверяем содержимое сертификата
    if kubectl get secret cloudbridge-main-tls -n "$CLOUDBRIDGE_NAMESPACE" &> /dev/null; then
        info "Проверка содержимого сертификата..."
        kubectl get secret cloudbridge-main-tls -n "$CLOUDBRIDGE_NAMESPACE" -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout | grep -A 1 "Subject:"
        success "Сертификат содержит правильные данные"
    else
        warning "Основной сертификат не найден"
    fi
}

# Создание документации
create_documentation() {
    info "Создание документации..."
    
    cat > "CERT_MANAGER_SETUP_COMPLETE.md" << EOF
# cert-manager Setup Complete

## Установленные компоненты

### cert-manager
- **Namespace**: $NAMESPACE
- **Version**: $CERT_MANAGER_VERSION
- **Installation Method**: $([ "$USE_HELM" = true ] && echo "Helm" || echo "kubectl")

### ClusterIssuers
- **letsencrypt-prod**: Production Let's Encrypt
- **letsencrypt-staging**: Staging Let's Encrypt (for testing)

### Certificates
- **cloudbridge-main-tls**: Main certificate for *.2gc.ru
- **cloudbridge-grpc-tls**: gRPC mTLS certificate
- **cloudbridge-api-tls**: API certificate
- **cloudbridge-monitoring-tls**: Monitoring certificate
- **cloudbridge-zitadel-tls**: Zitadel certificate

## Команды для управления

### Проверка статуса
\`\`\`bash
# Статус cert-manager
kubectl get pods -n $NAMESPACE

# Статус сертификатов
kubectl get certificates -n $CLOUDBRIDGE_NAMESPACE

# Детали сертификата
kubectl describe certificate cloudbridge-main-tls -n $CLOUDBRIDGE_NAMESPACE
\`\`\`

### Обновление сертификатов
\`\`\`bash
# Принудительное обновление
kubectl annotate certificate cloudbridge-main-tls -n $CLOUDBRIDGE_NAMESPACE cert-manager.io/renew-before="24h"

# Проверка обновления
kubectl get certificate cloudbridge-main-tls -n $CLOUDBRIDGE_NAMESPACE -o yaml
\`\`\`

### Мониторинг
\`\`\`bash
# Логи cert-manager
kubectl logs -f deployment/cert-manager -n $NAMESPACE

# Метрики
kubectl port-forward -n $NAMESPACE svc/cert-manager 9402:9402
curl http://localhost:9402/metrics
\`\`\`

## Автоматическое обновление

cert-manager автоматически обновляет сертификаты за 15 дней до истечения.

## Troubleshooting

### Проблемы с DNS
Если DNS-01 challenge не работает, используйте HTTP-01:
\`\`\`bash
kubectl patch certificate cloudbridge-main-tls -n $CLOUDBRIDGE_NAMESPACE --type merge -p '{"spec":{"issuerRef":{"name":"letsencrypt-staging"}}}'
\`\`\`

### Проверка ACME
\`\`\`bash
kubectl describe clusterissuer letsencrypt-prod
kubectl get secret letsencrypt-prod -n $NAMESPACE
\`\`\`
EOF

    success "Документация создана: CERT_MANAGER_SETUP_COMPLETE.md"
}

# Основная функция
main() {
    info "CloudBridge cert-manager Setup"
    info "=============================="
    echo ""
    
    check_dependencies
    create_namespace
    
    if [ "$USE_HELM" = true ]; then
        install_cert_manager_helm
    else
        install_cert_manager_kubectl
    fi
    
    verify_installation
    create_cluster_issuer
    create_certificates
    update_deployment
    create_monitoring
    create_alerts
    test_certificates
    create_documentation
    
    success "cert-manager setup завершен!"
    echo ""
    security_info "Следующие шаги:"
    echo "1. Проверьте статус сертификатов: kubectl get certificates -n $CLOUDBRIDGE_NAMESPACE"
    echo "2. Проверьте работу HTTPS endpoints"
    echo "3. Настройте мониторинг в Prometheus/Grafana"
    echo "4. Прочитайте документацию: CERT_MANAGER_SETUP_COMPLETE.md"
}

# Запуск
main "$@"
