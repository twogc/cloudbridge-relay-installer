# 🚀 Быстрый старт миграции CloudBridge Relay в Kubernetes

## 📋 Предварительные требования

### ✅ Инфраструктура
- **Kubernetes кластер**: `cloudbridge-cluster` в VK Cloud (уже развернут)
- **API сервер**: https://91.219.226.200:6443
- **Prometheus Stack**: Установлен и настроен
- **Docker Registry**: Готов к использованию

### ✅ Зависимости
```bash
# Проверка зависимостей
go version          # Go 1.23+
docker --version    # Docker 20.10+
kubectl version     # kubectl 1.28+
helm version        # Helm 3.12+
```

## 🎯 Быстрая миграция (30 минут)

### Шаг 1: Запуск автоматической миграции
```bash
cd cloudbridge-relay-installer
./scripts/migrate-to-kubernetes.sh
```

Скрипт автоматически выполнит:
- ✅ Анализ текущего кода
- ✅ Создание микросервисов
- ✅ Сборку Docker образов
- ✅ Создание Kubernetes манифестов
- ✅ Развертывание в кластере
- ✅ Проверку работоспособности

### Шаг 2: Проверка развертывания
```bash
# Проверка статуса
kubectl get all -n cloudbridge-relay

# Проверка логов
kubectl logs -f deployment/wireguard-service -n cloudbridge-relay
kubectl logs -f deployment/quic-service -n cloudbridge-relay
kubectl logs -f deployment/cadence-service -n cloudbridge-relay
kubectl logs -f deployment/ai-service -n cloudbridge-relay
kubectl logs -f deployment/quantum-service -n cloudbridge-relay
```

### Шаг 3: Тестирование API
```bash
# Health checks
curl http://localhost:8080/health  # WireGuard Service
curl http://localhost:8081/health  # QUIC Service
curl http://localhost:8082/health  # Cadence Service
curl http://localhost:8083/health  # AI Service
curl http://localhost:8084/health  # Quantum Service

# API endpoints
curl http://localhost:8080/tunnels     # WireGuard tunnels
curl http://localhost:8081/connections # QUIC connections
curl http://localhost:8082/workflows   # Cadence workflows
curl http://localhost:8083/anomalies   # AI anomalies
curl http://localhost:8084/keys        # Quantum keys
```

## 🔧 Ручная миграция (если автоматическая не работает)

### 1. Создание микросервисов
```bash
# Создание структуры
mkdir -p cmd/{wireguard-service,quic-service,cadence-service}
mkdir -p k8s/{wireguard-service,quic-service,cadence-service}

# Создание main.go для каждого сервиса
# (см. полный план миграции для деталей)
```

### 2. Сборка образов
```bash
# WireGuard Service
docker build -f Dockerfile.wireguard -t cloudbridge/wireguard-service:v1.0.0 .

# QUIC Service
docker build -f Dockerfile.quic -t cloudbridge/quic-service:v1.0.0 .

# Cadence Service
docker build -f Dockerfile.cadence -t cloudbridge/cadence-service:v1.0.0 .
```

### 3. Развертывание в Kubernetes
```bash
# Создание namespace
kubectl apply -f k8s/namespace.yaml

# Развертывание сервисов
kubectl apply -f k8s/wireguard-service/deployment.yaml
kubectl apply -f k8s/quic-service/deployment.yaml
kubectl apply -f k8s/cadence-service/deployment.yaml

# Настройка Ingress
kubectl apply -f k8s/ingress.yaml
```

## 📊 Мониторинг

### Доступ к Grafana
```bash
# Port-forward к Grafana
kubectl port-forward svc/kube-prometheus-stack-grafana 3000:80 -n prometheus-monitoring

# Откройте http://localhost:3000
# Логин: admin
# Пароль: CloudBridge2024!
```

### Полезные метрики
```promql
# CPU использование
rate(container_cpu_usage_seconds_total{container=~"wireguard-service|quic-service|cadence-service"}[5m])

# Memory использование
container_memory_usage_bytes{container=~"wireguard-service|quic-service|cadence-service"}

# Количество подов
kube_deployment_status_replicas{deployment=~"wireguard-service|quic-service|cadence-service"}
```

## 🚨 Устранение неполадок

### Проблема: Поды не запускаются
```bash
# Проверка событий
kubectl get events -n cloudbridge-relay --sort-by='.lastTimestamp'

# Проверка логов
kubectl describe pod <pod-name> -n cloudbridge-relay
kubectl logs <pod-name> -n cloudbridge-relay
```

### Проблема: Health checks не проходят
```bash
# Проверка readiness/liveness probes
kubectl describe deployment wireguard-service -n cloudbridge-relay

# Ручная проверка health endpoint
kubectl port-forward pod/<pod-name> 8080:8080 -n cloudbridge-relay
curl http://localhost:8080/health
```

### Проблема: Недостаточно ресурсов
```bash
# Проверка ресурсов узлов
kubectl top nodes
kubectl top pods -n cloudbridge-relay

# Масштабирование
kubectl scale deployment wireguard-service --replicas=1 -n cloudbridge-relay
```

## 🔄 Масштабирование

### Автоматическое масштабирование
```bash
# Создание HPA
kubectl apply -f k8s/hpa.yaml

# Проверка HPA
kubectl get hpa -n cloudbridge-relay
```

### Ручное масштабирование
```bash
# Масштабирование сервисов
kubectl scale deployment wireguard-service --replicas=3 -n cloudbridge-relay
kubectl scale deployment quic-service --replicas=5 -n cloudbridge-relay
kubectl scale deployment cadence-service --replicas=2 -n cloudbridge-relay
kubectl scale deployment ai-service --replicas=2 -n cloudbridge-relay
kubectl scale deployment quantum-service --replicas=2 -n cloudbridge-relay
```

## 🎯 Следующие шаги

### 1. Настройка DNS
```bash
# Добавьте запись в /etc/hosts для тестирования
echo "127.0.0.1 api.cloudbridge.local" | sudo tee -a /etc/hosts
```

### 2. Настройка SSL сертификатов
```bash
# Применение Cert Manager
kubectl apply -f k8s/cert-manager.yaml

# Создание ClusterIssuer
kubectl apply -f k8s/cluster-issuer.yaml
```

### 3. Настройка алертов
```bash
# Применение PrometheusRules
kubectl apply -f k8s/monitoring/prometheusrules.yaml
```

### 4. Настройка Service Mesh (опционально)
```bash
# Установка Istio
kubectl apply -f k8s/istio/install.yaml

# Применение VirtualService
kubectl apply -f k8s/istio/virtual-service.yaml
```

## 📋 Чек-лист завершения

- [ ] ✅ Микросервисы развернуты и работают
- [ ] ✅ Health checks проходят
- [ ] ✅ API endpoints отвечают
- [ ] ✅ Мониторинг настроен
- [ ] ✅ Логи доступны
- [ ] ✅ Масштабирование работает
- [ ] ✅ SSL сертификаты настроены
- [ ] ✅ Алерты настроены

## 🔧 Полезные команды

```bash
# Полная информация о развертывании
kubectl get all -n cloudbridge-relay -o wide

# Логи всех подов
kubectl logs -l app=wireguard-service -n cloudbridge-relay
kubectl logs -l app=quic-service -n cloudbridge-relay
kubectl logs -l app=cadence-service -n cloudbridge-relay
kubectl logs -l app=ai-service -n cloudbridge-relay
kubectl logs -l app=quantum-service -n cloudbridge-relay

# Ресурсы
kubectl top pods -n cloudbridge-relay
kubectl describe nodes

# События
kubectl get events -n cloudbridge-relay --sort-by='.lastTimestamp'

# Удаление (если нужно)
kubectl delete namespace cloudbridge-relay
```

---

**⏱️ Время выполнения**: 30 минут  
**🎯 Статус**: ГОТОВ К ЗАПУСКУ  
**📞 Поддержка**: См. полный план миграции для деталей
