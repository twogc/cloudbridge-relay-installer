#!/bin/bash

# Установка Zabbix Agent на виртуальную машину CloudBridge Relay
# Этот скрипт нужно выполнить на виртуальной машине CloudBridge

set -e

# Конфигурация
ZABBIX_SERVER_IP="10.0.0.241"
ZABBIX_SERVER_PORT="10051"
ZABBIX_AGENT_PORT="10050"
HOSTNAME="CloudBridge-Relay-VM"

echo "🚀 Установка Zabbix Agent на CloudBridge Relay VM..."

# Определение операционной системы
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$NAME
    VER=$VERSION_ID
else
    echo "❌ Не удалось определить операционную систему"
    exit 1
fi

echo "Обнаружена ОС: $OS $VER"

# Обновление системы
echo "📦 Обновление системы..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    sudo apt update && sudo apt upgrade -y
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]] || [[ "$OS" == *"Rocky"* ]]; then
    sudo yum update -y
else
    echo "❌ Неподдерживаемая операционная система: $OS"
    exit 1
fi

# Установка необходимых пакетов
echo "📦 Установка необходимых пакетов..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    sudo apt install -y wget curl gnupg2 software-properties-common apt-transport-https ca-certificates lsb-release
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]] || [[ "$OS" == *"Rocky"* ]]; then
    sudo yum install -y wget curl gnupg2
fi

# Добавление репозитория Zabbix
echo "📥 Добавление репозитория Zabbix..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    wget https://repo.zabbix.com/zabbix/6.0/ubuntu/pool/main/z/zabbix-release/zabbix-release_6.0-4+ubuntu22.04_all.deb
    sudo dpkg -i zabbix-release_6.0-4+ubuntu22.04_all.deb
    sudo apt update
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]] || [[ "$OS" == *"Rocky"* ]]; then
    sudo rpm -Uvh https://repo.zabbix.com/zabbix/6.0/rhel/8/x86_64/zabbix-release-6.0-4.el8.noarch.rpm
    sudo yum clean all
fi

# Установка Zabbix Agent
echo "📦 Установка Zabbix Agent..."
if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
    sudo apt install -y zabbix-agent
elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]] || [[ "$OS" == *"Rocky"* ]]; then
    sudo yum install -y zabbix-agent
fi

# Создание директории для логов
echo "📁 Создание директорий..."
sudo mkdir -p /var/log/zabbix
sudo chown zabbix:zabbix /var/log/zabbix

# Настройка конфигурации Zabbix Agent
echo "⚙️ Настройка конфигурации Zabbix Agent..."
sudo tee /etc/zabbix/zabbix_agentd.conf > /dev/null <<EOF
# Основные настройки
PidFile=/var/run/zabbix/zabbix_agentd.pid
LogFile=/var/log/zabbix/zabbix_agentd.log
LogFileSize=10
DebugLevel=3

# Настройки сервера
Server=$ZABBIX_SERVER_IP
ServerActive=$ZABBIX_SERVER_IP
ListenPort=$ZABBIX_AGENT_PORT
ListenIP=0.0.0.0

# Имя хоста
Hostname=$HOSTNAME

# Настройки безопасности
EnableRemoteCommands=0
LogRemoteCommands=1
AllowRoot=0

# Настройки производительности
StartAgents=3
Timeout=4
BufferSize=100

# CloudBridge специфичные настройки
UserParameter=cloudbridge.relay.status,curl -s http://localhost:8080/health | grep -o '"status":"[^"]*"' | cut -d'"' -f4
UserParameter=cloudbridge.relay.connections,netstat -an | grep :8080 | grep ESTABLISHED | wc -l
UserParameter=cloudbridge.relay.uptime,ps -o etime= -p \$(pgrep -f cloudbridge-relay) | tr -d ' '
UserParameter=cloudbridge.docker.containers,docker ps --format "table {{.Names}}\t{{.Status}}" | grep -v NAMES | wc -l
UserParameter=cloudbridge.docker.images,docker images --format "table {{.Repository}}" | grep -v REPOSITORY | wc -l
UserParameter=cloudbridge.disk.usage,df -h / | awk 'NR==2 {print \$5}' | sed 's/%//'
UserParameter=cloudbridge.memory.usage,free | awk 'NR==2{printf "%.2f", \$3*100/\$2}'
UserParameter=cloudbridge.cpu.usage,top -bn1 | grep "Cpu(s)" | awk '{print \$2}' | awk -F'%' '{print \$1}'

# Kubernetes метрики (если используется)
UserParameter=k8s.pods.running,kubectl get pods --all-namespaces --field-selector=status.phase=Running | grep -v NAME | wc -l
UserParameter=k8s.pods.total,kubectl get pods --all-namespaces | grep -v NAME | wc -l
UserParameter=k8s.nodes.ready,kubectl get nodes --field-selector=status.conditions[0].type=Ready | grep -v NAME | wc -l

# WireGuard метрики
UserParameter=wireguard.interfaces,wg show | grep interface | wc -l
UserParameter=wireguard.peers,wg show | grep peer | wc -l

# Сетевые метрики
UserParameter=network.connections.total,netstat -an | grep ESTABLISHED | wc -l
UserParameter=network.connections.listening,netstat -an | grep LISTEN | wc -l
UserParameter=network.ports.8080,netstat -an | grep :8080 | grep LISTEN | wc -l
UserParameter=network.ports.8081,netstat -an | grep :8081 | grep LISTEN | wc -l
UserParameter=network.ports.3456,netstat -an | grep :3456 | grep LISTEN | wc -l

# Логи и ошибки
UserParameter=cloudbridge.log.errors,tail -n 100 /var/log/cloudbridge/error.log 2>/dev/null | grep -i error | wc -l || echo 0
UserParameter=cloudbridge.log.warnings,tail -n 100 /var/log/cloudbridge/error.log 2>/dev/null | grep -i warning | wc -l || echo 0

# Производительность приложения
UserParameter=cloudbridge.response.time,curl -o /dev/null -s -w '%{time_total}' http://localhost:8080/health
UserParameter=cloudbridge.api.requests,curl -s http://localhost:8081/metrics | grep 'http_requests_total' | awk '{print \$2}' || echo 0
EOF

# Создание пользовательских скриптов для мониторинга
echo "📝 Создание пользовательских скриптов мониторинга..."
sudo mkdir -p /etc/zabbix/scripts

# Скрипт для проверки статуса CloudBridge
sudo tee /etc/zabbix/scripts/cloudbridge_status.sh > /dev/null <<'EOF'
#!/bin/bash
# Скрипт для проверки статуса CloudBridge Relay

# Проверка статуса сервиса
if systemctl is-active --quiet cloudbridge-relay; then
    echo "1"
else
    echo "0"
fi
EOF

# Скрипт для проверки Docker контейнеров
sudo tee /etc/zabbix/scripts/docker_status.sh > /dev/null <<'EOF'
#!/bin/bash
# Скрипт для проверки статуса Docker контейнеров

# Проверка статуса Docker
if systemctl is-active --quiet docker; then
    # Количество запущенных контейнеров
    running=$(docker ps -q | wc -l)
    echo "$running"
else
    echo "0"
fi
EOF

# Скрипт для проверки Kubernetes
sudo tee /etc/zabbix/scripts/k8s_status.sh > /dev/null <<'EOF'
#!/bin/bash
# Скрипт для проверки статуса Kubernetes

# Проверка статуса Kubernetes
if command -v kubectl &> /dev/null; then
    # Количество подов в состоянии Running
    running_pods=$(kubectl get pods --all-namespaces --field-selector=status.phase=Running --no-headers | wc -l)
    echo "$running_pods"
else
    echo "0"
fi
EOF

# Установка прав на скрипты
sudo chmod +x /etc/zabbix/scripts/*.sh
sudo chown -R zabbix:zabbix /etc/zabbix/scripts

# Настройка файрвола
echo "🔥 Настройка файрвола..."
if command -v ufw &> /dev/null; then
    sudo ufw allow $ZABBIX_AGENT_PORT/tcp
    echo "UFW: Разрешен порт $ZABBIX_AGENT_PORT"
elif command -v firewall-cmd &> /dev/null; then
    sudo firewall-cmd --permanent --add-port=$ZABBIX_AGENT_PORT/tcp
    sudo firewall-cmd --reload
    echo "Firewalld: Разрешен порт $ZABBIX_AGENT_PORT"
elif command -v iptables &> /dev/null; then
    sudo iptables -A INPUT -p tcp --dport $ZABBIX_AGENT_PORT -j ACCEPT
    echo "iptables: Разрешен порт $ZABBIX_AGENT_PORT"
fi

# Запуск и включение Zabbix Agent
echo "🚀 Запуск Zabbix Agent..."
sudo systemctl start zabbix-agent
sudo systemctl enable zabbix-agent

# Проверка статуса
echo "Проверка статуса Zabbix Agent..."
sudo systemctl status zabbix-agent --no-pager -l

# Проверка подключения к серверу
echo "🔍 Проверка подключения к Zabbix Server..."
if nc -z $ZABBIX_SERVER_IP $ZABBIX_SERVER_PORT; then
    echo "Подключение к Zabbix Server ($ZABBIX_SERVER_IP:$ZABBIX_SERVER_PORT) успешно"
else
    echo "❌ Не удается подключиться к Zabbix Server ($ZABBIX_SERVER_IP:$ZABBIX_SERVER_PORT)"
    echo "   Проверьте сетевую связность и настройки файрвола"
fi

# Проверка порта агента
echo "🔍 Проверка порта Zabbix Agent..."
if netstat -tlnp | grep :$ZABBIX_AGENT_PORT; then
    echo "Zabbix Agent слушает на порту $ZABBIX_AGENT_PORT"
else
    echo "❌ Zabbix Agent не слушает на порту $ZABBIX_AGENT_PORT"
fi

# Тестирование пользовательских параметров
echo "🧪 Тестирование пользовательских параметров..."
echo "CloudBridge статус: $(sudo -u zabbix /usr/sbin/zabbix_agentd -t cloudbridge.relay.status)"
echo "Docker контейнеры: $(sudo -u zabbix /usr/sbin/zabbix_agentd -t cloudbridge.docker.containers)"
echo "Использование диска: $(sudo -u zabbix /usr/sbin/zabbix_agentd -t cloudbridge.disk.usage)%"
echo "Использование памяти: $(sudo -u zabbix /usr/sbin/zabbix_agentd -t cloudbridge.memory.usage)%"

echo ""
echo "🎉 Установка Zabbix Agent завершена!"
echo ""
echo "Информация:"
echo "   Hostname: $HOSTNAME"
echo "   Zabbix Server: $ZABBIX_SERVER_IP:$ZABBIX_SERVER_PORT"
echo "   Agent Port: $ZABBIX_AGENT_PORT"
echo "   Config: /etc/zabbix/zabbix_agentd.conf"
echo "   Logs: /var/log/zabbix/zabbix_agentd.log"
echo ""
echo " Полезные команды:"
echo "   sudo systemctl status zabbix-agent"
echo "   sudo systemctl restart zabbix-agent"
echo "   sudo tail -f /var/log/zabbix/zabbix_agentd.log"
echo "   sudo -u zabbix /usr/sbin/zabbix_agentd -t cloudbridge.relay.status"
echo ""
echo "  Следующие шаги:"
echo "1. Добавьте хост '$HOSTNAME' в Zabbix Server"
echo "2. Примените шаблон 'CloudBridge Relay Server'"
echo "3. Проверьте сбор метрик в веб-интерфейсе"
echo ""
echo "🌐 Веб-интерфейс Zabbix: http://$ZABBIX_SERVER_IP/zabbix"




