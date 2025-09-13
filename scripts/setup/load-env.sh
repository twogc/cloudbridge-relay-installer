#!/bin/bash
# Загрузка переменных окружения из .env файла

if [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo "Переменные окружения загружены из .env файла"
else
    echo "Файл .env не найден"
    exit 1
fi
