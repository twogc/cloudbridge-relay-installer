# НЕДЕЛЯ 1, ЗАДАЧА 1.1: УСТАНОВКА И ЗАПУСК ВСЕХ СЕРВИСОВ
**Duration**: 4 часов
**Status**: IN PROGRESS
**Deadline**: 2025-10-28 (Today)

---

## ЦЕЛЬ
Запустить все 11 CloudBridge сервисов локально в Docker контейнерах и убедиться, что они работают.

---

## СЕРВИСЫ КОТОРЫЕ БУДУТ ЗАПУЩЕНЫ

### Infrastructure (5 сервисов)
```
✓ NATS JetStream      (4222, 8222)
✓ Redis              (6379)
✓ VictoriaMetrics    (8428) ← Главный для Dashboard
✓ Prometheus         (9090)
✓ Grafana            (3001)
```

### CloudBridge Core (5 сервисов)
```
✓ Control Plane      (8080, 9091) ← Zitadel, Cadence
✓ DDoS Protection    (8081) ← ML-based
✓ AI Service         (8083) ← Routing & predictions
✓ Relay              (9092, 4433, 8082) ← QUIC, P2P
✓ Dashboard BFF      (8084) ← Backend for Web UI
```

### Frontend (1 сервис)
```
✓ Dashboard UI       (3000) ← React/Next.js
```

### Monitoring Exporters (2 сервиса)
```
✓ BGP Blackhole      (9101)
✓ Scrubbing          (9102)
```

---

## БЫСТРЫЙ СТАРТ (5 МИНУТ)

### Шаг 1: Перейти в каталог с docker-compose
```bash
cd /Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories
```

### Шаг 2: Запустить все сервисы
```bash
docker-compose up -d
```

### Шаг 3: Дождаться инициализации (60 сек)
```bash
# Смотреть статус контейнеров
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
```

### Шаг 4: Проверить здоровье сервисов
```bash
# Все контейнеры должны быть в статусе "Up"
docker compose ps

# Или детальная проверка
curl -s http://localhost:8428/health && echo "VictoriaMetrics OK"
curl -s http://localhost:9090/-/healthy && echo "Prometheus OK"
curl -s http://localhost:3001/api/health && echo "Grafana OK"
```

---

## ПРОВЕРКА СЕРВИСОВ (URL'ы)

После запуска, проверьте эти URL'ы в браузере:

### Monitoring & Metrics
| Сервис | URL | Логин | Пароль |
|--------|-----|-------|--------|
| **Grafana** | http://localhost:3001 | admin | admin |
| **Prometheus** | http://localhost:9090 | - | - |
| **VictoriaMetrics** | http://localhost:8428 | - | - |
| **NATS** | http://localhost:8222 | - | - |

### CloudBridge Services
| Сервис | Health Check | Port |
|--------|--------------|------|
| **Control Plane** | http://localhost:8080/healthz | 8080 |
| **DDoS Protection** | http://localhost:8081/healthz | 8081 |
| **AI Service** | http://localhost:8083/health/ | 8083 |
| **Relay** | http://localhost:9092/healthz | 9092 |
| **Dashboard BFF** | http://localhost:8084/healthz | 8084 |
| **Dashboard UI** | http://localhost:3000 | 3000 |

---

##  ДЕТАЛЬНЫЕ ШАГИ

### Шаг 1: Подготовка
```bash
# Убедиться что Docker работает
docker --version
docker-compose --version
docker ps  # Should show no errors

# Убедиться что нет конфликтов портов
lsof -i :8428  # VictoriaMetrics port
lsof -i :6379  # Redis port
lsof -i :3000  # Dashboard port
# Если порты заняты, остановите старые контейнеры
```

### Шаг 2: Запуск сервисов
```bash
# Перейти в каталог
cd /Users/maxlanies/Git/2GC/cloudbridge-relay-installer/oss-repositories

# Запустить фоном
docker-compose up -d

# Или запустить с логами (в отдельном терминале)
docker-compose up
```

### Шаг 3: Мониторинг запуска
```bash
# Смотреть логи определенного сервиса
docker-compose logs -f redis          # Redis logs
docker-compose logs -f victoriametrics # VictoriaMetrics logs
docker-compose logs -f control-plane   # Control Plane logs
docker-compose logs -f dashboard-backend # Dashboard BFF logs

# Выход: Ctrl+C
```

### Шаг 4: Проверка здоровья

#### Метод 1: Через Docker
```bash
# Проверить статус контейнеров
docker ps -a --format "table {{.Names}}\t{{.Status}}"

# Ожидаемый результат:
# cloudbridge-nats               Up 2 minutes (healthy)
# cloudbridge-redis              Up 2 minutes (healthy)
# cloudbridge-victoriametrics    Up 2 minutes (healthy)
# cloudbridge-prometheus         Up 2 minutes (healthy)
# cloudbridge-grafana            Up 2 minutes (healthy)
# cloudbridge-control-plane      Up 1 minute  (healthy)
# cloudbridge-ddos-protection    Up 1 minute  (healthy)
# cloudbridge-ai-service         Up 1 minute  (healthy)
# cloudbridge-relay              Up 1 minute  (healthy)
# cloudbridge-dashboard-backend  Up 1 minute  (healthy)
# cloudbridge-dashboard-frontend Up 1 minute  (healthy)
```

#### Метод 2: Через HTTP запросы
```bash
#!/bin/bash

echo "Checking VictoriaMetrics..."
curl -s http://localhost:8428/health && echo "✓ OK" || echo "✗ FAIL"

echo "Checking Prometheus..."
curl -s http://localhost:9090/-/healthy && echo "✓ OK" || echo "✗ FAIL"

echo "Checking Grafana..."
curl -s http://localhost:3001/api/health && echo "✓ OK" || echo "✗ FAIL"

echo "Checking Control Plane..."
curl -s http://localhost:8080/healthz && echo "✓ OK" || echo "✗ FAIL"

echo "Checking DDoS Protection..."
curl -s http://localhost:8081/healthz && echo "✓ OK" || echo "✗ FAIL"

echo "Checking AI Service..."
curl -s http://localhost:8083/health/ && echo "✓ OK" || echo "✗ FAIL"

echo "Checking Relay..."
curl -s http://localhost:9092/healthz && echo "✓ OK" || echo "✗ FAIL"

echo "Checking Dashboard BFF..."
curl -s http://localhost:8084/healthz && echo "✓ OK" || echo "✗ FAIL"
```

#### Метод 3: Через Redis CLI
```bash
# Проверить Redis
docker exec cloudbridge-redis redis-cli ping
# Ожидаемый результат: PONG
```

---

## 🐛 TROUBLESHOOTING

### Проблема: Docker daemon не работает
```bash
# Решение: Запустить Docker Desktop (macOS/Windows) или:
sudo systemctl start docker  # Linux
```

### Проблема: Контейнер не запускается
```bash
# Проверить логи
docker-compose logs <service_name>

# Пересоздать контейнер
docker-compose down <service_name>
docker-compose up -d <service_name>
```

### Проблема: Порт уже в использовании
```bash
# Найти какой процесс использует порт
lsof -i :8428  # VictoriaMetrics example

# Остановить старые контейнеры
docker-compose down

# Или убить процесс если это не контейнер
kill -9 <PID>
```

### Проблема: Out of memory
```bash
# Увеличить Docker memory limit
# macOS/Windows: Docker Desktop → Preferences → Resources → Memory

# Linux: Check current usage
docker stats

# Clean up unused images/containers
docker system prune -a
```

### Проблема: VictoriaMetrics пуст (нет метрик)
```bash
# Это нормально на начале - данные будут приходить когда
# сервисы начнут отправлять метрики в Prometheus

# Проверить что Prometheus scraping
curl -s http://localhost:9090/api/v1/targets | jq

# Проверить что VictoriaMetrics получает данные
curl -s 'http://localhost:8428/api/v1/labels' | jq
```

---

## ОЖИДАЕМЫЕ РЕЗУЛЬТАТЫ

### После успешного запуска:
```
✓ 11+ контейнеров работает
✓ Все health checks passing
✓ VictoriaMetrics доступен (метрики будут заполняться постепенно)
✓ Grafana доступен с дефолтными дашбордами
✓ Dashboard UI доступен на localhost:3000
✓ Все сервисы могут общаться между собой через сеть cloudbridge
```

### Доступные endpoints:
```
Infrastructure:
  NATS:          nats://localhost:4222
  Redis:         localhost:6379
  VictoriaMetrics: http://localhost:8428
  Prometheus:    http://localhost:9090
  Grafana:       http://localhost:3001

CloudBridge:
  Control Plane:  http://localhost:8080
  DDoS:          http://localhost:8081
  AI Service:    http://localhost:8083
  Relay:         http://localhost:9092 (HTTP), localhost:4433 (QUIC)
  Dashboard BFF: http://localhost:8084
  Dashboard UI:  http://localhost:3000
```

---

## СЛЕДУЮЩИЕ ШАГИ (После успешного запуска)

### Сразу после:
1. Зафиксировать что все сервисы работают (скриншот docker ps)
2. Проверить VictoriaMetrics доступен
3. Открыть Grafana и Dashboard
4. Отметить Task 1.1 как завершенную

### Переходить на Task 1.2:
1. Проверить API endpoints (Dashboard BFF)
2. Расширить API структуру
3. Добавить новые endpoints

---

## ⏱️ ВРЕМЕННАЯ ШКАЛА

```
T+0:00    Docker-compose up -d
T+1:00    Все контейнеры должны быть "healthy"
T+2:00    Все health checks passing
T+3:00    Готово к тестированию
T+4:00    Task 1.1 завершена
```

---

## ПОЛЕЗНЫЕ КОМАНДЫ

```bash
# Управление всеми сервисами
docker-compose up -d           # Запустить все
docker-compose down            # Остановить все
docker-compose restart         # Перезагрузить все
docker-compose logs -f         # Смотреть все логи

# Управление конкретным сервисом
docker-compose up -d redis
docker-compose down relay
docker-compose logs -f control-plane

# Проверка
docker-compose ps              # Статус контейнеров
docker exec <name> bash        # Зайти в контейнер
docker stats                   # CPU/Memory usage

# Очистка
docker-compose down -v         # Down + удалить volumes
docker system prune -a         # Удалить неиспользуемое
```

---

## CHECKLIST

- [ ] Docker установлен и работает
- [ ] Docker Compose установлен и работает
- [ ] Перешел в каталог `/oss-repositories`
- [ ] Выполнил `docker-compose up -d`
- [ ] Дождался инициализации (60-90 сек)
- [ ] Проверил `docker ps` - все контейнеры "Up"
- [ ] Проверил здоровье VictoriaMetrics (curl)
- [ ] Проверил здоровье Control Plane (curl)
- [ ] Проверил здоровье Relay (curl)
- [ ] Проверил здоровье Dashboard BFF (curl)
- [ ] Открыл Grafana (localhost:3001)
- [ ] Открыл Dashboard UI (localhost:3000)
- [ ] Все сервисы работают корректно

---

**Task Status**: IN PROGRESS
**Estimated Time**: 4 hours (1 hour setup + 3 hours validation & troubleshooting)
**Next Task**: Task 1.2 - API Structure & Expansion
