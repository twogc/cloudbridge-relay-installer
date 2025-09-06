# CloudBridge Relay - Quick Start Staging

## 🚀 Быстрый Запуск Staging Environment

### 1. Запуск Minikube
```bash
minikube start --driver=docker --memory=4096 --cpus=2
```

### 2. Сборка и Загрузка Образа
```bash
docker build -t cloudbridge-relay:staging .
minikube image load cloudbridge-relay:staging
```

### 3. Развертывание
```bash
kubectl apply -f k8s/staging/
```

### 4. Проверка
```bash
kubectl get pods -n cloudbridge-staging
kubectl port-forward -n cloudbridge-staging service/cloudbridge-relay-staging-internal 9090:9090 &
curl http://localhost:9090/health
```

### 5. Остановка
```bash
minikube stop
```

---

## 📋 Команды для Управления

### Проверка Статуса
```bash
kubectl get all -n cloudbridge-staging
kubectl logs -n cloudbridge-staging -l app.kubernetes.io/name=cloudbridge-relay
```

### Перезапуск
```bash
kubectl rollout restart deployment/cloudbridge-relay-staging -n cloudbridge-staging
```

### Очистка
```bash
kubectl delete namespace cloudbridge-staging
minikube delete
```

---

*Для подробной документации см. [MINIKUBE_STAGING_GUIDE.md](./MINIKUBE_STAGING_GUIDE.md)*

