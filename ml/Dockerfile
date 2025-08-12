# -----------------------------
# CloudBridge ML Anomaly Service
# -----------------------------
#
# Рекомендуемые ресурсы:
#   CPU: 1-2 ядра
#   RAM: 1-2 ГБ
#   Disk: 1 ГБ+
#   Port: 8000 (по умолчанию)
#
# Для production используйте ограничения:
#   --memory=1g --cpus=1.0
# -----------------------------

FROM python:3.10-slim

WORKDIR /app

# Системные зависимости (если нужны)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential gcc && \
    rm -rf /var/lib/apt/lists/*

# Копируем requirements и устанавливаем зависимости
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Копируем исходники и модель
COPY . .

# Открываем порт
EXPOSE 8000

# Переменные окружения (можно переопределять при запуске)
ENV MODEL_PATH=models.joblib
ENV REDIS_HOST=localhost
ENV REDIS_PORT=6379

# Запуск сервиса
CMD ["uvicorn", "anomaly_api:app", "--host", "0.0.0.0", "--port", "8000"] 