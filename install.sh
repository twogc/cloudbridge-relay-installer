#!/bin/bash

set -e

# Проверка root прав
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

# Создание пользователя и групп
if ! id "relay" &>/dev/null; then
    echo "Creating relay user..."
    useradd -r -s /bin/false relay
fi

# Создание директорий
echo "Creating directories..."
mkdir -p /opt/relay
mkdir -p /var/log/relay
chown -R relay:relay /opt/relay
chown -R relay:relay /var/log/relay

# Копирование файлов
echo "Installing files..."
cp opt/relay/relay /opt/relay/
cp etc/systemd/system/relay.service /etc/systemd/system/

# Установка прав
chmod 755 /opt/relay/relay
chmod 644 /etc/systemd/system/relay.service

# Перезагрузка systemd
echo "Reloading systemd..."
systemctl daemon-reload

# Запуск сервиса
echo "Starting relay service..."
systemctl enable relay
systemctl start relay

echo "Installation complete!"
echo "Check service status with: systemctl status relay"
echo "View logs with: journalctl -u relay"