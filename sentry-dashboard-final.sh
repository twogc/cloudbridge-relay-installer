#!/bin/bash
# Итоговый скрипт для работы с Sentry Dashboard

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🚀 Sentry Dashboard - Итоговое решение"
echo "====================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

echo "📋 Доступные варианты настройки:"
echo ""
echo "1. 🔧 Ручная настройка (рекомендуется):"
echo "   ./setup-sentry-dashboard-manual.sh"
echo ""
echo "2. 🤖 Автоматическая настройка через API:"
echo "   ./create-sentry-dashboard-api.sh"
echo ""
echo "3. 📊 Управление дашбордом:"
echo "   ./manage-sentry-dashboard-env.sh info"
echo "   ./manage-sentry-dashboard-env.sh dashboard"
echo "   ./manage-sentry-dashboard-env.sh metrics"
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
echo ""
echo "🎯 Рекомендация:"
echo "   Используйте ручную настройку через веб-интерфейс"
echo "   для максимальной гибкости и контроля"
