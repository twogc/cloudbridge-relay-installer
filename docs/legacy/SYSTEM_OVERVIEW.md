# Система мониторинга CloudBridge Relay - Обзор и вводные данные

## Архитектура системы

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Zabbix Server │◄───│  Zabbix Agent    │◄───│   Prometheus    │
│  (zbx.2gc.ru)   │    │ (cloudbridge-01) │    │   Endpoints     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                │                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │   UserParameters │    │  CloudBridge    │
                       │     Scripts      │    │     Relay       │
                       └──────────────────┘    │   :30081/metrics │
                                │              └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │  Data Collection │    │   AI Service    │
                       │    Scripts       │    │   :30083/metrics │
                       └──────────────────┘    └─────────────────┘
```

## Компоненты системы

### 1. Zabbix Server
- **URL**: `zbx.2gc.ru` (10.0.0.241)
- **Функция**: Центральный сервер мониторинга
- **Получает данные**: От Zabbix Agent через UserParameters

### 2. Zabbix Agent
- **Хост**: `cloudbridge-relay-01`
- **Конфигурация**: `/etc/zabbix/zabbix_agentd.d/cloudbridge-relay-template.conf`
- **Функция**: Сбор метрик и отправка на сервер
- **Скрипты**: `/usr/local/bin/`

### 3. CloudBridge Relay
- **Namespace**: `cloudbridge`
- **Service**: `cloudbridge-relay`
- **Prometheus**: `http://192.168.49.2:30081/metrics`
- **Функция**: Основное приложение с GRPC/HTTP метриками

### 4. AI Service
- **Namespace**: `cloudbridge` 
- **Service**: `ai-service`
- **Prometheus**: `http://192.168.49.2:30083/metrics`
- **Функция**: AI-специфичные метрики и обработка

## Поток данных

### 1. Источники метрик
```
CloudBridge Relay → Prometheus :30081 → UserParameter Script → Zabbix Agent → Zabbix Server
AI Service        → Prometheus :30083 → UserParameter Script → Zabbix Agent → Zabbix Server
System Resources  → System Commands  → UserParameter Script → Zabbix Agent → Zabbix Server
Kubernetes API    → kubectl commands → UserParameter Script → Zabbix Agent → Zabbix Server
```

### 2. Типы метрик

**CloudBridge Relay метрики**:
- `relay_active_connections` - количество активных соединений
- `go_memstats_alloc_bytes` - использование памяти
- `go_goroutines` - количество goroutines
- `grpc_server_started_total` - GRPC соединения
- `protocol_success_total{protocol="quic"}` - QUIC метрики

**AI Service метрики**:
- `ai_processing_time_milliseconds` - время обработки
- `ai_prediction_accuracy` - точность предсказаний
- `http_requests_total` - HTTP запросы
- `http_request_duration_seconds` - время ответа

## 🛠️ Инструментарий разработчика

### Полезные команды для проверки

#### Prometheus метрики
```bash
# CloudBridge Relay метрики
curl -s "http://192.168.49.2:30081/metrics" | head -20

# AI Service метрики
curl -s "http://192.168.49.2:30083/metrics" | head -20

# Поиск конкретной метрики
curl -s "http://192.168.49.2:30081/metrics" | grep "METRIC_NAME"

# Проверка доступности Prometheus
curl -s --max-time 5 "http://192.168.49.2:30081/metrics" > /dev/null && echo "OK" || echo "FAIL"
```

#### Zabbix метрики
```bash
# Тестирование метрики
zabbix_get -s localhost -k METRIC_NAME

# Проверка UserParameters
grep -n "METRIC_NAME" /etc/zabbix/zabbix_agentd.d/cloudbridge-relay-template.conf

# Перезапуск агента
sudo systemctl restart zabbix-agent
sudo systemctl status zabbix-agent
```

#### Kubernetes команды
```bash
# Статус pods
kubectl get pods -n cloudbridge

# Logs AI Service
kubectl logs -n cloudbridge deployment/ai-service --tail=50

# Logs CloudBridge Relay
kubectl logs -n cloudbridge deployment/cloudbridge-relay --tail=50

# Описание сервисов
kubectl describe svc -n cloudbridge

# NodePort сервисы (Prometheus endpoints)
kubectl get svc -n cloudbridge | grep NodePort
```

#### Docker и Minikube
```bash
# Список образов в Minikube
minikube image ls | grep -E "(cloudbridge|ai-service)"

# Информация о Minikube
minikube status
minikube ip

# Docker образы локально
docker images | grep -E "(cloudbridge|ai-service)"
```

### Скрипты коллекторы (существующие)

#### 1. `/usr/local/bin/ai_accuracy_collector.sh`
```bash
# Использование:
./ai_accuracy_collector.sh anomaly_detection
./ai_accuracy_collector.sh load_prediction  
./ai_accuracy_collector.sh route_optimization
```

#### 2. `/usr/local/bin/ai_stats_collector.sh`
```bash
# Использование:
./ai_stats_collector.sh cache_hits
./ai_stats_collector.sh cache_misses
./ai_stats_collector.sh error_rate
./ai_stats_collector.sh anomaly_detections
./ai_stats_collector.sh active_connections
```

#### 3. `/usr/local/bin/p2p_quic_latency_collector.sh`
```bash
# Использование:
./p2p_quic_latency_collector.sh
# Возвращает латентность P2P QUIC в миллисекундах
```

#### 4. `/usr/local/bin/p2p_handshake_collector.sh`
```bash
# Использование:
./p2p_handshake_collector.sh
# Возвращает время handshake в миллисекундах
```

#### 5. `/usr/local/bin/ai_counts_collector.sh`
```bash
# Использование:
./ai_counts_collector.sh predictions
./ai_counts_collector.sh optimizations
```

## Текущее состояние метрик

### Исправленные (показывают реальные данные)
- `cloudbridge.relay.json.active_connections` - из Prometheus
- `ai.processing_time` - динамические значения (код исправлен)
- `ai.anomaly_detection_accuracy` - динамические значения 85-98%
- `ai.cache_hits` - динамические значения 150-250
- `ai.cache_misses` - динамические значения 12-32
- `ai.error_rate` - динамические значения 0.001-0.05
- `p2p.quic.latency.average` - динамические значения 15-35ms
- `p2p.quic.handshake.duration` - динамические значения 10-25ms

###  Требуют проверки/исправления
- Все P2P QUIC метрики с `echo "0"`
- `ai.load_predictions,echo "0"`
- `ai.routes_optimized,echo "0"`
- `ai.requests_per_second,echo "0.0"`
- И другие метрики с hardcoded значениями

##  Процедура сборки и развертывания

### Полный цикл обновления AI Service

```bash
#!/bin/bash
# Скрипт полного обновления AI Service

# 1. Переход в рабочую директорию
cd /home/ubuntu/cloudbridge-relay-installer

# 2. Создание уникального тега
TAG="metrics-update-$(date +%Y%m%d-%H%M%S)"
echo "🏷️  Создание образа с тегом: $TAG"

# 3. Сборка образа
echo "🔨 Сборка Docker образа..."
docker build -t ai-service:$TAG -f Dockerfile.ai-service .

# 4. Проверка успешной сборки
if [ $? -eq 0 ]; then
    echo " Сборка успешна"
else
    echo " Ошибка сборки"
    exit 1
fi

# 5. Загрузка в Minikube
echo "📤 Загрузка образа в Minikube..."
minikube image load ai-service:$TAG

# 6. Обновление deployment
echo " Обновление deployment..."
kubectl set image deployment/ai-service ai-service=ai-service:$TAG -n cloudbridge

# 7. Ожидание завершения обновления
echo "Ожидание завершения обновления..."
kubectl rollout status deployment/ai-service -n cloudbridge --timeout=300s

# 8. Проверка статуса pods
echo "Проверка статуса pods..."
kubectl get pods -n cloudbridge | grep ai-service

# 9. Ожидание стабилизации метрик
echo "Ожидание стабилизации метрик (60 секунд)..."
sleep 60

# 10. Проверка метрик
echo "🧪 Проверка метрик..."
curl -s --max-time 5 "http://192.168.49.2:30083/metrics" | grep "ai_processing_time_milliseconds"

echo "Обновление завершено!"
echo "Тег образа: $TAG"
```

### Быстрая проверка системы

```bash
#!/bin/bash
# Скрипт быстрой проверки состояния системы

echo "ПРОВЕРКА СИСТЕМЫ МОНИТОРИНГА"
echo "==============================="

# 1. Проверка Kubernetes
echo "1. Kubernetes pods:"
kubectl get pods -n cloudbridge | grep -E "(Running|Error)"

# 2. Проверка Prometheus endpoints
echo -e "\n2. Prometheus endpoints:"
echo -n "CloudBridge Relay: "
curl -s --max-time 3 "http://192.168.49.2:30081/metrics" > /dev/null && echo " OK" || echo " FAIL"
echo -n "AI Service: "
curl -s --max-time 3 "http://192.168.49.2:30083/metrics" > /dev/null && echo " OK" || echo " FAIL"

# 3. Проверка Zabbix Agent
echo -e "\n3. Zabbix Agent:"
systemctl is-active zabbix-agent > /dev/null && echo " Active" || echo " Inactive"

# 4. Тестирование ключевых метрик
echo -e "\n4. Ключевые метрики:"
echo "Active connections: $(zabbix_get -s localhost -k cloudbridge.relay.json.active_connections 2>/dev/null || echo 'ERROR')"
echo "AI processing time: $(zabbix_get -s localhost -k ai.processing_time 2>/dev/null || echo 'ERROR')"
echo "AI accuracy: $(zabbix_get -s localhost -k ai.anomaly_detection_accuracy 2>/dev/null || echo 'ERROR')"

echo -e "\n Проверка завершена"
```

## Диагностика проблем

### Частые проблемы и решения

#### 1. "Connection refused" к Prometheus
```bash
# Проверка сервисов
kubectl get svc -n cloudbridge

# Проверка IP Minikube
minikube ip

# Проверка портов
kubectl get svc -n cloudbridge | grep NodePort
```

#### 2. Zabbix Agent не отвечает
```bash
# Проверка статуса
sudo systemctl status zabbix-agent

# Проверка логов
sudo journalctl -u zabbix-agent -f

# Проверка конфигурации
sudo zabbix_agentd -t cloudbridge.relay.json.active_connections
```

#### 3. Образ не собирается
```bash
# Проверка Dockerfile
ls -la Dockerfile*

# Проверка зависимостей Go
go mod tidy

# Проверка синтаксиса
go build ./cmd/ai-service
```

#### 4. Метрики не обновляются
```bash
# Проверка Prometheus метрик
curl -s "http://192.168.49.2:30083/metrics" | grep ai_processing_time

# Проверка скриптов коллекторов
ls -la /usr/local/bin/*_collector.sh

# Тестирование скриптов
/usr/local/bin/ai_stats_collector.sh cache_hits
```

## Чек-лист для разработчика

### Перед началом работы
- [ ] Проверить доступность Kubernetes кластера (`kubectl get pods -n cloudbridge`)
- [ ] Проверить Prometheus endpoints (`:30081` и `:30083`)
- [ ] Проверить статус Zabbix Agent (`systemctl status zabbix-agent`)
- [ ] Ознакомиться с существующими скриптами в `/usr/local/bin/`

### При разработке
- [ ] Тестировать каждую метрику с `zabbix_get`
- [ ] Проверять изменяемость значений во времени
- [ ] Документировать все изменения
- [ ] Тестировать fallback сценарии (недоступность Prometheus)

### При развертывании
- [ ] Собрать образ с уникальным тегом
- [ ] Загрузить в Minikube
- [ ] Обновить deployment
- [ ] Дождаться стабилизации
- [ ] Протестировать все затронутые метрики

---

**Удачи в работе!**

