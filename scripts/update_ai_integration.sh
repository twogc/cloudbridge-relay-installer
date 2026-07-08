#!/bin/bash

# Скрипт для обновления AI сервиса с интеграцией реальных данных

set -e

echo "Обновление AI сервиса с интеграцией реальных данных..."

# Переходим в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Останавливаем AI сервисы
echo "⏹️ Останавливаем AI сервисы..."
kubectl scale deployment ai-service -n cloudbridge --replicas=0
kubectl scale deployment ai-worker -n cloudbridge --replicas=0

# Ждем завершения остановки
echo "Ждем завершения остановки..."
kubectl wait --for=delete pod -l app.kubernetes.io/name=ai-service -n cloudbridge --timeout=60s
kubectl wait --for=delete pod -l app.kubernetes.io/name=ai-worker -n cloudbridge --timeout=60s

# Собираем новый образ AI сервиса
echo "🔨 Собираем новый образ AI сервиса..."
docker build -f Dockerfile.ai-service -t ai-service:latest .

# Загружаем образ в Minikube
echo "Загружаем образ в Minikube..."
minikube image load ai-service:latest

# Обновляем конфигурацию AI
echo "⚙️ Обновляем конфигурацию AI..."
kubectl create configmap ai-config --from-file=config/ai-config.yaml -n cloudbridge --dry-run=client -o yaml | kubectl apply -f -

# Запускаем AI сервисы
echo "▶️ Запускаем AI сервисы..."
kubectl scale deployment ai-service -n cloudbridge --replicas=2
kubectl scale deployment ai-worker -n cloudbridge --replicas=2

# Ждем готовности
echo "Ждем готовности сервисов..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ai-service -n cloudbridge --timeout=120s
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=ai-worker -n cloudbridge --timeout=120s

# Проверяем статус
echo "Проверяем статус сервисов..."
kubectl get pods -n cloudbridge | grep ai

echo "AI сервис обновлен с интеграцией реальных данных!"
echo "Теперь AI будет получать реальные метрики от Relay сервера"
echo "Интервал обработки увеличен до 60 секунд"
echo " Тестовый режим отключен"

# Показываем логи
echo "Последние логи AI сервиса:"
kubectl logs -n cloudbridge -l app.kubernetes.io/name=ai-service --tail=10
