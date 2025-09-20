# 📚 **CloudBridge Relay - Документация**

## 🎯 **Обзор**

Данная документация содержит полную информацию о проекте CloudBridge Relay - интеллектуальной P2P mesh сети с AI-оптимизацией и интеграцией Cadence Workflow Engine.

---

## 📁 **Структура документации**

### 📊 **Отчеты (reports/)**
- **[CADENCE_P2P_MIGRATION_REPORT.md](reports/CADENCE_P2P_MIGRATION_REPORT.md)** - Отчет о миграции Cadence с WireGuard на P2P QUIC
- **[MINIKUBE_DEPLOYMENT_REPORT.md](reports/MINIKUBE_DEPLOYMENT_REPORT.md)** - Отчет о развертывании в Minikube
- **[AI_TESTING_REPORT.md](reports/AI_TESTING_REPORT.md)** - Отчет о тестировании AI компонентов
- **[DIAGNOSTIC_REPORT.md](reports/DIAGNOSTIC_REPORT.md)** - Отчет о диагностике проблем
- **[FIXES_IMPLEMENTATION_REPORT.md](reports/FIXES_IMPLEMENTATION_REPORT.md)** - Отчет о реализации исправлений

### 📖 **Руководства (guides/)**
- **[MINIKUBE_BUILD_GUIDE.md](../MINIKUBE_BUILD_GUIDE.md)** - Инструкция по сборке и развертыванию в Minikube
- **[DEVELOPER_BRIEFING.md](../DEVELOPER_BRIEFING.md)** - Понимание проекта и архитектуры
- **[DEVELOPER_CADENCE_MIGRATION_TASK.md](../DEVELOPER_CADENCE_MIGRATION_TASK.md)** - Задание по миграции Cadence

### 🔌 **API (api/)**
- **[AI_API_ENDPOINTS.md](api/AI_API_ENDPOINTS.md)** - Документация AI API endpoints
- **[P2P_API_ENDPOINTS.md](api/P2P_API_ENDPOINTS.md)** - Документация P2P API endpoints
- **[CADENCE_API_ENDPOINTS.md](api/CADENCE_API_ENDPOINTS.md)** - Документация Cadence API endpoints

### 🚀 **Развертывание (deployment/)**
- **[KUBERNETES_CONFIGURATION.md](deployment/KUBERNETES_CONFIGURATION.md)** - Конфигурация Kubernetes
- **[MONITORING_SETUP.md](deployment/MONITORING_SETUP.md)** - Настройка мониторинга
- **[NETWORK_POLICIES.md](deployment/NETWORK_POLICIES.md)** - Сетевые политики

---

## 🏗️ **Архитектура системы**

### **Основные компоненты:**
```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                       │
│                                                             │
│ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐ │
│ │ P2P QUIC        │ │ AI Components   │ │ Cadence         │ │
│ │ Manager         │ │                 │ │ Workflow        │ │
│ │                 │ │ • AI Service    │ │ Engine          │ │
│ │ • QUIC Server   │ │ • AI Analytics  │ │                 │ │
│ │ • STUN Server   │ │ • AI Worker     │ │ • P2P QUIC      │ │
│ │ • MASQUE Server │ │ • ML Models     │ │   Workflows     │ │
│ │ • ICE Support   │ │ • Predictions   │ │ • AI            │ │
│ │ • NAT Traversal │ │ • Optimization  │ │   Optimization  │ │
│ └─────────────────┘ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### **Ключевые возможности:**
- ✅ **P2P QUIC соединения** с поддержкой ICE/STUN/TURN
- ✅ **AI-оптимизированная маршрутизация** с ML предсказаниями
- ✅ **Cadence Workflow Engine** для оркестрации
- ✅ **Real-time аналитика** и мониторинг
- ✅ **Мультитенантная изоляция** и безопасность
- ✅ **Автоматическое масштабирование** и восстановление

---

## 🚀 **Быстрый старт**

### **1. Развертывание в Minikube:**
```bash
# Следовать инструкциям в MINIKUBE_BUILD_GUIDE.md
cd /home/ubuntu/cloudbridge-relay-installer
minikube start
docker build -t cloudbridge-relay:cadence-p2p-v1 .
minikube image load cloudbridge-relay:cadence-p2p-v1
kubectl apply -f k8s/deployment.yaml -n cloudbridge
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge
```

### **2. Проверка статуса:**
```bash
# Проверить pods
kubectl get pods -n cloudbridge

# Проверить services
kubectl get services -n cloudbridge

# Проверить логи
kubectl logs deployment/cloudbridge-relay -n cloudbridge
```

### **3. Тестирование API:**
```bash
# HTTP API
curl http://$(minikube ip):30080/health

# P2P API
curl http://$(minikube ip):32500/health

# AI Service
kubectl port-forward service/ai-service 8083:8083 -n cloudbridge &
curl http://localhost:8083/health
```

---

## 🔧 **Конфигурация**

### **Основные конфигурационные файлы:**
- `configs/config.json` - Основная конфигурация
- `k8s/deployment.yaml` - Kubernetes deployment
- `k8s/cloudbridge-relay-external-service.yaml` - Service конфигурация
- `k8s/network-policy-prometheus.yaml` - Сетевые политики
- `k8s/ai-service-endpoints.yaml` - AI API endpoints
- `k8s/monitoring-config.yaml` - Конфигурация мониторинга

### **Переменные окружения:**
```bash
CADENCE_ENABLED=true
CADENCE_HOST_PORT=cadence:7933
CADENCE_DOMAIN=cloudbridge-p2p
CADENCE_TASK_LIST=p2p-quic-tasks
```

---

## 📊 **Мониторинг**

### **Доступные метрики:**
- **AI Service метрики** - точность предсказаний, использование ресурсов
- **P2P метрики** - количество соединений, латентность, пропускная способность
- **Cadence метрики** - статус workflows, время выполнения
- **Системные метрики** - CPU, память, диск, сеть

### **Алерты:**
- HighLatency - низкая точность AI предсказаний
- AIServiceDown - AI Service недоступен
- HighMemoryUsage - высокое использование памяти
- HighCPUUsage - высокое использование CPU
- AnomaliesDetected - обнаружено много аномалий
- CadenceWorkflowFailed - сбой Cadence workflows

---

## 🧪 **Тестирование**

### **API тестирование:**
```bash
# AI Service
curl http://localhost:8083/api/v1/ai/status
curl http://localhost:8083/api/v1/ai/models

# AI Analytics
curl http://localhost:8084/health
curl http://localhost:8084/api/v1/analytics/metrics

# P2P API
curl http://localhost:8082/health
curl -X POST http://localhost:8082/api/v1/tenants/test/peers/register \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"public_key": "test-key", "allowed_ips": ["10.0.0.1/32"]}'
```

### **Cadence тестирование:**
```bash
# Cadence Web UI
kubectl port-forward service/cadence-web-service 8088:8088 -n cloudbridge
# Открыть http://localhost:8088
```

---

## 🔍 **Отладка**

### **Полезные команды:**
```bash
# Логи компонентов
kubectl logs deployment/cloudbridge-relay -n cloudbridge
kubectl logs deployment/ai-service -n cloudbridge
kubectl logs deployment/ai-analytics -n cloudbridge
kubectl logs deployment/cadence -n cloudbridge

# Статус ресурсов
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
kubectl get configmaps -n cloudbridge
kubectl get networkpolicies -A

# Сетевые подключения
kubectl exec -it deployment/ai-service -n cloudbridge -c ai-service -- nslookup prometheus-service
kubectl exec -it deployment/ai-service -n cloudbridge -c ai-service -- wget -qO- http://prometheus-service:9090/api/v1/query?query=up
```

---

## 📞 **Поддержка**

### **Контакты:**
- **Документация:** `/home/ubuntu/cloudbridge-relay-installer/docs/`
- **Конфигурации:** `/home/ubuntu/cloudbridge-relay-installer/k8s/`
- **Логи:** `kubectl logs deployment/cloudbridge-relay -n cloudbridge`

### **Полезные ссылки:**
- [Minikube Documentation](https://minikube.sigs.k8s.io/docs/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Cadence Documentation](https://cadenceworkflow.io/docs/)
- [QUIC Protocol](https://quicwg.org/)

---

**🎯 CloudBridge Relay - Интеллектуальная P2P mesh сеть с AI-оптимизацией!**
