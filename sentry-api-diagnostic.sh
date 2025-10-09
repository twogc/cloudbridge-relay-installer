#!/bin/bash
# Диагностика проблем с Sentry API для CloudBridge Relay

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔍 Sentry API Diagnostic - CloudBridge Relay"
echo "============================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Функция для выполнения API запросов с диагностикой
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local auth_type="$4"
    
    echo "🔗 Запрос: $method $endpoint"
    
    if [ "$auth_type" = "token" ]; then
        if [ -n "$data" ]; then
            curl -s -X "$method" \
                -H "Authorization: Bearer $SENTRY_API_TOKEN" \
                -H "Content-Type: application/json" \
                -d "$data" \
                -w "HTTP Status: %{http_code}\n" \
                "$endpoint"
        else
            curl -s -X "$method" \
                -H "Authorization: Bearer $SENTRY_API_TOKEN" \
                -w "HTTP Status: %{http_code}\n" \
                "$endpoint"
        fi
    elif [ "$auth_type" = "dsn" ]; then
        local dsn_auth=$(echo $SENTRY_DSN | cut -d'@' -f1)
        if [ -n "$data" ]; then
            curl -s -X "$method" \
                -u "$dsn_auth:" \
                -H "Content-Type: application/json" \
                -d "$data" \
                -w "HTTP Status: %{http_code}\n" \
                "$endpoint"
        else
            curl -s -X "$method" \
                -u "$dsn_auth:" \
                -w "HTTP Status: %{http_code}\n" \
                "$endpoint"
        fi
    fi
}

echo "1. 🔐 Проверка токена и прав доступа..."
echo "======================================"

# Проверяем базовую информацию о токене
echo "📋 Информация о токене:"
TOKEN_INFO=$(api_request "GET" "https://sentry.io/api/0/" "" "token")
echo "$TOKEN_INFO"
echo ""

# Проверяем доступ к организации
echo "🏢 Проверка доступа к организации:"
ORG_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/" "" "token")
echo "$ORG_RESPONSE"
echo ""

# Проверяем доступ к проекту
echo "📁 Проверка доступа к проекту:"
PROJECT_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/" "" "token")
echo "$PROJECT_RESPONSE"
echo ""

echo "2. 🔑 Проверка DSN аутентификации..."
echo "=================================="

# Проверяем DSN аутентификацию
echo "📋 DSN аутентификация:"
DSN_AUTH=$(echo $SENTRY_DSN | cut -d'@' -f1)
echo "DSN Auth: $DSN_AUTH"
echo ""

# Проверяем доступ к проекту через DSN
echo "📁 Проверка доступа к проекту через DSN:"
PROJECT_DSN_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/" "" "dsn")
echo "$PROJECT_DSN_RESPONSE"
echo ""

echo "3. 📊 Проверка доступных endpoints..."
echo "==================================="

# Проверяем различные endpoints
echo "🔗 Проверка доступных endpoints:"

# События
echo "📋 События:"
EVENTS_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/events/" "" "token")
echo "$EVENTS_RESPONSE"
echo ""

# Релизы
echo "🏷️ Релизы:"
RELEASES_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/projects/$ORG_SLUG/cloudbridge-relay/releases/" "" "token")
echo "$RELEASES_RESPONSE"
echo ""

# Дашборды
echo "📊 Дашборды:"
DASHBOARDS_RESPONSE=$(api_request "GET" "https://sentry.io/api/0/organizations/$ORG_SLUG/dashboards/" "" "token")
echo "$DASHBOARDS_RESPONSE"
echo ""

echo "4. 🔍 Анализ проблем..."
echo "======================"

# Анализируем проблемы
echo "❌ Обнаруженные проблемы:"
echo ""

if echo "$TOKEN_INFO" | grep -q "org:ci"; then
    echo "⚠️ Токен имеет ограниченные права: org:ci"
    echo "   Это означает, что токен предназначен только для CI/CD"
    echo "   Необходимы права: org:read, org:write, project:read, project:write"
fi

if echo "$ORG_RESPONSE" | grep -q "404"; then
    echo "⚠️ Организация не найдена или нет доступа"
    echo "   Проверьте правильность Organization ID: $ORG_SLUG"
fi

if echo "$PROJECT_RESPONSE" | grep -q "404"; then
    echo "⚠️ Проект не найден или нет доступа"
    echo "   Проверьте правильность Project ID: cloudbridge-relay"
fi

echo ""
echo "💡 Рекомендации по решению:"
echo "=========================="
echo ""
echo "1. 🔑 Получите новый API токен с полными правами:"
echo "   https://sentry.io/settings/account/api/auth-tokens/"
echo "   Права: org:read, org:write, project:read, project:write, event:read, event:write"
echo ""
echo "2. 🔍 Проверьте правильность Organization ID:"
echo "   Текущий: $ORG_SLUG"
echo "   Проверьте в Sentry Dashboard"
echo ""
echo "3. 📁 Проверьте правильность Project ID:"
echo "   Текущий: cloudbridge-relay"
echo "   Проверьте в Sentry Dashboard"
echo ""
echo "4. 🔧 Альтернативное решение:"
echo "   Используйте ручную настройку дашборда:"
echo "   ./setup-cloudbridge-sentry-dashboard.sh"
echo ""
echo "5. 📊 Проверьте доступ к Sentry Dashboard:"
echo "   https://sentry.io/organizations/$ORG_SLUG/dashboards/"
echo ""

echo "🎯 Следующие шаги:"
echo "================="
echo "1. Получите новый API токен с полными правами"
echo "2. Обновите SENTRY_API_TOKEN в .env файле"
echo "3. Запустите диагностику снова: ./sentry-api-diagnostic.sh"
echo "4. Если проблемы остаются, используйте ручную настройку"
