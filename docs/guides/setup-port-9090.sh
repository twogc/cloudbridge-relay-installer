#!/bin/bash

# CloudBridge Relay - Настройка порта 9090 для Cloudflare Origin
# Скрипт для настройки CloudBridge Relay на порту 9090

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Функция для вывода сообщений
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN} $1${NC}"
}

warning() {
    echo -e "${YELLOW} $1${NC}"
}

error() {
    echo -e "${RED} $1${NC}"
}

# Проверка прав root
if [[ $EUID -ne 0 ]]; then
   error "Этот скрипт должен быть запущен с правами root"
   exit 1
fi

log "Настройка CloudBridge Relay на порту 9090 для Cloudflare Origin"

# 1. Проверка Magic WAN туннеля
log "Проверка Magic WAN туннеля..."
if sudo swanctl --list-sas | grep -q "ESTABLISHED"; then
    success "Magic WAN туннель активен"
else
    error "Magic WAN туннель не активен. Запустите туннель сначала."
    exit 1
fi

# 2. Проверка CloudBridge Relay
log "Проверка CloudBridge Relay..."
if kubectl get pods -n cloudbridge -l app=cloudbridge-relay | grep -q "Running"; then
    success "CloudBridge Relay поды работают"
else
    error "CloudBridge Relay поды не работают"
    exit 1
fi

# 3. Проверка nginx
log "Проверка nginx..."
if systemctl is-active --quiet nginx; then
    success "nginx работает"
else
    error "nginx не работает"
    exit 1
fi

# 4. Создание резервной копии текущей конфигурации nginx
log "💾 Создание резервной копии nginx конфигурации..."
cp /etc/nginx/conf.d/edge-2gc-ru.conf /etc/nginx/conf.d/edge-2gc-ru.conf.backup.$(date +%Y%m%d_%H%M%S)
success "Резервная копия создана"

# 5. Установка новой nginx конфигурации для порта 9090
log " Установка nginx конфигурации для порта 9090..."
if [ -f "nginx/cloudflare-port-9090.conf" ]; then
    cp nginx/cloudflare-port-9090.conf /etc/nginx/conf.d/edge-2gc-ru.conf
    success "nginx конфигурация обновлена для порта 9090"
else
    error "Файл nginx/cloudflare-port-9090.conf не найден"
    exit 1
fi

# 6. Проверка nginx конфигурации
log "Проверка nginx конфигурации..."
if nginx -t; then
    success "nginx конфигурация корректна"
else
    error "nginx конфигурация содержит ошибки"
    log "Восстановление резервной копии..."
    cp /etc/nginx/conf.d/edge-2gc-ru.conf.backup.$(date +%Y%m%d_%H%M%S) /etc/nginx/conf.d/edge-2gc-ru.conf
    exit 1
fi

# 7. Открытие порта 9090 в файрволе
log "🔥 Настройка файрвола для порта 9090..."
ufw allow 9090/tcp comment "CloudBridge Relay HTTPS"
success "Порт 9090 открыт в файрволе"

# 8. Перезагрузка nginx
log " Перезагрузка nginx..."
systemctl reload nginx
success "nginx перезагружен"

# 9. Проверка доступности endpoints на порту 9090
log "Проверка доступности endpoints на порту 9090..."

# Health check
if curl -k -s https://edge.2gc.ru:9090/health | grep -q "healthy\|ok"; then
    success "Health endpoint доступен на порту 9090"
else
    warning "Health endpoint недоступен на порту 9090"
fi

# API endpoint
if curl -k -s https://edge.2gc.ru:9090/api/ | grep -q "api\|json"; then
    success "API endpoint доступен на порту 9090"
else
    warning "API endpoint недоступен на порту 9090"
fi

# 10. Создание скрипта мониторинга для порта 9090
log "Создание скрипта мониторинга для порта 9090..."
cat > /usr/local/bin/cloudbridge-monitor-9090.sh << 'EOF'
#!/bin/bash

# CloudBridge Relay Monitor Script для порта 9090
# Мониторинг CloudBridge Relay и Magic WAN туннеля

echo "CloudBridge Relay Status (Port 9090):"
echo "========================================"

# Magic WAN туннель
echo "Magic WAN Tunnel:"
sudo swanctl --list-sas | grep -E "(ESTABLISHED|INSTALLED)" || echo " Tunnel not active"

# CloudBridge Relay поды
echo ""
echo "CloudBridge Relay Pods:"
kubectl get pods -n cloudbridge -l app=cloudbridge-relay

# nginx статус
echo ""
echo "nginx Status:"
systemctl is-active nginx

# Health check на порту 9090
echo ""
echo "💚 Health Check (Port 9090):"
curl -k -s https://edge.2gc.ru:9090/health || echo " Health check failed"

# API check на порту 9090
echo ""
echo " API Check (Port 9090):"
curl -k -s https://edge.2gc.ru:9090/api/ | head -c 100 || echo " API check failed"

# Проверка порта 9090
echo ""
echo " Port 9090 Check:"
netstat -tlnp | grep :9090 || echo " Port 9090 not listening"

echo ""
echo "Monitoring complete (Port 9090)"
EOF

chmod +x /usr/local/bin/cloudbridge-monitor-9090.sh
success "Скрипт мониторинга создан: /usr/local/bin/cloudbridge-monitor-9090.sh"

# 11. Создание systemd сервиса для мониторинга
log "Создание systemd сервиса для мониторинга..."
cat > /etc/systemd/system/cloudbridge-monitor-9090.service << 'EOF'
[Unit]
Description=CloudBridge Relay Monitor (Port 9090)
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/local/bin/cloudbridge-monitor-9090.sh
User=root

[Install]
WantedBy=multi-user.target
EOF

# 12. Создание cron задачи для мониторинга
log "⏰ Создание cron задачи для мониторинга..."
echo "*/5 * * * * /usr/local/bin/cloudbridge-monitor-9090.sh >> /var/log/cloudbridge-monitor-9090.log 2>&1" | crontab -
success "Cron задача создана"

# 13. Создание лог файла
touch /var/log/cloudbridge-monitor-9090.log
chmod 644 /var/log/cloudbridge-monitor-9090.log

# 14. Финальная проверка
log "Финальная проверка системы..."

# Проверка всех компонентов
echo ""
echo "Статус системы (Port 9090):"
echo "=============================="

# Magic WAN
if sudo swanctl --list-sas | grep -q "ESTABLISHED"; then
    success "Magic WAN туннель: АКТИВЕН"
else
    error "Magic WAN туннель: НЕ АКТИВЕН"
fi

# CloudBridge Relay
if kubectl get pods -n cloudbridge -l app=cloudbridge-relay | grep -q "Running"; then
    success "CloudBridge Relay: РАБОТАЕТ"
else
    error "CloudBridge Relay: НЕ РАБОТАЕТ"
fi

# nginx
if systemctl is-active --quiet nginx; then
    success "nginx: РАБОТАЕТ"
else
    error "nginx: НЕ РАБОТАЕТ"
fi

# Health endpoint на порту 9090
if curl -k -s https://edge.2gc.ru:9090/health | grep -q "healthy\|ok"; then
    success "Health endpoint (Port 9090): ДОСТУПЕН"
else
    warning "Health endpoint (Port 9090): НЕ ДОСТУПЕН"
fi

# API endpoint на порту 9090
if curl -k -s https://edge.2gc.ru:9090/api/ | grep -q "api\|json"; then
    success "API endpoint (Port 9090): ДОСТУПЕН"
else
    warning "API endpoint (Port 9090): НЕ ДОСТУПЕН"
fi

# Проверка порта 9090
if netstat -tlnp | grep -q :9090; then
    success "Порт 9090: ОТКРЫТ"
else
    error "Порт 9090: НЕ ОТКРЫТ"
fi

echo ""
log "Настройка CloudBridge Relay на порту 9090 завершена!"
echo ""
echo "Следующие шаги:"
echo "=================="
echo "1. Настройте Cloudflare Tunnel в Dashboard:"
echo "   - Hostname: app.client.company.ru"
echo "   - Service: https://edge.2gc.ru:9090"
echo ""
echo "2. Настройте cloudflared клиенты с конфигурацией:"
echo "   - Файл: cloudflared-client-configs/port-9090-client.yml"
echo ""
echo "3. Проверьте работу:"
echo "   - curl -k https://edge.2gc.ru:9090/health"
echo "   - curl -k https://edge.2gc.ru:9090/api/"
echo ""
echo "4. Мониторинг:"
echo "   - /usr/local/bin/cloudbridge-monitor-9090.sh"
echo "   - /var/log/cloudbridge-monitor-9090.log"
echo ""
echo "CloudBridge Relay готов к работе на порту 9090!"

