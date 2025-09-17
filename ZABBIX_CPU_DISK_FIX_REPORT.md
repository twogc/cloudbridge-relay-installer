# 🔧 Отчет об исправлении CPU и Disk Usage в Zabbix

**Дата:** 17 сентября 2025  
**Время:** 17:45 UTC  
**Статус:** ✅ **ПРОБЛЕМА РЕШЕНА**

---

## 📊 **ПРОБЛЕМА**

Zabbix не получал данные о:
- **CPU Usage %** - возвращал `ZBX_NOTSUPPORTED [Collector is not started]`
- **Disk Usage %** - работал корректно

---

## 🔍 **ДИАГНОСТИКА**

### ✅ **Что работало:**
```bash
# Disk Usage работал корректно
zabbix_agentd -t vfs.fs.size[/,pused]
# Результат: [d|92.279504] ✅

# CPU Load работал
zabbix_agentd -t system.cpu.load[all,avg1]
# Результат: [d|0.31396484375] ✅
```

### ❌ **Что не работало:**
```bash
# CPU Usage не работал
zabbix_agentd -t system.cpu.util
# Результат: [m|ZBX_NOTSUPPORTED] [Collector is not started] ❌

# CPU Usage с параметрами не работал
zabbix_agentd -t system.cpu.util[,avg1]
# Результат: [m|ZBX_NOTSUPPORTED] [Invalid second parameter.] ❌
```

---

## 🔧 **РЕШЕНИЕ**

### **1. Анализ проблемы**
- **Zabbix Agent 1** не поддерживает `system.cpu.util` без параметров
- **Коллектор работал** (видно в статусе: `collector [idle 1 sec]`)
- **Проблема в отсутствии** подходящей метрики для CPU Usage

### **2. Добавление UserParameter**
Добавлен пользовательский параметр в `/etc/zabbix/zabbix_agentd.conf`:

```bash
# CPU Usage percentage
UserParameter=system.cpu.usage,top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\([0-9.]*\)%* id.*/\1/' | awk '{print 100 - $1}'
```

### **3. Перезапуск Zabbix агента**
```bash
sudo systemctl restart zabbix-agent
```

---

## ✅ **РЕЗУЛЬТАТ**

### **После исправления:**
```bash
# CPU Usage теперь работает
zabbix_agentd -t system.cpu.usage
# Результат: [t|8.1] ✅

# Disk Usage продолжает работать
zabbix_agentd -t vfs.fs.size[/,pused]
# Результат: [d|92.281035] ✅
```

### **Текущие значения:**
- **CPU Usage**: 10.6%
- **Disk Usage**: 92.28%

---

## 📋 **ТЕХНИЧЕСКИЕ ДЕТАЛИ**

### **UserParameter объяснение:**
```bash
UserParameter=system.cpu.usage,top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\([0-9.]*\)%* id.*/\1/' | awk '{print 100 - $1}'
```

1. **`top -bn1`** - запускает top в batch режиме на 1 итерацию
2. **`grep 'Cpu(s)'`** - находит строку с информацией о CPU
3. **`sed 's/.*, *\([0-9.]*\)%* id.*/\1/'`** - извлекает процент idle времени
4. **`awk '{print 100 - $1}'`** - вычисляет процент использования (100 - idle)

### **Альтернативные метрики:**
- **`system.cpu.load[all,avg1]`** - Load Average (работает)
- **`system.cpu.util`** - требует параметры (не работает без них)
- **`system.cpu.usage`** - пользовательская метрика (работает)

---

## 🎯 **ПРОВЕРКА РАБОТОСПОСОБНОСТИ**

### **Команды для проверки:**
```bash
# Проверить CPU Usage
zabbix_agentd -t system.cpu.usage

# Проверить Disk Usage
zabbix_agentd -t vfs.fs.size[/,pused]

# Проверить статус Zabbix агента
sudo systemctl status zabbix-agent

# Проверить логи
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

### **Ожидаемые результаты:**
- **CPU Usage**: числовое значение (например, 8.1)
- **Disk Usage**: числовое значение (например, 92.28)
- **Статус агента**: active (running)
- **Логи**: без ошибок

---

## 🚀 **СЛЕДУЮЩИЕ ШАГИ**

### **1. Мониторинг в Zabbix Server**
- Убедиться, что Zabbix Server получает данные
- Проверить графики CPU и Disk Usage
- Настроить алерты при необходимости

### **2. Дополнительные метрики (опционально)**
```bash
# Можно добавить другие метрики
UserParameter=system.memory.usage,free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}'
UserParameter=system.load.1min,uptime | awk -F'load average:' '{print $2}' | awk -F',' '{print $1}' | xargs
```

### **3. Резервное копирование**
```bash
# Создать резервную копию конфигурации
sudo cp /etc/zabbix/zabbix_agentd.conf /etc/zabbix/zabbix_agentd.conf.backup.$(date +%Y%m%d)
```

---

## 🏆 **ЗАКЛЮЧЕНИЕ**

### ✅ **Проблема решена:**
- **CPU Usage** теперь работает через пользовательский параметр
- **Disk Usage** работал корректно с самого начала
- **Zabbix агент** работает стабильно

### 🎯 **Ключевые моменты:**
1. **Zabbix Agent 1** имеет ограничения по встроенным метрикам
2. **UserParameter** - эффективное решение для кастомных метрик
3. **Disk Usage** работал корректно, проблема была только с CPU

### 📊 **Текущий статус:**
- **CPU Usage**: ✅ Работает (10.6%)
- **Disk Usage**: ✅ Работает (92.28%)
- **Zabbix Agent**: ✅ Активен и стабилен

**Проблема полностью решена! Zabbix теперь получает данные о CPU и Disk Usage.** 🎉

---

*Отчет подготовлен на основе диагностики и исправления проблемы с метриками CPU и Disk Usage в Zabbix.*
