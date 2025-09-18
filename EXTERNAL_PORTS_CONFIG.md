# 🌐 Конфигурация внешних портов для QUIC + ICE/STUN/TURN

## 🎯 **Порты, которые нужно открыть наружу:**

### **✅ Обязательные порты:**

| Порт | Протокол | Назначение | Приоритет |
|------|----------|------------|-----------|
| **8080** | TCP | HTTP API (Health, Registration) | 🔴 Критический |
| **8082** | TCP | **P2P API (ICE signaling)** | 🔴 Критический |
| **9090** | UDP | **QUIC Transport** | 🔴 Критический |
| **19302** | UDP | **STUN Server** | 🔴 Критический |

### **✅ Рекомендуемые порты:**

| Порт | Протокол | Назначение | Приоритет |
|------|----------|------------|-----------|
| **8443** | TCP | HTTPS/MASQUE (Corporate proxy) | 🟡 Важный |
| **9092** | UDP | Enhanced QUIC (Advanced features) | 🟡 Важный |

## 🔧 **Обновленная конфигурация Kubernetes:**

### **Service Configuration:**
```yaml
apiVersion: v1
kind: Service
metadata:
  name: cloudbridge-relay-service
  namespace: cloudbridge-relay
spec:
  type: LoadBalancer
  ports:
  # HTTP API - Health, Registration
  - port: 8080
    targetPort: 8080
    protocol: TCP
    name: http-api
  
  # P2P API - ICE signaling, P2P management
  - port: 8082
    targetPort: 8082
    protocol: TCP
    name: p2p-api
  
  # QUIC Transport - Main P2P data transport
  - port: 9090
    targetPort: 9090
    protocol: UDP
    name: quic-transport
  
  # STUN Server - NAT discovery
  - port: 19302
    targetPort: 19302
    protocol: UDP
    name: stun-server
  
  # HTTPS/MASQUE - Corporate proxy support
  - port: 8443
    targetPort: 8443
    protocol: TCP
    name: https-masque
  
  # Enhanced QUIC - Advanced features
  - port: 9092
    targetPort: 9092
    protocol: UDP
    name: enhanced-quic
```

## 🚀 **Порты для разных сценариев:**

### **1. Минимальная конфигурация (только P2P):**
```bash
# Открыть только критически важные порты
8080/tcp  # HTTP API
8082/tcp  # P2P API
9090/udp  # QUIC Transport
19302/udp # STUN Server
```

### **2. Полная конфигурация (все возможности):**
```bash
# Открыть все порты для максимальной функциональности
8080/tcp  # HTTP API
8082/tcp  # P2P API
8443/tcp  # HTTPS/MASQUE
9090/udp  # QUIC Transport
9092/udp  # Enhanced QUIC
19302/udp # STUN Server
```

### **3. Корпоративная конфигурация (через прокси):**
```bash
# Только HTTP/HTTPS порты для корпоративных сетей
8080/tcp  # HTTP API
8082/tcp  # P2P API
8443/tcp  # HTTPS/MASQUE (основной транспорт)
```

## 🔒 **Безопасность портов:**

### **Публичные порты (доступны всем):**
- **8080/tcp** - HTTP API (только health, регистрация)
- **19302/udp** - STUN Server (стандартный порт)

### **Аутентифицированные порты (требуют JWT):**
- **8082/tcp** - P2P API (ICE signaling)
- **9090/udp** - QUIC Transport
- **9092/udp** - Enhanced QUIC
- **8443/tcp** - HTTPS/MASQUE

## 🌐 **DNS и Load Balancer:**

### **Рекомендуемые DNS записи:**
```
relay.cloudbridge.com     A    <LOAD_BALANCER_IP>
p2p.cloudbridge.com       A    <LOAD_BALANCER_IP>
stun.cloudbridge.com      A    <LOAD_BALANCER_IP>
```

### **Load Balancer конфигурация:**
```yaml
# AWS ALB/NLB, GCP Load Balancer, Azure Load Balancer
ports:
  - 8080/tcp   # HTTP API
  - 8082/tcp   # P2P API
  - 8443/tcp   # HTTPS/MASQUE
  - 9090/udp   # QUIC Transport
  - 9092/udp   # Enhanced QUIC
  - 19302/udp  # STUN Server
```

## 🎯 **Итоговая рекомендация:**

### **Для продакшена открыть:**
1. **8080/tcp** - HTTP API (Health, Registration)
2. **8082/tcp** - P2P API (ICE signaling)
3. **9090/udp** - QUIC Transport (основной P2P)
4. **19302/udp** - STUN Server (NAT discovery)
5. **8443/tcp** - HTTPS/MASQUE (корпоративные сети)

### **Опционально:**
6. **9092/udp** - Enhanced QUIC (продвинутые функции)

**Клиенты подключаются outbound - никаких входящих портов не требуется!** 🎉
