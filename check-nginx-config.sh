#!/bin/bash

# 🔍 Скрипт проверки конфигурации nginx на продакшен сервере
# Использование: ./check-nginx-config.sh [SERVER_IP]

set -euo pipefail

# Конфигурация
SERVER_IP="${1:-95.163.250.190}"
SERVER_USER="root"

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

# Проверка статуса nginx
check_nginx_status() {
    log_info "Проверка статуса nginx..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Статус nginx ==="
        systemctl status nginx --no-pager || echo "nginx не запущен"
        
        echo "=== Проверка конфигурации nginx ==="
        nginx -t 2>&1 || echo "Ошибка в конфигурации nginx"
        
        echo "=== Процессы nginx ==="
        ps aux | grep nginx | grep -v grep || echo "Процессы nginx не найдены"
        
        echo "=== Открытые порты nginx ==="
        netstat -tlnp | grep nginx || echo "Порты nginx не найдены"
EOF
}

# Проверка конфигурационных файлов nginx
check_nginx_configs() {
    log_info "Проверка конфигурационных файлов nginx..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Основная конфигурация nginx ==="
        if [ -f "/etc/nginx/nginx.conf" ]; then
            echo "Файл: /etc/nginx/nginx.conf"
            cat /etc/nginx/nginx.conf
        else
            echo "Файл /etc/nginx/nginx.conf не найден"
        fi
        
        echo "=== Конфигурации сайтов ==="
        if [ -d "/etc/nginx/sites-available" ]; then
            echo "Доступные сайты:"
            ls -la /etc/nginx/sites-available/
            
            echo "Активные сайты:"
            ls -la /etc/nginx/sites-enabled/
            
            echo "=== Конфигурация для edge.2gc.ru ==="
            if [ -f "/etc/nginx/sites-available/edge.2gc.ru" ]; then
                cat /etc/nginx/sites-available/edge.2gc.ru
            elif [ -f "/etc/nginx/sites-available/default" ]; then
                cat /etc/nginx/sites-available/default
            else
                echo "Конфигурация для edge.2gc.ru не найдена"
            fi
        fi
        
        echo "=== Конфигурации в conf.d ==="
        if [ -d "/etc/nginx/conf.d" ]; then
            ls -la /etc/nginx/conf.d/
            for file in /etc/nginx/conf.d/*; do
                if [ -f "$file" ]; then
                    echo "=== Файл: $file ==="
                    cat "$file"
                fi
            done
        fi
EOF
}

# Проверка текущих портов и сервисов
check_current_services() {
    log_info "Проверка текущих сервисов и портов..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Все открытые порты ==="
        netstat -tlnp | grep LISTEN
        
        echo "=== Проверка портов relay ==="
        netstat -tlnp | grep -E ":(8080|8082|19302|9090|8443|9092)" || echo "Порты relay не открыты"
        
        echo "=== Проверка systemd сервисов ==="
        systemctl list-units --type=service | grep -E "(relay|nginx|cloudbridge)" || echo "Сервисы не найдены"
        
        echo "=== Проверка Docker контейнеров ==="
        docker ps -a | grep -E "(relay|nginx|cloudbridge)" || echo "Docker контейнеры не найдены"
        
        echo "=== Проверка процессов ==="
        ps aux | grep -E "(relay|nginx|cloudbridge)" | grep -v grep || echo "Процессы не найдены"
EOF
}

# Проверка логов nginx
check_nginx_logs() {
    log_info "Проверка логов nginx..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== Последние логи nginx (access) ==="
        if [ -f "/var/log/nginx/access.log" ]; then
            tail -20 /var/log/nginx/access.log
        else
            echo "Лог файл /var/log/nginx/access.log не найден"
        fi
        
        echo "=== Последние логи nginx (error) ==="
        if [ -f "/var/log/nginx/error.log" ]; then
            tail -20 /var/log/nginx/error.log
        else
            echo "Лог файл /var/log/nginx/error.log не найден"
        fi
        
        echo "=== Логи systemd для nginx ==="
        journalctl -u nginx --no-pager -n 20 || echo "Логи systemd для nginx не найдены"
EOF
}

# Проверка SSL сертификатов
check_ssl_certificates() {
    log_info "Проверка SSL сертификатов..."
    
    ssh $SERVER_USER@$SERVER_IP << 'EOF'
        echo "=== SSL сертификаты для edge.2gc.ru ==="
        if command -v openssl > /dev/null; then
            echo "Проверка сертификата через openssl:"
            echo | openssl s_client -servername edge.2gc.ru -connect edge.2gc.ru:443 2>/dev/null | openssl x509 -noout -dates 2>/dev/null || echo "Не удается получить сертификат"
        fi
        
        echo "=== Файлы сертификатов ==="
        find /etc/nginx -name "*.crt" -o -name "*.pem" -o -name "*.key" 2>/dev/null | head -10
        
        echo "=== Let's Encrypt сертификаты ==="
        if [ -d "/etc/letsencrypt/live/edge.2gc.ru" ]; then
            ls -la /etc/letsencrypt/live/edge.2gc.ru/
        else
            echo "Let's Encrypt сертификаты для edge.2gc.ru не найдены"
        fi
EOF
}

# Тестирование текущих endpoints
test_current_endpoints() {
    log_info "Тестирование текущих endpoints..."
    
    echo "=== Тест HTTP (80) ==="
    curl -I -s --connect-timeout 5 "http://edge.2gc.ru" || echo "HTTP (80) недоступен"
    
    echo "=== Тест HTTPS (443) ==="
    curl -I -k -s --connect-timeout 5 "https://edge.2gc.ru" || echo "HTTPS (443) недоступен"
    
    echo "=== Тест порта 8080 ==="
    curl -I -k -s --connect-timeout 5 "https://edge.2gc.ru:8080" || echo "Порт 8080 недоступен"
    
    echo "=== Тест порта 8082 ==="
    curl -I -k -s --connect-timeout 5 "https://edge.2gc.ru:8082" || echo "Порт 8082 недоступен"
    
    echo "=== Тест STUN порта 19302 ==="
    echo -e "\x00\x01\x00\x00\x21\x12\xa4\x42\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00" | nc -u -w 3 edge.2gc.ru 19302 | wc -c || echo "STUN порт 19302 недоступен"
}

# Создание рекомендаций по обновлению nginx
create_nginx_recommendations() {
    log_info "Создание рекомендаций по обновлению nginx..."
    
    cat > nginx-update-recommendations.md << 'EOF'
# 🔧 Рекомендации по обновлению nginx для QUIC + ICE/STUN/TURN

## 📋 **Текущая ситуация:**
- nginx работает как reverse proxy
- Новые порты (8080, 8082, 19302, 9090, 8443, 9092) не настроены
- Нужно добавить проксирование для новых сервисов

## 🎯 **Рекомендуемые изменения:**

### **1. Обновить конфигурацию nginx для edge.2gc.ru:**

```nginx
# /etc/nginx/sites-available/edge.2gc.ru
server {
    listen 80;
    listen 443 ssl http2;
    server_name edge.2gc.ru;
    
    # SSL конфигурация (существующая)
    ssl_certificate /etc/letsencrypt/live/edge.2gc.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/edge.2gc.ru/privkey.pem;
    
    # Основной API (существующий)
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # P2P API (новый)
    location /p2p/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    # ICE API (новый)
    location /ice/ {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Отдельный сервер для прямого доступа к портам (если нужно)
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
```

### **2. Альтернативный вариант - прямое проксирование портов:**

```nginx
# /etc/nginx/conf.d/relay-ports.conf
# HTTP API
upstream relay_http_api {
    server 127.0.0.1:8080;
}

# P2P API
upstream relay_p2p_api {
    server 127.0.0.1:8082;
}

server {
    listen 8080;
    server_name edge.2gc.ru;
    
    location / {
        proxy_pass http://relay_http_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 8082;
    server_name edge.2gc.ru;
    
    location / {
        proxy_pass http://relay_p2p_api;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### **3. Для UDP портов (STUN, QUIC):**
UDP порты (19302, 9090, 9092) должны быть доступны напрямую, без nginx проксирования.

## 🚀 **План обновления:**

1. **Создать бэкап текущей конфигурации:**
   ```bash
   sudo cp -r /etc/nginx /etc/nginx.backup.$(date +%Y%m%d_%H%M%S)
   ```

2. **Обновить конфигурацию nginx:**
   ```bash
   sudo nano /etc/nginx/sites-available/edge.2gc.ru
   ```

3. **Проверить конфигурацию:**
   ```bash
   sudo nginx -t
   ```

4. **Перезагрузить nginx:**
   ```bash
   sudo systemctl reload nginx
   ```

5. **Проверить работу:**
   ```bash
   curl -k "https://edge.2gc.ru/health"
   curl -k "https://edge.2gc.ru:8080/health"
   curl -k "https://edge.2gc.ru:8082/health"
   ```

## ⚠️ **Важные замечания:**

1. **SSL сертификаты:** Убедитесь, что SSL сертификаты действительны
2. **Firewall:** Проверьте, что все необходимые порты открыты
3. **Backend сервисы:** Убедитесь, что relay сервисы запущены на портах 8080 и 8082
4. **Логи:** Мониторьте логи nginx после изменений

## 🎯 **Ожидаемый результат:**

После обновления nginx:
- `https://edge.2gc.ru/health` → HTTP API (8080)
- `https://edge.2gc.ru:8080/health` → HTTP API (прямой доступ)
- `https://edge.2gc.ru:8082/health` → P2P API (прямой доступ)
- `edge.2gc.ru:19302` → STUN Server (UDP, прямой доступ)
- `edge.2gc.ru:9090` → QUIC Transport (UDP, прямой доступ)
- `edge.2gc.ru:8443` → MASQUE Proxy (HTTPS, прямой доступ)
- `edge.2gc.ru:9092` → Enhanced QUIC (UDP, прямой доступ)
EOF

    log_success "Рекомендации сохранены в nginx-update-recommendations.md"
}

# Основная функция
main() {
    log_info "🔍 Начинаем проверку конфигурации nginx на продакшен сервере"
    log_info "Сервер: $SERVER_IP"
    log_info "Пользователь: $SERVER_USER"
    
    # Проверки
    check_server_connection
    check_nginx_status
    check_nginx_configs
    check_current_services
    check_nginx_logs
    check_ssl_certificates
    test_current_endpoints
    create_nginx_recommendations
    
    log_success "🎉 Проверка завершена!"
    log_info "Рекомендации по обновлению nginx сохранены в nginx-update-recommendations.md"
    log_info "Следующий шаг: обновить конфигурацию nginx согласно рекомендациям"
}

# Запуск
main "$@"
