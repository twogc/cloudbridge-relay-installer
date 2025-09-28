#!/bin/bash

# Генерация самоподписанных сертификатов для разработки CloudBridge

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Конфигурация
CERT_DIR="./certs"
CERT_FILE="$CERT_DIR/dev-cert.pem"
KEY_FILE="$CERT_DIR/dev-key.pem"
DAYS_VALID=365

# Информация о сертификате
COUNTRY="US"
STATE="California"
CITY="San Francisco"
ORGANIZATION="CloudBridge Development"
ORGANIZATIONAL_UNIT="Development Team"
COMMON_NAME="localhost"
EMAIL="dev@cloudbridge.local"

# Дополнительные домены для SAN (Subject Alternative Names)
SAN_DOMAINS="localhost,127.0.0.1,::1,cloudbridge.local,api.cloudbridge.local"

# Создание директории для сертификатов
create_cert_directory() {
    info "Создание директории для сертификатов..."
    
    if [ ! -d "$CERT_DIR" ]; then
        mkdir -p "$CERT_DIR"
        success "Директория создана: $CERT_DIR"
    else
        info "Директория уже существует: $CERT_DIR"
    fi
}

# Проверка существующих сертификатов
check_existing_certs() {
    if [ -f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
        warning "Сертификаты уже существуют:"
        echo "  Сертификат: $CERT_FILE"
        echo "  Ключ: $KEY_FILE"
        
        # Проверяем срок действия
        if openssl x509 -checkend 86400 -noout -in "$CERT_FILE" >/dev/null 2>&1; then
            info "Сертификат действителен еще минимум 24 часа"
            
            read -p "Перегенерировать сертификаты? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                info "Использование существующих сертификатов"
                exit 0
            fi
        else
            warning "Сертификат истек или истекает в течение 24 часов"
        fi
    fi
}

# Создание конфигурационного файла OpenSSL
create_openssl_config() {
    info "Создание конфигурационного файла OpenSSL..."
    
    cat > "$CERT_DIR/openssl.conf" << EOF
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = dn
req_extensions = v3_req

[dn]
C=$COUNTRY
ST=$STATE
L=$CITY
O=$ORGANIZATION
OU=$ORGANIZATIONAL_UNIT
CN=$COMMON_NAME
emailAddress=$EMAIL

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
EOF

    # Добавляем SAN домены
    IFS=',' read -ra DOMAINS <<< "$SAN_DOMAINS"
    DNS_COUNT=1
    IP_COUNT=1
    
    for domain in "${DOMAINS[@]}"; do
        domain=$(echo "$domain" | xargs) # trim whitespace
        
        # Проверяем, является ли это IP адресом
        if [[ $domain =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]] || [[ $domain == "::1" ]]; then
            echo "IP.$IP_COUNT = $domain" >> "$CERT_DIR/openssl.conf"
            ((IP_COUNT++))
        else
            echo "DNS.$DNS_COUNT = $domain" >> "$CERT_DIR/openssl.conf"
            ((DNS_COUNT++))
        fi
    done
    
    success "Конфигурационный файл создан: $CERT_DIR/openssl.conf"
}

# Генерация приватного ключа
generate_private_key() {
    info "Генерация приватного ключа..."
    
    openssl genrsa -out "$KEY_FILE" 2048
    chmod 600 "$KEY_FILE"
    
    success "Приватный ключ создан: $KEY_FILE"
}

# Генерация самоподписанного сертификата
generate_certificate() {
    info "Генерация самоподписанного сертификата..."
    
    openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" \
        -days $DAYS_VALID \
        -config "$CERT_DIR/openssl.conf" \
        -extensions v3_req
    
    chmod 644 "$CERT_FILE"
    
    success "Сертификат создан: $CERT_FILE"
}

# Проверка созданного сертификата
verify_certificate() {
    info "Проверка созданного сертификата..."
    
    # Основная информация о сертификате
    echo "=== Информация о сертификате ==="
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Subject:"
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Issuer:"
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 2 "Validity"
    
    # Subject Alternative Names
    echo ""
    echo "=== Subject Alternative Names ==="
    openssl x509 -in "$CERT_FILE" -text -noout | grep -A 1 "Subject Alternative Name" || echo "SAN не найдены"
    
    # Проверка соответствия ключа и сертификата
    CERT_MODULUS=$(openssl x509 -noout -modulus -in "$CERT_FILE" | openssl md5)
    KEY_MODULUS=$(openssl rsa -noout -modulus -in "$KEY_FILE" | openssl md5)
    
    if [ "$CERT_MODULUS" = "$KEY_MODULUS" ]; then
        success "Сертификат и ключ соответствуют друг другу"
    else
        error "Сертификат и ключ НЕ соответствуют друг другу!"
        exit 1
    fi
}

# Создание информационного файла
create_info_file() {
    info "Создание информационного файла..."
    
    cat > "$CERT_DIR/README.md" << EOF
# CloudBridge Development Certificates

Этот каталог содержит самоподписанные сертификаты для разработки CloudBridge.

## Файлы

- \`dev-cert.pem\` - Сертификат (публичный ключ)
- \`dev-key.pem\` - Приватный ключ
- \`openssl.conf\` - Конфигурация OpenSSL, использованная для генерации

## Информация о сертификате

- **Срок действия**: $DAYS_VALID дней с $(date)
- **Общее имя (CN)**: $COMMON_NAME
- **Организация**: $ORGANIZATION
- **Поддерживаемые домены**: $SAN_DOMAINS

## Использование

### В браузере

1. Откройте https://localhost:8443
2. Браузер покажет предупреждение о небезопасном соединении
3. Нажмите "Advanced" -> "Proceed to localhost (unsafe)"

### Добавление в доверенные (опционально)

#### macOS
\`\`\`bash
sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_FILE
\`\`\`

#### Linux (Ubuntu/Debian)
\`\`\`bash
sudo cp $CERT_FILE /usr/local/share/ca-certificates/cloudbridge-dev.crt
sudo update-ca-certificates
\`\`\`

#### Windows
1. Откройте \`$CERT_FILE\` в проводнике
2. Нажмите "Install Certificate"
3. Выберите "Local Machine" -> "Place all certificates in the following store"
4. Выберите "Trusted Root Certification Authorities"

## Безопасность

⚠️ **ВАЖНО**: Эти сертификаты предназначены ТОЛЬКО для разработки!

- НЕ используйте в продакшене
- НЕ добавляйте в систему контроля версий
- Регулярно обновляйте (рекомендуется каждые 90 дней)

## Регенерация

Для создания новых сертификатов запустите:
\`\`\`bash
./scripts/generate-dev-certs.sh
\`\`\`

Создано: $(date)
EOF

    success "Информационный файл создан: $CERT_DIR/README.md"
}

# Создание .gitignore для сертификатов
create_gitignore() {
    if [ ! -f "$CERT_DIR/.gitignore" ]; then
        cat > "$CERT_DIR/.gitignore" << EOF
# Игнорируем все сертификаты и ключи
*.pem
*.crt
*.key
*.p12
*.pfx

# Но разрешаем README и конфигурацию
!README.md
!.gitignore
EOF
        success "Создан .gitignore для сертификатов"
    fi
}

# Тестирование HTTPS соединения
test_https_connection() {
    info "Тестирование HTTPS соединения..."
    
    # Проверяем, запущен ли сервер на порту 8443
    if ! nc -z localhost 8443 2>/dev/null; then
        warning "Сервер не запущен на порту 8443"
        info "Для тестирования запустите CloudBridge с HTTPS поддержкой"
        return
    fi
    
    # Тестируем соединение
    if curl -k -s https://localhost:8443/health >/dev/null 2>&1; then
        success "HTTPS соединение работает!"
    else
        warning "HTTPS соединение не удалось протестировать"
    fi
}

# Основная функция
main() {
    info "CloudBridge Development Certificate Generator"
    info "==========================================="
    
    # Проверяем наличие OpenSSL
    if ! command -v openssl &> /dev/null; then
        error "OpenSSL не найден. Установите OpenSSL для продолжения."
        exit 1
    fi
    
    create_cert_directory
    check_existing_certs
    create_openssl_config
    generate_private_key
    generate_certificate
    verify_certificate
    create_info_file
    create_gitignore
    
    success "Сертификаты для разработки успешно созданы!"
    echo ""
    info "Файлы сертификатов:"
    echo "  Сертификат: $CERT_FILE"
    echo "  Ключ: $KEY_FILE"
    echo ""
    info "Для использования в CloudBridge API обновите конфигурацию TLS:"
    echo "  CertFile: \"$CERT_FILE\""
    echo "  KeyFile: \"$KEY_FILE\""
    echo ""
    test_https_connection
}

# Запуск
main "$@"
