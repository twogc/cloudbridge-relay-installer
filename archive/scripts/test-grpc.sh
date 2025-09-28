#!/bin/bash

# Простой тест gRPC подключения
echo "🔍 Тестируем gRPC подключение..."

# Обновляем конфигурацию клиента для использования NodePort 30845
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client

# Создаем временную конфигурацию
cat > config-no-tls.yaml << 'EOF'
# Relay Configuration
relay:
  host: "192.168.49.2"
  ports:
    http_api: 30081
    grpc: 30845
  timeout: "30s"
  tls:
    enabled: false

# Authentication Configuration  
auth:
  type: "jwt"
  secret: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
  skip_validation: true

# API Configuration
api:
  base_url: "grpc://192.168.49.2:30845"
  timeout: "30s"
  insecure_skip_verify: true
EOF

echo "✅ Конфигурация создана. Тестируем клиент..."

# Тестируем клиент
timeout 15s ./cloudbridge-client tunnel --config config-no-tls.yaml --token simple-token.txt --tunnel-id test-tunnel --transport grpc --insecure-skip-tls-verify --log-level debug
