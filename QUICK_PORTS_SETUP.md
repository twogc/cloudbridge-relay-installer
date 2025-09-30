# CloudBridge Relay - Быстрая настройка портов

## 🚀 Команды для открытия портов (копировать и выполнить):

### 1. Критические порты (ОБЯЗАТЕЛЬНО):
```bash
sudo ufw allow 5552/tcp comment "P2P API"
sudo ufw allow 5553/udp comment "P2P QUIC"
sudo ufw allow 19302/udp comment "STUN"
```

### 2. Рекомендуемые порты (для полной функциональности):
```bash
sudo ufw allow 5551/tcp comment "Health Check"
sudo ufw allow 9090/udp comment "QUIC Transport"
sudo ufw allow 8443/tcp comment "MASQUE HTTP/3"
```

### 3. Дополнительные порты (опционально):
```bash
sudo ufw allow 51820/udp comment "WebRTC P2P"
```

## 📋 Полная команда (все порты сразу):
```bash
sudo ufw allow 5552/tcp comment "P2P API"
sudo ufw allow 5553/udp comment "P2P QUIC"
sudo ufw allow 19302/udp comment "STUN"
sudo ufw allow 5551/tcp comment "Health Check"
sudo ufw allow 9090/udp comment "QUIC Transport"
sudo ufw allow 8443/tcp comment "MASQUE HTTP/3"
sudo ufw allow 51820/udp comment "WebRTC P2P"
```

## 🔍 Проверка:
```bash
# Проверить статус
sudo ufw status

# Проверить открытые порты
sudo netstat -tlnp | grep -E "(5551|5552|5553|8443|9090|19302|51820)"
```

## ⚠️ Важно:
- **5552/TCP** - самый критический порт для P2P API
- **5553/UDP** - необходим для P2P соединений
- **19302/UDP** - критичен для NAT traversal
- **8443/TCP** - для HTTP/3 MASQUE прокси
- **5551/TCP** - для health check и админки
- **9090/UDP** - для метрик и дополнительного QUIC
- **51820/UDP** - для WebRTC P2P соединений
