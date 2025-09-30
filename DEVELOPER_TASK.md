# Техническое задание: Исправление авторизации в CloudBridge Relay

## Проблема

В текущей реализации CloudBridge Relay есть критические проблемы с авторизацией:

1. **Сервер всегда требует авторизацию** - даже когда `AUTH_ENABLED=false`, сервер возвращает 401 Unauthorized
2. **Клиенты не могут подключиться** - получают ошибку "Invalid token" при попытке регистрации
3. **P2P mesh не работает** - клиенты не могут установить соединение через relay

## Текущее состояние

### Сервер (cloudbridge-relay)
- **Порт**: 32500 (NodePort в Minikube)
- **API Endpoint**: `http://192.168.58.2:32500/api/v1/tenants/{tenant_id}/peers/register`
- **Проблема**: Функции `validateToken` в `internal/api/peer_handlers.go` и `internal/api/p2p_handlers.go` не проверяют переменную окружения `AUTH_ENABLED`

### Клиенты (cloudbridge-client)
- **Конфигурация**: `test/relay-client/config-test-server-a.yaml` и `config-test-server-b.yaml`
- **Тип авторизации**: `auth.type: "none"` (отключена)
- **Проблема**: Клиенты генерируют токены в формате `"server-id:server-a"`, но сервер их не принимает

## Что нужно исправить

### 1. Исправить серверную авторизацию

**Файлы для изменения:**
- `internal/api/peer_handlers.go` - функция `validateToken`
- `internal/api/p2p_handlers.go` - функция `validateToken`

**Требуемые изменения:**

```go
func (pm *PeerManager) validateToken(c *gin.Context) (*P2PClaims, error) {
    // Проверяем, включена ли авторизация
    if os.Getenv("AUTH_ENABLED") == "false" {
        // Возвращаем фиктивные claims для отключенной авторизации
        tenantID := c.Param("tenant_id")
        if tenantID == "" {
            tenantID = "default"
        }
        return &P2PClaims{
            Subject:        "anonymous",
            TenantID:       tenantID,
            ServerID:       "anonymous",
            ConnectionType: "quic",
            ProtocolType:   "p2p-mesh",
            Permissions:    []string{"relay_connect", "relay_tunnel"},
        }, nil
    }
    
    // ... остальная логика авторизации
}
```

### 2. Исправить обработку server_id в токенах

**Проблема**: Клиенты отправляют токены в формате `"server-id:server-a"`, но сервер не извлекает из них `server_id`.

**Решение**: В функции `RegisterPeer` в `internal/api/peer_handlers.go` добавить логику извлечения `server_id`:

```go
// Генерируем peer_id из server_id в токене
pm.logger.Info("DEBUG: ServerID from token", zap.String("server_id", claims.ServerID))

// Check if token is in special format "server-id:server-a"
var actualServerID string
if strings.HasPrefix(claims.ServerID, "server-id:") {
    actualServerID = strings.TrimPrefix(claims.ServerID, "server-id:")
    pm.logger.Info("DEBUG: Extracted server_id from special format", zap.String("extracted_server_id", actualServerID))
} else {
    actualServerID = claims.ServerID
}

peerID := generatePeerID(actualServerID)
pm.logger.Info("DEBUG: Generated peer_id", zap.String("peer_id", peerID))
```

### 3. Обновить deployment

**Файл**: `k8s/deployment-clean.yaml`

**Изменения:**
1. Обновить image tag на новый
2. Убедиться, что переменные окружения установлены:
   ```yaml
   env:
   - name: AUTH_ENABLED
     value: "false"
   - name: RELAY_JWT_SECRET
     value: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
   ```

## Как тестировать

### 1. Подготовка окружения

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать новый образ
TAG="auth-fix-$(date +%Y%m%d-%H%M%S)"
docker build -t cloudbridge-relay:$TAG .
docker save cloudbridge-relay:$TAG | minikube image load

# Обновить deployment
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=60s
```

### 2. Тестирование сервера

```bash
# Проверить, что сервер работает
kubectl get pods -n cloudbridge

# Проверить логи сервера
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20

# Тестировать API напрямую
curl -X POST http://192.168.58.2:32500/api/v1/tenants/mesh-network-test/peers/register \
  -H "Content-Type: application/json" \
  -d '{"peer_id": "test-peer", "endpoint": "192.168.1.100:5553"}'
```

**Ожидаемый результат**: HTTP 200 OK с ответом, содержащим выделенный IP-адрес.

### 3. Тестирование клиентов

```bash
# Убить все запущенные клиенты
pkill -f cloudbridge-client

# Запустить Server A
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
timeout 10s ./cloudbridge-client p2p --config config-test-server-a.yaml --log-level debug

# Запустить Server B (в другом терминале)
timeout 10s ./cloudbridge-client p2p --config config-test-server-b.yaml --log-level debug
```

**Ожидаемый результат**: 
- Клиенты успешно регистрируются (статус 200)
- Получают выделенные IP-адреса
- Устанавливают P2P соединение

### 4. Проверка P2P mesh

```bash
# Проверить логи сервера на регистрацию клиентов
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 | grep -E "(POST|register|peer_id)"

# Проверить выделенные IP-адреса
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 | grep -E "(allocated|IP|peer)"
```

## Генерация токенов для клиентов

### Текущая реализация (auth.type: "none")

Клиенты используют упрощенную авторизацию без JWT:

```yaml
# config-test-server-a.yaml
auth:
  type: "none"
  token: ""
  secret: ""
  skip_validation: true
  server_id: "server-a"  # Динамический server_id
```

**Как это работает:**
1. Клиент читает `server_id` из конфигурации
2. Генерирует токен в формате `"server-id:server-a"`
3. Отправляет токен в заголовке `Authorization: Bearer server-id:server-a`
4. Сервер извлекает `server_id` из токена и использует для генерации `peer_id`

### Альтернативная реализация (JWT токены)

Если нужна полная JWT авторизация:

```yaml
# config-test-server-a.yaml
auth:
  type: "jwt"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  secret: "Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh"
  server_id: "server-a"
```

**Генерация JWT токена:**
```bash
# Используя jwt.io или аналогичный инструмент
# Header: {"alg": "HS256", "typ": "JWT"}
# Payload: {"sub": "server-a", "tenant_id": "mesh-network-test", "server_id": "server-a"}
# Secret: Wahm7Iti1VAoXsFAV1MBeMfPv1yope2IJ47utB30Pk4FgZyJFZK0ayfX4zNstkUh
```

## Критерии успеха

1. **Сервер принимает запросы без авторизации** когда `AUTH_ENABLED=false`
2. **Клиенты успешно регистрируются** и получают IP-адреса
3. **P2P mesh работает** - клиенты могут обмениваться данными
4. **Логи показывают успешную регистрацию** пиров
5. **Выделенные IP-адреса** корректно назначаются клиентам

## Дополнительные заметки

- **Порты**: Убедиться, что health check использует правильный порт (5552, а не 5551)
- **Конфигурация**: Проверить, что ConfigMap не переопределяет настройки авторизации
- **Отладка**: Добавить подробные логи для отслеживания процесса авторизации
- **Тестирование**: Использовать уникальные теги Docker для каждой сборки

## Файлы для изменения

1. `internal/api/peer_handlers.go` - основная логика авторизации
2. `internal/api/p2p_handlers.go` - P2P API авторизация  
3. `k8s/deployment-clean.yaml` - конфигурация deployment
4. `test/relay-client/config-test-server-*.yaml` - конфигурация клиентов

## Команды для быстрого тестирования

```bash
# Полная пересборка и тестирование
cd /home/ubuntu/cloudbridge-relay-installer
TAG="test-$(date +%Y%m%d-%H%M%S)"
docker build -t cloudbridge-relay:$TAG .
docker save cloudbridge-relay:$TAG | minikube image load
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=60s

# Тестирование клиентов
pkill -f cloudbridge-client
cd test/relay-client
timeout 15s ./cloudbridge-client p2p --config config-test-server-a.yaml --log-level debug
```
