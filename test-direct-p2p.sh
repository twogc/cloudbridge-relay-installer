#!/bin/bash

# Скрипт для тестирования прямого P2P соединения
# Без использования CloudBridge Relay

set -e

echo "=== ПРЯМОЕ P2P СОЕДИНЕНИЕ БЕЗ RELAY ==="
echo "Тестирование прямого P2P соединения между серверами А и Б"
echo ""

# Проверка конфигураций
echo "=== Проверка конфигураций ==="
if [ ! -f "test/relay-client/config-direct-p2p-a.yaml" ]; then
    echo "[ERROR] Конфигурация сервера А не найдена"
    exit 1
fi

if [ ! -f "test/relay-client/config-direct-p2p-b.yaml" ]; then
    echo "[ERROR] Конфигурация сервера Б не найдена"
    exit 1
fi

echo "[INFO] Конфигурации найдены"

# Проверка CloudBridge клиента
if [ ! -f "test/relay-client/cloudbridge-client" ]; then
    echo "[ERROR] CloudBridge клиент не найден"
    exit 1
fi

echo "[INFO] CloudBridge клиент готов"

# Очистка старых логов
echo "=== Очистка старых логов ==="
rm -f direct-server-a.log direct-server-b.log

echo "=== Запуск прямого P2P тестирования ==="
echo "Запускаем два сервера в фоновом режиме..."
echo "Для остановки нажмите Ctrl+C"
echo ""

# Запуск сервера А
echo "[INFO] Запуск сервера А (прямое P2P)..."
timeout 30 ./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-direct-p2p-a.yaml --log-level debug --verbose > direct-server-a.log 2>&1 &
SERVER_A_PID=$!

# Небольшая задержка перед запуском сервера Б
sleep 2

# Запуск сервера Б
echo "[INFO] Запуск сервера Б (прямое P2P)..."
timeout 30 ./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-direct-p2p-b.yaml --log-level debug --verbose > direct-server-b.log 2>&1 &
SERVER_B_PID=$!

echo "[INFO] Серверы запущены:"
echo "  Сервер А (PID: $SERVER_A_PID) - логи в direct-server-a.log"
echo "  Сервер Б (PID: $SERVER_B_PID) - логи в direct-server-b.log"
echo ""

echo "Для мониторинга логов используйте:"
echo "  tail -f direct-server-a.log"
echo "  tail -f direct-server-b.log"
echo ""

echo "Для остановки нажмите Ctrl+C"
echo ""

# Функция для остановки серверов
cleanup() {
    echo ""
    echo "=== Остановка серверов ==="
    kill $SERVER_A_PID 2>/dev/null || true
    kill $SERVER_B_PID 2>/dev/null || true
    echo "[INFO] Серверы остановлены"
    exit 0
}

# Обработка сигналов
trap cleanup SIGINT SIGTERM

# Ожидание завершения
wait
