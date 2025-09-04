#!/bin/bash

# Скрипт для настройки тестового tenant'а relay-test-tenant
# CloudBridge Relay Server

set -e

echo "🔧 Настройка тестового tenant'а relay-test-tenant..."

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Проверка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker не установлен${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose не установлен${NC}"
    exit 1
fi

# Проверка статуса relay сервера
echo -e "${BLUE}📊 Проверка статуса relay сервера...${NC}"
if ! docker-compose ps cloudbridge-relay | grep -q "Up"; then
    echo -e "${RED}❌ Relay сервер не запущен${NC}"
    echo "Запустите: docker-compose up -d"
    exit 1
fi

echo -e "${GREEN}✅ Relay сервер запущен${NC}"

# Создание директории для логов
mkdir -p logs

# Проверка конфигурации
echo -e "${BLUE}🔍 Проверка конфигурации...${NC}"

if [ ! -f "configs/config.json" ]; then
    echo -e "${RED}❌ Файл configs/config.json не найден${NC}"
    exit 1
fi

if [ ! -f "configs/enhanced_quic.yaml" ]; then
    echo -e "${RED}❌ Файл configs/enhanced_quic.yaml не найден${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Конфигурационные файлы найдены${NC}"

# Проверка, что tenant уже настроен
echo -e "${BLUE}🔍 Проверка существующего tenant'а...${NC}"

if grep -q "relay-test-tenant" configs/config.json && grep -q "relay-test-tenant" configs/enhanced_quic.yaml; then
    echo -e "${GREEN}✅ Tenant relay-test-tenant уже настроен${NC}"
else
    echo -e "${YELLOW}⚠️  Tenant relay-test-tenant не найден в конфигурации${NC}"
    echo "Добавьте конфигурацию tenant'а в файлы конфигурации"
    exit 1
fi

# Перезапуск relay сервера для применения новой конфигурации
echo -e "${BLUE}🔄 Перезапуск relay сервера...${NC}"
docker-compose restart cloudbridge-relay

# Ожидание запуска сервера
echo -e "${BLUE}⏳ Ожидание запуска сервера...${NC}"
sleep 10

# Проверка статуса после перезапуска
if ! docker-compose ps cloudbridge-relay | grep -q "Up"; then
    echo -e "${RED}❌ Relay сервер не запустился после перезапуска${NC}"
    docker-compose logs cloudbridge-relay | tail -20
    exit 1
fi

echo -e "${GREEN}✅ Relay сервер перезапущен успешно${NC}"

# Проверка доступности API
echo -e "${BLUE}🌐 Проверка доступности API...${NC}"

# Проверка health endpoint
if curl -s http://localhost:9090/health | grep -q "healthy"; then
    echo -e "${GREEN}✅ Health endpoint доступен${NC}"
else
    echo -e "${RED}❌ Health endpoint недоступен${NC}"
    exit 1
fi

# Проверка metrics endpoint
if curl -s http://localhost:8084/metrics | head -1 | grep -q "#"; then
    echo -e "${GREEN}✅ Metrics endpoint доступен${NC}"
else
    echo -e "${RED}❌ Metrics endpoint недоступен${NC}"
    exit 1
fi

# Тестирование подключения с новым tenant'ом
echo -e "${BLUE}🧪 Тестирование подключения с tenant relay-test-tenant...${NC}"

# Создание тестового JWT токена (для демонстрации)
TEST_JWT="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJyZWxheS10ZXN0LXRlbmFudCIsInVzZXJfaWQiOiJkZXYtdXNlciIsInBlcm1pc3Npb25zIjpbInR1bm5lbCIsInJlYWQiLCJ3cml0ZSJdLCJpc3MiOiIyZ2MucnUiLCJhdWQiOiJyZWxheS10ZXN0LXRlbmFudCIsImV4cCI6MTczNTY4NjQwMCwiaWF0IjoxNzM1NjgyODAwfQ.test-signature"

echo "Тестовый JWT токен: $TEST_JWT"

# Тест handshake с новым tenant'ом
echo -e "${BLUE}📡 Тест handshake...${NC}"

# Отправка hello сообщения
HELLO_RESPONSE=$(echo -e '{"type":"hello","version":"1.0.0","tenant_id":"relay-test-tenant"}' | nc -w 5 localhost 9090)

if echo "$HELLO_RESPONSE" | grep -q "hello"; then
    echo -e "${GREEN}✅ Hello handshake успешен${NC}"
    echo "Ответ: $HELLO_RESPONSE"
else
    echo -e "${RED}❌ Hello handshake не удался${NC}"
    echo "Ответ: $HELLO_RESPONSE"
fi

# Тест auth с новым tenant'ом
echo -e "${BLUE}🔐 Тест аутентификации...${NC}"

AUTH_RESPONSE=$(echo -e "{\"type\":\"hello\",\"version\":\"1.0.0\"}\n{\"type\":\"auth\",\"token\":\"$TEST_JWT\",\"tenant_id\":\"relay-test-tenant\"}" | nc -w 10 localhost 9090)

if echo "$AUTH_RESPONSE" | grep -q "auth_response"; then
    echo -e "${GREEN}✅ Auth handshake успешен${NC}"
    echo "Ответ: $AUTH_RESPONSE"
else
    echo -e "${RED}❌ Auth handshake не удался${NC}"
    echo "Ответ: $AUTH_RESPONSE"
fi

# Проверка логов
echo -e "${BLUE}📋 Проверка логов...${NC}"
docker-compose logs cloudbridge-relay | grep -i "relay-test-tenant" | tail -5 || echo "Логи tenant'а не найдены"

# Итоговая информация
echo -e "${GREEN}🎉 Настройка тестового tenant'а завершена!${NC}"
echo ""
echo -e "${BLUE}📋 Информация о tenant'е:${NC}"
echo "  ID: relay-test-tenant"
echo "  Название: Relay Test Tenant"
echo "  Описание: Test tenant for DevOps development"
echo "  Лимиты:"
echo "    - Максимум туннелей: 10"
echo "    - Максимум соединений: 20"
echo "    - Пропускная способность: 100 MB"
echo "    - Rate limit: 500 запросов/сек"
echo "    - Burst: 100"
echo ""
echo -e "${BLUE}🔑 JWT токен для тестирования:${NC}"
echo "  Issuer: 2gc.ru"
echo "  Audience: relay-test-tenant"
echo "  Claims: tenant_id, user_id, permissions"
echo ""
echo -e "${BLUE}🌐 Endpoints для тестирования:${NC}"
echo "  - Main API: http://localhost:9090"
echo "  - Metrics: http://localhost:8084/metrics"
echo "  - Health: http://localhost:9090/health"
echo ""
echo -e "${YELLOW}⚠️  Важно:${NC}"
echo "  - JWT токен должен содержать поле 'tenant_id': 'relay-test-tenant'"
echo "  - Токен должен быть подписан с issuer '2gc.ru'"
echo "  - Токен должен иметь audience 'relay-test-tenant'"
echo ""
echo -e "${GREEN}✅ Tenant relay-test-tenant готов к использованию!${NC}"

