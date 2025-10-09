#!/bin/bash
# Управление Sentry Dashboard используя данные из .env

source .env

# Извлекаем данные из DSN
ORG_SLUG=$(echo $SENTRY_DSN | cut -d'@' -f2 | cut -d'.' -f1)
PROJECT_ID=$(echo $SENTRY_DSN | cut -d'/' -f4)

echo "🔧 Sentry Dashboard Management"
echo "============================="
echo "Organization: $ORG_SLUG"
echo "Project ID: $PROJECT_ID"
echo "Environment: $SENTRY_ENVIRONMENT"
echo "Release: $SENTRY_RELEASE"
echo ""

# Показываем доступные команды
case "$1" in
    "info")
        echo "📊 Информация о проекте:"
        echo "   DSN: $SENTRY_DSN"
        echo "   Organization: $ORG_SLUG"
        echo "   Project ID: $PROJECT_ID"
        echo "   Environment: $SENTRY_ENVIRONMENT"
        echo "   Release: $SENTRY_RELEASE"
        echo "   Server Name: $SENTRY_SERVER_NAME"
        echo "   Debug: $SENTRY_DEBUG"
        echo ""
        echo "🔗 Ссылки:"
        echo "   Dashboard: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
        echo "   Issues: https://sentry.io/organizations/$ORG_SLUG/issues/"
        echo "   Performance: https://sentry.io/organizations/$ORG_SLUG/performance/"
        echo "   Releases: https://sentry.io/organizations/$ORG_SLUG/releases/"
        ;;
    "dashboard")
        echo "📊 Sentry Dashboard:"
        echo "   URL: https://sentry.io/organizations/$ORG_SLUG/dashboards/"
        echo ""
        echo "📋 Рекомендуемые виджеты:"
        echo "   1. Error Rate - график ошибок"
        echo "   2. Response Time - время ответа"
        echo "   3. Throughput - пропускная способность"
        echo "   4. Memory Usage - использование памяти"
        echo "   5. CPU Usage - использование CPU"
        ;;
    "alerts")
        echo "🚨 Настройка алертов:"
        echo "   1. High Error Rate - >10 ошибок в минуту"
        echo "   2. Slow Response Time - >5 секунд"
        echo "   3. Memory Usage - >80%"
        echo "   4. CPU Usage - >90%"
        echo ""
        echo "🔗 Настройка: https://sentry.io/organizations/$ORG_SLUG/alerts/"
        ;;
    "metrics")
        echo "📈 Метрики Sentry:"
        echo "   Sample Rate: $SENTRY_SAMPLE_RATE"
        echo "   Traces Sample Rate: $SENTRY_TRACES_SAMPLE_RATE"
        echo "   Max Breadcrumbs: $SENTRY_MAX_BREADCRUMBS"
        echo ""
        echo "📊 Доступные метрики:"
        echo "   - Error Rate"
        echo "   - Response Time"
        echo "   - Throughput"
        echo "   - Memory Usage"
        echo "   - CPU Usage"
        echo "   - Custom Metrics"
        ;;
    "help"|"")
        echo "🔧 Доступные команды:"
        echo "   info      - Информация о проекте"
        echo "   dashboard - Ссылки на дашборд"
        echo "   alerts    - Настройка алертов"
        echo "   metrics   - Информация о метриках"
        echo "   help      - Показать эту справку"
        echo ""
        echo "Примеры:"
        echo "   $0 info"
        echo "   $0 dashboard"
        echo "   $0 alerts"
        echo "   $0 metrics"
        ;;
    *)
        echo "❌ Неизвестная команда: $1"
        echo "Используйте: $0 help"
        ;;
esac
