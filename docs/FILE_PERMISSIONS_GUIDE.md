# 🔐 Руководство по правам доступа к файлам

## 🚨 **Проблема: Permission denied при сохранении**

### **Причина:**
Файл `/usr/local/bin/ai_metrics_collector.sh` принадлежит пользователю `root`, а вы редактируете от имени `ubuntu`.

---

## 🛠️ **Решение:**

### **1. Редактирование файла:**
```bash
# Редактируйте файл в рабочей директории
nano scripts/ai_metrics_collector.sh
# или
code scripts/ai_metrics_collector.sh
```

### **2. Обновление файла в системе:**
```bash
# Запустите скрипт обновления
./scripts/update_ai_metrics_script.sh
```

### **3. Альтернативный способ (если нужно редактировать напрямую):**
```bash
# Измените владельца файла
sudo chown ubuntu:ubuntu /usr/local/bin/ai_metrics_collector.sh

# Отредактируйте файл
nano /usr/local/bin/ai_metrics_collector.sh

# Верните права root
sudo chown root:root /usr/local/bin/ai_metrics_collector.sh
```

---

## 📁 **Структура файлов:**

### **Рабочие файлы (для редактирования):**
- `scripts/ai_metrics_collector.sh` - копия для редактирования
- `scripts/update_ai_metrics_script.sh` - скрипт обновления

### **Системные файлы (исполняемые):**
- `/usr/local/bin/ai_metrics_collector.sh` - исполняемый файл (root:root)

---

## 🔧 **Полезные команды:**

### **Проверка прав доступа:**
```bash
ls -la /usr/local/bin/ai_metrics_collector.sh
ls -la scripts/ai_metrics_collector.sh
```

### **Обновление скрипта:**
```bash
# После редактирования scripts/ai_metrics_collector.sh
./scripts/update_ai_metrics_script.sh
```

### **Тестирование:**
```bash
# Тест локальной копии
./scripts/ai_metrics_collector.sh health

# Тест системной копии
/usr/local/bin/ai_metrics_collector.sh health
```

---

## ⚠️ **Важные замечания:**

1. **Всегда редактируйте** файл в `scripts/ai_metrics_collector.sh`
2. **Используйте скрипт обновления** для применения изменений
3. **Не редактируйте напрямую** `/usr/local/bin/ai_metrics_collector.sh`
4. **Проверяйте права** после обновления

---

## 🎯 **Рекомендуемый workflow:**

1. **Редактирование:**
   ```bash
   code scripts/ai_metrics_collector.sh
   ```

2. **Обновление:**
   ```bash
   ./scripts/update_ai_metrics_script.sh
   ```

3. **Тестирование:**
   ```bash
   zabbix_agentd -t ai.health
   ```

4. **Проверка в Zabbix:**
   - Откройте веб-интерфейс Zabbix
   - Проверьте Latest data для хоста cloudbridge-relay-01

---

## 🚀 **Готово!**

Теперь вы можете безопасно редактировать скрипт сбора метрик без проблем с правами доступа.
