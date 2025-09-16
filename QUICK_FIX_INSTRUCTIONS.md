# 🚨 БЫСТРОЕ ИСПРАВЛЕНИЕ: Docker предупреждения в Zabbix

## ⚡ **АВТОМАТИЧЕСКОЕ ИСПРАВЛЕНИЕ (1 команда):**

```bash
cd /home/ubuntu/cloudbridge-relay-installer && sudo ./scripts/fix_docker_warning_issue.sh
```

---

## 📋 **РУЧНОЕ ИСПРАВЛЕНИЕ (5 команд):**

### **1. Остановить Zabbix Agent:**
```bash
sudo pkill -f zabbix_agentd
```

### **2. Исправить скрипт сбора метрик:**
```bash
cd /home/ubuntu/cloudbridge-relay-installer
sudo sed -i 's/docker ps --format "{{.Names}}" 2>\/dev\/null | wc -l/docker ps --format "{{.Names}}" 2>\/dev\/null | grep -v "WARNING" | grep -v "Error" | wc -l/g' /usr/local/bin/ai_metrics_collector.sh
```

### **3. Исправить права Docker:**
```bash
sudo usermod -aG docker zabbix
```

### **4. Запустить Zabbix Agent:**
```bash
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &
```

### **5. Проверить результат:**
```bash
zabbix_agentd -t cloudbridge.docker.containers.running
```

---

## 📁 **ВАЖНЫЕ ФАЙЛЫ:**

- **Скрипт сбора метрик:** `/usr/local/bin/ai_metrics_collector.sh`
- **Рабочая копия:** `/home/ubuntu/cloudbridge-relay-installer/scripts/ai_metrics_collector.sh`
- **Конфигурация агента:** `/etc/zabbix/zabbix_agentd.conf`
- **Логи агента:** `/var/log/zabbix/zabbix_agentd.log`

---

## 🔧 **ПОЛНАЯ ИНСТРУКЦИЯ:**

См. файл: `docs/DEVOPS_TROUBLESHOOTING_GUIDE.md`

---

## ✅ **ОЖИДАЕМЫЙ РЕЗУЛЬТАТ:**

После исправления команда `zabbix_agentd -t cloudbridge.docker.containers.running` должна возвращать только число (например, `[t|2]`) без предупреждений.
