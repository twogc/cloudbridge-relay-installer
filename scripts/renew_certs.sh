#!/bin/bash

# Устанавливаем переменную окружения
export RELAY_DOMAIN="your-domain.com"

# Запускаем обновление сертификатов
/home/ubuntu/cloudbridge-relay-installer/scripts/update_certs.sh

# Перезапускаем сервис
sudo systemctl restart relay 