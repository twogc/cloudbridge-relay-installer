#!/bin/bash

# Демонстрационный скрипт Threat Intelligence для CloudBridge Relay Security
# Показывает возможности интеграции с внешними сервисами Threat Intelligence

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Функции логирования
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo -e "${PURPLE}================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}================================${NC}"
}

# Проверка зависимостей
check_dependencies() {
    log_info "Проверка зависимостей..."
    
    if ! command -v python3 &> /dev/null; then
        log_error "Python3 не установлен"
        exit 1
    fi
    
    if ! command -v curl &> /dev/null; then
        log_error "curl не установлен"
        exit 1
    fi
    
    log_success "Все зависимости установлены"
}

# Демонстрация проверки IP адресов
demo_ip_check() {
    log_header "ДЕМОНСТРАЦИЯ: Проверка IP адресов"
    
    # Список тестовых IP адресов
    local test_ips=(
        "8.8.8.8"           # Google DNS (чистый)
        "185.220.101.1"     # Потенциально подозрительный
        "1.1.1.1"           # Cloudflare DNS (чистый)
        "192.168.1.1"       # Локальный IP
    )
    
    for ip in "${test_ips[@]}"; do
        log_info "Проверка IP: $ip"
        echo "----------------------------------------"
        python3 scripts/security/threat_intelligence.py --check-ip "$ip"
        echo ""
        sleep 1
    done
}

# Демонстрация проверки доменов
demo_domain_check() {
    log_header "ДЕМОНСТРАЦИЯ: Проверка доменов"
    
    # Список тестовых доменов
    local test_domains=(
        "google.com"        # Чистый домен
        "microsoft.com"     # Чистый домен
        "example.com"       # Тестовый домен
    )
    
    for domain in "${test_domains[@]}"; do
        log_info "Проверка домена: $domain"
        echo "----------------------------------------"
        python3 scripts/security/threat_intelligence.py --check-domain "$domain"
        echo ""
        sleep 1
    done
}

# Демонстрация проверки хешей файлов
demo_hash_check() {
    log_header "ДЕМОНСТРАЦИЯ: Проверка хешей файлов"
    
    # Список тестовых хешей
    local test_hashes=(
        "d41d8cd98f00b204e9800998ecf8427e"  # MD5 пустой строки
        "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"  # SHA256 пустой строки
    )
    
    for hash in "${test_hashes[@]}"; do
        log_info "Проверка хеша: $hash"
        echo "----------------------------------------"
        python3 scripts/security/threat_intelligence.py --check-hash "$hash"
        echo ""
        sleep 1
    done
}

# Демонстрация интеграции с Enhanced Security Monitor
demo_enhanced_monitor() {
    log_header "ДЕМОНСТРАЦИЯ: Enhanced Security Monitor"
    
    log_info "Запуск расширенного мониторинга безопасности..."
    echo "----------------------------------------"
    ./scripts/security/enhanced_security_monitor.sh --once
}

# Демонстрация статистики Threat Intelligence
demo_statistics() {
    log_header "ДЕМОНСТРАЦИЯ: Статистика Threat Intelligence"
    
    log_info "Получение статистики кэша..."
    echo "----------------------------------------"
    python3 -c "
from scripts.security.threat_intelligence import ThreatIntelligence, create_threat_intelligence_config

config = create_threat_intelligence_config()
ti = ThreatIntelligence(config)
stats = ti.get_cache_stats()
print(f'Размер кэша: {stats[\"cache_size\"]}')
print(f'TTL кэша: {stats[\"cache_ttl\"]} секунд')
"
}

# Демонстрация API лимитов
demo_api_limits() {
    log_header "ДЕМОНСТРАЦИЯ: Информация об API лимитах"
    
    echo "VirusTotal API:"
    echo "- Бесплатный план: 4 запроса в минуту"
    echo "- Платный план: 500 запросов в минуту"
    echo ""
    
    echo "AbuseIPDB API:"
    echo "- Бесплатный план: 1000 запросов в день"
    echo "- Платный план: Неограниченно"
    echo ""
    
    echo "AlienVault OTX API:"
    echo "- Бесплатный план: Неограниченно"
    echo "- Требует регистрации для получения ключа"
    echo ""
}

# Главная функция
main() {
    log_header "ДЕМОНСТРАЦИЯ THREAT INTELLIGENCE"
    log_info "CloudBridge Relay Security"
    log_info "Версия: 1.0"
    echo ""
    
    # Проверка зависимостей
    check_dependencies
    echo ""
    
    # Демонстрации
    demo_ip_check
    echo ""
    
    demo_domain_check
    echo ""
    
    demo_hash_check
    echo ""
    
    demo_statistics
    echo ""
    
    demo_api_limits
    echo ""
    
    # Спрашиваем пользователя о запуске Enhanced Monitor
    read -p "Запустить демонстрацию Enhanced Security Monitor? (y/n): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        demo_enhanced_monitor
    fi
    
    echo ""
    log_success "Демонстрация завершена!"
    log_info "Для получения справки используйте:"
    echo "  python3 scripts/security/threat_intelligence.py --help"
    echo "  ./scripts/security/enhanced_security_monitor.sh --help"
}

# Обработка аргументов командной строки
case "${1:-}" in
    --help|-h)
        echo "Демонстрационный скрипт Threat Intelligence"
        echo ""
        echo "Использование:"
        echo "  $0                    # Полная демонстрация"
        echo "  $0 --ip-only          # Только проверка IP"
        echo "  $0 --domain-only      # Только проверка доменов"
        echo "  $0 --hash-only        # Только проверка хешей"
        echo "  $0 --monitor-only     # Только Enhanced Monitor"
        echo "  $0 --help             # Эта справка"
        exit 0
        ;;
    --ip-only)
        check_dependencies
        demo_ip_check
        ;;
    --domain-only)
        check_dependencies
        demo_domain_check
        ;;
    --hash-only)
        check_dependencies
        demo_hash_check
        ;;
    --monitor-only)
        check_dependencies
        demo_enhanced_monitor
        ;;
    "")
        main
        ;;
    *)
        log_error "Неизвестный аргумент: $1"
        echo "Используйте --help для получения справки"
        exit 1
        ;;
esac 