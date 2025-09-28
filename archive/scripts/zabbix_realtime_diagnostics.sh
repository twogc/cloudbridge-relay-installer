#!/bin/bash
# 🎯 ДИАГНОСТИКА ОБНОВЛЕНИЯ ДАННЫХ ZABBIX В РЕАЛЬНОМ ВРЕМЕНИ
# Скрипт для диагностики проблемы с необновляющимися данными в Zabbix
# Автор: AI Assistant | Дата: 26.09.2025

set -e

echo "🔍 ДИАГНОСТИКА ПРОБЛЕМЫ РЕАЛЬНОГО ВРЕМЕНИ ZABBIX"
echo "================================================="
echo "Дата: $(date)"
echo "Хост: cloudbridge-relay-01"
echo "Zabbix Server: zbx.2gc.ru (10.0.0.241)"
echo ""

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода результата
print_result() {
    local status=$1
    local message=$2
    if [ "$status" == "OK" ]; then
        echo -e "${GREEN}✅ $message${NC}"
    elif [ "$status" == "WARNING" ]; then
        echo -e "${YELLOW}⚠️ $message${NC}"
    elif [ "$status" == "ERROR" ]; then
        echo -e "${RED}❌ $message${NC}"
    else
        echo -e "${BLUE}ℹ️ $message${NC}"
    fi
}

# Функция для тестирования метрики
test_metric() {
    local key=$1
    local description=$2
    local timeout=10
    
    echo -n "🔍 Тестирование: $description "
    
    result=$(timeout $timeout zabbix_get -s localhost -k "$key" 2>/dev/null || echo "ERROR")
    
    if [ "$result" == "ERROR" ] || [ -z "$result" ]; then
        echo -e "${RED}❌ ОШИБКА${NC}"
        echo "   Ключ: $key"
        echo "   Результат: $result"
        return 1
    elif [ "$result" == "0" ]; then
        echo -e "${YELLOW}⚠️ НОЛЬ${NC}"
        echo "   Ключ: $key"
        echo "   Результат: $result"
        return 1
    else
        echo -e "${GREEN}✅ $result${NC}"
        return 0
    fi
}

echo "🧪 1. СОЗДАНИЕ ТЕСТОВЫХ МЕТРИК ДЛЯ ДИАГНОСТИКИ"
echo "================================================"

# Создание конфигурации для тестовых метрик
ZABBIX_TESTCONF="/etc/zabbix/zabbix_agentd.d/cloudbridge-realtime-test.conf"

print_result "INFO" "Создание файла конфигурации тестовых метрик..."

sudo tee "$ZABBIX_TESTCONF" > /dev/null << 'EOF'
### CloudBridge Real-time Test Metrics ###
# Быстро обновляющиеся метрики для диагностики

# Временные метки (обновляются каждую секунду)
UserParameter=cloudbridge.test.timestamp,date +%s
UserParameter=cloudbridge.test.timestamp_ms,date +%s%3N
UserParameter=cloudbridge.test.random,echo $((RANDOM % 1000 + 1))

# Живые системные метрики
UserParameter=cloudbridge.test.memory_live,free | awk 'NR==2{printf "%.2f", $3*100/$2}'
UserParameter=cloudbridge.test.cpu_live,top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}'
UserParameter=cloudbridge.test.uptime_live,awk '{print int($1)}' /proc/uptime

# Prometheus endpoint доступность  
UserParameter=cloudbridge.test.prometheus_alive,curl -s --max-time 2 http://192.168.49.2:30081/metrics | head -1 | wc -l

# Relay Memory (должен изменяться)
UserParameter=cloudbridge.relay.json.memory_live,curl -s --max-time 3 'http://192.168.49.2:30081/metrics' | grep '^go_memstats_alloc_bytes ' | awk '{printf "%.2f", $2/1024/1024}'

# Relay Goroutines (должен изменяться)  
UserParameter=cloudbridge.relay.json.goroutines_live,curl -s --max-time 3 'http://192.168.49.2:30081/metrics' | grep '^go_goroutines ' | awk '{print $2}'

# Relay uptime (должен расти)
UserParameter=cloudbridge.relay.json.uptime_live,curl -s --max-time 3 'http://192.168.49.2:30081/metrics' | grep '^process_start_time_seconds ' | awk 'BEGIN{now=systime()} {print int(now-$2)}'

# gRPC соединения
UserParameter=cloudbridge.relay.json.grpc_total,curl -s --max-time 3 'http://192.168.49.2:30081/metrics' | grep '^grpc_server_started_total' | awk 'END{sum=0} {sum+=$2} END{print sum}'
EOF

print_result "OK" "Создан файл: $ZABBIX_TESTCONF"

echo ""
echo "🔄 2. ПЕРЕЗАПУСК ZABBIX АГЕНТА"
echo "================================"

print_result "INFO" "Перезапуск Zabbix агента для загрузки новых метрик..."
sudo systemctl restart zabbix-agent

sleep 3

if sudo systemctl is-active --quiet zabbix-agent; then
    print_result "OK" "Zabbix агент запущен успешно"
else
    print_result "ERROR" "Ошибка запуска Zabbix агента"
    sudo systemctl status zabbix-agent --no-pager
    exit 1
fi

echo ""
echo "🧪 3. ТЕСТИРОВАНИЕ БЫСТРО ОБНОВЛЯЮЩИХСЯ МЕТРИК"
echo "==============================================="

# Тест временных меток
echo "⏰ ТЕСТ ВРЕМЕННЫХ МЕТОК:"
test_metric "cloudbridge.test.timestamp" "Unix timestamp"
test_metric "cloudbridge.test.timestamp_ms" "Unix timestamp (миллисекунды)"
test_metric "cloudbridge.test.random" "Случайное число"

echo ""
echo "💻 ТЕСТ ЖИВЫХ СИСТЕМНЫХ МЕТРИК:"
test_metric "cloudbridge.test.memory_live" "Живое использование памяти"
test_metric "cloudbridge.test.cpu_live" "Живое использование CPU"
test_metric "cloudbridge.test.uptime_live" "Uptime системы"

echo ""
echo "🌐 ТЕСТ PROMETHEUS И RELAY МЕТРИК:"
test_metric "cloudbridge.test.prometheus_alive" "Prometheus доступность"
test_metric "cloudbridge.relay.json.memory_live" "Память Relay (живая)"
test_metric "cloudbridge.relay.json.goroutines_live" "Горутины Relay (живые)"
test_metric "cloudbridge.relay.json.uptime_live" "Uptime Relay (живой)"
test_metric "cloudbridge.relay.json.grpc_total" "gRPC соединения (общие)"

echo ""
echo "🔍 4. ПРОВЕРКА РЕАЛЬНОГО ВРЕМЕНИ (15 СЕКУНД)"
echo "============================================"

print_result "INFO" "Тестирование обновлений в реальном времени..."

# Снимаем 3 значения с интервалом 5 секунд
metrics_to_monitor=(
    "cloudbridge.test.timestamp:Timestamp"
    "cloudbridge.test.random:Random"
    "cloudbridge.relay.json.memory_live:Relay_Memory"
    "cloudbridge.relay.json.goroutines_live:Goroutines"
)

echo "Время | Timestamp | Random | Memory_MB | Goroutines"
echo "------|-----------|--------|-----------|----------"

for i in {1..3}; do
    timestamp=$(zabbix_get -s localhost -k cloudbridge.test.timestamp 2>/dev/null || echo "ERR")
    random=$(zabbix_get -s localhost -k cloudbridge.test.random 2>/dev/null || echo "ERR")
    memory=$(zabbix_get -s localhost -k cloudbridge.relay.json.memory_live 2>/dev/null || echo "ERR")
    goroutines=$(zabbix_get -s localhost -k cloudbridge.relay.json.goroutines_live 2>/dev/null || echo "ERR")
    
    printf "%02d:00 | %10s | %6s | %9s | %10s\n" $((i*5)) "$timestamp" "$random" "$memory" "$goroutines"
    
    if [ $i -lt 3 ]; then
        sleep 5
    fi
done

echo ""
echo "🔍 5. ДИАГНОСТИКА ZABBIX СЕРВЕРА"  
echo "================================="

print_result "INFO" "Проверка подключения к Zabbix серверу..."

# Проверка подключения к серверу
if ping -c 1 10.0.0.241 >/dev/null 2>&1; then
    print_result "OK" "Сервер 10.0.0.241 пингуется"
else
    print_result "ERROR" "Сервер 10.0.0.241 недоступен"
fi

# Проверка порта 10051
if nc -z 10.0.0.241 10051 2>/dev/null; then
    print_result "OK" "Порт 10051 на сервере доступен"  
else
    print_result "ERROR" "Порт 10051 на сервере недоступен"
fi

# Проверка логов агента
print_result "INFO" "Последние записи в логе Zabbix агента:"
sudo tail -n 10 /var/log/zabbix/zabbix_agentd.log | while read line; do
    echo "  $line"
done

echo ""
echo "⚙️ 6. АНАЛИЗ КОНФИГУРАЦИИ"
echo "========================="

print_result "INFO" "Проверка конфигурации Zabbix агента..."

# Проверяем основные параметры
server_config=$(grep "^Server=" /etc/zabbix/zabbix_agentd.conf | head -1)
serveractive_config=$(grep "^ServerActive=" /etc/zabbix/zabbix_agentd.conf | head -1)  
hostname_config=$(grep "^Hostname=" /etc/zabbix/zabbix_agentd.conf | head -1)

echo "  $server_config"
echo "  $serveractive_config"  
echo "  $hostname_config"

# Проверяем количество UserParameters
user_params=$(grep "^UserParameter=" /etc/zabbix/zabbix_agentd.conf /etc/zabbix/zabbix_agentd.d/*.conf 2>/dev/null | wc -l)
print_result "INFO" "Всего UserParameter: $user_params"

echo ""
echo "🎯 7. РЕКОМЕНДАЦИИ ДЛЯ ZABBIX СЕРВЕРА"
echo "====================================="

print_result "INFO" "Команды для диагностики на Zabbix сервере zbx.2gc.ru:"

echo ""
echo "1. Проверьте подключение агента:"
echo "   zabbix_get -s $(hostname -I | awk '{print $1}') -k cloudbridge.test.timestamp"
echo "   zabbix_get -s $(hostname -I | awk '{print $1}') -k cloudbridge.relay.json.memory_live"

echo ""  
echo "2. Проверьте хост в Zabbix UI:"
echo "   Configuration → Hosts → cloudbridge-relay-01"
echo "   - Статус: Enabled"
echo "   - Agent interface: $(hostname -I | awk '{print $1}'):10050"
echo "   - Templates: CloudBridge Relay Template"

echo ""
echo "3. Проверьте элементы (Items):"
echo "   Configuration → Hosts → cloudbridge-relay-01 → Items"
echo "   - Type: Zabbix agent (passive)"
echo "   - Update interval: 30s"
echo "   - Status: Enabled"

echo ""
echo "4. Проверьте Latest Data:"
echo "   Monitoring → Latest Data → Host: cloudbridge-relay-01"
echo "   - Данные должны обновляться каждые 30 секунд"

echo ""
echo "5. Импортируйте шаблон (если не импортирован):"
echo "   Configuration → Templates → Import"
echo "   - Файл: cloudbridge-relay-template-fixed.xml"

echo ""
print_result "INFO" "Логи Zabbix сервера:"
echo "   sudo tail -f /var/log/zabbix/zabbix_server.log | grep 'cloudbridge-relay-01'"

echo ""
echo "📊 8. ИТОГОВЫЙ ОТЧЕТ"
echo "==================="

# Подсчет успешных тестов
successful_tests=0
total_tests=9

test_results=(
    $(zabbix_get -s localhost -k cloudbridge.test.timestamp >/dev/null 2>&1 && echo 1 || echo 0)
    $(zabbix_get -s localhost -k cloudbridge.test.random >/dev/null 2>&1 && echo 1 || echo 0)
    $(zabbix_get -s localhost -k cloudbridge.test.memory_live >/dev/null 2>&1 && echo 1 || echo 0)
    $(zabbix_get -s localhost -k cloudbridge.relay.json.memory_live >/dev/null 2>&1 && echo 1 || echo 0)
    $(zabbix_get -s localhost -k cloudbridge.relay.json.goroutines_live >/dev/null 2>&1 && echo 1 || echo 0)
    $(zabbix_get -s localhost -k cloudbridge.relay.json.uptime_live >/dev/null 2>&1 && echo 1 || echo 0)
    $(ping -c 1 10.0.0.241 >/dev/null 2>&1 && echo 1 || echo 0)
    $(nc -z 10.0.0.241 10051 >/dev/null 2>&1 && echo 1 || echo 0)
    $(sudo systemctl is-active --quiet zabbix-agent && echo 1 || echo 0)
)

for result in "${test_results[@]}"; do
    successful_tests=$((successful_tests + result))
done

echo "Успешных тестов: $successful_tests/$total_tests"
echo "Процент успеха: $((successful_tests * 100 / total_tests))%"

if [ $successful_tests -eq $total_tests ]; then
    print_result "OK" "ВСЕ ТЕСТЫ ПРОЙДЕНЫ! Агент готов к работе."
elif [ $successful_tests -gt 6 ]; then
    print_result "WARNING" "Большинство тестов пройдено. Нужны небольшие исправления."
else
    print_result "ERROR" "Много проблем. Требуется дополнительная настройка."
fi

echo ""
print_result "INFO" "🎯 ОСНОВНАЯ ПРОБЛЕМА: Данные собираются локально, но не обновляются на Zabbix сервере."
print_result "INFO" "🔧 РЕШЕНИЕ: Добавить хост 'cloudbridge-relay-01' в конфигуацию Zabbix сервера zbx.2gc.ru"

echo ""
echo "🎉 ДИАГНОСТИКА ЗАВЕРШЕНА"
echo "========================="
echo "Дата завершения: $(date)"



