#!/bin/bash

# Простой скрипт для отправки UDP данных на порт 51820

HOST=${1:-"localhost"}
PORT=${2:-"51820"}
MESSAGE=${3:-"Hello P2P Mesh!"}

echo "📤 Отправка UDP данных на $HOST:$PORT"
echo "📝 Сообщение: $MESSAGE"
echo ""

# Способ 1: netcat
echo "🔧 Способ 1: netcat"
echo "$MESSAGE" | nc -u -w 3 $HOST $PORT
echo ""

# Способ 2: /dev/udp (bash)
echo "🔧 Способ 2: /dev/udp"
echo "$MESSAGE" > /dev/udp/$HOST/$PORT
echo ""

# Способ 3: socat (если установлен)
if command -v socat &> /dev/null; then
    echo "🔧 Способ 3: socat"
    echo "$MESSAGE" | socat - UDP:$HOST:$PORT
    echo ""
else
    echo "⚠️  socat не установлен"
fi

# Способ 4: Python (если доступен)
if command -v python3 &> /dev/null; then
    echo "🔧 Способ 4: Python"
    python3 -c "
import socket
import sys
sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.sendto('$MESSAGE'.encode(), ('$HOST', $PORT))
print('Отправлено через Python')
sock.close()
"
    echo ""
fi

echo "✅ Все способы отправки выполнены!"


