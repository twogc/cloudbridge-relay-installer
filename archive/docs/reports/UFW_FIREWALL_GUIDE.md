# 🔥 **CloudBridge Relay - UFW Firewall Guide**

## **Обзор**

Этот документ описывает настройку UFW (Uncomplicated Firewall) для CloudBridge Relay с красивыми портами 5550-5555.

## **Быстрый старт**

### **1. Автоматическая настройка:**
```bash
# Добавить все правила CloudBridge Relay
sudo ./scripts/ufw-cloudbridge.sh add

# Показать текущие правила
sudo ./scripts/ufw-cloudbridge.sh show

# Удалить все правила CloudBridge Relay
sudo ./scripts/ufw-cloudbridge.sh remove
```

### **2. Ручная настройка:**
```bash
# TCP порты
sudo ufw allow 5550/tcp comment "CloudBridge Relay - Main TCP Server"
sudo ufw allow 5551/tcp comment "CloudBridge Relay - Metrics (Prometheus)"
sudo ufw allow 5552/tcp comment "CloudBridge Relay - P2P API"
sudo ufw allow 5553/tcp comment "CloudBridge Relay - HTTP API"
sudo ufw allow 5554/tcp comment "CloudBridge Relay - HTTPS API"

# UDP порты
sudo ufw allow 5555/udp comment "CloudBridge Relay - QUIC Transport"
sudo ufw allow 9090/udp comment "CloudBridge Relay - QUIC Legacy"
sudo ufw allow 9092/udp comment "CloudBridge Relay - Enhanced QUIC"
sudo ufw allow 19302/udp comment "CloudBridge Relay - STUN Server"
sudo ufw allow 8443/udp comment "CloudBridge Relay - MASQUE Protocol"
sudo ufw allow 5554/udp comment "CloudBridge Relay - UDP Transport"
```

##  **Порты CloudBridge Relay**

### **TCP порты:**
| Порт | Протокол | Назначение | Описание |
|------|----------|------------|----------|
| 5550 | TCP | Main Server | Основной TCP сервер CloudBridge Relay |
| 5551 | TCP | Metrics | Prometheus метрики |
| 5552 | TCP | P2P API | P2P API для управления |
| 5553 | TCP | HTTP API | HTTP API для клиентов |
| 5554 | TCP | HTTPS API | HTTPS API с TLS |

### **UDP порты:**
| Порт | Протокол | Назначение | Описание |
|------|----------|------------|----------|
| 5555 | UDP | QUIC Transport | Основной QUIC транспорт |
| 9090 | UDP | QUIC Legacy | Legacy QUIC поддержка |
| 9092 | UDP | Enhanced QUIC | Расширенный QUIC с AI |
| 19302 | UDP | STUN Server | STUN сервер для NAT traversal |
| 8443 | UDP | MASQUE | MASQUE протокол для туннелей |
| 5554 | UDP | UDP Transport | Дополнительный UDP транспорт |

## 🛠️ **Управление UFW**

### **Основные команды:**
```bash
# Проверить статус
sudo ufw status verbose

# Включить UFW
sudo ufw enable

# Отключить UFW
sudo ufw disable

# Сбросить все правила
sudo ufw --force reset
```

### **Управление правилами:**
```bash
# Добавить правило
sudo ufw allow 5550/tcp

# Удалить правило
sudo ufw delete allow 5550/tcp

# Добавить правило с комментарием
sudo ufw allow 5551/tcp comment "Metrics"

# Разрешить с определенного IP
sudo ufw allow from 192.168.1.100 to any port 5551
```

## **Диагностика**

### **Проверка открытых портов:**
```bash
# Проверить слушающие порты
sudo netstat -tlnp | grep -E "(555[0-5]|909[0-2]|19302|8443)"

# Проверить UDP порты
sudo netstat -ulnp | grep -E "(555[0-5]|909[0-2]|19302|8443)"

# Проверить с ss
sudo ss -tlnp | grep -E "(555[0-5]|909[0-2]|19302|8443)"
```

### **Тестирование подключений:**
```bash
# Тест TCP порта
telnet localhost 5553

# Тест UDP порта
nc -u localhost 5555

# Тест с nmap
nmap -p 5550-5555 localhost
```

## **Безопасность**

### **Рекомендации:**
1. **Ограничить доступ к метрикам** (порт 5551):
   ```bash
   # Только для внутренней сети
   sudo ufw delete allow 5551/tcp
   sudo ufw allow from 10.0.0.0/8 to any port 5551
   sudo ufw allow from 192.168.0.0/16 to any port 5551
   ```

2. **Ограничить доступ к P2P API** (порт 5552):
   ```bash
   # Только для администраторов
   sudo ufw delete allow 5552/tcp
   sudo ufw allow from 192.168.1.0/24 to any port 5552
   ```

3. **Логирование:**
   ```bash
   # Включить логирование
   sudo ufw logging on
   
   # Проверить логи
   sudo tail -f /var/log/ufw.log
   ```

## **Мониторинг**

### **Проверка трафика:**
```bash
# Мониторинг подключений
sudo netstat -an | grep -E "(555[0-5]|909[0-2]|19302|8443)"

# Статистика UFW
sudo ufw status numbered

# Логи UFW
sudo grep UFW /var/log/syslog | tail -20
```

## 🔄 **Автоматизация**

### **Скрипт управления:**
```bash
# Показать справку
./scripts/ufw-cloudbridge.sh help

# Добавить правила
sudo ./scripts/ufw-cloudbridge.sh add

# Удалить правила
sudo ./scripts/ufw-cloudbridge.sh remove

# Показать правила
sudo ./scripts/ufw-cloudbridge.sh show

# Полный статус
sudo ./scripts/ufw-cloudbridge.sh status
```

## **Устранение неполадок**

### **Частые проблемы:**

1. **UFW блокирует подключения:**
   ```bash
   # Проверить статус
   sudo ufw status
   
   # Временно отключить для тестирования
   sudo ufw disable
   ```

2. **Порты не открываются:**
   ```bash
   # Проверить, что приложение слушает
   sudo netstat -tlnp | grep 5550
   
   # Проверить правила UFW
   sudo ufw status | grep 5550
   ```

3. **Проблемы с IPv6:**
   ```bash
   # Отключить IPv6 правила
   sudo ufw delete allow 5550/tcp
   sudo ufw allow 5550/tcp
   ```

## 📝 **Чек-лист**

### **Настройка UFW:**
- [ ] UFW включен
- [ ] Правила для портов 5550-5555 добавлены
- [ ] UDP порты для QUIC настроены
- [ ] SSH доступ сохранен
- [ ] Логирование включено

### **Тестирование:**
- [ ] TCP порты отвечают
- [ ] UDP порты работают
- [ ] Метрики доступны
- [ ] API endpoints отвечают
- [ ] QUIC транспорт работает

### **Безопасность:**
- [ ] Доступ к метрикам ограничен
- [ ] P2P API защищен
- [ ] Логи мониторятся
- [ ] Правила протестированы

---

**Результат: UFW правильно настроен для CloudBridge Relay с портами 5550-5555!**
