# Техническое задание: Исправление передачи server_id в CloudBridge Relay

## Проблема

Несмотря на то, что авторизация работает и клиенты успешно регистрируются, **server_id не передается корректно** от клиента к серверу.

### Текущее состояние:
- ✅ Клиенты получают статус 200 OK
- ✅ Сервер выделяет IP-адреса (10.100.77.10)
- ❌ **Проблема**: `peer_id` генерируется как `peer_anonymous` вместо `peer_server-a`

### Логи сервера показывают:
```json
{"level":"INFO","timestamp":"2025-09-30T07:17:57.573Z","caller":"api/peer_handlers.go:263","msg":"Allocated IP for peer","peer_id":"peer_anonymous","tenant_id":"mesh-network-test","allocated_ip":"10.100.77.10"}
```

### Логи клиента показывают:
```
DEBUG: P2P.ServerID from config: server-a
DEBUG: Using server_id: server-a
DEBUG: Creating token with server_id: server-a
DEBUG: CreateTokenWithServerID called with server_id: server-a, generated token: server-id:server-a
```

## Анализ проблемы

### 1. Клиентская сторона работает правильно:
- Клиент читает `server_id: "server-a"` из конфигурации
- Генерирует токен в формате `"server-id:server-a"`
- Отправляет токен в заголовке `Authorization: Bearer server-id:server-a`

### 2. Серверная сторона не извлекает server_id:
- Сервер получает токен `"server-id:server-a"`
- Но в функции `validateToken` возвращает `ServerID: "anonymous"`
- В результате `peer_id` генерируется как `peer_anonymous`

## Что нужно исправить

### Файл: `internal/api/peer_handlers.go`

**Проблема в функции `validateToken`:**

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
            ServerID:       "anonymous", // ❌ ПРОБЛЕМА: всегда "anonymous"
            ConnectionType: "quic",
            ProtocolType:   "p2p-mesh",
            Permissions:    []string{"relay_connect", "relay_tunnel"},
        }, nil
    }
    // ... остальная логика
}
```

**Решение:**

```go
func (pm *PeerManager) validateToken(c *gin.Context) (*P2PClaims, error) {
    // Проверяем, включена ли авторизация
    if os.Getenv("AUTH_ENABLED") == "false" {
        // Извлекаем server_id из токена
        authHeader := c.GetHeader("Authorization")
        serverID := "anonymous" // значение по умолчанию
        
        if authHeader != "" {
            token := strings.TrimPrefix(authHeader, "Bearer ")
            if token != authHeader && strings.HasPrefix(token, "server-id:") {
                serverID = strings.TrimPrefix(token, "server-id:")
                pm.logger.Info("DEBUG: Extracted server_id from token", zap.String("server_id", serverID))
            }
        }
        
        tenantID := c.Param("tenant_id")
        if tenantID == "" {
            tenantID = "default"
        }
        return &P2PClaims{
            Subject:        "anonymous",
            TenantID:       tenantID,
            ServerID:       serverID, // ✅ ИСПРАВЛЕНО: используем извлеченный server_id
            ConnectionType: "quic",
            ProtocolType:   "p2p-mesh",
            Permissions:    []string{"relay_connect", "relay_tunnel"},
        }, nil
    }
    // ... остальная логика
}
```

### Файл: `internal/api/p2p_handlers.go`

**Аналогичная проблема в функции `validateToken`:**

```go
func (pm *P2PManager) validateToken(c *gin.Context) (*P2PClaims, error) {
    // Проверяем, включена ли авторизация
    if os.Getenv("AUTH_ENABLED") == "false" {
        // Извлекаем server_id из токена
        authHeader := c.GetHeader("Authorization")
        serverID := "anonymous" // значение по умолчанию
        
        if authHeader != "" {
            token := strings.TrimPrefix(authHeader, "Bearer ")
            if token != authHeader && strings.HasPrefix(token, "server-id:") {
                serverID = strings.TrimPrefix(token, "server-id:")
                pm.logger.Info("DEBUG: Extracted server_id from token", zap.String("server_id", serverID))
            }
        }
        
        tenantID := c.Param("tenant_id")
        if tenantID == "" {
            tenantID = "default"
        }
        return &P2PClaims{
            Subject:        "anonymous",
            TenantID:       tenantID,
            ServerID:       serverID, // ✅ ИСПРАВЛЕНО: используем извлеченный server_id
            ConnectionType: "quic",
            ProtocolType:   "p2p-mesh",
            Permissions:    []string{"relay_connect", "relay_tunnel"},
        }, nil
    }
    // ... остальная логика
}
```

## Как тестировать

### 1. Внести изменения в код
- Обновить `internal/api/peer_handlers.go`
- Обновить `internal/api/p2p_handlers.go`

### 2. Пересобрать и развернуть
```bash
cd /home/ubuntu/cloudbridge-relay-installer
TAG="server-id-fix-$(date +%Y%m%d-%H%M%S)"
docker build -t cloudbridge-relay:$TAG .
docker save cloudbridge-relay:$TAG | minikube image load
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=60s
```

### 3. Протестировать клиентов
```bash
# Убить старые клиенты
pkill -f cloudbridge-client

# Запустить Server A
cd /home/ubuntu/cloudbridge-relay-installer/test/relay-client
timeout 10s ./cloudbridge-client p2p --config config-test-server-a.yaml --log-level debug

# Запустить Server B
timeout 10s ./cloudbridge-client p2p --config config-test-server-b.yaml --log-level debug
```

### 4. Проверить логи сервера
```bash
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=20 | grep -E "(peer_id|server_id|allocated)"
```

## Ожидаемый результат

### До исправления:
```json
{"level":"INFO","msg":"Allocated IP for peer","peer_id":"peer_anonymous","allocated_ip":"10.100.77.10"}
```

### После исправления:
```json
{"level":"INFO","msg":"Allocated IP for peer","peer_id":"peer_server-a","allocated_ip":"10.100.77.10"}
{"level":"INFO","msg":"Allocated IP for peer","peer_id":"peer_server-b","allocated_ip":"10.100.77.11"}
```

## Дополнительные улучшения

### 1. Добавить отладочные логи
```go
pm.logger.Info("DEBUG: Processing token", 
    zap.String("auth_header", authHeader),
    zap.String("extracted_server_id", serverID),
    zap.String("generated_peer_id", peerID))
```

### 2. Валидация server_id
```go
// Проверяем, что server_id не пустой
if serverID == "" || serverID == "anonymous" {
    pm.logger.Warn("Empty or anonymous server_id, using default")
    serverID = "default"
}
```

### 3. Улучшенная обработка токенов
```go
// Поддержка разных форматов токенов
if strings.HasPrefix(token, "server-id:") {
    serverID = strings.TrimPrefix(token, "server-id:")
} else if strings.HasPrefix(token, "server:") {
    serverID = strings.TrimPrefix(token, "server:")
} else {
    // Попытка извлечь из JSON токена
    serverID = extractServerIDFromJSON(token)
}
```

## Критерии успеха

1. **✅ Server A**: `peer_id` должен быть `peer_server-a`
2. **✅ Server B**: `peer_id` должен быть `peer_server-b`
3. **✅ Логи**: Должны показывать правильные `server_id` и `peer_id`
4. **✅ P2P mesh**: Клиенты должны видеть друг друга по правильным `peer_id`
5. **✅ IP-адреса**: Каждый клиент должен получить уникальный IP

## Файлы для изменения

1. `internal/api/peer_handlers.go` - основная логика извлечения server_id
2. `internal/api/p2p_handlers.go` - P2P API логика извлечения server_id

## Команды для быстрого тестирования

```bash
# Полная пересборка и тестирование
cd /home/ubuntu/cloudbridge-relay-installer
TAG="server-id-fix-$(date +%Y%m%d-%H%M%S)"
docker build -t cloudbridge-relay:$TAG .
docker save cloudbridge-relay:$TAG | minikube image load
kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
kubectl rollout status deployment/cloudbridge-relay -n cloudbridge --timeout=60s

# Тестирование
pkill -f cloudbridge-client
cd test/relay-client
timeout 15s ./cloudbridge-client p2p --config config-test-server-a.yaml --log-level debug
timeout 15s ./cloudbridge-client p2p --config config-test-server-b.yaml --log-level debug

# Проверка логов
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50 | grep -E "(peer_id|server_id|allocated)"
```

## Приоритет: ВЫСОКИЙ

Эта проблема критична для правильной работы P2P mesh, так как:
- Клиенты не могут идентифицировать друг друга
- P2P соединения не устанавливаются корректно
- Мониторинг и отладка затруднены
