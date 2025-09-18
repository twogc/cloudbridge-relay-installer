# 🚀 Руководство по развёртыванию в Minikube

## 📋 Содержание
1. [Быстрый старт](#быстрый-старт)
2. [Сборка образов](#сборка-образов)
3. [Развёртывание сервисов](#развёртывание-сервисов)
4. [Типичные ошибки и решения](#типичные-ошибки-и-решения)
5. [Полезные команды](#полезные-команды)

## 🚀 Быстрый старт

### 1. Проверка Minikube
```bash
# Проверить статус Minikube
minikube status

# Запустить Minikube (если не запущен)
minikube start

# Проверить доступные образы
minikube image ls
```

### 2. Сборка и загрузка образа
```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать образ с уникальным тегом
docker build -t cloudbridge-relay:$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD) .

# Загрузить образ в Minikube
minikube image load cloudbridge-relay:$(date +%Y%m%d-%H%M%S)-$(git rev-parse --short HEAD)
```

### 3. Развёртывание
```bash
# Применить deployment
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment.yaml

# Проверить статус
kubectl get pods -n cloudbridge
```

## 🔨 Сборка образов

### Метод 1: Локальная сборка + загрузка
```bash
# Собрать образ локально
docker build -t cloudbridge-relay:latest .

# Загрузить в Minikube
minikube image load cloudbridge-relay:latest
```

### Метод 2: Прямая сборка в Minikube
```bash
# Собрать образ напрямую в Minikube
minikube image build -t cloudbridge-relay:latest .
```

### Метод 3: Сборка с Dockerfile.p2p
```bash
# Собрать P2P образ
docker build -t cloudbridge-relay-p2p:latest -f Dockerfile.p2p .

# Загрузить в Minikube
minikube image load cloudbridge-relay-p2p:latest
```

## 🚀 Развёртывание сервисов

### 1. Обновление существующего deployment
```bash
# Обновить образ в deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:latest -n cloudbridge

# Проверить статус rollout
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

### 2. Создание нового deployment
```bash
# Применить новый deployment
kubectl apply -f k8s/deployments/cloudbridge-relay-p2p-deployment.yaml

# Проверить статус
kubectl get pods -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay-p2p
```

### 3. Удаление и пересоздание
```bash
# Удалить deployment
kubectl delete deployment cloudbridge-relay -n cloudbridge

# Применить новый
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment.yaml
```

## ❌ Типичные ошибки и решения

### 1. Ошибка: `exec ./cloudbridge-relay: no such file or directory`

**Причина**: Бинарный файл собран для glibc, а контейнер использует musl libc (Alpine)

**Решение**:
```bash
# Проверить зависимости
docker run --rm cloudbridge-relay:latest ldd /app/cloudbridge-relay

# Если видите ошибки типа "symbol not found", измените базовый образ
# В Dockerfile замените:
FROM alpine:3.18
# На:
FROM ubuntu:22.04
```

### 2. Ошибка: `CrashLoopBackOff`

**Причина**: Контейнер постоянно падает

**Диагностика**:
```bash
# Посмотреть логи
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20

# Описать под для детальной информации
kubectl describe pod -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay
```

**Типичные причины**:
- Отсутствуют переменные окружения
- Неправильные права доступа к файлам
- Отсутствуют конфигурационные файлы

### 3. Ошибка: `ImagePullBackOff`

**Причина**: Minikube не может найти образ

**Решение**:
```bash
# Проверить, что образ загружен
minikube image ls | grep cloudbridge-relay

# Если образа нет, загрузить
minikube image load cloudbridge-relay:latest
```

### 4. Ошибка: `Invalid value: "": may not be specified when value is not empty`

**Причина**: В deployment есть переменные окружения с конфликтующими значениями

**Решение**:
```yaml
# Неправильно:
env:
- name: JWT_SECRET
  value: "test-secret"
  valueFrom:
    secretKeyRef:
      name: cloudbridge-secrets
      key: jwt-secret

# Правильно (выберите один способ):
env:
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: cloudbridge-secrets
      key: jwt-secret
```

### 5. Ошибка: `Service has no endpoints`

**Причина**: Labels в deployment не совпадают с selector в service

**Решение**:
```bash
# Проверить labels deployment
kubectl get deployment cloudbridge-relay -n cloudbridge -o yaml | grep -A 10 labels

# Проверить selector service
kubectl get service cloudbridge-relay -n cloudbridge -o yaml | grep -A 5 selector

# Исправить labels в deployment файле
```

### 6. Ошибка: `readOnlyRootFilesystem: true`

**Причина**: Контейнер пытается писать в read-only файловую систему

**Решение**:
```yaml
# В deployment изменить:
securityContext:
  readOnlyRootFilesystem: false  # Изменить на false
```

### 7. Ошибка: `Failed to load configuration: JWT secret must be set`

**Причина**: Отсутствует переменная окружения `RELAY_JWT_SECRET`

**Решение**:
```yaml
# Добавить в deployment:
env:
- name: RELAY_JWT_SECRET
  value: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
```

## 🧹 Очистка и обслуживание

### Очистка старых образов
```bash
# Удалить неиспользуемые образы в Minikube
minikube ssh "docker image prune -f"

# Удалить конкретные образы
minikube image rm cloudbridge-relay:old-tag
```

### Очистка deployment
```bash
# Удалить deployment
kubectl delete deployment cloudbridge-relay -n cloudbridge

# Удалить все ресурсы
kubectl delete all -l app.kubernetes.io/name=cloudbridge-relay -n cloudbridge
```

## 🔍 Полезные команды

### Мониторинг
```bash
# Посмотреть все поды
kubectl get pods -n cloudbridge

# Посмотреть логи
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 -f

# Посмотреть события
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'

# Описать ресурс
kubectl describe deployment cloudbridge-relay -n cloudbridge
```

### Отладка
```bash
# Запустить debug под
kubectl run debug-pod --rm -i --tty --image=cloudbridge-relay:latest --restart=Never -- /bin/sh

# Проверить конфигурацию
kubectl get configmap cloudbridge-config -n cloudbridge -o yaml

# Проверить секреты
kubectl get secret cloudbridge-secrets -n cloudbridge -o yaml
```

### Сетевые проверки
```bash
# Проверить сервисы
kubectl get services -n cloudbridge

# Проверить endpoints
kubectl get endpoints -n cloudbridge

# Port forward для тестирования
kubectl port-forward -n cloudbridge deployment/cloudbridge-relay 9090:9090
```

## 📝 Шаблоны deployment

### Минимальный deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cloudbridge-relay
  namespace: cloudbridge
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cloudbridge-relay
  template:
    metadata:
      labels:
        app: cloudbridge-relay
    spec:
      containers:
      - name: cloudbridge-relay
        image: cloudbridge-relay:latest
        ports:
        - containerPort: 9090
        env:
        - name: JWT_SECRET
          value: "your-jwt-secret"
        - name: RELAY_JWT_SECRET
          value: "your-relay-jwt-secret"
        securityContext:
          readOnlyRootFilesystem: false
          runAsNonRoot: true
          runAsUser: 1001
          runAsGroup: 1001
```

## 🎯 Лучшие практики

1. **Всегда используйте уникальные теги** для образов
2. **Проверяйте логи** при возникновении проблем
3. **Используйте readOnlyRootFilesystem: false** для отладки
4. **Проверяйте совместимость** базового образа с бинарным файлом
5. **Очищайте старые образы** для экономии места
6. **Используйте health checks** для мониторинга
7. **Проверяйте переменные окружения** перед развёртыванием

## 🆘 Экстренное восстановление

Если всё сломалось:

```bash
# 1. Остановить все deployment
kubectl delete deployment --all -n cloudbridge

# 2. Очистить образы
minikube ssh "docker image prune -a -f"

# 3. Перезапустить Minikube
minikube stop
minikube start

# 4. Пересобрать и развернуть
docker build -t cloudbridge-relay:latest .
minikube image load cloudbridge-relay:latest
kubectl apply -f k8s/deployments/cloudbridge-relay-deployment.yaml
```

---

**Примечание**: Этот гайд основан на реальном опыте развёртывания cloudbridge-relay в Minikube и содержит решения для всех типичных проблем, с которыми можно столкнуться.
