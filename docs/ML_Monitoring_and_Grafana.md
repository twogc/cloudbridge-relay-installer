# Техническое задание: Мониторинг ML-сервиса через Prometheus и Grafana

## 1. Экспорт ML-метрик в Prometheus

- В ML FastAPI сервисе реализовать endpoint `/metrics` с помощью библиотеки `prometheus_client`.
- Экспортировать следующие метрики:
  - `anomaly_detected_total` — количество обнаруженных аномалий
  - `error_rate_last` — последнее значение error_rate
  - `detect_requests_total` — количество запросов к /detect

## 2. Настройка Prometheus

- В файл `prometheus/prometheus.yml` добавить job для ML-сервиса:
  ```yaml
  - job_name: 'ml-anomaly-api'
    static_configs:
      - targets: ['ml:8000']  # или другой host:port
  ```
- Перезапустить Prometheus после изменения конфига.

## 3. Добавление панелей в Grafana

- В существующий dashboard (или новый) добавить панели:
  - **Статистика аномалий:**  
    - Тип: Stat/Graph  
    - Запрос: `anomaly_detected_total`
  - **Error Rate:**  
    - Тип: Gauge/Graph  
    - Запрос: `error_rate_last`
  - **Количество запросов:**  
    - Тип: Stat  
    - Запрос: `detect_requests_total`
- Сохранить dashboard в формате JSON в папку `deploy/grafana/dashboards/` (например, `ml_anomaly_dashboard.json`).

## 4. Документация

- Описать в этом файле:
  - Как работает экспорт метрик.
  - Как настроить Prometheus и Grafana для сбора и визуализации ML-метрик.
  - Примеры запросов для панелей Grafana.
  - Пример скриншота дашборда (по возможности).

---

### Пример панели для Grafana (Stat):

- **Title:** Detected Anomalies
- **Query:** `anomaly_detected_total`
- **Datasource:** Prometheus

---

### Пример scrape-конфига для Prometheus

```yaml
scrape_configs:
  - job_name: 'ml-anomaly-api'
    static_configs:
      - targets: ['ml:8000']
```

---

### Пример запроса для панели Error Rate

- **Query:** `error_rate_last`
- **Type:** Gauge/Graph

---

### Пример запроса для панели Количество запросов

- **Query:** `detect_requests_total`
- **Type:** Stat

---

### Примечания
- Названия метрик и порт сервиса уточняйте по фактической реализации.
- Для production рекомендуется добавить алерты на превышение порогов error_rate и числа аномалий. 

---

## Реализация и настройка мониторинга ML метрик

### Экспорт метрик в FastAPI ML сервисе

В файле `ml/anomaly_api.py` реализован FastAPI сервис с endpoint `/metrics` с помощью библиотеки `prometheus_client`. Экспортируются следующие метрики:
- `anomaly_detected_total` — количество обнаруженных аномалий (Counter)
- `error_rate_last` — последнее значение error_rate (Gauge)
- `detect_requests_total` — количество запросов к /detect (Counter)

Пример кода:
```python
from prometheus_client import Counter, Gauge, generate_latest, CONTENT_TYPE_LATEST
from fastapi import Response

anomaly_counter = Counter('anomaly_detected_total', 'Total detected anomalies')
error_rate_gauge = Gauge('error_rate_last', 'Last error rate value')
requests_total = Counter('detect_requests_total', 'Total /detect requests')

@app.get("/metrics")
def metrics():
    return Response(generate_latest(), media_type=CONTENT_TYPE_LATEST)
```

### Настройка Prometheus

В файл `prometheus/prometheus.yml` добавлен job для ML-сервиса:
```yaml
  - job_name: 'ml-anomaly-api'
    static_configs:
      - targets: ['ml:8000']
    metrics_path: '/metrics'
    scheme: 'http'
    scrape_interval: 15s
    scrape_timeout: 5s
```
После изменения конфига необходимо перезапустить Prometheus.

### Настройка Grafana

Создайте новый dashboard или добавьте панели в существующий. Примеры панелей:
- **Статистика аномалий**
  - Тип: Stat/Graph
  - Запрос: `anomaly_detected_total`
- **Error Rate**
  - Тип: Gauge/Graph
  - Запрос: `error_rate_last`
- **Количество запросов**
  - Тип: Stat
  - Запрос: `detect_requests_total`

Сохраните dashboard в формате JSON в папку `deploy/grafana/dashboards/` (например, `ml_anomaly_dashboard.json`).

#### Примеры запросов для панелей Grafana
- Detected Anomalies: `anomaly_detected_total`
- Error Rate: `error_rate_last`
- Количество запросов: `detect_requests_total`

#### Пример скриншота дашборда
_Скриншот будет добавлен после создания и наполнения дашборда._ 

---

## Автоматическая загрузка dashboard в Grafana

Для автоматической загрузки dashboard используется механизм provisioning Grafana:
- provisioning-файл: `grafana/provisioning/dashboards/default.yaml`
- путь для dashboard'ов: `deploy/grafana/dashboards/`
- docker-compose монтирует provisioning и dashboard'ы в контейнер Grafana:
  - `./grafana/provisioning/dashboards:/etc/grafana/provisioning/dashboards`
  - `./deploy/grafana/dashboards:/var/lib/grafana/dashboards`

После перезапуска Grafana все dashboard'ы из папки автоматически появляются в интерфейсе.

**Порядок действий для production:**
1. Убедитесь, что provisioning-файл и dashboard лежат в нужных папках.
2. Перезапустите Grafana: `docker restart <container_id>`
3. Проверьте появление dashboard в UI Grafana.

--- 