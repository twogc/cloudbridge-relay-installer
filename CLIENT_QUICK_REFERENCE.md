# 🚀 Краткая справка для разработчика клиента

## 🎯 **Основные изменения:**

### **❌ Удалить:**
- WireGuard компоненты
- Порты 51820/udp
- Kernel интерфейсы

### **✅ Добавить:**
- QUIC + ICE/STUN/TURN
- Новые порты и endpoints

---

## 🌐 **DNS endpoints для edge.2gc.ru:**

```
edge.2gc.ru:8080   # HTTP API (Health, Registration)
edge.2gc.ru:8082   # P2P API (ICE signaling)
edge.2gc.ru:9090   # QUIC Transport (Main P2P)
edge.2gc.ru:19302  # STUN Server (NAT discovery)
edge.2gc.ru:8443   # HTTPS/MASQUE (Corporate proxy)
edge.2gc.ru:9092   # Enhanced QUIC (Advanced features)
```

---

## 📡 **Ключевые API запросы:**

### **1. Регистрация:**
```http
POST http://edge.2gc.ru:8080/api/v1/peers/register
```

### **2. ICE Signaling:**
```http
POST http://edge.2gc.ru:8082/api/v1/ice/candidates
```

### **3. P2P Connection:**
```http
POST http://edge.2gc.ru:8082/api/v1/p2p/connect
```

---

## 🔧 **Новые зависимости:**

```go
require (
    github.com/quic-go/quic-go v0.40.0
    github.com/pion/ice/v2 v2.3.0
    github.com/pion/stun v0.6.0
    github.com/pion/turn/v2 v2.1.0
)
```

---

## ⏱️ **Срок выполнения: 2-3 недели**

**Результат: P2P mesh network без открытия портов на клиентской стороне!** 🎉

