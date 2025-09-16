# 🔧 Решение проблем с Zabbix Agent

## 🚨 Проблема: "Unknown metric ai.health"

## 🚨 Проблема: "Value of type 'string' is not suitable for value type 'Numeric'"

### **Причина:**
Zabbix Agent не может получить доступ к Docker API из-за отсутствия прав доступа.

### **Решение:**

#### **1. Добавить пользователя zabbix в группу docker:**
```bash
sudo usermod -aG docker zabbix
```

#### **2. Проверить доступ к Docker:**
```bash
sudo -u zabbix docker ps
```

#### **3. Перезапустить Zabbix Agent:**
```bash
sudo pkill -f zabbix_agentd
sudo ./scripts/start_zabbix_agent.sh
```

#### **4. Автоматическое исправление:**
```bash
sudo ./scripts/fix_zabbix_docker_permissions.sh
```

## 🚨 Проблема: "Unknown metric ai.health"

### **Причина:**
Zabbix сервер не может получить данные от клиента из-за конфликта между Zabbix Agent и Zabbix Agent 2.

### **Решение:**

#### **1. Остановить Zabbix Agent 2:**
```bash
sudo systemctl stop zabbix-agent2
sudo systemctl disable zabbix-agent2
```

#### **2. Остановить все процессы Zabbix Agent:**
```bash
sudo pkill -f zabbix_agentd
```

#### **3. Запустить правильный Zabbix Agent:**
```bash
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &
```

#### **4. Проверить работу:**
```bash
# Проверить, что Agent слушает на порту 10050
sudo netstat -tlnp | grep 10050

# Тестировать AI метрики
zabbix_agentd -t ai.health
zabbix_agentd -t ai.status
```

## 🚀 Автоматическое решение

Используйте готовый скрипт:
```bash
sudo ./scripts/start_zabbix_agent.sh
```

## 📊 Проверка работы

### **1. Проверка порта:**
```bash
sudo netstat -tlnp | grep 10050
# Должно показать: zabbix_agen (не zabbix_agent2)
```

### **2. Тестирование метрик:**
```bash
zabbix_agentd -t ai.health
# Должно вернуть: [t|1] или [t|0]
```

### **3. Проверка логов:**
```bash
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

## 🔍 Диагностика проблем

### **Проблема: Порт 10050 занят**
```bash
# Найти процесс, занимающий порт
sudo lsof -i :10050

# Остановить процесс
sudo fuser -k 10050/tcp
```

### **Проблема: UserParameter не работает**
```bash
# Проверить права на скрипт
ls -la /usr/local/bin/ai_metrics_collector.sh

# Тестировать скрипт от имени пользователя zabbix
sudo -u zabbix /usr/local/bin/ai_metrics_collector.sh health
```

### **Проблема: AI сервис недоступен**
```bash
# Проверить доступность AI API
curl -s http://localhost:8083/api/v1/ai/status

# Проверить, что AI сервис запущен
kubectl get pods -n cloudbridge | grep ai
```

## 📋 Конфигурация

### **Правильная конфигурация Zabbix Agent:**
```bash
# /etc/zabbix/zabbix_agentd.conf
Server=10.0.0.241
ServerActive=10.0.0.241
Hostname=cloudbridge-relay-01

# AI Metrics UserParameters
UserParameter=ai.status,/usr/local/bin/ai_metrics_collector.sh status
UserParameter=ai.health,/usr/local/bin/ai_metrics_collector.sh health
# ... остальные метрики
```

## 🎯 Результат

После исправления:
- ✅ Zabbix Agent работает на порту 10050
- ✅ AI метрики доступны для Zabbix сервера
- ✅ Сбор данных работает корректно
- ✅ Нет конфликтов с Zabbix Agent 2

## 🔧 Полезные команды

```bash
# Запуск агента
sudo ./scripts/start_zabbix_agent.sh

# Остановка агента
sudo pkill -f zabbix_agentd

# Проверка статуса
sudo netstat -tlnp | grep 10050

# Тест метрик
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health

# Просмотр логов
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

