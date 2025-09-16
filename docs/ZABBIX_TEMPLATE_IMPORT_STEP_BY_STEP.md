# 📋 Пошаговый импорт Zabbix шаблона для AI метрик

## 🚨 Проблема с импортом

Ошибка `Invalid tag "/zabbix_export/templates/template(1)": unexpected tag "triggers"` возникает из-за неправильной структуры XML в Zabbix 6.0.

## ✅ Решение: Пошаговый импорт

### **Шаг 1: Импорт основного шаблона с Items**

1. Откройте веб-интерфейс Zabbix: http://10.0.0.241/zabbix
2. Войдите в систему (Admin / zabbix)
3. Перейдите в **Configuration → Templates → Import**
4. Нажмите **Choose File**
5. Выберите файл `scripts/cloudbridge_ai_zabbix_template_items_only.xml`
6. Нажмите **Import**

**Ожидаемый результат:** ✅ Template imported successfully

### **Шаг 2: Проверка импорта Items**

1. Перейдите в **Configuration → Templates**
2. Найдите шаблон **CloudBridge AI Monitoring**
3. Нажмите на название шаблона
4. Перейдите на вкладку **Items**
5. Убедитесь, что все 10 AI метрик импортированы:
   - AI Service Status
   - AI Service Health
   - Anomaly Detections
   - Load Predictions
   - Routes Optimized
   - AI Processing Time
   - AI Cache Hits
   - AI Cache Misses
   - AI Circuit Breaker State
   - AI Error Rate

### **Шаг 3: Применение шаблона к хосту**

1. Перейдите в **Configuration → Hosts**
2. Найдите хост CloudBridge Relay
3. Нажмите на имя хоста
4. Перейдите на вкладку **Templates**
5. В поле **Link new templates** выберите **CloudBridge AI Monitoring**
6. Нажмите **Update**

### **Шаг 4: Создание триггеров вручную**

Поскольку импорт триггеров через XML может вызывать проблемы, создадим их вручную:

#### **4.1. Создание триггера "AI Service Down"**

1. Перейдите в **Configuration → Templates**
2. Нажмите на **CloudBridge AI Monitoring**
3. Перейдите на вкладку **Triggers**
4. Нажмите **Create trigger**
5. Заполните поля:
   - **Name:** `AI Service Down`
   - **Expression:** `last(/CloudBridge AI Monitoring/ai.health)=0`
   - **Severity:** `High`
   - **Description:** `AI service is not responding`
6. Нажмите **Add**

#### **4.2. Создание триггера "AI Circuit Breaker Open"**

1. Нажмите **Create trigger**
2. Заполните поля:
   - **Name:** `AI Circuit Breaker Open`
   - **Expression:** `last(/CloudBridge AI Monitoring/ai.circuit_breaker_state)="open"`
   - **Severity:** `High`
   - **Description:** `AI circuit breaker is open`
3. Нажмите **Add**

#### **4.3. Создание триггера "High AI Processing Time"**

1. Нажмите **Create trigger**
2. Заполните поля:
   - **Name:** `High AI Processing Time`
   - **Expression:** `last(/CloudBridge AI Monitoring/ai.processing_time)>1000`
   - **Severity:** `Average`
   - **Description:** `AI processing time is above 1000ms`
3. Нажмите **Add**

#### **4.4. Создание триггера "High AI Error Rate"**

1. Нажмите **Create trigger**
2. Заполните поля:
   - **Name:** `High AI Error Rate`
   - **Expression:** `last(/CloudBridge AI Monitoring/ai.error_rate)>10`
   - **Severity:** `High`
   - **Description:** `AI error rate is above 10%`
3. Нажмите **Add**

### **Шаг 5: Тестирование**

#### **5.1. Проверка сбора данных**

1. Перейдите в **Monitoring → Latest data**
2. Выберите хост CloudBridge
3. Убедитесь, что AI метрики собираются:
   - AI Service Status
   - AI Service Health
   - Anomaly Detections
   - Load Predictions
   - Routes Optimized
   - AI Processing Time
   - AI Cache Hits
   - AI Cache Misses
   - AI Circuit Breaker State
   - AI Error Rate

#### **5.2. Проверка триггеров**

1. Перейдите в **Monitoring → Problems**
2. Убедитесь, что триггеры работают корректно
3. Если есть проблемы, проверьте выражения триггеров

## 🔧 Альтернативный способ: Создание шаблона с нуля

Если импорт все еще не работает, создайте шаблон полностью вручную:

### **Создание шаблона**

1. **Configuration → Templates → Create template**
2. Заполните поля:
   - **Template name:** `CloudBridge AI Monitoring`
   - **Groups:** `CloudBridge` (создайте группу, если не существует)
   - **Description:** `Template for monitoring CloudBridge AI components and metrics`

### **Добавление Items**

Для каждого Item:
1. **Configuration → Templates → CloudBridge AI Monitoring → Items → Create item**
2. Заполните поля согласно таблице:

| Name | Key | Type | Value Type | Units | History | Trends |
|------|-----|------|------------|-------|---------|--------|
| AI Service Status | ai.status | Zabbix agent | Text | - | 7d | 365d |
| AI Service Health | ai.health | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| Anomaly Detections | ai.anomaly_detections | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| Load Predictions | ai.load_predictions | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| Routes Optimized | ai.routes_optimized | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| AI Processing Time | ai.processing_time | Zabbix agent | Numeric (float) | ms | 7d | 365d |
| AI Cache Hits | ai.cache_hits | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| AI Cache Misses | ai.cache_misses | Zabbix agent | Numeric (unsigned) | - | 7d | 365d |
| AI Circuit Breaker State | ai.circuit_breaker_state | Zabbix agent | Text | - | 7d | 365d |
| AI Error Rate | ai.error_rate | Zabbix agent | Numeric (float) | % | 7d | 365d |

## 🧪 Финальное тестирование

### **1. Проверка на хосте**

```bash
# На хосте CloudBridge
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
```

### **2. Проверка в Zabbix**

1. **Monitoring → Latest data**
2. Выберите хост CloudBridge
3. Проверьте, что все AI метрики собираются

### **3. Проверка триггеров**

1. **Monitoring → Problems**
2. Убедитесь, что триггеры работают

## 🎯 Результат

После успешного выполнения всех шагов у вас будет:

- ✅ **10 AI метрик** в Zabbix
- ✅ **4 критических триггера**
- ✅ **Автоматический мониторинг** AI компонентов
- ✅ **Уведомления** о проблемах

## 🚨 Устранение неполадок

### **Проблема: Items не собирают данные**

**Решение:**
1. Убедитесь, что Zabbix Agent запущен
2. Проверьте конфигурацию UserParameter'ов
3. Перезапустите Zabbix Agent: `sudo systemctl restart zabbix-agent`

### **Проблема: Ошибки в логах Zabbix Agent**

**Диагностика:**
```bash
sudo tail -f /var/log/zabbix/zabbix_agentd.log
```

**Решение:**
1. Проверьте синтаксис скрипта `/usr/local/bin/ai_metrics_collector.sh`
2. Убедитесь, что AI Service доступен
3. Проверьте права доступа к скрипту

## 📞 Поддержка

Если проблемы продолжаются:
1. Проверьте версию Zabbix (должна быть 6.0 или выше)
2. Проверьте права доступа к файлам
3. Проверьте логи: `/var/log/zabbix/zabbix_server.log`

**AI метрики теперь полностью интегрированы с Zabbix!** 🚀

