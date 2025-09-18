# 🚀 Быстрое развертывание на продакшене

## 📋 **Проблема:**
Клиент готов к работе с новой QUIC + ICE/STUN/TURN архитектурой, но на продакшен сервере `edge.2gc.ru` еще не развернута новая архитектура.

## ✅ **Решение:**

### **Вариант 1: Автоматическое развертывание (рекомендуется)**

```bash
# Запустить скрипт автоматического развертывания
./deploy-to-production.sh

# Или указать конкретный IP сервера
./deploy-to-production.sh 95.163.250.190
```

**Что делает скрипт:**
1. ✅ Проверяет подключение к серверу
2. ✅ Останавливает старые сервисы
3. ✅ Обновляет код до коммита `341fa1f6`
4. ✅ Устанавливает Go 1.25
5. ✅ Настраивает окружение и JWT секрет
6. ✅ Собирает приложение
7. ✅ Настраивает systemd сервис
8. ✅ Настраивает UFW firewall
9. ✅ Запускает сервис
10. ✅ Тестирует endpoints

### **Вариант 2: Ручное развертывание**

```bash
# 1. Подключиться к серверу
ssh root@95.163.250.190

# 2. Остановить старые сервисы
systemctl stop cloudbridge-relay
docker stop $(docker ps -q --filter "name=relay")

# 3. Обновить код
cd /opt/cloudbridge-relay-installer
git pull origin main

# 4. Установить Go 1.25
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 5. Настроить окружение
cat > .env << EOF
RELAY_JWT_SECRET=fallback-key
RELAY_HTTP_PORT=8080
RELAY_P2P_PORT=8082
RELAY_QUIC_PORT=9090
RELAY_STUN_PORT=19302
RELAY_MASQUE_PORT=8443
RELAY_ENHANCED_QUIC_PORT=9092
EOF

# 6. Собрать приложение
go build -o cloudbridge-relay ./cmd/relay
chmod +x cloudbridge-relay

# 7. Настроить systemd
sudo cat > /etc/systemd/system/cloudbridge-relay.service << EOF
[Unit]
Description=CloudBridge Relay Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/cloudbridge-relay-installer
EnvironmentFile=/opt/cloudbridge-relay-installer/.env
ExecStart=/opt/cloudbridge-relay-installer/cloudbridge-relay
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable cloudbridge-relay

# 8. Настроить UFW
sudo ufw --force reset
sudo ufw allow 22
sudo ufw allow 10050
sudo ufw allow 8080/tcp
sudo ufw allow 8082/tcp
sudo ufw allow 19302/udp
sudo ufw allow 9090/udp
sudo ufw allow 8443/tcp
sudo ufw allow 9092/udp
sudo ufw --force enable

# 9. Запустить сервис
sudo systemctl start cloudbridge-relay

# 10. Проверить статус
systemctl status cloudbridge-relay
netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)"
```

---

## 🧪 **Тестирование после развертывания:**

### **1. Проверить endpoints:**
```bash
# HTTP API
curl -k "https://edge.2gc.ru:8080/health"

# P2P API  
curl -k "https://edge.2gc.ru:8082/health"

# STUN сервер
echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302
```

### **2. Проверить с клиента:**
```bash
# Тест подключения
curl -k -v "https://edge.2gc.ru:8080/health"
curl -k -v "https://edge.2gc.ru:8082/health"

# Тест STUN
nc -u -v edge.2gc.ru 19302
```

---

## 📊 **Ожидаемый результат:**

После успешного развертывания:

### **✅ Порты должны быть открыты:**
- `8080/tcp` - HTTP API ✅
- `8082/tcp` - P2P API ✅
- `19302/udp` - STUN Server ✅
- `9090/udp` - QUIC Transport ✅
- `8443/tcp` - MASQUE Proxy ✅
- `9092/udp` - Enhanced QUIC ✅

### **✅ Endpoints должны отвечать:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-20T...",
  "services": {
    "http_api": "running",
    "p2p_api": "running",
    "stun_server": "running",
    "quic_transport": "running",
    "masque_proxy": "running",
    "enhanced_quic": "running"
  }
}
```

### **✅ Клиент должен успешно:**
1. Подключиться к HTTP API для регистрации
2. Получить STUN ответы для NAT traversal
3. Установить QUIC соединения для P2P mesh
4. Использовать MASQUE для обхода корпоративных прокси

---

## 🚨 **Важные замечания:**

1. **JWT секрет:** Убедитесь, что `RELAY_JWT_SECRET=fallback-key` настроен
2. **Бэкап:** Сделайте бэкап текущей конфигурации перед обновлением
3. **Мониторинг:** Настройте мониторинг новых сервисов
4. **Логи:** Проверьте логи: `journalctl -u cloudbridge-relay -f`

---

## 📞 **Поддержка:**

При возникновении проблем:
1. Проверить логи: `journalctl -u cloudbridge-relay -f`
2. Проверить порты: `netstat -tlnp`
3. Проверить firewall: `sudo ufw status`
4. Проверить статус: `systemctl status cloudbridge-relay`

---

## 🎯 **Результат:**

После успешного развертывания клиент сможет работать с новой QUIC + ICE/STUN/TURN архитектурой и получать "почти LAN" опыт без открытия портов на клиентской стороне! 🚀
