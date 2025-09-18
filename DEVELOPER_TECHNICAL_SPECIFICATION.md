# 🚀 **CloudBridge Relay - Техническое задание для разработчика**

## 📋 **Обзор проекта**

### **Цель проекта:**
CloudBridge Relay - это P2P mesh сеть для создания "почти LAN" соединений между удаленными серверами клиентов через централизованный relay сервер.

### **Проблема, которую решаем:**
- **NAT/Firewall блокируют прямые P2P соединения** между удаленными серверами
- **Корпоративные сети блокируют P2P протоколы** (BitTorrent, WebRTC, etc.)
- **Сложность настройки традиционных VPN решений** (OpenVPN, WireGuard)
- **Необходимость изоляции** между разными пользователями/тенантами

### **Наше решение:**
- **Все соединения outbound**: Клиент подключается к нам, мы соединяем с другими
- **Современные протоколы**: QUIC + ICE/STUN/TURN для максимальной совместимости
- **Автоматический fallback**: Если прямое соединение не работает, используем relay
- **Мультитунельность**: Разные пользователи могут объединять свои сервера изолированно
- **Простота настройки**: Клиент просто подключается к нашему серверу

## 🏗️ **Архитектура системы**

### **Компоненты:**
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Client A      │    │  CloudBridge     │    │   Client B      │
│  (Office A)     │◄──►│     Relay        │◄──►│  (Office B)     │
│                 │    │                  │    │                 │
│ - QUIC Client   │    │ - P2P API        │    │ - QUIC Client   │
│ - ICE/STUN      │    │ - QUIC Relay     │    │ - ICE/STUN      │
│ - JWT Auth      │    │ - STUN Server    │    │ - JWT Auth      │
└─────────────────┘    │ - MASQUE Proxy   │    └─────────────────┘
                       │ - Multi-tenant   │
                       └──────────────────┘
```

### **Порты и протоколы:**
| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| 8080 | TCP/HTTPS | HTTP API (Health, Registration) | ✅ Работает |
| 8082 | TCP/HTTPS | P2P API (ICE signaling) | ✅ Работает |
| 19302 | UDP | STUN Server | ✅ Работает |
| 9090 | UDP | QUIC Transport | ✅ Работает |
| 8443 | TCP/HTTPS | HTTPS/MASQUE (Corporate proxy) | ✅ Работает |
| 9092 | UDP | Enhanced QUIC (Advanced features) | ✅ Работает |

## 🔍 **Текущее состояние (по результатам тестирования)**

### **✅ ЧТО РАБОТАЕТ:**

1. **Базовая инфраструктура:**
   - ✅ Клиенты успешно запускаются и регистрируются
   - ✅ Сетевые интерфейсы настроены (10.100.77.77, 10.100.77.78)
   - ✅ Relay сервер доступен и отвечает
   - ✅ Echo функциональность работает
   - ✅ JWT аутентификация работает
   - ✅ Все порты доступны с внешней сети

2. **API endpoints:**
   - ✅ HTTP API: `http://edge.2gc.ru:8080/health`
   - ✅ P2P API: `http://edge.2gc.ru:8082/health`
   - ✅ Регистрация пиров работает

### **❌ ЧТО НЕ РАБОТАЕТ (требует доработки):**

1. **P2P маршрутизация:**
   - ❌ Relay не пересылает пакеты между клиентами
   - ❌ Только echo функциональность, не полноценный P2P relay
   - ❌ Клиенты не могут общаться друг с другом

2. **Heartbeat/Keep-alive:**
   - ❌ Клиенты не остаются онлайн после регистрации
   - ❌ P2P API показывает `"online_peers":0,"total_peers":0`
   - ❌ Нет механизма поддержания соединения

3. **QUIC + ICE соединения:**
   - ❌ Клиенты не могут установить QUIC соединения через relay
   - ❌ ICE кандидаты не обмениваются
   - ❌ Нет реализации P2P mesh сети

## 🎯 **Требуемая функциональность**

### **1. Мультитунельность и изоляция:**

#### **Архитектура изоляции:**
```
┌─────────────────────────────────────────────────────────────┐
│                    CloudBridge Relay                        │
├─────────────────────────────────────────────────────────────┤
│  Tenant A (tenant-216420165)                               │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   Client A1     │◄──►│   Client A2     │                │
│  │  (Office A1)    │    │  (Office A2)    │                │
│  └─────────────────┘    └─────────────────┘                │
├─────────────────────────────────────────────────────────────┤
│  Tenant B (tenant-123456789)                               │
│  ┌─────────────────┐    ┌─────────────────┐                │
│  │   Client B1     │◄──►│   Client B2     │                │
│  │  (Office B1)    │    │  (Office B2)    │                │
│  └─────────────────┘    └─────────────────┘                │
└─────────────────────────────────────────────────────────────┘
```

#### **Изоляция по уровням:**
- **JWT токены** содержат `tenant_id` для идентификации
- **Сетевые пространства** изолированы по tenant
- **P2P соединения** только внутри одного tenant
- **Маршрутизация** на основе tenant_id

### **2. P2P Mesh сеть:**

#### **Требуемый функционал:**
```go
// Пример структуры для P2P соединений
type P2PConnection struct {
    ConnectionID    string    `json:"connection_id"`
    TenantID        string    `json:"tenant_id"`
    SourcePeer      string    `json:"source_peer"`
    TargetPeer      string    `json:"target_peer"`
    ConnectionType  string    `json:"connection_type"` // "quic", "quic-ice"
    Status          string    `json:"status"`          // "connecting", "established", "disconnected"
    CreatedAt       time.Time `json:"created_at"`
    LastActivity    time.Time `json:"last_activity"`
    QUICConn        *quic.Conn `json:"-"`
    ICEConn         *ice.Conn  `json:"-"`
}

type Peer struct {
    PeerID          string    `json:"peer_id"`
    TenantID        string    `json:"tenant_id"`
    PublicKey       string    `json:"public_key"`
    AllowedIPs      []string  `json:"allowed_ips"`
    Status          string    `json:"status"`          // "online", "offline"
    LastSeen        time.Time `json:"last_seen"`
    HeartbeatTicker *time.Ticker `json:"-"`
    Connections     map[string]*P2PConnection `json:"-"`
}
```

### **3. Heartbeat/Keep-alive система:**

```go
// Heartbeat механизм
func (p *Peer) StartHeartbeat() {
    p.HeartbeatTicker = time.NewTicker(30 * time.Second)
    go func() {
        for range p.HeartbeatTicker.C {
            // Отправляем heartbeat в relay
            p.SendHeartbeat()
        }
    }()
}

func (p *Peer) SendHeartbeat() {
    // POST /api/v1/tenants/{tenant_id}/peers/{peer_id}/heartbeat
    // Обновляем LastSeen и Status
}
```

## 🛠️ **План доработки**

### **Этап 1: P2P маршрутизация (Приоритет: ВЫСОКИЙ)**

#### **1.1 Реализовать P2P соединения:**
```go
// В cmd/relay/main.go добавить:
func (s *Server) handleP2PConnection(conn *quic.Conn, peerID string) {
    // 1. Определяем tenant по peerID
    tenantID := s.getTenantByPeerID(peerID)
    
    // 2. Находим все активные соединения в этом tenant
    activeConnections := s.getActiveConnections(tenantID)
    
    // 3. Создаем P2P маршрутизацию
    for _, targetConn := range activeConnections {
        if targetConn.PeerID != peerID {
            s.establishP2PConnection(conn, targetConn)
        }
    }
}
```

#### **1.2 Реализовать QUIC relay:**
```go
func (s *Server) relayQUICPacket(sourceConn *quic.Conn, targetPeerID string, data []byte) {
    // 1. Находим целевое соединение
    targetConn := s.findConnectionByPeerID(targetPeerID)
    
    // 2. Пересылаем пакет
    if targetConn != nil {
        targetConn.Write(data)
    }
}
```

### **Этап 2: Heartbeat система (Приоритет: ВЫСОКИЙ)**

#### **2.1 Добавить heartbeat endpoint:**
```go
// В internal/api/peer_handlers.go:
func (h *PeerHandler) HandleHeartbeat(c *gin.Context) {
    tenantID := c.Param("tenant_id")
    peerID := c.Param("peer_id")
    
    // Обновляем LastSeen и Status
    peer := h.peerManager.GetPeer(tenantID, peerID)
    if peer != nil {
        peer.LastSeen = time.Now()
        peer.Status = "online"
    }
    
    c.JSON(200, gin.H{"status": "ok"})
}
```

#### **2.2 Реализовать cleanup офлайн пиров:**
```go
func (s *Server) cleanupOfflinePeers() {
    ticker := time.NewTicker(60 * time.Second)
    go func() {
        for range ticker.C {
            s.peerManager.CleanupOfflinePeers(5 * time.Minute)
        }
    }()
}
```

### **Этап 3: ICE/STUN интеграция (Приоритет: СРЕДНИЙ)**

#### **3.1 Реализовать ICE кандидаты:**
```go
type ICECandidate struct {
    Candidate     string `json:"candidate"`
    SDPMLineIndex int    `json:"sdpMLineIndex"`
    SDPMid        string `json:"sdpMid"`
}

func (s *Server) exchangeICECandidates(sourcePeer, targetPeer string, candidates []ICECandidate) {
    // Обмениваем ICE кандидаты между пирами
}
```

### **Этап 4: Мультитунельность (Приоритет: СРЕДНИЙ)**

#### **4.1 Изоляция по tenant:**
```go
type TenantManager struct {
    tenants map[string]*Tenant
    mutex   sync.RWMutex
}

type Tenant struct {
    TenantID    string
    Peers       map[string]*Peer
    Connections map[string]*P2PConnection
    NetworkCIDR string
}
```

## 🐳 **Инструкция по сборке Minikube**

### **1. Установка Minikube:**

```bash
# Установка Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Установка kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Запуск Minikube
minikube start --driver=docker --memory=4096 --cpus=2
```

### **2. Сборка и загрузка Docker образа:**

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Собрать Docker образ с новым тегом
docker build -t cloudbridge-relay:quic-ice-v2 .

# Загрузить образ в Minikube
minikube image load cloudbridge-relay:quic-ice-v2

# Проверить загруженные образы
minikube image ls | grep cloudbridge-relay
```

### **3. Применение Kubernetes конфигураций:**

```bash
# Создать namespace
kubectl create namespace cloudbridge

# Применить конфигурации
kubectl apply -f k8s/deployment.yaml -n cloudbridge
kubectl apply -f k8s/cloudbridge-relay-external-service.yaml -n cloudbridge

# Проверить статус
kubectl get pods -n cloudbridge
kubectl get services -n cloudbridge
```

### **4. Проверка работы:**

```bash
# Проверить логи
kubectl logs -f deployment/cloudbridge-relay -n cloudbridge

# Проверить порты
kubectl port-forward service/cloudbridge-relay-external 8080:8080 -n cloudbridge

# Тестировать API
curl http://localhost:8080/health
```

## 📊 **Метрики успеха**

### **Критерии готовности:**

1. **P2P соединения:**
   - ✅ Клиенты могут общаться друг с другом через relay
   - ✅ QUIC пакеты пересылаются между пирами
   - ✅ Latency < 100ms для локальных соединений

2. **Мультитунельность:**
   - ✅ Пиры из разных tenant не могут общаться
   - ✅ Изоляция на уровне JWT токенов
   - ✅ Отдельные сетевые пространства

3. **Стабильность:**
   - ✅ Heartbeat система работает
   - ✅ Автоматический cleanup офлайн пиров
   - ✅ Graceful shutdown соединений

4. **Масштабируемость:**
   - ✅ Поддержка 100+ одновременных соединений
   - ✅ Memory usage < 500MB
   - ✅ CPU usage < 50%

## 🚀 **Следующие шаги**

### **Для разработчика:**

1. **Немедленно (сегодня):**
   - Изучить текущий код в `cmd/relay/`
   - Понять структуру JWT токенов
   - Проанализировать существующие API endpoints

2. **Этап 1 (1-2 дня):**
   - Реализовать P2P маршрутизацию
   - Добавить heartbeat систему
   - Протестировать базовую P2P связность

3. **Этап 2 (2-3 дня):**
   - Реализовать мультитунельность
   - Добавить ICE/STUN интеграцию
   - Протестировать изоляцию

4. **Этап 3 (1 день):**
   - Оптимизация производительности
   - Добавление метрик
   - Финальное тестирование

### **Для DevOps:**

1. **Подготовить тестовую среду:**
   - Настроить Minikube
   - Подготовить тестовые JWT токены
   - Настроить мониторинг

2. **Тестирование:**
   - Тестировать P2P соединения между разными серверами
   - Проверить изоляцию между tenant
   - Нагрузочное тестирование

## 📞 **Контакты и поддержка**

- **Техническая документация:** `PROJECT_OVERVIEW.md`
- **Быстрые исправления:** `QUICK_FIX_INSTRUCTIONS.md`
- **DevOps задачи:** `DEVOPS_TASK.md`
- **Текущий статус:** Все порты работают, нужна доработка P2P маршрутизации

---

**🎯 Цель: Создать полноценную P2P mesh сеть, где удаленные сервера клиентов могут общаться как в локальной сети, с полной изоляцией между разными пользователями!**
