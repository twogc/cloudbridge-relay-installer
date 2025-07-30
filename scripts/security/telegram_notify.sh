#!/bin/bash
# Скрипт для отправки сообщений в Telegram
TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
CHAT_ID="-1002403555405"
MESSAGE="$1"
MODE="${2:-HTML}"

# Экранируем только амперсанд для HTML-форматирования
if [[ "$MODE" == "HTML" ]]; then
    MESSAGE=$(echo "$MESSAGE" | sed 's/&/\&amp;/g')
fi

# Используем printf для правильной обработки \n
FORMATTED_MESSAGE=$(printf '%b' "$MESSAGE")

RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot$TOKEN/sendMessage" \
    -d chat_id="$CHAT_ID" \
    -d text="$FORMATTED_MESSAGE" \
    -d parse_mode="$MODE")

# Логируем ответ Telegram API
if echo "$RESPONSE" | grep -q '"ok":true'; then
    echo "[telegram_notify.sh] Успешно: $RESPONSE" >&2
else
    echo "[telegram_notify.sh] Ошибка: $RESPONSE" >&2
fi 