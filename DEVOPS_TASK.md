# 🚀 Задание для DevOps: Исправление CloudBridge Relay

## 📋 **Суть проекта CloudBridge Relay**

### **🎯 Назначение:**
CloudBridge Relay - это P2P mesh сеть для создания "почти LAN" соединений между серверами клиентов через наш relay сервер. Проект позволяет клиентам подключаться друг к другу без открытия портов на их стороне, используя современные протоколы QUIC + ICE/STUN/TURN.

### **🔧 Архитектура:**
- **Relay сервер** (наш): Принимает входящие соединения от клиентов
- **Клиенты**: Подключаются к relay серверу и получают доступ к другим клиентам
- **P2P mesh**: Клиенты могут общаться напрямую через relay или через прямое соединение

### **🌐 Протоколы:**
- **QUIC**: Основной транспорт для P2P соединений
- **ICE/STUN/TURN**: NAT traversal для обхода файрволов
- **MASQUE**: HTTP/3 прокси для корпоративных сетей
- **WebSocket/TCP**: Fallback для обхода блокировок

---

## 🚨 **Диагностированная проблема:**

### **❌ Текущее состояние:**
1. **Сервис запускается, но с предупреждениями** - использует дефолтный JWT секрет
2. **Переменная окружения `RELAY_JWT_SECRET` не настроена** в .env файле
3. **Порт 8082 (P2P API) работает** - сервис успешно запускается
4. **UDP порты настроены** - STUN, QUIC transport доступны
5. **Клиент может подключиться** к relay серверу

### **✅ Что работает:**
- nginx работает на портах 80, 443
- Порт 8080 открыт (HTTP API)
- Порт 8082 открыт (P2P API) ✅
- Порт 8443 открыт (MASQUE)
- Порт 9090 открыт (QUIC)
- Сервис cloudbridge-relay запущен и функционирует

---

## 🎯 **Задание для DevOps:**

### **1. Исправить переменные окружения**

**Проблема:** В .env файле используется `JWT_SECRET`, но сервис ожидает `RELAY_JWT_SECRET`

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Добавить недостающую переменную в .env файл
echo "RELAY_JWT_SECRET=fallback-key" >> .env

# Или создать полный .env файл с правильными переменными
cat > .env << EOF
# CloudBridge Relay Environment Variables
# Этот файл содержит секретные данные и НЕ должен попадать в git

# JWT Configuration (исправлено)
JWT_SECRET=1Pc6bvrTUDtXMKm6mMF6+7UJi/EYAdOcGPkXCRYV4xI=
RELAY_JWT_SECRET=fallback-key

# API Configuration
API_KEY=cb-api-key-1757634513

# Database Configuration
CASSANDRA_USERNAME=cassandra
CASSANDRA_PASSWORD=cb-cassandra-1757634513

# WireGuard Configuration
WIREGUARD_PRIVATE_KEY=iB7ZFl4fqYxgu818pHYqmytiU2fzReWZoJ/fdFSWXmI=
WIREGUARD_PUBLIC_KEY=LG2l0xUInTlYGJzIPeI6//t/SYlBkd+N6f6g5Hsr6Wc=

# Monitoring Configuration
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=cb-grafana-1757634513
PROMETHEUS_BASIC_AUTH=admin:cb-prometheus-1757634513

# Development Configuration
LOG_LEVEL=info
DEBUG=false

# Relay Service Ports
RELAY_HTTP_PORT=8080
RELAY_P2P_PORT=8082
RELAY_QUIC_PORT=9090
RELAY_STUN_PORT=19302
RELAY_MASQUE_PORT=8443
RELAY_ENHANCED_QUIC_PORT=9092
EOF
```

### **2. Перезапустить сервис с правильными переменными**

```bash
# Перейти в директорию проекта
cd /home/ubuntu/cloudbridge-relay-installer

# Остановить текущие процессы (если нужно)
sudo pkill -f cloudbridge-relay

# Загрузить переменные из .env файла
source .env

# Проверить, что переменная установлена
echo "RELAY_JWT_SECRET: $RELAY_JWT_SECRET"

# Запустить сервис с правильными переменными
./cloudbridge-relay --config configs/config.json &
```

### **3. Проверить, что сервис запустился без предупреждений**

```bash
# Проверить логи запуска
sleep 5
ps aux | grep cloudbridge-relay

# Проверить, что нет предупреждений о JWT_SECRET
journalctl -u cloudbridge-relay --since "1 minute ago" | grep -i "jwt_secret\|warning"
```

### **4. Проверить, что порты открыты**

```bash
# Проверить TCP порты
sudo netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)"

# Проверить UDP порты
sudo netstat -ulnp | grep -E ":(19302|9090|9092)"

# Ожидаемый результат (порты уже работают через nginx/haproxy):
# tcp  0  0 0.0.0.0:8080  0.0.0.0:*  LISTEN  [PID]/docker-proxy
# tcp  0  0 0.0.0.0:8082  0.0.0.0:*  LISTEN  [PID]/cloudbridge-relay
# tcp  0  0 0.0.0.0:8443  0.0.0.0:*  LISTEN  [PID]/docker-proxy
# tcp  0  0 0.0.0.0:9090  0.0.0.0:*  LISTEN  [PID]/haproxy
```

### **5. Настроить systemd сервис (опционально)**

```bash
# Создать systemd unit файл
sudo cat > /etc/systemd/system/cloudbridge-relay.service << EOF
[Unit]
Description=CloudBridge Relay Server
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/cloudbridge-relay-installer
EnvironmentFile=/home/ubuntu/cloudbridge-relay-installer/.env
ExecStart=/home/ubuntu/cloudbridge-relay-installer/cloudbridge-relay --config configs/config.json
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

# Перезагрузить systemd
sudo systemctl daemon-reload
sudo systemctl enable cloudbridge-relay
sudo systemctl start cloudbridge-relay

# Проверить статус
sudo systemctl status cloudbridge-relay
```

**Примечание:** Сервис уже запущен и работает. Systemd настройка нужна только для автоматического запуска при перезагрузке сервера.

### **6. Настроить UFW firewall (если нужно)**

```bash
# Проверить текущий статус firewall
sudo ufw status

# Если firewall не настроен, добавить правила:
sudo ufw allow 22        # SSH
sudo ufw allow 10050     # Zabbix
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 8080/tcp  # HTTP API
sudo ufw allow 8082/tcp  # P2P API
sudo ufw allow 19302/udp # STUN Server
sudo ufw allow 9090/udp  # QUIC Transport
sudo ufw allow 8443/tcp  # MASQUE Proxy
sudo ufw allow 9092/udp  # Enhanced QUIC

# Включить firewall (если не включен)
sudo ufw --force enable

# Проверить статус
sudo ufw status
```

**Примечание:** Firewall может быть уже настроен. Проверьте текущий статус перед внесением изменений.

### **7. Тестирование endpoints**

```bash
# Тест HTTP API
curl -k "https://edge.2gc.ru:8080/health"

# Тест P2P API
curl -k "https://edge.2gc.ru:8082/health"

# Тест STUN сервера
echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302

# Тест QUIC порта
nc -u -w 3 edge.2gc.ru 9090 < /dev/null
```

---

## 📊 **Ожидаемый результат:**

### **✅ После исправления должно работать:**

1. **HTTP API (8080)**: `https://edge.2gc.ru:8080/health` → `{"status":"ok"}`
2. **P2P API (8082)**: `https://edge.2gc.ru:8082/health` → `{"status":"ok"}`
3. **STUN Server (19302)**: UDP ответы на STUN Binding Request
4. **QUIC Transport (9090)**: UDP порт открыт для QUIC соединений
5. **MASQUE Proxy (8443)**: HTTPS прокси для HTTP/3
6. **Enhanced QUIC (9092)**: UDP порт для расширенного QUIC

### **🎯 Клиент сможет:**
- Подключиться к HTTP API для регистрации
- Получить STUN ответы для NAT traversal
- Установить QUIC соединения для P2P mesh
- Использовать MASQUE для обхода корпоративных прокси

---

## 🚨 **Важные замечания:**

1. **JWT секрет**: Используйте `fallback-key` для совместимости с клиентом
2. **Права доступа**: Убедитесь, что пользователь `ubuntu` имеет права на запуск
3. **Логи**: Проверяйте логи: `journalctl -u cloudbridge-relay -f`
4. **Мониторинг**: Настройте мониторинг новых портов
5. **Безопасность**: Убедитесь, что все порты защищены

---

## 📞 **Поддержка:**

При возникновении проблем:
1. Проверить логи: `journalctl -u cloudbridge-relay -f`
2. Проверить порты: `netstat -tlnp`
3. Проверить firewall: `sudo ufw status`
4. Проверить статус: `systemctl status cloudbridge-relay`

---

## 🎉 **Результат:**

После успешного исправления клиент сможет подключаться к relay серверу и использовать P2P mesh сеть для создания "почти LAN" соединений без открытия портов на клиентской стороне!

**Это позволит клиентам получать "почти LAN" опыт на L3 уровне через наш relay сервер!** 🚀
