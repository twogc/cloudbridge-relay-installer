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