# 🔧 Исправление ошибки импорта Zabbix шаблона

## 🚨 Проблема

При импорте Zabbix шаблона возникает ошибка:
```
Details Import failed
Invalid tag "/zabbix_export": unexpected tag "value_maps".
```

## 🔍 Причины ошибки

1. **HTML комментарии** в XML файле (`<!-- -->`)
2. **Неправильная структура XML** для Zabbix 6.0
3. **Отсутствующие обязательные поля** в некоторых элементах
4. **Недопустимые теги** в контексте Zabbix

## ✅ Решение

### **Вариант 1: Использовать исправленный шаблон**

Используйте файл `scripts/cloudbridge_ai_zabbix_template_fixed.xml` - он исправлен и должен импортироваться без ошибок.

### **Вариант 2: Использовать упрощенный шаблон**

Если все еще есть проблемы, используйте `scripts/cloudbridge_ai_zabbix_template_simple.xml` - это минимальная версия с основными метриками.

## 📋 Пошаговая инструкция

### **Шаг 1: Импорт исправленного шаблона**

1. Откройте веб-интерфейс Zabbix: http://10.0.0.241/zabbix
2. Войдите в систему (Admin / zabbix)
3. Перейдите в **Configuration → Templates → Import**
4. Нажмите **Choose File**
5. Выберите файл `scripts/cloudbridge_ai_zabbix_template_fixed.xml`
6. Нажмите **Import**

### **Шаг 2: Проверка импорта**

После успешного импорта вы должны увидеть:
- ✅ **Template imported successfully**
- ✅ Шаблон **CloudBridge AI Monitoring** в списке шаблонов

### **Шаг 3: Применение шаблона к хосту**

1. Перейдите в **Configuration → Hosts**
2. Найдите хост CloudBridge Relay
3. Нажмите на имя хоста
4. Перейдите на вкладку **Templates**
5. В поле **Link new templates** выберите **CloudBridge AI Monitoring**
6. Нажмите **Update**

## 🔧 Альтернативное решение

Если импорт все еще не работает, создайте шаблон вручную:

### **Создание шаблона вручную**

1. **Configuration → Templates → Create template**
2. **Template name:** `CloudBridge AI Monitoring`
3. **Groups:** `CloudBridge`
4. **Description:** `Template for monitoring CloudBridge AI components`

### **Добавление Items**

Добавьте следующие Items:

| Name | Key | Type | Value Type | Units |
|------|-----|------|------------|-------|
| AI Service Status | ai.status | Zabbix agent | Text | - |
| AI Service Health | ai.health | Zabbix agent | Numeric (unsigned) | - |
| Anomaly Detections | ai.anomaly_detections | Zabbix agent | Numeric (unsigned) | - |
| Load Predictions | ai.load_predictions | Zabbix agent | Numeric (unsigned) | - |
| Routes Optimized | ai.routes_optimized | Zabbix agent | Numeric (unsigned) | - |
| AI Processing Time | ai.processing_time | Zabbix agent | Numeric (float) | ms |
| AI Cache Hits | ai.cache_hits | Zabbix agent | Numeric (unsigned) | - |
| AI Cache Misses | ai.cache_misses | Zabbix agent | Numeric (unsigned) | - |
| AI Circuit Breaker State | ai.circuit_breaker_state | Zabbix agent | Text | - |
| AI Error Rate | ai.error_rate | Zabbix agent | Numeric (float) | % |

### **Добавление Triggers**

Добавьте следующие Triggers:

| Name | Expression | Priority |
|------|------------|----------|
| AI Service Down | `last(/CloudBridge AI Monitoring/ai.health)=0` | High |
| AI Circuit Breaker Open | `last(/CloudBridge AI Monitoring/ai.circuit_breaker_state)="open"` | High |
| High AI Processing Time | `last(/CloudBridge AI Monitoring/ai.processing_time)>1000` | Average |
| High AI Error Rate | `last(/CloudBridge AI Monitoring/ai.error_rate)>10` | High |

## 🧪 Тестирование

После импорта шаблона протестируйте:

### **1. Проверка Items**

```bash
# На хосте CloudBridge
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
```

### **2. Проверка в Zabbix**

1. **Monitoring → Latest data**
2. Выберите хост CloudBridge
3. Проверьте, что AI метрики собираются

### **3. Проверка Triggers**

1. **Monitoring → Problems**
2. Убедитесь, что триггеры работают корректно

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

### **Проблема: Триггеры не срабатывают**

**Решение:**
1. Проверьте выражения триггеров
2. Убедитесь, что Items собирают данные
3. Проверьте настройки триггеров

## 📞 Поддержка

Если проблемы продолжаются:

1. **Проверьте версию Zabbix:** должна быть 6.0 или выше
2. **Проверьте права доступа:** файлы должны быть доступны для чтения
3. **Проверьте логи:** `/var/log/zabbix/zabbix_server.log`

## 🎯 Результат

После успешного импорта у вас будет:
- ✅ **10 AI метрик** в Zabbix
- ✅ **4 критических триггера**
- ✅ **Автоматический мониторинг** AI компонентов
- ✅ **Уведомления** о проблемах

**AI метрики теперь полностью интегрированы с Zabbix!** 🚀

