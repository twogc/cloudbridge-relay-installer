#!/bin/bash
# Полный набор команд для работы с Sentry Dashboard

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 Sentry Dashboard Commands"
echo "============================"
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo ""

# Показываем все доступные команды
echo "📋 Доступные команды:"
echo ""
echo "1. Информация о проекте:"
echo "   ./manage-sentry-dashboard-env.sh info"
echo ""
echo "2. Ссылки на дашборд:"
echo "   ./manage-sentry-dashboard-env.sh dashboard"
echo ""
echo "3. Настройка алертов:"
echo "   ./manage-sentry-dashboard-env.sh alerts"
echo ""
echo "4. Информация о метриках:"
echo "   ./manage-sentry-dashboard-env.sh metrics"
echo ""
echo "5. Создание дашборда (нужен API токен):"
echo "   ./setup-sentry-dashboard-full.sh"
echo ""
echo "🔗 Прямые ссылки:"
echo "   Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
echo "   Issues: https://sentry.io/organizations/$ORG_SLUG/issues/"
echo "   Performance: https://sentry.io/organizations/$ORG_SLUG/performance/"
echo "   Releases: https://sentry.io/organizations/$ORG_SLUG/releases/"
echo ""
echo "📊 Текущие настройки:"
echo "   Environment: $SENTRY_ENVIRONMENT"
echo "   Release: $SENTRY_RELEASE"
echo "   Debug: $SENTRY_DEBUG"
echo "   Sample Rate: $SENTRY_SAMPLE_RATE"
echo "   Traces Sample Rate: $SENTRY_TRACES_SAMPLE_RATE"
