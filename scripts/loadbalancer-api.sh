#!/usr/bin/env bash

# Скрипт для управления балансировщиком нагрузки через API VK Cloud
# Источник: terraform-relay/docs-public/docs/ru/networks/balancing/instructions/manage-lb/manage-lb.md

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

# Функция для получения списка сетей
get_networks() {
    echo "Получение списка сетей..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/networks" | jq '.networks[] | {id, name, subnets}'
}

# Функция для получения списка подсетей
get_subnets() {
    echo "Получение списка подсетей..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/subnets" | jq '.subnets[] | {id, name, cidr, network_id}'
}

# Функция для создания балансировщика
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

# Функция для получения списка балансировщиков
list_loadbalancers() {
    echo "Получение списка балансировщиков..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers" | jq '.loadbalancers[] | {id, name, vip_address, status}'
}

# Функция для получения информации о балансировщике
show_loadbalancer() {
    local lb_id=$1
    echo "Информация о балансировщике $lb_id..."
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers/$lb_id" | jq '.loadbalancer'
}

# Функция для создания прослушивателя
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

# Функция для создания пула
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

# Функция для добавления участника в пул
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

# Функция для удаления балансировщика
delete_loadbalancer() {
    local lb_id=$1
    echo "Удаление балансировщика $lb_id..."
    curl -s -X DELETE \
        -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers/$lb_id"
    echo "Балансировщик удален"
}

# Главное меню
main_menu() {
    echo "=== Управление балансировщиком нагрузки VK Cloud ==="
    echo "1. Получить токен"
    echo "2. Показать сети"
    echo "3. Показать подсети"
    echo "4. Создать балансировщик"
    echo "5. Показать балансировщики"
    echo "6. Показать информацию о балансировщике"
    echo "7. Создать прослушиватель"
    echo "8. Создать пул"
    echo "9. Добавить участника"
    echo "10. Удалить балансировщик"
    echo "0. Выход"
    echo "Выберите действие: "
    read choice
    
    case $choice in
        1) get_token ;;
        2) get_networks ;;
        3) get_subnets ;;
        4) 
            echo "Введите название балансировщика: "
            read name
            echo "Введите ID подсети: "
            read subnet_id
            create_loadbalancer $name $subnet_id
            ;;
        5) list_loadbalancers ;;
        6) 
            echo "Введите ID балансировщика: "
            read lb_id
            show_loadbalancer $lb_id
            ;;
        7)
            echo "Введите ID балансировщика: "
            read lb_id
            echo "Введите название прослушивателя: "
            read name
            echo "Введите протокол (HTTP/HTTPS/TCP): "
            read protocol
            echo "Введите порт: "
            read port
            create_listener $lb_id $name $protocol $port
            ;;
        8)
            echo "Введите ID прослушивателя: "
            read listener_id
            echo "Введите название пула: "
            read name
            echo "Введите протокол (HTTP/HTTPS/TCP): "
            read protocol
            echo "Введите метод балансировки (ROUND_ROBIN/LEAST_CONNECTIONS/SOURCE_IP): "
            read lb_method
            create_pool $listener_id $name $protocol $lb_method
            ;;
        9)
            echo "Введите ID пула: "
            read pool_id
            echo "Введите IP адрес участника: "
            read address
            echo "Введите порт: "
            read port
            echo "Введите ID подсети: "
            read subnet_id
            add_member $pool_id $address $port $subnet_id
            ;;
        10)
            echo "Введите ID балансировщика: "
            read lb_id
            delete_loadbalancer $lb_id
            ;;
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

