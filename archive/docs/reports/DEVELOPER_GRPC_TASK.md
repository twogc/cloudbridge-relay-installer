# 🚀 ЗАДАНИЕ РАЗРАБОТЧИКУ: Исправление gRPC подключения клиента

## 📋 ПРОБЛЕМА
Клиент `cloudbridge-client` не может подключиться к gRPC серверу CloudBridge Relay. Сервер запускается, но клиент получает ошибку `context deadline exceeded`.

## 🔍 ЧТО НУЖНО ПРОВЕРИТЬ

### 1. Статус gRPC сервера
```bash
# Проверить статус подов
kubectl get pods -n cloudbridge | grep cloudbridge-relay

# Проверить логи запуска gRPC сервера
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 | grep -i grpc

# Должны увидеть:
# "gRPC Integration started","address":"[::]:8444"
# "gRPC server integration started successfully"
```

### 2. Конфигурация gRPC сервера
**Файл:** `config/config.json`
```json
"grpc": {
  "enabled": true,
  "port": 8444,
  "tls_enabled": false,  // ← ВАЖНО: отключить TLS для тестирования
  "cert_file": "/app/certs/tls.crt",
  "key_file": "/app/certs/tls.key",
  "mtls_enabled": false,
  "keepalive_ping": "15s",
  "keepalive_timeout": "10s",
  "max_recv_msg_size": 4194304,
  "max_send_msg_size": 4194304
}
```

### 3. Проверка сетевого подключения
```bash
# Проверить, слушает ли gRPC сервер порт 8444 внутри пода
kubectl exec -n cloudbridge deployment/cloudbridge-relay -- netstat -tlnp | grep 8444

# Настроить port-forward для тестирования
kubectl port-forward -n cloudbridge service/cloudbridge-relay-external 8444:8444 &

# Проверить доступность порта на хосте
netstat -tlnp | grep 8444

# Тестировать подключение
telnet localhost 8444
```

### 4. Конфигурация клиента
**Файл:** `test/relay-client/config.yaml`
```yaml
# API Configuration
api:
  base_url: "grpc://localhost:8444"          # gRPC протокол + port-forward
  timeout: "30s"
  insecure_skip_verify: true

# Authentication Configuration  
auth:
  type: "jwt"  # или "none" для тестирования
  secret: "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
  skip_validation: true  # для тестирования
```

## 🔑 ТОКЕНЫ ДЛЯ АВТОРИЗАЦИИ

### 1. JWT Secret (основной)
```
85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=
```

### 2. Тестовые токены
**Файл:** `test/relay-client/real-token.txt`
```
eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItMTc1ODI3NjE4NjM5NyIsImNvbm5lY3Rpb25fdHlwZSI6InF1aWMiLCJwZXJtaXNzaW9ucyI6WyJwMnBfY29ubmVjdCIsIm1lc2hfam9pbiIsIm1lc2hfbWFuYWdlIl0sImlhdCI6MTc1ODMyNjUyMywiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLTE3NTgyNzYxODYzOTciLCJqdGkiOiJqd3RfMTc1ODMyNjUyM194dmoxbmM0YiJ9.AlnsRyngVs5BWZqhJHCBKFUxq9Qs5grkFCo4BA7VUrA
```

**Декодированный payload:**
```json
{
  "protocol_type": "p2p-mesh",
  "connection_type": "quic",
  "tenant_id": "tenant-216420165",
  "server_id": "server-1758276186397",
  "permissions": ["p2p_connect", "mesh_join", "mesh_manage"]
}
```

### 3. Генерация новых токенов
**CF-SaaS токен генератор:** `test/cf-saas/`
```bash
cd test/cf-saas
npm run dev
# Открыть: http://localhost:4321/api/v1/tokens/generate-zitadel
```

**Создание тестового токена:**
```bash
cd test/relay-client
node create-test-token.js
```

## 📂 КЛЮЧЕВЫЕ ФАЙЛЫ КОДА

### 1. gRPC Сервер
**Основной файл:** `cmd/relay/main.go`
- Строки ~2000-2100: Инициализация gRPC сервера
- Функция: `setupGRPCIntegration()`

**gRPC Integration:** `internal/transport/grpc/integration.go`
- Функция: `Start()` - запуск gRPC сервера
- Функция: `Stop()` - остановка gRPC сервера

**gRPC Server:** `internal/transport/grpc/server.go`
- Функция: `NewServer()` - создание gRPC сервера
- Функция: `Start()` - запуск сервера на порту 8444

### 2. gRPC Клиент
**Основной файл:** `test/relay-client/` (бинарный файл)
- Конфигурация: `config.yaml`
- Команда: `./cloudbridge-client tunnel --config config.yaml --transport grpc`

### 3. Аутентификация
**JWT Validator:** `internal/zitadel/validator.go`
- Функция: `ValidateToken()` - валидация JWT токенов
- Использует secret: `85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE=`

**Auth Adapter:** `internal/auth/adapter.go`
- Интерфейс: `Provider` - провайдер аутентификации

## 🧪 ТЕСТИРОВАНИЕ

### 1. Запуск сервера с отключенным TLS
```bash
# Применить конфигурацию без TLS
kubectl apply -f k8s/configmap-no-tls.yaml
kubectl apply -f k8s/deployment-no-tls.yaml
kubectl apply -f k8s/service-no-tls.yaml

# Настроить port-forward
kubectl port-forward -n cloudbridge service/cloudbridge-relay-no-tls 8444:8444 &
```

### 2. Тестирование клиента
```bash
cd test/relay-client

# Создать простую конфигурацию
cat > config-test.yaml << 'EOF'
api:
  base_url: "grpc://localhost:8444"
  timeout: "30s"
  insecure_skip_verify: true

auth:
  type: "none"
  skip_validation: true

relay:
  host: "localhost"
  ports:
    grpc: 8444
  tls:
    enabled: false
EOF

# Запустить клиент
./cloudbridge-client tunnel \
  --config config-test.yaml \
  --tunnel-id test-tunnel \
  --transport grpc \
  --insecure-skip-tls-verify \
  --log-level debug
```

### 3. Проверка логов
```bash
# Логи сервера
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=50

# Логи с фильтрацией
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls --tail=100 | grep -i -E "(grpc|connection|client|error|tunnel)"

# Проверка входящих соединений
kubectl logs -n cloudbridge -l app=cloudbridge-relay-no-tls -f
# Затем в другом терминале запустить клиент
```

## ❌ ИЗВЕСТНЫЕ ПРОБЛЕМЫ

### 1. TLS сертификаты
- **Проблема:** Неправильные пути к сертификатам
- **Решение:** Отключить TLS (`"tls_enabled": false`) или исправить пути

### 2. NodePort не работает
- **Проблема:** Minikube не пробрасывает NodePort на хост
- **Решение:** Использовать `kubectl port-forward`

### 3. Валидация токенов
- **Проблема:** Клиент не может пройти аутентификацию
- **Решение:** Временно отключить валидацию (`"skip_validation": true`)

## 🎯 ЦЕЛЬ
**Добиться успешного подключения клиента к gRPC серверу без ошибок `context deadline exceeded`.**

## 📞 РЕЗУЛЬТАТ
После исправления должны увидеть в логах сервера:
```
[grpc] Incoming connection from client
[tunnel] New tunnel created: test-tunnel
[grpc] Client authenticated successfully
```

А клиент должен вывести:
```
[relay-client] INFO: Connected to gRPC server successfully
[relay-client] INFO: Tunnel established: test-tunnel
```
