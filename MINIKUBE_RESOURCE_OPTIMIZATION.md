# 🔧 Оптимизация ресурсов Minikube для стабильной работы

## ❌ Проблема
При сборке образов в Minikube перестают работать текущие поды (Zitadel отвечает 503).

## 🔍 Причины
1. **Конкуренция за ресурсы** - сборка Docker образов создает высокую нагрузку
2. **Перегрузка VM** - Load Average > 30, постоянные рестарты подов
3. **Проблемы с образами** - `ErrImageNeverPull` после пересборки

## ✅ Решения

### 1. **Увеличить ресурсы Minikube**
```bash
# Остановить Minikube
minikube stop

# Перезапустить с увеличенными ресурсами
minikube start \
  --driver=docker \
  --memory=8192 \
  --cpus=4 \
  --disk-size=50g \
  --container-runtime=containerd

# Проверить новые ресурсы
minikube status
kubectl top nodes
```

### 2. **Оптимизировать процесс сборки**
```bash
# Создать скрипт безопасной сборки
cat > /home/ubuntu/safe-build.sh << 'EOF'
#!/bin/bash

echo "=== БЕЗОПАСНАЯ СБОРКА В MINIKUBE ==="

# 1. Проверяем нагрузку системы
LOAD=$(uptime | awk -F'load average:' '{ print $2 }' | awk '{ print $1 }' | sed 's/,//')
if (( $(echo "$LOAD > 2.0" | bc -l) )); then
    echo "⚠️  Высокая нагрузка: $LOAD. Ждем снижения..."
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

# 5. Собираем с ограничением ресурсов
echo "Собираем образ..."
docker build \
  --memory=2g \
  --cpus=2 \
  -t cloudbridge/ai-analytics:$TAG \
  -f docker/images/Dockerfile.ai-analytics .

if [ $? -eq 0 ]; then
    echo "✅ Образ собран: $TAG"
    echo "TAG=$TAG" > .last-build-tag
    
    # 6. Обновляем deployment
    kubectl set image deployment/ai-analytics \
      ai-analytics=cloudbridge/ai-analytics:$TAG \
      -n cloudbridge
    
    echo "✅ Deployment обновлен"
else
    echo "❌ Ошибка сборки"
    exit 1
fi
EOF

chmod +x /home/ubuntu/safe-build.sh
```

### 3. **Настроить мониторинг ресурсов**
```bash
# Скрипт для мониторинга во время сборки
cat > /home/ubuntu/monitor-build.sh << 'EOF'
#!/bin/bash

echo "=== МОНИТОРИНГ СБОРКИ ==="

while true; do
    echo "$(date): Load: $(uptime | awk -F'load average:' '{print $2}')"
    echo "Memory: $(free -h | grep Mem | awk '{print $3"/"$2}')"
    echo "Pods status:"
    kubectl get pods -n zitadel --no-headers | grep -v Running | wc -l | xargs echo "Non-running pods:"
    echo "---"
    sleep 30
done
EOF

chmod +x /home/ubuntu/monitor-build.sh
```

### 4. **Исправить проблемы с образами**
```bash
# Очистить старые образы
docker image prune -f

# Проверить образы в Minikube
minikube image ls | grep cloudbridge

# Удалить поды с ErrImageNeverPull
kubectl delete pods -l app=ai-service -n cloudbridge --field-selector=status.phase=Failed
kubectl delete pods -l app=ai-telegram-scheduler -n cloudbridge --field-selector=status.phase=Failed
```

### 5. **Использовать раздельную сборку**
```bash
# Собирать образы вне Minikube, затем загружать
# 1. Обычная сборка
docker build -t cloudbridge/ai-analytics:$TAG .

# 2. Загрузка в Minikube
minikube image load cloudbridge/ai-analytics:$TAG

# 3. Обновление deployment
kubectl set image deployment/ai-analytics ai-analytics=cloudbridge/ai-analytics:$TAG -n cloudbridge
```

## 📊 Мониторинг
```bash
# Проверка здоровья системы
watch -n 5 'echo "=== SYSTEM HEALTH ===" && uptime && echo && kubectl get pods -n zitadel | grep -v Running'
```

## 🎯 Результат
- Стабильная работа Zitadel и других сервисов
- Отсутствие 503 ошибок
- Успешная сборка без влияния на production поды
