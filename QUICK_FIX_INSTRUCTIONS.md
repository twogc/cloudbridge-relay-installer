# ⚡ Быстрое исправление CloudBridge Relay

## 🚨 **Проблема:**
Сервис не запускается из-за отсутствия переменной окружения `RELAY_JWT_SECRET`

## ✅ **Решение (5 минут):**

### **1. Создать переменные окружения:**
```bash
cd /home/ubuntu/cloudbridge-relay-installer
cat > .env << EOF
RELAY_JWT_SECRET=fallback-key
RELAY_HTTP_PORT=8080
RELAY_P2P_PORT=8082
RELAY_QUIC_PORT=9090
RELAY_STUN_PORT=19302
RELAY_MASQUE_PORT=8443
RELAY_ENHANCED_QUIC_PORT=9092
EOF
```

### **2. Остановить старые процессы:**
```bash
sudo pkill -f cloudbridge-relay
```

### **3. Запустить сервис:**
```bash
export RELAY_JWT_SECRET=fallback-key
./cloudbridge-relay --config configs/config.json &
```

### **4. Проверить порты:**
```bash
sudo netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)"
```

### **5. Тестировать endpoints:**
```bash
# HTTP API
curl -k "https://edge.2gc.ru:8080/health"

# P2P API
curl -k "https://edge.2gc.ru:8082/health"

# STUN
echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302
```

## 🎯 **Ожидаемый результат:**
- Порт 8080: HTTP API ✅
- Порт 8082: P2P API ✅
- Порт 19302: STUN Server ✅
- Порт 9090: QUIC Transport ✅
- Порт 8443: MASQUE Proxy ✅
- Порт 9092: Enhanced QUIC ✅

## 🚀 **Клиент сможет подключиться!**
