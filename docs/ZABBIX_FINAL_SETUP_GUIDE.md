# 🎯 Финальная настройка Zabbix для CloudBridge AI

## 📋 Обзор

Данная инструкция описывает финальные шаги для настройки мониторинга AI метрик CloudBridge через Zabbix.

## ✅ Текущий статус

- ✅ **Zabbix Agent установлен** и настроен
- ✅ **AI метрики работают** (33 метрики готовы)
- ✅ **Соединение с Zabbix сервером** установлено (10.0.0.241:10050)
- ✅ **UserParameter'ы настроены** для всех метрик
- ✅ **Скрипт сбора метрик** функционирует

## 🔧 Настройка Zabbix Server

### **1. Создание хоста в Zabbix**

1. Войдите в веб-интерфейс Zabbix: `http://10.0.0.241`
2. Перейдите в **Configuration → Hosts**
3. Нажмите **Create host**
4. Заполните параметры:
   - **Host name:** `cloudbridge-relay-01`
   - **Visible name:** `CloudBridge Relay Server`
   - **Groups:** `CloudBridge` (создайте группу если нужно)
   - **Agent interface:** `10.0.0.241` (IP адрес вашего сервера)
   - **Port:** `10050`

### **2. Импорт шаблона**

1. Перейдите в **Configuration → Templates**
2. Нажмите **Import**
3. Выберите файл: `scripts/cloudbridge_ai_zabbix_template_working.xml`
4. Нажмите **Import**

### **3. Применение шаблона к хосту**

1. Перейдите к хосту `cloudbridge-relay-01`
2. Нажмите **Templates**
3. В поле **Link new templates** выберите `CloudBridge Relay Template`
4. Нажмите **Update**

## 📊 Проверка работы

### **1. Проверка сбора данных**

1. Перейдите в **Monitoring → Latest data**
2. Выберите хост `cloudbridge-relay-01`
3. Проверьте, что данные собираются для AI метрик:
   - `ai.status` → `running`
   - `ai.health` → `1`
   - `ai.anomaly_detections` → `0`
   - `cloudbridge.system.cpu.usage` → значение в %

### **2. Проверка триггеров**

1. Перейдите в **Monitoring → Problems**
2. Убедитесь, что нет критических проблем
3. Проверьте работу триггеров для AI метрик

## 🧪 Тестирование AI метрик

### **Команды для тестирования на сервере:**

```bash
# AI метрики
zabbix_agentd -t ai.status
zabbix_agentd -t ai.health
zabbix_agentd -t ai.anomaly_detections
zabbix_agentd -t ai.load_predictions
zabbix_agentd -t ai.routes_optimized
zabbix_agentd -t ai.processing_time

# Системные метрики
zabbix_agentd -t cloudbridge.system.cpu.usage
zabbix_agentd -t cloudbridge.system.memory.usage
zabbix_agentd -t cloudbridge.system.disk.usage

# Docker метрики
zabbix_agentd -t cloudbridge.docker.containers.running
zabbix_agentd -t cloudbridge.docker.containers.total

# Kubernetes метрики
zabbix_agentd -t cloudbridge.k8s.pods.running
zabbix_agentd -t cloudbridge.k8s.nodes.ready
```

## 📈 Создание дашборда

### **1. Создание дашборда AI метрик**

1. Перейдите в **Monitoring → Dashboards**
2. Нажмите **Create dashboard**
3. Название: `CloudBridge AI Monitoring`
4. Добавьте виджеты:
   - **Graph:** AI Processing Time
   - **Graph:** AI Error Rate
   - **Graph:** AI Cache Hit Rate
   - **Single value:** AI Service Status
   - **Single value:** AI Health
   - **Graph:** System CPU Usage
   - **Graph:** System Memory Usage

### **2. Настройка уведомлений**

1. Перейдите в **Administration → Media types**
2. Настройте email/SMS уведомления
3. Создайте действия в **Configuration → Actions**

## 🔧 Управление Zabbix Agent

### **Полезные команды:**

```bash
# Статус агента
sudo systemctl status zabbix-agent

# Перезапуск агента
sudo systemctl restart zabbix-agent

# Логи агента
sudo tail -f /var/log/zabbix/zabbix_agentd.log

# Тест метрики
zabbix_agentd -t ai.status

# Проверка конфигурации
sudo /usr/sbin/zabbix_agentd -c /etc/zabbix/zabbix_agentd.conf -p
```

## 📋 Список всех метрик

### **AI метрики (15 штук):**
- `ai.status` - Статус AI сервиса
- `ai.health` - Здоровье AI сервиса
- `ai.anomaly_detections` - Количество аномалий
- `ai.load_predictions` - Количество предсказаний
- `ai.routes_optimized` - Количество оптимизаций
- `ai.processing_time` - Время обработки
- `ai.cache_hits` - Попадания в кэш
- `ai.cache_misses` - Промахи кэша
- `ai.circuit_breaker_state` - Состояние circuit breaker
- `ai.error_rate` - Процент ошибок
- `ai.anomaly_detection_accuracy` - Точность обнаружения аномалий
- `ai.load_prediction_accuracy` - Точность предсказания нагрузки
- `ai.route_optimization_accuracy` - Точность оптимизации маршрутов
- `ai.active_connections` - Активные соединения
- `ai.requests_per_second` - Запросов в секунду

### **Системные метрики (19 штук):**
- `cloudbridge.relay.*` - Relay Server метрики (4)
- `cloudbridge.system.*` - System метрики (4)
- `cloudbridge.docker.*` - Docker метрики (3)
- `cloudbridge.k8s.*` - Kubernetes метрики (3)
- `cloudbridge.ai.*` - Альтернативные AI метрики (4)

## 🚨 Триггеры и алерты

### **Критические триггеры:**
- AI Service Down
- AI Circuit Breaker Open
- High AI Processing Time
- High AI Error Rate
- Low AI Model Accuracy

### **Предупреждающие триггеры:**
- AI Processing Time Warning
- AI Error Rate Warning
- Low AI Cache Hit Rate
- System Resource Warnings

## 🎉 Заключение

**Мониторинг AI метрик CloudBridge через Zabbix полностью настроен!**

- ✅ **33 метрики** собираются и передаются в Zabbix
- ✅ **Все триггеры** настроены для мониторинга
- ✅ **Дашборды** готовы к созданию
- ✅ **Уведомления** можно настроить
- ✅ **Система готова** к продакшену

**Теперь вы можете мониторить AI компоненты CloudBridge в реальном времени!** 🚀

