#!/bin/bash

# Тестовый скрипт для проверки P2P маршрутизации CloudBridge Relay
# Этот скрипт тестирует базовую P2P функциональность

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация
RELAY_HOST="localhost"
HTTP_API_PORT="8080"
P2P_API_PORT="8082"
QUIC_PORT="9090"
STUN_PORT="19302"

# JWT токен для тестирования
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsIm9yZ19pZCI6InRlbmFudC0yMTY0MjAxNjUiLCJ0ZW5hbnRfaWQiOiJ0ZW5hbnQtMjE2NDIwMTY1Iiwic2VydmVyX2lkIjoic2VydmVyLTE3NTgyMTkxOTU3MDMiLCJjb25uZWN0aW9uX3R5cGUiOiJxdWljLWljZSIsIm1heF9wZWVycyI6IjEwIiwicGVybWlzc2lvbnMiOlsibWVzaF9qb2luIiwibWVzaF9tYW5hZ2UiXSwibmV0d29ya19jb25maWciOnsic3VibmV0IjoiMTAuMC4wLjAvMjQiLCJnYXRld2F5IjoiMTAuMC4wLjEiLCJkbnMiOlsiOC44LjguOCIsIjEuMS4xLjEiXSwibXR1IjoxNDIwLCJmaXJld2FsbF9ydWxlcyI6WyJhbGxvd19zc2giLCJhbGxvd19odHRwIl0sImVuYWJsZV9pcHY2IjpmYWxzZX0sInF1aWNfY29uZmlnIjp7ImhhbmRzaGFrZV90aW1lb3V0IjoiMTBzIiwiaWRsZV90aW1lb3V0IjoiMzBzIiwibWF4X3N0cmVhbXMiOjEwMCwiY29ubmVjdGlvbl9taWdyYXRpb24iOnRydWUsIm11bHRpcGxleGluZyI6dHJ1ZX0sImljZV9jb25maWciOnsic3R1bl9zZXJ2ZXJzIjpbImVkZ2UuMmdjLnJ1OjE5MzAyIl0sInR1cm5fc2VydmVycyI6W10sImljZV90aW1lb3V0IjoiMzBzIiwiY29ubmVjdGlvbl90aW1lb3V0IjoiMTBzIiwiZ2F0aGVyaW5nX3RpbWVvdXQiOiI1cyIsImNhbmRpZGF0ZV90eXBlcyI6WyJob3N0Iiwic3JmbHgiLCJyZWxheSJdfSwibWFzcXVlX2NvbmZpZyI6eyJlbmFibGVkIjp0cnVlLCJwcm94eV91cmwiOiJodHRwczovL2VkZ2UuMmdjLnJ1Ojg0NDMiLCJhbHBuX3Byb3RvY29scyI6WyJoMyIsImgzLTI5Il0sImZhbGxiYWNrX2VuYWJsZWQiOnRydWV9LCJtZXNoX2NvbmZpZyI6eyJuZXR3b3JrX2lkIjoibWVzaC1uZXR3b3JrLTAwMSIsInN1Ym5ldCI6IjEwLjAuMC4wLzE2IiwicmVnaXN0cnlfdXJsIjoiaHR0cHM6Ly9lZGdlLjJnYy5ydTo4MDgwIiwiaGVhcnRiZWF0X2ludGVydmFsIjoiMzBzIiwibWF4X3BlZXJzIjoxMCwicm91dGluZ19zdHJhdGVneSI6InBlcmZvcm1hbmNlX29wdGltYWwiLCJlbmFibGVfYXV0b19kaXNjb3ZlcnkiOnRydWUsInRydXN0X2xldmVsIjoiYmFzaWMifSwicGVlcl93aGl0ZWxpc3QiOlsicGVlci0wMDEiLCJwZWVyLTAwMiIsInBlZXItMDAzIl0sInJlbGF5X2VuZHBvaW50cyI6eyJodHRwX2FwaSI6ImVkZ2UuMmdjLnJ1OjgwODAiLCJwMnBfYXBpIjoiZWRnZS4yZ2MucnU6ODA4MiIsInF1aWNfdHJhbnNwb3J0IjoiZWRnZS4yZ2MucnU6OTA5MCIsInN0dW5fc2VydmVyIjoiZWRnZS4yZ2MucnU6MTkzMDIiLCJtYXNxdWVfcHJveHkiOiJlZGdlLjJnYy5ydTo4NDQzIiwiZW5oYW5jZWRfcXVpYyI6ImVkZ2UuMmdjLnJ1OjkwOTIifSwiaWF0IjoxNzU4MjE5MTk4LCJpc3MiOiJodHRwczovL2F1dGguMmdjLnJ1L3JlYWxtcy9jbG91ZGJyaWRnZSIsImF1ZCI6ImFjY291bnQiLCJzdWIiOiJzZXJ2ZXItY2xpZW50LXNlcnZlci0xNzU4MjE5MTk1NzAzIiwianRpIjoiand0XzE3NTgyMTkxOTgxMDlfbXoxdDhyc2RhIn0.d-o1yXtZOlnyv53Uhg5zEI3l3IRbQV27VdnEQIn5w_4"

# Функции для вывода
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Функция для проверки доступности сервиса
check_service() {
    local service_name=$1
    local url=$2
    local expected_status=${3:-200}
    
    print_info "Проверяем $service_name..."
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_status"; then
        print_success "$service_name доступен"
        return 0
    else
        print_error "$service_name недоступен"
        return 1
    fi
}

# Функция для тестирования API endpoint
test_api_endpoint() {
    local method=$1
    local url=$2
    local headers=$3
    local data=$4
    local expected_status=${5:-200}
    local description=$6
    
    print_info "Тестируем: $description"
    
    local response
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "$headers" -d "$data" "$url")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "$headers" "$url")
    fi
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_status" ]; then
        print_success "$description - HTTP $http_code"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 0
    else
        print_error "$description - HTTP $http_code (ожидался $expected_status)"
        echo "$body"
        return 1
    fi
}

# Основная функция тестирования
main() {
    print_header "CloudBridge Relay P2P Routing Test"
    
    # Проверяем доступность сервисов
    print_header "1. Проверка доступности сервисов"
    
    if ! check_service "HTTP API" "http://$RELAY_HOST:$HTTP_API_PORT/health"; then
        print_error "HTTP API недоступен. Убедитесь, что сервер запущен."
        exit 1
    fi
    
    if ! check_service "P2P API" "http://$RELAY_HOST:$P2P_API_PORT/health"; then
        print_error "P2P API недоступен. Убедитесь, что сервер запущен."
        exit 1
    fi
    
    # Тестируем базовые API endpoints
    print_header "2. Тестирование базовых API endpoints"
    
    # Health check
    test_api_endpoint "GET" "http://$RELAY_HOST:$HTTP_API_PORT/health" "" "" 200 "Health Check"
    test_api_endpoint "GET" "http://$RELAY_HOST:$P2P_API_PORT/health" "" "" 200 "P2P API Health Check"
    
    # Тестируем регистрацию пиров
    print_header "3. Тестирование регистрации пиров"
    
    local tenant_id="tenant-216420165"
    local peer1_id="peer-001"
    local peer2_id="peer-002"
    
    # Регистрация первого пира
    local peer1_data='{
        "public_key": "peer-001-public-key",
        "allowed_ips": ["10.100.77.10/32"],
        "peer_info": {
            "name": "Test Peer 1",
            "location": "Test Location 1"
        }
    }'
    
    test_api_endpoint "POST" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/register" \
        "Authorization: Bearer $JWT_TOKEN" \
        "$peer1_data" \
        200 \
        "Регистрация первого пира"
    
    # Регистрация второго пира
    local peer2_data='{
        "public_key": "peer-002-public-key",
        "allowed_ips": ["10.100.77.11/32"],
        "peer_info": {
            "name": "Test Peer 2",
            "location": "Test Location 2"
        }
    }'
    
    test_api_endpoint "POST" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/register" \
        "Authorization: Bearer $JWT_TOKEN" \
        "$peer2_data" \
        200 \
        "Регистрация второго пира"
    
    # Тестируем heartbeat систему
    print_header "4. Тестирование heartbeat системы"
    
    # Heartbeat для первого пира
    test_api_endpoint "POST" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/$peer1_id/heartbeat" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Heartbeat для первого пира"
    
    # Heartbeat для второго пира
    test_api_endpoint "POST" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/$peer2_id/heartbeat" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Heartbeat для второго пира"
    
    # Проверяем статус пиров
    test_api_endpoint "GET" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/$peer1_id/status" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Статус первого пира"
    
    test_api_endpoint "GET" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/$peer2_id/status" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Статус второго пира"
    
    # Тестируем статистику tenant
    print_header "5. Тестирование статистики tenant"
    
    test_api_endpoint "GET" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/stats" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Статистика tenant"
    
    test_api_endpoint "GET" "http://$RELAY_HOST:$P2P_API_PORT/api/v1/tenants/$tenant_id/peers/online" \
        "Authorization: Bearer $JWT_TOKEN" \
        "" \
        200 \
        "Список онлайн пиров"
    
    # Тестируем STUN сервер
    print_header "6. Тестирование STUN сервера"
    
    print_info "Тестируем STUN сервер на порту $STUN_PORT..."
    
    # Простой тест STUN сервера
    if timeout 5 bash -c "echo > /dev/udp/$RELAY_HOST/$STUN_PORT" 2>/dev/null; then
        print_success "STUN сервер доступен"
    else
        print_warning "STUN сервер может быть недоступен (это нормально для UDP)"
    fi
    
    # Тестируем QUIC порт
    print_header "7. Тестирование QUIC порта"
    
    print_info "Проверяем доступность QUIC порта $QUIC_PORT..."
    
    if nc -z -u "$RELAY_HOST" "$QUIC_PORT" 2>/dev/null; then
        print_success "QUIC порт доступен"
    else
        print_warning "QUIC порт может быть недоступен"
    fi
    
    # Итоговый отчет
    print_header "8. Итоговый отчет"
    
    print_success "Все основные тесты P2P маршрутизации выполнены!"
    print_info "Проверьте логи сервера для детальной информации о P2P соединениях"
    print_info "Для полного тестирования P2P маршрутизации запустите QUIC клиенты"
    
    print_header "Следующие шаги для полного тестирования:"
    echo "1. Запустите QUIC клиенты для тестирования P2P соединений"
    echo "2. Проверьте пересылку пакетов между пирами"
    echo "3. Протестируйте мультитунельность с разными tenant"
    echo "4. Проверьте ICE/STUN интеграцию"
}

# Запуск тестов
main "$@"
