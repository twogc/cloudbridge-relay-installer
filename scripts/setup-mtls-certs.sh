#!/bin/bash
set -euo pipefail

# CloudBridge Relay mTLS Certificate Setup
# Скрипт для генерации сертификатов для mTLS аутентификации с ZITADEL

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Конфигурация
CERT_DIR="$PROJECT_ROOT/certs/relay-mtls"
CA_NAME="cloudbridge-relay-ca"
SERVER_NAME="relay-server"
CLIENT_NAME="relay-client"

# Domains для SAN
RELAY_DOMAINS=(
    "relay.local"
    "relay.2gc.space"  
    "b1.2gc.space"
    "*.relay.2gc.space"
    "localhost"
)

# IP адреса для SAN
RELAY_IPS=(
    "127.0.0.1"
    "::1"
    "192.168.1.1"
    "10.0.0.1"
)

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $*${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $*${NC}"; }
info() { echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $*${NC}"; }

# Создание директории для сертификатов
setup_cert_directory() {
    log "Настройка директории сертификатов..."
    
    mkdir -p "$CERT_DIR"
    chmod 755 "$CERT_DIR"
    
    log "Директория создана: $CERT_DIR"
}

# Генерация CA сертификата
generate_ca_certificate() {
    log "Генерация корневого CA сертификата..."
    
    local ca_key="$CERT_DIR/${CA_NAME}-key.pem"
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    
    if [[ -f "$ca_cert" && -f "$ca_key" ]]; then
        warn "CA сертификат уже существует, пропускаем генерацию"
        return 0
    fi
    
    # Генерация приватного ключа CA
    openssl genrsa -out "$ca_key" 4096
    chmod 600 "$ca_key"
    
    # Создание конфигурации CA
    cat > "$CERT_DIR/ca.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow
O = CloudBridge Relay
OU = Certificate Authority
CN = CloudBridge Relay CA
emailAddress = ca@cloudbridge.local

[v3_ca]
basicConstraints = critical,CA:true
keyUsage = critical, digitalSignature, cRLSign, keyCertSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer
EOF

    # Генерация самоподписанного CA сертификата
    openssl req -new -x509 -days 3650 -key "$ca_key" -out "$ca_cert" \
        -config "$CERT_DIR/ca.conf" -extensions v3_ca
    
    log "CA сертификат создан:"
    log "   Ключ: $ca_key"
    log "   Сертификат: $ca_cert"
    
    # Показать информацию о сертификате
    openssl x509 -in "$ca_cert" -text -noout | grep -E "(Subject:|Not After)"
}

# Генерация серверного сертификата
generate_server_certificate() {
    log "Генерация серверного сертификата для relay-server..."
    
    local server_key="$CERT_DIR/${SERVER_NAME}-key.pem"
    local server_csr="$CERT_DIR/${SERVER_NAME}.csr"
    local server_cert="$CERT_DIR/${SERVER_NAME}-cert.pem"
    local ca_key="$CERT_DIR/${CA_NAME}-key.pem"
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    
    if [[ -f "$server_cert" && -f "$server_key" ]]; then
        warn "Серверный сертификат уже существует, пропускаем генерацию"
        return 0
    fi
    
    # Генерация приватного ключа сервера
    openssl genrsa -out "$server_key" 2048
    chmod 600 "$server_key"
    
    # Создание SAN списка
    local san_dns=""
    local san_ip=""
    
    for domain in "${RELAY_DOMAINS[@]}"; do
        if [[ -z "$san_dns" ]]; then
            san_dns="DNS:$domain"
        else
            san_dns="$san_dns,DNS:$domain"
        fi
    done
    
    for ip in "${RELAY_IPS[@]}"; do
        if [[ -z "$san_ip" ]]; then
            san_ip="IP:$ip"
        else
            san_ip="$san_ip,IP:$ip"
        fi
    done
    
    # Создание конфигурации сервера
    cat > "$CERT_DIR/server.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow  
O = CloudBridge Relay
OU = Server
CN = $SERVER_NAME
emailAddress = server@cloudbridge.local

[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = $san_dns,$san_ip
EOF

    # Генерация CSR
    openssl req -new -key "$server_key" -out "$server_csr" \
        -config "$CERT_DIR/server.conf"
    
    # Подписание сертификата CA
    openssl x509 -req -in "$server_csr" -CA "$ca_cert" -CAkey "$ca_key" \
        -CAcreateserial -out "$server_cert" -days 365 \
        -extensions v3_req -extfile "$CERT_DIR/server.conf"
    
    # Удаление CSR
    rm -f "$server_csr"
    
    log "Серверный сертификат создан:"
    log "   Ключ: $server_key"
    log "   Сертификат: $server_cert"
    
    # Показать SAN
    openssl x509 -in "$server_cert" -text -noout | grep -A1 "Subject Alternative Name"
}

# Генерация клиентского сертификата  
generate_client_certificate() {
    log "Генерация клиентского сертификата для relay-client..."
    
    local client_key="$CERT_DIR/${CLIENT_NAME}-key.pem"
    local client_csr="$CERT_DIR/${CLIENT_NAME}.csr"
    local client_cert="$CERT_DIR/${CLIENT_NAME}-cert.pem"
    local ca_key="$CERT_DIR/${CA_NAME}-key.pem"
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    
    if [[ -f "$client_cert" && -f "$client_key" ]]; then
        warn "Клиентский сертификат уже существует, пропускаем генерацию"
        return 0
    fi
    
    # Генерация приватного ключа клиента
    openssl genrsa -out "$client_key" 2048
    chmod 600 "$client_key"
    
    # Создание конфигурации клиента
    cat > "$CERT_DIR/client.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = RU
ST = Moscow
L = Moscow
O = CloudBridge Relay
OU = Client
CN = $CLIENT_NAME
emailAddress = client@cloudbridge.local

[v3_req]  
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
subjectAltName = DNS:$CLIENT_NAME,DNS:relay-client.local
EOF

    # Генерация CSR
    openssl req -new -key "$client_key" -out "$client_csr" \
        -config "$CERT_DIR/client.conf"
    
    # Подписание сертификата CA
    openssl x509 -req -in "$client_csr" -CA "$ca_cert" -CAkey "$ca_key" \
        -CAcreateserial -out "$client_cert" -days 365 \
        -extensions v3_req -extfile "$CERT_DIR/client.conf"
    
    # Удаление CSR
    rm -f "$client_csr"
    
    log "Клиентский сертификат создан:"
    log "   Ключ: $client_key" 
    log "   Сертификат: $client_cert"
}

# Генерация PKCS#12 файлов
generate_pkcs12_files() {
    log "Генерация PKCS#12 файлов..."
    
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    local server_key="$CERT_DIR/${SERVER_NAME}-key.pem"
    local server_cert="$CERT_DIR/${SERVER_NAME}-cert.pem"
    local client_key="$CERT_DIR/${CLIENT_NAME}-key.pem"
    local client_cert="$CERT_DIR/${CLIENT_NAME}-cert.pem"
    
    # PKCS#12 для сервера
    openssl pkcs12 -export -out "$CERT_DIR/${SERVER_NAME}.p12" \
        -inkey "$server_key" -in "$server_cert" -certfile "$ca_cert" \
        -name "$SERVER_NAME" -passout pass:cloudbridge
    
    # PKCS#12 для клиента
    openssl pkcs12 -export -out "$CERT_DIR/${CLIENT_NAME}.p12" \
        -inkey "$client_key" -in "$client_cert" -certfile "$ca_cert" \
        -name "$CLIENT_NAME" -passout pass:cloudbridge
    
    log "PKCS#12 файлы созданы:"
    log "   Сервер: $CERT_DIR/${SERVER_NAME}.p12"
    log "   Клиент: $CERT_DIR/${CLIENT_NAME}.p12"
    log "   Пароль: cloudbridge"
}

# Создание Kubernetes секретов
create_kubernetes_secrets() {
    log "Создание Kubernetes секретов..."
    
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    local server_key="$CERT_DIR/${SERVER_NAME}-key.pem"
    local server_cert="$CERT_DIR/${SERVER_NAME}-cert.pem"
    local client_key="$CERT_DIR/${CLIENT_NAME}-key.pem"
    local client_cert="$CERT_DIR/${CLIENT_NAME}-cert.pem"
    
    # Секрет для CA сертификата
    kubectl create secret generic relay-mtls-ca \
        --from-file=ca.crt="$ca_cert" \
        -n cloudbridge \
        --dry-run=client -o yaml > "$CERT_DIR/ca-secret.yaml"
    
    # Секрет для серверного сертификата
    kubectl create secret tls relay-mtls-server \
        --cert="$server_cert" \
        --key="$server_key" \
        -n cloudbridge \
        --dry-run=client -o yaml > "$CERT_DIR/server-secret.yaml"
    
    # Секрет для клиентского сертификата  
    kubectl create secret tls relay-mtls-client \
        --cert="$client_cert" \
        --key="$client_key" \
        -n cloudbridge \
        --dry-run=client -o yaml > "$CERT_DIR/client-secret.yaml"
    
    log "Kubernetes манифесты созданы:"
    log "   CA Secret: $CERT_DIR/ca-secret.yaml"
    log "   Server Secret: $CERT_DIR/server-secret.yaml"
    log "   Client Secret: $CERT_DIR/client-secret.yaml"
    log ""
    log "Для применения:"
    log "   kubectl apply -f $CERT_DIR/ca-secret.yaml"
    log "   kubectl apply -f $CERT_DIR/server-secret.yaml"
    log "   kubectl apply -f $CERT_DIR/client-secret.yaml"
}

# Генерация конфигурационных файлов
generate_config_files() {
    log "Генерация конфигурационных файлов..."
    
    # Go конфигурация для mTLS клиента
    cat > "$CERT_DIR/mtls-config.go" <<EOF
// CloudBridge Relay mTLS Configuration
package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// MTLSConfig конфигурация для mTLS соединения
type MTLSConfig struct {
	CACertFile     string
	ClientCertFile string
	ClientKeyFile  string
	ServerName     string
	InsecureSkipVerify bool
}

// LoadTLSConfig загружает TLS конфигурацию для mTLS
func (c *MTLSConfig) LoadTLSConfig() (*tls.Config, error) {
	// Загрузка CA сертификата
	caCert, err := ioutil.ReadFile(c.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}
	
	// Загрузка клиентского сертификата
	clientCert, err := tls.LoadX509KeyPair(c.ClientCertFile, c.ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}
	
	// Создание TLS конфигурации
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:           caCertPool,
		ServerName:        c.ServerName,
		InsecureSkipVerify: c.InsecureSkipVerify,
		MinVersion:        tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
	
	return tlsConfig, nil
}
EOF

    # YAML конфигурация для CloudBridge Relay
    cat > "$CERT_DIR/relay-mtls.yaml" <<EOF
# CloudBridge Relay mTLS Configuration

mtls:
  enabled: true
  ca_cert_file: "${CERT_DIR}/${CA_NAME}-cert.pem"
  server_cert_file: "${CERT_DIR}/${SERVER_NAME}-cert.pem"
  server_key_file: "${CERT_DIR}/${SERVER_NAME}-key.pem"
  client_cert_file: "${CERT_DIR}/${CLIENT_NAME}-cert.pem"
  client_key_file: "${CERT_DIR}/${CLIENT_NAME}-key.pem"
  server_name: "$SERVER_NAME"
  insecure_skip_verify: false
  
  # TLS настройки
  min_version: "1.2"
  max_version: "1.3"
  cipher_suites:
    - "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"
    - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
    - "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305"
    - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
  
  # Проверка сертификатов
  verify_client_cert: true
  require_client_cert: true
  
  # Разрешенные клиенты (по CN)
  allowed_clients:
    - "relay-client"
    - "cloudbridge-admin"
    - "kubectl-plugin"

zitadel_integration:
  # RFC 8705: OAuth 2.0 Mutual-TLS Client Authentication
  mtls_endpoint_alias: "https://edge.2gc.ru:8443"
  client_cert_bound_access_tokens: true
  
  # Token Exchange настройки
  token_exchange:
    enabled: true
    subject_token_type: "urn:ietf:params:oauth:token-type:access_token"
    requested_token_type: "urn:ietf:params:oauth:token-type:access_token"
    audience: "cloudbridge-relay"
EOF

    log "Конфигурационные файлы созданы:"
    log "   Go config: $CERT_DIR/mtls-config.go"
    log "   YAML config: $CERT_DIR/relay-mtls.yaml"
}

# Проверка сертификатов
verify_certificates() {
    log "Проверка созданных сертификатов..."
    
    local ca_cert="$CERT_DIR/${CA_NAME}-cert.pem"
    local server_cert="$CERT_DIR/${SERVER_NAME}-cert.pem"
    local client_cert="$CERT_DIR/${CLIENT_NAME}-cert.pem"
    
    # Проверка CA сертификата
    if openssl x509 -in "$ca_cert" -text -noout | grep -q "CA:TRUE"; then
        log "CA сертификат валиден"
    else
        error "CA сертификат невалиден"
    fi
    
    # Проверка серверного сертификата
    if openssl verify -CAfile "$ca_cert" "$server_cert" &>/dev/null; then
        log "Серверный сертификат валиден"
    else
        error "Серверный сертификат невалиден"
    fi
    
    # Проверка клиентского сертификата
    if openssl verify -CAfile "$ca_cert" "$client_cert" &>/dev/null; then
        log "Клиентский сертификат валиден"
    else
        error "Клиентский сертификат невалиден"
    fi
    
    # Показать сроки действия
    log "Сроки действия сертификатов:"
    echo "CA: $(openssl x509 -in "$ca_cert" -noout -enddate | cut -d= -f2)"
    echo "Server: $(openssl x509 -in "$server_cert" -noout -enddate | cut -d= -f2)"
    echo "Client: $(openssl x509 -in "$client_cert" -noout -enddate | cut -d= -f2)"
}

# Основная функция
main() {
    log "Настройка mTLS сертификатов для CloudBridge Relay"
    log "📁 Директория сертификатов: $CERT_DIR"
    log ""
    
    setup_cert_directory
    generate_ca_certificate
    generate_server_certificate
    generate_client_certificate
    generate_pkcs12_files
    create_kubernetes_secrets
    generate_config_files
    verify_certificates
    
    log "Настройка mTLS сертификатов завершена!"
    log ""
    log "📁 Созданные файлы:"
    log "   CA: $CERT_DIR/${CA_NAME}-{cert,key}.pem"
    log "    Server: $CERT_DIR/${SERVER_NAME}-{cert,key}.pem"
    log "   👤 Client: $CERT_DIR/${CLIENT_NAME}-{cert,key}.pem"
    log "   PKCS#12: $CERT_DIR/{${SERVER_NAME},${CLIENT_NAME}}.p12"
    log "   ⚙️  Configs: $CERT_DIR/{mtls-config.go,relay-mtls.yaml}"
    log "   ☸️  K8s: $CERT_DIR/{ca,server,client}-secret.yaml"
    log ""
    log "📝 Следующие шаги:"
    log "   1. Примените Kubernetes секреты:"
    log "      kubectl apply -f $CERT_DIR/*-secret.yaml"
    log "   2. Обновите конфигурацию CloudBridge Relay"
    log "   3. Настройте ZITADEL для использования mTLS"
    log "   4. Протестируйте mTLS соединение"
}

# Запуск скрипта
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
