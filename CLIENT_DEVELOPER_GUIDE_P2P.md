# 🚀 **CloudBridge Relay - Руководство для разработчика клиентского приложения**

## 📋 **Обзор изменений**

CloudBridge Relay перешел на новую архитектуру P2P mesh сети с использованием QUIC протокола. Все WireGuard зависимости удалены, теперь используется единый стандарт токенов и современные протоколы.

## 🏗️ **Новая архитектура**

### **Схема соединения:**
```
Client A (Office A) ←→ QUIC:9090 ←→ CloudBridge Relay ←→ QUIC:9090 ←→ Client B (Office B)
```

### **Ключевые изменения:**
- ❌ **Удалено:** WireGuard, OpenVPN, сложные конфигурации
- ✅ **Добавлено:** QUIC протокол, JWT аутентификация, упрощенные токены
- ✅ **Улучшено:** Производительность, безопасность, масштабируемость

## 🔑 **Новый формат JWT токенов**

### **Структура токена:**
```json
{
  "protocol_type": "p2p-mesh",
  "scope": "p2p-mesh-claims",
  "tenant_id": "tenant-216420165",
  "server_id": "server-client-server-1758105689169",
  "connection_type": "quic",
  "permissions": ["p2p_connect", "mesh_join"],
  "iat": 1758245498,
  "iss": "https://auth.2gc.ru/realms/cloudbridge",
  "aud": "account",
  "sub": "server-client-server-1758105689169",
  "jti": "jwt_1758245498_unified"
}
```

### **Обязательные поля:**
- `protocol_type`: Должен быть `"p2p-mesh"`
- `connection_type`: Должен быть `"quic"`
- `tenant_id`: Идентификатор организации
- `server_id`: Уникальный идентификатор сервера
- `permissions`: Массив разрешений, должен содержать `"p2p_connect"`

## 🌐 **Протокол подключения**

### **1. Регистрация пира через API**

```bash
POST /api/v1/tenants/{tenant_id}/peers/register
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "public_key": "test-peer-public-key",
  "allowed_ips": ["10.0.0.0/24"]
}
```

**Ответ:**
```json
{
  "success": true,
  "peer_id": "peer_server-client-server-1758105689169",
  "relay_session_id": "rs_a1da47cab11b665646bc39b16eb38d71",
  "registered_at": "2025-09-19T01:31:43.629456228Z"
}
```

### **2. Discovery пиров**

```bash
GET /api/v1/tenants/{tenant_id}/peers/discover
Authorization: Bearer <jwt_token>
```

**Ответ:**
```json
{
  "success": true,
  "tenant_id": "tenant-216420165",
  "peers": [
    {
      "id": "peer_server-client-server-1758105689169",
      "peer_id": "peer_server-client-server-1758105689169",
      "tenant_id": "tenant-216420165",
      "device_id": "server-client-server-1758105689169",
      "connection_type": "quic",
      "relay_session_id": "rs_a1da47cab11b665646bc39b16eb38d71",
      "is_online": true,
      "last_seen": "2025-09-19T01:31:43.629456141Z",
      "registered_at": "2025-09-19T01:31:43.629456228Z",
      "mesh_status": "connected"
    }
  ]
}
```

## 🔌 **QUIC соединение**

### **1. Подключение к relay серверу**

```go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
    
    "github.com/quic-go/quic-go"
)

func connectToRelay(relayAddr, jwtToken string) (*quic.Connection, error) {
    // Конфигурация TLS
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true, // Для тестирования
        NextProtos:         []string{"cloudbridge-p2p"},
    }
    
    // Подключение к relay серверу
    conn, err := quic.DialAddr(context.Background(), relayAddr, tlsConfig, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to relay: %v", err)
    }
    
    return &conn, nil
}
```

### **2. Аутентификация**

```go
func authenticate(conn quic.Connection, jwtToken string) error {
    // Открываем поток для аутентификации
    authStream, err := conn.OpenStreamSync(context.Background())
    if err != nil {
        return fmt.Errorf("failed to open auth stream: %v", err)
    }
    defer authStream.Close()
    
    // Отправляем JWT токен
    authMessage := fmt.Sprintf("AUTH %s", jwtToken)
    _, err = authStream.Write([]byte(authMessage))
    if err != nil {
        return fmt.Errorf("failed to send auth message: %v", err)
    }
    
    // Читаем подтверждение
    buffer := make([]byte, 1024)
    n, err := authStream.Read(buffer)
    if err != nil {
        return fmt.Errorf("failed to read auth response: %v", err)
    }
    
    response := string(buffer[:n])
    if response != "AUTH_OK" {
        return fmt.Errorf("authentication failed: %s", response)
    }
    
    return nil
}
```

## 📨 **P2P сообщения**

### **Формат сообщений:**
```
"TO:<peer_id>:<payload>"
```

### **Примеры сообщений:**

```go
// Ping до другого пира
message := "TO:peer_server-client-server-1758105689169:ping 10.100.77.11"

// SSH соединение
message := "TO:peer_server-client-server-1758105689169:ssh admin@10.100.77.11"

// HTTP запрос
message := "TO:peer_server-client-server-1758105689169:GET /api/data HTTP/1.1"

// Передача файла
message := "TO:peer_server-client-server-1758105689169:file_transfer data.bin"
```

### **Отправка сообщения:**

```go
func sendP2PMessage(conn quic.Connection, targetPeerID, payload string) error {
    // Открываем поток для данных
    dataStream, err := conn.OpenStreamSync(context.Background())
    if err != nil {
        return fmt.Errorf("failed to open data stream: %v", err)
    }
    defer dataStream.Close()
    
    // Формируем сообщение
    message := fmt.Sprintf("TO:%s:%s", targetPeerID, payload)
    
    // Отправляем
    _, err = dataStream.Write([]byte(message))
    if err != nil {
        return fmt.Errorf("failed to send message: %v", err)
    }
    
    return nil
}
```

## 🌍 **Сетевая архитектура**

### **Выделение IP адресов:**
```bash
# Каждый tenant получает свою подсеть
tenant-216420165 → 10.100.77.0/24
tenant-123456789 → 10.100.78.0/24
tenant-987654321 → 10.100.79.0/24
```

### **Автоматическое назначение IP:**
```bash
# При регистрации пира
Client A → 10.100.77.10
Client B → 10.100.77.11
Client C → 10.100.77.12
```

### **Прямое общение:**
```bash
# Клиенты могут напрямую обращаться друг к другу
ping 10.100.77.11
ssh user@10.100.77.11
curl http://10.100.77.11:8080
```

## 🔄 **Heartbeat система**

### **Endpoint для heartbeat:**
```bash
POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat
Authorization: Bearer <jwt_token>
```

### **Интервал heartbeat:**
- **Рекомендуемый:** 30 секунд
- **Минимальный:** 10 секунд
- **Максимальный:** 60 секунд

### **Реализация heartbeat клиента:**

```go
func startHeartbeat(peerID, tenantID, jwtToken string) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            err := sendHeartbeat(peerID, tenantID, jwtToken)
            if err != nil {
                log.Printf("Heartbeat failed: %v", err)
            }
        }
    }
}

func sendHeartbeat(peerID, tenantID, jwtToken string) error {
    url := fmt.Sprintf("http://relay-server:8082/api/v1/tenants/%s/peers/%s/heartbeat", 
                      tenantID, peerID)
    
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return err
    }
    
    req.Header.Set("Authorization", "Bearer "+jwtToken)
    
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
    }
    
    return nil
}
```

## 🛠️ **Полный пример клиента**

```go
package main

import (
    "context"
    "crypto/tls"
    "fmt"
    "log"
    "time"
    
    "github.com/quic-go/quic-go"
)

type CloudBridgeClient struct {
    relayAddr    string
    jwtToken     string
    tenantID     string
    serverID     string
    peerID       string
    conn         quic.Connection
}

func NewCloudBridgeClient(relayAddr, jwtToken, tenantID, serverID string) *CloudBridgeClient {
    return &CloudBridgeClient{
        relayAddr: relayAddr,
        jwtToken:  jwtToken,
        tenantID:  tenantID,
        serverID:  serverID,
        peerID:    fmt.Sprintf("peer_%s", serverID),
    }
}

func (c *CloudBridgeClient) Connect() error {
    // Подключение к relay
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true,
        NextProtos:         []string{"cloudbridge-p2p"},
    }
    
    conn, err := quic.DialAddr(context.Background(), c.relayAddr, tlsConfig, nil)
    if err != nil {
        return fmt.Errorf("failed to connect: %v", err)
    }
    
    c.conn = conn
    
    // Аутентификация
    err = c.authenticate()
    if err != nil {
        return fmt.Errorf("authentication failed: %v", err)
    }
    
    // Запуск heartbeat
    go c.startHeartbeat()
    
    // Запуск обработки входящих сообщений
    go c.handleIncomingMessages()
    
    return nil
}

func (c *CloudBridgeClient) authenticate() error {
    authStream, err := c.conn.OpenStreamSync(context.Background())
    if err != nil {
        return err
    }
    defer authStream.Close()
    
    authMessage := fmt.Sprintf("AUTH %s", c.jwtToken)
    _, err = authStream.Write([]byte(authMessage))
    if err != nil {
        return err
    }
    
    buffer := make([]byte, 1024)
    n, err := authStream.Read(buffer)
    if err != nil {
        return err
    }
    
    response := string(buffer[:n])
    if response != "AUTH_OK" {
        return fmt.Errorf("auth failed: %s", response)
    }
    
    return nil
}

func (c *CloudBridgeClient) SendMessage(targetPeerID, payload string) error {
    dataStream, err := c.conn.OpenStreamSync(context.Background())
    if err != nil {
        return err
    }
    defer dataStream.Close()
    
    message := fmt.Sprintf("TO:%s:%s", targetPeerID, payload)
    _, err = dataStream.Write([]byte(message))
    return err
}

func (c *CloudBridgeClient) handleIncomingMessages() {
    for {
        stream, err := c.conn.AcceptStream(context.Background())
        if err != nil {
            log.Printf("Failed to accept stream: %v", err)
            continue
        }
        
        go func(s quic.Stream) {
            defer s.Close()
            
            buffer := make([]byte, 4096)
            n, err := s.Read(buffer)
            if err != nil {
                log.Printf("Failed to read message: %v", err)
                return
            }
            
            message := string(buffer[:n])
            log.Printf("Received message: %s", message)
            
            // Обработка входящего сообщения
            c.processIncomingMessage(message)
        }(stream)
    }
}

func (c *CloudBridgeClient) processIncomingMessage(message string) {
    // Парсинг и обработка входящего сообщения
    log.Printf("Processing message: %s", message)
}

func (c *CloudBridgeClient) startHeartbeat() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            // Отправка heartbeat через API
            // Реализация sendHeartbeat из примера выше
        }
    }
}

func (c *CloudBridgeClient) Close() error {
    if c.conn != nil {
        return c.conn.CloseWithError(0, "client closing")
    }
    return nil
}

// Пример использования
func main() {
    client := NewCloudBridgeClient(
        "relay-server:9090",
        "your-jwt-token-here",
        "tenant-216420165",
        "server-client-server-1758105689169",
    )
    
    err := client.Connect()
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer client.Close()
    
    // Отправка сообщения
    err = client.SendMessage("peer_other-server", "Hello from client!")
    if err != nil {
        log.Printf("Failed to send message: %v", err)
    }
    
    // Ожидание
    select {}
}
```

## 📋 **Чек-лист для миграции**

### **Что нужно изменить в существующем коде:**

- [ ] **Удалить WireGuard зависимости** - больше не нужны
- [ ] **Обновить JWT токены** - использовать новый формат
- [ ] **Заменить протокол** - с WireGuard на QUIC
- [ ] **Обновить аутентификацию** - использовать новый протокол AUTH
- [ ] **Изменить формат сообщений** - использовать "TO:peer_id:payload"
- [ ] **Добавить heartbeat** - для поддержания соединений
- [ ] **Обновить API вызовы** - использовать новые endpoints

### **Новые зависимости:**
```go
// Добавить в go.mod
require (
    github.com/quic-go/quic-go v0.40.0
)
```

## 🚀 **Готово к использованию!**

Новая архитектура CloudBridge Relay готова к использованию. Все компоненты протестированы и работают в Minikube. Клиентские приложения могут подключаться к relay серверу и общаться друг с другом как в локальной сети.

**Для получения JWT токенов обратитесь к разработчику личного кабинета!**
