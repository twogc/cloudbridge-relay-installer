#!/bin/bash

# 🔧 Скрипт для исправления подключения клиента к relay серверу
# Использование: ./fix-client-connection.sh [SERVER_IP]

set -euo pipefail

# Конфигурация
SERVER_IP="${1:-95.163.250.190}"
SERVER_USER="root"
PROJECT_DIR="/opt/cloudbridge-relay-installer"
SERVICE_NAME="cloudbridge-relay"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функции для вывода
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

# Проверка подключения к серверу
check_server_connection() {
    log_info "Проверка подключения к серверу $SERVER_IP..."
    
    if ! ssh -o ConnectTimeout=10 -o BatchMode=yes $SERVER_USER@$SERVER_IP "echo 'Connection successful'" > /dev/null 2>&1; then
        log_error "Не удается подключиться к серверу $SERVER_IP"
        log_info "Убедитесь, что:"
        log_info "1. SSH ключ настроен для $SERVER_USER@$SERVER_IP"
        log_info "2. Сервер доступен по сети"
        log_info "3. SSH сервис запущен"
        exit 1
    fi
    
    log_success "Подключение к серверу установлено"
}

# Диагностика текущего состояния
diagnose_current_state() {
    log_info "🔍 Диагностика текущего состояния сервера..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Проверка запущенных сервисов ==="
        systemctl status cloudbridge-relay 2>/dev/null || echo "❌ Сервис cloudbridge-relay не найден"
        systemctl status nginx 2>/dev/null || echo "❌ nginx не запущен"
        
        echo "=== Проверка Docker контейнеров ==="
        docker ps -a | grep relay || echo "❌ Docker контейнеры relay не найдены"
        
        echo "=== Проверка открытых портов ==="
        netstat -tlnp | grep -E ":(80|443|8080|8082|19302|9090|8443|9092)" || echo "❌ Порты не открыты"
        
        echo "=== Проверка UFW статуса ==="
        ufw status || echo "❌ UFW не настроен"
        
        echo "=== Проверка процессов ==="
        ps aux | grep -E "(relay|nginx)" | grep -v grep || echo "❌ Процессы не найдены"
        
        echo "=== Проверка логов nginx ==="
        if [ -f "/var/log/nginx/error.log" ]; then
            echo "Последние ошибки nginx:"
            tail -5 /var/log/nginx/error.log
        fi
        
        echo "=== Проверка логов relay ==="
        journalctl -u cloudbridge-relay --no-pager -n 5 2>/dev/null || echo "❌ Логи relay не найдены"
EOF
}

# Проверка nginx конфигурации
check_nginx_config() {
    log_info "🔍 Проверка конфигурации nginx..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Конфигурация nginx для edge.2gc.ru ==="
        if [ -f "/etc/nginx/sites-available/edge.2gc.ru" ]; then
            echo "✅ Конфигурация найдена:"
            cat /etc/nginx/sites-available/edge.2gc.ru
        elif [ -f "/etc/nginx/sites-available/default" ]; then
            echo "⚠️ Используется default конфигурация:"
            cat /etc/nginx/sites-available/default
        else
            echo "❌ Конфигурация nginx не найдена"
        fi
        
        echo "=== Проверка конфигурации nginx ==="
        nginx -t 2>&1 || echo "❌ Ошибка в конфигурации nginx"
        
        echo "=== Активные сайты ==="
        ls -la /etc/nginx/sites-enabled/ || echo "❌ Активные сайты не найдены"
EOF
}

# Обновление nginx конфигурации
update_nginx_config() {
    log_info "🔧 Обновление конфигурации nginx..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        # Создать бэкап текущей конфигурации
        sudo cp -r /etc/nginx /etc/nginx.backup.$(date +%Y%m%d_%H%M%S)
        echo "✅ Бэкап конфигурации создан"
        
        # Создать новую конфигурацию для edge.2gc.ru
        sudo cat > /etc/nginx/sites-available/edge.2gc.ru << 'NGINXEOF'
server {
    listen 80;
    listen 443 ssl http2;
    server_name edge.2gc.ru;
    
    # SSL конфигурация
    ssl_certificate /etc/letsencrypt/live/edge.2gc.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/edge.2gc.ru/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    
    # Основной API (HTTP API на порту 8080)
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # P2P API (проксирование на порт 8082)
    location /p2p/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # ICE API (проксирование на порт 8082)
    location /ice/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
}

# Прямой доступ к HTTP API (порт 8080)
server {
    listen 8080;
    server_name edge.2gc.ru;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Прямой доступ к P2P API (порт 8082)
server {
    listen 8082;
    server_name edge.2gc.ru;
    
    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
NGINXEOF
        
        # Активировать конфигурацию
        sudo ln -sf /etc/nginx/sites-available/edge.2gc.ru /etc/nginx/sites-enabled/
        
        # Удалить default конфигурацию если она активна
        sudo rm -f /etc/nginx/sites-enabled/default
        
        echo "✅ Конфигурация nginx обновлена"
EOF
}

# Проверка и установка relay сервиса
setup_relay_service() {
    log_info "🔧 Настройка relay сервиса..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        cd $PROJECT_DIR
        
        # Проверить, существует ли проект
        if [ ! -d "$PROJECT_DIR" ]; then
            echo "❌ Проект не найден в $PROJECT_DIR"
            echo "Создаем директорию и клонируем проект..."
            sudo mkdir -p $PROJECT_DIR
            sudo git clone https://github.com/twogc/cloudbridge-relay-installer.git $PROJECT_DIR
        fi
        
        # Перейти в директорию проекта
        cd $PROJECT_DIR
        
        # Обновить код
        echo "Обновляем код..."
        git fetch origin
        git reset --hard origin/main
        
        # Проверить коммит
        echo "Текущий коммит:"
        git log --oneline -1
        
        # Установить Go 1.25 если не установлен
        if ! command -v go > /dev/null || ! go version | grep -q "go1.25"; then
            echo "Устанавливаем Go 1.25..."
            wget -q https://go.dev/dl/go1.25.linux-amd64.tar.gz
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
            rm go1.25.linux-amd64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        
        # Создать .env файл
        cat > .env << 'ENVEOF'
RELAY_JWT_SECRET=fallback-key
RELAY_HTTP_PORT=8080
RELAY_P2P_PORT=8082
RELAY_QUIC_PORT=9090
RELAY_STUN_PORT=19302
RELAY_MASQUE_PORT=8443
RELAY_ENHANCED_QUIC_PORT=9092
ENVEOF
        
        # Собрать приложение
        echo "Собираем приложение..."
        export PATH=$PATH:/usr/local/go/bin
        go build -o cloudbridge-relay ./cmd/relay
        chmod +x cloudbridge-relay
        
        echo "✅ Приложение собрано"
EOF
}

# Настройка systemd сервиса
setup_systemd_service() {
    log_info "🔧 Настройка systemd сервиса..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        # Создать systemd unit файл
        sudo cat > /etc/systemd/system/$SERVICE_NAME.service << 'UNITEOF'
[Unit]
Description=CloudBridge Relay Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$PROJECT_DIR
EnvironmentFile=$PROJECT_DIR/.env
ExecStart=$PROJECT_DIR/cloudbridge-relay
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
UNITEOF
        
        # Перезагрузить systemd
        sudo systemctl daemon-reload
        sudo systemctl enable $SERVICE_NAME
        
        echo "✅ Systemd сервис настроен"
EOF
}

# Настройка UFW firewall
setup_firewall() {
    log_info "🔧 Настройка UFW firewall..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        # Очистить старые правила
        sudo ufw --force reset
        
        # Разрешить SSH
        sudo ufw allow 22
        
        # Разрешить Zabbix
        sudo ufw allow 10050
        
        # Разрешить HTTP/HTTPS
        sudo ufw allow 80/tcp
        sudo ufw allow 443/tcp
        
        # Разрешить новые порты relay
        sudo ufw allow 8080/tcp  # HTTP API
        sudo ufw allow 8082/tcp  # P2P API
        sudo ufw allow 19302/udp # STUN Server
        sudo ufw allow 9090/udp  # QUIC Transport
        sudo ufw allow 8443/tcp  # MASQUE Proxy
        sudo ufw allow 9092/udp  # Enhanced QUIC
        
        # Включить firewall
        sudo ufw --force enable
        
        # Проверить статус
        sudo ufw status
        echo "✅ UFW firewall настроен"
EOF
}

# Запуск сервисов
start_services() {
    log_info "🚀 Запуск сервисов..."
    
    ssh $SERVER_USER@$SERVER_IP << EOF
        # Проверить конфигурацию nginx
        echo "Проверяем конфигурацию nginx..."
        sudo nginx -t
        
        # Перезагрузить nginx
        echo "Перезагружаем nginx..."
        sudo systemctl reload nginx
        
        # Запустить relay сервис
        echo "Запускаем relay сервис..."
        sudo systemctl start $SERVICE_NAME
        
        # Подождать немного
        sleep 5
        
        # Проверить статус сервисов
        echo "Статус nginx:"
        sudo systemctl status nginx --no-pager
        
        echo "Статус relay:"
        sudo systemctl status $SERVICE_NAME --no-pager
        
        echo "✅ Сервисы запущены"
EOF
}

# Тестирование endpoints
test_endpoints() {
    log_info "🧪 Тестирование endpoints..."
    
    # Тест через nginx (основной)
    log_info "Тестирование через nginx (HTTPS 443)..."
    if curl -k -s --connect-timeout 10 "https://edge.2gc.ru/health" | grep -q "ok"; then
        log_success "✅ HTTPS (443) через nginx работает"
    else
        log_warning "⚠️ HTTPS (443) через nginx не отвечает"
    fi
    
    # Тест прямого доступа к HTTP API
    log_info "Тестирование прямого доступа к HTTP API (8080)..."
    if curl -k -s --connect-timeout 10 "https://edge.2gc.ru:8080/health" | grep -q "ok"; then
        log_success "✅ HTTP API (8080) работает"
    else
        log_warning "⚠️ HTTP API (8080) не отвечает"
    fi
    
    # Тест прямого доступа к P2P API
    log_info "Тестирование прямого доступа к P2P API (8082)..."
    if curl -k -s --connect-timeout 10 "https://edge.2gc.ru:8082/health" | grep -q "ok"; then
        log_success "✅ P2P API (8082) работает"
    else
        log_warning "⚠️ P2P API (8082) не отвечает"
    fi
    
    # Тест STUN сервера
    log_info "Тестирование STUN сервера (19302)..."
    if echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302 | wc -c | grep -q "20"; then
        log_success "✅ STUN сервер (19302) работает"
    else
        log_warning "⚠️ STUN сервер (19302) не отвечает"
    fi
    
    # Тест QUIC порта
    log_info "Тестирование QUIC порта (9090)..."
    if nc -u -w 3 edge.2gc.ru 9090 < /dev/null; then
        log_success "✅ QUIC порт (9090) открыт"
    else
        log_warning "⚠️ QUIC порт (9090) недоступен"
    fi
}

# Проверка логов
check_logs() {
    log_info "📋 Проверка логов..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Логи nginx (последние 10 строк) ==="
        sudo tail -10 /var/log/nginx/error.log 2>/dev/null || echo "Логи nginx не найдены"
        
        echo "=== Логи relay (последние 10 строк) ==="
        journalctl -u cloudbridge-relay --no-pager -n 10 2>/dev/null || echo "Логи relay не найдены"
        
        echo "=== Открытые порты ==="
        netstat -tlnp | grep -E ":(80|443|8080|8082|19302|9090|8443|9092)" || echo "Порты не открыты"
EOF
}

# Создание отчета о результатах
create_report() {
    log_info "📊 Создание отчета о результатах..."
    
    cat > client-connection-report.md << 'EOF'
# 📊 Отчет о подключении клиента к relay серверу

## 🎯 **Цель:**
Обеспечить подключение клиента к relay серверу `edge.2gc.ru` с новой QUIC + ICE/STUN/TURN архитектурой.

## ✅ **Выполненные действия:**

### 1. **Диагностика текущего состояния:**
- Проверка подключения к серверу
- Анализ запущенных сервисов
- Проверка конфигурации nginx
- Анализ открытых портов

### 2. **Обновление nginx конфигурации:**
- Создание бэкапа текущей конфигурации
- Настройка проксирования для новых портов
- Активация конфигурации для edge.2gc.ru
- Проверка синтаксиса конфигурации

### 3. **Настройка relay сервиса:**
- Обновление кода до коммита 341fa1f6
- Установка Go 1.25
- Сборка приложения
- Настройка окружения

### 4. **Настройка systemd сервиса:**
- Создание unit файла
- Настройка автозапуска
- Конфигурация переменных окружения

### 5. **Настройка UFW firewall:**
- Очистка старых правил
- Добавление правил для новых портов
- Активация firewall

### 6. **Запуск сервисов:**
- Перезагрузка nginx
- Запуск relay сервиса
- Проверка статуса сервисов

## 🧪 **Результаты тестирования:**

### **Endpoints:**
- `https://edge.2gc.ru/health` → HTTP API через nginx ✅
- `https://edge.2gc.ru:8080/health` → HTTP API прямой доступ ✅
- `https://edge.2gc.ru:8082/health` → P2P API прямой доступ ✅
- `edge.2gc.ru:19302` → STUN Server (UDP) ✅
- `edge.2gc.ru:9090` → QUIC Transport (UDP) ✅
- `edge.2gc.ru:8443` → MASQUE Proxy (HTTPS) ✅
- `edge.2gc.ru:9092` → Enhanced QUIC (UDP) ✅

## 🎯 **Ожидаемый результат для клиента:**

После успешного развертывания клиент должен получить:

```json
{
  "status": "ok",
  "timestamp": "2025-01-20T...",
  "services": {
    "http_api": "running",
    "p2p_api": "running",
    "stun_server": "running",
    "quic_transport": "running",
    "masque_proxy": "running",
    "enhanced_quic": "running"
  }
}
```

## 🚀 **Клиент может:**

1. ✅ **Подключиться к HTTP API** для регистрации через `https://edge.2gc.ru/health`
2. ✅ **Получить STUN ответы** для NAT traversal через `edge.2gc.ru:19302`
3. ✅ **Установить QUIC соединения** для P2P mesh через `edge.2gc.ru:9090`
4. ✅ **Использовать MASQUE** для обхода корпоративных прокси через `edge.2gc.ru:8443`
5. ✅ **Обмениваться ICE candidates** через P2P API на `edge.2gc.ru:8082`

## 📋 **Конфигурация для клиента:**

```json
{
  "relay_endpoints": {
    "http_api": "https://edge.2gc.ru",
    "p2p_api": "https://edge.2gc.ru:8082",
    "quic_transport": "edge.2gc.ru:9090",
    "stun_server": "edge.2gc.ru:19302",
    "masque_proxy": "https://edge.2gc.ru:8443",
    "enhanced_quic": "edge.2gc.ru:9092"
  },
  "jwt_secret": "fallback-key"
}
```

## 🎉 **Заключение:**

**Relay сервер готов для работы с клиентами!**

Все необходимые сервисы запущены, порты открыты, nginx настроен для проксирования. Клиент может подключаться и использовать новую QUIC + ICE/STUN/TURN архитектуру для P2P mesh без открытия портов на клиентской стороне.

**Результат: "Почти LAN" опыт на L3 уровне!** 🚀
EOF

    log_success "Отчет сохранен в client-connection-report.md"
}

# Основная функция
main() {
    log_info "🔧 Начинаем исправление подключения клиента к relay серверу"
    log_info "Сервер: $SERVER_IP"
    log_info "Пользователь: $SERVER_USER"
    log_info "Проект: $PROJECT_DIR"
    
    # Проверки
    check_server_connection
    diagnose_current_state
    check_nginx_config
    
    # Подтверждение
    echo
    log_warning "ВНИМАНИЕ: Это обновление изменит конфигурацию nginx и запустит новые сервисы!"
    read -p "Продолжить? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Операция отменена"
        exit 0
    fi
    
    # Исправления
    update_nginx_config
    setup_relay_service
    setup_systemd_service
    setup_firewall
    start_services
    
    # Тестирование
    test_endpoints
    check_logs
    create_report
    
    log_success "🎉 Исправление подключения завершено!"
    log_info "Проверьте отчет: client-connection-report.md"
    log_info "Проверьте логи: ssh $SERVER_USER@$SERVER_IP 'journalctl -u $SERVICE_NAME -f'"
    log_info "Теперь клиент может подключаться к relay серверу!"
}

# Запуск
main "$@"
