# ML Anomaly Detection Service

## Назначение
Сервис предназначен для обнаружения аномалий в сетевом трафике, безопасности и производительности на основе исторических данных. Реализован на Python с использованием scikit-learn, FastAPI и pandas.

---

## Быстрый старт

### 1. Установка зависимостей
```bash
pip install -r ml/requirements.txt
```

### 2. Подготовка исторических данных
- Формат: CSV, все необходимые поля (см. пример `ml/historical_data_example.csv`).
- Минимум 100 строк для обучения.

### 3. Обучение моделей
```bash
cd ml
python3 train_anomaly_detector.py historical_data_example.csv
```
- После обучения появится файл `models.joblib`.

### 4. Запуск REST API
```bash
uvicorn anomaly_api:app --reload
```
- API доступен на http://127.0.0.1:8000
- Swagger UI: http://127.0.0.1:8000/docs

### 5. Пример запроса к API
```json
POST /detect
{
  "timestamp": "2025-07-08T12:00:00",
  "bytes_sent": 1000,
  "bytes_received": 2000,
  "packets_sent": 10,
  "packets_received": 20,
  "connection_duration": 30,
  "requests_per_second": 5,
  "error_rate": 0.01,
  "user_id_hash": 12345,
  "session_duration": 300,
  "requests_in_session": 50,
  "unique_endpoints": 5,
  "failed_auth_attempts": 0,
  "suspicious_ips": 0,
  "rate_limit_violations": 0,
  "ssl_errors": 0
}
```

---

## Переменные окружения и конфигурирование

- `MODEL_PATH` — путь к файлу модели (по умолчанию `models.joblib`)
- `REDIS_HOST`, `REDIS_PORT` — параметры Redis (по умолчанию localhost:6379)
- `IFOREST_CONTAMINATION`, `IFOREST_N_ESTIMATORS` — параметры IsolationForest
- `MLP_HIDDEN_LAYERS`, `MLP_MAX_ITER` — параметры MLPRegressor

Можно задать через `.env` или переменные окружения (см. пример `ml/.env.example`).

---

## Запуск в Docker

### Dockerfile
В папке `ml/` есть готовый Dockerfile. Пример сборки и запуска:
```bash
cd ml
# Сборка
sudo docker build -t anomaly-ml .
# Запуск
sudo docker run --rm -p 8000:8000 \
  -e MODEL_PATH=models.joblib \
  -e REDIS_HOST=localhost \
  -e REDIS_PORT=6379 \
  anomaly-ml
```

### Docker Compose
Пример файла `ml/docker-compose.yml`:
```yaml
version: '3.8'
services:
  anomaly-ml:
    build: .
    ports:
      - "8000:8000"
    environment:
      - MODEL_PATH=models.joblib
      - REDIS_HOST=localhost
      - REDIS_PORT=6379
    volumes:
      - ./models.joblib:/app/models.joblib:ro
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1.0'
          memory: 1g
```

---

## Рекомендации по ресурсам
- **Минимум:** 1 CPU, 1 ГБ RAM, 100 МБ диск
- **Рекомендуется:** 2 CPU, 2 ГБ RAM, 1 ГБ+ диск
- **Порт:** 8000 (можно изменить)
- **GPU не требуется**

---

## Обновление и переобучение модели
- Для обновления модели обучите новую модель (`train_anomaly_detector.py`) и замените файл `models.joblib`.
- Сервис автоматически подхватит новую модель при перезапуске контейнера или сервиса.
- Для hot-swap без остановки — реализуйте atomic-rename файла модели и health-check.

---

## Примеры ошибок и валидации
- Все числовые поля должны быть неотрицательными, error_rate — от 0 до 1.
- Пример ошибки:
```json
{"detail":[{"type":"int_parsing","loc":["body","bytes_sent"],"msg":"Input should be a valid integer, unable to parse string as an integer","input":"not_a_number"}]}
```
- Если отсутствует обязательное поле:
```json
{"detail":[{"type":"missing","loc":["body","packets_sent"],"msg":"Field required","input":{...}}]}
```

---

## Интеграция с backend
Пример на Python (requests):
```python
import requests
url = "http://127.0.0.1:8000/detect"
data = { ... }  # как в примере выше
response = requests.post(url, json=data)
print(response.json())
```

---

## Тестирование
- Unit-тесты: `pytest ml/test_anomaly_detector.py`
- Swagger UI: ручная проверка через http://127.0.0.1:8000/docs

---

## Советы и развитие
- Для production используйте отдельное окружение (venv, Docker).
- Регулярно обновляйте зависимости и переобучайте модели на новых данных.
- Для расширения — добавьте новые признаки, типы аномалий, интеграцию с алертингом (Prometheus, Alertmanager).
- Возможна автоматизация переобучения (скрипт, endpoint, cron). 