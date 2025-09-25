#!/usr/bin/env bash

# Скрипт для управления Anycast IP через API VK Cloud
# Источник: terraform-relay/docs-public/docs/ru/networks/vnet/instructions/ip/anycast-ip/anycast-ip.md

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

# Функция для получения списка Anycast IP
list_anycast_ips() {
    echo "Получение списка Anycast IP..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips" | jq '.anycastips[] | {id, name, description, status}'
}

# Функция для создания Anycast IP
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

# Функция для получения информации об Anycast IP
show_anycast_ip() {
    local anycast_id=$1
    echo "Информация об Anycast IP $anycast_id..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id" | jq '.anycastip'
}

# Функция для получения списка портов
list_ports() {
    echo "Получение списка портов..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/ports" | jq '.ports[] | {id, name, fixed_ips, device_id}'
}

# Функция для получения списка Floating IP
list_floating_ips() {
    echo "Получение списка Floating IP..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/floatingips" | jq '.floatingips[] | {id, floating_ip_address, port_id, status}'
}

# Функция для привязки IP адреса к Anycast IP
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

# Функция для отвязки IP адреса от Anycast IP
disassociate_ip_from_anycast() {
    local anycast_id=$1
    local port_id=$2
    
    echo "Отвязка порта $port_id от Anycast IP $anycast_id..."
    
    RESULT=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"port_id\": \"$port_id\"
        }" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id/disassociate")
    
    if echo "$RESULT" | jq -e '.anycastip' > /dev/null; then
        echo "IP адрес успешно отвязан от Anycast IP"
        echo "$RESULT" | jq '.anycastip'
    else
        echo "❌ Ошибка отвязки:"
        echo "$RESULT"
    fi
}

# Функция для добавления healthcheck
add_healthcheck() {
    local anycast_id=$1
    local port=${2:-80}
    local protocol=${3:-"tcp"}
    
    echo "Добавление healthcheck для Anycast IP $anycast_id..."
    echo "Порт: $port, Протокол: $protocol"
    
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

# Функция для получения информации о healthcheck
show_healthcheck() {
    local anycast_id=$1
    echo "Информация о healthcheck для Anycast IP $anycast_id..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id/healthcheck" | jq '.healthcheck'
}

# Функция для удаления Anycast IP
delete_anycast_ip() {
    local anycast_id=$1
    echo "Удаление Anycast IP $anycast_id..."
    curl -s -X DELETE \
        -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id"
    echo "Anycast IP удален"
}

# Функция для проверки статуса Anycast IP
check_anycast_status() {
    local anycast_id=$1
    echo "Проверка статуса Anycast IP $anycast_id..."
    
    ANYCAST_INFO=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id")
    
    if echo "$ANYCAST_INFO" | jq -e '.anycastip' > /dev/null; then
        echo "=== Информация об Anycast IP ==="
        echo "$ANYCAST_INFO" | jq '.anycastip | {id, name, description, status, created_at}'
        
        echo ""
        echo "=== Привязанные IP адреса ==="
        echo "$ANYCAST_INFO" | jq '.anycastip.associated_ips[]? | {port_id, ip_address, status}'
        
        echo ""
        echo "=== Healthcheck ==="
        HEALTHCHECK=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
            "$NEUTRON_URL/v2.0/anycastips/$anycast_id/healthcheck")
        if echo "$HEALTHCHECK" | jq -e '.healthcheck' > /dev/null; then
            echo "$HEALTHCHECK" | jq '.healthcheck | {port, protocol, status}'
        else
            echo "Healthcheck не настроен"
        fi
    else
        echo "❌ Anycast IP не найден"
    fi
}

# Функция для автоматической настройки Anycast IP для CloudBridge Relay
setup_cloudbridge_anycast() {
    echo "🚀 Автоматическая настройка Anycast IP для CloudBridge Relay..."
    
    # Создаем Anycast IP
    echo "1. Создание Anycast IP..."
    ANYCAST_ID=$(create_anycast_ip "cloudbridge-relay-anycast" "Anycast IP для CloudBridge Relay - отказоустойчивая маршрутизация")
    
    if [ -z "$ANYCAST_ID" ]; then
        echo "❌ Ошибка создания Anycast IP"
        exit 1
    fi
    
    echo "2. Получение списка LoadBalancer сервисов..."
    # Здесь можно добавить логику для получения IP адресов LoadBalancer
    # из Kubernetes или других источников
    
    echo "3. Добавление healthcheck..."
    add_healthcheck "$ANYCAST_ID" 80 "tcp"
    
    echo "4. Проверка статуса..."
    check_anycast_status "$ANYCAST_ID"
    
    echo ""
    echo "Anycast IP настроен!"
    echo "📍 ID: $ANYCAST_ID"
    echo ""
    echo " Для привязки IP адресов используйте:"
    echo "   ./anycast-api.sh associate $ANYCAST_ID <port_id>"
}

# Главное меню
main_menu() {
    echo "=== Управление Anycast IP VK Cloud ==="
    echo "1. Получить токен"
    echo "2. Показать Anycast IP"
    echo "3. Создать Anycast IP"
    echo "4. Показать информацию об Anycast IP"
    echo "5. Показать порты"
    echo "6. Показать Floating IP"
    echo "7. Привязать IP к Anycast"
    echo "8. Отвязать IP от Anycast"
    echo "9. Добавить healthcheck"
    echo "10. Показать healthcheck"
    echo "11. Проверить статус Anycast IP"
    echo "12. Удалить Anycast IP"
    echo "13. Автоматическая настройка для CloudBridge Relay"
    echo "0. Выход"
    echo "Выберите действие: "
    read choice
    
    case $choice in
        1) get_token ;;
        2) list_anycast_ips ;;
        3) 
            echo "Введите название Anycast IP: "
            read name
            echo "Введите описание (опционально): "
            read description
            create_anycast_ip "$name" "$description"
            ;;
        4) 
            echo "Введите ID Anycast IP: "
            read anycast_id
            show_anycast_ip "$anycast_id"
            ;;
        5) list_ports ;;
        6) list_floating_ips ;;
        7)
            echo "Введите ID Anycast IP: "
            read anycast_id
            echo "Введите ID порта: "
            read port_id
            associate_ip_to_anycast "$anycast_id" "$port_id"
            ;;
        8)
            echo "Введите ID Anycast IP: "
            read anycast_id
            echo "Введите ID порта: "
            read port_id
            disassociate_ip_from_anycast "$anycast_id" "$port_id"
            ;;
        9)
            echo "Введите ID Anycast IP: "
            read anycast_id
            echo "Введите порт (по умолчанию 80): "
            read port
            port=${port:-80}
            echo "Введите протокол (tcp/icmp, по умолчанию tcp): "
            read protocol
            protocol=${protocol:-tcp}
            add_healthcheck "$anycast_id" "$port" "$protocol"
            ;;
        10)
            echo "Введите ID Anycast IP: "
            read anycast_id
            show_healthcheck "$anycast_id"
            ;;
        11)
            echo "Введите ID Anycast IP: "
            read anycast_id
            check_anycast_status "$anycast_id"
            ;;
        12)
            echo "Введите ID Anycast IP: "
            read anycast_id
            delete_anycast_ip "$anycast_id"
            ;;
        13) setup_cloudbridge_anycast ;;
        0) exit 0 ;;
        *) echo "Неверный выбор" ;;
    esac
}

# Проверяем наличие jq
if ! command -v jq &> /dev/null; then
    echo "Ошибка: jq не установлен. Установите jq для работы с JSON."
    exit 1
fi

# Запускаем меню
main_menu

