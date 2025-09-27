#!/bin/bash
# 🎯 КОМАНДЫ ДЛЯ ДИАГНОСТИКИ НА ZABBIX СЕРВЕРЕ zbx.2gc.ru
# Скрипт с командами для выполнения на Zabbix сервере 10.0.0.241
# Автор: AI Assistant | Дата: 26.09.2025

echo "🎯 КОМАНДЫ ДЛЯ ВЫПОЛНЕНИЯ НА ZABBIX СЕРВЕРЕ zbx.2gc.ru"
echo "======================================================"
echo ""

# Получаем IP адрес хоста агента
AGENT_IP=$(hostname -I | awk '{print $1}')
echo "🔍 IP адрес агента: $AGENT_IP"
echo "📱 Имя хоста: cloudbridge-relay-01"
echo ""

echo "⚠️ ВНИМАНИЕ: Все команды ниже нужно выполнять НА ZABBIX СЕРВЕРЕ zbx.2gc.ru (10.0.0.241)"
echo "=============================================================================="
echo ""

echo "1️⃣ ТЕСТ ПОДКЛЮЧЕНИЯ К АГЕНТУ"
echo "=============================="
echo ""
cat << EOF
# Тест подключения с Zabbix сервера к агенту:
zabbix_get -s $AGENT_IP -k agent.ping

# Проверка быстро обновляющихся метрик:
zabbix_get -s $AGENT_IP -k cloudbridge.test.timestamp
zabbix_get -s $AGENT_IP -k cloudbridge.test.random  
zabbix_get -s $AGENT_IP -k cloudbridge.relay.json.memory_live

# Проверка основных CloudBridge метрик:
zabbix_get -s $AGENT_IP -k cloudbridge.relay.json.goroutines_live
zabbix_get -s $AGENT_IP -k cloudbridge.relay.json.uptime_live
zabbix_get -s $AGENT_IP -k cloudbridge.system.cpu.usage
zabbix_get -s $AGENT_IP -k cloudbridge.system.memory.usage

# Kubernetes метрики:  
zabbix_get -s $AGENT_IP -k cloudbridge.k8s.pods.running
zabbix_get -s $AGENT_IP -k cloudbridge.k8s.nodes.ready

# Docker метрики:
zabbix_get -s $AGENT_IP -k cloudbridge.docker.containers.running
zabbix_get -s $AGENT_IP -k cloudbridge.docker.containers.total
EOF

echo ""
echo "2️⃣ ПРОВЕРКА ЛОГОВ ZABBIX СЕРВЕРА"
echo "================================="
echo ""
cat << EOF
# Проверка логов подключения к агенту:
sudo tail -f /var/log/zabbix/zabbix_server.log | grep "cloudbridge-relay-01"

# Поиск ошибок с агентом:
sudo grep "cloudbridge-relay-01" /var/log/zabbix/zabbix_server.log | tail -20

# Поиск проблем с подключением:
sudo grep "$AGENT_IP" /var/log/zabbix/zabbix_server.log | tail -10
EOF

echo ""
echo "3️⃣ НАСТРОЙКА ХОСТА В ZABBIX WEB UI"
echo "==================================="
echo ""
cat << EOF
В Zabbix Web интерфейсе (http://zbx.2gc.ru):

Configuration → Hosts → Create host

Основные настройки:
- Host name: cloudbridge-relay-01
- Groups: CloudBridge (или создать новую группу)
- Agent interfaces: 
  * IP address: $AGENT_IP
  * Port: 10050
  * Статус: Enabled

Templates (Шаблоны):
- Импортировать: cloudbridge-relay-template-fixed.xml
- Применить шаблон: CloudBridge Relay Template
EOF

echo ""
echo "4️⃣ ИМПОРТ ШАБЛОНА"
echo "=================="
echo ""
cat << EOF
Configuration → Templates → Import

1. Загрузить файл: cloudbridge-relay-template-fixed.xml
2. Настройки импорта:
   - Update existing: Yes
   - Create new: Yes
   - Delete missing: No

3. После импорта применить к хосту cloudbridge-relay-01
EOF

echo ""
echo "5️⃣ ПРОВЕРКА ЭЛЕМЕНТОВ (ITEMS)"
echo "============================="
echo ""
cat << EOF
Configuration → Hosts → cloudbridge-relay-01 → Items

Ключевые элементы для проверки:
- cloudbridge.test.timestamp (должен быть Type: Zabbix agent, Update interval: 30s)
- cloudbridge.test.random (должен изменяться каждый раз)
- cloudbridge.relay.json.memory_live (должен показывать ~20-25 MB)
- cloudbridge.relay.json.goroutines_live (должен показывать количество горутин)
- cloudbridge.system.cpu.usage (использование CPU)
- cloudbridge.k8s.pods.running (количество запущенных pods)

Настройки каждого элемента:
- Type: Zabbix agent (passive)
- Update interval: 30s или меньше для тестирования
- Status: Enabled
EOF

echo ""
echo "6️⃣ МОНИТОРИНГ LATEST DATA"
echo "=========================="
echo ""
cat << EOF
Monitoring → Latest Data

Фильтры:
- Hosts: cloudbridge-relay-01
- Name: (оставить пустым для всех метрик)

Что проверить:
- Последнее время обновления (Last check)
- Значения должны изменяться в реальном времени
- Отсутствие ошибок (No errors)

Быстро изменяющиеся метрики для тестирования:
- cloudbridge.test.timestamp (каждые 30 сек новое значение)
- cloudbridge.test.random (случайное число 1-1000)
- cloudbridge.relay.json.memory_live (может изменяться)
EOF

echo ""
echo "7️⃣ СОЗДАНИЕ ТЕСТОВОГО ГРАФИКА"
echo "=============================="
echo ""
cat << EOF
Monitoring → Graphs → Create graph

Настройки:
- Name: CloudBridge Test Metrics
- Width: 900
- Height: 200

Items (элементы):
- cloudbridge.test.timestamp
- cloudbridge.relay.json.memory_live  
- cloudbridge.k8s.pods.running

График должен показывать изменения во времени.
EOF

echo ""
echo "8️⃣ ДИАГНОСТИКА ПРОБЛЕМ"
echo "======================="
echo ""
cat << EOF
Если данные не обновляются:

A) Проверить конфигурацию хоста:
   - IP адрес правильный: $AGENT_IP
   - Порт доступен: 10050  
   - Hostname совпадает: cloudbridge-relay-01

B) Проверить элементы:
   - Type = Zabbix agent (НЕ Zabbix agent (active))
   - Update interval не слишком большой (30s)
   - Status = Enabled

C) Проверить сетевое подключение:
   - ping $AGENT_IP
   - telnet $AGENT_IP 10050
   - iptables / firewall правила

D) Проверить логи:
   - /var/log/zabbix/zabbix_server.log
   - Ошибки подключения или таймауты

E) Перезапустить сервер (если необходимо):
   - sudo systemctl restart zabbix-server
EOF

echo ""
echo "9️⃣ КОМАНДЫ БЫСТРОЙ ДИАГНОСТИКИ"
echo "==============================="
echo ""
cat << EOF
# Все команды выполнять на сервере zbx.2gc.ru:

# Быстрая проверка агента:
zabbix_get -s $AGENT_IP -k agent.ping

# Тест изменяющихся данных (выполнить несколько раз):
zabbix_get -s $AGENT_IP -k cloudbridge.test.timestamp

# Проверка основных метрик:
for key in cloudbridge.test.random cloudbridge.relay.json.memory_live cloudbridge.k8s.pods.running; do
  echo "Testing \$key: \$(zabbix_get -s $AGENT_IP -k \$key)"
done

# Проверка логов (последние 20 строк):
sudo tail -20 /var/log/zabbix/zabbix_server.log

# Проверка активных соединений:
sudo netstat -an | grep :10051
EOF

echo ""
echo "🔟 ВОЗМОЖНЫЕ ПРИЧИНЫ ПРОБЛЕМЫ"
echo "============================="
echo ""
cat << EOF
1. ❌ Хост не создан в Zabbix UI
   ✅ Решение: Создать хост cloudbridge-relay-01

2. ❌ Неправильный IP адрес хоста  
   ✅ Решение: Использовать IP $AGENT_IP

3. ❌ Элементы в режиме Active вместо Passive
   ✅ Решение: Type = Zabbix agent (passive)

4. ❌ Шаблон не импортирован
   ✅ Решение: Импортировать cloudbridge-relay-template-fixed.xml

5. ❌ Слишком большой интервал обновления
   ✅ Решение: Update interval = 30s

6. ❌ Firewall блокирует порт 10050
   ✅ Решение: Открыть порт или отключить firewall

7. ❌ Zabbix server перегружен
   ✅ Решение: Проверить ресурсы сервера, перезапустить службы
EOF

echo ""
echo "🎉 ОЖИДАЕМЫЙ РЕЗУЛЬТАТ"
echo "======================"
echo ""
cat << EOF
После выполнения всех настроек:

✅ Хост cloudbridge-relay-01 должен быть активен (зеленый)
✅ Latest Data показывает обновляющиеся значения каждые 30 сек
✅ Графики отображают изменения во времени
✅ Логи сервера не содержат ошибок подключения

Метрики, которые должны изменяться в реальном времени:
- cloudbridge.test.timestamp: Unix timestamp (растет каждую секунду)
- cloudbridge.test.random: 1-1000 (случайное при каждом запросе)  
- cloudbridge.relay.json.memory_live: ~20-25 MB (может изменяться)
- cloudbridge.relay.json.goroutines_live: количество горутин
- cloudbridge.k8s.pods.running: количество запущенных pods

Если эти метрики работают, значит проблема решена!
EOF

echo ""
echo "📞 ТЕХНИЧЕСКАЯ ПОДДЕРЖКА"
echo "========================"
echo ""
echo "Если проблема не решена, предоставьте следующую информацию:"
echo "- Версия Zabbix сервера"
echo "- Результат команды: zabbix_get -s $AGENT_IP -k agent.ping"
echo "- Логи из /var/log/zabbix/zabbix_server.log"
echo "- Скриншот Latest Data с хостом cloudbridge-relay-01"
echo ""
echo "🎯 Агент готов! Все метрики работают локально."
echo "   Нужно только настроить их сбор на Zabbix сервере."



