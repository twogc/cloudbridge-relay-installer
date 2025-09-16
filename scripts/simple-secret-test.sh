#!/bin/bash

# 🔐 Простое тестирование fallback секрета
# Без зависимостей от внешних инструментов

set -e

# Цвета
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔐 ПРОСТОЕ ТЕСТИРОВАНИЕ FALLBACK СЕКРЕТА${NC}"
echo "=========================================="

# Реальный токен
REAL_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsIm9yZ19pZCI6ImIyZGJiMjkyLWQ4MzMtNGRlMi05MjFhLTJlMTJhZWFjZjg4MCIsInRlbmFudF9pZCI6ImIyZGJiMjkyLWQ4MzMtNGRlMi05MjFhLTJlMTJhZWFjZjg4MCIsInNlcnZlcl9pZCI6InNlcnZlci0xNzU3NzY5MjYwODI0IiwibWF4X3BlZXJzIjoiMTAiLCJwZXJtaXNzaW9ucyI6WyJtZXNoX2pvaW4iLCJtZXNoX21hbmFnZSJdLCJpYXQiOjE3NTc4NjQ3NTgsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci1jbGllbnQtc2VydmVyLTE3NTc3NjkyNjA4MjQiLCJqdGkiOiJqd3RfMTc1Nzg2NDc1ODQ1Ml9mM2FhZTRsOWwifQ.LMk61cPqnqx2-BubRhTwhfbV0JT5hPdMURiZi8yiGcc"

echo -e "${YELLOW}📋 Тестируемые секреты:${NC}"
echo ""

# Список секретов для тестирования
SECRETS=(
    "85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="
    "85Sk_NfLq3gqzCXzmBKbJCpL-f5BssXz3G8dVi3sPiE"
    "test-secret"
    "fallback-secret"
    "keycloak-fallback"
    "cloudbridge-secret"
    "1Pc6bvrTUDtXMKm6mMF6+7UJi/EYAdOcGPkXCRYV4xI="
    "test-secret-key-for-development"
    "your-secret-key"
)

# Функция для создания тестового токена с секретом
create_test_token() {
    local secret="$1"
    local header='{"alg":"HS256","typ":"JWT","kid":"fallback-key"}'
    local payload='{"protocol_type":"p2p-mesh","scope":"p2p-mesh-claims","tenant_id":"test","iat":1757864758,"iss":"https://auth.2gc.ru/realms/cloudbridge","aud":"account","sub":"test"}'
    
    # Кодируем header и payload
    local encoded_header=$(echo -n "$header" | base64 | tr -d '=')
    local encoded_payload=$(echo -n "$payload" | base64 | tr -d '=')
    
    # Создаем подпись (упрощенная версия)
    local message="$encoded_header.$encoded_payload"
    local signature=$(echo -n "$message" | openssl dgst -sha256 -hmac "$secret" -binary | base64 | tr -d '=')
    
    echo "$message.$signature"
}

# Тестируем каждый секрет
for secret in "${SECRETS[@]}"; do
    echo -e "${BLUE}🧪 Тестируем секрет: $secret${NC}"
    
    # Создаем тестовый токен
    test_token=$(create_test_token "$secret")
    
    # Проверяем, совпадает ли подпись с реальным токеном
    real_signature=$(echo "$REAL_TOKEN" | cut -d'.' -f3)
    test_signature=$(echo "$test_token" | cut -d'.' -f3)
    
    if [ "$real_signature" = "$test_signature" ]; then
        echo -e "${GREEN}✅ УСПЕХ: Секрет подходит!${NC}"
        echo -e "${GREEN}   Правильный секрет: $secret${NC}"
        exit 0
    else
        echo -e "${RED}❌ Не подходит${NC}"
    fi
    echo ""
done

echo -e "${YELLOW}⚠️  Ни один из тестируемых секретов не подошел${NC}"
echo -e "${YELLOW}   Требуется проверка конфигурации Keycloak${NC}"

echo ""
echo -e "${BLUE}📊 АНАЛИЗ ТОКЕНА${NC}"
echo "=================="

# Декодируем части токена
echo -e "${YELLOW}Header:${NC}"
echo "$REAL_TOKEN" | cut -d'.' -f1 | base64 -d 2>/dev/null || echo "Ошибка декодирования"

echo ""
echo -e "${YELLOW}Payload:${NC}"
echo "$REAL_TOKEN" | cut -d'.' -f2 | base64 -d 2>/dev/null || echo "Ошибка декодирования"

echo ""
echo -e "${YELLOW}Signature:${NC}"
echo "$REAL_TOKEN" | cut -d'.' -f3

echo ""
echo -e "${BLUE}🔧 РЕКОМЕНДАЦИИ${NC}"
echo "================"
echo "1. Проверить настройки Keycloak realm 'cloudbridge'"
echo "2. Найти правильный fallback секрет в конфигурации"
echo "3. Проверить логи Keycloak"
echo "4. Убедиться, что токен создан с правильным секретом"


