#!/bin/bash

echo "ПОЛНОЕ ИСПРАВЛЕНИЕ gRPC ПОДКЛЮЧЕНИЯ"
echo "======================================="

cd /home/ubuntu/cloudbridge-relay-installer

echo "Шаг 1: Применяем ConfigMap без TLS..."
kubectl apply -f k8s/configmap-no-tls.yaml

echo "Шаг 2: Применяем Deployment без TLS..."
kubectl apply -f k8s/deployment-no-tls.yaml

echo "Шаг 3: Применяем Service без TLS..."
kubectl apply -f k8s/service-no-tls.yaml

echo "Шаг 4: Ждем запуска пода (60 сек)..."
kubectl wait --for=condition=ready pod -l app=cloudbridge-relay-no-tls -n cloudbridge --timeout=60s

echo "Шаг 5: Проверяем статус подов..."
kubectl get pods -n cloudbridge | grep no-tls

echo "Шаг 6: Проверяем сервисы..."
kubectl get services -n cloudbridge | grep no-tls

echo "Шаг 7: Проверяем логи запуска..."
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=20 | grep -i grpc

echo "Шаг 8: Настраиваем port-forward..."
pkill -f "kubectl port-forward.*8444" 2>/dev/null || true
sleep 2
nohup kubectl port-forward -n cloudbridge service/cloudbridge-relay-no-tls 8444:8444 > port-forward-no-tls.log 2>&1 &
sleep 3

echo "Шаг 9: Проверяем доступность порта..."
netstat -tlnp | grep 8444 || echo "Порт 8444 не доступен"

echo "🧪 Шаг 10: Тестируем gRPC подключение..."
cd test/relay-client

# Создаем простую конфигурацию без TLS
cat > config-simple-no-tls.yaml << 'EOF'
relay:
  host: "localhost"
  ports:
    grpc: 8444
  timeout: "30s"
  tls:
    enabled: false

auth:
  type: "none"
  skip_validation: true

api:
  base_url: "grpc://localhost:8444"
  timeout: "30s"
  insecure_skip_verify: true

logging:
  level: "debug"
EOF

echo "Шаг 11: ЗАПУСКАЕМ КЛИЕНТ..."
echo "Конфигурация клиента:"
cat config-simple-no-tls.yaml

echo ""
echo "🔥 ТЕСТИРУЕМ ПОДКЛЮЧЕНИЕ..."
timeout 20s ./cloudbridge-client tunnel \
  --config config-simple-no-tls.yaml \
  --tunnel-id test-tunnel \
  --transport grpc \
  --insecure-skip-tls-verify \
  --log-level debug \
  --local-port 8080 \
  --remote-host 127.0.0.1 \
  --remote-port 80

echo ""
echo "ТЕСТ ЗАВЕРШЕН!"
echo "Если есть ошибки, проверяем логи сервера:"
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=10
