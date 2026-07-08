#!/bin/bash

echo "=== БЕЗОПАСНАЯ СБОРКА В MINIKUBE ==="

# 1. Проверяем нагрузку системы
LOAD=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
if (( $(echo "$LOAD > 2.0" | bc -l) )); then
    echo " Высокая нагрузка: $LOAD. Ждем снижения..."
    while (( $(echo "$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//') > 2.0" | bc -l) )); do
        sleep 10
    done
fi

# 2. Устанавливаем лимиты Docker
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# 3. Настраиваем Docker для Minikube
eval $(minikube docker-env)

# 4. Создаем уникальный тег
TAG="build-$(date +%Y%m%d-%H%M%S)"
echo "Используем тег: $TAG"

# 5. Переходим в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# 6. Собираем с ограничением ресурсов
echo "Собираем образ..."
docker build \
  --memory=2g \
  --cpus=2 \
  -t cloudbridge/ai-analytics:$TAG \
  -f docker/images/Dockerfile.ai-analytics .

if [ $? -eq 0 ]; then
    echo "Образ собран: $TAG"
    echo "TAG=$TAG" > .last-build-tag
    
    # 7. Обновляем deployment
    kubectl set image deployment/ai-analytics \
      ai-analytics=cloudbridge/ai-analytics:$TAG \
      -n cloudbridge
    
    echo "Deployment обновлен"
else
    echo "Ошибка сборки"
    exit 1
fi
