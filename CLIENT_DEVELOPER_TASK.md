# 🎯 Техническое задание для разработчика клиента Relay

## 📋 **Задача: Адаптация клиента для QUIC + ICE/STUN/TURN архитектуры**

### **🎯 Цель:**
Перевести клиент с WireGuard на новую архитектуру QUIC + ICE/STUN/TURN для обеспечения P2P mesh network без открытия портов на клиентской стороне.

---

## 🔄 **Изменения в архитектуре**

### **❌ Старая архитектура (WireGuard):**
```
Client ←→ WireGuard:51820 ←→ Relay ←→ WireGuard:51820 ←→ Client
```

### **✅ Новая архитектура (QUIC + ICE/STUN/TURN):**
```
Client ←→ QUIC:9090 + ICE:8082 + STUN:19302 ←→ Relay ←→ QUIC:9090 + ICE:8082 + STUN:19302 ←→ Client
```

---

## 🌐 **DNS и порты для edge.2gc.ru**

### **Обязательные endpoints:**
```
edge.2gc.ru:8080   # HTTP API (Health, Registration)
edge.2gc.ru:8082   # P2P API (ICE signaling)
edge.2gc.ru:9090   # QUIC Transport (Main P2P)
edge.2gc.ru:19302  # STUN Server (NAT discovery)
```

### **Рекомендуемые endpoints:**
```
edge.2gc.ru:8443   # HTTPS/MASQUE (Corporate proxy)
edge.2gc.ru:9092   # Enhanced QUIC (Advanced features)
```

---

## 🔧 **Изменения в коде клиента**

### **1. Удалить WireGuard компоненты:**

#### **Удалить файлы:**
```bash
# Удалить все WireGuard связанные файлы
rm -rf wireguard/
rm -f wg0.conf
rm -f wg0.key
rm -f wg0.pub
```

#### **Удалить зависимости:**
```go
// Удалить из go.mod
// github.com/WireGuard/wireguard-go
// github.com/vishvananda/netlink
// github.com/vishvananda/netns
```

#### **Удалить импорты:**
```go
// Удалить из всех .go файлов
import (
    // "github.com/WireGuard/wireguard-go"
    // "github.com/vishvananda/netlink"
    // "github.com/vishvananda/netns"
)
```

### **2. Добавить QUIC + ICE/STUN/TURN компоненты:**

#### **Добавить зависимости в go.mod:**
```go
require (
    github.com/quic-go/quic-go v0.40.0
    github.com/pion/ice/v2 v2.3.0
    github.com/pion/stun v0.6.0
    github.com/pion/turn/v2 v2.1.0
    github.com/pion/webrtc/v3 v3.2.0
)
```

#### **Создать новые файлы:**

**`internal/ice/agent.go`**
```go
package ice

import (
    "context"
    "net"
    "github.com/pion/ice/v2"
    "github.com/pion/stun"
)

type ICEAgent struct {
    agent     *ice.Agent
    stunServers []string
    turnServers []string
}

func NewICEAgent(stunServers, turnServers []string) *ICEAgent {
    return &ICEAgent{
        stunServers: stunServers,
        turnServers: turnServers,
    }
}

func (a *ICEAgent) Start() error {
    config := &ice.AgentConfig{
        NetworkTypes: []ice.NetworkType{ice.NetworkTypeUDP4},
        STUNServers:  a.stunServers,
        TURNServers:  a.turnServers,
    }
    
    agent, err := ice.NewAgent(config)
    if err != nil {
        return err
    }
    
    a.agent = agent
    return nil
}

func (a *ICEAgent) GatherCandidates() ([]*ice.Candidate, error) {
    return a.agent.GetLocalCandidates()
}

func (a *ICEAgent) AddRemoteCandidate(candidate *ice.Candidate) error {
    return a.agent.AddRemoteCandidate(candidate)
}

func (a *ICEAgent) StartConnectivityChecks() error {
    return a.agent.StartConnectivityChecks()
}
```

**`internal/quic/connection.go`**
```go
package quic

import (
    "context"
    "crypto/tls"
    "net"
    "github.com/quic-go/quic-go"
)

type QUICConnection struct {
    conn    quic.Connection
    streams map[string]quic.Stream
}

func NewQUICConnection() *QUICConnection {
    return &QUICConnection{
        streams: make(map[string]quic.Stream),
    }
}

func (q *QUICConnection) Connect(addr string) error {
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true, // Для разработки
    }
    
    conn, err := quic.DialAddr(context.Background(), addr, tlsConfig, nil)
    if err != nil {
        return err
    }
    
    q.conn = conn
    return nil
}

func (q *QUICConnection) CreateStream() (quic.Stream, error) {
    return q.conn.OpenStreamSync(context.Background())
}

func (q *QUICConnection) AcceptStream() (quic.Stream, error) {
    return q.conn.AcceptStream(context.Background())
}
```

**`internal/p2p/manager.go`**
```go
package p2p

import (
    "context"
    "encoding/json"
    "net/http"
    "github.com/quic-go/quic-go"
)

type P2PManager struct {
    iceAgent    *ice.ICEAgent
    quicConn    *quic.QUICConnection
    relayURL    string
    sessionID   string
    peerID      string
}

func NewP2PManager(relayURL, peerID string) *P2PManager {
    return &P2PManager{
        relayURL: relayURL,
        peerID:   peerID,
    }
}

func (p *P2PManager) Initialize() error {
    // Инициализация ICE агента
    stunServers := []string{"edge.2gc.ru:19302"}
    turnServers := []string{} // Пока не используем TURN
    
    p.iceAgent = ice.NewICEAgent(stunServers, turnServers)
    if err := p.iceAgent.Start(); err != nil {
        return err
    }
    
    // Инициализация QUIC соединения
    p.quicConn = quic.NewQUICConnection()
    
    return nil
}

func (p *P2PManager) ConnectToPeer(targetPeerID string) error {
    // 1. Собрать ICE candidates
    candidates, err := p.iceAgent.GatherCandidates()
    if err != nil {
        return err
    }
    
    // 2. Отправить candidates на relay
    if err := p.sendCandidatesToRelay(candidates); err != nil {
        return err
    }
    
    // 3. Получить remote candidates
    remoteCandidates, err := p.getRemoteCandidatesFromRelay()
    if err != nil {
        return err
    }
    
    // 4. Добавить remote candidates
    for _, candidate := range remoteCandidates {
        if err := p.iceAgent.AddRemoteCandidate(candidate); err != nil {
            return err
        }
    }
    
    // 5. Начать connectivity checks
    if err := p.iceAgent.StartConnectivityChecks(); err != nil {
        return err
    }
    
    // 6. Установить QUIC соединение
    return p.establishQUICConnection()
}

func (p *P2PManager) sendCandidatesToRelay(candidates []*ice.Candidate) error {
    url := p.relayURL + ":8082/api/v1/ice/candidates"
    
    payload := map[string]interface{}{
        "session_id":  p.sessionID,
        "peer_id":     p.peerID,
        "candidates":  candidates,
    }
    
    jsonData, _ := json.Marshal(payload)
    
    resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

### **3. Обновить конфигурацию:**

**`config/config.yaml`**
```yaml
relay:
  host: "edge.2gc.ru"
  ports:
    http_api: 8080      # HTTP API
    p2p_api: 8082      # P2P API (ICE signaling)
    quic: 9090         # QUIC Transport
    stun: 19302        # STUN Server
    masque: 8443       # HTTPS/MASQUE
    enhanced_quic: 9092 # Enhanced QUIC

ice:
  stun_servers:
    - "edge.2gc.ru:19302"
  turn_servers: []
  timeout: "30s"

quic:
  handshake_timeout: "10s"
  idle_timeout: "30s"
  max_streams: 100

p2p:
  max_connections: 1000
  session_timeout: "300s"
```

### **4. Обновить main.go:**

**`cmd/client/main.go`**
```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "github.com/your-org/cloudbridge-client/internal/p2p"
    "github.com/your-org/cloudbridge-client/internal/config"
)

func main() {
    // Загрузить конфигурацию
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Создать P2P Manager
    p2pManager := p2p.NewP2PManager(
        cfg.Relay.Host,
        cfg.Client.PeerID,
    )
    
    // Инициализировать P2P Manager
    if err := p2pManager.Initialize(); err != nil {
        log.Fatal("Failed to initialize P2P Manager:", err)
    }
    
    // Зарегистрироваться на relay
    if err := p2pManager.Register(); err != nil {
        log.Fatal("Failed to register on relay:", err)
    }
    
    // Ожидать сигналов
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    
    log.Println("CloudBridge Client started successfully")
    log.Println("Press Ctrl+C to stop")
    
    <-c
    log.Println("Shutting down...")
}
```

---

## 📡 **API Endpoints для edge.2gc.ru**

### **1. Регистрация клиента:**
```http
POST http://edge.2gc.ru:8080/api/v1/peers/register
Content-Type: application/json
Authorization: Bearer <JWT_TOKEN>

{
    "peer_id": "client-123",
    "public_key": "client-public-key",
    "endpoints": ["192.168.1.100:9090"],
    "capabilities": ["quic", "ice", "masque"]
}
```

### **2. ICE Signaling:**
```http
POST http://edge.2gc.ru:8082/api/v1/ice/candidates
Content-Type: application/json
Authorization: Bearer <JWT_TOKEN>

{
    "session_id": "session-123",
    "peer_id": "client-123",
    "candidates": [
        {
            "foundation": "1",
            "component": 1,
            "transport": "udp",
            "priority": 2130706431,
            "address": "192.168.1.100",
            "port": 9090,
            "type": "host"
        }
    ]
}
```

### **3. P2P Connection:**
```http
POST http://edge.2gc.ru:8082/api/v1/p2p/connect
Content-Type: application/json
Authorization: Bearer <JWT_TOKEN>

{
    "session_id": "session-123",
    "target_peer_id": "client-456",
    "protocol": "quic"
}
```

### **4. Health Check:**
```http
GET http://edge.2gc.ru:8080/health
```

---

## 🔧 **Порядок реализации:**

### **Этап 1: Подготовка (1-2 дня)**
1. ✅ Удалить WireGuard компоненты
2. ✅ Добавить новые зависимости
3. ✅ Создать базовую структуру ICE/QUIC

### **Этап 2: ICE Implementation (3-5 дней)**
1. ✅ Реализовать ICE Agent
2. ✅ Интеграция с STUN сервером
3. ✅ Сбор и обмен candidates

### **Этап 3: QUIC Implementation (3-5 дней)**
1. ✅ Реализовать QUIC соединения
2. ✅ Управление потоками
3. ✅ Connection migration

### **Этап 4: P2P Manager (2-3 дня)**
1. ✅ Интеграция ICE + QUIC
2. ✅ API взаимодействие с relay
3. ✅ Управление сессиями

### **Этап 5: Testing (2-3 дня)**
1. ✅ Unit тесты
2. ✅ Integration тесты
3. ✅ End-to-end тесты

---

## 🎯 **Критерии готовности:**

### **✅ Функциональные требования:**
- [ ] Клиент подключается к edge.2gc.ru без WireGuard
- [ ] ICE candidates собираются и обмениваются
- [ ] QUIC соединения устанавливаются
- [ ] P2P соединения работают через relay
- [ ] Connection migration работает
- [ ] MASQUE support для корпоративных сетей

### **✅ Технические требования:**
- [ ] Нет открытых портов на клиентской стороне
- [ ] Все соединения outbound
- [ ] NAT traversal работает
- [ ] Автоматический fallback на MASQUE
- [ ] Логирование и мониторинг

### **✅ Производительность:**
- [ ] Время установки соединения < 5 секунд
- [ ] Пропускная способность > 100 Mbps
- [ ] Задержка < 50ms
- [ ] Стабильность соединения > 99%

---

## 🚀 **Готово к продакшену:**

После выполнения всех этапов клиент будет готов для работы с новой QUIC + ICE/STUN/TURN архитектурой и обеспечит:

- ✅ **P2P mesh network** без открытия портов
- ✅ **"Почти LAN" опыт** на L3 уровне
- ✅ **NAT traversal** через ICE/STUN/TURN
- ✅ **Connection migration** через QUIC
- ✅ **Корпоративная совместимость** через MASQUE

**Срок выполнения: 2-3 недели** 🎯
