# 🤖 Финальный AI Zabbix шаблон - Руководство

## 🎯 Обзор

Создан **рабочий AI шаблон** для Zabbix на основе проверенной структуры. Шаблон содержит все необходимые AI метрики и триггеры с правильными константами Zabbix 6.0.

## ✅ Что включено

### **14 AI метрик:**
1. **AI Service Status** - статус AI сервиса (текст)
2. **AI Service Health** - здоровье сервиса (0/1)
3. **AI Anomaly Detections** - количество обнаруженных аномалий
4. **AI Load Predictions** - количество предсказаний нагрузки
5. **AI Routes Optimized** - количество оптимизированных маршрутов
6. **AI Processing Time** - время обработки (мс)
7. **AI Cache Hits** - попадания в кэш
8. **AI Cache Misses** - промахи кэша
9. **AI Circuit Breaker State** - состояние circuit breaker
10. **AI Error Rate** - процент ошибок
11. **AI Anomaly Detection Accuracy** - точность обнаружения аномалий
12. **AI Load Prediction Accuracy** - точность предсказания нагрузки
13. **AI Route Optimization Accuracy** - точность оптимизации маршрутов
14. **AI Active Connections** - активные соединения
15. **AI Requests per Second** - запросов в секунду

### **11 триггеров:**
- **6 критических** (HIGH priority)
- **5 предупреждающих** (WARNING priority)

### **Макросы:**
- `{$CB.AI.STATUS}` = "running"
- `{$CB.AI.PROCESSING_TIME.WARN}` = 500ms
- `{$CB.AI.PROCESSING_TIME.HIGH}` = 1000ms
- `{$CB.AI.ERROR_RATE.WARN}` = 5%
- `{$CB.AI.ERROR_RATE.HIGH}` = 10%
- `{$CB.AI.ACCURACY.WARN}` = 80%
- `{$CB.AI.ACCURACY.HIGH}` = 70%
- `{$CB.NODATA.WINDOW}` = 2m

## 📋 Пошаговый импорт

### **Шаг 1: Импорт шаблона**

1. Откройте веб-интерфейс Zabbix: http://10.0.0.241/zabbix
2. Войдите в систему (Admin / zabbix)
3. Перейдите в **Configuration → Templates → Import**
4. Нажмите **Choose File**
5. Выберите файл `scripts/cloudbridge_ai_zabbix_template_working.xml`
6. Нажмите **Import**

**Ожидаемый результат:** ✅ Template imported successfully

### **Шаг 2: Проверка импорта**

1. Перейдите в **Configuration → Templates**
2. Найдите шаблон **CloudBridge AI Monitoring**
3. Нажмите на название шаблона
4. Проверьте:
   - **Items:** 14 AI метрик
   - **Triggers:** 11 триггеров
   - **Macros:** 8 макросов

### **Шаг 3: Применение к хосту**

1. Перейдите в **Configuration → Hosts**
2. Найдите хост CloudBridge Relay
3. Нажмите на имя хоста
4. Перейдите на вкладку **Templates**
5. В поле **Link new templates** выберите **CloudBridge AI Monitoring**
6. Нажмите **Update**

### **Шаг 4: Настройка UserParameter'ов**

Убедитесь, что на хосте CloudBridge настроены UserParameter'ы:

```bash
# Запустите скрипт интеграции
sudo ./scripts/ai_zabbix_integration.sh
```

## 🧪 Тестирование

### **1. Проверка сбора данных**

```bash
# На хосте CloudBridge
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
zabbix_agentd -t ai.load_predictions
zabbix_agentd -t ai.routes_optimized
zabbix_agentd -t ai.processing_time
zabbix_agentd -t ai.cache_hits
zabbix_agentd -t ai.cache_misses
zabbix_agentd -t ai.circuit_breaker_state
zabbix_agentd -t ai.error_rate
zabbix_agentd -t ai.anomaly_detection_accuracy
zabbix_agentd -t ai.load_prediction_accuracy
zabbix_agentd -t ai.route_optimization_accuracy
zabbix_agentd -t ai.active_connections
zabbix_agentd -t ai.requests_per_second
```

### **2. Проверка в Zabbix**

1. **Monitoring → Latest data**
2. Выберите хост CloudBridge
3. Проверьте, что все AI метрики собираются

### **3. Проверка триггеров**

1. **Monitoring → Problems**
2. Убедитесь, что триггеры работают корректно

## 🚨 Критические триггеры

### **HIGH Priority:**
1. **AI Service Down** - AI сервис недоступен
2. **AI Service Status Not Running** - статус не "running"
3. **AI Circuit Breaker Open** - circuit breaker открыт
4. **High AI Processing Time** - время обработки > 1000ms
5. **High AI Error Rate** - процент ошибок > 10%
6. **AI Service: no data** - нет данных от AI сервиса

### **WARNING Priority:**
1. **AI Processing Time Warning** - время обработки > 500ms
2. **AI Error Rate Warning** - процент ошибок > 5%
3. **Low AI Anomaly Detection Accuracy** - точность < 80%
4. **Low AI Load Prediction Accuracy** - точность < 80%
5. **Low AI Route Optimization Accuracy** - точность < 80%

## 🔧 Настройка макросов

Вы можете настроить пороговые значения через макросы:

1. **Configuration → Templates → CloudBridge AI Monitoring → Macros**
2. Измените значения по необходимости:
   - `{$CB.AI.PROCESSING_TIME.WARN}` - предупреждение по времени обработки
   - `{$CB.AI.PROCESSING_TIME.HIGH}` - критическое время обработки
   - `{$CB.AI.ERROR_RATE.WARN}` - предупреждение по ошибкам
   - `{$CB.AI.ERROR_RATE.HIGH}` - критический процент ошибок
   - `{$CB.AI.ACCURACY.WARN}` - предупреждение по точности

## 📊 Мониторинг

### **Доступные метрики:**

#### **Доступность:**
- AI Service Status
- AI Service Health

#### **Производительность:**
- AI Processing Time
- AI Requests per Second
- AI Active Connections

#### **Качество:**
- AI Anomaly Detection Accuracy
- AI Load Prediction Accuracy
- AI Route Optimization Accuracy

#### **Надежность:**
- AI Circuit Breaker State
- AI Error Rate

#### **Кэширование:**
- AI Cache Hits
- AI Cache Misses

#### **Активность:**
- AI Anomaly Detections
- AI Load Predictions
- AI Routes Optimized

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
3. Проверьте настройки макросов

## 📞 Поддержка

Если проблемы продолжаются:
1. Проверьте версию Zabbix (должна быть 6.0 или выше)
2. Проверьте логи: `/var/log/zabbix/zabbix_server.log`
3. Убедитесь, что AI Service запущен и доступен

## 🎯 Результат

После успешного импорта у вас будет:
- ✅ **14 AI метрик** с правильными типами данных
- ✅ **11 триггеров** для мониторинга
- ✅ **8 настраиваемых макросов**
- ✅ **Автоматический мониторинг** AI компонентов
- ✅ **Уведомления** о проблемах
- ✅ **Полная интеграция** с Zabbix 6.0

**AI метрики теперь полностью интегрированы с Zabbix!** 🚀

