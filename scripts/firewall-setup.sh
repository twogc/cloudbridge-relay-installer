#!/bin/bash

# 🔥 CloudBridge Relay Firewall Setup Script
# Настройка Security Groups через API VK Cloud

# set -e  # Временно отключено для отладки

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    log "Проверяем зависимости..."
    
    if ! command -v curl &> /dev/null; then
        error "curl не установлен"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq не установлен"
        exit 1
    fi
    
    success "Все зависимости установлены"
}

# Загрузка переменных окружения
load_env() {
    log "Загружаем переменные окружения..."
    
    if [ -f "setup-env.sh" ]; then
        source setup-env.sh
    else
        error "Файл setup-env.sh не найден"
        exit 1
    fi
    
    # Проверка обязательных переменных
    if [ -z "$OS_TOKEN" ]; then
        error "OS_TOKEN не установлен"
        exit 1
    fi
    
    if [ -z "$NEUTRON_URL" ]; then
        error "NEUTRON_URL не установлен"
        exit 1
    fi
    
    success "Переменные окружения загружены"
}

# Функция для создания Security Group
create_security_group() {
    local name=$1
    local description=$2
    
    log "Создаем Security Group: $name"
    
    response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"security_group\": {
                \"name\": \"$name\",
                \"description\": \"$description\"
            }
        }" \
        "$NEUTRON_URL/v2.0/security-groups")
    
    if [ $? -ne 0 ]; then
        error "Ошибка при создании Security Group"
        return 1
    fi
    
    sg_id=$(echo $response | jq -r '.security_group.id')
    
    if [ "$sg_id" = "null" ] || [ -z "$sg_id" ]; then
        error "Не удалось получить ID Security Group"
        echo "Response: $response"
        return 1
    fi
    
    success "Security Group создана с ID: $sg_id"
    echo $sg_id
}

# Функция для создания правила
create_security_rule() {
    local sg_id=$1
    local direction=$2
    local protocol=$3
    local port_min=$4
    local port_max=$5
    local remote_ip=$6
    
    log "Создаем правило: $protocol:$port_min-$port_max ($direction)"
    
    response=$(curl -s -X POST \
        -H "X-Auth-Token: $OS_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"security_group_rule\": {
                \"security_group_id\": \"$sg_id\",
                \"direction\": \"$direction\",
                \"protocol\": \"$protocol\",
                \"port_range_min\": $port_min,
                \"port_range_max\": $port_max,
                \"remote_ip_prefix\": \"$remote_ip\"
            }
        }" \
        "$NEUTRON_URL/v2.0/security-group-rules")
    
    if [ $? -ne 0 ]; then
        error "Ошибка при создании правила"
        return 1
    fi
    
    rule_id=$(echo $response | jq -r '.security_group_rule.id')
    
    if [ "$rule_id" = "null" ] || [ -z "$rule_id" ]; then
        error "Не удалось создать правило"
        echo "Response: $response"
        return 1
    fi
    
    success "Правило создано с ID: $rule_id"
}

# Функция для проверки существования Security Group
check_security_group_exists() {
    local name=$1
    
    log "Проверяем существование Security Group: $name"
    
    response=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/security-groups?name=$name")
    
    sg_id=$(echo $response | jq -r '.security_groups[0].id')
    
    if [ "$sg_id" != "null" ] && [ -n "$sg_id" ]; then
        warning "Security Group '$name' уже существует с ID: $sg_id"
        echo $sg_id
        return 0
    fi
    
    return 1
}

# Основная функция настройки
setup_firewall() {
    log "🚀 Начинаем настройку firewall для CloudBridge Relay"
    
    # Создаем основную Security Group
    relay_sg_id=$(check_security_group_exists "cloudbridge-relay-sg")
    if [ -z "$relay_sg_id" ]; then
        relay_sg_id=$(create_security_group "cloudbridge-relay-sg" "Security group for CloudBridge Relay services")
    fi
    
    # Создаем административную Security Group
    admin_sg_id=$(check_security_group_exists "cloudbridge-admin-sg")
    if [ -z "$admin_sg_id" ]; then
        admin_sg_id=$(create_security_group "cloudbridge-admin-sg" "Administrative security group for CloudBridge Relay")
    fi
    
    log "Настраиваем правила для основной Security Group..."
    
    # Входящие правила для Relay API
    create_security_rule $relay_sg_id "ingress" "tcp" 8080 8080 "0.0.0.0/0"
    create_security_rule $relay_sg_id "ingress" "tcp" 8081 8081 "0.0.0.0/0"
    create_security_rule $relay_sg_id "ingress" "udp" 51820 51820 "0.0.0.0/0"
    create_security_rule $relay_sg_id "ingress" "udp" 9090 9090 "0.0.0.0/0"
    create_security_rule $relay_sg_id "ingress" "tcp" 80 80 "0.0.0.0/0"
    create_security_rule $relay_sg_id "ingress" "tcp" 443 443 "0.0.0.0/0"
    
    # Исходящие правила
    create_security_rule $relay_sg_id "egress" "tcp" 1 65535 "0.0.0.0/0"
    create_security_rule $relay_sg_id "egress" "udp" 1 65535 "0.0.0.0/0"
    
    log "Настраиваем правила для административной Security Group..."
    
    # Административные правила (ограниченный доступ)
    create_security_rule $admin_sg_id "ingress" "tcp" 22 22 "0.0.0.0/0"
    create_security_rule $admin_sg_id "ingress" "tcp" 8080 8080 "0.0.0.0/0"
    create_security_rule $admin_sg_id "ingress" "tcp" 8081 8081 "0.0.0.0/0"
    create_security_rule $admin_sg_id "egress" "tcp" 1 65535 "0.0.0.0/0"
    create_security_rule $admin_sg_id "egress" "udp" 1 65535 "0.0.0.0/0"
    
    success "Firewall настроен успешно!"
    
    echo ""
    echo "Созданные Security Groups:"
    echo "  - cloudbridge-relay-sg: $relay_sg_id"
    echo "  - cloudbridge-admin-sg: $admin_sg_id"
    echo ""
    echo "🔧 Следующие шаги:"
    echo "  1. Применить Security Groups к Load Balancer"
    echo "  2. Применить Security Groups к Kubernetes узлам"
    echo "  3. Протестировать доступность сервисов"
    echo "  4. Настроить мониторинг"
}

# Функция для проверки статуса
check_status() {
    log "Проверяем статус Security Groups..."
    
    response=$(curl -s -H "X-Auth-Token: $OS_TOKEN" \
        "$NEUTRON_URL/v2.0/security-groups")
    
    echo $response | jq -r '.security_groups[] | "\(.name): \(.id)"'
}

# Функция для удаления Security Group
delete_security_group() {
    local name=$1
    
    log "Удаляем Security Group: $name"
    
    sg_id=$(check_security_group_exists "$name")
    if [ $? -eq 0 ]; then
        curl -s -X DELETE \
            -H "X-Auth-Token: $OS_TOKEN" \
            "$NEUTRON_URL/v2.0/security-groups/$sg_id"
        
        success "Security Group '$name' удалена"
    else
        warning "Security Group '$name' не найдена"
    fi
}

# Главная функция
main() {
    case "${1:-setup}" in
        "setup")
            check_dependencies
            load_env
            setup_firewall
            ;;
        "status")
            load_env
            check_status
            ;;
        "delete")
            load_env
            delete_security_group "cloudbridge-relay-sg"
            delete_security_group "cloudbridge-admin-sg"
            ;;
        "help"|"-h"|"--help")
            echo "Использование: $0 [команда]"
            echo ""
            echo "Команды:"
            echo "  setup   - Настроить firewall (по умолчанию)"
            echo "  status  - Показать статус Security Groups"
            echo "  delete  - Удалить Security Groups"
            echo "  help    - Показать эту справку"
            ;;
        *)
            error "Неизвестная команда: $1"
            echo "Используйте '$0 help' для справки"
            exit 1
            ;;
    esac
}

# Запуск скрипта
main "$@"
