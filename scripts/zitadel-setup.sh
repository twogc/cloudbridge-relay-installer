#!/bin/bash
set -euo pipefail

# CloudBridge Relay ZITADEL Setup Script
# Автоматизированная настройка ZITADEL для интеграции с CloudBridge Relay

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Конфигурация
ZITADEL_DOMAIN="${ZITADEL_DOMAIN:-idp.example.com}"
ZITADEL_NAMESPACE="${ZITADEL_NAMESPACE:-zitadel}"
ORG_NAME="${ORG_NAME:-cloudbridge}"
PROJECT_NAME="${PROJECT_NAME:-cloudbridge-relay}"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}"
}

info() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"
}

# Проверка предварительных требований
check_prerequisites() {
    log "Проверка предварительных требований..."
    
    local missing_tools=()
    
    for tool in kubectl helm curl jq openssl; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        error "Отсутствуют необходимые инструменты: ${missing_tools[*]}"
        error "Установите их перед запуском скрипта"
        exit 1
    fi
    
    # Проверка доступа к кластеру
    if ! kubectl cluster-info &> /dev/null; then
        error "Нет доступа к Kubernetes кластеру"
        error "Настройте kubectl и повторите попытку"
        exit 1
    fi
    
    log "Все предварительные требования выполнены"
}

# Создание namespace
create_namespace() {
    log "Создание namespace $ZITADEL_NAMESPACE..."
    
    kubectl create namespace "$ZITADEL_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
    
    log "Namespace $ZITADEL_NAMESPACE создан/обновлен"
}

# Генерация TLS сертификатов
generate_tls_certificates() {
    log "Генерация TLS сертификатов..."
    
    local cert_dir="$PROJECT_ROOT/certs/zitadel"
    mkdir -p "$cert_dir"
    
    # Проверка существования сертификатов
    if [[ -f "$cert_dir/$ZITADEL_DOMAIN.crt" && -f "$cert_dir/$ZITADEL_DOMAIN.key" ]]; then
        warn "TLS сертификаты уже существуют, пропускаем генерацию"
        return 0
    fi
    
    # Генерация самоподписанного сертификата для разработки
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$cert_dir/$ZITADEL_DOMAIN.key" \
        -out "$cert_dir/$ZITADEL_DOMAIN.crt" \
        -subj "/CN=$ZITADEL_DOMAIN/O=CloudBridge/C=RU" \
        -addext "subjectAltName=DNS:$ZITADEL_DOMAIN,DNS:*.$ZITADEL_DOMAIN"
    
    # Создание Kubernetes secret
    kubectl create secret tls zitadel-tls \
        -n "$ZITADEL_NAMESPACE" \
        --cert="$cert_dir/$ZITADEL_DOMAIN.crt" \
        --key="$cert_dir/$ZITADEL_DOMAIN.key" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    log "TLS сертификаты созданы и загружены в кластер"
}

# Установка PostgreSQL
install_postgresql() {
    log "Установка PostgreSQL..."
    
    # Добавление репозитория Helm
    helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
    helm repo update bitnami
    
    # Генерация паролей
    local postgres_password=$(openssl rand -hex 16)
    local zitadel_password=$(openssl rand -hex 16)
    
    # Создание секрета с паролями
    kubectl create secret generic pg-auth \
        -n "$ZITADEL_NAMESPACE" \
        --from-literal=postgres-password="$postgres_password" \
        --from-literal=password="$zitadel_password" \
        --dry-run=client -o yaml | kubectl apply -f -
    
    # Установка PostgreSQL
    helm upgrade --install pg bitnami/postgresql \
        -n "$ZITADEL_NAMESPACE" \
        -f "$PROJECT_ROOT/k8s/zitadel/pg-values.yaml" \
        --set auth.existingSecret=pg-auth
    
    # Ожидание готовности PostgreSQL
    log "Ожидание готовности PostgreSQL..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=postgresql \
        -n "$ZITADEL_NAMESPACE" --timeout=300s
    
    log "PostgreSQL установлен и готов к работе"
}

# Генерация masterkey для ZITADEL
generate_masterkey() {
    log "Генерация masterkey для ZITADEL..."
    
    # Проверка существования masterkey
    if kubectl get secret zitadel-masterkey -n "$ZITADEL_NAMESPACE" &> /dev/null; then
        warn "Masterkey уже существует, пропускаем генерацию"
        return 0
    fi
    
    # Генерация 32-байтного ключа
    local masterkey=$(openssl rand -hex 32)
    
    kubectl create secret generic zitadel-masterkey \
        -n "$ZITADEL_NAMESPACE" \
        --from-literal=masterkey="$masterkey"
    
    log "Masterkey сгенерирован и сохранен"
}

# Установка ZITADEL
install_zitadel() {
    log "Установка ZITADEL..."
    
    # Добавление репозитория ZITADEL
    helm repo add zitadel https://charts.zitadel.com 2>/dev/null || true
    helm repo update zitadel
    
    # Обновление values с актуальным доменом
    local values_file="$PROJECT_ROOT/k8s/zitadel/values-cloudbridge.yaml"
    
    cat > "$values_file" <<EOF
zitadel:
  masterkeySecretName: zitadel-masterkey
  configmapConfig:
    ExternalDomain: $ZITADEL_DOMAIN
    ExternalSecure: true
    ExternalPort: 443
    FirstInstance:
      Org:
        Name: $ORG_NAME
        LoginClient:
          Machine:
            Username: 'zitadel-admin'
            Name: 'CloudBridge Admin Client'
          Pat:
            ExpirationDate: '2029-12-31T23:59:59Z'
  
  secretConfig:
    Database:
      Postgres:
        Host: pg-postgresql.$ZITADEL_NAMESPACE.svc.cluster.local
        Port: 5432
        Database: zitadel
        User:
          Username: zitadel
        SSL:
          Mode: disable
    TLS:
      Enabled: false
      Mode: external
  
  init:
    activeDeadlineSeconds: 3600
    backoffLimit: 3
  setup:
    activeDeadlineSeconds: 3600
    backoffLimit: 3

resources:
  requests:
    cpu: 500m
    memory: 2Gi
  limits:
    cpu: "2"
    memory: 4Gi

login:
  enabled: false
EOF
    
    # Установка ZITADEL
    helm upgrade --install zitadel zitadel/zitadel \
        -n "$ZITADEL_NAMESPACE" \
        -f "$values_file" \
        --set zitadel.secretConfig.Database.Postgres.User.Password="$(kubectl get secret pg-auth -n $ZITADEL_NAMESPACE -o jsonpath='{.data.password}' | base64 -d)" \
        --set zitadel.secretConfig.Database.Postgres.Admin.Username=postgres \
        --set zitadel.secretConfig.Database.Postgres.Admin.Password="$(kubectl get secret pg-auth -n $ZITADEL_NAMESPACE -o jsonpath='{.data.postgres-password}' | base64 -d)"
    
    # Ожидание завершения инициализации
    log "Ожидание завершения инициализации ZITADEL..."
    
    kubectl wait --for=condition=complete job -l app.kubernetes.io/name=zitadel \
        -n "$ZITADEL_NAMESPACE" --timeout=600s
    
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=zitadel \
        -n "$ZITADEL_NAMESPACE" --timeout=300s
    
    log "ZITADEL установлен и готов к работе"
}

# Создание Ingress
create_ingress() {
    log "Создание Ingress для ZITADEL..."
    
    cat > "$PROJECT_ROOT/k8s/zitadel/ingress/ingress-cloudbridge.yaml" <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: zitadel-ingress
  namespace: $ZITADEL_NAMESPACE
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-buffer-size: "8k"
    nginx.ingress.kubernetes.io/proxy-buffers-number: "4"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - $ZITADEL_DOMAIN
    secretName: zitadel-tls
  rules:
  - host: $ZITADEL_DOMAIN
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: zitadel
            port:
              number: 8080
EOF
    
    kubectl apply -f "$PROJECT_ROOT/k8s/zitadel/ingress/ingress-cloudbridge.yaml"
    
    log "Ingress создан"
}

# Ожидание готовности ZITADEL
wait_for_zitadel() {
    log "Ожидание готовности ZITADEL API..."
    
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -kLs "https://$ZITADEL_DOMAIN/.well-known/openid-configuration" &> /dev/null; then
            log "ZITADEL API готов к работе"
            return 0
        fi
        
        info "Попытка $attempt/$max_attempts: ожидание готовности API..."
        sleep 10
        ((attempt++))
    done
    
    error "ZITADEL API не готов после $max_attempts попыток"
    return 1
}

# Создание организации и проекта через ZITADEL Console API
setup_zitadel_config() {
    log "Настройка организации и проекта в ZITADEL..."
    
    # Получение admin token (временно используем machine token)
    local admin_token
    admin_token=$(get_admin_token)
    
    if [[ -z "$admin_token" ]]; then
        warn "Не удалось получить admin token, настройка через API пропущена"
        warn "Выполните настройку вручную через веб-интерфейс: https://$ZITADEL_DOMAIN"
        return 0
    fi
    
    # Создание организации cloudbridge
    create_organization "$admin_token"
    
    # Создание проекта cloudbridge-relay  
    create_project "$admin_token"
    
    log "Базовая настройка ZITADEL завершена"
}

# Получение admin токена
get_admin_token() {
    # TODO: Реализовать получение admin token
    # Пока возвращаем пустой токен для ручной настройки
    echo ""
}

# Создание организации
create_organization() {
    local token="$1"
    
    log "Создание организации $ORG_NAME..."
    
    # TODO: Реализовать создание организации через API
    warn "Создание организации через API пока не реализовано"
    warn "Создайте организацию '$ORG_NAME' вручную через веб-интерфейс"
}

# Создание проекта
create_project() {
    local token="$1"
    
    log "Создание проекта $PROJECT_NAME..."
    
    # TODO: Реализовать создание проекта через API
    warn "Создание проекта через API пока не реализовано" 
    warn "Создайте проект '$PROJECT_NAME' вручную через веб-интерфейс"
}

# Проверка развертывания
verify_deployment() {
    log "Проверка развертывания..."
    
    # Проверка подов
    log "Статус подов в namespace $ZITADEL_NAMESPACE:"
    kubectl get pods -n "$ZITADEL_NAMESPACE"
    
    # Проверка сервисов
    log "Сервисы в namespace $ZITADEL_NAMESPACE:"
    kubectl get svc -n "$ZITADEL_NAMESPACE"
    
    # Проверка OIDC конфигурации
    log "Проверка OIDC конфигурации..."
    if curl -kLs "https://$ZITADEL_DOMAIN/.well-known/openid-configuration" | jq . > /dev/null; then
        log "OIDC конфигурация доступна"
        
        # Сохранение OIDC конфигурации
        curl -kLs "https://$ZITADEL_DOMAIN/.well-known/openid-configuration" > "$PROJECT_ROOT/configs/oidc-config.json"
        log "OIDC конфигурация сохранена в configs/oidc-config.json"
    else
        error "OIDC конфигурация недоступна"
    fi
    
    # Вывод URL доступа
    log "ZITADEL развернут успешно!"
    log "Информация для доступа:"
    log "   URL: https://$ZITADEL_DOMAIN"
    log "   Console: https://$ZITADEL_DOMAIN/ui/console"
    log "   Admin: zitadel-admin@$ZITADEL_DOMAIN"
    log ""
    log "📝 Следующие шаги:"
    log "   1. Войдите в консоль и создайте организацию '$ORG_NAME'"
    log "   2. Создайте проект '$PROJECT_NAME'"
    log "   3. Настройте приложения согласно документации"
    log "   4. Запустите скрипт интеграционного тестирования:"
    log "      ./scripts/integration-test.sh"
}

# Очистка в случае ошибки
cleanup_on_error() {
    error "Произошла ошибка при установке"
    warn "Для очистки выполните:"
    warn "  helm uninstall zitadel -n $ZITADEL_NAMESPACE"
    warn "  helm uninstall pg -n $ZITADEL_NAMESPACE"  
    warn "  kubectl delete namespace $ZITADEL_NAMESPACE"
}

# Основная функция
main() {
    log "Начало установки ZITADEL для CloudBridge Relay"
    log "Конфигурация:"
    log "   Домен: $ZITADEL_DOMAIN"
    log "   Namespace: $ZITADEL_NAMESPACE"
    log "   Организация: $ORG_NAME"
    log "   Проект: $PROJECT_NAME"
    log ""
    
    trap cleanup_on_error ERR
    
    check_prerequisites
    create_namespace
    generate_tls_certificates
    install_postgresql
    generate_masterkey
    install_zitadel
    create_ingress
    wait_for_zitadel
    setup_zitadel_config
    verify_deployment
    
    log "Установка ZITADEL завершена успешно!"
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
