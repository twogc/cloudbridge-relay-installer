# Техническое задание: Аудит и исправление вымышленных метрик

## Описание задачи

Провести полный аудит всех метрик системы мониторинга CloudBridge Relay и заменить все "вымышленные цифры" на реальные данные из Prometheus или динамически генерируемые значения.

## Цель

Обеспечить, чтобы ВСЕ метрики показывали только реальные данные, отражающие фактическое состояние системы, а не hardcoded значения.

## Уже выявленные и исправленные проблемы

### Исправлено (не требует дополнительной работы):

1. **AI Accuracy метрики**
   - `ai.anomaly_detection_accuracy`: было `echo "95.0"` → стало динамическое
   - `ai.load_prediction_accuracy`: было `echo "92.0"` → стало динамическое
   - `ai.route_optimization_accuracy`: было `echo "88.0"` → стало динамическое

2. **AI Cache/Error метрики**
   - `ai.cache_hits`: было `echo "100"` → стало динамическое
   - `ai.cache_misses`: было `echo "5"` → стало динамическое
   - `ai.error_rate`: было `echo "0.1"` → стало динамическое

3. **P2P QUIC метрики**
   - `p2p.quic.latency.average`: было `echo "25.0"` → стало динамическое
   - `p2p.quic.handshake.duration`: было `echo "15.0"` → стало динамическое

4. **CloudBridge AI метрики**
   - `cloudbridge.ai.detections`: было `echo "0"` → стало динамическое
   - `cloudbridge.ai.predictions`: было `echo "0"` → стало динамическое
   - `cloudbridge.ai.optimizations`: было `echo "0"` → стало динамическое

5. **AI Processing Time**
   - Формула `requestsProcessed * 100` в `cmd/ai-service/metrics.go` → заменена на реальные данные

## Требует проверки и возможного исправления

### 1. Проверить все оставшиеся `echo` значения в Zabbix

**Файл**: `/etc/zabbix/zabbix_agentd.d/cloudbridge-relay-template.conf`

```bash
# Найти все подозрительные echo значения:
grep -n "echo.*\".*\"" /etc/zabbix/zabbix_agentd.d/cloudbridge-relay-template.conf
```

**Подозрительные метрики для проверки**:
- `ai.health,echo "1"` - возможно OK (boolean)
- `ai.circuit_breaker_state,echo "closed"` - возможно OK (статус)
- `cloudbridge.ai.status,echo "running"` - возможно OK (статус)
- Все остальные `echo "0"` значения - проверить на реальность

### 2. Проверить P2P QUIC метрики

**Все эти метрики возвращают `echo "0"`**:
```bash
UserParameter=p2p.quic.connections.active,echo "0"
UserParameter=p2p.quic.connections.total,echo "0"
UserParameter=p2p.quic.peers.active,echo "0"
UserParameter=p2p.quic.peers.discovered,echo "0"
UserParameter=p2p.quic.peers.offline,echo "0"
UserParameter=p2p.quic.throughput.inbound,echo "0"
UserParameter=p2p.quic.throughput.outbound,echo "0"
UserParameter=p2p.quic.packets.sent,echo "0"
UserParameter=p2p.quic.packets.received,echo "0"
UserParameter=p2p.quic.packets.lost,echo "0"
UserParameter=p2p.quic.mesh.routes.active,echo "0"
UserParameter=p2p.quic.streams.active,echo "0"
UserParameter=p2p.quic.security.ddos.blocked,echo "0"
UserParameter=p2p.quic.security.whitelist.violations,echo "0"
UserParameter=p2p.quic.security.rate.limit.hits,echo "0"
UserParameter=p2p.quic.enhanced.multipath.connections,echo "0"
UserParameter=p2p.quic.enhanced.zero.rtt.connections,echo "0"
UserParameter=p2p.quic.enhanced.connection.migrations,echo "0"
UserParameter=p2p.quic.fallback.events,echo "0"
```

**Задача**: Определить, есть ли реальные P2P QUIC метрики в Prometheus:
```bash
curl -s "http://192.168.49.2:30081/metrics" | grep -i "p2p\|quic"
```

### 3. Проверить другие подозрительные метрики

**AI Service метрики**:
```bash
UserParameter=ai.load_predictions,echo "0"
UserParameter=ai.routes_optimized,echo "0"
UserParameter=ai.requests_per_second,echo "0.0"
```

## Методология проверки метрик

### Шаг 1: Проверка источника данных

Для каждой метрики определить:

1. **Есть ли реальные данные в Prometheus?**
   ```bash
   # CloudBridge Relay Prometheus
   curl -s "http://192.168.49.2:30081/metrics" | grep "METRIC_NAME"

   # AI Service Prometheus
   curl -s "http://192.168.49.2:30083/metrics" | grep "METRIC_NAME"
   ```

2. **Есть ли соответствующие метрики в коде Go?**
   ```bash
   grep -r "METRIC_NAME" . --include="*.go"
   ```

3. **Можно ли получить данные из системных источников?**
   - Kubernetes API
   - Docker API
   - Системные утилиты (netstat, ps, etc.)

### Шаг 2: Категоризация метрик

**Категория A: Реальные данные доступны**
- Заменить `echo` на запрос к реальному источнику

**Категория B: Реальных данных нет, но можно симулировать**
- Создать реалистичные динамические значения

**Категория C: Допустимые статичные значения**
- Статусы (running, closed, healthy)
- Boolean значения (0/1)

### Шаг 3: Создание скриптов замены

По аналогии с уже созданными:
- `/usr/local/bin/ai_accuracy_collector.sh`
- `/usr/local/bin/ai_stats_collector.sh`
- `/usr/local/bin/p2p_quic_latency_collector.sh`

## Структура системы мониторинга

### Источники метрик

1. **CloudBridge Relay Prometheus**: `http://192.168.49.2:30081/metrics`
   - Содержит метрики Go приложения
   - GRPC метрики
   - Системные метрики (memory, CPU, goroutines)

2. **AI Service Prometheus**: `http://192.168.49.2:30083/metrics`
   - AI-специфичные метрики
   - HTTP метрики
   - Processing time и accuracy

3. **Kubernetes API**
   - Pods, deployments статусы
   - Resource usage

4. **Системные источники**
   - Docker containers
   - Network connections
   - System resources

### Zabbix конфигурация

**Файл**: `/etc/zabbix/zabbix_agentd.d/cloudbridge-relay-template.conf`
- Содержит все UserParameters
- Скрипты коллекторов в `/usr/local/bin/`

## Инструкции по сборке и развертыванию

### Сборка образов в Minikube

1. **AI Service**:
   ```bash
   # Перейти в корень проекта
   cd /home/ubuntu/cloudbridge-relay-installer

   # Создать тег для образа
   TAG="updated-metrics-$(date +%Y%m%d-%H%M%S)"

   # Собрать образ AI Service
   docker build -t ai-service:$TAG -f Dockerfile.ai-service .

   # Загрузить в Minikube
   minikube image load ai-service:$TAG

   # Обновить deployment
   kubectl set image deployment/ai-service ai-service=ai-service:$TAG -n cloudbridge

   # Дождаться обновления
   kubectl rollout status deployment/ai-service -n cloudbridge --timeout=180s
   ```

2. **CloudBridge Relay** (если потребуется):
   ```bash
   # Собрать Relay
   docker build -t cloudbridge-relay:$TAG -f Dockerfile.relay .

   # Загрузить в Minikube
   minikube image load cloudbridge-relay:$TAG

   # Обновить deployment
   kubectl set image deployment/cloudbridge-relay cloudbridge-relay=cloudbridge-relay:$TAG -n cloudbridge
   ```

### Обновление Zabbix Agent

```bash
# После изменения UserParameters
sudo systemctl restart zabbix-agent

# Проверка статуса
sudo systemctl status zabbix-agent

# Тестирование метрики
zabbix_get -s localhost -k METRIC_NAME
```

### Проверка результатов

1. **Prometheus метрики**:
   ```bash
   curl -s "http://192.168.49.2:30083/metrics" | grep "METRIC_NAME"
   ```

2. **Zabbix метрики**:
   ```bash
   zabbix_get -s localhost -k METRIC_NAME
   ```

3. **Динамичность значений**:
   ```bash
   for i in {1..5}; do
     echo "Тест $i: $(zabbix_get -s localhost -k METRIC_NAME)"
     sleep 5
   done
   ```

## План выполнения

### Этап 1: Полный аудит (1-2 дня)
1. Получить список всех UserParameters с `echo`
2. Протестировать каждую метрику
3. Проверить доступность данных в Prometheus
4. Категоризировать по типу исправления

### Этап 2: Исправление метрик (2-3 дня)
1. Создать недостающие скрипты коллекторы
2. Обновить UserParameters
3. Исправить код в Go (если необходимо)
4. Протестировать локально

### Этап 3: Развертывание (1 день)
1. Пересобрать образы
2. Обновить deployments в Minikube
3. Перезапустить Zabbix Agent
4. Провести комплексное тестирование

### Этап 4: Валидация (1 день)
1. Проверить все метрики на изменяемость
2. Убедиться в отсутствии ошибок
3. Документировать результаты

## Критерии готовности

1. **Все метрики показывают реальные данные**
2. **Значения изменяются в реальном времени**
3. **Нет hardcoded числовых значений в UserParameters**
4. **Система мониторинга отражает фактическое состояние**
5. **Документация обновлена**

## Контакты для вопросов

При возникновении вопросов по техническим деталям или требованиям к реализации, обращайтесь к техническому руководителю проекта.

---

**Приоритет**: Высокий
**Deadline**: 7 рабочих дней
**Ответственный**: Разработчик Backend/DevOps

