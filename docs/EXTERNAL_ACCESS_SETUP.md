# 🌐 Настройка внешнего доступа к микросервисам CloudBridge Relay

## 📊 Текущая ситуация

### Проблема с LoadBalancer
VK Cloud не может создать LoadBalancer из-за превышения квоты на floating IP адреса:
```
Error: Quota exceeded for resource: 'floatingip' (limit: 4, used: 4)
```

### Решение: NodePort сервисы
Вместо LoadBalancer используем NodePort сервисы для внешнего доступа.

## 🔧 Текущая конфигурация

### NodePort сервисы
```bash
# Auth Service
auth-service-external: NodePort 10.254.241.6
- HTTP: 8080:30080/TCP
- Metrics: 9090:30090/TCP

# Metrics Service  
metrics-service-external: NodePort 10.254.204.64
- HTTP: 8084:30084/TCP
- Metrics: 9090:30091/TCP
```

### Kubernetes узлы
```bash
# Master node
kubernetes-cluster-8926-master-0: 10.0.0.137

# Worker node
kubernetes-cluster-8926-default-group-0: 10.0.0.130
```

## 🌐 Варианты внешнего доступа

### Вариант 1: Прямой доступ через NodePort (Рекомендуется)

#### Для Auth Service:
```bash
# HTTP API
curl http://10.0.0.137:30080/health
curl http://10.0.0.130:30080/health

# Metrics
curl http://10.0.0.137:30090/metrics
curl http://10.0.0.130:30090/metrics
```

#### Для Metrics Service:
```bash
# HTTP API
curl http://10.0.0.137:30084/health
curl http://10.0.0.130:30084/health

# Metrics
curl http://10.0.0.137:30091/metrics
curl http://10.0.0.130:30091/metrics
```

### Вариант 2: Через Ingress Controller

#### Настройка DNS (локально для тестирования):
```bash
# Добавить в /etc/hosts
10.0.0.137 auth.cloudbridge.local
10.0.0.137 metrics.cloudbridge.local
10.0.0.137 api.cloudbridge.local
```

#### Доступ через Ingress:
```bash
# Auth Service
curl http://auth.cloudbridge.local/health
curl http://api.cloudbridge.local/auth/health

# Metrics Service
curl http://metrics.cloudbridge.local/health
curl http://api.cloudbridge.local/metrics/health
```

## 🚀 Настройка внешнего IP (Опционально)

### Для production использования рекомендуется:

#### 1. Запросить увеличение квоты в VK Cloud
```bash
# Обратиться в поддержку VK Cloud для увеличения лимита floating IP
# Текущий лимит: 4, требуется: минимум 1 дополнительный
```

#### 2. Назначить внешний IP узлу
```bash
# В VK Cloud Console:
# 1. Перейти к Compute -> Instances
# 2. Найти kubernetes-cluster-8926-master-0
# 3. Назначить Floating IP
```

#### 3. Обновить LoadBalancer
```bash
# После получения внешнего IP
kubectl apply -f k8s/cloudbridge-ingress.yaml
```

## 🔍 Тестирование доступа

### Проверка доступности сервисов:
```bash
# Проверка Auth Service
curl -v http://10.0.0.137:30080/health
curl -v http://10.0.0.137:30080/ready

# Проверка Metrics Service
curl -v http://10.0.0.137:30084/health
curl -v http://10.0.0.137:30084/ready

# Проверка метрик
curl -v http://10.0.0.137:30090/metrics
curl -v http://10.0.0.137:30091/metrics
```

### Проверка через Ingress:
```bash
# Добавить в /etc/hosts
echo "10.0.0.137 auth.cloudbridge.local metrics.cloudbridge.local api.cloudbridge.local" | sudo tee -a /etc/hosts

# Тестирование
curl -v http://auth.cloudbridge.local/health
curl -v http://metrics.cloudbridge.local/health
curl -v http://api.cloudbridge.local/auth/health
```

## 🔒 Безопасность

### Рекомендации:
1. **Firewall**: Настроить правила для доступа только к нужным портам
2. **SSL/TLS**: Настроить SSL сертификаты для production
3. **Authentication**: Добавить аутентификацию для внешнего доступа
4. **Rate Limiting**: Настроить ограничения на количество запросов

### Настройка firewall:
```bash
# Разрешить доступ только к NodePort портам
sudo ufw allow 30080/tcp  # Auth Service HTTP
sudo ufw allow 30084/tcp  # Metrics Service HTTP
sudo ufw allow 30090/tcp  # Auth Service Metrics
sudo ufw allow 30091/tcp  # Metrics Service Metrics
```

## 📊 Мониторинг доступа

### Логирование запросов:
```bash
# Просмотр логов Ingress Controller
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller

# Просмотр логов сервисов
kubectl logs -n cloudbridge deployment/auth-service
kubectl logs -n cloudbridge deployment/metrics-service
```

### Метрики доступа:
```bash
# Метрики Ingress Controller
curl http://10.0.0.137:30090/metrics | grep nginx

# Метрики сервисов
curl http://10.0.0.137:30090/metrics
curl http://10.0.0.137:30091/metrics
```

## 🎯 Следующие шаги

### Для разработчика:
1. **Протестировать доступ** к сервисам через NodePort
2. **Настроить DNS** для удобного доступа
3. **Реализовать реальные сервисы** с health checks
4. **Настроить SSL** сертификаты

### Для production:
1. **Запросить увеличение квоты** floating IP в VK Cloud
2. **Настроить LoadBalancer** с внешним IP
3. **Настроить SSL** сертификаты
4. **Настроить мониторинг** и алерты

---

**🌐 Статус**: Внешний доступ настроен через NodePort  
**🔧 Готово к**: Тестированию и разработке  
**🚀 Следующий этап**: Реализация реальных сервисов

