#!/bin/bash
# Ручная настройка Sentry Dashboard через веб-интерфейс

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔧 Ручная настройка Sentry Dashboard"
echo "===================================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

echo "📋 Инструкции по созданию дашборда:"
echo ""
echo "1. Откройте Sentry Dashboard:"
echo "   https://sentry.io/organizations/$ORG_SLUG/dashboards/"
echo ""
echo "2. Нажмите 'Create Dashboard'"
echo ""
echo "3. Настройте дашборд:"
echo "   - Название: 'CloudBridge Relay Dashboard'"
echo "   - Описание: 'Мониторинг CloudBridge Relay'"
echo ""
echo "4. Добавьте виджеты:"
echo "   - Error Rate (график ошибок)"
echo "   - Response Time (время ответа)"
echo "   - Throughput (пропускная способность)"
echo "   - Memory Usage (использование памяти)"
echo "   - CPU Usage (использование CPU)"
echo ""
echo "5. Настройте фильтры:"
echo "   - Environment: $SENTRY_ENVIRONMENT"
echo "   - Release: $SENTRY_RELEASE"
echo "   - Project: cloudbridge-relay"
echo ""
echo "6. Сохраните дашборд"
echo ""
echo "🔗 Полезные ссылки:"
echo "   Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
echo "   Issues: https://sentry.io/organizations/$ORG_SLUG/issues/"
echo "   Performance: https://sentry.io/organizations/$ORG_SLUG/performance/"
echo "   Releases: https://sentry.io/organizations/$ORG_SLUG/releases/"
echo ""
echo "📊 Рекомендуемые настройки:"
echo "   - Time Range: Last 24 hours"
echo "   - Refresh: 5 minutes"
echo "   - Alerts: Enable for critical metrics"
