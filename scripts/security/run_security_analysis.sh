#!/bin/bash

# Запуск OSSEC, Snort, Suricata
sudo systemctl start ossec || true
sudo systemctl start snort || true
sudo systemctl start suricata || true

# Проверка статуса
for svc in ossec snort suricata; do
  if ! systemctl is-active --quiet $svc; then
    $(dirname "$0")/telegram_notify.sh "$svc не запущен!"
  fi
done

# Проверка новых алертов
$(dirname "$0")/snort_suricata_alert.sh

# Можно добавить другие проверки и анализы 