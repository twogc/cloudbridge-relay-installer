# 🚀 План развертывания QUIC + ICE/STUN/TURN на продакшене

## 📊 **Текущая ситуация:**

### ✅ **Что готово:**
- **Клиент:** Готов к работе с новой архитектурой
- **JWT токен:** Содержит все необходимые конфигурации
- **DNS:** `edge.2gc.ru` → `95.163.250.190` ✅
- **STUN порт 19302:** Открыт, но сервис не отвечает

### ❌ **Что нужно исправить:**
- **HTTP API (8080):** Connection refused
- **P2P API (8082):** Connection refused  
- **STUN сервер (19302):** Не отвечает на запросы
- **QUIC transport (9090):** Не запущен
- **MASQUE proxy (8443):** Не запущен
- **Enhanced QUIC (9092):** Не запущен

---

## 🎯 **План развертывания:**

### **Этап 1: Подготовка сервера**

#### 1.1 **Проверить текущее состояние:**
```bash
# Подключиться к серверу
ssh root@95.163.250.190

# Проверить запущенные сервисы
systemctl status cloudbridge-relay
docker ps -a | grep relay

# Проверить открытые порты
netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)"

# Проверить логи
journalctl -u cloudbridge-relay -f
```

#### 1.2 **Остановить старые сервисы:**
```bash
# Остановить старый relay
systemctl stop cloudbridge-relay
docker stop $(docker ps -q --filter "name=relay")
```

#### 1.3 **Обновить код:**
```bash
# Перейти в директорию проекта
cd /path/to/cloudbridge-relay-installer

# Получить последние изменения
git pull origin main

# Проверить коммит
git log --oneline -1
# Должен быть: 341fa1f6 🚀 Implement QUIC + ICE/STUN/TURN P2P architecture
```

### **Этап 2: Настройка окружения**

#### 2.1 **Установить зависимости:**
```bash
# Обновить Go до версии 1.25
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Проверить версию
go version
# Должно быть: go version go1.25 linux/amd64
```

#### 2.2 **Настроить JWT секрет:**
```bash
# Создать или обновить .env файл
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

### **Этап 3: Сборка и развертывание**

#### 3.1 **Собрать приложение:**
```bash
# Собрать бинарник
go build -o cloudbridge-relay ./cmd/relay

# Проверить размер и права
ls -la cloudbridge-relay
chmod +x cloudbridge-relay
```

#### 3.2 **Настроить systemd сервис:**
```bash
# Создать systemd unit файл
sudo cat > /etc/systemd/system/cloudbridge-relay.service << EOF
[Unit]
Description=CloudBridge Relay Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/path/to/cloudbridge-relay-installer
EnvironmentFile=/path/to/cloudbridge-relay-installer/.env
ExecStart=/path/to/cloudbridge-relay-installer/cloudbridge-relay
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Перезагрузить systemd
sudo systemctl daemon-reload
sudo systemctl enable cloudbridge-relay
```

#### 3.3 **Настроить UFW firewall:**
```bash
# Очистить старые правила
sudo ufw --force reset

# Разрешить SSH
sudo ufw allow 22

# Разрешить Zabbix
sudo ufw allow 10050

# Разрешить новые порты
sudo ufw allow 8080/tcp  # HTTP API
sudo ufw allow 8082/tcp  # P2P API
sudo ufw allow 19302/udp # STUN Server
sudo ufw allow 9090/udp  # QUIC Transport
sudo ufw allow 8443/tcp  # MASQUE Proxy
sudo ufw allow 9092/udp  # Enhanced QUIC

# Включить firewall
sudo ufw --force enable

# Проверить статус
sudo ufw status
```

### **Этап 4: Запуск и тестирование**

#### 4.1 **Запустить сервис:**
```bash
# Запустить relay
sudo systemctl start cloudbridge-relay

# Проверить статус
sudo systemctl status cloudbridge-relay

# Проверить логи
journalctl -u cloudbridge-relay -f
```

#### 4.2 **Проверить порты:**
```bash
# Проверить, что все порты открыты
netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)"

# Ожидаемый результат:
# tcp  0  0 0.0.0.0:8080  0.0.0.0:*  LISTEN  [PID]/cloudbridge-relay
# tcp  0  0 0.0.0.0:8082  0.0.0.0:*  LISTEN  [PID]/cloudbridge-relay
# udp  0  0 0.0.0.0:19302 0.0.0.0:*  [PID]/cloudbridge-relay
# udp  0  0 0.0.0.0:9090  0.0.0.0:*  [PID]/cloudbridge-relay
# tcp  0  0 0.0.0.0:8443  0.0.0.0:*  LISTEN  [PID]/cloudbridge-relay
# udp  0  0 0.0.0.0:9092  0.0.0.0:*  [PID]/cloudbridge-relay
```

#### 4.3 **Тестирование endpoints:**
```bash
# Тест HTTP API
curl -k "https://edge.2gc.ru:8080/health"
# Ожидаемый результат: {"status":"ok","timestamp":"..."}

# Тест P2P API
curl -k "https://edge.2gc.ru:8082/health"
# Ожидаемый результат: {"status":"ok","timestamp":"..."}

# Тест STUN сервера
echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302
# Ожидаемый результат: STUN Binding Response
```

---

## 🔧 **Альтернативный вариант: Docker развертывание**

### **Docker Compose:**
```yaml
# docker-compose.yml
version: '3.8'
services:
  cloudbridge-relay:
    build: .
    ports:
      - "8080:8080"   # HTTP API
      - "8082:8082"   # P2P API
      - "19302:19302/udp" # STUN Server
      - "9090:9090/udp"   # QUIC Transport
      - "8443:8443"   # MASQUE Proxy
      - "9092:9092/udp"   # Enhanced QUIC
    environment:
      - RELAY_JWT_SECRET=fallback-key
    restart: unless-stopped
```

### **Команды для Docker:**
```bash
# Собрать и запустить
docker-compose up -d --build

# Проверить статус
docker-compose ps

# Проверить логи
docker-compose logs -f
```

---

## 📋 **Чек-лист развертывания:**

### **✅ Подготовка:**
- [ ] Подключиться к серверу
- [ ] Остановить старые сервисы
- [ ] Обновить код до коммита 341fa1f6
- [ ] Установить Go 1.25
- [ ] Настроить JWT секрет

### **✅ Сборка:**
- [ ] Собрать приложение
- [ ] Настроить systemd сервис
- [ ] Настроить UFW firewall

### **✅ Запуск:**
- [ ] Запустить сервис
- [ ] Проверить статус
- [ ] Проверить порты
- [ ] Проверить логи

### **✅ Тестирование:**
- [ ] HTTP API (8080) - health check
- [ ] P2P API (8082) - health check
- [ ] STUN сервер (19302) - binding request
- [ ] QUIC transport (9090) - connection test
- [ ] MASQUE proxy (8443) - HTTP/3 test
- [ ] Enhanced QUIC (9092) - connection test

---

## 🎯 **Ожидаемый результат:**

После успешного развертывания клиент должен получить:

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

**И клиент сможет успешно:**
1. ✅ Подключиться к HTTP API для регистрации
2. ✅ Получить STUN ответы для NAT traversal
3. ✅ Установить QUIC соединения для P2P mesh
4. ✅ Использовать MASQUE для обхода корпоративных прокси

---

## 🚨 **Важные замечания:**

1. **Бэкап:** Сделать бэкап текущей конфигурации перед обновлением
2. **Мониторинг:** Настроить мониторинг новых сервисов
3. **Логи:** Проверить, что логи пишутся корректно
4. **Производительность:** Мониторить нагрузку на новые порты
5. **Безопасность:** Убедиться, что JWT секрет защищен

---

## 📞 **Поддержка:**

При возникновении проблем:
1. Проверить логи: `journalctl -u cloudbridge-relay -f`
2. Проверить порты: `netstat -tlnp`
3. Проверить firewall: `sudo ufw status`
4. Проверить статус: `systemctl status cloudbridge-relay`
