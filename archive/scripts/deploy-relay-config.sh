#!/bin/bash

# Скрипт для развертывания nginx конфигурации на relay сервере (10.0.0.249)

set -e

RELAY_SERVER="10.0.0.249"
RELAY_USER="root"  # или другой пользователь с sudo правами
CONFIG_FILE="nginx-relay-config.conf"
REMOTE_CONFIG="/etc/nginx/sites-available/cloudbridge-2gc"

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Проверка доступности relay сервера
check_relay_server() {
    info "Проверяем доступность relay сервера $RELAY_SERVER..."
    
    if timeout 5 bash -c "</dev/tcp/$RELAY_SERVER/22" 2>/dev/null; then
        success "✓ SSH порт доступен на $RELAY_SERVER"
    else
        error "✗ SSH порт недоступен на $RELAY_SERVER"
        return 1
    fi
    
    if timeout 5 bash -c "</dev/tcp/$RELAY_SERVER/80" 2>/dev/null; then
        success "✓ HTTP порт доступен на $RELAY_SERVER"
    else
        warning "⚠ HTTP порт недоступен на $RELAY_SERVER"
    fi
}

# Резервное копирование существующей конфигурации
backup_existing_config() {
    info "Создаем резервную копию существующей конфигурации..."
    
    ssh $RELAY_USER@$RELAY_SERVER "
        if [ -f $REMOTE_CONFIG ]; then
            sudo cp $REMOTE_CONFIG ${REMOTE_CONFIG}.backup-$(date +%Y%m%d_%H%M%S)
            echo 'Резервная копия создана'
        else
            echo 'Существующая конфигурация не найдена'
        fi
    "
}

# Копирование конфигурации на relay сервер
deploy_config() {
    info "Копируем nginx конфигурацию на relay сервер..."
    
    # Копируем файл
    scp "$CONFIG_FILE" "$RELAY_USER@$RELAY_SERVER:/tmp/cloudbridge-2gc.conf"
    
    # Устанавливаем конфигурацию
    ssh $RELAY_USER@$RELAY_SERVER "
        sudo mv /tmp/cloudbridge-2gc.conf $REMOTE_CONFIG
        sudo chown root:root $REMOTE_CONFIG
        sudo chmod 644 $REMOTE_CONFIG
    "
    
    success "✓ Конфигурация скопирована"
}

# Проверка конфигурации nginx
test_nginx_config() {
    info "Проверяем конфигурацию nginx..."
    
    ssh $RELAY_USER@$RELAY_SERVER "
        sudo nginx -t
    " && success "✓ Конфигурация nginx корректна" || {
        error "✗ Ошибка в конфигурации nginx"
        return 1
    }
}

# Активация конфигурации
activate_config() {
    info "Активируем конфигурацию..."
    
    ssh $RELAY_USER@$RELAY_SERVER "
        # Удаляем default конфигурацию если есть
        sudo rm -f /etc/nginx/sites-enabled/default
        
        # Создаем симлинк
        sudo ln -sf $REMOTE_CONFIG /etc/nginx/sites-enabled/cloudbridge-2gc
        
        # Перезагружаем nginx
        sudo systemctl reload nginx
        
        # Проверяем статус
        sudo systemctl status nginx --no-pager -l
    "
    
    success "✓ Конфигурация активирована"
}

# Тестирование конфигурации
test_configuration() {
    info "Тестируем новую конфигурацию..."
    
    # Тест health endpoint
    info "Тестируем health endpoint..."
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null "http://$RELAY_SERVER/health" || echo "000")
    
    case $HTTP_CODE in
        200)
            success "✓ Health endpoint работает (HTTP $HTTP_CODE)"
            ;;
        404)
            warning "⚠ Health endpoint не найден (HTTP $HTTP_CODE)"
            ;;
        502|503|504)
            warning "⚠ Проблемы с backend (HTTP $HTTP_CODE)"
            ;;
        000)
            error "✗ Сервер недоступен"
            ;;
        *)
            warning "⚠ Неожиданный ответ (HTTP $HTTP_CODE)"
            ;;
    esac
    
    # Тест nginx status
    info "Тестируем nginx status..."
    HTTP_CODE=$(curl -s -w "%{http_code}" -o /dev/null "http://$RELAY_SERVER/nginx_status" || echo "000")
    
    if [ "$HTTP_CODE" = "200" ]; then
        success "✓ Nginx status доступен"
    else
        warning "⚠ Nginx status недоступен (HTTP $HTTP_CODE)"
    fi
}

# Проверка логов
check_logs() {
    info "Проверяем логи nginx..."
    
    ssh $RELAY_USER@$RELAY_SERVER "
        echo 'Последние записи в error.log:'
        sudo tail -n 10 /var/log/nginx/error.log 2>/dev/null || echo 'Лог файл не найден'
        
        echo ''
        echo 'Последние записи в access.log:'
        sudo tail -n 5 /var/log/nginx/access.log 2>/dev/null || echo 'Лог файл не найден'
    "
}

# Создание мониторинга
setup_monitoring() {
    info "Настраиваем мониторинг..."
    
    ssh $RELAY_USER@$RELAY_SERVER "
        # Создаем скрипт для мониторинга
        sudo tee /usr/local/bin/cloudbridge-monitor.sh > /dev/null << 'EOF'
#!/bin/bash
# CloudBridge Monitoring Script

# Проверяем nginx
if ! systemctl is-active --quiet nginx; then
    echo \"[ERROR] Nginx не запущен\"
    exit 1
fi

# Проверяем backend
if ! curl -s --max-time 5 http://192.168.49.2:32500/health > /dev/null; then
    echo \"[ERROR] Backend недоступен\"
    exit 1
fi

echo \"[OK] Все сервисы работают\"
EOF

        sudo chmod +x /usr/local/bin/cloudbridge-monitor.sh
        
        # Добавляем в cron (каждые 5 минут)
        echo '*/5 * * * * /usr/local/bin/cloudbridge-monitor.sh >> /var/log/cloudbridge-monitor.log 2>&1' | sudo crontab -
    "
    
    success "✓ Мониторинг настроен"
}

# Основная функция
main() {
    info "CloudBridge Relay Configuration Deployment"
    info "=========================================="
    info "Target server: $RELAY_SERVER"
    info "Config file: $CONFIG_FILE"
    echo ""
    
    if [ ! -f "$CONFIG_FILE" ]; then
        error "Файл конфигурации $CONFIG_FILE не найден"
        exit 1
    fi
    
    check_relay_server
    echo ""
    backup_existing_config
    echo ""
    deploy_config
    echo ""
    test_nginx_config
    echo ""
    activate_config
    echo ""
    test_configuration
    echo ""
    check_logs
    echo ""
    setup_monitoring
    
    success "Развертывание конфигурации завершено!"
    
    info "Следующие шаги:"
    echo "1. Протестировать все endpoints через LoadBalancer"
    echo "2. Проверить логи: ssh $RELAY_USER@$RELAY_SERVER 'sudo tail -f /var/log/nginx/cloudbridge-2gc-*.log'"
    echo "3. Мониторить статус: ssh $RELAY_USER@$RELAY_SERVER '/usr/local/bin/cloudbridge-monitor.sh'"
}

# Запуск
if [ "$1" = "--dry-run" ]; then
    info "DRY RUN MODE - только проверка доступности"
    check_relay_server
else
    main "$@"
fi
