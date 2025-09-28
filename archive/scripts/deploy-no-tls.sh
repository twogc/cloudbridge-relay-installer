#!/bin/bash

echo "🔍 Применяем манифесты для gRPC без TLS..."

# Применяем ConfigMap
kubectl apply -f k8s/configmap-no-tls.yaml

# Применяем Deployment
kubectl apply -f k8s/deployment-no-tls.yaml

# Применяем Service
kubectl apply -f k8s/service-no-tls.yaml

echo "✅ Манифесты применены. Проверяем статус..."

# Проверяем статус
kubectl get pods -n cloudbridge | grep no-tls

echo "🔍 Ждем запуска пода..."
kubectl wait --for=condition=ready pod -l app=cloudbridge-relay-no-tls -n cloudbridge --timeout=60s

echo "✅ Готово! Проверяем логи..."
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=10
