# CloudBridge Relay - Minikube Staging Deployment Guide

## 📋 Обзор

Это руководство описывает процесс развертывания CloudBridge Relay в staging environment с использованием Minikube для локальной разработки и тестирования.

## 🎯 Цели

- Создать локальную staging среду для тестирования
- Обеспечить изоляцию от production environment
- Упростить процесс разработки и отладки
- Подготовить к production deployment

---

## 🛠️ Предварительные Требования

### Системные Требования
- **OS:** Linux (Ubuntu 22.04+), macOS, Windows
- **RAM:** Минимум 4GB (рекомендуется 8GB+)
- **CPU:** 2+ ядра
- **Disk:** 20GB+ свободного места

### Установленные Инструменты
```bash
# Docker
sudo apt-get update
sudo apt-get install docker.io
sudo systemctl start docker
sudo systemctl enable docker

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```

---

## 🚀 Запуск Minikube

### 1. Запуск Minikube
```bash
# Запуск с оптимальными настройками
minikube start --driver=docker --memory=4096 --cpus=2

# Проверка статуса
minikube status
kubectl get nodes
```

### 2. Настройка Docker Registry
```bash
# Настройка minikube для использования локального Docker registry
eval $(minikube docker-env)

# Проверка подключения
docker ps
```

---

## 🏗️ Развертывание Staging Environment

### 1. Сборка Docker Image
```bash
# Переход в корневую директорию проекта
cd /path/to/cloudbridge-relay-installer

# Сборка образа
docker build -t cloudbridge-relay:staging .

# Загрузка образа в minikube
minikube image load cloudbridge-relay:staging
```

### 2. Применение Kubernetes Manifests
```bash
# Создание namespace
kubectl apply -f k8s/staging/namespace.yaml

# Применение конфигурации
kubectl apply -f k8s/staging/configmap.yaml

# Применение секретов
kubectl apply -f k8s/staging/secrets.yaml

# Развертывание приложения
kubectl apply -f k8s/staging/deployment.yaml

# Применение сервисов
kubectl apply -f k8s/staging/services.yaml
```

### 3. Проверка Развертывания
```bash
# Проверка подов
kubectl get pods -n cloudbridge-staging

# Проверка сервисов
kubectl get services -n cloudbridge-staging

# Проверка логов
kubectl logs -n cloudbridge-staging -l app.kubernetes.io/name=cloudbridge-relay
```

---

## 🔧 Конфигурация

### Staging Environment Features
- **Debug Mode:** Включен
- **Mock Services:** Используются для изоляции
- **Test Data:** Включен для тестирования
- **Log Level:** Debug
- **Resource Limits:** Снижены для локальной разработки

### Порты и Сервисы
- **9090/TCP:** HTTP API
- **9090/UDP:** QUIC Protocol
- **51820/UDP:** WireGuard P2P Mesh
- **8084/TCP:** Metrics

---

## 🧪 Тестирование

### 1. Health Check
```bash
# Port forwarding
kubectl port-forward -n cloudbridge-staging service/cloudbridge-relay-staging-internal 9090:9090 &

# Проверка health endpoint
curl http://localhost:9090/health
```

### 2. Unit Tests
```bash
# Запуск unit тестов
go test ./... -v

# Запуск с покрытием
go test ./... -cover
```

### 3. Integration Tests
```bash
# Запуск integration тестов
go test ./internal/integration/... -v
```

---

## 📊 Мониторинг

### Доступные Endpoints
- **Health:** `http://localhost:9090/health`
- **Metrics:** `http://localhost:8084/metrics`
- **API:** `http://localhost:9090/api`

### Полезные Команды
```bash
# Просмотр логов в реальном времени
kubectl logs -n cloudbridge-staging -f deployment/cloudbridge-relay-staging

# Описание пода
kubectl describe pod -n cloudbridge-staging -l app.kubernetes.io/name=cloudbridge-relay

# Проверка ресурсов
kubectl top pods -n cloudbridge-staging
```

---

## 🔄 Управление

### Перезапуск Deployment
```bash
# Перезапуск deployment
kubectl rollout restart deployment/cloudbridge-relay-staging -n cloudbridge-staging

# Проверка статуса rollout
kubectl rollout status deployment/cloudbridge-relay-staging -n cloudbridge-staging
```

### Масштабирование
```bash
# Увеличение количества реплик
kubectl scale deployment/cloudbridge-relay-staging --replicas=3 -n cloudbridge-staging

# Проверка статуса
kubectl get pods -n cloudbridge-staging
```

### Обновление Конфигурации
```bash
# Редактирование ConfigMap
kubectl edit configmap cloudbridge-staging-config -n cloudbridge-staging

# Применение изменений
kubectl apply -f k8s/staging/configmap.yaml
kubectl rollout restart deployment/cloudbridge-relay-staging -n cloudbridge-staging
```

---

## 🛑 Остановка и Очистка

### Остановка Minikube
```bash
# Остановка minikube
minikube stop

# Проверка статуса
minikube status
```

### Очистка Ресурсов
```bash
# Удаление staging namespace
kubectl delete namespace cloudbridge-staging

# Очистка Docker образов
docker rmi cloudbridge-relay:staging
```

### Полная Очистка Minikube
```bash
# Удаление minikube кластера
minikube delete

# Очистка Docker
docker system prune -a
```

---

## 🚨 Troubleshooting

### Частые Проблемы

#### 1. Pod в состоянии ImagePullBackOff
```bash
# Проверка образа
minikube image ls | grep cloudbridge-relay

# Перезагрузка образа
minikube image load cloudbridge-relay:staging
```

#### 2. Port уже используется
```bash
# Поиск процесса
sudo netstat -tlnp | grep :9090

# Завершение процесса
sudo kill -9 <PID>
```

#### 3. Недостаточно ресурсов
```bash
# Увеличение памяти minikube
minikube stop
minikube start --driver=docker --memory=6144 --cpus=3
```

#### 4. Проблемы с Docker
```bash
# Перезапуск Docker
sudo systemctl restart docker

# Проверка статуса
sudo systemctl status docker
```

---

## 📈 Производительность

### Оптимизация для Staging
- **Memory:** 4GB для minikube
- **CPU:** 2 ядра
- **Storage:** 20GB
- **Network:** Bridge mode

### Мониторинг Ресурсов
```bash
# Использование ресурсов
kubectl top nodes
kubectl top pods -n cloudbridge-staging

# Детальная информация
kubectl describe node minikube
```

---

## 🔗 Полезные Ссылки

- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [CloudBridge Relay Architecture](../architecture/PROJECT_OVERVIEW.md)

---

## 📝 Заключение

Minikube staging environment предоставляет:
- ✅ Изолированную среду для тестирования
- ✅ Быстрое развертывание и отладку
- ✅ Полную совместимость с Kubernetes
- ✅ Простое управление ресурсами

**Рекомендация:** Используйте staging environment для всех разработок и тестирования перед production deployment.

---

*Документация обновлена: 6 сентября 2025*

