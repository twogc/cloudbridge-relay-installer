# 🔧 Исправление констант Zabbix 6.0

## 🚨 Проблема

Ошибка `Invalid tag "/zabbix_export/templates/template(1)/items/item(1)/type": unexpected constant "0"` возникает из-за использования старых числовых констант в Zabbix 6.0.

## 🔍 Изменения в Zabbix 6.0

В Zabbix 6.0 изменились константы с числовых на текстовые:

### **Типы элементов (type):**
- ❌ `0` → ✅ `ZABBIX_AGENT`
- ❌ `1` → ✅ `SNMP_AGENT`
- ❌ `2` → ✅ `IPMI_AGENT`
- ❌ `3` → ✅ `JMX_AGENT`
- ❌ `4` → ✅ `DEPENDENT_ITEM`
- ❌ `5` → ✅ `HTTP_AGENT`
- ❌ `6` → ✅ `SSH_AGENT`
- ❌ `7` → ✅ `TELNET_AGENT`
- ❌ `8` → ✅ `CALCULATED`
- ❌ `9` → ✅ `EXTERNAL_CHECK`
- ❌ `10` → ✅ `ODBC_AGENT`
- ❌ `11` → ✅ `IPMI_AGENT`
- ❌ `12` → ✅ `SSH_AGENT`
- ❌ `13` → ✅ `HTTP_AGENT`
- ❌ `14` → ✅ `SNMP_TRAP`
- ❌ `15` → ✅ `DEPENDENT_ITEM`
- ❌ `16` → ✅ `HTTP_AGENT`
- ❌ `17` → ✅ `SNMP_AGENT`
- ❌ `18` → ✅ `SNMP_AGENT`
- ❌ `19` → ✅ `SNMP_AGENT`
- ❌ `20` → ✅ `EXTERNAL_CHECK`

### **Типы значений (value_type):**
- ❌ `0` → ✅ `FLOAT`
- ❌ `1` → ✅ `CHAR`
- ❌ `2` → ✅ `LOG`
- ❌ `3` → ✅ `UINT64`
- ❌ `4` → ✅ `TEXT`

## ✅ Решение

### **Используйте исправленный шаблон**

Файл `scripts/cloudbridge_ai_zabbix_template_final.xml` содержит правильные константы для Zabbix 6.0:

```xml
<item>
    <uuid>ai-status-uuid</uuid>
    <name>AI Service Status</name>
    <type>ZABBIX_AGENT</type>          <!-- Вместо 0 -->
    <key>ai.status</key>
    <value_type>TEXT</value_type>      <!-- Вместо 1 -->
    <units/>
    <description>Current status of AI service</description>
</item>
```

### **Правильные константы для AI метрик:**

| Метрика | type | value_type | Описание |
|---------|------|------------|----------|
| AI Service Status | `ZABBIX_AGENT` | `TEXT` | Текстовый статус |
| AI Service Health | `ZABBIX_AGENT` | `UINT64` | Числовое значение (0/1) |
| Anomaly Detections | `ZABBIX_AGENT` | `UINT64` | Счетчик |
| Load Predictions | `ZABBIX_AGENT` | `UINT64` | Счетчик |
| Routes Optimized | `ZABBIX_AGENT` | `UINT64` | Счетчик |
| AI Processing Time | `ZABBIX_AGENT` | `FLOAT` | Время в миллисекундах |
| AI Cache Hits | `ZABBIX_AGENT` | `UINT64` | Счетчик |
| AI Cache Misses | `ZABBIX_AGENT` | `UINT64` | Счетчик |
| AI Circuit Breaker State | `ZABBIX_AGENT` | `TEXT` | Текстовое состояние |
| AI Error Rate | `ZABBIX_AGENT` | `FLOAT` | Процент ошибок |

## 📋 Пошаговый импорт

### **Шаг 1: Импорт исправленного шаблона**

1. Откройте веб-интерфейс Zabbix: http://10.0.0.241/zabbix
2. Войдите в систему (Admin / zabbix)
3. Перейдите в **Configuration → Templates → Import**
4. Нажмите **Choose File**
5. Выберите файл `scripts/cloudbridge_ai_zabbix_template_final.xml`
6. Нажмите **Import**

**Ожидаемый результат:** ✅ Template imported successfully

### **Шаг 2: Проверка импорта**

1. Перейдите в **Configuration → Templates**
2. Найдите шаблон **CloudBridge AI Monitoring**
3. Нажмите на название шаблона
4. Перейдите на вкладку **Items**
5. Убедитесь, что все 10 AI метрик импортированы с правильными типами

### **Шаг 3: Применение к хосту**

1. Перейдите в **Configuration → Hosts**
2. Найдите хост CloudBridge Relay
3. Нажмите на имя хоста
4. Перейдите на вкладку **Templates**
5. В поле **Link new templates** выберите **CloudBridge AI Monitoring**
6. Нажмите **Update**

### **Шаг 4: Создание триггеров**

Создайте триггеры вручную в веб-интерфейсе:

1. **Configuration → Templates → CloudBridge AI Monitoring → Triggers → Create trigger**

#### **Триггер 1: AI Service Down**
- **Name:** `AI Service Down`
- **Expression:** `last(/CloudBridge AI Monitoring/ai.health)=0`
- **Severity:** `High`

#### **Триггер 2: AI Circuit Breaker Open**
- **Name:** `AI Circuit Breaker Open`
- **Expression:** `last(/CloudBridge AI Monitoring/ai.circuit_breaker_state)="open"`
- **Severity:** `High`

#### **Триггер 3: High AI Processing Time**
- **Name:** `High AI Processing Time`
- **Expression:** `last(/CloudBridge AI Monitoring/ai.processing_time)>1000`
- **Severity:** `Average`

#### **Триггер 4: High AI Error Rate**
- **Name:** `High AI Error Rate`
- **Expression:** `last(/CloudBridge AI Monitoring/ai.error_rate)>10`
- **Severity:** `High`

## 🧪 Тестирование

### **1. Проверка сбора данных**

```bash
# На хосте CloudBridge
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
```

### **2. Проверка в Zabbix**

1. **Monitoring → Latest data**
2. Выберите хост CloudBridge
3. Проверьте, что AI метрики собираются с правильными типами данных

### **3. Проверка триггеров**

1. **Monitoring → Problems**
2. Убедитесь, что триггеры работают корректно

## 🔧 Альтернативное решение

Если импорт все еще не работает, создайте шаблон полностью вручную:

### **Создание шаблона**

1. **Configuration → Templates → Create template**
2. Заполните поля:
   - **Template name:** `CloudBridge AI Monitoring`
   - **Groups:** `CloudBridge`
   - **Description:** `Template for monitoring CloudBridge AI components and metrics`

### **Добавление Items**

Для каждого Item используйте правильные типы:

1. **Configuration → Templates → CloudBridge AI Monitoring → Items → Create item**
2. Заполните поля согласно таблице выше

## 🚨 Устранение неполадок

### **Проблема: Items не собирают данные**

**Решение:**
1. Убедитесь, что Zabbix Agent запущен
2. Проверьте конфигурацию UserParameter'ов
3. Перезапустите Zabbix Agent: `sudo systemctl restart zabbix-agent`

### **Проблема: Неправильные типы данных**

**Решение:**
1. Проверьте, что используете правильные константы Zabbix 6.0
2. Убедитесь, что value_type соответствует типу данных
3. Проверьте настройки Items в веб-интерфейсе

## 📞 Поддержка

Если проблемы продолжаются:
1. Проверьте версию Zabbix (должна быть 6.0 или выше)
2. Проверьте логи: `/var/log/zabbix/zabbix_server.log`
3. Убедитесь, что используете правильные константы

## 🎯 Результат

После успешного импорта у вас будет:
- ✅ **10 AI метрик** с правильными типами данных
- ✅ **4 критических триггера**
- ✅ **Автоматический мониторинг** AI компонентов
- ✅ **Уведомления** о проблемах

**AI метрики теперь полностью интегрированы с Zabbix 6.0!** 🚀

