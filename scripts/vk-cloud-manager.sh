#!/usr/bin/env bash

# Общий скрипт для управления балансировщиком и Anycast IP через API VK Cloud
# Источники: 
# - terraform-relay/docs-public/docs/ru/networks/balancing/instructions/manage-lb/manage-lb.md
# - terraform-relay/docs-public/docs/ru/networks/vnet/instructions/ip/anycast-ip/anycast-ip.md

# Загружаем переменные окружения
source ./setup-env.sh

# Функция для получения токена
get_token() {
    echo "Получение токена доступа..."
    TOKEN=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{
            \"auth\": {
                \"identity\": {
                    \"methods\": [\"password\"],
                    \"password\": {
                        \"user\": {
                            \"name\": \"$OS_USERNAME\",
                            \"domain\": {
                                \"name\": \"$OS_USER_DOMAIN_NAME\"
                            },
                            \"password\": \"$OS_PASSWORD\"
                        }
                    }
                },
                \"scope\": {
                    \"project\": {
                        \"id\": \"$OS_PROJECT_ID\"
                    }
                }
            }
        }" \
        "$OS_AUTH_URL/auth/tokens" | jq -r '.token.id')
    
    echo "Токен получен: ${TOKEN:0:20}..."
    export OS_TOKEN=$TOKEN
}

# === ФУНКЦИИ ДЛЯ БАЛАНСИРОВЩИКА НАГРУЗКИ ===

# Создание балансировщика
create_loadbalancer() {
    local name=$1
    local subnet_id=$2
    local availability_zone=${3:-"GZ1"}
    
    echo "Создание балансировщика '$name' в подсети $subnet_id..."
    
    LB_ID=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"loadbalancer\": {
                \"name\": \"$name\",
                \"vip_subnet_id\": \"$subnet_id\",
                \"availability_zone\": \"$availability_zone\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers" | jq -r '.loadbalancer.id')
    
    echo "Балансировщик создан с ID: $LB_ID"
    echo $LB_ID
}

# Список балансировщиков
list_loadbalancers() {
    echo "Получение списка балансировщиков..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers" | jq '.loadbalancers[] | {id, name, vip_address, status}'
}

# Создание прослушивателя
create_listener() {
    local lb_id=$1
    local name=$2
    local protocol=${3:-"HTTP"}
    local port=${4:-80}
    
    echo "Создание прослушивателя '$name' для балансировщика $lb_id..."
    
    LISTENER_ID=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"listener\": {
                \"name\": \"$name\",
                \"protocol\": \"$protocol\",
                \"protocol_port\": $port,
                \"loadbalancer_id\": \"$lb_id\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/listeners" | jq -r '.listener.id')
    
    echo "Прослушиватель создан с ID: $LISTENER_ID"
    echo $LISTENER_ID
}

# Создание пула
create_pool() {
    local listener_id=$1
    local name=$2
    local protocol=${3:-"HTTP"}
    local lb_method=${4:-"ROUND_ROBIN"}
    
    echo "Создание пула '$name' для прослушивателя $listener_id..."
    
    POOL_ID=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"pool\": {
                \"name\": \"$name\",
                \"protocol\": \"$protocol\",
                \"lb_algorithm\": \"$lb_method\",
                \"listener_id\": \"$listener_id\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/pools" | jq -r '.pool.id')
    
    echo "Пул создан с ID: $POOL_ID"
    echo $POOL_ID
}

# Добавление участника в пул
add_member() {
    local pool_id=$1
    local address=$2
    local port=${3:-80}
    local subnet_id=$4
    
    echo "Добавление участника $address:$port в пул $pool_id..."
    
    MEMBER_ID=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"member\": {
                \"address\": \"$address\",
                \"protocol_port\": $port,
                \"subnet_id\": \"$subnet_id\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/pools/$pool_id/members" | jq -r '.member.id')
    
    echo "Участник добавлен с ID: $MEMBER_ID"
    echo $MEMBER_ID
}

# === ФУНКЦИИ ДЛЯ ANYCAST IP ===

# Создание Anycast IP
create_anycast_ip() {
    local name=$1
    local description=${2:-"Anycast IP для отказоустойчивости"}
    
    echo "Создание Anycast IP '$name'..."
    
    ANYCAST_ID=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"anycastip\": {
                \"name\": \"$name\",
                \"description\": \"$description\"
            }
        }" \
        "$NEUTRON_URL/v2.0/anycastips" | jq -r '.anycastip.id')
    
    echo "Anycast IP создан с ID: $ANYCAST_ID"
    echo $ANYCAST_ID
}

# Список Anycast IP
list_anycast_ips() {
    echo "Получение списка Anycast IP..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips" | jq '.anycastips[] | {id, name, description, status}'
}

# Привязка IP к Anycast
associate_ip_to_anycast() {
    local anycast_id=$1
    local port_id=$2
    
    echo "Привязка порта $port_id к Anycast IP $anycast_id..."
    
    RESULT=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"port_id\": \"$port_id\"
        }" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id/associate")
    
    if echo "$RESULT" | jq -e '.anycastip' > /dev/null; then
        echo "IP адрес успешно привязан к Anycast IP"
        echo "$RESULT" | jq '.anycastip'
    else
        echo "❌ Ошибка привязки:"
        echo "$RESULT"
    fi
}

# Добавление healthcheck для Anycast
add_anycast_healthcheck() {
    local anycast_id=$1
    local port=${2:-80}
    local protocol=${3:-"tcp"}
    
    echo "Добавление healthcheck для Anycast IP $anycast_id..."
    
    RESULT=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"healthcheck\": {
                \"port\": $port,
                \"protocol\": \"$protocol\"
            }
        }" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id/healthcheck")
    
    if echo "$RESULT" | jq -e '.healthcheck' > /dev/null; then
        echo "Healthcheck успешно добавлен"
        echo "$RESULT" | jq '.healthcheck'
    else
        echo "❌ Ошибка добавления healthcheck:"
        echo "$RESULT"
    fi
}

# === ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ===

# Получение списка сетей
get_networks() {
    echo "Получение списка сетей..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/networks" | jq '.networks[] | {id, name, subnets}'
}

# Получение списка подсетей
get_subnets() {
    echo "Получение списка подсетей..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/subnets" | jq '.subnets[] | {id, name, cidr, network_id}'
}

# Получение списка портов
list_ports() {
    echo "Получение списка портов..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/ports" | jq '.ports[] | {id, name, fixed_ips, device_id}'
}

# Получение списка Floating IP
list_floating_ips() {
    echo "Получение списка Floating IP..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/floatingips" | jq '.floatingips[] | {id, floating_ip_address, port_id, status}'
}

# === АВТОМАТИЧЕСКАЯ НАСТРОЙКА ===

# Автоматическая настройка CloudBridge Relay с балансировщиком и Anycast
setup_cloudbridge_complete() {
    echo "🚀 Автоматическая настройка CloudBridge Relay с балансировщиком и Anycast..."
    
    # 1. Создание Anycast IP
    echo "1. Создание Anycast IP..."
    ANYCAST_ID=$(create_anycast_ip "cloudbridge-relay-anycast" "Anycast IP для CloudBridge Relay")
    
    if [ -z "$ANYCAST_ID" ]; then
        echo "❌ Ошибка создания Anycast IP"
        exit 1
    fi
    
    # 2. Добавление healthcheck для Anycast
    echo "2. Добавление healthcheck для Anycast IP..."
    add_anycast_healthcheck "$ANYCAST_ID" 80 "tcp"
    
    # 3. Получение информации о сетях
    echo "3. Получение информации о сетях..."
    echo "Доступные подсети:"
    get_subnets
    
    echo ""
    echo "Базовая настройка завершена!"
    echo "📍 Anycast IP ID: $ANYCAST_ID"
    echo ""
    echo " Следующие шаги:"
    echo "1. Создайте балансировщик: выберите 'Создать балансировщик'"
    echo "2. Привяжите IP адреса к Anycast: выберите 'Привязать IP к Anycast'"
    echo "3. Настройте правила балансировки"
}

# === ГЛАВНОЕ МЕНЮ ===

main_menu() {
    echo "=== VK Cloud Manager - Балансировщик и Anycast IP ==="
    echo ""
    echo "🔐 АУТЕНТИФИКАЦИЯ:"
    echo "1. Получить токен"
    echo ""
    echo "🌐 СЕТИ И РЕСУРСЫ:"
    echo "2. Показать сети"
    echo "3. Показать подсети"
    echo "4. Показать порты"
    echo "5. Показать Floating IP"
    echo ""
    echo "⚖️ БАЛАНСИРОВЩИК НАГРУЗКИ:"
    echo "6. Показать балансировщики"
    echo "7. Создать балансировщик"
    echo "8. Создать прослушиватель"
    echo "9. Создать пул"
    echo "10. Добавить участника в пул"
    echo ""
    echo "🌍 ANYCAST IP:"
    echo "11. Показать Anycast IP"
    echo "12. Создать Anycast IP"
    echo "13. Привязать IP к Anycast"
    echo "14. Добавить healthcheck для Anycast"
    echo ""
    echo "🚀 АВТОМАТИЧЕСКАЯ НАСТРОЙКА:"
    echo "15. Настроить CloudBridge Relay (полная настройка)"
    echo ""
    echo "0. Выход"
    echo ""
    echo "Выберите действие: "
    read choice
    
    case $choice in
        1) get_token ;;
        2) get_networks ;;
        3) get_subnets ;;
        4) list_ports ;;
        5) list_floating_ips ;;
        6) list_loadbalancers ;;
        7) 
            echo "Введите название балансировщика: "
            read name
            echo "Введите ID подсети: "
            read subnet_id
            echo "Введите зону доступности (по умолчанию GZ1): "
            read az
            az=${az:-GZ1}
            create_loadbalancer "$name" "$subnet_id" "$az"
            ;;
        8)
            echo "Введите ID балансировщика: "
            read lb_id
            echo "Введите название прослушивателя: "
            read name
            echo "Введите протокол (HTTP/HTTPS/TCP): "
            read protocol
            echo "Введите порт: "
            read port
            create_listener "$lb_id" "$name" "$protocol" "$port"
            ;;
        9)
            echo "Введите ID прослушивателя: "
            read listener_id
            echo "Введите название пула: "
            read name
            echo "Введите протокол (HTTP/HTTPS/TCP): "
            read protocol
            echo "Введите метод балансировки (ROUND_ROBIN/LEAST_CONNECTIONS/SOURCE_IP): "
            read lb_method
            create_pool "$listener_id" "$name" "$protocol" "$lb_method"
            ;;
        10)
            echo "Введите ID пула: "
            read pool_id
            echo "Введите IP адрес участника: "
            read address
            echo "Введите порт: "
            read port
            echo "Введите ID подсети: "
            read subnet_id
            add_member "$pool_id" "$address" "$port" "$subnet_id"
            ;;
        11) list_anycast_ips ;;
        12) 
            echo "Введите название Anycast IP: "
            read name
            echo "Введите описание (опционально): "
            read description
            create_anycast_ip "$name" "$description"
            ;;
        13)
            echo "Введите ID Anycast IP: "
            read anycast_id
            echo "Введите ID порта: "
            read port_id
            associate_ip_to_anycast "$anycast_id" "$port_id"
            ;;
        14)
            echo "Введите ID Anycast IP: "
            read anycast_id
            echo "Введите порт (по умолчанию 80): "
            read port
            port=${port:-80}
            echo "Введите протокол (tcp/icmp, по умолчанию tcp): "
            read protocol
            protocol=${protocol:-tcp}
            add_anycast_healthcheck "$anycast_id" "$port" "$protocol"
            ;;
        15) setup_cloudbridge_complete ;;
        0) exit 0 ;;
        *) echo "Неверный выбор" ;;
    esac
}

# Проверяем наличие jq
if ! command -v jq &> /dev/null; then
    echo "Ошибка: jq не установлен. Установите jq для работы с JSON."
    echo "Ubuntu/Debian: sudo apt-get install jq"
    echo "CentOS/RHEL: sudo yum install jq"
    exit 1
fi

# Запускаем меню
main_menu

