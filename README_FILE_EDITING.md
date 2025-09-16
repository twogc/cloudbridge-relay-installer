# 📝 Как редактировать ai_metrics_collector.sh

## 🚨 **Проблема решена!**

Теперь вы можете редактировать скрипт сбора метрик без проблем с правами доступа.

---

## 🔧 **Как редактировать:**

### **1. Редактирование:**
```bash
# Откройте файл для редактирования
code scripts/ai_metrics_collector.sh
# или
nano scripts/ai_metrics_collector.sh
```

### **2. Применение изменений:**
```bash
# Запустите скрипт обновления
./scripts/update_ai_metrics_script.sh
```

### **3. Тестирование:**
```bash
# Проверьте, что изменения применились
zabbix_agentd -t ai.health
```

---

## 📁 **Файлы:**

- **`scripts/ai_metrics_collector.sh`** - для редактирования (ubuntu:ubuntu)
- **`/usr/local/bin/ai_metrics_collector.sh`** - исполняемый файл (root:root)
- **`scripts/update_ai_metrics_script.sh`** - скрипт обновления

---

## ✅ **Готово!**

Теперь вы можете безопасно редактировать скрипт сбора метрик.
