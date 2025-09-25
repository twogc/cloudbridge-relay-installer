#!/bin/bash

# Тестирование Telegram API напрямую
BOT_TOKEN="8188460753:AAF31IbAFk1Z6Sc9hfjxvI9ZweQdHXoW0ug"
CHAT_ID="-1002403555405"

echo "Тестирование Telegram API напрямую..."

# Тест 1: Проверка бота
echo "1. Проверка информации о боте..."
curl -s "https://api.telegram.org/bot${BOT_TOKEN}/getMe" | jq .

echo -e "\n2. Отправка тестового сообщения..."
curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
  -H "Content-Type: application/json" \
  -d "{
    \"chat_id\": \"${CHAT_ID}\",
    \"text\": \" Тестовое сообщение от CloudBridge AI Service\\n\\nВремя: $(date)\\nСтатус: Система работает нормально\\n\\nЭто автоматическое тестовое сообщение для проверки интеграции Telegram уведомлений.\",
    \"parse_mode\": \"HTML\"
  }" | jq .

echo -e "\n3. Отправка алерта о ресурсах..."
curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
  -H "Content-Type: application/json" \
  -d "{
    \"chat_id\": \"${CHAT_ID}\",
    \"text\": \"ALERT: High Resource Usage\\n\\nType: CPU Usage\\nUsage: 85%\\nThreshold: 80%\\nTime: $(date)\\n\\nRecommendation: Check system load\",
    \"parse_mode\": \"HTML\"
  }" | jq .

echo -e "\n4. Отправка отчета о производительности..."
curl -s -X POST "https://api.telegram.org/bot${BOT_TOKEN}/sendMessage" \
  -H "Content-Type: application/json" \
  -d "{
    \"chat_id\": \"${CHAT_ID}\",
    \"text\": \"CloudBridge Performance Report\\n\\nTime: $(date)\\n\\nMetrics:\\n• Active Connections: 45\\n• Throughput: 145.2 Mbps\\n• Latency: 25.5 ms\\n• Error Rate: 0.02%\\n\\nStatus: System operating stably\",
    \"parse_mode\": \"HTML\"
  }" | jq .

echo -e "\nТестирование завершено!"
