#!/bin/bash

# CloudBridge Relay SSL/TLS Certificate Generation Script
# This script generates CA, server, and client certificates for CloudBridge Relay

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CERT_DIR="certs"
CONFIG_DIR="cert"
CA_KEY_SIZE=4096
SERVER_KEY_SIZE=2048
CLIENT_KEY_SIZE=2048
VALIDITY_DAYS=365

echo -e "${BLUE}🔒 CloudBridge Relay SSL/TLS Certificate Generation${NC}"
echo "=================================================="

# Create certificates directory
echo -e "${YELLOW}📁 Creating certificates directory...${NC}"
mkdir -p "$CERT_DIR"

# Check if OpenSSL is available
if ! command -v openssl &> /dev/null; then
    echo -e "${RED}❌ OpenSSL is not installed. Please install OpenSSL first.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ OpenSSL found: $(openssl version)${NC}"

# Generate CA private key
echo -e "${YELLOW}🔑 Generating CA private key...${NC}"
openssl genrsa -out "$CERT_DIR/ca.key" $CA_KEY_SIZE
chmod 600 "$CERT_DIR/ca.key"

# Generate CA certificate
echo -e "${YELLOW}📜 Generating CA certificate...${NC}"
openssl req -new -x509 -days $VALIDITY_DAYS -key "$CERT_DIR/ca.key" -out "$CERT_DIR/ca.crt" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/OU=CloudBridge/CN=CloudBridge CA"

# Generate server private key
echo -e "${YELLOW}🔑 Generating server private key...${NC}"
openssl genrsa -out "$CERT_DIR/server.key" $SERVER_KEY_SIZE
chmod 600 "$CERT_DIR/server.key"

# Generate server certificate signing request
echo -e "${YELLOW}📝 Generating server certificate signing request...${NC}"
openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" \
    -config "$CONFIG_DIR/openssl-cloudbridge.cnf"

# Generate server certificate
echo -e "${YELLOW}📜 Generating server certificate...${NC}"
openssl x509 -req -in "$CERT_DIR/server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial -out "$CERT_DIR/server.crt" -days $VALIDITY_DAYS \
    -extensions v3_ca -extfile "$CONFIG_DIR/openssl-cloudbridge.cnf"

# Generate client private key
echo -e "${YELLOW}🔑 Generating client private key...${NC}"
openssl genrsa -out "$CERT_DIR/client.key" $CLIENT_KEY_SIZE
chmod 600 "$CERT_DIR/client.key"

# Generate client certificate signing request
echo -e "${YELLOW}📝 Generating client certificate signing request...${NC}"
openssl req -new -key "$CERT_DIR/client.key" -out "$CERT_DIR/client.csr" \
    -subj "/C=RU/ST=Moscow/L=Moscow/O=2GC Technologies/OU=CloudBridge/CN=CloudBridge Client"

# Generate client certificate
echo -e "${YELLOW}📜 Generating client certificate...${NC}"
openssl x509 -req -in "$CERT_DIR/client.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" \
    -CAcreateserial -out "$CERT_DIR/client.crt" -days $VALIDITY_DAYS

# Generate Cassandra keystore and truststore
echo -e "${YELLOW}🗄️ Generating Cassandra keystore and truststore...${NC}"

# Create keystore for Cassandra
openssl pkcs12 -export -in "$CERT_DIR/server.crt" -inkey "$CERT_DIR/server.key" \
    -out "$CERT_DIR/cassandra-keystore.p12" -name "cassandra" -password pass:cassandra

# Create truststore for Cassandra
openssl pkcs12 -export -in "$CERT_DIR/ca.crt" -inkey "$CERT_DIR/ca.key" \
    -out "$CERT_DIR/cassandra-truststore.p12" -name "cassandra-ca" -password pass:cassandra

# Clean up CSR files
echo -e "${YELLOW}🧹 Cleaning up temporary files...${NC}"
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/client.csr"

# Set proper permissions
chmod 644 "$CERT_DIR"/*.crt
chmod 600 "$CERT_DIR"/*.key
chmod 644 "$CERT_DIR"/*.p12

# Display certificate information
echo -e "${GREEN}✅ Certificate generation completed successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Generated certificates:${NC}"
echo "  📜 CA Certificate: $CERT_DIR/ca.crt"
echo "  🔑 CA Private Key: $CERT_DIR/ca.key"
echo "  📜 Server Certificate: $CERT_DIR/server.crt"
echo "  🔑 Server Private Key: $CERT_DIR/server.key"
echo "  📜 Client Certificate: $CERT_DIR/client.crt"
echo "  🔑 Client Private Key: $CERT_DIR/client.key"
echo "  🗄️ Cassandra Keystore: $CERT_DIR/cassandra-keystore.p12"
echo "  🗄️ Cassandra Truststore: $CERT_DIR/cassandra-truststore.p12"
echo ""

# Display certificate details
echo -e "${BLUE}📋 Certificate Details:${NC}"
echo "CA Certificate:"
openssl x509 -in "$CERT_DIR/ca.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)"

echo ""
echo "Server Certificate:"
openssl x509 -in "$CERT_DIR/server.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After|DNS:|IP Address:)"

echo ""
echo "Client Certificate:"
openssl x509 -in "$CERT_DIR/client.crt" -text -noout | grep -E "(Subject:|Issuer:|Not Before|Not After)"

echo ""
echo -e "${GREEN}🎉 All certificates generated successfully!${NC}"
echo -e "${YELLOW}💡 Next steps:${NC}"
echo "  1. Update CloudBridge Relay configuration to enable TLS"
echo "  2. Create Kubernetes secrets for certificates"
echo "  3. Update deployment manifests to mount certificates"
echo "  4. Test TLS connections"
