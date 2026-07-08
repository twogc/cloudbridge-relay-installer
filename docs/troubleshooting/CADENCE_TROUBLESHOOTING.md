# 🚨 Cadence Troubleshooting - Быстрые решения

## Диагностика проблем

### 1. Быстрая проверка статуса
```bash
# Проверить все поды
kubectl get pods -n cloudbridge

# Проверить события
kubectl get events -n cloudbridge --sort-by='.lastTimestamp' | tail -10

# Проверить ресурсы
kubectl top pods -n cloudbridge
```

### 2. Проверка логов
```bash
# Логи Cadence
kubectl logs -f deployment/cadence -n cloudbridge

# Логи базы данных
kubectl logs -f deployment/postgres -n cloudbridge

# Логи с предыдущего контейнера
kubectl logs -f deployment/cadence -n cloudbridge --previous
```

## 🐛 Частые ошибки и быстрые решения

###  CrashLoopBackOff

#### Проблема: Cadence не запускается
```bash
# Диагностика
kubectl describe pod -l app.kubernetes.io/name=cadence -n cloudbridge

# Быстрое решение - перезапуск
kubectl rollout restart deployment/cadence -n cloudbridge

# Если не помогает - удалить и пересоздать
kubectl delete deployment cadence -n cloudbridge
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
```

#### Проблема: База данных не запускается
```bash
# Диагностика
kubectl logs -f deployment/postgres -n cloudbridge

# Быстрое решение - пересоздать PVC
kubectl delete pvc postgres-pvc -n cloudbridge
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge
```

###  ImagePullBackOff

#### Проблема: Не удается загрузить образ
```bash
# Проверить доступность образа
docker pull ubercadence/server:v0.22.4

# Исправить политику загрузки
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","imagePullPolicy":"IfNotPresent"}]}}}}'
```

###  Config validation failed

#### Проблема: Must provide one of VisibilityStore and AdvancedVisibilityStore
```bash
# Быстрое решение
kubectl patch deployment cadence -n cloudbridge -p '{"spec":{"template":{"spec":{"containers":[{"name":"cadence","env":[{"name":"VISIBILITY_STORE","value":"sqlite"},{"name":"VISIBILITY_DB_PATH","value":"/var/lib/cadence/visibility.db"}]}]}}}}'
```

###  Connection refused

#### Проблема: База данных недоступна
```bash
# Проверить подключение к PostgreSQL
kubectl exec -it deployment/cadence -n cloudbridge -- nslookup postgres-service

# Переключиться на SQLite
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
```

###  Port already in use

#### Проблема: Порт занят при port forwarding
```bash
# Найти процесс, использующий порт
lsof -i :8080

# Убить процесс
kill -9 <PID>

# Или использовать другой порт
kubectl port-forward service/cadence-nodeport 8081:7934 -n cloudbridge
```

##  Быстрые исправления

### 1. Сброс конфигурации Cadence
```bash
# Удалить deployment
kubectl delete deployment cadence -n cloudbridge

# Применить рабочую конфигурацию
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge -w
```

### 2. Сброс базы данных
```bash
# Удалить PostgreSQL
kubectl delete deployment postgres -n cloudbridge
kubectl delete pvc postgres-pvc -n cloudbridge

# Пересоздать
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge
```

### 3. Полный сброс
```bash
# Удалить namespace
kubectl delete namespace cloudbridge

# Пересоздать
kubectl create namespace cloudbridge
kubectl apply -f k8s/deployments/postgres-simple.yaml -n cloudbridge
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
kubectl apply -f k8s/services/cadence-nodeport.yaml -n cloudbridge
```

## Быстрый старт после проблем

### 1. Минимальная рабочая конфигурация
```bash
# Создать namespace
kubectl create namespace cloudbridge

# Развернуть Cadence с SQLite
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge

# Создать сервис
kubectl apply -f k8s/services/cadence-nodeport.yaml -n cloudbridge

# Проверить
kubectl get pods -n cloudbridge
```

### 2. Проверка работы
```bash
# Port forwarding
kubectl port-forward service/cadence-nodeport 7934:7934 -n cloudbridge &

# Тест
curl http://localhost:7934/health
```

## Мониторинг

### 1. Основные метрики
```bash
# Статус подов
kubectl get pods -n cloudbridge -o wide

# Использование ресурсов
kubectl top pods -n cloudbridge

# События
kubectl get events -n cloudbridge --sort-by='.lastTimestamp' | tail -20
```

### 2. Логи в реальном времени
```bash
# Все логи Cadence
kubectl logs -f deployment/cadence -n cloudbridge

# Логи с фильтром
kubectl logs -f deployment/cadence -n cloudbridge | grep -i error

# Логи базы данных
kubectl logs -f deployment/postgres -n cloudbridge
```

## Восстановление

### 1. Восстановление из бэкапа
```bash
# Если есть бэкап PVC
kubectl delete pvc cadence-pvc -n cloudbridge
kubectl apply -f backup/cadence-pvc.yaml -n cloudbridge

# Перезапустить Cadence
kubectl rollout restart deployment/cadence -n cloudbridge
```

### 2. Восстановление конфигурации
```bash
# Восстановить из git
git checkout HEAD -- k8s/deployments/cadence-working.yaml

# Применить
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
```

## 🧹 Очистка

### 1. Очистка ресурсов
```bash
# Удалить все поды
kubectl delete pods --all -n cloudbridge

# Удалить все deployments
kubectl delete deployments --all -n cloudbridge

# Удалить все PVC
kubectl delete pvc --all -n cloudbridge
```

### 2. Полная очистка
```bash
# Удалить namespace
kubectl delete namespace cloudbridge

# Очистить Docker образы
docker rmi ubercadence/server:v0.22.4
docker rmi postgres:13
```

## 📞 Экстренные команды

### 1. Когда ничего не работает
```bash
# Остановить Minikube
minikube stop

# Очистить Minikube
minikube delete

# Запустить заново
minikube start --driver=docker --memory=4096 --cpus=2

# Развернуть заново
kubectl create namespace cloudbridge
kubectl apply -f k8s/deployments/cadence-working.yaml -n cloudbridge
kubectl apply -f k8s/services/cadence-nodeport.yaml -n cloudbridge
```

### 2. Когда нужен быстрый доступ
```bash
# Port forwarding для всех сервисов
kubectl port-forward service/cadence-nodeport 7933:7933 -n cloudbridge &
kubectl port-forward service/cadence-nodeport 7934:7934 -n cloudbridge &

# Проверить доступность
curl http://localhost:7934/health
```

## Чек-лист для диагностики

- [ ] Minikube запущен: `minikube status`
- [ ] Namespace создан: `kubectl get namespaces | grep cloudbridge`
- [ ] Поды запущены: `kubectl get pods -n cloudbridge`
- [ ] Сервисы созданы: `kubectl get services -n cloudbridge`
- [ ] Логи без ошибок: `kubectl logs deployment/cadence -n cloudbridge`
- [ ] API доступен: `curl http://localhost:7934/health`
- [ ] Ресурсы достаточны: `kubectl top pods -n cloudbridge`

---

**💡 Совет: Сохраните этот файл в закладки для быстрого доступа к решениям!**
