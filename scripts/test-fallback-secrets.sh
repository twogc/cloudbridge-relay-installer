#!/bin/bash

# 🔐 Тестирование различных вариантов fallback секрета
# Для отладки проблем с JWT валидацией

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 ТЕСТИРОВАНИЕ FALLBACK СЕКРЕТОВ${NC}"
echo "=================================="

# Реальный токен от DevOps
REAL_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImZhbGxiYWNrLWtleSJ9.eyJwcm90b2NvbF90eXBlIjoicDJwLW1lc2giLCJzY29wZSI6InAycC1tZXNoLWNsYWltcyIsIm9yZ19pZCI6ImIyZGJiMjkyLWQ4MzMtNGRlMi05MjFhLTJlMTJhZWFjZjg4MCIsInRlbmFudF9pZCI6ImIyZGJiMjkyLWQ4MzMtNGRlMi05MjFhLTJlMTJhZWFjZjg4MCIsInNlcnZlcl9pZCI6InNlcnZlci0xNzU3NzY5MjYwODI0IiwibWF4X3BlZXJzIjoiMTAiLCJwZXJtaXNzaW9ucyI6WyJtZXNoX2pvaW4iLCJtZXNoX21hbmFnZSJdLCJpYXQiOjE3NTc4NjQ3NTgsImlzcyI6Imh0dHBzOi8vYXV0aC4yZ2MucnUvcmVhbG1zL2Nsb3VkYnJpZGdlIiwiYXVkIjoiYWNjb3VudCIsInN1YiI6InNlcnZlci1jbGllbnQtc2VydmVyLTE3NTc3NjkyNjA4MjQiLCJqdGkiOiJqd3RfMTc1Nzg2NDc1ODQ1Ml9mM2FhZTRsOWwifQ.LMk61cPqnqx2-BubRhTwhfbV0JT5hPdMURiZi8yiGcc"

# Наш предоставленный секрет
OUR_SECRET="85Sk/NfLq3gqzCXzmBKbJCpL+f5BssXz3G8dVi3sPiE="

echo -e "${YELLOW}📋 Тестируемые варианты секрета:${NC}"
echo ""

# Функция для тестирования секрета
test_secret() {
    local secret="$1"
    local description="$2"
    
    echo -e "${BLUE}🧪 Тестируем: $description${NC}"
    echo "Секрет: $secret"
    
    # Создаем временный файл с секретом
    echo "$secret" > /tmp/test_secret.txt
    
    # Тестируем с помощью jwt-cli (если установлен)
    if command -v jwt &> /dev/null; then
        if jwt verify "$REAL_TOKEN" --secret "$secret" &> /dev/null; then
            echo -e "${GREEN}✅ УСПЕХ: Токен валидируется с этим секретом!${NC}"
            return 0
        else
            echo -e "${RED}❌ ОШИБКА: Токен не валидируется${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  jwt-cli не установлен, пропускаем тест${NC}"
    fi
    
    echo ""
}

# Тестируем различные варианты секрета
echo -e "${YELLOW}1. Наш предоставленный секрет (Base64):${NC}"
test_secret "$OUR_SECRET" "Base64 секрет"

echo -e "${YELLOW}2. Base64 URL-safe версия:${NC}"
URL_SAFE_SECRET=$(echo "$OUR_SECRET" | tr '+/' '-_' | tr -d '=')
test_secret "$URL_SAFE_SECRET" "Base64 URL-safe секрет"

echo -e "${YELLOW}3. Hex версия:${NC}"
HEX_SECRET=$(echo "$OUR_SECRET" | base64 -d | xxd -p -c 256)
test_secret "$HEX_SECRET" "Hex секрет"

echo -e "${YELLOW}4. Raw bytes (base64 decoded):${NC}"
RAW_SECRET=$(echo "$OUR_SECRET" | base64 -d)
test_secret "$RAW_SECRET" "Raw bytes секрет"

echo -e "${YELLOW}5. Популярные тестовые секреты:${NC}"
test_secret "test-secret" "Тестовый секрет"
test_secret "fallback-secret" "Fallback секрет"
test_secret "keycloak-fallback" "Keycloak fallback"
test_secret "cloudbridge-secret" "CloudBridge секрет"

echo -e "${YELLOW}6. Секреты из конфигурации:${NC}"
test_secret "1Pc6bvrTUDtXMKm6mMF6+7UJi/EYAdOcGPkXCRYV4xI=" "Django JWT секрет"
test_secret "test-secret-key-for-development" "Docker Compose секрет"
test_secret "your-secret-key" "K8s ConfigMap секрет"

echo -e "${BLUE}📊 АНАЛИЗ ТОКЕНА${NC}"
echo "=================="

# Декодируем header и payload токена
echo -e "${YELLOW}Header токена:${NC}"
echo "$REAL_TOKEN" | cut -d'.' -f1 | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Ошибка декодирования header"

echo ""
echo -e "${YELLOW}Payload токена:${NC}"
echo "$REAL_TOKEN" | cut -d'.' -f2 | base64 -d 2>/dev/null | jq . 2>/dev/null || echo "Ошибка декодирования payload"

echo ""
echo -e "${BLUE}🔧 РЕКОМЕНДАЦИИ ДЛЯ DEVOPS${NC}"
echo "=============================="
echo "1. Проверить настройки Keycloak realm 'cloudbridge'"
echo "2. Найти правильный fallback секрет в конфигурации"
echo "3. Проверить логи Keycloak при создании токенов"
echo "4. Убедиться, что токен создан с правильным секретом"
echo ""
echo -e "${GREEN}✅ Скрипт завершен${NC}"


