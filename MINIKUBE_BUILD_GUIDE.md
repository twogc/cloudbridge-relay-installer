# 🐳 **CloudBridge Relay - Быстрая сборка в Minikube**

## ⚠️ **ВАЖНО: Уникальные теги образов**

**ОБЯЗАТЕЛЬНО** используйте уникальный тег для каждого Docker образа! Kubernetes не будет использовать новый образ, если тег не изменился.

```bash
# Всегда создавайте уникальный тег
TAG="mvp-$(date +%Y%m%d-%H%M%S)"
echo "Используем тег: $TAG"
```

## 🎯 **Цель**

Быстро развернуть CloudBridge Relay в Minikube для разработки и тестирования QUIC-first + MASQUE архитектуры.

## ⚡ **Быстрый старт**

### **1. Установка Minikube (если не установлен):**
```bash
# Установить kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Установить Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

### **2. Запуск Minikube:**
```bash
# Запустить Minikube
minikube start --driver=docker --memory=4096 --cpus=2

# Проверить статус
minikube status
```

## 🏗️ **Сборка и развертывание**

### **⚠️ ВАЖНО: Уникальный тег образа**

**ОБЯЗАТЕЛЬНО** используйте уникальный тег для каждого образа, иначе Kubernetes не будет использовать новый образ!

```bash
# Создать уникальный тег (например, с датой и временем)
TAG="mvp-$(date +%Y%m%d-%H%M%S)"
echo "Используем тег: $TAG"
```

### **1. Сборка Docker образа:**

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать образ с уникальным тегом
docker build -t cloudbridge-relay:$TAG .

# Проверить собранный образ
docker images | grep cloudbridge-relay
```

### **2. Загрузка в Minikube:**

```bash
# Загрузить образ в Minikube
minikube image load cloudbridge-relay:$TAG

# Проверить загруженные образы
minikube image ls | grep cloudbridge-relay
```

### **3. Развертывание в Kubernetes:**

```bash
# Создать namespace
kubectl create namespace cloudbridge

# Обновить deployment с новым тегом
sed "s|image: cloudbridge-relay:.*|image: cloudbridge-relay:$TAG|g" k8s/deployment.yaml | kubectl apply -f - -n cloudbridge

# Применить service
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

## 🚀 **Тестирование**

### **1. Port forwarding для тестирования:**

```bash
# Port forwarding для HTTP API
kubectl port-forward service/cloudbridge-relay-external 8080:8080 -n cloudbridge &

# Port forwarding для P2P API  
kubectl port-forward service/cloudbridge-relay-external 8082:8082 -n cloudbridge &

# Port forwarding для MASQUE
kubectl port-forward service/cloudbridge-relay-external 8443:8443 -n cloudbridge &
```

### **2. Проверка работы:**

```bash
# Тестировать HTTP API
curl http://localhost:8080/health

# Тестировать P2P API
curl http://localhost:8082/health

# Тестировать MASQUE
curl -k https://localhost:8443/.well-known/masque/udp/127.0.0.1:8080
```

### **3. Проверка логов:**

```bash
# Следить за логами в реальном времени
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge
```

## 🔄 **Обновление образа**

```bash
# Создать новый уникальный тег
NEW_TAG="mvp-$(date +%Y%m%d-%H%M%S)"
echo "Новый тег: $NEW_TAG"

# Пересобрать образ
docker build -t cloudbridge-relay:$NEW_TAG .

# Загрузить в Minikube
minikube image load cloudbridge-relay:$NEW_TAG

# Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$NEW_TAG -n cloudbridge

# Проверить обновление
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

### **⚠️ ВАЖНО: Восстановление переменных Sentry после обновления**

После обновления образа переменные Sentry могут сброситься. Восстановите их:

```bash
# Восстановить переменные Sentry из .env файла
source .env

# Применить переменные Sentry
kubectl patch deployment cloudbridge-relay -n cloudbridge -p "{
  \"spec\": {
    \"template\": {
      \"spec\": {
        \"containers\": [{
          \"name\": \"cloudbridge-relay\",
          \"env\": [
            {\"name\": \"SENTRY_DSN\", \"value\": \"$SENTRY_DSN\"},
            {\"name\": \"SENTRY_ENVIRONMENT\", \"value\": \"$SENTRY_ENVIRONMENT\"},
            {\"name\": \"SENTRY_RELEASE\", \"value\": \"$SENTRY_RELEASE\"},
            {\"name\": \"SENTRY_DEBUG\", \"value\": \"$SENTRY_DEBUG\"},
            {\"name\": \"SENTRY_SERVER_NAME\", \"value\": \"$SENTRY_SERVER_NAME\"},
            {\"name\": \"SENTRY_TRACES_SAMPLE_RATE\", \"value\": \"$SENTRY_TRACES_SAMPLE_RATE\"},
            {\"name\": \"SENTRY_SAMPLE_RATE\", \"value\": \"$SENTRY_SAMPLE_RATE\"},
            {\"name\": \"SENTRY_MAX_BREADCRUMBS\", \"value\": \"$SENTRY_MAX_BREADCRUMBS\"}
          ]
        }]
      }
    }
  }
}"

# Проверить, что Sentry работает
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "Integration installed"
```

## 📊 **Настройка Sentry (мониторинг и логирование)**

### **1. Конфигурация Sentry:**

CloudBridge Relay поддерживает интеграцию с Sentry для мониторинга ошибок и метрик.

```bash
# Настроить переменные Sentry в deployment
kubectl patch deployment cloudbridge-relay -n cloudbridge -p '{
  "spec": {
    "template": {
      "spec": {
        "containers": [{
          "name": "cloudbridge-relay",
          "env": [
            {"name": "SENTRY_DSN", "value": "https://YOUR_DSN@sentry.io/YOUR_PROJECT_ID"},
            {"name": "SENTRY_ENVIRONMENT", "value": "production"},
            {"name": "SENTRY_RELEASE", "value": "cloudbridge-relay@1.0.0"},
            {"name": "SENTRY_DEBUG", "value": "true"},
            {"name": "SENTRY_SERVER_NAME", "value": "cloudbridge-relay"},
            {"name": "SENTRY_TRACES_SAMPLE_RATE", "value": "0.1"},
            {"name": "SENTRY_SAMPLE_RATE", "value": "1.0"},
            {"name": "SENTRY_MAX_BREADCRUMBS", "value": "50"}
          ]
        }]
      }
    }
  }
}'
```

### **2. Проверка работы Sentry:**

```bash
# Проверить логи Sentry
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep -i sentry

# Проверить отправку событий в Sentry
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "Sending.*event"

# Проверить метрики Sentry
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "sentry/metrics.go"
```

### **3. Настройка из .env файла:**

Если у вас есть `.env` файл с конфигурацией Sentry:

```bash
# Загрузить переменные из .env файла
source .env

# Применить все переменные Sentry
kubectl patch deployment cloudbridge-relay -n cloudbridge -p "{
  \"spec\": {
    \"template\": {
      \"spec\": {
        \"containers\": [{
          \"name\": \"cloudbridge-relay\",
          \"env\": [
            {\"name\": \"SENTRY_DSN\", \"value\": \"$SENTRY_DSN\"},
            {\"name\": \"SENTRY_ENVIRONMENT\", \"value\": \"$SENTRY_ENVIRONMENT\"},
            {\"name\": \"SENTRY_RELEASE\", \"value\": \"$SENTRY_RELEASE\"},
            {\"name\": \"SENTRY_DEBUG\", \"value\": \"$SENTRY_DEBUG\"},
            {\"name\": \"SENTRY_SERVER_NAME\", \"value\": \"$SENTRY_SERVER_NAME\"},
            {\"name\": \"SENTRY_TRACES_SAMPLE_RATE\", \"value\": \"$SENTRY_TRACES_SAMPLE_RATE\"},
            {\"name\": \"SENTRY_SAMPLE_RATE\", \"value\": \"$SENTRY_SAMPLE_RATE\"},
            {\"name\": \"SENTRY_MAX_BREADCRUMBS\", \"value\": \"$SENTRY_MAX_BREADCRUMBS\"}
          ]
        }]
      }
    }
  }
}"
```

### **4. Проверка в Sentry Dashboard:**

1. Откройте [Sentry Dashboard](https://sentry.io/)
2. Перейдите в ваш проект
3. Проверьте разделы:
   - **Issues** - ошибки и исключения
   - **Performance** - метрики производительности
   - **Releases** - информация о релизах

### **5. Отладка Sentry:**

```bash
# Проверить инициализацию Sentry
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "Integration installed"

# Проверить отправку событий
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "Sending.*event.*to"

# Проверить ошибки Sentry
kubectl logs deployment/cloudbridge-relay -n cloudbridge | grep "Sentry.*error"
```

## 🐛 **Отладка**

```bash
# Проверить статус pods
kubectl get pods -n cloudbridge

# Проверить события
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'

# Проверить логи
kubectl logs deployment/cloudbridge-relay -n cloudbridge --tail=50
```

## 🧹 **Очистка**

```bash
# Удалить namespace
kubectl delete namespace cloudbridge

# Остановить Minikube
minikube stop

# Очистить образы
docker rmi cloudbridge-relay:$TAG
minikube image rm cloudbridge-relay:$TAG
```

## 📋 **Быстрый чек-лист**

### **✅ Основные проверки:**
- [ ] Minikube запущен
- [ ] Уникальный тег образа создан
- [ ] Docker образ собран
- [ ] Образ загружен в Minikube
- [ ] Deployment применен
- [ ] Pods работают
- [ ] Port forwarding настроен
- [ ] API endpoints отвечают

### **✅ Проверки Sentry:**
- [ ] Переменные Sentry настроены
- [ ] Sentry инициализирован (`Integration installed`)
- [ ] События отправляются в Sentry (`Sending.*event.*to`)
- [ ] Метрики собираются (`sentry/metrics.go`)
- [ ] Dashboard Sentry показывает данные

---

**🎯 Результат: CloudBridge Relay развернут в Minikube с QUIC-first + MASQUE архитектурой!**