# CloudBridge Relay - Конфигурации для тестового пилота

## 🎯 Готовые конфигурации для клиента

### 📋 Информация о токенах:
- **Алгоритм**: HS256
- **Срок действия**: Без ограничений (permanent)
- **Аудитория**: relay-test-tenant
- **Tenant ID**: mesh-network-test
- **Назначение**: P2P mesh connectivity

## 🚀 Конфигурация для Server A:

### Файл: `server-a-pilot.yaml`
```yaml
# CloudBridge Client Configuration - Server A (Pilot)
# Использует внешний доступ через поддомены

# P2P Configuration
p2p:
  server_id: "server-a"
  tenant_id: "mesh-network-test"
  # Внешние поддомены для интернет доступа
  relay_url: "https://relay.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
  admin_url: "https://admin.2gc.ru"
  monitor_url: "https://monitor.2gc.ru"
  api_url: "https://api.2gc.ru"

# Authentication Configuration
auth:
  type: "jwt"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWEiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIyMDU4NywiaWF0IjoxNzU5MjIwNTg3fQ.2BhA0Tlm6yhrUcHbZ7oICJB9lxn9_Gm1z2rdAGmxyZA"
  secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
  skip_validation: false

# Relay Configuration
relay:
  enabled: true
  # Основной релей через поддомен
  url: "https://relay.2gc.ru"
  # MASQUE прокси для HTTP/3
  masque_url: "https://masque.2gc.ru"
  # P2P discovery
  p2p_url: "https://p2p.2gc.ru"

# Network Configuration
network:
  # Локальная сеть для P2P
  local_network: "10.100.0.0/16"
  # P2P порт
  p2p_port: 51820
  # HTTP API порт
  api_port: 8080

# Logging
log:
  level: "debug"
  format: "json"

# Monitoring
monitoring:
  enabled: true
  metrics_url: "https://monitor.2gc.ru/metrics"
  health_url: "https://relay.2gc.ru/health"
```

## 🚀 Конфигурация для Server B:

### Файл: `server-b-pilot.yaml`
```yaml
# CloudBridge Client Configuration - Server B (Pilot)
# Использует внешний доступ через поддомены

# P2P Configuration
p2p:
  server_id: "server-b"
  tenant_id: "mesh-network-test"
  # Внешние поддомены для интернет доступа
  relay_url: "https://relay.2gc.ru"
  p2p_url: "https://p2p.2gc.ru"
  admin_url: "https://admin.2gc.ru"
  monitor_url: "https://monitor.2gc.ru"
  api_url: "https://api.2gc.ru"

# Authentication Configuration
auth:
  type: "jwt"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYiIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWIiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIyMDU4NywiaWF0IjoxNzU5MjIwNTg3fQ.e3SseNclshGsQexlIsr8KjD5vtZESo3mFTa5Xwhll9E"
  secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
  skip_validation: false

# Relay Configuration
relay:
  enabled: true
  # Основной релей через поддомен
  url: "https://relay.2gc.ru"
  # MASQUE прокси для HTTP/3
  masque_url: "https://masque.2gc.ru"
  # P2P discovery
  p2p_url: "https://p2p.2gc.ru"

# Network Configuration
network:
  # Локальная сеть для P2P
  local_network: "10.100.0.0/16"
  # P2P порт
  p2p_port: 51820
  # HTTP API порт
  api_port: 8080

# Logging
log:
  level: "debug"
  format: "json"

# Monitoring
monitoring:
  enabled: true
  metrics_url: "https://monitor.2gc.ru/metrics"
  health_url: "https://relay.2gc.ru/health"
```

## 🔧 Инструкции для клиента:

### 1. **Установка клиента:**
```bash
# Скачать и распаковать cloudbridge-client
# Убедиться, что бинарный файл исполняемый
chmod +x cloudbridge-client
```

### 2. **Запуск Server A:**
```bash
./cloudbridge-client p2p --config server-a-pilot.yaml --log-level debug
```

### 3. **Запуск Server B:**
```bash
./cloudbridge-client p2p --config server-b-pilot.yaml --log-level debug
```

### 4. **Проверка подключения:**
- Оба сервера должны подключиться к релею
- В логах должно появиться сообщение о регистрации
- Серверы должны увидеть друг друга в P2P mesh

## 🌐 Внешние endpoints для тестирования:

### **Health Check:**
```bash
curl -k https://edge.2gc.ru/health
```

### **API регистрация:**
```bash
curl -X POST -k https://edge.2gc.ru/api/v1/tenants/mesh-network-test/peers/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJzZXJ2ZXItYSIsInRlbmFudF9pZCI6Im1lc2gtbmV0d29yay10ZXN0Iiwic2VydmVyX2lkIjoic2VydmVyLWEiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljIiwicHJvdG9jb2xfdHlwZSI6InAycC1tZXNoIiwicGVybWlzc2lvbnMiOlsicmVsYXlfY29ubmVjdCIsInJlbGF5X3R1bm5lbCIsInAycF9kaXNjb3ZlciIsInAycF9jb25uZWN0Il0sImlzcyI6ImNsb3VkYnJpZGdlLXJlbGF5IiwiYXVkIjpbInJlbGF5LXRlc3QtdGVuYW50Il0sIm5iZiI6MTc1OTIyMDU4NywiaWF0IjoxNzU5MjIwNTg3fQ.2BhA0Tlm6yhrUcHbZ7oICJB9lxn9_Gm1z2rdAGmxyZA" \
  -d '{"peer_id": "test-peer", "public_key": "test-key", "allowed_ips": ["10.100.0.1/32"]}'
```

## 📊 Ожидаемые результаты:

### 1. **Server A подключение:**
- Регистрация с `server_id: "server-a"`
- Получение IP адреса в сети `10.100.0.0/16`
- Подключение к P2P mesh

### 2. **Server B подключение:**
- Регистрация с `server_id: "server-b"`
- Получение IP адреса в сети `10.100.0.0/16`
- Подключение к P2P mesh

### 3. **P2P соединение:**
- Server A и Server B должны увидеть друг друга
- Установление прямого P2P соединения
- Обмен данными через mesh сеть

## ⚠️ Важные замечания:

1. **Токены действительны** - без ограничений по времени
2. **Внешний доступ работает** - через поддомены
3. **Безопасность** - JWT токены с подписью
4. **Мониторинг** - доступен через https://monitor.2gc.ru
5. **Администрирование** - доступно через https://admin.2gc.ru

## 🎯 Готово к тестовому пилоту!

Конфигурации готовы для передачи клиенту. Оба сервера смогут подключиться к CloudBridge Relay и образовать P2P mesh сеть.
