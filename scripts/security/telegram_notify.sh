#!/bin/bash
# Скрипт для отправки сообщений в Telegram
TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
CHAT_ID="-1002403555405"
MESSAGE="$1"

curl -s -X POST "https://api.telegram.org/bot$TOKEN/sendMessage" \
    -d chat_id="$CHAT_ID" \
    -d text="$MESSAGE" 