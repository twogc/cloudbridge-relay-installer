#!/bin/bash
# Получение реальных метрик CloudBridge Relay

echo "📊 CloudBridge Relay - Анализ метрик"
echo "====================================="
echo "Дата: $(date)"
echo ""

# Функция для проверки доступности порта
check_port() {
    local host="$1"
    local port="$2"
    local service="$3"
    
    if timeout 5 bash -c "</dev/tcp/$host/$port" 2>/dev/null; then
        echo "✅ $service ($host:$port) - Доступен"
        return 0
    else
        echo "❌ $service ($host:$port) - Недоступен"
        return 1
    fi
}

# Функция для получения метрик
get_metrics() {
    local url="$1"
    local name="$2"
    
    echo "📈 Получение метрик $name..."
    echo "URL: $url"
    
    # Пробуем HTTP
    response=$(curl -s --connect-timeout 5 "$url" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "✅ HTTP метрики получены"
        echo "Размер ответа: $(echo "$response" | wc -c) байт"
        echo "Строк: $(echo "$response" | wc -l)"
        return 0
    fi
    
    # Пробуем HTTPS
    response=$(curl -k -s --connect-timeout 5 "$url" 2>/dev/null)
    if [ $? -eq 0 ] && [ -n "$response" ]; then
        echo "✅ HTTPS метрики получены"
        echo "Размер ответа: $(echo "$response" | wc -c) байт"
        echo "Строк: $(echo "$response" | wc -l)"
        return 0
    fi
    
    echo "❌ Не удалось получить метрики"
    return 1
}

echo "🔍 Проверка доступности сервисов..."
echo ""

# Проверяем основные порты CloudBridge Relay
check_port "localhost" "5550" "Main Server"
check_port "localhost" "5551" "Metrics"
check_port "localhost" "5552" "P2P API"
check_port "localhost" "8082" "External API"

echo ""

# Проверяем мониторинг порты
check_port "localhost" "9090" "Prometheus"
check_port "localhost" "18083" "AI Service"

echo ""

# Пробуем получить метрики
echo "📊 Попытка получения метрик..."
echo ""

# Основные endpoints для метрик
endpoints=(
    "http://localhost:5551/metrics:CloudBridge Relay Metrics"
    "http://localhost:5550/metrics:CloudBridge Relay Main"
    "http://localhost:5552/metrics:CloudBridge Relay P2P API"
    "http://localhost:8082/metrics:CloudBridge Relay External"
    "http://localhost:9090/metrics:Prometheus"
    "http://localhost:18083/metrics:AI Service"
)

for endpoint in "${endpoints[@]}"; do
    url=$(echo "$endpoint" | cut -d':' -f1-2)
    name=$(echo "$endpoint" | cut -d':' -f3)
    
    echo "🔗 Проверка $name..."
    get_metrics "$url" "$name"
    echo ""
done

# Проверяем процессы CloudBridge Relay
echo "🔍 Анализ процессов CloudBridge Relay..."
echo ""

relay_processes=$(ps aux | grep -E "(cloudbridge-relay|cloudbridge-client)" | grep -v grep | wc -l)
echo "📊 Активных процессов CloudBridge: $relay_processes"

if [ "$relay_processes" -gt 0 ]; then
    echo "✅ CloudBridge процессы активны"
    echo ""
    echo "Детали процессов:"
    ps aux | grep -E "(cloudbridge-relay|cloudbridge-client)" | grep -v grep | head -5
else
    echo "❌ CloudBridge процессы не найдены"
fi

echo ""

# Проверяем сетевые соединения
echo "🌐 Анализ сетевых соединений..."
echo ""

netstat_output=$(netstat -tlnp 2>/dev/null | grep -E "(5550|5551|5552|5553|5554|5555|8080|8081|8082|8083|9090)")
if [ -n "$netstat_output" ]; then
    echo "✅ Найдены активные сетевые соединения:"
    echo "$netstat_output"
else
    echo "❌ Активные сетевые соединения не найдены"
fi

echo ""

# Проверяем логи
echo "📝 Анализ логов..."
echo ""

if [ -f "server-a.log" ]; then
    echo "✅ Лог файл server-a.log найден"
    echo "Размер: $(wc -c < server-a.log) байт"
    echo "Строк: $(wc -l < server-a.log)"
    echo ""
    echo "Последние 5 строк лога:"
    tail -5 server-a.log
else
    echo "❌ Лог файл server-a.log не найден"
fi

echo ""

# Проверяем конфигурацию
echo "⚙️ Анализ конфигурации..."
echo ""

if [ -f ".env" ]; then
    echo "✅ Файл .env найден"
    echo "Sentry DSN: $(grep SENTRY_DSN .env | cut -d'=' -f2 | cut -c1-50)..."
    echo "Environment: $(grep SENTRY_ENVIRONMENT .env | cut -d'=' -f2)"
    echo "Release: $(grep SENTRY_RELEASE .env | cut -d'=' -f2)"
else
    echo "❌ Файл .env не найден"
fi

echo ""

# Итоговый отчет
echo "📋 Итоговый отчет:"
echo "=================="
echo ""

# Подсчитываем доступные сервисы
available_services=0
total_services=6

if check_port "localhost" "5550" "Main Server" >/dev/null 2>&1; then
    ((available_services++))
fi

if check_port "localhost" "5551" "Metrics" >/dev/null 2>&1; then
    ((available_services++))
fi

if check_port "localhost" "5552" "P2P API" >/dev/null 2>&1; then
    ((available_services++))
fi

if check_port "localhost" "8082" "External API" >/dev/null 2>&1; then
    ((available_services++))
fi

if check_port "localhost" "9090" "Prometheus" >/dev/null 2>&1; then
    ((available_services++))
fi

if check_port "localhost" "18083" "AI Service" >/dev/null 2>&1; then
    ((available_services++))
fi

echo "📊 Статистика:"
echo "   Доступных сервисов: $available_services/$total_services"
echo "   Активных процессов: $relay_processes"
echo "   Лог файлов: $([ -f "server-a.log" ] && echo "1" || echo "0")"
echo "   Конфигурация: $([ -f ".env" ] && echo "1" || echo "0")"

echo ""
echo "🎯 Рекомендации:"
echo "   1. Проверить доступность Prometheus метрик"
echo "   2. Настроить Grafana дашборды"
echo "   3. Добавить дополнительные алерты"
echo "   4. Оптимизировать производительность"
echo "   5. Настроить горизонтальное масштабирование"

echo ""
echo "✅ Анализ завершен!"
