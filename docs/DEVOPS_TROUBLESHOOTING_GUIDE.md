# 🔧 DevOps: Полное руководство по устранению проблем с Zabbix

## 🚨 **ПРОБЛЕМА: Docker предупреждения в метриках Zabbix**

### **Ошибка:**
```
Value of type "string" is not suitable for value type "Numeric (unsigned)". 
Value "WARNING: Error loading config file: open /root/.docker/config.json: permission denied\n2"
```

---

## 📁 **СТРУКТУРА ФАЙЛОВ И ИХ РАСПОЛОЖЕНИЕ:**

### **1. Основные файлы проекта:**
```
/home/ubuntu/cloudbridge-relay-installer/
├── scripts/
│   ├── ai_metrics_collector.sh              # РАБОЧАЯ КОПИЯ для редактирования
│   ├── update_ai_metrics_script.sh          # Скрипт обновления
│   ├── start_zabbix_agent.sh                # Скрипт запуска агента
│   └── fix_zabbix_docker_permissions.sh     # Исправление прав Docker
├── docs/
│   ├── DEVOPS_TROUBLESHOOTING_GUIDE.md      # ЭТА ИНСТРУКЦИЯ
│   ├── DOCKER_ERROR_FIX_REPORT.md           # Отчет об исправлении
│   └── FILE_PERMISSIONS_GUIDE.md            # Руководство по правам
└── README_FILE_EDITING.md                   # Краткая инструкция
```

### **2. Системные файлы:**
```
/usr/local/bin/ai_metrics_collector.sh       # ИСПОЛНЯЕМЫЙ ФАЙЛ (root:root)
/etc/zabbix/zabbix_agentd.conf               # Конфигурация Zabbix Agent
/var/log/zabbix/zabbix_agentd.log            # Логи Zabbix Agent
/run/zabbix/zabbix_agentd.pid                # PID файл агента
```

---

## 🔍 **ДИАГНОСТИКА ПРОБЛЕМЫ:**

### **Шаг 1: Проверка текущего состояния**
```bash
# Перейти в рабочую директорию
cd /home/ubuntu/cloudbridge-relay-installer

# Проверить статус Zabbix Agent
sudo netstat -tlnp | grep 10050

# Проверить логи агента
sudo tail -20 /var/log/zabbix/zabbix_agentd.log

# Тестировать проблемную метрику
zabbix_agentd -t cloudbridge.docker.containers.running
```

### **Шаг 2: Проверка скрипта сбора метрик**
```bash
# Тестировать скрипт напрямую
/usr/local/bin/ai_metrics_collector.sh docker_containers_running

# Тестировать от имени пользователя zabbix
sudo -u zabbix /usr/local/bin/ai_metrics_collector.sh docker_containers_running
```

### **Шаг 3: Проверка Docker команд**
```bash
# Проверить Docker команды
sudo -u zabbix docker ps --format "{{.Names}}" 2>&1

# Проверить права на Docker socket
ls -la /var/run/docker.sock

# Проверить группы пользователя zabbix
groups zabbix
```

---

## 🛠️ **ПОШАГОВОЕ ИСПРАВЛЕНИЕ:**

### **Шаг 1: Остановка Zabbix Agent**
```bash
# Остановить все процессы Zabbix Agent
sudo pkill -f zabbix_agentd

# Удалить PID файл
sudo rm -f /run/zabbix/zabbix_agentd.pid

# Проверить, что порт свободен
sudo netstat -tlnp | grep 10050
```

### **Шаг 2: Исправление скрипта сбора метрик**
```bash
# Перейти в рабочую директорию
cd /home/ubuntu/cloudbridge-relay-installer

# Отредактировать рабочий файл
nano scripts/ai_metrics_collector.sh
# или
code scripts/ai_metrics_collector.sh
```

**Найти секцию Docker метрик и исправить:**
```bash
# НАЙТИ ЭТИ СТРОКИ:
"docker_containers_running")
    docker ps --format "{{.Names}}" 2>/dev/null | wc -l || echo "2"
    ;;
"docker_containers_total")
    docker ps -a --format "{{.Names}}" 2>/dev/null | wc -l || echo "2"
    ;;

# ЗАМЕНИТЬ НА:
"docker_containers_running")
    docker ps --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
    ;;
"docker_containers_total")
    docker ps -a --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
    ;;
```

### **Шаг 3: Обновление скрипта в системе**
```bash
# Запустить скрипт обновления
./scripts/update_ai_metrics_script.sh
```

### **Шаг 4: Исправление прав Docker**
```bash
# Добавить пользователя zabbix в группу docker
sudo usermod -aG docker zabbix

# Проверить права
sudo -u zabbix docker ps
```

### **Шаг 5: Очистка конфигурации Zabbix Agent**
```bash
# Проверить дублирующиеся UserParameter'ы
grep -n "cloudbridge.docker.containers.running" /etc/zabbix/zabbix_agentd.conf

# Удалить дублирующиеся записи (если есть)
sudo sed -i '/^UserParameter=cloudbridge\.docker\.containers\.running,\/usr\/local\/bin\/ai_metrics_collector\.sh docker_containers_running$/d' /etc/zabbix/zabbix_agentd.conf
sudo sed -i '/^UserParameter=cloudbridge\.docker\.containers\.total,\/usr\/local\/bin\/ai_metrics_collector\.sh docker_containers_total$/d' /etc/zabbix/zabbix_agentd.conf
```

### **Шаг 6: Запуск Zabbix Agent**
```bash
# Запустить агент в фоновом режиме
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &

# Проверить запуск
sleep 3
sudo netstat -tlnp | grep 10050
```

### **Шаг 7: Тестирование**
```bash
# Тестировать все метрики
zabbix_agentd -t cloudbridge.docker.containers.running
zabbix_agentd -t cloudbridge.docker.containers.total
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
```

---

## 🔧 **АЛЬТЕРНАТИВНЫЕ РЕШЕНИЯ:**

### **Решение 1: Полная переустановка скрипта**
```bash
# Создать новый скрипт с исправлениями
sudo tee /usr/local/bin/ai_metrics_collector.sh > /dev/null << 'EOF'
#!/bin/bash

AI_API_URL="http://localhost:8083"
METRIC_TYPE="$1"

case "$METRIC_TYPE" in
    "docker_containers_running")
        docker ps --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
        ;;
    "docker_containers_total")
        docker ps -a --format "{{.Names}}" 2>/dev/null | grep -v "WARNING" | wc -l || echo "2"
        ;;
    *)
        echo "0"
        ;;
esac
EOF

# Установить права
sudo chown root:root /usr/local/bin/ai_metrics_collector.sh
sudo chmod +x /usr/local/bin/ai_metrics_collector.sh
```

### **Решение 2: Использование альтернативных команд**
```bash
# Заменить Docker команды на альтернативные
sudo tee -a /etc/zabbix/zabbix_agentd.conf > /dev/null << 'EOF'

# Альтернативные Docker метрики
UserParameter=cloudbridge.docker.containers.running,ps aux | grep docker | grep -v grep | wc -l
UserParameter=cloudbridge.docker.containers.total,ls /var/lib/docker/containers/ 2>/dev/null | wc -l
EOF
```

---

## 📋 **ПОЛЕЗНЫЕ КОМАНДЫ:**

### **Диагностика:**
```bash
# Проверить статус всех сервисов
sudo systemctl status zabbix-agent
sudo systemctl status zabbix-agent2

# Проверить процессы
ps aux | grep zabbix

# Проверить порты
sudo netstat -tlnp | grep 10050

# Проверить логи
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

### **Управление Zabbix Agent:**
```bash
# Остановить агент
sudo pkill -f zabbix_agentd

# Запустить агент
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &

# Перезапустить агент
sudo pkill -f zabbix_agentd && sleep 2 && sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &
```

### **Тестирование метрик:**
```bash
# Тестировать конкретную метрику
zabbix_agentd -t METRIC_NAME

# Тестировать все AI метрики
for metric in ai.status ai.health ai.anomaly_detections ai.load_predictions; do
    echo "Testing $metric:"
    zabbix_agentd -t $metric
done

# Тестировать все Docker метрики
for metric in cloudbridge.docker.containers.running cloudbridge.docker.containers.total; do
    echo "Testing $metric:"
    zabbix_agentd -t $metric
done
```

---

## 🚨 **ЭКСТРЕННЫЕ ДЕЙСТВИЯ:**

### **Если ничего не помогает:**
```bash
# 1. Полная остановка всех Zabbix процессов
sudo pkill -f zabbix
sudo systemctl stop zabbix-agent
sudo systemctl stop zabbix-agent2

# 2. Очистка PID файлов
sudo rm -f /run/zabbix/*.pid

# 3. Пересоздание скрипта с нуля
sudo rm -f /usr/local/bin/ai_metrics_collector.sh
sudo cp /home/ubuntu/cloudbridge-relay-installer/scripts/ai_metrics_collector.sh /usr/local/bin/
sudo chown root:root /usr/local/bin/ai_metrics_collector.sh
sudo chmod +x /usr/local/bin/ai_metrics_collector.sh

# 4. Запуск агента
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -f &
```

---

## 📞 **КОНТАКТЫ И ДОПОЛНИТЕЛЬНАЯ ПОМОЩЬ:**

### **Файлы для изучения:**
- `docs/DOCKER_ERROR_FIX_REPORT.md` - Подробный отчет об исправлении
- `docs/FILE_PERMISSIONS_GUIDE.md` - Руководство по правам доступа
- `README_FILE_EDITING.md` - Краткая инструкция по редактированию

### **Логи для анализа:**
- `/var/log/zabbix/zabbix_agentd.log` - Логи Zabbix Agent
- `/var/log/syslog` - Системные логи
- `journalctl -u zabbix-agent` - Логи службы

---

## 🎯 **ОЖИДАЕМЫЙ РЕЗУЛЬТАТ:**

После выполнения всех шагов:
- ✅ Docker метрики возвращают только числовые значения
- ✅ Нет предупреждений в выводе команд
- ✅ Zabbix Agent работает стабильно
- ✅ Все метрики собираются без ошибок

**Время выполнения:** 10-15 минут  
**Сложность:** Средняя  
**Требуемые права:** sudo
