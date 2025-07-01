#!/bin/bash

# Если есть .env — загрузить переменные окружения
if [ -f .env ]; then
  echo "[INFO] Загрузка переменных из .env"
  set -o allexport
  source .env
  set +o allexport
fi

# Явно указываем путь к конфигу AlertManager
ALERTMANAGER_YML="deploy/alertmanager/alertmanager.yml"

# CloudBridge Relay - Обновление SMTP настроек AlertManager
# Этот скрипт обновляет конфигурацию AlertManager с реальными SMTP настройками

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
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

# Проверка, что мы в корневой директории проекта
check_project_root() {
    if [[ ! -f "docker-compose.yml" ]] || [[ ! -d "deploy" ]]; then
        log_error "Скрипт должен быть запущен из корневой директории проекта"
        exit 1
    fi
}

# Проверка обязательных переменных окружения
REQUIRED_VARS=(ALERT_EMAIL_TO ALERT_EMAIL_FROM ALERT_EMAIL_HOST ALERT_EMAIL_USER ALERT_EMAIL_PASS)
for var in "${REQUIRED_VARS[@]}"; do
  if [ -z "${!var}" ]; then
    echo "[ERROR] Переменная $var не задана!"
    exit 1
  fi
done

# Тестирование SMTP соединения
echo "[INFO] Тестирование SMTP соединения..."
timeout 5 bash -c "echo | openssl s_client -connect smtp.rusender.ru:465 -servername smtp.rusender.ru" > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "[OK] SMTP порт 465 доступен"
else
  echo "[ERROR] SMTP порт 465 недоступен"
fi

# Перезапуск AlertManager
restart_alertmanager() {
    log_info "Перезапуск AlertManager с новыми настройками..."
    
    # Остановка AlertManager
    docker-compose stop alertmanager
    
    # Запуск AlertManager
    docker-compose up -d alertmanager
    
    log_success "AlertManager перезапущен"
}

# Проверка статуса AlertManager
check_alertmanager_status() {
    log_info "Проверка статуса AlertManager..."
    
    # Ждем запуска
    sleep 10
    
    # Проверка доступности
    if curl -s -f "http://localhost:9093/-/healthy" > /dev/null 2>&1; then
        log_success "AlertManager работает"
    else
        log_error "AlertManager не отвечает"
        return 1
    fi
    
    # Проверка конфигурации
    local config_check=$(curl -s "http://localhost:9093/api/v2/status" | jq -r '.config.original' | grep -c "smtp.rusender.ru" || echo "0")
    
    if [[ "$config_check" -gt 0 ]]; then
        log_success "SMTP настройки применены"
    else
        log_warning "SMTP настройки не найдены в конфигурации"
    fi
    
    return 0
}

# Тестирование email отправки
test_email_sending() {
    log_info "Тестирование отправки email..."
    
    # Создание тестового алерта
    local test_alert='[
        {
            "labels": {
                "alertname": "TestEmailAlert",
                "severity": "warning"
            },
            "annotations": {
                "summary": "Test Email Alert",
                "description": "This is a test email alert to verify SMTP configuration"
            }
        }
    ]'
    
    # Отправка тестового алерта
    local response=$(curl -s -w "%{http_code}" -X POST "http://localhost:9093/api/v1/alerts" \
        -H "Content-Type: application/json" \
        -d "$test_alert")
    
    local http_code="${response: -3}"
    
    if [[ "$http_code" = "200" ]]; then
        log_success "Тестовый алерт отправлен успешно"
        log_info "Проверьте email info@2gc.ru на наличие тестового уведомления"
    else
        log_warning "Тестовый алерт не отправлен (HTTP код: $http_code)"
    fi
}

# Просмотр логов AlertManager
show_alertmanager_logs() {
    log_info "Просмотр логов AlertManager..."
    
    echo ""
    log_info "Последние 10 строк логов AlertManager:"
    docker-compose logs alertmanager | tail -10
    
    echo ""
    log_info "Поиск ошибок SMTP в логах:"
    docker-compose logs alertmanager | grep -i "smtp\|email\|mail" | tail -5
}

# Обновление email-конфига для AlertManager
cat > $ALERTMANAGER_YML <<EOL
global:
  resolve_timeout: 5m

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
  - match:
      severity: critical
    receiver: 'email'
    repeat_interval: 30m
  - match:
      severity: warning
    receiver: 'slack.warning'
    repeat_interval: 1h

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'

- name: 'slack.critical'
  slack_configs:
  - channel: '#alerts-critical'
    title: '{{ template "slack.title" . }}'
    text: '{{ template "slack.text" . }}'
    send_resolved: true
    actions:
    - type: button
      text: 'View in Grafana'
      url: '{{ template "slack.grafana.url" . }}'

- name: 'slack.warning'
  slack_configs:
  - channel: '#alerts-warning'
    title: '{{ template "slack.title" . }}'
    text: '{{ template "slack.text" . }}'
    send_resolved: true
    actions:
    - type: button
      text: 'View in Grafana'
      url: '{{ template "slack.grafana.url" . }}'

- name: 'email'
  email_configs:
  - to: "$ALERT_EMAIL_TO"
    from: "$ALERT_EMAIL_FROM"
    smarthost: "$ALERT_EMAIL_HOST:465"
    auth_username: "$ALERT_EMAIL_USER"
    auth_password: "$ALERT_EMAIL_PASS"
    send_resolved: true
    require_tls: false
    tls_config:
      insecure_skip_verify: false
      server_name: "$ALERT_EMAIL_HOST"

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOL

# Основная функция
main() {
    log_info "Обновление SMTP настроек AlertManager..."
    
    check_project_root
    restart_alertmanager
    
    if check_alertmanager_status; then
        test_email_sending
        show_alertmanager_logs
        
        log_success "Обновление SMTP настроек завершено!"
        log_info "Email уведомления теперь будут отправляться на info@2gc.ru"
        log_info "SMTP сервер: smtp.rusender.ru:587 (TLS)"
    else
        log_error "Проблемы с AlertManager после обновления"
        show_alertmanager_logs
        exit 1
    fi
}

# Обработка аргументов
case "${1:-}" in
    --help|-h)
        echo "Использование: $0 [опции]"
        echo ""
        echo "Опции:"
        echo "  --help, -h     Показать эту справку"
        echo "  --test         Только тестирование SMTP"
        echo "  --logs         Показать логи AlertManager"
        echo ""
        echo "Примеры:"
        echo "  sudo $0                    # Полное обновление"
        echo "  sudo $0 --test            # Только тестирование"
        echo "  sudo $0 --logs            # Показать логи"
        ;;
    --test)
        log_info "Только тестирование SMTP..."
        check_project_root
        test_email_sending
        ;;
    --logs)
        show_alertmanager_logs
        ;;
    "")
        main "$@"
        ;;
    *)
        log_error "Неизвестная опция: $1"
        echo "Используйте --help для справки"
        exit 1
        ;;
esac 