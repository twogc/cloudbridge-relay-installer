# 🔗 Соответствие ключей AI метрик в Zabbix

## 📋 Обзор

Данный документ показывает соответствие между ключами в Zabbix шаблоне и UserParameter'ами в Zabbix Agent для корректной передачи AI данных.

## ✅ AI метрики - Соответствие ключей

### **Основные AI метрики:**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| AI Service Status | `ai.status` | `ai.status` | ✅ Статус AI сервиса |
| AI Service Health | `ai.health` | `ai.health` | ✅ Здоровье AI сервиса |
| AI Anomaly Detections | `ai.anomaly_detections` | `ai.anomaly_detections` | ✅ Количество аномалий |
| AI Load Predictions | `ai.load_predictions` | `ai.load_predictions` | ✅ Количество предсказаний |
| AI Routes Optimized | `ai.routes_optimized` | `ai.routes_optimized` | ✅ Количество оптимизаций |
| AI Processing Time | `ai.processing_time` | `ai.processing_time` | ✅ Время обработки |
| AI Cache Hits | `ai.cache_hits` | `ai.cache_hits` | ✅ Попадания в кэш |
| AI Cache Misses | `ai.cache_misses` | `ai.cache_misses` | ✅ Промахи кэша |
| AI Circuit Breaker State | `ai.circuit_breaker_state` | `ai.circuit_breaker_state` | ✅ Состояние circuit breaker |
| AI Error Rate | `ai.error_rate` | `ai.error_rate` | ✅ Процент ошибок |
| AI Anomaly Detection Accuracy | `ai.anomaly_detection_accuracy` | `ai.anomaly_detection_accuracy` | ✅ Точность обнаружения аномалий |
| AI Load Prediction Accuracy | `ai.load_prediction_accuracy` | `ai.load_prediction_accuracy` | ✅ Точность предсказания нагрузки |
| AI Route Optimization Accuracy | `ai.route_optimization_accuracy` | `ai.route_optimization_accuracy` | ✅ Точность оптимизации маршрутов |
| AI Active Connections | `ai.active_connections` | `ai.active_connections` | ✅ Активные соединения |
| AI Requests per Second | `ai.requests_per_second` | `ai.requests_per_second` | ✅ Запросов в секунду |

### **Альтернативные AI метрики (cloudbridge.ai.*):**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| AI Service Status | `cloudbridge.ai.status` | `cloudbridge.ai.status` | ✅ Статус AI сервиса |
| AI Detections | `cloudbridge.ai.detections` | `cloudbridge.ai.detections` | ✅ Количество аномалий |
| AI Predictions | `cloudbridge.ai.predictions` | `cloudbridge.ai.predictions` | ✅ Количество предсказаний |
| AI Optimizations | `cloudbridge.ai.optimizations` | `cloudbridge.ai.optimizations` | ✅ Количество оптимизаций |

## 🔧 Системные метрики

### **Relay Server метрики:**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| Relay Active Connections | `cloudbridge.relay.connections` | `cloudbridge.relay.connections` | ✅ Активные соединения |
| Relay Uptime | `cloudbridge.relay.uptime` | `cloudbridge.relay.uptime` | ✅ Время работы |
| Relay Memory Usage | `cloudbridge.relay.memory` | `cloudbridge.relay.memory` | ✅ Использование памяти |
| Relay Goroutines | `cloudbridge.relay.goroutines` | `cloudbridge.relay.goroutines` | ✅ Количество goroutines |

### **System метрики:**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| System CPU Usage | `cloudbridge.system.cpu.usage` | `cloudbridge.system.cpu.usage` | ✅ Использование CPU |
| System Memory Usage | `cloudbridge.system.memory.usage` | `cloudbridge.system.memory.usage` | ✅ Использование памяти |
| System Disk Usage | `cloudbridge.system.disk.usage` | `cloudbridge.system.disk.usage` | ✅ Использование диска |
| System Network Connections | `cloudbridge.system.network.connections` | `cloudbridge.system.network.connections` | ✅ Сетевые соединения |

### **Docker метрики:**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| Docker Containers Running | `cloudbridge.docker.containers.running` | `cloudbridge.docker.containers.running` | ✅ Запущенные контейнеры |
| Docker Containers Total | `cloudbridge.docker.containers.total` | `cloudbridge.docker.containers.total` | ✅ Всего контейнеров |
| Docker Images Total | `cloudbridge.docker.images.total` | `cloudbridge.docker.images.total` | ✅ Всего образов |

### **Kubernetes метрики:**

| Метрика в шаблоне | Ключ в шаблоне | UserParameter | Описание |
|-------------------|----------------|---------------|----------|
| K8s Pods Running | `cloudbridge.k8s.pods.running` | `cloudbridge.k8s.pods.running` | ✅ Запущенные поды |
| K8s Pods Total | `cloudbridge.k8s.pods.total` | `cloudbridge.k8s.pods.total` | ✅ Всего подов |
| K8s Nodes Ready | `cloudbridge.k8s.nodes.ready` | `cloudbridge.k8s.nodes.ready` | ✅ Готовые ноды |

## 🧪 Тестирование соответствий

### **Тест AI метрик:**

```bash
# Основные AI метрики
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

# Альтернативные AI метрики
zabbix_agentd -t cloudbridge.ai.status
zabbix_agentd -t cloudbridge.ai.detections
zabbix_agentd -t cloudbridge.ai.predictions
zabbix_agentd -t cloudbridge.ai.optimizations
```

### **Тест системных метрик:**

```bash
# Relay метрики
zabbix_agentd -t cloudbridge.relay.connections
zabbix_agentd -t cloudbridge.relay.uptime
zabbix_agentd -t cloudbridge.relay.memory
zabbix_agentd -t cloudbridge.relay.goroutines

# System метрики
zabbix_agentd -t cloudbridge.system.cpu.usage
zabbix_agentd -t cloudbridge.system.memory.usage
zabbix_agentd -t cloudbridge.system.disk.usage
zabbix_agentd -t cloudbridge.system.network.connections

# Docker метрики
zabbix_agentd -t cloudbridge.docker.containers.running
zabbix_agentd -t cloudbridge.docker.containers.total
zabbix_agentd -t cloudbridge.docker.images.total

# Kubernetes метрики
zabbix_agentd -t cloudbridge.k8s.pods.running
zabbix_agentd -t cloudbridge.k8s.pods.total
zabbix_agentd -t cloudbridge.k8s.nodes.ready
```

## 🚨 Проблемы и решения

### **Проблема: Метрика не собирается**

**Диагностика:**
```bash
# Проверка UserParameter
zabbix_agentd -t <metric_key>

# Проверка скрипта
/usr/local/bin/ai_metrics_collector.sh <metric_type>

# Проверка AI API
curl -s http://localhost:8083/api/v1/ai/status
curl -s http://localhost:8083/api/v1/ai/metrics
```

**Решение:**
1. Убедитесь, что UserParameter добавлен в конфигурацию
2. Проверьте, что скрипт `/usr/local/bin/ai_metrics_collector.sh` исполняемый
3. Убедитесь, что AI Service доступен
4. Перезапустите Zabbix Agent: `sudo systemctl restart zabbix-agent`

### **Проблема: Неправильные данные**

**Решение:**
1. Проверьте формат данных в AI API
2. Убедитесь, что jq установлен: `sudo apt install jq`
3. Проверьте права доступа к скрипту

## 📋 Установка

### **1. Запуск обновленной интеграции:**

```bash
sudo ./scripts/ai_zabbix_integration_updated.sh
```

### **2. Проверка установки:**

```bash
# Проверка UserParameter'ов
grep "UserParameter.*ai\." /etc/zabbix/zabbix_agentd.conf

# Проверка скрипта
ls -la /usr/local/bin/ai_metrics_collector.sh

# Тест метрик
zabbix_agentd -t ai.status
```

## 🎯 Результат

После правильной настройки все AI метрики будут корректно передаваться в Zabbix:

- ✅ **15 AI метрик** с правильными ключами
- ✅ **4 альтернативных AI метрики** для совместимости
- ✅ **4 Relay метрики** для мониторинга сервера
- ✅ **4 System метрики** для системного мониторинга
- ✅ **3 Docker метрики** для контейнеров
- ✅ **3 Kubernetes метрики** для кластера

**Всего: 33 метрики готовы к передаче в Zabbix!** 🚀

