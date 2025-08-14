# 🐳 Установка Docker Registry в VK Cloud

## 📋 Предварительные требования

### 1. Плавающий IP адрес
Вам нужно создать плавающий IP адрес в VK Cloud:

1. **VK Cloud Console** → **Сети** → **Плавающие IP**
2. **Создать плавающий IP**
3. **Скопировать IP адрес** (например: `95.163.250.128`)

### 2. Объектное хранилище S3
Учетные данные уже настроены:
- **Access Key**: `82d3zPrQ5X2WW8WDwWC8Wm`
- **Secret Key**: `6cjq48bKiUC52LFP62G3YF5e39BfabELZ8eGeCUDwKQc`
- **Bucket**: `dockerimages-8530029fa7b64199b2ad87c18d4620df-bucket`

## 🚀 Установка Docker Registry

### Шаг 1: Обновить конфигурацию
Откройте файл `configs/docker-registry-vkcloud.yaml` и замените:

```yaml
service:
  name: registry
  type: LoadBalancer
  loadBalancerIP: "95.163.250.128"  # ← Вставьте ваш плавающий IP
  port: 5000
  nodePort: 5000
```

### Шаг 2: Установить через VK Cloud Console

1. **VK Cloud Console** → **Kubernetes** → **kubernetes-cluster-8926** → **Аддоны**
2. **Найти** "docker-registry" → **Установить**
3. **Название приложения**: `cloudbridge-registry`
4. **Пространство имен**: `kube-system`
5. **Values**: Вставить содержимое `configs/docker-registry-vkcloud.yaml`
6. **Установить аддон**

### Шаг 3: Проверка установки

```bash
# Проверить поды
export KUBECONFIG=~/.kube/kubernetes-cluster-8926_kubeconfig.yaml
kubectl --insecure-skip-tls-verify get pods -n kube-system | grep registry

# Проверить сервис
kubectl --insecure-skip-tls-verify get services -n kube-system | grep registry
```

## 🔐 Настройка доступа к Registry

### Шаг 1: Войти в registry
```bash
# Замените <плавающий_IP> на ваш IP адрес
docker login <плавающий_IP>:5000 --username cloudbridge --password password
```

### Шаг 2: Загрузить образы
```bash
# Тегировать образы
docker tag localhost/cloudbridge-auth-service:latest <плавающий_IP>:5000/cloudbridge/auth-service:latest
docker tag localhost/cloudbridge-metrics-service:latest <плавающий_IP>:5000/cloudbridge/metrics-service:latest

# Загрузить образы
docker push <плавающий_IP>:5000/cloudbridge/auth-service:latest
docker push <плавающий_IP>:5000/cloudbridge/metrics-service:latest
```

### Шаг 3: Создать секрет для Kubernetes
```bash
# Создать секрет для доступа к registry
kubectl --insecure-skip-tls-verify create secret docker-registry cloudbridge-registry-secret \
  --docker-server=<плавающий_IP>:5000 \
  --docker-username=cloudbridge \
  --docker-password=password \
  --docker-email=info@2gc.ru \
  -n cloudbridge
```

## 🔧 Обновление deployment файлов

После установки registry обновите deployment файлы:

### Auth Service
```yaml
# k8s/auth-service-real-deployment.yaml
spec:
  template:
    spec:
      imagePullSecrets:
      - name: cloudbridge-registry-secret
      containers:
      - name: auth-service
        image: <плавающий_IP>:5000/cloudbridge/auth-service:latest
        imagePullPolicy: Always
```

### Metrics Service
```yaml
# k8s/metrics-service-real-deployment.yaml
spec:
  template:
    spec:
      imagePullSecrets:
      - name: cloudbridge-registry-secret
      containers:
      - name: metrics-service
        image: <плавающий_IP>:5000/cloudbridge/metrics-service:latest
        imagePullPolicy: Always
```

## 📊 Проверка работы

### 1. Проверить registry
```bash
# Список образов
curl -u cloudbridge:password http://<плавающий_IP>:5000/v2/_catalog

# Список тегов для образа
curl -u cloudbridge:password http://<плавающий_IP>:5000/v2/cloudbridge/auth-service/tags/list
```

### 2. Развернуть микросервисы
```bash
# Применить обновленные deployment
kubectl --insecure-skip-tls-verify apply -f k8s/auth-service-real-deployment.yaml
kubectl --insecure-skip-tls-verify apply -f k8s/metrics-service-real-deployment.yaml

# Проверить статус
kubectl --insecure-skip-tls-verify get pods -n cloudbridge
```

## 🚨 Устранение проблем

### Ошибка "неверные данные в поле IP address"
- Убедитесь, что плавающий IP создан в VK Cloud
- Проверьте, что IP адрес указан в правильном формате
- Убедитесь, что IP адрес не используется другими сервисами

### Ошибка доступа к registry
- Проверьте, что registry запущен: `kubectl get pods -n kube-system | grep registry`
- Проверьте внешний IP: `kubectl get services -n kube-system | grep registry`
- Убедитесь, что порт 5000 открыт

### Ошибка загрузки образов
- Проверьте логи registry: `kubectl logs -n kube-system <registry-pod-name>`
- Убедитесь, что S3 bucket существует и доступен
- Проверьте учетные данные S3

## 📚 Дополнительные ресурсы

- [Документация VK Cloud Registry](https://mcs.mail.ru/app/docs/kubernetes/addons/docker-registry)
- [Docker Registry Helm Chart](https://github.com/twuni/docker-registry.helm)
- [VK Cloud Storage S3](https://mcs.mail.ru/app/docs/object-storage)

---

**⏱️ Время установки**: 15-30 минут  
**🎯 Результат**: Приватный Docker Registry для CloudBridge микросервисов
