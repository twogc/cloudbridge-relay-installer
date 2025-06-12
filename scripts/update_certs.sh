#!/bin/bash

# Проверяем наличие certbot
if ! command -v certbot &> /dev/null; then
    echo "Installing certbot..."
    sudo apt-get update
    sudo apt-get install -y certbot
fi

# Проверяем наличие переменной окружения
if [ -z "$RELAY_DOMAIN" ]; then
    echo "Error: RELAY_DOMAIN environment variable is not set"
    exit 1
fi

DOMAIN=$RELAY_DOMAIN

# Получаем сертификат
echo "Obtaining certificate for $DOMAIN..."
sudo certbot certonly --standalone -d $DOMAIN --agree-tos --email admin@$DOMAIN --non-interactive

# Копируем сертификаты в нужную директорию
echo "Copying certificates..."
sudo cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem /etc/relay/certs/
sudo cp /etc/letsencrypt/live/$DOMAIN/privkey.pem /etc/relay/certs/

# Устанавливаем правильные права
sudo chown -R ubuntu:ubuntu /etc/relay/certs
sudo chmod 600 /etc/relay/certs/*

# Перезапускаем сервис
echo "Restarting relay service..."
sudo systemctl restart relay

echo "Certificate update completed!" 