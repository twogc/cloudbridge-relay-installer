#!/usr/bin/env bash

# Комплексный скрипт для настройки CloudBridge Relay с балансировщиком и Anycast IP
# Автор: VK Cloud API Manager
# Дата: $(date)

set -e

# Загружаем переменные окружения
source ./setup-env.sh

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Функция для ожидания готовности балансировщика
wait_for_loadbalancer() {
    local lb_id=$1
    local max_attempts=30
    local attempt=1
    
    log_info "Ожидание готовности балансировщика $lb_id..."
    
    while [ $attempt -le $max_attempts ]; do
        status=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/loadbalancers/$lb_id" | \
            jq -r '.loadbalancer.provisioning_status')
        
        if [ "$status" = "ACTIVE" ]; then
            log_success "Балансировщик $lb_id готов!"
            return 0
        elif [ "$status" = "ERROR" ]; then
            log_error "Ошибка создания балансировщика $lb_id"
            return 1
        else
            log_info "Попытка $attempt/$max_attempts: статус $status"
            sleep 10
            ((attempt++))
        fi
    done
    
    log_error "Таймаут ожидания готовности балансировщика"
    return 1
}

# Функция для получения информации о сетях
get_network_info() {
    log_info "Получение информации о сетях..."
    
    # Получаем cloudbridge-network
    cloudbridge_network=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/networks" | \
        jq -r '.networks[] | select(.name == "cloudbridge-network") | {id, name, subnets}')
    
    if [ -z "$cloudbridge_network" ]; then
        log_error "Сеть cloudbridge-network не найдена"
        return 1
    fi
    
    network_id=$(echo "$cloudbridge_network" | jq -r '.id')
    subnet_id=$(echo "$cloudbridge_network" | jq -r '.subnets[0]')
    
    log_success "Найдена сеть: $network_id, подсеть: $subnet_id"
    
    export CLOUDBRIDGE_NETWORK_ID=$network_id
    export CLOUDBRIDGE_SUBNET_ID=$subnet_id
}

# Функция для получения информации об Anycast IP
get_anycast_info() {
    log_info "Получение информации об Anycast IP..."
    
    anycast_info=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/anycastips" | \
        jq -r '.anycastips[] | select(.name == "2GC_RELAY")')
    
    if [ -z "$anycast_info" ]; then
        log_warning "Anycast IP 2GC_RELAY не найден"
        return 1
    fi
    
    anycast_id=$(echo "$anycast_info" | jq -r '.id')
    anycast_ip=$(echo "$anycast_info" | jq -r '.ip_address')
    
    log_success "Найден Anycast IP: $anycast_id ($anycast_ip)"
    
    export ANYCAST_ID=$anycast_id
    export ANYCAST_IP=$anycast_ip
}

# Функция для создания балансировщика
create_loadbalancer() {
    local name=$1
    local subnet_id=$2
    local availability_zone=${3:-"MS1"}
    
    log_info "Создание балансировщика '$name' в подсети $subnet_id..."
    
    lb_response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"loadbalancer\": {
                \"name\": \"$name\",
                \"vip_subnet_id\": \"$subnet_id\",
                \"availability_zone\": \"$availability_zone\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers")
    
    lb_id=$(echo "$lb_response" | jq -r '.loadbalancer.id')
    lb_vip=$(echo "$lb_response" | jq -r '.loadbalancer.vip_address')
    
    if [ "$lb_id" = "null" ] || [ -z "$lb_id" ]; then
        log_error "Ошибка создания балансировщика"
        echo "$lb_response"
        return 1
    fi
    
    log_success "Балансировщик создан: $lb_id ($lb_vip)"
    
    # Ждем готовности балансировщика
    wait_for_loadbalancer "$lb_id"
    
    export LOADBALANCER_ID=$lb_id
    export LOADBALANCER_VIP=$lb_vip
}

# Функция для создания прослушивателя
create_listener() {
    local lb_id=$1
    local name=$2
    local protocol=${3:-"HTTP"}
    local port=${4:-80}
    
    log_info "Создание прослушивателя '$name' для балансировщика $lb_id..."
    
    listener_response=$(curl -s -X POST \
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
        "$OCTAVIA_URL/v2.0/lbaas/listeners")
    
    listener_id=$(echo "$listener_response" | jq -r '.listener.id')
    
    if [ "$listener_id" = "null" ] || [ -z "$listener_id" ]; then
        log_error "Ошибка создания прослушивателя"
        echo "$listener_response"
        return 1
    fi
    
    log_success "Прослушиватель создан: $listener_id"
    
    export LISTENER_ID=$listener_id
}

# Функция для создания пула
create_pool() {
    local listener_id=$1
    local name=$2
    local protocol=${3:-"HTTP"}
    local lb_method=${4:-"ROUND_ROBIN"}
    
    log_info "Создание пула '$name' для прослушивателя $listener_id..."
    
    pool_response=$(curl -s -X POST \
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
        "$OCTAVIA_URL/v2.0/lbaas/pools")
    
    pool_id=$(echo "$pool_response" | jq -r '.pool.id')
    
    if [ "$pool_id" = "null" ] || [ -z "$pool_id" ]; then
        log_error "Ошибка создания пула"
        echo "$pool_response"
        return 1
    fi
    
    log_success "Пул создан: $pool_id"
    
    export POOL_ID=$pool_id
}

# Функция для добавления участника в пул
add_member() {
    local pool_id=$1
    local address=$2
    local port=${3:-80}
    local subnet_id=$4
    
    log_info "Добавление участника $address:$port в пул $pool_id..."
    
    member_response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"member\": {
                \"address\": \"$address\",
                \"protocol_port\": $port,
                \"subnet_id\": \"$subnet_id\"
            }
        }" \
        "$OCTAVIA_URL/v2.0/lbaas/pools/$pool_id/members")
    
    member_id=$(echo "$member_response" | jq -r '.member.id')
    
    if [ "$member_id" = "null" ] || [ -z "$member_id" ]; then
        log_error "Ошибка добавления участника"
        echo "$member_response"
        return 1
    fi
    
    log_success "Участник добавлен: $member_id"
    
    export MEMBER_ID=$member_id
}

# Функция для привязки балансировщика к Anycast IP
associate_loadbalancer_to_anycast() {
    local anycast_id=$1
    local lb_id=$2
    
    log_info "Привязка балансировщика $lb_id к Anycast IP $anycast_id..."
    
    # Получаем порт балансировщика
    lb_port=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers/$lb_id" | \
        jq -r '.loadbalancer.vip_port_id')
    
    if [ "$lb_port" = "null" ] || [ -z "$lb_port" ]; then
        log_error "Не удалось получить порт балансировщика"
        return 1
    fi
    
    # Привязываем к Anycast IP
    associate_response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"port_id\": \"$lb_port\"
        }" \
        "$NEUTRON_URL/v2.0/anycastips/$anycast_id/associate")
    
    if echo "$associate_response" | jq -e '.anycastip' > /dev/null; then
        log_success "Балансировщик успешно привязан к Anycast IP"
    else
        log_error "Ошибка привязки к Anycast IP"
        echo "$associate_response"
        return 1
    fi
}

# Функция для отображения статуса
show_status() {
    log_info "=== Статус CloudBridge Relay ==="
    
    echo ""
    echo "Anycast IP:"
    if [ -n "$ANYCAST_ID" ]; then
        curl -s -H "X-Auth-Token: $OS_TOKEN" \
            "$NEUTRON_URL/v2.0/anycastips/$ANYCAST_ID" | \
            jq '.anycastip | {id, name, ip_address, status, healthcheck}'
    else
        echo "Не настроен"
    fi
    
    echo ""
    echo "⚖️ Балансировщики:"
    curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$OCTAVIA_URL/v2.0/lbaas/loadbalancers" | \
        jq '.loadbalancers[] | {id, name, vip_address, status, availability_zone}'
    
    echo ""
    echo " Прослушиватели:"
    if [ -n "$LISTENER_ID" ]; then
        curl -s -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/listeners/$LISTENER_ID" | \
            jq '.listener | {id, name, protocol, protocol_port, status}'
    else
        echo "Не настроены"
    fi
    
    echo ""
    echo "🏊 Пул:"
    if [ -n "$POOL_ID" ]; then
        curl -s -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/pools/$POOL_ID" | \
            jq '.pool | {id, name, protocol, lb_algorithm, status}'
    else
        echo "Не настроен"
    fi
}

# Главная функция настройки
setup_cloudbridge_relay() {
    log_info "Настройка CloudBridge Relay с балансировщиком и Anycast IP"
    
    # 1. Получаем информацию о сетях
    get_network_info
    
    # 2. Получаем информацию об Anycast IP
    get_anycast_info
    
    # 3. Создаем балансировщик
    create_loadbalancer "cloudbridge-relay-lb" "$CLOUDBRIDGE_SUBNET_ID" "MS1"
    
    # 4. Создаем прослушиватель
    create_listener "$LOADBALANCER_ID" "cloudbridge-relay-listener" "HTTP" 80
    
    # 5. Создаем пул
    create_pool "$LISTENER_ID" "cloudbridge-relay-pool" "HTTP" "ROUND_ROBIN"
    
    # 6. Добавляем участников (пример с IP адресами)
    # add_member "$POOL_ID" "192.168.1.10" 80 "$CLOUDBRIDGE_SUBNET_ID"
    # add_member "$POOL_ID" "192.168.1.11" 80 "$CLOUDBRIDGE_SUBNET_ID"
    
    # 7. Привязываем к Anycast IP
    if [ -n "$ANYCAST_ID" ]; then
        associate_loadbalancer_to_anycast "$ANYCAST_ID" "$LOADBALANCER_ID"
    fi
    
    # 8. Показываем статус
    show_status
    
    log_success "Настройка CloudBridge Relay завершена!"
    
    echo ""
    echo "Сводка:"
    echo "Anycast IP: $ANYCAST_IP"
    echo "Балансировщик: $LOADBALANCER_VIP"
    echo "Прослушиватель: $LISTENER_ID"
    echo "Пул: $POOL_ID"
}

# Функция для очистки
cleanup() {
    log_warning "Удаление ресурсов CloudBridge Relay..."
    
    if [ -n "$POOL_ID" ]; then
        curl -s -X DELETE -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/pools/$POOL_ID"
        log_info "Пул $POOL_ID удален"
    fi
    
    if [ -n "$LISTENER_ID" ]; then
        curl -s -X DELETE -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/listeners/$LISTENER_ID"
        log_info "Прослушиватель $LISTENER_ID удален"
    fi
    
    if [ -n "$LOADBALANCER_ID" ]; then
        curl -s -X DELETE -H "X-Auth-Token: $OS_TOKEN" \
            "$OCTAVIA_URL/v2.0/lbaas/loadbalancers/$LOADBALANCER_ID"
        log_info "Балансировщик $LOADBALANCER_ID удален"
    fi
}

# Главное меню
main_menu() {
    echo "=== CloudBridge Relay Setup ==="
    echo "1. Полная настройка CloudBridge Relay"
    echo "2. Показать статус"
    echo "3. Очистка ресурсов"
    echo "4. Тест подключения"
    echo "0. Выход"
    echo ""
    echo "Выберите действие: "
    read choice
    
    case $choice in
        1) setup_cloudbridge_relay ;;
        2) show_status ;;
        3) cleanup ;;
        4) 
            log_info "Тест подключения к Anycast IP..."
            if [ -n "$ANYCAST_IP" ]; then
                curl -s --connect-timeout 5 "http://$ANYCAST_IP" || log_warning "Anycast IP недоступен"
            else
                log_warning "Anycast IP не настроен"
            fi
            ;;
        0) exit 0 ;;
        *) log_error "Неверный выбор" ;;
    esac
}

# Проверяем наличие jq
if ! command -v jq &> /dev/null; then
    log_error "jq не установлен. Установите jq для работы с JSON."
    exit 1
fi

# Запускаем меню
main_menu

