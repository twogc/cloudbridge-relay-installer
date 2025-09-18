# 🚀 Архитектура портов: QUIC + ICE/STUN/TURN

## 📊 Текущая конфигурация портов

### 🔧 **Основные порты сервера:**

| Порт | Протокол | Назначение | Статус |
|------|----------|------------|--------|
| **8080** | TCP | HTTP API (Health, Metrics) | ✅ Активен |
| **8082** | TCP | P2P API (ICE signaling, P2P connect) | ✅ Активен |
| **8443** | TCP | HTTPS/TLS (MASQUE CONNECT-UDP) | ✅ Активен |
| **9090** | UDP | QUIC транспорт | ✅ Активен |
| **9091** | TCP | HTTP API (внутренний) | ✅ Активен |
| **9092** | UDP | Enhanced QUIC | ✅ Активен |
| **19302** | UDP | STUN сервер | ✅ Активен |

### 🌐 **Внешние STUN серверы:**
```
stun.l.google.com:19302
stun1.l.google.com:19302
```

## 🔄 **Архитектура соединений**

### **1. P2P Connection Flow:**
```
Client A ←→ QUIC:9090 ←→ Relay ←→ QUIC:9090 ←→ Client B
```

### **2. ICE Signaling Flow:**
```
Client A ←→ HTTP:8082 ←→ Relay ←→ HTTP:8082 ←→ Client B
```

### **3. STUN Discovery Flow:**
```
Client A ←→ UDP:19302 ←→ STUN Server ←→ UDP:19302 ←→ Client A
```

### **4. MASQUE Tunneling Flow:**
```
Client A ←→ HTTPS:8443 ←→ Relay ←→ UDP:9090 ←→ Client B
```

## 🎯 **Ключевые порты для P2P**

### **Порт 8082 - P2P API Server**
- **ICE signaling** - обмен кандидатами
- **P2P connection management** - установка соединений
- **Session management** - управление сессиями

### **Порт 9090 - QUIC Transport**
- **Основной транспорт** для P2P данных
- **Connection migration** - смена IP без разрыва
- **Multiplexing** - множественные потоки

### **Порт 19302 - STUN**
- **NAT discovery** - определение внешнего IP
- **Port mapping** - определение доступных портов
- **Connectivity checks** - проверка связности

### **Порт 8443 - MASQUE**
- **Corporate proxy** - обход корпоративных файрволов
- **UDP over HTTP/3** - туннелирование UDP
- **Fallback transport** - резервный транспорт

## 🔧 **Конфигурация в Kubernetes**

### **Service Ports:**
```yaml
ports:
- port: 8080    # HTTP API
- port: 8443    # HTTPS/MASQUE  
- port: 19302   # STUN
```

### **Container Ports:**
```yaml
ports:
- containerPort: 8080   # HTTP API
- containerPort: 8082   # P2P API (internal)
- containerPort: 8443   # HTTPS/MASQUE
- containerPort: 9090   # QUIC
- containerPort: 9091   # Internal API
- containerPort: 9092   # Enhanced QUIC
- containerPort: 19302  # STUN
```

## 🚀 **Преимущества новой архитектуры**

### **Без WireGuard (старая):**
```
Client ←→ WireGuard:51820 ←→ Relay ←→ WireGuard:51820 ←→ Client
```

### **С QUIC + ICE/STUN/TURN (новая):**
```
Client ←→ QUIC:9090 + ICE:8082 + STUN:19302 ←→ Relay ←→ QUIC:9090 + ICE:8082 + STUN:19302 ←→ Client
```

## 📋 **Порты для клиентов**

### **Клиент должен открыть:**
- ❌ **Никаких портов!** - все через outbound соединения

### **Клиент подключается к:**
- ✅ **8080** - Health check, регистрация
- ✅ **8082** - ICE signaling, P2P API
- ✅ **9090** - QUIC транспорт
- ✅ **19302** - STUN discovery
- ✅ **8443** - MASQUE (через корпоративные прокси)

## 🎯 **Итог**

**Основные порты для P2P функциональности:**
- 🚀 **8082** - ICE signaling и P2P API
- 🚀 **9090** - QUIC транспорт для данных
- 🚀 **19302** - STUN для NAT traversal
- 🚀 **8443** - MASQUE для корпоративных сетей

**Клиенты не открывают портов - все через outbound соединения!** 🎉

