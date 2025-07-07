#!/bin/bash

echo "🚀 CloudBridge SIEM System Startup"
echo "=================================="
echo ""

# Функции логирования
log_info() {
    echo "[INFO] $1"
}

log_warning() {
    echo "[WARNING] $1"
}

log_error() {
    echo "[ERROR] $1"
}

log_success() {
    echo "[SUCCESS] $1"
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    # Проверка Python
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 не установлен"
        exit 1
    fi
    
    # Проверка Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker не установлен"
        exit 1
    fi
    
    # Проверка Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose не установлен"
        exit 1
    fi
    
    # Проверка портов
    if netstat -tuln 2>/dev/null | grep -q ":8081 "; then
        log_warning "Порт 8081 уже занят. Убедитесь, что SIEM не запущен."
    fi
    
    log_success "Зависимости проверены"
}

# Создание директорий
create_directories() {
    log_info "Создание директорий..."
    mkdir -p logs
    mkdir -p backups
    mkdir -p data
    mkdir -p config
    log_success "Директории созданы"
}

# Создание конфигурации
create_config() {
    if [ ! -f "config/siem_config.yaml" ]; then
        log_info "Создание конфигурации по умолчанию..."
        cat > config/siem_config.yaml << 'EOF'
# CloudBridge SIEM Configuration
collection:
  interval: 60
  sources:
    - relay
    - keycloak
    - system

detection:
  brute_force:
    enabled: true
    threshold: 5
    window: 300
  suspicious_ip:
    enabled: true
    blacklist_sources:
      - "https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset"
  unusual_activity:
    enabled: true
    threshold: 10

alerting:
  telegram:
    enabled: false
    bot_token: ""
    chat_id: ""
  email:
    enabled: false
    smtp_server: ""
    smtp_port: 587
    username: ""
    password: ""
    recipients: []
  slack:
    enabled: false
    webhook_url: ""

storage:
  elasticsearch:
    hosts: ["elasticsearch:9200"]
    index_prefix: "siem"
  retention:
    days: 90

monitoring:
  prometheus:
    enabled: true
    port: 9091
  grafana:
    enabled: true
    port: 3001
EOF
        log_success "Конфигурация создана"
    else
        log_info "Конфигурация уже существует"
    fi
}

# Установка зависимостей Python
install_python_deps() {
    log_info "Установка зависимостей Python..."
    python3 -m pip install --user -r requirements.txt
    if [ $? -eq 0 ]; then
        log_success "Зависимости установлены"
    else
        log_warning "Некоторые зависимости не установлены, продолжаем..."
    fi
}

# Запуск Docker Compose
start_docker_services() {
    log_info "Запуск Docker Compose..."
    docker-compose -f docker-compose.siem.yml up -d
    if [ $? -eq 0 ]; then
        log_success "Docker сервисы запущены"
    else
        log_error "Ошибка запуска Docker сервисов"
        exit 1
    fi
}

# Ожидание запуска сервисов
wait_for_services() {
    log_info "Ожидание запуска сервисов..."
    sleep 30
    
    # Проверка статуса
    log_info "Проверка статуса сервисов..."
    docker-compose -f docker-compose.siem.yml ps
}

# Создание индексов Elasticsearch
setup_elasticsearch() {
    log_info "Настройка Elasticsearch..."
    
    # Ожидание готовности Elasticsearch
    for i in {1..15}; do
        if curl -s http://localhost:9200/_cluster/health > /dev/null 2>&1; then
            log_success "Elasticsearch готов"
            break
        else
            log_info "Ожидание Elasticsearch... ($i/15)"
            sleep 5
        fi
    done
    
    # Создание индексов
    log_info "Создание индексов SIEM..."
    
    # Индекс событий
    curl -X PUT "localhost:9200/siem-events" -H "Content-Type: application/json" -d '{
      "mappings": {
        "properties": {
          "timestamp": {"type": "date"},
          "source": {"type": "keyword"},
          "event_type": {"type": "keyword"},
          "severity": {"type": "keyword"},
          "message": {"type": "text"},
          "ip_address": {"type": "ip"},
          "user_id": {"type": "keyword"},
          "indicators": {"type": "keyword"},
          "tags": {"type": "keyword"}
        }
      }
    }' 2>/dev/null && log_info "Индекс siem-events создан" || log_warning "Индекс siem-events уже существует"
    
    # Индекс уведомлений
    curl -X PUT "localhost:9200/siem-alerts" -H "Content-Type: application/json" -d '{
      "mappings": {
        "properties": {
          "timestamp": {"type": "date"},
          "alert_type": {"type": "keyword"},
          "severity": {"type": "keyword"},
          "message": {"type": "text"},
          "source": {"type": "keyword"},
          "status": {"type": "keyword"},
          "indicators": {"type": "keyword"}
        }
      }
    }' 2>/dev/null && log_info "Индекс siem-alerts создан" || log_warning "Индекс siem-alerts уже существует"
}

# Запуск тестов
run_tests() {
    log_info "Запуск тестов системы..."
    if command -v python3 &> /dev/null; then
        python3 test_siem.py
    else
        log_warning "Python3 не найден, пропускаем тесты"
    fi
}

# Основная функция
main() {
    check_dependencies
    create_directories
    create_config
    install_python_deps
    start_docker_services
    wait_for_services
    setup_elasticsearch
    run_tests
    
    echo ""
    echo "🎉 CloudBridge SIEM запущен!"
    echo ""
    echo "📊 Доступные сервисы:"
    echo "   - SIEM Web UI: http://localhost:8081"
    echo "   - Kibana: http://localhost:5601"
    echo "   - Grafana: http://localhost:3001"
    echo "   - Prometheus: http://localhost:9091"
    echo "   - Elasticsearch: http://localhost:9200"
    echo ""
    echo "📋 Управление:"
    echo "   - Статус: docker-compose -f docker-compose.siem.yml ps"
    echo "   - Логи: docker-compose -f docker-compose.siem.yml logs -f"
    echo "   - Остановка: docker-compose -f docker-compose.siem.yml down"
    echo ""
}

# Запуск основной функции
main 