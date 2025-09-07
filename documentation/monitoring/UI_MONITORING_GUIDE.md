# 🖥️ UI Monitoring Guide - CloudBridge Relay

## 📋 Обзор

Этот документ описывает, как мониторить Cadence и Minikube через пользовательские интерфейсы (UI) в CloudBridge Relay.

---

## 🚀 Быстрый старт

### **Автоматический запуск всех UI:**
```bash
# Запуск всех интерфейсов мониторинга
./scripts/monitoring-ui.sh start

# Открытие всех интерфейсов в браузере
./scripts/monitoring-ui.sh open

# Проверка статуса
./scripts/monitoring-ui.sh status

# Остановка всех интерфейсов
./scripts/monitoring-ui.sh stop
```

---

## 📊 Доступные UI интерфейсы

### **1. 🎨 Grafana Dashboard**
- **URL**: http://localhost:3000
- **Описание**: Основной дашборд для визуализации метрик
- **Порт**: 3000
- **Функции**:
  - Дашборды для Cadence Workflow Engine
  - Дашборды для Minikube Infrastructure
  - Визуализация метрик Prometheus
  - Настраиваемые графики и алерты

### **2. 📈 Prometheus Metrics**
- **URL**: http://localhost:9090
- **Описание**: Интерфейс для просмотра метрик и алертов
- **Порт**: 9090
- **Функции**:
  - Просмотр всех метрик системы
  - Выполнение PromQL запросов
  - Просмотр алертов и правил
  - Статус targets и scraping

### **3. ⚙️ Cadence CLI Monitoring**
- **Описание**: Мониторинг Cadence Workflow Engine через CLI
- **Доступ**: `./scripts/cadence-monitor.sh`
- **Функции**:
  - Просмотр доменов и их статуса
  - Мониторинг workflow executions
  - Просмотр task lists
  - Информация о кластере
  - Логи и метрики

### **4. 🗄️ Cassandra Metrics**
- **URL**: http://localhost:9500
- **Описание**: Метрики базы данных Cassandra
- **Порт**: 9500
- **Функции**:
  - Метрики производительности
  - Статус кластера
  - Информация о таблицах
  - Мониторинг запросов

### **5. 🌐 CloudBridge Relay**
- **URL**: http://localhost:8084
- **Описание**: Основное приложение CloudBridge Relay
- **Порт**: 8084
- **Функции**:
  - Health check endpoint
  - Метрики приложения
  - Статус сервиса
  - API endpoints

---

## 🎯 Мониторинг Cadence

### **Через Grafana Dashboard:**

#### **1. Cadence Workflow Engine Dashboard:**
- **Request Rate**: Количество запросов в секунду
- **Request Latency**: Время отклика запросов
- **Workflow Executions**: Активные выполнения workflow
- **Error Rate**: Частота ошибок

#### **2. Ключевые метрики:**
```promql
# Количество запросов
rate(cadence_requests[5m])

# Время отклика
histogram_quantile(0.95, rate(cadence_request_latency_bucket[5m]))

# Выполнения workflow
cadence_workflow_executions

# Частота ошибок
rate(cadence_errors[5m])
```

### **Через Cadence CLI:**

#### **1. Мониторинг доменов:**
```bash
# Просмотр всех доменов
./scripts/cadence-monitor.sh domains

# Описание конкретного домена
kubectl exec -n cloudbridge cadence-xxx -- cadence --address 10.244.0.172:7933 --domain cadence-system domain describe
```

#### **2. Workflow Executions:**
```bash
# Просмотр workflow executions
./scripts/cadence-monitor.sh workflows

# Список workflow в домене
kubectl exec -n cloudbridge cadence-xxx -- cadence --address 10.244.0.172:7933 --domain cadence-system workflow list
```

#### **3. Task Lists и кластер:**
```bash
# Просмотр task lists
./scripts/cadence-monitor.sh tasklists

# Информация о кластере
./scripts/cadence-monitor.sh cluster

# Логи Cadence
./scripts/cadence-monitor.sh logs
```

---

## 🖥️ Мониторинг Minikube

### **Через Grafana Dashboard:**

#### **1. Minikube Infrastructure Dashboard:**
- **CPU Usage**: Использование процессора
- **Memory Usage**: Использование памяти
- **Running Pods**: Количество запущенных подов
- **Disk Usage**: Использование дискового пространства

#### **2. Ключевые метрики:**
```promql
# Использование CPU
100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)

# Использование памяти
(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100

# Количество подов
kubelet_running_pods

# Использование диска
node_filesystem_avail_bytes{mountpoint="/"}
```

### **Через Prometheus:**

#### **1. Kubernetes Metrics:**
- Статус нод кластера
- Информация о подах
- Использование ресурсов
- Сетевые метрики

#### **2. System Metrics:**
- CPU, Memory, Disk
- Network I/O
- File system usage
- Process metrics

---

## 🛠️ Управление UI интерфейсами

### **Команды скрипта:**

```bash
# Запуск всех интерфейсов
./scripts/monitoring-ui.sh start

# Запуск конкретного интерфейса
./scripts/monitoring-ui.sh grafana
./scripts/monitoring-ui.sh prometheus
./scripts/monitoring-ui.sh cassandra
./scripts/monitoring-ui.sh relay

# Мониторинг Cadence через CLI
./scripts/cadence-monitor.sh domains
./scripts/cadence-monitor.sh workflows
./scripts/cadence-monitor.sh logs

# Открытие всех в браузере
./scripts/monitoring-ui.sh open

# Проверка статуса
./scripts/monitoring-ui.sh status

# Остановка всех
./scripts/monitoring-ui.sh stop

# Справка
./scripts/monitoring-ui.sh help
```

### **Ручное управление port-forward:**

```bash
# Grafana
kubectl port-forward -n cloudbridge-monitoring svc/grafana 3000:3000

# Prometheus
kubectl port-forward -n cloudbridge-monitoring svc/prometheus-optimized-service 9090:9090

# Cadence
kubectl port-forward -n cloudbridge svc/cadence-service 7934:7934

# Cassandra
kubectl port-forward -n cloudbridge svc/cassandra-service 9500:9500

# CloudBridge Relay
kubectl port-forward -n cloudbridge svc/cloudbridge-relay 8084:8084
```

---

## 📊 Дашборды Grafana

### **1. Cadence Workflow Engine Dashboard:**
- **Request Rate**: График запросов в реальном времени
- **Request Latency**: 95-й и 50-й процентили времени отклика
- **Workflow Executions**: Количество активных выполнений
- **Error Rate**: Частота ошибок по типам

### **2. Minikube Infrastructure Dashboard:**
- **CPU Usage**: Использование процессора в процентах
- **Memory Usage**: Использование памяти в процентах
- **Running Pods**: Количество запущенных подов
- **Disk Usage**: Доступное и общее дисковое пространство

### **3. Настройка дашбордов:**
- Автоматическое обновление каждые 30 секунд
- Настраиваемые временные диапазоны
- Интерактивные графики
- Экспорт данных

---

## 🔍 Диагностика и troubleshooting

### **Проверка доступности UI:**

```bash
# Проверка статуса подов
kubectl get pods -n cloudbridge-monitoring
kubectl get pods -n cloudbridge

# Проверка сервисов
kubectl get services -n cloudbridge-monitoring
kubectl get services -n cloudbridge

# Проверка port-forward
./scripts/monitoring-ui.sh status
```

### **Частые проблемы:**

#### **1. UI не открывается:**
```bash
# Проверка port-forward
ps aux | grep "kubectl port-forward"

# Перезапуск port-forward
./scripts/monitoring-ui.sh stop
./scripts/monitoring-ui.sh start
```

#### **2. Нет данных в Grafana:**
```bash
# Проверка Prometheus
curl http://localhost:9090/api/v1/targets

# Проверка метрик
curl http://localhost:9090/api/v1/query?query=up
```

#### **3. Cadence UI недоступен:**
```bash
# Проверка логов Cadence
kubectl logs -n cloudbridge cadence-xxx

# Проверка сервиса
kubectl get svc -n cloudbridge cadence-service
```

---

## 📈 Лучшие практики

### **1. Мониторинг:**
- Регулярно проверяйте дашборды Grafana
- Настройте алерты для критических метрик
- Мониторьте использование ресурсов Minikube
- Отслеживайте производительность Cadence

### **2. Производительность:**
- Используйте фильтры в Prometheus для оптимизации запросов
- Настройте автоматическое обновление дашбордов
- Мониторьте размер хранилища Prometheus
- Регулярно очищайте старые данные

### **3. Безопасность:**
- Ограничьте доступ к UI интерфейсам
- Используйте HTTPS для production
- Регулярно обновляйте компоненты
- Мониторьте доступ к интерфейсам

---

## 🎯 Заключение

**UI мониторинг CloudBridge Relay обеспечивает полную видимость системы:**

### **Доступные интерфейсы:**
- ✅ **Grafana** - визуализация и дашборды
- ✅ **Prometheus** - метрики и алерты
- ✅ **Cadence Web UI** - управление workflow
- ✅ **Cassandra Metrics** - мониторинг БД
- ✅ **CloudBridge Relay** - основное приложение

### **Ключевые возможности:**
- 🚀 **Автоматический запуск** всех интерфейсов
- 📊 **Специализированные дашборды** для Cadence и Minikube
- 🔍 **Детальный мониторинг** всех компонентов
- 🛠️ **Простое управление** через скрипт

**Система готова для комплексного мониторинга через UI!** 🎉

---

*Руководство создано: 7 сентября 2025*  
*Версия: UI Monitoring v1.0*  
*Статус: ✅ Production Ready*
