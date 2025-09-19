# Анализ портов для внешнего доступа P2P маршрутизации

## 🔍 Текущая конфигурация портов

### P2P Deployment порты:
- **8080/TCP** - HTTP API (внутренний)
- **8082/TCP** - P2P API (внутренний)
- **9090/UDP** - QUIC P2P (внутренний)
- **8443/TCP** - HTTPS (внутренний)
- **19302/UDP** - STUN (внутренний)

### Существующие сервисы:
- **p2p-api-service** (NodePort: 30085) - P2P API
- **p2p-api-loadbalancer** (LoadBalancer: 30502) - P2P API
- **cloudbridge-relay-loadbalancer** - основной LoadBalancer

## 🚀 Рекомендуемые порты для внешнего доступа

### 1. P2P API (HTTP) - КРИТИЧНО
**Порт**: 8082/TCP
**Назначение**: HTTP API для регистрации пиров, обнаружения, heartbeat
**Внешний доступ**: ✅ **УЖЕ ОТКРЫТ**
- NodePort: 30085
- LoadBalancer: 30502

### 2. QUIC P2P (UDP) - КРИТИЧНО
**Порт**: 9090/UDP
**Назначение**: P2P соединения через QUIC протокол
**Внешний доступ**: ⚠️ **ТРЕБУЕТ ОТКРЫТИЯ**
- Текущий статус: только внутренний доступ
- Рекомендация: открыть через NodePort или LoadBalancer

### 3. STUN (UDP) - ВАЖНО
**Порт**: 19302/UDP
**Назначение**: STUN сервер для NAT traversal
**Внешний доступ**: ⚠️ **ТРЕБУЕТ ОТКРЫТИЯ**
- Текущий статус: только внутренний доступ
- Рекомендация: открыть для ICE/STUN функциональности

### 4. HTTPS (TCP) - РЕКОМЕНДУЕТСЯ
**Порт**: 8443/TCP
**Назначение**: Безопасный HTTPS доступ
**Внешний доступ**: ⚠️ **ТРЕБУЕТ ОТКРЫТИЯ**
- Текущий статус: только внутренний доступ
- Рекомендация: открыть для безопасного доступа

### 5. HTTP API (TCP) - ОПЦИОНАЛЬНО
**Порт**: 8080/TCP
**Назначение**: Основной HTTP API
**Внешний доступ**: ⚠️ **ТРЕБУЕТ ОТКРЫТИЯ**
- Текущий статус: только внутренний доступ
- Рекомендация: открыть для мониторинга и управления

## 🔧 Конфигурация для внешнего доступа

### Создание NodePort сервиса для P2P
```yaml
apiVersion: v1
kind: Service
metadata:
  name: p2p-external-access
  namespace: cloudbridge
spec:
  type: NodePort
  ports:
  - name: p2p-api
    port: 8082
    targetPort: 8082
    nodePort: 30085
    protocol: TCP
  - name: quic-p2p
    port: 9090
    targetPort: 9090
    nodePort: 30090
    protocol: UDP
  - name: stun
    port: 19302
    targetPort: 19302
    nodePort: 30093
    protocol: UDP
  - name: https
    port: 8443
    targetPort: 8443
    nodePort: 30044
    protocol: TCP
  - name: http-api
    port: 8080
    targetPort: 8080
    nodePort: 30080
    protocol: TCP
  selector:
    app.kubernetes.io/name: cloudbridge-relay-p2p
    app.kubernetes.io/component: relay
```

### Создание LoadBalancer сервиса
```yaml
apiVersion: v1
kind: Service
metadata:
  name: p2p-external-loadbalancer
  namespace: cloudbridge
spec:
  type: LoadBalancer
  ports:
  - name: p2p-api
    port: 8082
    targetPort: 8082
    protocol: TCP
  - name: quic-p2p
    port: 9090
    targetPort: 9090
    protocol: UDP
  - name: stun
    port: 19302
    targetPort: 19302
    protocol: UDP
  - name: https
    port: 8443
    targetPort: 8443
    protocol: TCP
  - name: http-api
    port: 8080
    targetPort: 8080
    protocol: TCP
  selector:
    app.kubernetes.io/name: cloudbridge-relay-p2p
    app.kubernetes.io/component: relay
```

## 🌐 Внешний доступ через Minikube

### Текущий доступ:
```bash
# P2P API (уже работает)
curl http://192.168.49.2:30085/health

# Minikube IP
minikube ip  # 192.168.49.2
```

### После открытия всех портов:
```bash
# P2P API
curl http://192.168.49.2:30085/api/v1/tenants/...

# QUIC P2P (UDP)
# Клиенты подключаются к 192.168.49.2:30090

# STUN
# Клиенты подключаются к 192.168.49.2:30093

# HTTPS
curl https://192.168.49.2:30044/health

# HTTP API
curl http://192.168.49.2:30080/health
```

## 🔒 Безопасность

### Firewall правила:
```bash
# Разрешить входящие соединения
sudo ufw allow 30085/tcp  # P2P API
sudo ufw allow 30090/udp  # QUIC P2P
sudo ufw allow 30093/udp  # STUN
sudo ufw allow 30044/tcp  # HTTPS
sudo ufw allow 30080/tcp  # HTTP API
```

### Рекомендации по безопасности:
1. **P2P API (8082)** - защищен JWT токенами
2. **QUIC P2P (9090)** - использует TLS шифрование
3. **STUN (19302)** - публичный протокол, безопасен
4. **HTTPS (8443)** - TLS шифрование
5. **HTTP API (8080)** - рекомендуется использовать только для мониторинга

## 📊 Приоритеты открытия портов

### 🔴 КРИТИЧНО (открыть немедленно):
1. **QUIC P2P (9090/UDP)** - для P2P соединений
2. **STUN (19302/UDP)** - для NAT traversal

### 🟡 ВАЖНО (открыть в ближайшее время):
3. **HTTPS (8443/TCP)** - для безопасного доступа
4. **HTTP API (8080/TCP)** - для мониторинга

### 🟢 ОПЦИОНАЛЬНО (можно отложить):
5. **P2P API (8082/TCP)** - уже открыт

## 🎯 Следующие шаги

1. **Создать NodePort сервис** для QUIC P2P и STUN
2. **Настроить LoadBalancer** для всех портов
3. **Обновить firewall** правила
4. **Протестировать внешний доступ** к QUIC P2P
5. **Настроить мониторинг** внешних портов

## 💡 Рекомендации

- **QUIC P2P (9090/UDP)** - самый важный порт для P2P функциональности
- **STUN (19302/UDP)** - необходим для NAT traversal
- **HTTPS (8443/TCP)** - рекомендуется для продакшена
- **HTTP API (8080/TCP)** - полезен для мониторинга и отладки
