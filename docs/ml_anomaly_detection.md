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

## Структура файлов
- `ml/anomaly_detector.py` — основная логика ML
- `ml/train_anomaly_detector.py` — обучение моделей на CSV
- `ml/anomaly_api.py` — REST API (FastAPI)
- `ml/historical_data_example.csv` — пример данных
- `ml/requirements.txt` — зависимости

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

## Советы
- Для production используйте отдельное окружение (venv).
- Регулярно обновляйте зависимости и переобучайте модели на новых данных.
- Для расширения — добавьте новые признаки, типы аномалий, интеграцию с алертингом. 