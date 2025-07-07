FROM python:3.11-slim

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Создание рабочей директории
WORKDIR /app

# Копирование requirements и установка зависимостей
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода
COPY alerting/ ./alerting/
COPY config/ ./config/

# Создание пользователя
RUN useradd -m -u 1000 siem && chown -R siem:siem /app
USER siem

# Запуск alert manager
CMD ["python", "alerting/alert_manager.py"] 