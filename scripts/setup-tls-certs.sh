#!/bin/bash

# Настройка TLS сертификатов для CloudBridge (edge.2gc.ru)

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
DOMAIN="edge.2gc.ru"
CERT_DIR="/etc/ssl/cloudbridge"
CERT_FILE="$CERT_DIR/$DOMAIN.crt"
KEY_FILE="$CERT_DIR/$DOMAIN.key"
CHAIN_FILE="$CERT_DIR/$DOMAIN-chain.crt"
FULLCHAIN_FILE="$CERT_DIR/$DOMAIN-fullchain.crt"

# Kubernetes конфигурация
K8S_NAMESPACE="cloudbridge"
K8S_SECRET_NAME="edge-2gc-ru-tls"

# Создание директории для сертификатов
create_cert_directory() {
    info "Создание директории для сертификатов..."
    
    if [ ! -d "$CERT_DIR" ]; then
        sudo mkdir -p "$CERT_DIR"
        sudo chmod 755 "$CERT_DIR"
        success "Директория создана: $CERT_DIR"
    else
        info "Директория уже существует: $CERT_DIR"
    fi
}

# Проверка существующих сертификатов
check_existing_certs() {
    info "Проверка существующих сертификатов для $DOMAIN..."
    
    if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
        success "Найдены существующие сертификаты:"
        echo "  Сертификат: $CERT_FILE"
        echo "  Ключ: $KEY_FILE"
        
        # Проверяем срок действия
        if openssl x509 -checkend 86400 -noout -in "$CERT_FILE" >/dev/null 2>&1; then
            EXPIRY_DATE=$(openssl x509 -enddate -noout -in "$CERT_FILE" | cut -d= -f2)
            success "Сертификат действителен до: $EXPIRY_DATE"
            return 0
        else
            warning "Сертификат истек или истекает в течение 24 часов"
            return 1
        fi
    else
        warning "Сертификаты для $DOMAIN не найдены"
        return 1
    fi
}

# Вариант 1: Let's Encrypt с Certbot
setup_letsencrypt() {
    info "=== ВАРИАНТ 1: Let's Encrypt (Certbot) ==="
    
    # Проверяем установку certbot
    if ! command -v certbot &> /dev/null; then
        info "Устанавливаем Certbot..."
        sudo apt update
        sudo apt install -y certbot python3-certbot-nginx
    fi
    
    info "Получение сертификата Let's Encrypt для $DOMAIN..."
    warning "ВНИМАНИЕ: Для этого нужно:"
    echo "1. Домен $DOMAIN должен указывать на этот сервер"
    echo "2. Порт 80 должен быть открыт"
    echo "3. Nginx или Apache должен быть настроен"
    
    read -p "Продолжить с Let's Encrypt? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Standalone режим (без веб-сервера)
        sudo certbot certonly --standalone \
            --preferred-challenges http \
            -d "$DOMAIN" \
            --email "admin@2gc.ru" \
            --agree-tos \
            --non-interactive
        
        if [ $? -eq 0 ]; then
            # Копируем сертификаты в нашу директорию
            sudo cp "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" "$FULLCHAIN_FILE"
            sudo cp "/etc/letsencrypt/live/$DOMAIN/privkey.pem" "$KEY_FILE"
            sudo cp "/etc/letsencrypt/live/$DOMAIN/cert.pem" "$CERT_FILE"
            sudo cp "/etc/letsencrypt/live/$DOMAIN/chain.pem" "$CHAIN_FILE"
            
            # Устанавливаем правильные права
            sudo chown root:ssl-cert "$CERT_FILE" "$KEY_FILE" "$CHAIN_FILE" "$FULLCHAIN_FILE"
            sudo chmod 644 "$CERT_FILE" "$CHAIN_FILE" "$FULLCHAIN_FILE"
            sudo chmod 640 "$KEY_FILE"
            
            success "Let's Encrypt сертификат установлен!"
            return 0
        else
            error "Не удалось получить Let's Encrypt сертификат"
            return 1
        fi
    fi
}

# Вариант 2: Загрузка существующих сертификатов
upload_existing_certs() {
    info "=== ВАРИАНТ 2: Загрузка существующих сертификатов ==="
    
    echo "Укажите пути к вашим сертификатам:"
    read -p "Путь к файлу сертификата (.crt/.pem): " USER_CERT_FILE
    read -p "Путь к файлу приватного ключа (.key): " USER_KEY_FILE
    read -p "Путь к файлу цепочки сертификатов (опционально): " USER_CHAIN_FILE
    
    # Проверяем файлы
    if [ ! -f "$USER_CERT_FILE" ]; then
        error "Файл сертификата не найден: $USER_CERT_FILE"
        return 1
    fi
    
    if [ ! -f "$USER_KEY_FILE" ]; then
        error "Файл ключа не найден: $USER_KEY_FILE"
        return 1
    fi
    
    # Проверяем валидность сертификата
    if ! openssl x509 -in "$USER_CERT_FILE" -text -noout >/dev/null 2>&1; then
        error "Неверный формат сертификата"
        return 1
    fi
    
    # Проверяем соответствие ключа и сертификата
    CERT_MODULUS=$(openssl x509 -noout -modulus -in "$USER_CERT_FILE" | openssl md5)
    KEY_MODULUS=$(openssl rsa -noout -modulus -in "$USER_KEY_FILE" | openssl md5)
    
    if [ "$CERT_MODULUS" != "$KEY_MODULUS" ]; then
        error "Сертификат и ключ не соответствуют друг другу!"
        return 1
    fi
    
    # Копируем файлы
    info "Копирование сертификатов..."
    sudo cp "$USER_CERT_FILE" "$CERT_FILE"
    sudo cp "$USER_KEY_FILE" "$KEY_FILE"
    
    if [ -f "$USER_CHAIN_FILE" ]; then
        sudo cp "$USER_CHAIN_FILE" "$CHAIN_FILE"
        # Создаем fullchain
        sudo cat "$CERT_FILE" "$CHAIN_FILE" > "$FULLCHAIN_FILE"
    else
        sudo cp "$CERT_FILE" "$FULLCHAIN_FILE"
    fi
    
    # Устанавливаем права
    sudo chown root:ssl-cert "$CERT_FILE" "$KEY_FILE" "$FULLCHAIN_FILE"
    sudo chmod 644 "$CERT_FILE" "$FULLCHAIN_FILE"
    sudo chmod 640 "$KEY_FILE"
    
    if [ -f "$CHAIN_FILE" ]; then
        sudo chown root:ssl-cert "$CHAIN_FILE"
        sudo chmod 644 "$CHAIN_FILE"
    fi
    
    success "Сертификаты успешно установлены!"
    return 0
}

# Вариант 3: Самоподписанный сертификат (только для тестирования)
create_self_signed() {
    info "=== ВАРИАНТ 3: Самоподписанный сертификат (ТОЛЬКО ДЛЯ ТЕСТИРОВАНИЯ) ==="
    
    warning "ВНИМАНИЕ: Самоподписанные сертификаты НЕ рекомендуются для продакшена!"
    
    read -p "Создать самоподписанный сертификат для тестирования? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        info "Создание самоподписанного сертификата..."
        
        # Создаем конфигурационный файл
        cat > /tmp/openssl-$DOMAIN.conf << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=RU
ST=Moscow
L=Moscow
O=2GC
OU=IT Department
CN=$DOMAIN
emailAddress=admin@2gc.ru

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = $DOMAIN
DNS.2 = *.$DOMAIN
DNS.3 = api.$DOMAIN
DNS.4 = www.$DOMAIN
EOF

        # Генерируем ключ и сертификат
        sudo openssl req -new -x509 -key /dev/stdin -out "$CERT_FILE" \
            -days 365 \
            -config /tmp/openssl-$DOMAIN.conf \
            -extensions v3_req << EOF
$(sudo openssl genrsa 2048 | sudo tee "$KEY_FILE")
EOF

        # Создаем fullchain (для самоподписанного это тот же файл)
        sudo cp "$CERT_FILE" "$FULLCHAIN_FILE"
        
        # Устанавливаем права
        sudo chown root:ssl-cert "$CERT_FILE" "$KEY_FILE" "$FULLCHAIN_FILE"
        sudo chmod 644 "$CERT_FILE" "$FULLCHAIN_FILE"
        sudo chmod 640 "$KEY_FILE"
        
        # Очищаем временный файл
        rm -f /tmp/openssl-$DOMAIN.conf
        
        success "Самоподписанный сертификат создан!"
        warning "Не забудьте заменить его на реальный сертификат в продакшене!"
        return 0
    fi
}

# Создание Kubernetes Secret
create_k8s_secret() {
    info "Создание Kubernetes Secret для TLS сертификатов..."
    
    if [ ! -f "$CERT_FILE" ] || [ ! -f "$KEY_FILE" ]; then
        error "Сертификаты не найдены. Сначала установите сертификаты."
        return 1
    fi
    
    # Удаляем существующий secret если есть
    kubectl delete secret "$K8S_SECRET_NAME" -n "$K8S_NAMESPACE" 2>/dev/null || true
    
    # Создаем новый secret
    kubectl create secret tls "$K8S_SECRET_NAME" \
        --cert="$FULLCHAIN_FILE" \
        --key="$KEY_FILE" \
        -n "$K8S_NAMESPACE"
    
    if [ $? -eq 0 ]; then
        success "Kubernetes Secret создан: $K8S_SECRET_NAME"
        
        # Показываем информацию о secret
        kubectl describe secret "$K8S_SECRET_NAME" -n "$K8S_NAMESPACE"
        return 0
    else
        error "Не удалось создать Kubernetes Secret"
        return 1
    fi
}

# Обновление CloudBridge конфигурации
update_cloudbridge_config() {
    info "Обновление конфигурации CloudBridge..."
    
    # Создаем ConfigMap с TLS конфигурацией
    cat > /tmp/tls-config.yaml << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cloudbridge-tls-config
  namespace: $K8S_NAMESPACE
data:
  tls-config.yaml: |
    tls:
      cert_file: "/etc/ssl/certs/tls.crt"
      key_file: "/etc/ssl/private/tls.key"
      min_version: 1.2
      max_version: 1.3
      auto_redirect: true
      http_port: 80
      https_port: 443
    security:
      environment: "production"
      https:
        force_https: true
        hsts_max_age: 31536000
        hsts_include_subdomain: true
        hsts_preload: true
      cors:
        allowed_origins:
          - "https://$DOMAIN"
          - "https://api.$DOMAIN"
          - "https://www.$DOMAIN"
EOF

    kubectl apply -f /tmp/tls-config.yaml
    rm -f /tmp/tls-config.yaml
    
    success "Конфигурация CloudBridge обновлена"
}

# Проверка сертификата
verify_certificate() {
    info "Проверка установленного сертификата..."
    
    if [ ! -f "$CERT_FILE" ]; then
        error "Сертификат не найден: $CERT_FILE"
        return 1
    fi
    
    echo "=== Информация о сертификате ==="
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Subject:"
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Issuer:"
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 2 "Validity"
    
    echo ""
    echo "=== Subject Alternative Names ==="
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Subject Alternative Name" || echo "SAN не найдены"
    
    echo ""
    echo "=== Проверка домена ==="
    if openssl x509 -in "$CERT_FILE" -text -noout | grep -q "$DOMAIN"; then
        success "Сертификат содержит домен $DOMAIN"
    else
        warning "Сертификат НЕ содержит домен $DOMAIN"
    fi
}

# Тестирование HTTPS
test_https() {
    info "Тестирование HTTPS соединения..."
    
    # Проверяем, запущен ли CloudBridge
    if kubectl get pods -n "$K8S_NAMESPACE" | grep -q "cloudbridge-relay.*Running"; then
        info "CloudBridge запущен, настраиваем port-forward для тестирования..."
        
        # Port forward для HTTPS
        kubectl port-forward -n "$K8S_NAMESPACE" svc/cloudbridge-relay-external 8443:8443 &
        PF_PID=$!
        sleep 3
        
        # Тестируем соединение
        if curl -k -s https://localhost:8443/health >/dev/null 2>&1; then
            success "HTTPS соединение работает!"
        else
            warning "HTTPS соединение не удалось протестировать"
        fi
        
        # Останавливаем port-forward
        kill $PF_PID 2>/dev/null || true
    else
        warning "CloudBridge не запущен. Запустите его для тестирования HTTPS."
    fi
}

# Создание инструкции по обновлению deployment
create_deployment_instructions() {
    info "Создание инструкций по обновлению deployment..."
    
    cat > "update-deployment-tls.md" << EOF
# Обновление CloudBridge Deployment для TLS

## 1. Обновите deployment для использования TLS secret:

\`\`\`yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: $K8S_NAMESPACE
spec:
  template:
    spec:
      containers:
      - name: cloudbridge-relay
        volumeMounts:
        - name: tls-certs
          mountPath: /etc/ssl/certs
          readOnly: true
        - name: tls-config
          mountPath: /etc/cloudbridge/tls
          readOnly: true
        env:
        - name: TLS_CERT_FILE
          value: "/etc/ssl/certs/tls.crt"
        - name: TLS_KEY_FILE
          value: "/etc/ssl/certs/tls.key"
        - name: CLOUDBRIDGE_ENV
          value: "production"
      volumes:
      - name: tls-certs
        secret:
          secretName: $K8S_SECRET_NAME
      - name: tls-config
        configMap:
          name: cloudbridge-tls-config
\`\`\`

## 2. Примените изменения:

\`\`\`bash
kubectl apply -f k8s/deployment.yaml
kubectl rollout restart deployment/cloudbridge-relay -n $K8S_NAMESPACE
\`\`\`

## 3. Проверьте статус:

\`\`\`bash
kubectl rollout status deployment/cloudbridge-relay -n $K8S_NAMESPACE
kubectl logs -f deployment/cloudbridge-relay -n $K8S_NAMESPACE
\`\`\`
EOF

    success "Инструкции сохранены в: update-deployment-tls.md"
}

# Основная функция
main() {
    info "CloudBridge TLS Certificate Setup"
    info "================================="
    info "Домен: $DOMAIN"
    echo ""
    
    create_cert_directory
    
    if check_existing_certs; then
        info "Сертификаты уже установлены и действительны."
        read -p "Переустановить сертификаты? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Используем существующие сертификаты"
            verify_certificate
            create_k8s_secret
            update_cloudbridge_config
            create_deployment_instructions
            test_https
            return 0
        fi
    fi
    
    echo "Выберите способ установки сертификатов:"
    echo "1) Let's Encrypt (автоматически, требует DNS и открытый порт 80)"
    echo "2) Загрузить существующие сертификаты"
    echo "3) Создать самоподписанный сертификат (только для тестирования)"
    echo "4) Выход"
    
    read -p "Ваш выбор (1-4): " -n 1 -r
    echo
    
    case $REPLY in
        1)
            if setup_letsencrypt; then
                verify_certificate
                create_k8s_secret
                update_cloudbridge_config
                create_deployment_instructions
                test_https
            fi
            ;;
        2)
            if upload_existing_certs; then
                verify_certificate
                create_k8s_secret
                update_cloudbridge_config
                create_deployment_instructions
                test_https
            fi
            ;;
        3)
            if create_self_signed; then
                verify_certificate
                create_k8s_secret
                update_cloudbridge_config
                create_deployment_instructions
                test_https
            fi
            ;;
        4)
            info "Выход..."
            exit 0
            ;;
        *)
            error "Неверный выбор"
            exit 1
            ;;
    esac
    
    success "Настройка TLS сертификатов завершена!"
    echo ""
    security_info "Следующие шаги:"
    echo "1. Обновите deployment согласно инструкциям в update-deployment-tls.md"
    echo "2. Проверьте работу HTTPS endpoints"
    echo "3. Настройте автоматическое обновление сертификатов (для Let's Encrypt)"
}

# Запуск
main "$@"
