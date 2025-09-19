# 🎯 Задание для разработчика сервера CloudBridge Relay

## 📋 **Задача: Унификация токенов и удаление WireGuard**

### **Контекст:**
После анализа архитектуры CloudBridge Relay и изучения подхода Cloudflare, необходимо обновить сервер для поддержки упрощенных токенов и удаления поддержки WireGuard (больше не используется).

## 🔄 **Что нужно изменить**

### **1. Унифицировать структуру токенов**

#### **❌ Текущая проблема:**
Сервер ожидает разные поля в зависимости от `connection_type`:
- `connection_type=wireguard` → требует `wireguard_config`
- `connection_type=quic` → не поддерживается
- `protocol_type=p2p-mesh` → не поддерживается с `connection_type=quic`

#### **✅ Новая унифицированная структура:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169",
  "connection_type": "quic",
  "permissions": ["p2p_connect", "mesh_join"],
  "iat": 1758244379,
  "iss": "https://auth.2gc.ru/realms/cloudbridge",
  "aud": "account",
  "sub": "server-client-server-1758105689169",
  "jti": "jwt_1758244379_okc19k5d9"
}
```

### **2. Удалить поддержку WireGuard**

#### **Файлы для изменения:**
- `internal/token/validator.go` - валидация токенов
- `internal/token/p2p_claims.go` - структуры P2P claims
- `internal/api/peer_handlers.go` - обработка регистрации пиров
- `cmd/relay/main.go` - основная логика сервера

## 🔧 **Технические изменения**

### **1. Обновить `internal/token/p2p_claims.go`**

#### **Убрать WireGuard поля:**
```go
// ❌ УДАЛИТЬ
type WireGuardConfig struct {
    InterfaceName string   `json:"interface_name"`
    ListenPort    int      `json:"listen_port"`
    Address       string   `json:"address"`
    MTU           int      `json:"mtu"`
    AllowedIPs    []string `json:"allowed_ips"`
}
```

#### **Обновить P2PClaims:**
```go
type P2PClaims struct {
    jwt.RegisteredClaims
    ProtocolType   string           `json:"protocol_type"`
    Scope          string           `json:"scope"`
    TenantID       string           `json:"tenant_id"`
    ServerID       string           `json:"server_id"`
    ConnectionType string           `json:"connection_type"`
    Permissions    []string         `json:"permissions"`
    // ❌ УДАЛИТЬ WireGuardConfig
    // ❌ УДАЛИТЬ NetworkConfig
    // ❌ УДАЛИТЬ MeshConfig
    // ❌ УДАЛИТЬ PeerWhitelist
}
```

### **2. Обновить `internal/token/validator.go`**

#### **Упростить валидацию:**
```go
func (v *TokenValidator) ValidateP2PToken(tokenString string) (*P2PClaims, error) {
    // 1. Парсить JWT токен
    token, err := jwt.ParseWithClaims(tokenString, &P2PClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(v.secretKey), nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("invalid token: %v", err)
    }
    
    claims, ok := token.Claims.(*P2PClaims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }
    
    // 2. Проверить обязательные поля
    if claims.ProtocolType != "p2p-mesh" {
        return nil, fmt.Errorf("unsupported protocol: %s", claims.ProtocolType)
    }
    
    if claims.ConnectionType != "quic" {
        return nil, fmt.Errorf("unsupported connection type: %s", claims.ConnectionType)
    }
    
    if claims.TenantID == "" {
        return nil, fmt.Errorf("missing tenant_id")
    }
    
    if claims.ServerID == "" {
        return nil, fmt.Errorf("missing server_id")
    }
    
    // 3. Проверить разрешения
    if !contains(claims.Permissions, "p2p_connect") {
        return nil, fmt.Errorf("missing p2p_connect permission")
    }
    
    return claims, nil
}
```

### **3. Обновить `internal/api/peer_handlers.go`**

#### **Упростить регистрацию пира:**
```go
func (h *PeerHandler) handleRegisterPeer(c *gin.Context) {
    // 1. Получить токен из заголовка
    tokenString := c.GetHeader("Authorization")
    if !strings.HasPrefix(tokenString, "Bearer ") {
        c.JSON(401, gin.H{"error": "Missing or invalid authorization header"})
        return
    }
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")
    
    // 2. Валидировать токен
    claims, err := h.tokenValidator.ValidateP2PToken(tokenString)
    if err != nil {
        c.JSON(401, gin.H{"error": "Invalid token", "message": err.Error()})
        return
    }
    
    // 3. Получить данные регистрации
    var req RelayAssistedRegistrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid request format", "message": err.Error()})
        return
    }
    
    // 4. Зарегистрировать пира
    peerID := fmt.Sprintf("peer_%s", claims.ServerID)
    relaySessionID := fmt.Sprintf("rs_%s", generateRandomID())
    
    // 5. Вернуть ответ
    c.JSON(200, gin.H{
        "peer_id": peerID,
        "relay_session_id": relaySessionID,
        "tenant_id": claims.TenantID,
        "connection_type": claims.ConnectionType,
        "status": "registered",
    })
}
```

## 📁 **Файлы для изменения**

### **Основные файлы:**
```
internal/token/
├── validator.go          # Валидация токенов
├── p2p_claims.go         # Структуры P2P claims
└── ...

internal/api/
├── peer_handlers.go      # Обработка регистрации пиров
├── peer_models.go        # Модели данных
└── ...

cmd/relay/
├── main.go               # Основная логика сервера
└── ...
```

### **Конфигурационные файлы:**
```
k8s/
├── deployments/
│   └── cloudbridge-relay-p2p-deployment.yaml
└── services/
    └── p2p-api-service.yaml
```

## 🐳 **Инструкция по сборке Minikube**

### **1. Обновить код:**
```bash
# Внести изменения в файлы выше
# Убрать поддержку WireGuard
# Упростить валидацию токенов
```

### **2. Собрать образ:**
```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать Docker образ
docker build -t cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens .

# Загрузить в Minikube
minikube image load cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens
```

### **3. Обновить deployment:**
```bash
# Обновить образ в deployment
kubectl set image deployment/cloudbridge-relay-p2p \
  cloudbridge-relay-p2p=cloudbridge-relay-p2p:$(date +%Y%m%d-%H%M%S)-unified-tokens \
  -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl logs -f deployment/cloudbridge-relay-p2p -n cloudbridge
```

### **4. Протестировать:**
```bash
# Тестировать с новым токеном
curl -X POST "http://edge.2gc.ru:8083/api/v1/tenants/tenant-216420165/peers/register" \
  -H "Authorization: Bearer NEW_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "test-peer-public-key",
    "allowed_ips": ["10.0.0.0/24"]
  }'
```

## 📋 **Чек-лист для разработчика**

### **Код:**
- [ ] Убрать `WireGuardConfig` из `p2p_claims.go`
- [ ] Убрать `NetworkConfig` из `p2p_claims.go`
- [ ] Убрать `MeshConfig` из `p2p_claims.go`
- [ ] Убрать `PeerWhitelist` из `p2p_claims.go`
- [ ] Упростить валидацию в `validator.go`
- [ ] Поддержать `connection_type=quic` с `protocol_type=p2p-mesh`
- [ ] Обновить обработку регистрации в `peer_handlers.go`

### **Тестирование:**
- [ ] Собрать новый Docker образ
- [ ] Загрузить в Minikube
- [ ] Обновить deployment
- [ ] Протестировать с новым токеном
- [ ] Проверить, что старые токены не работают
- [ ] Проверить, что новые токены работают

## 🎯 **Ожидаемый результат**

### **После выполнения:**
1. ✅ **Единый стандарт токенов** - только необходимые поля
2. ✅ **Поддержка QUIC** - `connection_type=quic` работает
3. ✅ **Удален WireGuard** - больше не поддерживается
4. ✅ **Упрощенная валидация** - меньше ошибок
5. ✅ **Совместимость с новыми токенами** - работает с упрощенной структурой

### **Пример рабочего токена:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsInRlbmFudF9pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJzZXJ2ZXJfaWQiOiJzZXJ2ZXItY2xpZW50LXNlcnZlci0xNzU4MTA1Njg5MTY5IiwiY29ubmVjdGlvbl90eXBlIjoicXVpYyIsInBlcm1pc3Npb25zIjpbInAycF9jb25uZWN0IiwibWVzaF9qb2luIl0sImlhdCI6MTc1ODI0NDM3OSwiaXNzIjoiaHR0cHM6Ly9hdXRoLjJnYy5ydS9yZWFsbXMvY2xvdWRicmlkZ2UiLCJhdWQiOiJhY2NvdW50Iiwic3ViIjoic2VydmVyLWNsaWVudC1zZXJ2ZXItMTc1ODEwNTY4OTE2OSIsImp0aSI6Imp3dF8xNzU4MjQ0Mzc5X29rYzE5azVkOSJ9.iyuA6zUT7CyqW3HPUIpStc4XLdyl7YMzHJ7qo_3GmIY
```

## 📞 **Поддержка**

- **Документация:** `CORRECT_MINIMAL_APPROACH.md`
- **Пример токена:** `generate-jwt-token.go`
- **Тестирование:** CloudBridge Relay на `edge.2gc.ru:8083`

---

**🎯 Цель: Создать единый стандарт токенов для P2P mesh сети CloudBridge Relay без поддержки WireGuard!**
