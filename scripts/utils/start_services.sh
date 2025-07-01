#!/bin/bash
set -e

echo -e "\033[1;34m[INFO]\033[0m Запуск сервисов..."

# Создаем необходимые директории
mkdir -p keycloak/data keycloak/imports logs certs backups
chmod 755 keycloak/data keycloak/imports logs
chmod 700 certs
chmod 755 backups

# Останавливаем существующие контейнеры
if docker-compose ps | grep -q "Up"; then
    echo -e "\033[1;34m[INFO]\033[0m Остановка существующих контейнеров..."
    docker-compose down
fi

# Запускаем сервисы
echo -e "\033[1;34m[INFO]\033[0m Запуск docker-compose..."
docker-compose --env-file .env up -d

echo -e "\033[0;32m[SUCCESS]\033[0m Сервисы запущены"

# Ждем готовности сервисов
echo -e "\033[1;34m[INFO]\033[0m Ожидание готовности сервисов..."
sleep 10

# Проверяем статус
if docker-compose ps | grep -q "Up"; then
    echo -e "\033[0;32m[SUCCESS]\033[0m Все сервисы запущены успешно"
else
    echo -e "\033[0;31m[ERROR]\033[0m Не все сервисы запустились"
    docker-compose ps
    exit 1
fi 