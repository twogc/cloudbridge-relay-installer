# CloudBridge Relay - Финальное руководство по развертыванию

## Обзор

Это финальное руководство по развертыванию CloudBridge Relay с TURN Pool функциональностью в Minikube, следуя принципам MINIKUBE_BUILD_GUIDE.md.

## Что развертывается

- **CloudBridge Relay** с TURN Pool Manager (2 реплики)
- **Cassandra** база данных
- **AI Service** с Telegram интеграцией
- **AI Analytics** сервис
- **22 сервиса** включая NodePort и LoadBalancer
- **3 namespace**: cloudbridge, cloudbridge-monitoring, monitoring

## Быстрый старт

### **Автоматическое развертывание (рекомендуется):**
```bash
cd /home/ubuntu/cloudbridge-relay-installer
./deploy-final.sh
```

### **Проверка статуса:**
```bash
./check-status.sh
```

### **Очистка:**
```bash
./cleanup.sh
```

## Доступные скрипты

| Скрипт | Описание | Следует MINIKUBE_BUILD_GUIDE.md |
|--------|----------|--------------------------------|
| `./deploy-final.sh` | Финальное развертывание | Да |
| `./deploy-existing-images.sh` | Развертывание с существующими образами |  Нет |
| `./deploy-with-build-guide.sh` | Развертывание с пересборкой | Да |
| `./check-status.sh` | Проверка статуса | Да |
| `./cleanup.sh` | Очистка развертывания | Да |

##  Принципы MINIKUBE_BUILD_GUIDE.md

### **1. Уникальные теги образов**
```bash
TAG="turn-pool-$(date +%Y%m%d-%H%M%S)"
echo "Используем тег: $TAG"
```

### **2. Сборка Docker образа**
```bash
docker build -t cloudbridge-relay:$TAG .
```

### **3. Загрузка в Minikube**
```bash
minikube image load cloudbridge-relay:$TAG
```

### **4. Развертывание в Kubernetes**
```bash
kubectl apply -f deployment.yaml -n cloudbridge
```

## Доступ к сервисам

### **Port Forward (для тестирования):**
```bash
# CloudBridge Relay
kubectl port-forward svc/cloudbridge-relay 9090:9090 -n cloudbridge

# AI Service
kubectl port-forward svc/ai-service 8083:8083 -n cloudbridge

# Cassandra
kubectl port-forward svc/cassandra-service 9042:9042 -n cloudbridge
```

### **NodePort (внешний доступ):**
```bash
# Получение IP Minikube
minikube ip

# CloudBridge Relay
curl http://$(minikube ip):30990/health

# AI Service
curl http://$(minikube ip):30083/health
```

## 🧪 Тестирование

### **Health Checks:**
```bash
# CloudBridge Relay
curl -v http://localhost:9090/health

# AI Service
curl -v http://localhost:8083/health

# Cassandra
kubectl exec -n cloudbridge deployment/cassandra -- cqlsh -e "describe keyspaces"
```

### **TURN Pool тестирование:**
```bash
# Проверка логов TURN Pool
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i "turn\|pool"

# Проверка конфигурации
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- cat /app/configs/config.json | grep -A 10 "turn_pool"
```

## Troubleshooting

### **Общие проблемы:**

#### **Проблема: TURN Pool отключен**
```bash
# Проверка конфигурации
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- cat /app/configs/config.json | jq '.turn_pool.enabled'

# Проверка логов
kubectl logs -n cloudbridge deployment/cloudbridge-relay | grep -i "turn"
```

#### **Проблема: Pod не запускается**
```bash
kubectl describe pod <pod-name> -n cloudbridge
kubectl logs <pod-name> -n cloudbridge
```

#### **Проблема: Сервис недоступен**
```bash
kubectl get endpoints -n cloudbridge
kubectl describe service <service-name> -n cloudbridge
```

### **Полезные команды:**
```bash
# Просмотр всех ресурсов
kubectl get all -n cloudbridge

# Мониторинг ресурсов
kubectl top pods -n cloudbridge

# Вход в под
kubectl exec -it <pod-name> -n cloudbridge -- /bin/bash

# Проверка событий
kubectl get events -n cloudbridge --sort-by='.lastTimestamp'
```

## Ожидаемый результат

После успешного развертывания вы должны увидеть:

```
NAME                                 READY   STATUS    RESTARTS   AGE
ai-analytics-578f9bbd5d-7zf8b        1/1     Running   0          2h
ai-service-7b9c657dbc-x6p48          1/1     Running   0          1h
cassandra-776f646b95-qnr7l           1/1     Running   0          1h
cloudbridge-relay-5fd7988f86-cwjfp   1/1     Running   0          1h
cloudbridge-relay-5fd7988f86-jjb4r   1/1     Running   0          1h
```

### **TURN Pool логи:**
```
{"level":"INFO","timestamp":"2025-09-25T02:27:04.704Z","caller":"logger/logger.go:132","msg":"TURN Pool Manager initialized","default_pool":"primary","pools":2,"strategy":"weighted"}
{"level":"INFO","timestamp":"2025-09-25T02:27:04.704Z","caller":"logger/logger.go:132","msg":"Added TURN server to pool","pool":"primary","server":"turn1.cloudbridge.ru:3478","region":"ru-central1"}
```

## Ключевые особенности

### **TURN Pool Manager:**
- Load balancing с весовыми коэффициентами
- Health checks для TURN серверов
- Failover между пулами
- Метрики и мониторинг
- Интеграция с ICE Agent

### **CloudBridge Relay:**
- QUIC-first архитектура
- MASQUE протокол
- P2P API
- AI Analytics
- Post-Quantum Cryptography

### **AI Service:**
- Telegram интеграция
- Anomaly Detection
- Predictive Analytics
- Intelligent Routing
- Security Intelligence

## Готово!

CloudBridge Relay с TURN Pool функциональностью успешно развернут по принципам MINIKUBE_BUILD_GUIDE.md!

### **Следующие шаги:**
1. Протестировать TURN Pool функциональность
2. Настроить мониторинг
3. Оптимизировать производительность
4. Настроить production окружение

---

*Документ создан: 2025-09-25*  
*Статус: Готов к использованию*
