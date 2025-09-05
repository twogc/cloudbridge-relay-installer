#!/bin/bash

# CloudBridge Relay - Автоматизированное развертывание
# Этот скрипт полностью автоматизирует развертывание CloudBridge Relay сервера

set -e

echo "🚀 CloudBridge Relay - Автоматизированное развертывание"
echo "====================================================="

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода с цветом
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка зависимостей
check_dependencies() {
    print_status "Проверка зависимостей..."
    
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform не установлен"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker не установлен"
        exit 1
    fi
    
    if [ ! -f "~/.ssh/cloudbridge-relay" ]; then
        print_error "SSH ключ ~/.ssh/cloudbridge-relay не найден"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
}

# Проверка аутентификации
check_auth() {
    print_status "Проверка аутентификации VK Cloud..."
    
    if ! source ../auth.sh; then
        print_error "Ошибка аутентификации в VK Cloud"
        exit 1
    fi
    
    print_success "Аутентификация в VK Cloud успешна"
}

# Инициализация Terraform
init_terraform() {
    print_status "Инициализация Terraform..."
    
    if [ ! -d ".terraform" ]; then
        terraform init
    else
        print_status "Terraform уже инициализирован"
    fi
    
    print_success "Terraform инициализирован"
}

# Планирование развертывания
plan_deployment() {
    print_status "Планирование развертывания..."
    
    terraform plan -out=tfplan
    
    print_success "План развертывания создан"
}

# Развертывание инфраструктуры
deploy_infrastructure() {
    print_status "Развертывание инфраструктуры..."
    
    terraform apply tfplan
    
    print_success "Инфраструктура развернута"
}

# Ожидание готовности VM
wait_for_vm() {
    print_status "Ожидание готовности виртуальной машины..."
    
    # Получаем IP адрес из Terraform output
    VM_IP=$(terraform output -raw relay_ddos_fip_address 2>/dev/null || echo "")
    
    if [ -z "$VM_IP" ]; then
        print_error "Не удалось получить IP адрес VM"
        exit 1
    fi
    
    print_status "IP адрес VM: $VM_IP"
    
    # Ждем, пока VM будет доступна по SSH
    print_status "Ожидание доступности SSH..."
    for i in {1..30}; do
        if ssh -i ~/.ssh/cloudbridge-relay -o ConnectTimeout=10 -o StrictHostKeyChecking=no ubuntu@$VM_IP "echo 'SSH доступен'" 2>/dev/null; then
            print_success "SSH доступен на VM"
            break
        fi
        
        if [ $i -eq 30 ]; then
            print_error "SSH недоступен после 30 попыток"
            exit 1
        fi
        
        print_status "Попытка $i/30 - SSH недоступен, ожидание..."
        sleep 10
    done
}

# Автоматическая установка CloudBridge Relay
install_cloudbridge_relay() {
    print_status "Автоматическая установка CloudBridge Relay..."
    
    VM_IP=$(terraform output -raw relay_ddos_fip_address)
    
    # Ждем завершения cloud-init
    print_status "Ожидание завершения cloud-init..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "cloud-init status --wait"
    
    # Устанавливаем Docker и зависимости
    print_status "Установка Docker и зависимостей..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "
        sudo apt update
        sudo apt install -y docker.io docker-compose curl wget git htop vim net-tools ufw nginx fail2ban ca-certificates gnupg lsb-release
        sudo usermod -aG docker ubuntu
        sudo systemctl start docker && sudo systemctl enable docker
        echo 'Docker установлен успешно!'
    "
    
    # Настраиваем UFW firewall
    print_status "Настройка UFW firewall..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "
        sudo ufw --force reset
        sudo ufw default deny incoming
        sudo ufw default allow outgoing
        sudo ufw allow 22/tcp comment 'SSH'
        sudo ufw allow 80/tcp comment 'HTTP'
        sudo ufw allow 443/tcp comment 'HTTPS'
        sudo ufw allow 8080/tcp comment 'Relay Server API'
        sudo ufw allow 8081/tcp comment 'Relay Server Metrics'
        sudo ufw allow 3456/tcp comment 'Relay Server Custom'
        sudo ufw allow 9090/udp comment 'QUIC Protocol'
        sudo ufw allow 51820/udp comment 'WireGuard P2P Mesh'
        sudo ufw --force enable
        echo 'UFW настроен успешно!'
    "
    
    # Создаем директории и устанавливаем CloudBridge Relay
    print_status "Установка CloudBridge Relay..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "
        sudo mkdir -p /opt/cloudbridge-relay
        sudo mkdir -p /opt/cloudbridge-relay/configs
        sudo mkdir -p /opt/cloudbridge-relay/logs
        sudo chown -R ubuntu:ubuntu /opt/cloudbridge-relay
        cd /opt/cloudbridge-relay
        echo 'Создание docker-compose.yml...'
        cat > docker-compose.yml << 'EOF'
version: '3.8'
services:
  cloudbridge-relay:
    image: registry.vkcs.ru/cloudbridge-relay:latest
    container_name: cloudbridge-relay
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    devices:
      - /dev/net/tun:/dev/net/tun
    ports:
      - '8080:8080'
      - '8081:8081'
      - '3456:3456'
      - '9090:9090/udp'
      - '51820:51820/udp'
    volumes:
      - ./configs:/app/configs:ro
      - ./logs:/app/logs
      - /etc/wireguard:/etc/wireguard:ro
      - /var/run/wireguard:/var/run/wireguard
    environment:
      - CONFIG_PATH=/app/configs/config.json
      - LOG_LEVEL=info
      - WIREGUARD_ENABLED=true
      - TZ=UTC
    healthcheck:
      test: ['CMD', 'curl', '-f', 'http://localhost:8080/health']
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '0.5'
          memory: 512M
    logging:
      driver: 'json-file'
      options:
        max-size: '10m'
        max-file: '3'
    networks:
      - cloudbridge-network

networks:
  cloudbridge-network:
    driver: bridge
EOF
        echo 'Docker Compose конфигурация создана!'
    "
    
    # Запускаем CloudBridge Relay
    print_status "Запуск CloudBridge Relay..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "
        cd /opt/cloudbridge-relay
        sudo docker-compose up -d
        echo 'CloudBridge Relay запущен!'
    "
    
    # Создаем скрипты управления
    print_status "Создание скриптов управления..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "
        cat > /tmp/relay-status.sh << 'EOF'
#!/bin/bash
echo '=== CloudBridge Relay Status ==='
echo 'Nginx: \$(systemctl is-active nginx)'
echo 'UFW: \$(sudo ufw status | grep Status)'
echo 'Fail2ban: \$(systemctl is-active fail2ban)'
echo 'Docker: \$(systemctl is-active docker)'
echo ''
echo '=== Listening Ports ==='
sudo ss -tlnp | grep -E ':(80|443|8080|8081|3456|9090|51820)'
echo ''
echo '=== Docker Containers ==='
sudo docker ps -a
EOF

        cat > /tmp/ddos-status.sh << 'EOF'
#!/bin/bash
echo '=== DDoS Protection Status ==='
echo 'nginx: \$(systemctl is-active nginx)'
echo 'fail2ban: \$(systemctl is-active fail2ban)'
echo 'ufw: \$(sudo ufw status | grep Status)'
echo 'docker: \$(systemctl is-active docker)'
echo ''
echo '=== Active UFW Rules ==='
sudo ufw status numbered
echo ''
echo '=== Docker Containers ==='
sudo docker ps -a
EOF

        sudo mv /tmp/relay-status.sh /usr/local/bin/relay-status
        sudo mv /tmp/ddos-status.sh /usr/local/bin/ddos-status
        sudo chmod +x /usr/local/bin/relay-status
        sudo chmod +x /usr/local/bin/ddos-status
        echo 'Скрипты управления созданы!'
    "
    
    print_success "CloudBridge Relay установлен и настроен"
}

# Проверка статуса развертывания
check_deployment_status() {
    print_status "Проверка статуса развертывания..."
    
    VM_IP=$(terraform output -raw relay_ddos_fip_address)
    
    print_status "Проверка сервисов..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "sudo relay-status"
    
    print_status "Проверка DDoS защиты..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "sudo ddos-status"
    
    print_status "Проверка Docker контейнеров..."
    ssh -i ~/.ssh/cloudbridge-relay -o StrictHostKeyChecking=no ubuntu@$VM_IP "sudo docker ps -a"
    
    print_success "Проверка статуса завершена"
}

# Вывод информации о развертывании
show_deployment_info() {
    print_status "Информация о развертывании..."
    
    VM_IP=$(terraform output -raw relay_ddos_fip_address)
    
    echo ""
    echo "🎉 CloudBridge Relay успешно развернут!"
    echo "====================================="
    echo "IP адрес: $VM_IP"
    echo "SSH подключение: ssh ubuntu@$VM_IP"
    echo ""
    echo "Доступные сервисы:"
    echo "- HTTP: http://$VM_IP/"
    echo "- Relay API: http://$VM_IP:8080/"
    echo "- Metrics: http://$VM_IP:8081/"
    echo ""
    echo "Управление:"
    echo "- Статус: sudo relay-status"
    echo "- DDoS защита: sudo ddos-status"
    echo "- Docker: sudo docker ps -a"
    echo ""
    echo "Автоматизация: Полностью автоматизированное развертывание!"
}

# Основная функция
main() {
    echo ""
    print_status "Начинаем автоматизированное развертывание..."
    echo ""
    
    check_dependencies
    echo ""
    
    check_auth
    echo ""
    
    init_terraform
    echo ""
    
    plan_deployment
    echo ""
    
    deploy_infrastructure
    echo ""
    
    wait_for_vm
    echo ""
    
    install_cloudbridge_relay
    echo ""
    
    check_deployment_status
    echo ""
    
    show_deployment_info
    echo ""
    
    print_success "Развертывание завершено успешно!"
}

# Запуск основной функции
main "$@"
