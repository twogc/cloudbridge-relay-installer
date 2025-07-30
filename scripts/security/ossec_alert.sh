#!/bin/bash
# Скрипт вызывается OSSEC при алерте
ALERT="$1"
$(dirname "$0")/telegram_notify.sh "$ALERT" 