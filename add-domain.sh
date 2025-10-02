#!/bin/bash

# Скрипт для добавления любого домена в hosts файл
# Использование: ./add-domain.sh example.com

if [ $# -eq 0 ]; then
    echo "Использование: $0 <domain>"
    echo "Пример: $0 example.com"
    exit 1
fi

DOMAIN=$1

echo "Добавляем домен: $DOMAIN"

# Получаем IP через TCP DNS
IP=$(dig +tcp @1.1.1.1 $DOMAIN +short | head -1)

if [ -n "$IP" ] && [[ $IP =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Найден IP для $DOMAIN: $IP"
    
    # Проверяем, есть ли уже запись
    if ! grep -q "$DOMAIN" /etc/hosts; then
        echo "$IP $DOMAIN" | sudo tee -a /etc/hosts
        echo "Добавлен $DOMAIN -> $IP"
    else
        echo "$DOMAIN уже есть в hosts файле"
    fi
else
    echo "Не удалось получить IP для $DOMAIN"
    exit 1
fi

echo "Готово! Проверьте: ping $DOMAIN"

