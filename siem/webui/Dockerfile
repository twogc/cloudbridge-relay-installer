FROM python:3.11-slim

# Установка системных зависимостей
RUN apt-get update && apt-get install -y \
    curl \
    wget \
    nginx \
    && rm -rf /var/lib/apt/lists/*

# Создание рабочей директории
WORKDIR /app

# Копирование requirements и установка зависимостей
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копирование кода
COPY webui/ ./webui/
COPY config/ ./config/

# Настройка nginx
COPY webui/nginx.conf /etc/nginx/nginx.conf

# Создание пользователя
RUN useradd -m -u 1000 siem && chown -R siem:siem /app

# Скрипт запуска
COPY webui/start.sh /start.sh
RUN chmod +x /start.sh

EXPOSE 80

CMD ["/start.sh"] 