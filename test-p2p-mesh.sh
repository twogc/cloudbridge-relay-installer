#!/bin/bash

# CloudBridge Relay P2P Mesh Test Script
# Запуск двух серверов для тестирования P2P mesh сети

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

# Проверяем, что CloudBridge Relay работает
print_header "Проверка CloudBridge Relay"
if ! curl -s https://edge.2gc.ru:5552/health > /dev/null; then
    print_error "CloudBridge Relay не доступен на https://edge.2gc.ru:5552"
fi
print_status "CloudBridge Relay работает"

# Проверяем конфигурационные файлы
print_header "Проверка конфигураций"
if [ ! -f "test/relay-client/config-server-a.yaml" ]; then
    print_error "Конфигурация сервера А не найдена"
fi
if [ ! -f "test/relay-client/config-server-b.yaml" ]; then
    print_error "Конфигурация сервера Б не найдена"
fi
print_status "Конфигурации найдены"

# Проверяем клиент
if [ ! -f "test/relay-client/cloudbridge-client" ]; then
    print_error "CloudBridge клиент не найден"
fi
print_status "CloudBridge клиент готов"

print_header "Запуск P2P Mesh тестирования"
echo "Запускаем два сервера в фоновом режиме..."
echo "Для остановки нажмите Ctrl+C"
echo

# Функция для очистки при выходе
cleanup() {
    print_warning "Останавливаем серверы..."
    kill $SERVER_A_PID 2>/dev/null || true
    kill $SERVER_B_PID 2>/dev/null || true
    exit 0
}

# Устанавливаем обработчик сигналов
trap cleanup SIGINT SIGTERM

# Запускаем сервер А в фоне
print_status "Запуск сервера А..."
./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-server-a.yaml --log-level debug > server-a.log 2>&1 &
SERVER_A_PID=$!

# Ждем немного
sleep 2

# Запускаем сервер Б в фоне
print_status "Запуск сервера Б..."
./test/relay-client/cloudbridge-client p2p --config test/relay-client/config-server-b.yaml --log-level debug > server-b.log 2>&1 &
SERVER_B_PID=$!

print_status "Серверы запущены:"
echo "  Сервер А (PID: $SERVER_A_PID) - логи в server-a.log"
echo "  Сервер Б (PID: $SERVER_B_PID) - логи в server-b.log"
echo
echo "Для мониторинга логов используйте:"
echo "  tail -f server-a.log"
echo "  tail -f server-b.log"
echo
echo "Для остановки нажмите Ctrl+C"
echo

# Ждем завершения
wait
