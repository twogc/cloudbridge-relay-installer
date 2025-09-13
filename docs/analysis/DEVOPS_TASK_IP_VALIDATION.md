# 🎯 DEVOPS ЗАДАНИЕ: НАСТРОЙКА IP ВАЛИДАЦИИ ДЛЯ WIREGUARD ENDPOINTS

## 📋 **ОПИСАНИЕ ПРОБЛЕМЫ**

CloudBridge Relay сервер отклоняет регистрацию пиров с ошибкой:
```
{"error":"Invalid WireGuard endpoint","code":400,"message":"endpoint IP not allowed: 192.168.1.100"}
```

**Причина:** IP валидатор настроен с пустым whitelist'ом разрешенных IP адресов.

## 🔍 **АНАЛИЗ ПРОЕКТА (ОБЯЗАТЕЛЬНО ПРОЧИТАТЬ!)**

### **Архитектура проекта:**
- **Язык:** Go (Golang)
- **Фреймворк:** Gin (HTTP router)
- **Контейнеризация:** Docker + Kubernetes (Minikube)
- **Основной сервис:** CloudBridge Relay (P2P mesh network)

### **Ключевые файлы для изучения:**
1. `internal/api/peer_handlers.go` - обработка регистрации пиров
2. `internal/api/ip_validator.go` - валидация IP адресов
3. `configs/config.json` - основная конфигурация
4. `k8s/deployments/cloudbridge-relay-deployment.yaml` - K8s deployment
5. `k8s/secrets/cloudbridge-secrets.yaml` - секреты

### **Текущее состояние:**
- ✅ JWT токены работают корректно
- ✅ Health endpoints доступны
- ✅ Peer discovery функционирует
- ❌ IP валидация блокирует все endpoint'ы

## 🎯 **ЗАДАЧИ**

### **1. ИЗУЧЕНИЕ ПРОЕКТА (ОБЯЗАТЕЛЬНО!)**

**Перед любыми изменениями:**
```bash
# 1. Изучить структуру проекта
find . -name "*.go" -path "*/api/*" | head -10
find . -name "*.yaml" -path "*/k8s/*" | head -10

# 2. Понять текущую конфигурацию
cat configs/config.json | jq '.security.allowed_ips'
cat configs/config.json | jq '.tenants."relay-test-tenant".allowed_ips'

# 3. Изучить IP валидатор
grep -n "IsIPAllowed\|AddAllowedIP\|AddAllowedCIDR" internal/api/ip_validator.go

# 4. Понять инициализацию PeerManager
grep -A 20 "NewPeerManager" internal/api/peer_handlers.go
```

### **2. НАСТРОЙКА IP WHITELIST**

**Вариант A: Через конфигурацию (РЕКОМЕНДУЕТСЯ)**
```json
{
  "security": {
    "allowed_ips": [
      "10.0.0.0/8",
      "172.16.0.0/12", 
      "192.168.0.0/16",
      "203.0.113.0/24",
      "198.51.100.0/24"
    ]
  }
}
```

**Вариант B: Через код (если конфигурация не поддерживается)**
```go
// В internal/api/peer_handlers.go в функции NewPeerManager
ipValidator.AddAllowedCIDR("10.0.0.0/8")
ipValidator.AddAllowedCIDR("172.16.0.0/12")
ipValidator.AddAllowedCIDR("192.168.0.0/16")
```

### **3. ТЕСТИРОВАНИЕ**

**Команды для проверки:**
```bash
# 1. Проверить health
curl "http://95.163.250.190:30082/health"

# 2. Проверить peer discovery
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://95.163.250.190:30082/api/v1/peers/discover?tenant_id=test-tenant-123"

# 3. Проверить регистрацию пира
curl -X POST \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"peer_id":"test-peer-001","tenant_id":"test-tenant-123","org_id":"test-org-456","device_id":"device_1757631924","connection_type":"hybrid","wireguard_config":{"endpoint":"192.168.1.100:51820","public_key":"dGVzdC1wdWJsaWMta2V5LWZvci10ZXN0aW5nLW9ubHk=","allowed_ips":["10.0.0.0/24"],"listen_port":51820}}' \
  "http://95.163.250.190:30082/api/v1/peers/register"
```

## ⚠️ **КРИТИЧЕСКИ ВАЖНО - НЕ СЛОМАТЬ!**

### **Что НЕЛЬЗЯ трогать:**
1. **JWT секреты** - уже настроены правильно
2. **Kubernetes секреты** - содержат рабочие ключи
3. **Docker образы** - уже собраны и работают
4. **Основную логику** - только IP whitelist!

### **Что можно изменять:**
1. **Конфигурационные файлы** - `configs/config.json`
2. **IP валидатор** - только добавление разрешенных IP
3. **Kubernetes ConfigMaps** - если нужно

### **Проверки перед деплоем:**
```bash
# 1. Проверить синтаксис Go
go build -o cloudbridge-relay cmd/relay/*.go

# 2. Проверить JSON конфигурацию
jq . configs/config.json

# 3. Проверить Kubernetes манифесты
kubectl apply --dry-run=client -f k8s/
```

## 🚀 **ПЛАН ВЫПОЛНЕНИЯ**

### **Этап 1: Изучение (30 мин)**
1. Прочитать документацию проекта
2. Изучить текущую конфигурацию
3. Понять архитектуру IP валидации

### **Этап 2: Настройка (15 мин)**
1. Добавить разрешенные IP диапазоны
2. Обновить конфигурацию
3. Пересобрать приложение

### **Этап 3: Тестирование (15 мин)**
1. Проверить health endpoints
2. Протестировать peer registration
3. Убедиться, что ничего не сломалось

### **Этап 4: Деплой (10 мин)**
1. Обновить Docker образ
2. Загрузить в Minikube
3. Перезапустить deployment

## 📊 **ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ**

### **До исправления:**
```json
{"error":"Invalid WireGuard endpoint","code":400,"message":"endpoint IP not allowed: 192.168.1.100"}
```

### **После исправления:**
```json
{"success":true,"peer_id":"test-peer-001","registered_at":"2025-09-12T01:00:00Z","discovery_endpoint":"wss://relay.2gc.ru/ws/peers"}
```

## 🔧 **ПОЛЕЗНЫЕ КОМАНДЫ**

### **Мониторинг:**
```bash
# Логи приложения
kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay --tail=20

# Статус подов
kubectl get pods -n cloudbridge

# Статус deployment
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge
```

### **Отладка:**
```bash
# Проверить конфигурацию в поде
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- cat /app/configs/config.json

# Проверить переменные окружения
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- env | grep JWT
```

## 📝 **КРИТЕРИИ УСПЕХА**

✅ **Задание выполнено, если:**
1. Peer registration работает с IP `192.168.1.100:51820`
2. Health endpoints остались доступными
3. JWT токены продолжают работать
4. Peer discovery функционирует
5. Логи не содержат ошибок валидации IP

## 🆘 **ПОДДЕРЖКА**

**При возникновении проблем:**
1. Проверить логи: `kubectl logs -n cloudbridge -l app.kubernetes.io/name=cloudbridge-relay`
2. Проверить конфигурацию: `kubectl get configmap -n cloudbridge`
3. Проверить секреты: `kubectl get secret -n cloudbridge`

**Важно:** Не изменять JWT секреты и Kubernetes секреты без крайней необходимости!

---

## 🎯 **ИТОГОВАЯ ЦЕЛЬ**

Настроить IP whitelist для WireGuard endpoints, чтобы peer registration работал с IP адресами из частных сетей (192.168.x.x, 10.x.x.x, 172.16.x.x), не нарушив при этом существующую функциональность.

**Время выполнения:** ~1 час
**Приоритет:** Высокий
**Риск:** Низкий (только конфигурационные изменения)
